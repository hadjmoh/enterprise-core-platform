package compliance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
)

// PrivacyVault manages tokenization of sensitive data
type PrivacyVault struct {
	mu          sync.RWMutex
	tokenMap    map[string]string // data -> token
	reverseMap  map[string]string // token -> data
	salt        string
}

// NewPrivacyVault creates a new privacy vault
func NewPrivacyVault(salt string) *PrivacyVault {
	return &PrivacyVault{
		tokenMap:   make(map[string]string),
		reverseMap: make(map[string]string),
		salt:       salt,
	}
}

// Tokenize replaces sensitive data with a deterministic token
// Uses format-preserving-like strategy (simple hash with prefix for PoC)
func (v *PrivacyVault) Tokenize(data string) string {
	if data == "" {
		return ""
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	// Check if already tokenized
	if token, ok := v.tokenMap[data]; ok {
		return token
	}

	// Generate new token
	hash := sha256.Sum256([]byte(data + v.salt))
	token := fmt.Sprintf("detok_%s", hex.EncodeToString(hash[:8]))

	v.tokenMap[data] = token
	v.reverseMap[token] = data

	return token
}

// Detokenize retrieves the original data for a given token
// Requires high-privilege access (enforced by caller/middleware)
func (v *PrivacyVault) Detokenize(token string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	data, ok := v.reverseMap[token]
	return data, ok
}
