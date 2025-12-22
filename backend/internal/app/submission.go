package app

import (
	"fmt"
	"sync"
	"time"

	"enterprise-core/backend/pkg/logger"
)

// SubmissionStatus represents the state of a submission
type SubmissionStatus string

const (
	StatusPending   SubmissionStatus = "pending"
	StatusReviewing SubmissionStatus = "reviewing"
	StatusApproved  SubmissionStatus = "approved"
	StatusRejected  SubmissionStatus = "rejected"
	StatusPublished SubmissionStatus = "published"
)

// Submission represents an app submission
type Submission struct {
	ID               string            `json:"id"`
	AppName          string            `json:"app_name"`
	Manifest         *AppManifest      `json:"manifest"`
	BundlePath       string            `json:"-"` // Don't expose file path
	Status           SubmissionStatus  `json:"status"`
	Submitter        string            `json:"submitter"`
	SubmittedAt      time.Time         `json:"submitted_at"`
	ReviewedAt       *time.Time        `json:"reviewed_at,omitempty"`
	ReviewedBy       string            `json:"reviewed_by,omitempty"`
	ReviewNotes      string            `json:"review_notes,omitempty"`
	ValidationResult *ValidationResult `json:"validation_result"`
}

// SubmissionManager manages app submissions
type SubmissionManager struct {
	mu          sync.RWMutex
	submissions map[string]*Submission // submission ID -> submission
	byStatus    map[SubmissionStatus][]*Submission
	validator   *Validator
	logger      *logger.Logger
}

// NewSubmissionManager creates a new submission manager
func NewSubmissionManager(log *logger.Logger) *SubmissionManager {
	return &SubmissionManager{
		submissions: make(map[string]*Submission),
		byStatus:    make(map[SubmissionStatus][]*Submission),
		validator:   NewValidator(),
		logger:      log,
	}
}

// SubmitApp creates a new submission
func (sm *SubmissionManager) SubmitApp(manifest *AppManifest, bundlePath, submitter string) (*Submission, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate the app
	validationResult := sm.validator.ValidateAll(manifest, bundlePath)

	submission := &Submission{
		ID:               generateSubmissionID(),
		AppName:          manifest.Metadata.Name,
		Manifest:         manifest,
		BundlePath:       bundlePath,
		Status:           StatusPending,
		Submitter:        submitter,
		SubmittedAt:      time.Now(),
		ValidationResult: validationResult,
	}

	sm.submissions[submission.ID] = submission
	sm.byStatus[StatusPending] = append(sm.byStatus[StatusPending], submission)

	sm.logger.Info(fmt.Sprintf("New app submission: %s (ID: %s)", manifest.Metadata.Name, submission.ID))
	return submission, nil
}

// GetSubmission retrieves a submission by ID
func (sm *SubmissionManager) GetSubmission(id string) (*Submission, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	submission, exists := sm.submissions[id]
	if !exists {
		return nil, fmt.Errorf("submission not found: %s", id)
	}

	return submission, nil
}

// ListSubmissions returns submissions filtered by status
func (sm *SubmissionManager) ListSubmissions(status SubmissionStatus) []*Submission {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if status == "" {
		// Return all submissions
		all := make([]*Submission, 0, len(sm.submissions))
		for _, sub := range sm.submissions {
			all = append(all, sub)
		}
		return all
	}

	// Return submissions with specific status
	submissions := sm.byStatus[status]
	result := make([]*Submission, len(submissions))
	copy(result, submissions)
	return result
}

// ApproveSubmission approves a submission
func (sm *SubmissionManager) ApproveSubmission(id, reviewer, notes string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	submission, exists := sm.submissions[id]
	if !exists {
		return fmt.Errorf("submission not found: %s", id)
	}

	if submission.Status != StatusPending && submission.Status != StatusReviewing {
		return fmt.Errorf("submission cannot be approved in status: %s", submission.Status)
	}

	// Update status
	sm.removeFromStatusList(submission.Status, id)
	submission.Status = StatusApproved
	now := time.Now()
	submission.ReviewedAt = &now
	submission.ReviewedBy = reviewer
	submission.ReviewNotes = notes
	sm.byStatus[StatusApproved] = append(sm.byStatus[StatusApproved], submission)

	sm.logger.Info(fmt.Sprintf("Submission approved: %s by %s", id, reviewer))
	return nil
}

// RejectSubmission rejects a submission
func (sm *SubmissionManager) RejectSubmission(id, reviewer, notes string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	submission, exists := sm.submissions[id]
	if !exists {
		return fmt.Errorf("submission not found: %s", id)
	}

	if submission.Status == StatusApproved || submission.Status == StatusPublished {
		return fmt.Errorf("cannot reject submission in status: %s", submission.Status)
	}

	// Update status
	sm.removeFromStatusList(submission.Status, id)
	submission.Status = StatusRejected
	now := time.Now()
	submission.ReviewedAt = &now
	submission.ReviewedBy = reviewer
	submission.ReviewNotes = notes
	sm.byStatus[StatusRejected] = append(sm.byStatus[StatusRejected], submission)

	sm.logger.Info(fmt.Sprintf("Submission rejected: %s by %s", id, reviewer))
	return nil
}

// UpdateStatus changes the submission status
func (sm *SubmissionManager) UpdateStatus(id string, status SubmissionStatus) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	submission, exists := sm.submissions[id]
	if !exists {
		return fmt.Errorf("submission not found: %s", id)
	}

	sm.removeFromStatusList(submission.Status, id)
	submission.Status = status
	sm.byStatus[status] = append(sm.byStatus[status], submission)

	return nil
}

// Helper to remove submission from status list
func (sm *SubmissionManager) removeFromStatusList(status SubmissionStatus, id string) {
	list := sm.byStatus[status]
	for i, sub := range list {
		if sub.ID == id {
			sm.byStatus[status] = append(list[:i], list[i+1:]...)
			break
		}
	}
}

// Helper function
func generateSubmissionID() string {
	return "sub_" + time.Now().Format("20060102150405") + "_" + randomString(8)
}
