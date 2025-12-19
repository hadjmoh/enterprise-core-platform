package writer

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DiskWriter persists events to disk in rotating segments
type DiskWriter struct {
	basePath      string
	currentFile   *os.File
	currentSize   int64
	maxSegmentSize int64
	logger        *logger.Logger
	mu            sync.Mutex
}

func NewDiskWriter(basePath string, maxSegmentSizeMB int, logger *logger.Logger) (*DiskWriter, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &DiskWriter{
		basePath:       basePath,
		maxSegmentSize: int64(maxSegmentSizeMB) * 1024 * 1024,
		logger:         logger,
	}, nil
}

func (d *DiskWriter) Handle(event buffer.Event) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Rotate if needed
	if d.currentFile == nil || d.currentSize >= d.maxSegmentSize {
		if err := d.rotate(); err != nil {
			return err
		}
	}

	// Write event as JSON line
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	n, err := d.currentFile.Write(data)
	if err != nil {
		return err
	}

	d.currentSize += int64(n)
	return nil
}

func (d *DiskWriter) rotate() error {
	// Close current file
	if d.currentFile != nil {
		d.currentFile.Close()
	}

	// Create new segment file
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(d.basePath, fmt.Sprintf("events-%s.jsonl", timestamp))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	d.currentFile = file
	d.currentSize = 0
	d.logger.Info("Rotated to new segment", "file", filename)

	return nil
}

func (d *DiskWriter) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.currentFile != nil {
		return d.currentFile.Close()
	}
	return nil
}
