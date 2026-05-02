package gitea

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	body := []byte(`{"action":"opened"}`)
	secret := "devhub-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !VerifySignature(body, secret, signature) {
		t.Fatal("expected signature to verify")
	}
	if !VerifySignature(body, secret, "sha256="+signature) {
		t.Fatal("expected prefixed signature to verify")
	}
	if VerifySignature(body, secret, "bad-signature") {
		t.Fatal("expected invalid signature to fail")
	}
}
