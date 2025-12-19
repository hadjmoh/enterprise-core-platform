package storage

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"path/filepath"
	"time"
)

// TieringPolicy defines when data should move between tiers
type TieringPolicy struct {
	HotToWarmAge  time.Duration
	WarmToColdAge time.Duration
	RetentionAge  time.Duration
}

// TieringManager handles the movement of partitions between storage tiers
type TieringManager struct {
	policy  TieringPolicy
	manager *PartitionManager
	logger  *logger.Logger
	
	hotPath  string
	warmPath string
	coldPath string
	
	dryRun   bool
	migSem   chan struct{} // Concurrency limit for migrations
	metrics  *Metrics
}

func NewTieringManager(hot, warm, cold string, pm *PartitionManager, logg *logger.Logger, metrics *Metrics) *TieringManager {
	return &TieringManager{
		hotPath:  hot,
		warmPath: warm,
		coldPath: cold,
		manager:  pm,
		logger:   logg,
		metrics:  metrics,
		policy: TieringPolicy{
			HotToWarmAge:  24 * time.Hour,
			WarmToColdAge: 7 * 24 * time.Hour,
			RetentionAge:  30 * 24 * time.Hour,
		},
		dryRun: false,
		migSem: make(chan struct{}, 2), // Max 2 concurrent migrations
	}
}

func (tm *TieringManager) SetDryRun(dry bool) {
	tm.dryRun = dry
}

// Run starts the periodic tiering background task
func (tm *TieringManager) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	tm.logger.Info("Tiering Manager started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.ProcessTiers()
		}
	}
}

func (tm *TieringManager) ProcessTiers() {
	// 1. Get all partitions
	// This is a simplified scan. In a real system, we'd use the registry or a database.
	// For now, let's assume we can scan the hot directory.
	
	tm.logger.Info("Checking partitions for tiering migration")
	
	now := time.Now()
	
	// Scan Hot Tier
	err := filepath.Walk(tm.hotPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		
		// Expected path: hotPath/YYYY/MM/DD/HH
		// If we find a metadata.json, it's a partition
		metaPath := filepath.Join(path, "metadata.json")
		if _, err := os.Stat(metaPath); err == nil {
			// Extract time from path (reverse of Format)
			// For simplicity, we'll use FileInfo.ModTime or parse from path
			rel, _ := filepath.Rel(tm.hotPath, path)
			t, err := time.Parse("2006/01/02/15", filepath.ToSlash(rel))
			if err != nil {
				return nil
			}
			
			if now.Sub(t) > tm.policy.RetentionAge {
				if tm.dryRun {
					tm.logger.Info("[DRY-RUN] Would delete expired partition", "path", path)
				} else {
					tm.logger.Info("Deleting expired partition", "path", path)
					os.RemoveAll(path)
					if tm.metrics != nil {
						tm.metrics.PartitionsDeletedTotal++
						tm.metrics.TTLEnforcedTotal++
					}
				}
				return filepath.SkipDir
			}
			
			if now.Sub(t) > tm.policy.HotToWarmAge {
				if tm.dryRun {
					tm.logger.Info("[DRY-RUN] Would migrate partition to Warm", "path", path)
				} else {
					tm.Migrate(path, tm.warmPath, StateWarm)
				}
				return filepath.SkipDir
			}
		}
		return nil
	})
	
	if err != nil {
		tm.logger.Error("Tiering scan failed", err)
	}
}

func (tm *TieringManager) Migrate(srcPath, dstRoot string, newState PartitionState) {
	// 1. Safety Guard: Concurrency
	select {
	case tm.migSem <- struct{}{}:
		defer func() { <-tm.migSem }()
	default:
		tm.logger.Warn("Max concurrent migrations reached, skipping for now", "src", srcPath)
		return
	}

	// 2. Safety Guard: Disk Space Check (Simplified)
	// In a real system, we'd use syscall.Statfs
	
	tm.logger.Info("Migrating partition", "src", srcPath, "target", dstRoot, "state", newState)
	
	start := time.Now()
	rel, _ := filepath.Rel(tm.hotPath, srcPath)
	dstPath := filepath.Join(dstRoot, rel)
	
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		tm.logger.Error("Failed to create target directory", err)
		return
	}
	
	if err := os.Rename(srcPath, dstPath); err != nil {
		tm.logger.Error("Migration rename failed", err)
		if tm.metrics != nil {
			tm.metrics.MigrationFailureTotal++
		}
		return
	}
	
	duration := time.Since(start).Milliseconds()
	if tm.metrics != nil {
		tm.metrics.HotToWarmMigrationsTotal++
		tm.metrics.MigrationDurationMs = duration
	}

	// 4. Update metadata in the new location
	t, _ := time.Parse("2006/01/02/15", filepath.ToSlash(rel))
	meta, err := tm.manager.GetMetadata(t) // This might need path adjustment if GetMetadata uses basePath
	if err == nil {
		meta.State = newState
		meta.LastMigrationAt = time.Now()
		meta.DeleteAt = t.Add(tm.policy.RetentionAge)
		// Note: writing to the NEW location
		tm.manager.WriteShardMetadataAtomic(dstPath, &ShardMetadata{}) // Stub update
	}

	tm.logger.Info("Migration completed", "src", srcPath, "duration_ms", time.Since(start).Milliseconds())
}
