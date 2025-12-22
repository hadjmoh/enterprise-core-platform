package api

import (
	"encoding/json"
	"enterprise-core/backend/internal/cluster"
	"net/http"
)

type ClusterHandler struct {
	scaler      *cluster.AutoScaler
	monitor     *cluster.ResourceMonitor
	nodeMgr     *cluster.NodeManager
	shardMgr    *cluster.ShardManager
	shc         *cluster.SHCManager
	deploySrv   *cluster.DeploymentServer
	rollbackMgr *cluster.RollbackManager
	drCoord     *cluster.DRCoordinator
}

func NewClusterHandler(scaler *cluster.AutoScaler, monitor *cluster.ResourceMonitor, nodeMgr *cluster.NodeManager, shardMgr *cluster.ShardManager, shc *cluster.SHCManager, deploySrv *cluster.DeploymentServer, rollbackMgr *cluster.RollbackManager, drCoord *cluster.DRCoordinator) *ClusterHandler {
	return &ClusterHandler{
		scaler:      scaler,
		monitor:     monitor,
		nodeMgr:     nodeMgr,
		shardMgr:    shardMgr,
		shc:         shc,
		deploySrv:   deploySrv,
		rollbackMgr: rollbackMgr,
		drCoord:     drCoord,
	}
}

func (h *ClusterHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.scaler.GetStatus()
	status["current_metrics"] = h.monitor.GetClusterMetrics()
	status["nodes"] = h.nodeMgr.GetNodes()
	status["shards"] = h.shardMgr.GetShards()

	if h.shc != nil {
		status["shc_state"] = map[string]interface{}{
			"search_jobs": h.shc.GetStore().GetAll("search_job"),
			"configs":     h.shc.GetStore().GetAll("config"),
		}
	}

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

func (h *ClusterHandler) DeployConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Version string            `json:"version"`
		Content map[string]string `json:"content"`
		Author  string            `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pkg, err := h.deploySrv.CreateConfig(req.Version, req.Content, req.Author)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.deploySrv.DeployConfig(pkg.Version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"version": pkg.Version,
		"checksum": pkg.Checksum,
	})
}

func (h *ClusterHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	versions := h.deploySrv.GetStore().ListVersions()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"versions": versions,
	})
}

func (h *ClusterHandler) RollbackConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.rollbackMgr.RollbackToVersion(req.Version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"version": req.Version,
	})
}

func (h *ClusterHandler) GetDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	history := h.deploySrv.GetDeploymentHistory()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}

func (h *ClusterHandler) GetDeploymentStatus(w http.ResponseWriter, r *http.Request) {
	status := h.rollbackMgr.GetRollbackStatus()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ClusterHandler) GetDRStatus(w http.ResponseWriter, r *http.Request) {
	status := h.drCoord.GetDRStatus()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *ClusterHandler) TriggerFailover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TargetSite string `json:"target_site"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.drCoord.TriggerFailover(req.TargetSite); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"target_site": req.TargetSite,
	})
}
