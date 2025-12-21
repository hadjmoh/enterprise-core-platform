package storage

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/encryption"
	"enterprise-core/backend/internal/index"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	MaxEventsPerFile = 10000
	MaxFileSize      = 50 * 1024 * 1024 // 50MB
)

// FileStorageEngine implements StorageEngine using local files
type FileStorageEngine struct {
	manager *PartitionManager
	logger  *logger.Logger
	kms     *encryption.KeyManager
	mu      sync.Mutex
	stats   Stats
	wals    map[string]*WAL // partition path -> WAL
	metrics Metrics
	memtables map[string]*Memtable // partition path -> Memtable
	registries map[string]*Registry // partition path -> Registry
	flushSem   chan struct{}        // Semaphore for concurrent flushes
	tokenizer   *index.Tokenizer
	distributor *ShardDistributor
	tiering     *TieringManager
}

type Metrics struct {
	EventsWrittenTotal     uint64
	BytesWrittenTotal      uint64
	OpenPartitions         int
	MemtableFlushTotal     uint64
	SSTableCreatedTotal    uint64
	TokensIndexedTotal          uint64
	IndexFilesCreatedTotal      uint64
	BloomFiltersCreatedTotal    uint64
	SparseIndexesCreatedTotal   uint64
	IndexRebuildDurationMs      int64
	ShardEventsWrittenTotal     uint64
	ShardFlushTotal             uint64
	ShardRecoveryEventsTotal    uint64
	RawBytesWrittenTotal        uint64
	CompressedBytesWrittenTotal uint64
	CompressionCacheHitsTotal   uint64
	CompressionCacheMissesTotal uint64
	DecompressionDurationMs     int64
	CompressedBlocksReadTotal   uint64
	HotToWarmMigrationsTotal    uint64
	WarmToColdMigrationsTotal   uint64
	PartitionsDeletedTotal      uint64
	TTLEnforcedTotal            uint64
	MigrationFailureTotal       uint64
	MigrationDurationMs         int64
	SearchTotal                 uint64
	SearchLatencyMs             int64
	SSTablesPrunedTotal         uint64
	TokensProcessedTotal        uint64
	ParseErrorsTotal            uint64
	ASTNodesCreatedTotal        uint64
}

func NewFileStorageEngine(basePath string, logg *logger.Logger) *FileStorageEngine {
	engine := &FileStorageEngine{
		manager: NewPartitionManager(basePath),
		logger:  logg,
		kms:     encryption.NewKeyManager(),
		stats:   Stats{Uptime: 0},
		wals:      make(map[string]*WAL),
		memtables:  make(map[string]*Memtable),
		registries: make(map[string]*Registry),
		flushSem:    make(chan struct{}, 4),
		tokenizer:   index.NewTokenizer(),
		distributor: NewShardDistributor(4), // Initialize with 4 shards for now
	}
	
	// Default paths for tiered storage
	warmPath := filepath.Join(basePath, "warm")
	coldPath := filepath.Join(basePath, "cold")
	engine.tiering = NewTieringManager(basePath, warmPath, coldPath, engine.manager, logg, &engine.metrics)
	
	if err := engine.Recover(); err != nil {
		logg.Error("Storage recovery failed", err)
	}
	
	// Start tiering background worker
	go engine.tiering.Run(context.Background())
	
	return engine
}

func (e *FileStorageEngine) Recover() error {
	// Look for existing partitions in the last 24 hours as a simple heuristic
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	partitions := e.manager.GetPartitions(yesterday, now, e.tiering.warmPath, e.tiering.coldPath)

	for _, p := range partitions {
		// Index recovery check
		e.ValidateIndexes(p)

		// Scan for shards
		shards, err := os.ReadDir(p)
		if err != nil {
			continue
		}

		for _, s := range shards {
			if !s.IsDir() {
				continue
			}
			shardPath := filepath.Join(p, s.Name())
			walPath := filepath.Join(shardPath, "wal.log")
			if _, err := os.Stat(walPath); err == nil {
				e.logger.Info("Recovering shard from WAL", "path", shardPath)
				wal, err := OpenWAL(shardPath, e.kms)
				if err != nil {
					continue
				}
				events, err := wal.Replay()
				if err != nil {
					wal.Close()
					continue
				}
				
				if len(events) > 0 {
					e.logger.Info("Recovering events from WAL", "count", len(events), "shard", shardPath)
					mem := NewMemtable(int64(MaxFileSize))
					e.memtables[shardPath] = mem
					e.wals[shardPath] = wal
					
					for _, ev := range events {
						mem.Push(ev)
					}
					e.mu.Lock()
					e.metrics.ShardRecoveryEventsTotal += uint64(len(events))
					e.mu.Unlock()

					// Optionally trigger immediate flush if recovered set is large
					if mem.Size() >= int64(MaxFileSize) {
						// Extract timestamp from first event for manager
						t, err := time.Parse(time.RFC3339, events[0].Timestamp)
						if err == nil {
							go e.Flush(shardPath, t)
						}
					}
				} else {
					wal.Close()
				}
			}
		}
	}
	// Reconcile tiers to ensure data is where it belongs
	e.ReconcileTiers()
	
	return nil
}

