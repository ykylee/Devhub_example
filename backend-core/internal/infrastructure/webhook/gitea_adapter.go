package webhook

import (
	"encoding/json"
	"strings"
)

// GiteaWebhookAdapter 는 provider_type='scm' 의 Gitea/Forgejo/Gogs webhook envelope
// parser (X-2 sprint, release_v0-1_roadmap.md §3.5 X-2).
//
// 정공법:
//   - 기존 gitea_webhook.go (legacy) 와 IngestIntegrationProviderWebhook (newer) 의
//     envelope 파싱 로직을 Gitea adapter 로 통합.
//   - Gitea/Forgejo 가 X-Gitea-* header + repository/sender/sender envelope 전송.
//     Gogs 가 X-Gogs-* header + 동일 envelope schema.
//   - ExternalRef = repository.full_name (PR 의 경우 "org/repo#123" 형식) + issue/PR
//     의 number field. AutoRouteVocRegistration 의 giteaExternalRefPattern (^GITEA-<digits>$)
//     와 cross-ref.
//
// Dispatch:
//   - IngestIntegrationProviderWebhook handler 가 provider_type='scm' 인 경우
//     GiteaWebhookAdapter.ExtractEvent 호출 (X-2 multi-provider 일반화).
//   - JIRA/Generic adapter 는 별도 follow-up.
type GiteaWebhookAdapter struct{}

// NewGiteaWebhookAdapter 는 GiteaWebhookAdapter 의 constructor.
func NewGiteaWebhookAdapter() *GiteaWebhookAdapter {
	return &GiteaWebhookAdapter{}
}

// ProviderType returns "scm" (integration-registry 의 IntegrationProviderType).
func (a *GiteaWebhookAdapter) ProviderType() string {
	return "scm"
}

// giteaEnvelope 는 Gitea/Forgejo/Gogs 의 webhook payload schema.
// 기존 gitea_webhook.go 의 giteaWebhookEnvelope 와 동일 shape (확장: pull_request / issues
// 의 number field + action field).
type giteaEnvelope struct {
	Action   string `json:"action"` // "opened" | "closed" | "synchronize" | "reopened" | etc
	Number   int64  `json:"number"` // issue/PR number
	Repository *struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"repository"`
	Sender *struct {
		Login    string `json:"login"`
		UserName string `json:"username"`
	} `json:"sender"`
}

// ExtractEvent 는 Gitea webhook payload + headers 의 normalized WebhookEvent 추출.
// invalid payload (JSON parse error) → error.
// empty event type / delivery ID → "" 로 채움 (caller 가 dedupe key / event_type 결정).
func (a *GiteaWebhookAdapter) ExtractEvent(providerKey string, payload []byte, headers map[string]string) (WebhookEvent, error) {
	var env giteaEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return WebhookEvent{}, err
	}
	eventType := firstHeader(headers, "X-Integration-Event", "X-Gitea-Event", "X-Gogs-Event")
	deliveryID := firstHeader(headers, "X-Integration-Delivery", "X-Gitea-Delivery", "X-Gogs-Delivery")
	actorLogin := ""
	if env.Sender != nil {
		actorLogin = env.Sender.Login
		if actorLogin == "" {
			actorLogin = env.Sender.UserName
		}
	}
	externalRef := buildGiteaExternalRef(env, eventType)
	return WebhookEvent{
		ProviderType: a.ProviderType(),
		ProviderKey:  providerKey,
		EventType:    eventType,
		DeliveryID:   deliveryID,
		ExternalRef:  externalRef,
		ActorLogin:   actorLogin,
		PayloadHash:  payloadHashHex(payload),
		RawPayload:   payload,
	}, nil
}

// buildGiteaExternalRef 는 Gitea event 의 external_ref 정공법.
// pull_request / issues / release / push 등의 event 별로 number field 매핑.
// AutoRoute 의 giteaExternalRefPattern (^GITEA-\d+$) 와 정합 — PR/issue 의
// external_ref = "GITEA-<number>" 형식.
func buildGiteaExternalRef(env giteaEnvelope, eventType string) string {
	// issue / pull_request / release = "GITEA-<number>"
	if env.Number > 0 {
		if strings.HasPrefix(eventType, "pull_request") || strings.HasPrefix(eventType, "issues") || strings.HasPrefix(eventType, "release") {
			return "GITEA-" + intToString(env.Number)
		}
	}
	// push / repository / 기타 = "GITEA-<repository.full_name>"
	if env.Repository != nil && env.Repository.FullName != "" {
		return "GITEA-" + env.Repository.FullName
	}
	return ""
}

// intToString 는 int64 → string 변환. strconv.FormatInt 의 가벼운 wrapper.
func intToString(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
