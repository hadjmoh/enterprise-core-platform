package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"enterprise-core/backend/internal/app"
	"enterprise-core/backend/pkg/logger"
)

type SubmissionHandler struct {
	manager *app.SubmissionManager
	logger  *logger.Logger
}

func NewSubmissionHandler(manager *app.SubmissionManager, log *logger.Logger) *SubmissionHandler {
	return &SubmissionHandler{
		manager: manager,
		logger:  log,
	}
}

// SubmitApp handles app submission
func (h *SubmissionHandler) SubmitApp(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("bundle")
	if err != nil {
		http.Error(w, "Bundle file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get submitter email
	submitter := r.FormValue("submitter")
	if submitter == "" {
		submitter = "unknown@example.com"
	}

	// Save bundle to temp location
	tempDir := filepath.Join(os.TempDir(), "submissions")
	os.MkdirAll(tempDir, 0755)
	bundlePath := filepath.Join(tempDir, header.Filename)

	out, err := os.Create(bundlePath)
	if err != nil {
		http.Error(w, "Failed to save bundle", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save bundle", http.StatusInternalServerError)
		return
	}

	// Parse manifest from bundle (simplified - in production, extract from tar.gz)
	// For now, we'll create a mock manifest
	manifest := &app.AppManifest{
		APIVersion: "prospect/v1alpha1",
		Kind:       "App",
		Metadata: app.AppMetadata{
			Name:        strings.TrimSuffix(header.Filename, ".tar.gz"),
			DisplayName: "Submitted App",
			Version:     "1.0.0",
			Description: "App submitted for review",
			Author:      submitter,
		},
		Permissions: []string{},
	}

	// Create submission
	submission, err := h.manager.SubmitApp(manifest, bundlePath, submitter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
}

// ListSubmissions returns all submissions
func (h *SubmissionHandler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	
	var submissions []*app.Submission
	if status != "" {
		submissions = h.manager.ListSubmissions(app.SubmissionStatus(status))
	} else {
		submissions = h.manager.ListSubmissions("")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submissions)
}

// GetSubmission returns a specific submission
func (h *SubmissionHandler) GetSubmission(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/submissions/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	submissionID := parts[4]

	submission, err := h.manager.GetSubmission(submissionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(submission)
}

// ApproveSubmission approves a submission
func (h *SubmissionHandler) ApproveSubmission(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/submissions/{id}/approve
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	submissionID := parts[4]

	var req struct {
		Reviewer string `json:"reviewer"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reviewer == "" {
		req.Reviewer = "admin"
	}

	if err := h.manager.ApproveSubmission(submissionID, req.Reviewer, req.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

// RejectSubmission rejects a submission
func (h *SubmissionHandler) RejectSubmission(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/submissions/{id}/reject
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 6 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	submissionID := parts[4]

	var req struct {
		Reviewer string `json:"reviewer"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reviewer == "" {
		req.Reviewer = "admin"
	}

	if req.Notes == "" {
		http.Error(w, "Rejection notes are required", http.StatusBadRequest)
		return
	}

	if err := h.manager.RejectSubmission(submissionID, req.Reviewer, req.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
}
