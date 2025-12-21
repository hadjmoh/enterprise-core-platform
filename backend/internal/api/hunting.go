package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/security"
	"net/http"
	"strings"
)

type HuntingHandler struct {
	manager *security.HuntingManager
}

func NewHuntingHandler(mgr *security.HuntingManager) *HuntingHandler {
	return &HuntingHandler{manager: mgr}
}

func (h *HuntingHandler) GetHunts(w http.ResponseWriter, r *http.Request) {
	hunts := h.manager.GetHunts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hunts)
}

func (h *HuntingHandler) SaveHunt(w http.ResponseWriter, r *http.Request) {
	var hunt security.Hunt
	if err := json.NewDecoder(r.Body).Decode(&hunt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.manager.SaveHunt(&hunt)
	w.WriteHeader(http.StatusCreated)
}

func (h *HuntingHandler) GetNotebooks(w http.ResponseWriter, r *http.Request) {
	notebooks := h.manager.GetNotebooks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notebooks)
}

func (h *HuntingHandler) SaveNotebook(w http.ResponseWriter, r *http.Request) {
	var nb security.Notebook
	if err := json.NewDecoder(r.Body).Decode(&nb); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.manager.SaveNotebook(&nb)
	w.WriteHeader(http.StatusCreated)
}

func (h *HuntingHandler) GetNotebook(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		http.Error(w, "Notebook ID required", http.StatusBadRequest)
		return
	}
	id := parts[4]

	nb, ok := h.manager.GetNotebook(id)
	if !ok {
		http.Error(w, "Notebook not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nb)
}
