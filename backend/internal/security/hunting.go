package security

import (
	"sync"
	"time"
)

// Hunt represents a specialized proactive search with a hypothesis
type Hunt struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	SPL         string    `json:"spl"`
	Description string    `json:"description"`
	Hypothesis  string    `json:"hypothesis"`
	Status      string    `json:"status"` // draft, active, completed
	Findings    string    `json:"findings"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NotebookCell represents a single unit of work in a notebook
type NotebookCell struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // markdown, spl
	Content string `json:"content"`
}

// Notebook represents a collaborative investigation document
type Notebook struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Author    string         `json:"author"`
	Cells     []NotebookCell `json:"cells"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// HuntingManager manages the persistence and lifecycle of hunts and notebooks
type HuntingManager struct {
	hunts     map[string]*Hunt
	notebooks map[string]*Notebook
	mu        sync.RWMutex
}

func NewHuntingManager() *HuntingManager {
	return &HuntingManager{
		hunts:     make(map[string]*Hunt),
		notebooks: make(map[string]*Notebook),
	}
}

func (m *HuntingManager) SaveHunt(h *Hunt) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h.UpdatedAt = time.Now()
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now()
	}
	m.hunts[h.ID] = h
}

func (m *HuntingManager) GetHunts() []*Hunt {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make([]*Hunt, 0, len(m.hunts))
	for _, h := range m.hunts {
		results = append(results, h)
	}
	return results
}

func (m *HuntingManager) SaveNotebook(nb *Notebook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	nb.UpdatedAt = time.Now()
	if nb.CreatedAt.IsZero() {
		nb.CreatedAt = time.Now()
	}
	m.notebooks[nb.ID] = nb
}

func (m *HuntingManager) GetNotebooks() []*Notebook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	results := make([]*Notebook, 0, len(m.notebooks))
	for _, nb := range m.notebooks {
		results = append(results, nb)
	}
	return results
}

func (m *HuntingManager) GetNotebook(id string) (*Notebook, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nb, ok := m.notebooks[id]
	return nb, ok
}
