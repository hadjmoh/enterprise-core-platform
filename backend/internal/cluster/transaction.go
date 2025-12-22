package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"sync"
	"time"
)

// TransactionState represents the state of a distributed transaction
type TransactionState string

const (
	TxPending   TransactionState = "pending"
	TxPrepared  TransactionState = "prepared"
	TxCommitted TransactionState = "committed"
	TxAborted   TransactionState = "aborted"
)

// Transaction represents a distributed operation across the cluster
type Transaction struct {
	ID          string                  `json:"id"`
	Type        string                  `json:"type"` // "deployment", "config_update", etc.
	State       TransactionState        `json:"state"`
	Participants []string               `json:"participants"` // Node IDs
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	PreparedNodes []string             `json:"prepared_nodes"`
}

// TransactionLog tracks transaction history for audit
type TransactionLog struct {
	TxID      string           `json:"tx_id"`
	Timestamp time.Time        `json:"timestamp"`
	State     TransactionState `json:"state"`
	Message   string           `json:"message"`
}

// TransactionCoordinator implements two-phase commit protocol
type TransactionCoordinator struct {
	transactions map[string]*Transaction
	logs         []TransactionLog
	nodeMgr      *NodeManager
	logger       *logger.Logger
	mu           sync.RWMutex
}

func NewTransactionCoordinator(nodeMgr *NodeManager, l *logger.Logger) *TransactionCoordinator {
	return &TransactionCoordinator{
		transactions: make(map[string]*Transaction),
		logs:         make([]TransactionLog, 0),
		nodeMgr:      nodeMgr,
		logger:       l,
	}
}

// BeginTransaction starts a new distributed transaction
func (tc *TransactionCoordinator) BeginTransaction(txType string, participants []string, payload map[string]interface{}) (*Transaction, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	txID := fmt.Sprintf("tx-%d", time.Now().UnixNano())
	tx := &Transaction{
		ID:           txID,
		Type:         txType,
		State:        TxPending,
		Participants: participants,
		Payload:      payload,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		PreparedNodes: make([]string, 0),
	}

	tc.transactions[txID] = tx
	tc.logTransaction(txID, TxPending, "Transaction initiated")
	tc.logger.Info("Transaction started", "tx_id", txID, "type", txType, "participants", len(participants))

	return tx, nil
}

// PreparePhase executes the prepare phase of 2PC
func (tc *TransactionCoordinator) PreparePhase(txID string) error {
	tc.mu.Lock()
	tx, ok := tc.transactions[txID]
	if !ok {
		tc.mu.Unlock()
		return fmt.Errorf("transaction %s not found", txID)
	}
	tc.mu.Unlock()

	tc.logger.Info("Starting prepare phase", "tx_id", txID, "participants", len(tx.Participants))

	// In a real implementation, this would send prepare requests to all nodes
	// For now, we simulate successful preparation
	tc.mu.Lock()
	tx.PreparedNodes = tx.Participants
	tx.State = TxPrepared
	tx.UpdatedAt = time.Now()
	tc.logTransaction(txID, TxPrepared, fmt.Sprintf("All %d nodes prepared", len(tx.Participants)))
	tc.mu.Unlock()

	return nil
}

// CommitPhase executes the commit phase of 2PC
func (tc *TransactionCoordinator) CommitPhase(txID string) error {
	tc.mu.Lock()
	tx, ok := tc.transactions[txID]
	if !ok {
		tc.mu.Unlock()
		return fmt.Errorf("transaction %s not found", txID)
	}

	if tx.State != TxPrepared {
		tc.mu.Unlock()
		return fmt.Errorf("transaction %s not in prepared state", txID)
	}
	tc.mu.Unlock()

	tc.logger.Info("Starting commit phase", "tx_id", txID)

	// In a real implementation, this would send commit requests to all nodes
	tc.mu.Lock()
	tx.State = TxCommitted
	tx.UpdatedAt = time.Now()
	tc.logTransaction(txID, TxCommitted, "Transaction committed successfully")
	tc.mu.Unlock()

	return nil
}

// AbortTransaction rolls back a transaction
func (tc *TransactionCoordinator) AbortTransaction(txID string, reason string) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tx, ok := tc.transactions[txID]
	if !ok {
		return fmt.Errorf("transaction %s not found", txID)
	}

	tc.logger.Warn("Aborting transaction", "tx_id", txID, "reason", reason)

	tx.State = TxAborted
	tx.UpdatedAt = time.Now()
	tc.logTransaction(txID, TxAborted, fmt.Sprintf("Transaction aborted: %s", reason))

	return nil
}

// GetTransaction retrieves a transaction by ID
func (tc *TransactionCoordinator) GetTransaction(txID string) (*Transaction, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	tx, ok := tc.transactions[txID]
	return tx, ok
}

// GetTransactionLogs returns the audit log for a transaction
func (tc *TransactionCoordinator) GetTransactionLogs(txID string) []TransactionLog {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	logs := make([]TransactionLog, 0)
	for _, log := range tc.logs {
		if log.TxID == txID {
			logs = append(logs, log)
		}
	}
	return logs
}

func (tc *TransactionCoordinator) logTransaction(txID string, state TransactionState, message string) {
	log := TransactionLog{
		TxID:      txID,
		Timestamp: time.Now(),
		State:     state,
		Message:   message,
	}
	tc.logs = append(tc.logs, log)
}
