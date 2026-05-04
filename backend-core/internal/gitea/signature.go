package gitea

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func VerifySignature(body []byte, secret, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}

	signature = strings.TrimSpace(signature)
	signature = strings.TrimPrefix(signature, "sha256=")
	received, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)
	return hmac.Equal(received, expected)
}

func PayloadHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