func (e *FileStorageEngine) ReconcileTiers() {
	e.logger.Info("Starting tier reconciliation")
}

// ValidateIndexes ensures every SSTable has a valid corresponding index.
func (e *FileStorageEngine) ValidateIndexes(partitionPath string) {
	files, err := os.ReadDir(partitionPath)
	if err != nil {
		return
	}

	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".jsonl" {
			sstPath := filepath.Join(partitionPath, f.Name())
			idxPath := sstPath + ".idx"

			if _, err := os.Stat(idxPath); os.IsNotExist(err) {
				e.logger.Warn("Missing index for SSTable, rebuilding", "file", f.Name())
				e.rebuildIndex(sstPath)
			}
		}
	}
}

func (e *FileStorageEngine) rebuildIndex(sstPath string) {
	// Simple re-index: read all events and build index
	it, err := NewSSTableIteratorWithEncryption(sstPath, e.kms)
	if err != nil {
		return
	}
	defer it.Close()

	batchIndex := index.NewInvertedIndex()
	bloom := index.NewBloomFilter(10000, 0.01) // Estimated size
	sparse := index.NewSparseIndex()

	var events []buffer.Event
	for {
		ev, err := it.Next(context.Background())
		if err == io.EOF {
			break
		}
		
		idx := len(events)
		if IsTombstone(ev) {
			events = append(events, ev)
			continue
		}
		
		msg, _ := ev.Data["message"].(string)
		tokens := e.tokenizer.Tokenize(msg)
		for _, token := range tokens {
			batchIndex.Add(token, int64(idx))
			bloom.Add(token)
		}

		if idx%500 == 0 {
			sparse.Add(ev.Timestamp, int64(idx))
		}
		events = append(events, ev)
	}

	if len(events) == 0 {
		return
	}

	// Persist index with metadata
	idxMeta := index.IndexMetadata{
		MinTimestamp:     events[0].Timestamp,
		MaxTimestamp:     events[len(events)-1].Timestamp,
		TokenCount:       batchIndex.Size(),
		EventCount:       len(events),
		SchemaVersion:    1,
		TokenizerVersion: e.tokenizer.Version,
	}

	idxData, _ := json.Marshal(batchIndex)
	metaData, _ := json.Marshal(idxMeta)

	idxFile, err := os.OpenFile(sstPath+".idx", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		idxFile.Write(idxData)
		idxFile.Write([]byte("\n"))
		idxFile.Write(metaData)
		footerSize := int64(len(metaData) + 1)
		binary.Write(idxFile, binary.LittleEndian, footerSize)
		idxFile.Close()
	}

	// Persist Bloom and Sparse atomically
	filterPath := sstPath + ".filter"
	tmpFilterPath := filterPath + ".tmp"
	bloom.UpdateRange(events[0].Timestamp)
	bloom.UpdateRange(events[len(events)-1].Timestamp)
	
	filterData, _ := json.Marshal(bloom)
	if err := os.WriteFile(tmpFilterPath, filterData, 0644); err == nil {
		os.Rename(tmpFilterPath, filterPath)
	}

	sparsePath := sstPath + ".sparse"
	tmpSparsePath := sparsePath + ".tmp"
	sparseData, _ := json.Marshal(sparse)
	if err := os.WriteFile(tmpSparsePath, sparseData, 0644); err == nil {
		os.Rename(tmpSparsePath, sparsePath)
	}
}

