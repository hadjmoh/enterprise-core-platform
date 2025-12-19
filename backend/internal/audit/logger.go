package audit

import (
	"encoding/json"
	"enterprise-core/backend/pkg/logger"
	"os"
	"sync"
	"time"
)

// Logger provides immutable audit logging for governance
type Logger struct {
	filePath string
	logger   *logger.Logger
	mu       sync.Mutex
}

type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Status    string                 `json:"status"` // success, failure
	Details   map[string]interface{} `json:"details,omitempty"`
}

func NewLogger(filePath string, logger *logger.Logger) *Logger {
	return &Logger{
		filePath: filePath,
		logger:   logger,
	}
}

func (a *Logger) Log(actor, action, resource, status string, details map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now(),
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		Status:    status,
		Details:   details,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}

	a.logger.Info("Audit log entry created", "action", action, "actor", actor)
	return nil
}
