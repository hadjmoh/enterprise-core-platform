package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KeyManager handles Master Keys and Data Keys
type KeyManager struct {
	masterKey []byte // Mock master key
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		masterKey: []byte("very-secret-master-key-32-chars-!"), // 32 bytes for AES-256
	}
}

// GenerateDataKey creates a new 32-byte data key and returns it along with its encrypted version
func (k *KeyManager) GenerateDataKey() ([]byte, []byte, error) {
	dataKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dataKey); err != nil {
		return nil, nil, err
	}

	encryptedKey, err := k.encrypt(dataKey, k.masterKey)
	if err != nil {
		return nil, nil, err
	}

	return dataKey, encryptedKey, nil
}

// DecryptDataKey removes the master key protection from a data key
func (k *KeyManager) DecryptDataKey(encryptedKey []byte) ([]byte, error) {
	return k.decrypt(encryptedKey, k.masterKey)
}

// Low-level helper for AES-GCM encryption
func (k *KeyManager) encrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Low-level helper for AES-GCM decryption
func (k *KeyManager) decrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, encryptedData, nil)
}

// Encrypter returns a GCM cipher interface for a data key
func (k *KeyManager) NewCipher(dataKey []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