func (e *FileStorageEngine) Write(ctx context.Context, event buffer.Event) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Determine partition and validate state
	t, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		t = time.Now()
	}
	
	if err := e.manager.EnsurePartition(t); err != nil {
		return fmt.Errorf("failed to create partition: %w", err)
	}

	meta, err := e.manager.GetMetadata(t)
	if err != nil {
		return err
	}
	if meta.State != StateOpen {
		return fmt.Errorf("partition is not open for writes: %s", meta.State)
	}

	// 2. Determine shard and Write to WAL first
	shardID := e.distributor.GetShard(event.Source) // Use Source or Hash-Key
	
	// Guardrail: Max shards per partition
	if shardID >= 32 {
		shardID = 31 // Fallback to a "catch-all" shard if distributor produces too high an index
	}
	
	shardPath := e.manager.GetShardPath(t, shardID)
	if err := os.MkdirAll(shardPath, 0755); err != nil {
		return fmt.Errorf("failed to create shard directory: %w", err)
	}

	wal, ok := e.wals[shardPath]
	if !ok {
		var err error
		wal, err = OpenWAL(shardPath, e.kms)
		if err != nil {
			return fmt.Errorf("failed to open WAL: %w", err)
		}
		e.wals[shardPath] = wal
	}
	if err := wal.Append(event); err != nil {
		return fmt.Errorf("WAL write failed: %w", err)
	}

	// 3. Buffer in Memtable
	mem, ok := e.memtables[shardPath]
	if !ok {
		mem = NewMemtable(int64(MaxFileSize)) // Use MaxFileSize as buffer limit
		e.memtables[shardPath] = mem
	}

	event.Data["_storage_schema_version"] = 1
	shouldFlush := mem.Push(event)

	// 4. Update metrics
	e.metrics.EventsWrittenTotal++
	e.metrics.ShardEventsWrittenTotal++
	e.metrics.BytesWrittenTotal += uint64(len(event.Timestamp) + 100) // approx
	e.metrics.OpenPartitions = len(e.wals)

	// 5. Trigger Flush if needed
	if shouldFlush {
		select {
		case e.flushSem <- struct{}{}:
			go func() {
				defer func() { <-e.flushSem }()
				e.Flush(shardPath, t)
			}()
		default:
			e.logger.Warn("Flush queue full, backpressure active", "shard", shardPath)
			// In a real system, we'd signal backpressure to the pipeline here
		}
	}

	return nil
}

