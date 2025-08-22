package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"filippo.io/age"
)

// EncryptWithPassphrase encrypts data using age with a passphrase
func EncryptWithPassphrase(data []byte, passphrase string) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient: %v", err)
	}

	var buf bytes.Buffer
	writer, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypt writer: %v", err)
	}

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	return buf.Bytes(), nil
}

// DecryptWithPassphrase decrypts data using age with a passphrase
func DecryptWithPassphrase(encryptedData []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %v", err)
	}

	reader, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create decrypt reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %v", err)
	}

	return data, nil
}

// EncryptWithKey encrypts data using age with a symmetric key
func EncryptWithKey(data []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes")
	}

	// For symmetric encryption, we'll use a simple approach
	// In a real implementation, you might want to use a proper symmetric cipher
	// For now, we'll use age with a passphrase derived from the key
	passphrase := base64.StdEncoding.EncodeToString(key)

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipient: %v", err)
	}

	var buf bytes.Buffer
	writer, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypt writer: %v", err)
	}

	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %v", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	return buf.Bytes(), nil
}

// DecryptWithKey decrypts data using age with a symmetric key
func DecryptWithKey(encryptedData []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes")
	}

	// For symmetric decryption, use the same approach as encryption
	passphrase := base64.StdEncoding.EncodeToString(key)

	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity: %v", err)
	}

	reader, err := age.Decrypt(bytes.NewReader(encryptedData), identity)
	if err != nil {
		return nil, fmt.Errorf("failed to create decrypt reader: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decrypted data: %v", err)
	}

	return data, nil
}

// GenerateProjectKey generates a new 32-byte project key
func GenerateProjectKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}
	return key, nil
}

// GenerateKeyID generates a unique key ID
func GenerateKeyID() (string, error) {
	// Generate 16 random bytes for the key ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", fmt.Errorf("failed to generate key ID: %v", err)
	}
	
	// Convert to base64 for a readable ID
	return base64.StdEncoding.EncodeToString(idBytes), nil
}

// GenerateDataKey generates a random 32-byte key for symmetric encryption
func GenerateDataKey() ([]byte, error) {
	return GenerateProjectKey()
}

// KeyFromBase64 decodes a base64-encoded key
func KeyFromBase64(keyB64 string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %v", err)
	}
	
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	
	return key, nil
}

// KeyToBase64 encodes a key to base64
func KeyToBase64(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}
