package encryption

import (
	"encoding/json"
	"enterprise-core/backend/pkg/logger"
	"fmt"
)

// FieldEncryptor handles field-level encryption for sensitive data
type FieldEncryptor struct {
	encryptor      *AESEncryptor
	sensitiveFields []string
	logger         *logger.Logger
}

func NewFieldEncryptor(encryptor *AESEncryptor, sensitiveFields []string, logger *logger.Logger) *FieldEncryptor {
	return &FieldEncryptor{
		encryptor:      encryptor,
		sensitiveFields: sensitiveFields,
		logger:         logger,
	}
}

// EncryptFields encrypts sensitive fields in the data map
func (f *FieldEncryptor) EncryptFields(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for key, value := range data {
		if f.isSensitiveField(key) {
			// Encrypt the field
			encrypted, err := f.encryptValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", key, err)
			}
			result[key] = encrypted
			result[key+"_encrypted"] = true
		} else {
			result[key] = value
		}
	}

	return result, nil
}

// DecryptFields decrypts encrypted fields in the data map
func (f *FieldEncryptor) DecryptFields(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	for key, value := range data {
		if encrypted, ok := data[key+"_encrypted"].(bool); ok && encrypted {
			// Decrypt the field
			decrypted, err := f.decryptValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt field %s: %w", key, err)
			}
			result[key] = decrypted
			// Remove encryption marker
			delete(result, key+"_encrypted")
		} else {
			result[key] = value
		}
	}

	return result, nil
}

func (f *FieldEncryptor) isSensitiveField(field string) bool {
	for _, sensitive := range f.sensitiveFields {
		if field == sensitive {
			return true
		}
	}
	return false
}

func (f *FieldEncryptor) encryptValue(value interface{}) (string, error) {
	// Convert value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	// Encrypt
	encrypted, err := f.encryptor.Encrypt(data)
	if err != nil {
		return "", err
	}

	// Return as base64 string
	return fmt.Sprintf("encrypted:%x", encrypted), nil
}

func (f *FieldEncryptor) decryptValue(value interface{}) (interface{}, error) {
	// Extract encrypted data
	encryptedStr, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("encrypted value is not a string")
	}

	// Parse hex
	var encrypted []byte
	if _, err := fmt.Sscanf(encryptedStr, "encrypted:%x", &encrypted); err != nil {
		return nil, err
	}

	// Decrypt
	decrypted, err := f.encryptor.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	// Parse JSON back to original type
	var result interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Common sensitive field names
var DefaultSensitiveFields = []string{
	"password",
	"ssn",
	"credit_card",
	"api_key",
	"secret",
	"token",
	"private_key",
	"email",
	"phone",
	"address",
}