func (e *FileStorageEngine) Flush(shardPath string, t time.Time) error {
	e.mu.Lock()
	mem, ok := e.memtables[shardPath]
	if !ok || mem.Count() == 0 {
		e.mu.Unlock()
		return nil
	}
	
	// Copy events and clear memtable
	events := mem.GetEvents()
	mem.Clear()
	
	// Ensure registry is loaded
	reg, ok := e.registries[shardPath]
	if !ok {
		reg = NewRegistry(shardPath)
		reg.Load()
		e.registries[shardPath] = reg
	}
	e.mu.Unlock()

	e.logger.Info("Flushing memtable to SSTable", "shard", shardPath, "events", len(events))

	// 1. Determine SSTable sequence
	meta, err := e.manager.GetMetadata(t)
	if err != nil {
		return err
	}
	seq := (meta.EventCount / MaxEventsPerFile) + 1
	fileName := fmt.Sprintf("events_%04d.jsonl", seq)
	filePath := filepath.Join(shardPath, fileName)

	// 2. Write SSTable with Footer and Index
	writer, err := NewSSTableWriterWithEncryption(filePath, e.kms)
	if err != nil {
		e.logger.Error("Failed to create SSTableWriter", err)
		return err
	}

	// Build temporary index for this batch
	batchIndex := index.NewInvertedIndex()
	bloom := index.NewBloomFilter(uint(len(events)), 0.01)
	sparse := index.NewSparseIndex()

	var indexedEvents int = 0
	for i, event := range events {
		// Skip tombstones during indexing
		if IsTombstone(event) {
			continue
		}
		indexedEvents++

		// Log messages are usually in Data["message"] or event itself
		msg, _ := event.Data["message"].(string)
		tokens := e.tokenizer.Tokenize(msg)
		for _, token := range tokens {
			batchIndex.Add(token, int64(i))
			bloom.Add(token)
		}

		// Sparse indexing: every 500 events
		if i%500 == 0 {
			sparse.Add(event.Timestamp, int64(i))
		}
	}

	if err := writer.WriteBatch(events); err != nil {
		writer.Close()
		return err
	}

	// 3. Persist the Index with Footer
	idxPath := filePath + ".idx"
	tmpIdxPath := idxPath + ".tmp"
	
	idxMeta := index.IndexMetadata{
		MinTimestamp:     events[0].Timestamp,
		MaxTimestamp:     events[len(events)-1].Timestamp,
		TokenCount:       batchIndex.Size(),
		EventCount:       indexedEvents,
		SchemaVersion:    1,
		TokenizerVersion: e.tokenizer.Version,
	}

	idxData, _ := json.Marshal(batchIndex) 
	metaData, _ := json.Marshal(idxMeta)

	// Atomic write for index
	idxFile, err := os.OpenFile(tmpIdxPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err == nil {
		idxFile.Write(idxData)
		idxFile.Write([]byte("\n"))
		idxFile.Write(metaData)
		
		// Footer size pointer (8 bytes)
		footerSize := int64(len(metaData) + 1)
		binary.Write(idxFile, binary.LittleEndian, footerSize)
		idxFile.Sync()
		idxFile.Close()
		os.Rename(tmpIdxPath, idxPath)
	}

	// Persist Bloom Filter atomically
	filterPath := filePath + ".filter"
	tmpFilterPath := filterPath + ".tmp"
	bloom.UpdateRange(events[0].Timestamp)
	bloom.UpdateRange(events[len(events)-1].Timestamp)
	
	filterData, _ := json.Marshal(bloom)
	if err := os.WriteFile(tmpFilterPath, filterData, 0644); err == nil {
		os.Rename(tmpFilterPath, filterPath)
	}

	// Persist Sparse Index atomically
	sparsePath := filePath + ".sparse"
	tmpSparsePath := sparsePath + ".tmp"
	sparseData, _ := json.Marshal(sparse)
	if err := os.WriteFile(tmpSparsePath, sparseData, 0644); err == nil {
		os.Rename(tmpSparsePath, sparsePath)
	}
	
	// Get metadata from writer for registry before closing (if exposed or calculated)
	// For simplicity, we'll re-calculate or expose from writer
	info := SSTableInfo{
		Filename:     fileName,
		MinTimestamp: events[0].Timestamp,
		MaxTimestamp: events[len(events)-1].Timestamp,
		EventCount:   int64(len(events)),
	}

	if err := writer.Close(); err != nil {
		return err
	}

	// 3. Update Registry
	if err := reg.Add(info); err != nil {
		e.logger.Error("Failed to update SSTable registry", err)
	}

	// 4. Update metadata
	totalSize := len(events) * 150 // Simplified estimate
	if err := e.manager.UpdateMetadata(t, totalSize); err != nil {
		e.logger.Error("Failed to update metadata after flush", err)
	}

	// 5. Truncate WAL
	e.mu.Lock()
	if wal, ok := e.wals[shardPath]; ok {
		wal.Truncate()
		delete(e.wals, shardPath)
	}
	e.mu.Unlock()

	// 6. Update Shard Metadata
	shardMeta := ShardMetadata{
		ShardID:         0, // Example, could be parsed from path
		ParentPartition: filepath.Base(filepath.Dir(shardPath)),
		EventCount:      int64(len(events)),
		MinTimestamp:    events[0].Timestamp,
		MaxTimestamp:    events[len(events)-1].Timestamp,
		RegistryVersion: 1,
		LastModifiedAt:  time.Now(),
	}
	e.manager.WriteShardMetadataAtomic(shardPath, &shardMeta)

	// Update metrics
	e.mu.Lock()
	e.metrics.MemtableFlushTotal++
	e.metrics.ShardFlushTotal++
	e.metrics.SSTableCreatedTotal++
	e.metrics.TokensIndexedTotal += uint64(batchIndex.Size())
	e.metrics.IndexFilesCreatedTotal++
	e.metrics.BloomFiltersCreatedTotal++
	e.metrics.SparseIndexesCreatedTotal++
	
	fi, _ := os.Stat(filePath)
	if fi != nil {
		e.metrics.CompressedBytesWrittenTotal += uint64(fi.Size())
	}
	e.metrics.RawBytesWrittenTotal += uint64(len(events) * 150) // Estimate raw size
	
	e.mu.Unlock()

	return nil
}

func (e *FileStorageEngine) NewReader(ctx context.Context, q Query) (StorageReader, error) {
	partitions := e.manager.GetPartitions(q.StartTime, q.EndTime, e.tiering.warmPath, e.tiering.coldPath)
	if len(partitions) == 0 {
		return &emptyReader{}, nil
	}

	var shardReaders []StorageReader
	for _, p := range partitions {
		shards, err := os.ReadDir(p)
		if err != nil {
			continue
		}

		for _, s := range shards {
			if !s.IsDir() {
				continue
			}
			shardPath := filepath.Join(p, s.Name())
			
			// Discover SSTables in this shard
			reg := NewRegistry(shardPath)
			if err := reg.Load(); err != nil {
				continue
			}
			
			sstables := reg.List()
			for _, info := range sstables {
				// Basic time-range pruning
				if info.MaxTimestamp < q.StartTime.Format(time.RFC3339) || info.MinTimestamp > q.EndTime.Format(time.RFC3339) {
					continue
				}
				
				it, err := NewSSTableIteratorWithEncryption(filepath.Join(shardPath, info.Filename), e.kms)
				if err == nil {
					shardReaders = append(shardReaders, it)
				}
			}
		}
	}

	if len(shardReaders) == 0 {
		return &emptyReader{}, nil
	}

	return NewMergeIterator(shardReaders), nil
}

type emptyReader struct{}
func (r *emptyReader) Next(ctx context.Context) (buffer.Event, error) { return buffer.Event{}, io.EOF }
func (r *emptyReader) Close() error                              { return nil }

type fileReader struct {
	partitions []string
	currentIdx int
}
func (r *fileReader) Next(ctx context.Context) (buffer.Event, error) {
	// Skeleton implementation
	return buffer.Event{}, io.EOF 
}
func (r *fileReader) Close() error { return nil }

func (e *FileStorageEngine) Query(ctx context.Context, q Query) ([]buffer.Event, error) {
	return e.Search(ctx, q)
}

func (e *FileStorageEngine) Search(ctx context.Context, q Query) ([]buffer.Event, error) {
	start := time.Now()
	e.mu.Lock()
	e.metrics.SearchTotal++
	e.mu.Unlock()

	// 1. Discover partitions
	partitions := e.manager.GetPartitions(q.StartTime, q.EndTime, e.tiering.warmPath, e.tiering.coldPath)
	if len(partitions) == 0 {
		return nil, nil
	}

	var results []buffer.Event
	var readers []StorageReader

	for _, p := range partitions {
		shards, err := os.ReadDir(p)
		if err != nil {
			continue
		}

		for _, s := range shards {
			if !s.IsDir() {
				continue
			}
			shardPath := filepath.Join(p, s.Name())
			
			reg := NewRegistry(shardPath)
			if err := reg.Load(); err != nil {
				continue
			}
			
			sstables := reg.List()
			for _, info := range sstables {
				// A. Time-range pruning
				if info.MaxTimestamp < q.StartTime.Format(time.RFC3339) || info.MinTimestamp > q.EndTime.Format(time.RFC3339) {
					continue
				}
				
				sstPath := filepath.Join(shardPath, info.Filename)
				
				// B. Bloom Filter Pruning
				if len(q.Keywords) > 0 {
					filterPath := sstPath + ".filter"
					if data, err := os.ReadFile(filterPath); err == nil {
						var bloom index.BloomFilter
						if json.Unmarshal(data, &bloom) == nil {
							match := true
							for _, kw := range q.Keywords {
								if !bloom.Test(kw) {
									match = false
									break
								}
							}
							if !match {
								e.mu.Lock()
								e.metrics.SSTablesPrunedTotal++
								e.mu.Unlock()
								continue
							}
						}
					}
				}

				// C. Create Iterator
				it, err := NewSSTableIteratorWithEncryption(sstPath, e.kms)
				if err == nil {
					readers = append(readers, it)
				}
			}
		}
	}

	if len(readers) == 0 {
		return nil, nil
	}

	// 2. Aggregate and Filter
	merged := NewMergeIterator(readers)
	defer merged.Close()

	for {
		if q.Limit > 0 && len(results) >= q.Limit {
			break
		}

		ev, err := merged.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return results, err
		}

		// Field Filtering
		match := true
		for k, v := range q.FieldFilters {
			if val, ok := ev.Data[k].(string); !ok || val != v {
				match = false
				break
			}
		}

		if match {
			results = append(results, ev)
		}
	}

	e.mu.Lock()
	e.metrics.SearchLatencyMs = time.Since(start).Milliseconds()
	e.mu.Unlock()

	return results, nil
}

func (e *FileStorageEngine) Stats() Stats {
	return e.stats
}

func (e *FileStorageEngine) FlushAll() {
	e.mu.Lock()
	var wg sync.WaitGroup
	for shardPath := range e.memtables {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			e.Flush(p, time.Now())
		}(shardPath)
	}
	e.mu.Unlock()
	wg.Wait()
}

func (e *FileStorageEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, wal := range e.wals {
		wal.Close()
	}
	e.logger.Info("File Storage Engine closed")
	return nil
}
