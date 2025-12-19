package security

import (
	"bytes"
	"fmt"
	"io"

	"golang.org/x/crypto/openpgp"
)

// PGPDecryptor handles PGP/GPG encrypted data
type PGPDecryptor struct {
	keyring openpgp.EntityList
}

func NewPGPDecryptor(privateKeyData []byte) (*PGPDecryptor, error) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(privateKeyData))
	if err != nil {
		return nil, fmt.Errorf("failed to read PGP key: %w", err)
	}

	return &PGPDecryptor{keyring: keyring}, nil
}

func (p *PGPDecryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	md, err := openpgp.ReadMessage(bytes.NewReader(encryptedData), p.keyring, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read PGP message: %w", err)
	}

	decrypted, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return decrypted, nil
}

// IsEncrypted checks if data appears to be PGP encrypted
func IsEncrypted(data []byte) bool {
	return bytes.Contains(data, []byte("-----BEGIN PGP MESSAGE-----"))
}
