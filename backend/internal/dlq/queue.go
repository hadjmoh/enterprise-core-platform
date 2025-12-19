package dlq

import (
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Queue stores failed events for manual review
type Queue struct {
	basePath string
	logger   *logger.Logger
	mu       sync.Mutex
	count    int64
}

type FailedEvent struct {
	Event     buffer.Event `json:"event"`
	Error     string       `json:"error"`
	Timestamp time.Time    `json:"timestamp"`
	Retries   int          `json:"retries"`
}

func NewQueue(basePath string, logger *logger.Logger) (*Queue, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &Queue{
		basePath: basePath,
		logger:   logger,
	}, nil
}

func (q *Queue) Add(event buffer.Event, err error, retries int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	failed := FailedEvent{
		Event:     event,
		Error:     err.Error(),
		Timestamp: time.Now(),
		Retries:   retries,
	}

	// Write to file
	filename := filepath.Join(q.basePath, time.Now().Format("20060102-150405.000000")+".json")
	data, marshalErr := json.MarshalIndent(failed, "", "  ")
	if marshalErr != nil {
		return marshalErr
	}

	if writeErr := os.WriteFile(filename, data, 0644); writeErr != nil {
		return writeErr
	}

	q.count++
	q.logger.Info("Event added to DLQ", "file", filename, "error", err.Error())
	return nil
}

func (q *Queue) List() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(q.basePath, "*.json"))
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (q *Queue) Get(filename string) (*FailedEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var failed FailedEvent
	if err := json.Unmarshal(data, &failed); err != nil {
		return nil, err
	}

	return &failed, nil
}

func (q *Queue) Remove(filename string) error {
	return os.Remove(filename)
}

func (q *Queue) Size() (int64, error) {
	files, err := q.List()
	if err != nil {
		return 0, err
	}
	return int64(len(files)), nil
}

func (q *Queue) Reprocess(filename string, handler func(buffer.Event) error) error {
	failed, err := q.Get(filename)
	if err != nil {
		return err
	}

	if err := handler(failed.Event); err != nil {
		// Update retries and re-add
		return q.Add(failed.Event, err, failed.Retries+1)
	}

	// Success - remove from DLQ
	return q.Remove(filename)
}
