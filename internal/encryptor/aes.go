// Package encryptor provides helpers for symmetric encryption.
package encryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// AESEncrypter wraps AES-GCM operations.
type AESEncrypter struct{ gcm cipher.AEAD }

// NewAESEncrypter constructs an AESEncrypter from a hex-encoded key.
func NewAESEncrypter(hexKey string) (*AESEncrypter, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("decode aes key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}
	return &AESEncrypter{gcm: gcm}, nil
}

// Encrypt encrypts plaintext with the provided nonce.
func (a *AESEncrypter) Encrypt(plaintext []byte, nonce []byte) ([]byte, error) {
	if len(nonce) != a.gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size")
	}
	return a.gcm.Seal(nil, nonce, plaintext, nil), nil
}

// Decrypt decrypts ciphertext with the provided nonce.
func (a *AESEncrypter) Decrypt(
	ciphertext []byte, nonce []byte) ([]byte, error) {
	if len(nonce) != a.gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size")
	}
	return a.gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptToJSON returns a JSON object with base64 nonce and cipher for payload.
func (a *AESEncrypter) EncryptToJSON(
	payload json.RawMessage) (json.RawMessage, error) {
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	cipherText, err := a.Encrypt([]byte(payload), nonce)
	if err != nil {
		return nil, err
	}
	obj := map[string]string{
		"nonce":  base64.StdEncoding.EncodeToString(nonce),
		"cipher": base64.StdEncoding.EncodeToString(cipherText),
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// DecryptFromJSON expects base64 fields and returns decrypted bytes.
func (a *AESEncrypter) DecryptFromJSON(enc json.RawMessage) ([]byte, error) {
	type wrapper struct{ Nonce, Cipher string }
	var w wrapper
	if err := json.Unmarshal(enc, &w); err != nil {
		return nil, err
	}
	nonce, err := base64.StdEncoding.DecodeString(w.Nonce)
	if err != nil {
		return nil, err
	}
	cipherB, err := base64.StdEncoding.DecodeString(w.Cipher)
	if err != nil {
		return nil, err
	}
	return a.Decrypt(cipherB, nonce)
}
