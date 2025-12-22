package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/governance"
	"net/http"
)

type GovernanceHandler struct {
	governor      *governance.Governor
	driftDetector *governance.DriftDetector
	policyEngine  *governance.PolicyEngine
}

func NewGovernanceHandler(gov *governance.Governor, drift *governance.DriftDetector, pe *governance.PolicyEngine) *GovernanceHandler {
	return &GovernanceHandler{
		governor:      gov,
		driftDetector: drift,
		policyEngine:  pe,
	}
}

func (h *GovernanceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.governor.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *GovernanceHandler) Lockdown(w http.ResponseWriter, r *http.Request) {
	h.governor.EnterEmergencyLockdown()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "lockdown_initiated"})
}

func (h *GovernanceHandler) ReleaseLockdown(w http.ResponseWriter, r *http.Request) {
	h.governor.ReleaseLockdown()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "lockdown_released"})
}

func (h *GovernanceHandler) ToggleKillSwitch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Subsystem string `json:"subsystem"`
		Active    bool   `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.governor.ToggleKillSwitch(governance.Subsystem(req.Subsystem), req.Active)
	w.WriteHeader(http.StatusOK)
}

func (h *GovernanceHandler) GetDrift(w http.ResponseWriter, r *http.Request) {
	report := h.driftDetector.Check()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
