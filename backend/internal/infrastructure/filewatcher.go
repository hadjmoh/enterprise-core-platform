package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"path/filepath"
	"time"
)

type FileWatcher struct {
	paths  []string
	logger *logger.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func NewFileWatcher(paths []string, logger *logger.Logger) *FileWatcher {
	return &FileWatcher{
		paths:  paths,
		logger: logger,
	}
}

func (f *FileWatcher) Start() error {
	f.ctx, f.cancel = context.WithCancel(context.Background())

	for _, path := range f.paths {
		go f.watchFile(path)
	}

	f.logger.Info("File Watcher started", "paths", len(f.paths))
	return nil
}

func (f *FileWatcher) watchFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		f.logger.Error("Failed to open file", err)
		return
	}
	defer file.Close()

	// Seek to end of file
	file.Seek(0, os.SEEK_END)

	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-f.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				f.logger.Info("File watcher line",
					"file", filepath.Base(path),
					"length", len(line),
				)
				// TODO: Send to ingestion pipeline
			}
		}
	}
}

func (f *FileWatcher) Stop() {
	if f.cancel != nil {
		f.cancel()
		f.logger.Info("File Watcher stopped")
	}
}

func (f *FileWatcher) Protocol() string {
	return "FileWatcher"
}
