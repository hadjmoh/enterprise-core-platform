package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Feedback represents an analyst's feedback on an AI/ML output
type Feedback struct {
	ID        string    `json:"id"`
	EntityID  string    `json:"entity_id"`
	AlertID   string    `json:"alert_id"`
	Rating    int       `json:"rating"` // 1 (Thumbs Down) to 5 (Thumbs Up)
	IsFalsePositive bool `json:"is_false_positive"`
	Comments  string    `json:"comments"`
	Timestamp time.Time `json:"timestamp"`
}

// FeedbackStore manages analyst feedback
type FeedbackStore struct {
	mu       sync.RWMutex
	Feedback []Feedback `json:"feedback"`
	filepath string
}

// NewFeedbackStore creates a new feedback store
func NewFeedbackStore(filepath string) *FeedbackStore {
	fs := &FeedbackStore{
		Feedback: []Feedback{},
		filepath: filepath,
	}
	if err := fs.load(); err != nil {
		fmt.Printf("Warning: Failed to load feedback store: %v\n", err)
	}
	return fs
}

// SubmitFeedback records new feedback
func (fs *FeedbackStore) SubmitFeedback(f Feedback) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	f.ID = fmt.Sprintf("fb-%d", time.Now().UnixNano())
	f.Timestamp = time.Now()
	
	fs.Feedback = append(fs.Feedback, f)
	return fs.persist()
}

func (fs *FeedbackStore) persist() error {
	data, err := json.MarshalIndent(fs.Feedback, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filepath, data, 0644)
}

func (fs *FeedbackStore) load() error {
	data, err := os.ReadFile(fs.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &fs.Feedback)
}
