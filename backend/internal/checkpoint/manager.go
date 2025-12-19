package checkpoint

import (
	"encoding/json"
	"enterprise-core/backend/pkg/logger"
	"os"
	"sync"
	"time"
)

// Manager tracks processing offsets for replay capability
type Manager struct {
	checkpoints map[string]*Checkpoint
	filePath    string
	logger      *logger.Logger
	mu          sync.RWMutex
	autoSave    bool
	saveInterval time.Duration
	stopChan    chan struct{}
}

type Checkpoint struct {
	Source    string    `json:"source"`
	Offset    int64     `json:"offset"`
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func NewManager(filePath string, logger *logger.Logger) *Manager {
	return &Manager{
		checkpoints:  make(map[string]*Checkpoint),
		filePath:     filePath,
		logger:       logger,
		autoSave:     true,
		saveInterval: 10 * time.Second,
		stopChan:     make(chan struct{}),
	}
}

func (m *Manager) Start() error {
	// Load existing checkpoints
	if err := m.Load(); err != nil {
		m.logger.Error("Failed to load checkpoints", err)
	}

	// Start auto-save goroutine
	if m.autoSave {
		go m.autoSaveLoop()
	}

	m.logger.Info("Checkpoint Manager started", "file", m.filePath)
	return nil
}

func (m *Manager) UpdateCheckpoint(source string, offset int64, metadata map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkpoints[source] = &Checkpoint{
		Source:    source,
		Offset:    offset,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}

func (m *Manager) GetCheckpoint(source string) (*Checkpoint, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cp, ok := m.checkpoints[source]
	return cp, ok
}

func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.MarshalIndent(m.checkpoints, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		return err
	}

	m.logger.Info("Checkpoints saved", "count", len(m.checkpoints))
	return nil
}

func (m *Manager) Load() error {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No checkpoints file yet
		}
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := json.Unmarshal(data, &m.checkpoints); err != nil {
		return err
	}

	m.logger.Info("Checkpoints loaded", "count", len(m.checkpoints))
	return nil
}

func (m *Manager) autoSaveLoop() {
	ticker := time.NewTicker(m.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.Save(); err != nil {
				m.logger.Error("Auto-save failed", err)
			}
		case <-m.stopChan:
			return
		}
	}
}

func (m *Manager) Stop() error {
	close(m.stopChan)
	return m.Save()
}

func (m *Manager) GetAllCheckpoints() map[string]*Checkpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy
	result := make(map[string]*Checkpoint)
	for k, v := range m.checkpoints {
		result[k] = v
	}
	return result
}
