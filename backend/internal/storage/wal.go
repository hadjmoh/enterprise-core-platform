package storage

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/encryption"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// WAL handles encrypted write-ahead logging for a partition
type WAL struct {
	path   string
	file   *os.File
	kms    *encryption.KeyManager
	cipher cipher.AEAD
}

func OpenWAL(partitionPath string, kms *encryption.KeyManager) (*WAL, error) {
	walPath := filepath.Join(partitionPath, "wal.log")
	keyPath := filepath.Join(partitionPath, "wal.key")

	var encryptedKey []byte
	var dataKey []byte
	var err error

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		// New WAL, generate new data key
		var rawKey []byte
		rawKey, encryptedKey, err = kms.GenerateDataKey()
		if err != nil {
			return nil, err
		}
		dataKey = rawKey
		if err := os.WriteFile(keyPath, encryptedKey, 0600); err != nil {
			return nil, err
		}
	} else {
		// Existing WAL, load and decrypt data key
		encryptedKey, err = os.ReadFile(keyPath)
		if err != nil {
			return nil, err
		}
		dataKey, err = kms.DecryptDataKey(encryptedKey)
		if err != nil {
			return nil, err
		}
	}

	c, err := kms.NewCipher(dataKey)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(walPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	return &WAL{path: walPath, file: f, kms: kms, cipher: c}, nil
}

func (w *WAL) Append(event buffer.Event) error {
	plaintext, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Encrypt using AES-GCM
	nonce := make([]byte, w.cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	
	ciphertext := w.cipher.Seal(nil, nonce, plaintext, nil)

	// Combine: NonceSize (1 byte) + Nonce + CiphertextSize (4 bytes) + Ciphertext
	// For simplicity in this mock, we'll store hex-encoded lines: "nonce:ciphertext\n"
	encoded := fmt.Sprintf("%x:%x\n", nonce, ciphertext)
	
	_, err = w.file.WriteString(encoded)
	if err != nil {
		return err
	}
	return w.file.Sync()
}

func (w *WAL) Replay() ([]buffer.Event, error) {
	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	var events []buffer.Event
	var line string
	for {
		_, err := fmt.Fscanln(w.file, &line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		nonce, _ := hex.DecodeString(parts[0])
		ciphertext, _ := hex.DecodeString(parts[1])

		plaintext, err := w.cipher.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, fmt.Errorf("decryption failed during replay: %w", err)
		}

		var event buffer.Event
		if err := json.Unmarshal(plaintext, &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (w *WAL) Truncate() error {
	w.file.Close()
	return os.Remove(w.path)
}

func (w *WAL) Close() error {
	return w.file.Close()
}
