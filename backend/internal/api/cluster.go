package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/cluster"
	"net/http"
)

type ClusterHandler struct {
	scaler *cluster.AutoScaler
	monitor *cluster.ResourceMonitor
}

func NewClusterHandler(scaler *cluster.AutoScaler, monitor *cluster.ResourceMonitor) *ClusterHandler {
	return &ClusterHandler{
		scaler: scaler,
		monitor: monitor,
	}
}

func (h *ClusterHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.scaler.GetStatus()
	status["current_metrics"] = h.monitor.GetClusterMetrics()
	status["nodes"] = h.monitor.GetNodes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ClusterHandler) ManualScale(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Action string `json:"action"`
		Delta  int    `json:"delta"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.scaler.Scale(r.Context(), req.Action, req.Delta, req.Reason)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ClusterHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p cluster.ScalingPolicy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.scaler.UpdatePolicy(p)
	w.WriteHeader(http.StatusOK)
}

func (h *ClusterHandler) ToggleKillSwitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.scaler.ToggleKillSwitch(req.Enabled)
	w.WriteHeader(http.StatusOK)
}
