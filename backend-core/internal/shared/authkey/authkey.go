// Package authkey 는 api_keys 의 cryptographic helper 들을 둔다.
//
// service.APIKey (생성 + repo write) 와 view.AuthenticateActor (read + verify)
// 양쪽에서 동일 helper 를 사용해야 — hash 알고리즘 drift 가 발생하면 모든
// 키가 다음 rotation 때 silent invalid. shared layer 에 두어 양쪽이 단일
// source-of-truth 를 참조.
package authkey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
)

// APIKeyPrefix 는 모든 API key 의 visible prefix. `dhk_` (DevHub Key) —
// 운영자가 key format 을 즉시 식별 가능. raw key material 자체는 노출 X.
const APIKeyPrefix = "dhk_"

// APIKeyRandomLength 는 random material 의 byte 수. 32 byte (256 bit) —
// HS256 JWT signing key 와 동일 entropy. base64-URL 인코딩 (no padding) 으로
// 43 자.
const APIKeyRandomLength = 32

// APIKeyDisplayPrefixLength 는 raw key 의 leading chars (예: `dhk_aB3x`).
// 8 자 = prefix 4 + random 4 — list view 에서 두 key disambiguate 가능, but
// prefix 누출은 raw key 복원에 사용 불가.
const APIKeyDisplayPrefixLength = 8

// HashAPIKey 는 raw key 의 sha256 hash (32 byte) 를 반환. view.auth.go 의
// auth middleware hot path 와 service.GenerateAPIKey 가 동일 알고리즘
// 사용 — drift 시 silent invalid.
func HashAPIKey(rawKey string) []byte {
	sum := sha256.Sum256([]byte(rawKey))
	return sum[:]
}

// GenerateRawAPIKey 는 crypto/rand 32 byte 기반의 raw API key + sha256 hash +
// display prefix (앞 8자) 를 한꺼번에 반환. service.GenerateAPIKey 가 thin
// re-export 하지만, view layer 가 service 의존 회피 시 직접 호출 가능.
//
// rawKey 예: `dhk_3J7x...` (총 47 자: prefix 4 + base64url(32) 43).
// keyHash 예: 32 byte sha256 binary.
// keyPrefix 예: `dhk_3J7x` (앞 8 자).
func GenerateRawAPIKey() (rawKey string, keyHash []byte, keyPrefix string, err error) {
	randomBytes := make([]byte, APIKeyRandomLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, "", fmt.Errorf("failed to read crypto/rand: %w", err)
	}
	randomEncoded := base64.RawURLEncoding.EncodeToString(randomBytes)
	rawKey = APIKeyPrefix + randomEncoded
	keyHash = HashAPIKey(rawKey)
	keyPrefix = rawKey[:APIKeyDisplayPrefixLength]
	return rawKey, keyHash, keyPrefix, nil
}

// IsCIDRAllowed 는 client IP 가 CIDR allowlist 에 match 하는지 확인. nil
// allowlist = all IPs 허용. 잘못된 CIDR / IP 는 error return.
func IsCIDRAllowed(clientIP string, allowedCIDRs []string) (bool, error) {
	if len(allowedCIDRs) == 0 {
		return true, nil
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false, fmt.Errorf("invalid client IP %q", clientIP)
	}
	for _, cidr := range allowedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return false, fmt.Errorf("invalid CIDR %q in allowlist: %w", cidr, err)
		}
		if ipNet.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}
