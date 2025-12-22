package audit

import (
	"encoding/json"
	"enterprise-core/backend/pkg/logger"
	"os"
	"sync"
	"time"
)

// AuditLogger provides immutable audit logging for governance
type AuditLogger struct {
	filePath string
	logger   *logger.Logger
	mu       sync.Mutex
}

type Entry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Actor      string                 `json:"actor"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Status     string                 `json:"status"`      // success, failure
	ResultHash string                 `json:"result_hash,omitempty"` // Cumulative SHA-256 of results
	Details    map[string]interface{} `json:"details,omitempty"`
}

func NewLogger(filePath string, logger *logger.Logger) *AuditLogger {
	return &AuditLogger{
		filePath: filePath,
		logger:   logger,
	}
}

func (a *AuditLogger) Log(actor, action, resource, status string, resultHash string, details map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := Entry{
		Timestamp:  time.Now(),
		Actor:      actor,
		Action:     action,
		Resource:   resource,
		Status:     status,
		ResultHash: resultHash,
		Details:    details,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Use O_SYNC to ensure log durability
	f, err := os.OpenFile(a.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}

	a.logger.Info("Audit log entry created", "action", action, "actor", actor, "status", status)
	return nil
}
