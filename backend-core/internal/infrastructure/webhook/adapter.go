// Package webhook 는 integration-registry 의 provider webhook ingest 의
// multi-provider adapter 패턴 (X-2 sprint, release_v0-1_roadmap.md §3.5 X-2).
//
// 정공법:
//   - WebhookAdapter interface: provider_type 별 webhook envelope parser.
//   - GiteaWebhookAdapter: provider_type='scm' 의 default adapter (Gitea/Forgejo/Gogs
//     의 X-Gitea-* header + repository/sender envelope 파싱).
//   - JiraWebhookAdapter (별도 follow-up): provider_type='alm' 의 Jira webhook
//     (X-Atlassian-Webhook-Identifier + issue envelope).
//   - GenericWebhookAdapter (별도 follow-up): provider_type='other' 의 custom
//     adapter (사용자 정의 event type + external_ref pattern).
//
// 본 패키지의 adapter 는 IngestIntegrationProviderWebhook handler (integration-registry
// view) 가 provider_type 별로 dispatch. multi-provider webhook endpoint 일반화의
// foundation.
package webhook

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// WebhookEvent 는 adapter 가 추출한 normalized webhook event.
// provider-agnostic 한 공통 struct — frontend / audit / event store 가 본 struct 만
// 의존하면 provider 무관하게 동작.
type WebhookEvent struct {
	ProviderType string // "scm" | "alm" | "ci_cd" | "doc" | "infra" | "task_tracker" | "other"
	ProviderKey  string // e.g., "gitea-main", "jira-prod"
	EventType    string // provider-native event type (e.g., "pull_request", "jira:issue_created")
	DeliveryID   string // provider-native delivery ID (dedupe key 의 SSOT)
	ExternalRef  string // provider-native external_ref (예: "GITEA-123", "DEV-456", "#789")
	ActorLogin   string // webhook 발신자 login
	PayloadHash  string // sha256(payload) hex
	RawPayload   []byte // 원본 payload (audit / replay)
}

// ErrAdapterNotFound 는 dispatcher 가 provider_type 에 매칭되는 adapter 를 못 찾았을 때.
var ErrAdapterNotFound = errors.New("webhook adapter not found for provider_type")

// WebhookAdapter 는 provider_type 별 webhook envelope parser.
// 본 interface 의 구현은 GiteaWebhookAdapter (default scm) + Jira/Generic
// (별도 follow-up) 가 위치. IngestIntegrationProviderWebhook handler 가
// provider.ProviderType 별로 dispatch.
type WebhookAdapter interface {
	// ProviderType 은 adapter 가 handling 하는 provider_type.
	// 예: "scm" (GiteaWebhookAdapter) | "alm" (JiraWebhookAdapter) | "other" (GenericWebhookAdapter).
	ProviderType() string

	// ExtractEvent 는 provider-native payload + headers 에서 normalized WebhookEvent 추출.
	// invalid payload → error 반환. headers 는 X-*-Event / X-*-Delivery 등 provider-native header.
	ExtractEvent(providerKey string, payload []byte, headers map[string]string) (WebhookEvent, error)
}

// GetAdapterForProviderType 는 provider_type 별로 등록된 adapter 를 반환.
// 본 함수 는 dispatcher 의 SSOT — IngestIntegrationProviderWebhook 가 호출.
// 미등록 provider_type → ErrAdapterNotFound.
var adapterRegistry = map[string]WebhookAdapter{}

// RegisterAdapter 는 adapter 를 registry 에 등록. main.go 또는 init() 에서 호출.
func RegisterAdapter(a WebhookAdapter) {
	adapterRegistry[a.ProviderType()] = a
}

// GetAdapterForProviderType 는 provider_type 의 adapter 반환. 미등록 → ErrAdapterNotFound.
func GetAdapterForProviderType(providerType string) (WebhookAdapter, error) {
	a, ok := adapterRegistry[providerType]
	if !ok {
		return nil, ErrAdapterNotFound
	}
	return a, nil
}

// payloadHashHex 는 payload 의 sha256 hex (32 byte → 64 hex char).
// WebhookEvent.PayloadHash field 채움용.
func payloadHashHex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// firstHeader 는 headers map 에서 첫 번째 non-empty 값 반환.
// nil map 안전. event type / delivery ID 의 alias 지원 (X-Gitea-Event / X-Atlassian-Event).
func firstHeader(headers map[string]string, names ...string) string {
	for _, name := range names {
		if v, ok := headers[name]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
