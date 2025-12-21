package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/security"
	"net/http"
)

type MitreHandler struct {
	manager *security.MitreManager
}

func NewMitreHandler(mgr *security.MitreManager) *MitreHandler {
	return &MitreHandler{manager: mgr}
}

func (h *MitreHandler) GetCoverage(w http.ResponseWriter, r *http.Request) {
	coverage := h.manager.GetCoverage()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coverage)
}
