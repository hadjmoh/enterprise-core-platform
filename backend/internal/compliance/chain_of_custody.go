package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"enterprise-core/backend/pkg/logger"
	"fmt"
)

// ChainOfCustody tracks event lineage and integrity
type ChainOfCustody struct {
	logger *logger.Logger
}

func NewChainOfCustody(logger *logger.Logger) *ChainOfCustody {
	return &ChainOfCustody{logger: logger}
}

func (c *ChainOfCustody) Sign(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *ChainOfCustody) Verify(data []byte, signature string) bool {
	return c.Sign(data) == signature
}

func (c *ChainOfCustody) AddLineage(metadata map[string]string, stage, nodeID string) map[string]string {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[fmt.Sprintf("lineage_%s", stage)] = nodeID
	return metadata
}
