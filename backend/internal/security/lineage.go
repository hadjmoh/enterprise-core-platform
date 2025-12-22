package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"time"
)

// LineageTracker handles the creation and verification of lineage steps
type LineageTracker struct {
	secretKey []byte
	nodeID    string
}

func NewLineageTracker(secretKey string, nodeID string) *LineageTracker {
	return &LineageTracker{
		secretKey: []byte(secretKey),
		nodeID:    nodeID,
	}
}

// CreateStep generates a signed lineage step for the current stage
func (t *LineageTracker) CreateStep(stage, action string, data map[string]interface{}) buffer.LineageStep {
	timestamp := time.Now()
	dataJSON, _ := json.Marshal(data)
	
	h := sha256.New()
	h.Write(dataJSON)
	dataHash := hex.EncodeToString(h.Sum(nil))

	step := buffer.LineageStep{
		Stage:     stage,
		NodeID:    t.nodeID,
		Timestamp: timestamp,
		Action:    action,
		DataHash:  dataHash,
	}

	step.Signature = t.signStep(step)
	return step
}

func (t *LineageTracker) signStep(step buffer.LineageStep) string {
	payload := step.Stage + step.NodeID + step.Timestamp.Format(time.RFC3339) + step.Action + step.DataHash
	mac := hmac.New(sha256.New, t.secretKey)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyStep checks if a lineage step's signature is valid
func (t *LineageTracker) VerifyStep(step buffer.LineageStep) bool {
	expectedSignature := t.signStep(step)
	return hmac.Equal([]byte(step.Signature), []byte(expectedSignature))
}
