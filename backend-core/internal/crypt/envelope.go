package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrInvalidEnvelope = errors.New("invalid envelope format")
	ErrNoMasterKey     = errors.New("DEVHUB_ENCRYPTION_KEY environment variable is missing")
)

var (
	globalCrypter *EnvelopeCrypter
	cryptErr      error
)

func init() {
	// KEK 가 주입되어 있을 때만 기동. 비어 있으면 bypass plaintext fallback
	keyStr := os.Getenv("DEVHUB_ENCRYPTION_KEY")
	if keyStr == "" {
		return
	}
	globalCrypter, cryptErr = NewEnvelopeCrypter()
	if cryptErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Error: DEVHUB_ENCRYPTION_KEY invalid: %v\n", cryptErr)
		panic(cryptErr)
	}
}

// Encrypt plaintext. Falls back to original if no master key is present.
func Encrypt(plaintext string) (string, error) {
	if globalCrypter == nil {
		return plaintext, nil // Bypass mode
	}
	return globalCrypter.Encrypt(plaintext)
}

// Decrypt ciphertext. Falls back to original if no master key is present or not an envelope.
func Decrypt(value string) (string, error) {
	if globalCrypter == nil {
		return value, nil // Bypass mode
	}
	return globalCrypter.Decrypt(value)
}

type EnvelopeCrypter struct {
	kek []byte // Master Key (32 bytes for AES-256)
}

func NewEnvelopeCrypter() (*EnvelopeCrypter, error) {
	keyStr := os.Getenv("DEVHUB_ENCRYPTION_KEY")
	if keyStr == "" {
		return nil, ErrNoMasterKey
	}
	
	// Try standard base64 decoding first
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil || len(key) != 32 {
		// Fallback to hex decoding
		hexKey, hexErr := hex.DecodeString(keyStr)
		if hexErr == nil && len(hexKey) == 32 {
			key = hexKey
		} else {
			if err != nil {
				return nil, fmt.Errorf("decode master key: failed base64 and hex parsing")
			}
			return nil, fmt.Errorf("master key must be exactly 32 bytes, got %d", len(key))
		}
	}
	
	return &EnvelopeCrypter{kek: key}, nil
}

// Encrypt plaintext into $env$v1$<wrapped_dek_b64>$<nonce_b64>$<ciphertext_b64>
func (ec *EnvelopeCrypter) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 1. Generate random 32-byte DEK
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return "", fmt.Errorf("generate dek: %w", err)
	}

	// 2. Encrypt plaintext with DEK using AES-GCM
	ciphertext, nonce, err := aesGCMEncrypt(dek, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("encrypt plaintext: %w", err)
	}

	// 3. Encrypt DEK with KEK (master key) using AES-GCM
	wrappedDek, dekNonce, err := aesGCMEncrypt(ec.kek, dek)
	if err != nil {
		return "", fmt.Errorf("wrap dek: %w", err)
	}

	// 4. Compose envelope string
	wrappedDekB64 := base64.StdEncoding.EncodeToString(append(dekNonce, wrappedDek...))
	nonceB64 := base64.StdEncoding.EncodeToString(nonce)
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertext)

	return fmt.Sprintf("$env$v1$%s$%s$%s", wrappedDekB64, nonceB64, ciphertextB64), nil
}

// Decrypt envelope back to plaintext. Falls back to original value if not an envelope.
func (ec *EnvelopeCrypter) Decrypt(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, "$env$v1$") {
		return value, nil // Legacy Plaintext fallback
	}

	parts := strings.Split(value, "$")
	if len(parts) != 6 { // "", "env", "v1", "wrappedDekB64", "nonceB64", "ciphertextB64" (splitted by $)
		return "", ErrInvalidEnvelope
	}

	wrappedDekData, err := base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return "", fmt.Errorf("decode wrapped dek: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}

	// Extract wrapped DEK nonce (12 bytes) + encrypted DEK
	if len(wrappedDekData) < 12 {
		return "", ErrInvalidEnvelope
	}
	dekNonce := wrappedDekData[:12]
	encryptedDek := wrappedDekData[12:]

	// 1. Decrypt DEK with KEK
	dek, err := aesGCMDecrypt(ec.kek, encryptedDek, dekNonce)
	if err != nil {
		return "", fmt.Errorf("unwrap dek: %w", err)
	}

	// 2. Decrypt ciphertext with DEK
	plaintext, err := aesGCMDecrypt(dek, ciphertext, nonce)
	if err != nil {
		return "", fmt.Errorf("decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}

// Internal AES-GCM Helpers
func aesGCMEncrypt(key, plaintext []byte) (ciphertext []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func aesGCMDecrypt(key, ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, nonce, ciphertext, nil)
}
