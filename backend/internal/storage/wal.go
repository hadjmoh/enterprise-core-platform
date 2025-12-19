package storage

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"io"
	"os"
	"path/filepath"
)

// WAL handles write-ahead logging for a partition
type WAL struct {
	path string
	file *os.File
}

func OpenWAL(partitionPath string) (*WAL, error) {
	path := filepath.Join(partitionPath, "wal.log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &WAL{path: path, file: f}, nil
}

func (w *WAL) Append(event buffer.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = w.file.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	return w.file.Sync() // fsync for durability
}

func (w *WAL) Replay() ([]buffer.Event, error) {
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var events []buffer.Event
	dec := json.NewDecoder(w.file)
	for {
		var event buffer.Event
		if err := dec.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (w *WAL) Truncate() error {
	w.file.Close()
	return os.Remove(w.path)
}

func (w *WAL) Close() error {
	return w.file.Close()
}
