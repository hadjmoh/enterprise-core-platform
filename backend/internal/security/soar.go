package security

import (
	"context"
	"enterprise-core/backend/internal/audit"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// Action defines the interface for a response capability
type Action interface {
	Name() string
	Execute(ctx context.Context, params map[string]interface{}) error
}

// PendingAction represents an automated response awaiting approval
type PendingAction struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
	StagedAt  time.Time              `json:"staged_at"`
	Severity  string                 `json:"severity"`
	EntityID  string                 `json:"entity_id"`
	Status    string                 `json:"status"` // pending, approved, denied, executed
}

// Orchestrator manages the registration and execution of security playbooks
type Orchestrator struct {
	actions      map[string]Action
	pending      map[string]*PendingAction
	history      []*PendingAction
	auditor      *audit.AuditLogger
	logger       *logger.Logger
	mu           sync.RWMutex
}

func NewOrchestrator(auditor *audit.AuditLogger, logg *logger.Logger) *Orchestrator {
	return &Orchestrator{
		actions: make(map[string]Action),
		pending: make(map[string]*PendingAction),
		history: make([]*PendingAction, 0),
		auditor: auditor,
		logger:  logg,
	}
}

func (o *Orchestrator) RegisterAction(action Action) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.actions[action.Name()] = action
	o.logger.Info("Registered SOAR action", "name", action.Name())
}

func (o *Orchestrator) StageAction(actionName string, entityID string, severity string, params map[string]interface{}) string {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := fmt.Sprintf("ACT-%d", time.Now().UnixNano())
	pa := &PendingAction{
		ID:       id,
		Action:   actionName,
		Params:   params,
		StagedAt: time.Now(),
		Severity: severity,
		EntityID: entityID,
		Status:   "pending",
	}

	o.pending[id] = pa
	
	o.logger.Warn("SOAR action staged for approval", "id", id, "action", actionName, "entity", entityID)
	
	// Audit the staging
	o.auditor.Log("system", "SOAR_ACTION_STAGED", id, "success", "", map[string]interface{}{
		"action":    actionName,
		"entity_id": entityID,
		"severity":  severity,
	})

	return id
}

func (o *Orchestrator) ApproveAction(ctx context.Context, id string) error {
	o.mu.Lock()
	pa, ok := o.pending[id]
	if !ok {
		o.mu.Unlock()
		return fmt.Errorf("action %s not found", id)
	}
	delete(o.pending, id)
	o.mu.Unlock()

	o.mu.RLock()
	action, ok := o.actions[pa.Action]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("handler for action %s not found", pa.Action)
	}

	pa.Status = "executing"
	err := action.Execute(ctx, pa.Params)
	
	pa.Status = "executed"
	if err != nil {
		pa.Status = "failed"
	}

	o.mu.Lock()
	o.history = append(o.history, pa)
	o.mu.Unlock()

	// Audit the execution
	status := "success"
	if err != nil {
		status = "failed"
	}
	o.auditor.Log("system", "SOAR_ACTION_EXECUTED", id, status, "", map[string]interface{}{
		"action":    pa.Action,
		"entity_id": pa.EntityID,
		"error":     err,
	})

	return err
}

func (o *Orchestrator) GetPending() []*PendingAction {
	o.mu.RLock()
	defer o.mu.RUnlock()
	results := make([]*PendingAction, 0, len(o.pending))
	for _, p := range o.pending {
		results = append(results, p)
	}
	return results
}

func (o *Orchestrator) GetHistory() []*PendingAction {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.history
}

// Mock Actions

type BlockIPAction struct {
	Logger *logger.Logger
}

func (a *BlockIPAction) Name() string { return "block_ip" }
func (a *BlockIPAction) Execute(ctx context.Context, params map[string]interface{}) error {
	ip := params["ip"].(string)
	a.Logger.Warn("SOAR EXECUTE: Blocking IP on perimeter firewall", "ip", ip)
	return nil
}

type DisableUserAction struct {
	Logger *logger.Logger
}

func (a *DisableUserAction) Name() string { return "disable_user" }
func (a *DisableUserAction) Execute(ctx context.Context, params map[string]interface{}) error {
	user := params["user"].(string)
	a.Logger.Warn("SOAR EXECUTE: Disabling user in Active Directory", "user", user)
	return nil
}
