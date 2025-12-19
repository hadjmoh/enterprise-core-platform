package encryption

import (
	"encoding/base64"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"os"
	"sync"
)

// KMS provides key management capabilities
type KMS struct {
	keys   map[string][]byte
	logger *logger.Logger
	mu     sync.RWMutex
}

func NewKMS(logger *logger.Logger) *KMS {
	return &KMS{
		keys:   make(map[string][]byte),
		logger: logger,
	}
}

// GetKey retrieves a key by ID
func (k *KMS) GetKey(keyID string) ([]byte, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	key, ok := k.keys[keyID]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	return key, nil
}

// StoreKey stores a key with the given ID
func (k *KMS) StoreKey(keyID string, key []byte) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.keys[keyID] = key
	k.logger.Info("Key stored", "keyID", keyID)
	return nil
}

// GenerateKey creates a new key and stores it
func (k *KMS) GenerateKey(keyID string) ([]byte, error) {
	key, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	if err := k.StoreKey(keyID, key); err != nil {
		return nil, err
	}

	return key, nil
}

// RotateKey generates a new key for the given ID
func (k *KMS) RotateKey(keyID string) ([]byte, error) {
	k.logger.Info("Rotating key", "keyID", keyID)
	return k.GenerateKey(keyID)
}

// LoadFromEnv loads keys from environment variables
func (k *KMS) LoadFromEnv() error {
	// Load master key from environment
	if masterKeyB64 := os.Getenv("MASTER_ENCRYPTION_KEY"); masterKeyB64 != "" {
		key, err := base64.StdEncoding.DecodeString(masterKeyB64)
		if err != nil {
			return fmt.Errorf("failed to decode master key: %w", err)
		}
		
		if err := k.StoreKey("master", key); err != nil {
			return err
		}
		
		k.logger.Info("Master key loaded from environment")
	}

	return nil
}

// TODO: Integrate with external KMS providers
// - AWS KMS
// - HashiCorp Vault
// - Azure Key Vault
// - Google Cloud KMS
