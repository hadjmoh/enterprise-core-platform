package wal

import (
	"bufio"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Log implements Write-Ahead Logging for durability
type Log struct {
	basePath    string
	currentFile *os.File
	writer      *bufio.Writer
	logger      *logger.Logger
	mu          sync.Mutex
	segmentSize int64
	currentSize int64
	sequence    int64
}

type Entry struct {
	Sequence  int64        `json:"sequence"`
	Timestamp time.Time    `json:"timestamp"`
	Event     buffer.Event `json:"event"`
}

func NewLog(basePath string, segmentSizeMB int, logger *logger.Logger) (*Log, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &Log{
		basePath:    basePath,
		logger:      logger,
		segmentSize: int64(segmentSizeMB) * 1024 * 1024,
		sequence:    0,
	}, nil
}

func (w *Log) Append(event buffer.Event) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Rotate if needed
	if w.currentFile == nil || w.currentSize >= w.segmentSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	// Create entry
	w.sequence++
	entry := Entry{
		Sequence:  w.sequence,
		Timestamp: time.Now(),
		Event:     event,
	}

	// Write to WAL
	data, err := json.Marshal(entry)
	if err != nil {
		return 0, err
	}

	data = append(data, '\n')
	n, err := w.writer.Write(data)
	if err != nil {
		return 0, err
	}

	// Flush to disk for durability
	if err := w.writer.Flush(); err != nil {
		return 0, err
	}

	w.currentSize += int64(n)
	return entry.Sequence, nil
}

func (w *Log) rotate() error {
	// Flush and close current file
	if w.writer != nil {
		w.writer.Flush()
	}
	if w.currentFile != nil {
		w.currentFile.Close()
	}

	// Create new segment
	timestamp := time.Now().Format("20060102-150405")
	filename := filepath.Join(w.basePath, fmt.Sprintf("wal-%s.log", timestamp))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	w.currentFile = file
	w.writer = bufio.NewWriter(file)
	w.currentSize = 0
	w.logger.Info("WAL rotated", "file", filename)

	return nil
}

func (w *Log) Replay(fromSequence int64, handler func(Entry) error) error {
	// Read all WAL files
	files, err := filepath.Glob(filepath.Join(w.basePath, "wal-*.log"))
	if err != nil {
		return err
	}

	for _, filename := range files {
		if err := w.replayFile(filename, fromSequence, handler); err != nil {
			return err
		}
	}

	return nil
}

func (w *Log) replayFile(filename string, fromSequence int64, handler func(Entry) error) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry Entry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			w.logger.Error("Failed to parse WAL entry", err)
			continue
		}

		if entry.Sequence > fromSequence {
			if err := handler(entry); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func (w *Log) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer != nil {
		w.writer.Flush()
	}
	if w.currentFile != nil {
		return w.currentFile.Close()
	}
	return nil
}
