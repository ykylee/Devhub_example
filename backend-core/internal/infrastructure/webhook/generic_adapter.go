package webhook

import (
	"encoding/json"
	"regexp"
)

// GenericWebhookAdapter 는 provider_type='other' 의 custom adapter (X-2 sprint,
// release_v0-1_roadmap.md §3.5 X-2).
//
// 정공법:
//   - 'other' provider 는 사용자 정의 webhook payload schema (Gitea/Jira 가 아닌 임의
//     시스템 — GitHub/GitLab/Bitbucket/Custom CI 등).
//   - custom event type 의 external_ref pattern (InboundSourceRoutingConfig.CustomExternalRefPattern)
//     가 AutoRoute 의 pattern matcher 와 cross-ref.
//   - 본 adapter 는 단순 envelope — issue/PR/merge_request 등의 표준 field 가 아닌
//     {event, external_ref, actor} 의 3-field minimal schema 가정.
//   - Gitea/Forgejo/Gogs/GitHub 의 #<n> / GitLab 의 !<n> / custom 시스템 의 ticket-XXX 등
//     다양한 external_ref 정공법 지원 — AutoRoute 의 InboundSourceRoutingConfig 의
//     custom_external_ref_pattern 으로 매칭.
//
// 정공법 예시 payload (사용자 정의):
//   {
//     "event": "deployment_completed",
//     "external_ref": "CUSTOM-999",
//     "actor": "alice@example.com",
//     "metadata": {"service": "auth", "version": "1.2.3"}
//   }
type GenericWebhookAdapter struct{}

// NewGenericWebhookAdapter 는 GenericWebhookAdapter 의 constructor.
func NewGenericWebhookAdapter() *GenericWebhookAdapter {
	return &GenericWebhookAdapter{}
}

// ProviderType returns "other" (integration-registry 의 IntegrationProviderType).
func (a *GenericWebhookAdapter) ProviderType() string {
	return "other"
}

// genericEnvelope 는 'other' provider 의 minimal envelope. 3 field 정공법:
//   - event: string (사용자 정의 event type)
//   - external_ref: string (provider-native external reference, e.g., "CUSTOM-999")
//   - actor: string (webhook 발신자 login or email)
type genericEnvelope struct {
	Event       string `json:"event"`
	ExternalRef string `json:"external_ref"`
	Actor       string `json:"actor"`
}

// ExtractEvent 는 generic webhook payload + headers 의 normalized WebhookEvent 추출.
// invalid payload (JSON parse error) → error.
// event type 의 priority: payload.event > X-Integration-Event > empty.
func (a *GenericWebhookAdapter) ExtractEvent(providerKey string, payload []byte, headers map[string]string) (WebhookEvent, error) {
	var env genericEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return WebhookEvent{}, err
	}
	eventType := env.Event
	if eventType == "" {
		eventType = firstHeader(headers, "X-Integration-Event", "X-Webhook-Event", "X-Event-Type")
	}
	deliveryID := firstHeader(headers, "X-Integration-Delivery", "X-Webhook-Delivery", "X-Delivery-ID")
	return WebhookEvent{
		ProviderType: a.ProviderType(),
		ProviderKey:  providerKey,
		EventType:    eventType,
		DeliveryID:   deliveryID,
		ExternalRef:  env.ExternalRef,
		ActorLogin:   env.Actor,
		PayloadHash:  payloadHashHex(payload),
		RawPayload:   payload,
	}, nil
}

// MatchExternalRefPattern 는 (X-2) 의 generic provider 의 custom external_ref pattern matcher.
// InboundSourceRoutingConfig.CustomExternalRefPattern 의 regex 와 external_ref 매칭.
// 1차 PR #586 의 auto_route.go 의 gitea/jira/github/gitlab provider-specific pattern 의
// 일반화 — 'other' provider 의 custom regex.
func (a *GenericWebhookAdapter) MatchExternalRefPattern(externalRef, customPattern string) bool {
	if externalRef == "" || customPattern == "" {
		return false
	}
	re, err := regexp.Compile(customPattern)
	if err != nil {
		return false // invalid pattern = silent skip (1차 PR #586 의 InvalidCustomPattern 케이스 정합)
	}
	return re.MatchString(externalRef)
}
