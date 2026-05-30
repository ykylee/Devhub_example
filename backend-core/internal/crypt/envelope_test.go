package crypt

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestEnvelopeCrypter_MasterKeyValidation(t *testing.T) {
	// 1. Missing Key
	os.Setenv("DEVHUB_ENCRYPTION_KEY", "")
	defer os.Unsetenv("DEVHUB_ENCRYPTION_KEY")

	_, err := NewEnvelopeCrypter()
	if !errors.Is(err, ErrNoMasterKey) {
		t.Errorf("expected ErrNoMasterKey, got %v", err)
	}

	// 2. Invalid size (e.g. 16 bytes)
	shortKey := base64.StdEncoding.EncodeToString(make([]byte, 16))
	os.Setenv("DEVHUB_ENCRYPTION_KEY", shortKey)
	_, err = NewEnvelopeCrypter()
	if err == nil || !strings.Contains(err.Error(), "master key must be exactly 32 bytes") {
		t.Errorf("expected 32-byte size validation error, got %v", err)
	}

	// 3. Completely invalid base64 and hex
	os.Setenv("DEVHUB_ENCRYPTION_KEY", "not-a-valid-base64-or-hex-key!")
	_, err = NewEnvelopeCrypter()
	if err == nil || !strings.Contains(err.Error(), "failed base64 and hex parsing") {
		t.Errorf("expected decode error, got %v", err)
	}
}

func TestEnvelopeCrypter_EncryptDecrypt_Base64(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i)
	}
	encodedKey := base64.StdEncoding.EncodeToString(rawKey)
	os.Setenv("DEVHUB_ENCRYPTION_KEY", encodedKey)
	defer os.Unsetenv("DEVHUB_ENCRYPTION_KEY")

	crypter, err := NewEnvelopeCrypter()
	if err != nil {
		t.Fatalf("failed to create EnvelopeCrypter: %v", err)
	}

	plaintext := "my-secret-access-token"

	// Encrypt
	ciphertext1, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if !strings.HasPrefix(ciphertext1, "$env$v1$") {
		t.Errorf("ciphertext lacks expected prefix: %s", ciphertext1)
	}

	// Decrypt
	decrypted, err := crypter.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected decrypted text %q, got %q", plaintext, decrypted)
	}

	// Nonce uniqueness verification
	ciphertext2, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt 2 failed: %v", err)
	}

	if ciphertext1 == ciphertext2 {
		t.Errorf("nonces should be dynamic, but ciphertexts are identical")
	}
}

func TestEnvelopeCrypter_EncryptDecrypt_Hex(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i + 10)
	}
	encodedKey := hex.EncodeToString(rawKey)
	os.Setenv("DEVHUB_ENCRYPTION_KEY", encodedKey)
	defer os.Unsetenv("DEVHUB_ENCRYPTION_KEY")

	crypter, err := NewEnvelopeCrypter()
	if err != nil {
		t.Fatalf("failed to create EnvelopeCrypter: %v", err)
	}

	plaintext := "hex-key-secret"
	ciphertext, err := crypter.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	decrypted, err := crypter.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected decrypted %q, got %q", plaintext, decrypted)
	}
}

func TestEnvelopeCrypter_LegacyPlaintextFallback(t *testing.T) {
	rawKey := make([]byte, 32)
	encodedKey := base64.StdEncoding.EncodeToString(rawKey)
	os.Setenv("DEVHUB_ENCRYPTION_KEY", encodedKey)
	defer os.Unsetenv("DEVHUB_ENCRYPTION_KEY")

	crypter, err := NewEnvelopeCrypter()
	if err != nil {
		t.Fatalf("failed to create EnvelopeCrypter: %v", err)
	}

	// Plaintext that has no envelope prefix should bypass decryption gracefully
	legacyValue := "unencrypted_legacy_password"
	decrypted, err := crypter.Decrypt(legacyValue)
	if err != nil {
		t.Fatalf("legacy decrypt should not fail: %v", err)
	}

	if decrypted != legacyValue {
		t.Errorf("expected legacy fallback to return %q, got %q", legacyValue, decrypted)
	}
}

func TestEnvelopeCrypter_InvalidEnvelope(t *testing.T) {
	rawKey := make([]byte, 32)
	encodedKey := base64.StdEncoding.EncodeToString(rawKey)
	os.Setenv("DEVHUB_ENCRYPTION_KEY", encodedKey)
	defer os.Unsetenv("DEVHUB_ENCRYPTION_KEY")

	crypter, err := NewEnvelopeCrypter()
	if err != nil {
		t.Fatalf("failed to create EnvelopeCrypter: %v", err)
	}

	// 1. Missing segments
	_, err = crypter.Decrypt("$env$v1$only-one-segment")
	if !errors.Is(err, ErrInvalidEnvelope) {
		t.Errorf("expected ErrInvalidEnvelope for missing segments, got %v", err)
	}

	// 2. Invalid base64 in wrapped DEK segment
	_, err = crypter.Decrypt("$env$v1$notbase64!$YW55IG5vbmNl$YW55IGNpcGhlcg==")
	if err == nil || !strings.Contains(err.Error(), "decode wrapped dek") {
		t.Errorf("expected wrapped dek decode error, got %v", err)
	}
}

func TestGlobalBypassMode(t *testing.T) {
	// If globalCrypter is nil (meaning DEVHUB_ENCRYPTION_KEY was empty at startup),
	// crypt.Encrypt and crypt.Decrypt should transparently bypass.
	// Since init() runs once, we can test by asserting globalCrypter.
	// We'll call the global package level functions.
	
	originalGlobal := globalCrypter
	defer func() { globalCrypter = originalGlobal }()
	globalCrypter = nil

	input := "bypass-secret"
	
	enc, err := Encrypt(input)
	if err != nil {
		t.Errorf("global Encrypt failed in bypass: %v", err)
	}
	if enc != input {
		t.Errorf("expected global Encrypt bypass to return %q, got %q", input, enc)
	}

	dec, err := Decrypt(input)
	if err != nil {
		t.Errorf("global Decrypt failed in bypass: %v", err)
	}
	if dec != input {
		t.Errorf("expected global Decrypt bypass to return %q, got %q", input, dec)
	}
}
