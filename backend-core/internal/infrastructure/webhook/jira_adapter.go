package webhook

import (
	"encoding/json"
	"strings"
)

// JiraWebhookAdapter 는 provider_type='alm' 의 Jira/Atlassian webhook envelope
// parser (X-2 sprint, release_v0-1_roadmap.md §3.5 X-2).
//
// 정공법:
//   - Atlassian webhook 의 X-Atlassian-Webhook-Identifier header (delivery ID).
//   - Jira 의 issue webhook envelope (issue.key + issue.fields.summary + user.emailAddress).
//   - ExternalRef = Jira issue key (예: "DEV-456", "PROJ-123") — AutoRoute 의
//     jiraExternalRefPattern (^([A-Z][A-Z0-9_]{1,9})-([0-9]+)$) 와 1:1 정합.
//   - Sprint/project reference 는 issue.fields.project.key (Jira project key = DEV/PROJ).
//   - payload envelope = Jira Cloud webhook spec (https://developer.atlassian.com/cloud/jira/platform/webhooks/).
//
// Dispatch:
//   - IngestIntegrationProviderWebhook handler 가 provider_type='alm' 인 경우
//     JiraWebhookAdapter.ExtractEvent 호출.
//   - Gitea adapter 와 동일 dispatcher pattern (X-2 multi-provider 일반화).
type JiraWebhookAdapter struct{}

// NewJiraWebhookAdapter 는 JiraWebhookAdapter 의 constructor.
func NewJiraWebhookAdapter() *JiraWebhookAdapter {
	return &JiraWebhookAdapter{}
}

// ProviderType returns "alm" (integration-registry 의 IntegrationProviderType).
func (a *JiraWebhookAdapter) ProviderType() string {
	return "alm"
}

// jiraEnvelope 는 Atlassian Jira Cloud webhook 의 payload schema.
// Jira webhook spec 의 standard envelope (webhookEvent + issue + user).
type jiraEnvelope struct {
	WebhookEvent string `json:"webhookEvent"` // "jira:issue_created" | "jira:issue_updated" | "jira:issue_deleted" | etc.
	Issue *struct {
		Key     string `json:"key"`     // "DEV-456"
		ID      string `json:"id"`      // Jira internal numeric ID
		Fields  *struct {
			Summary string `json:"summary"`
		} `json:"fields,omitempty"`
	} `json:"issue,omitempty"`
	User *struct {
		AccountID    string `json:"accountId"`
		EmailAddress string `json:"emailAddress"`
		DisplayName  string `json:"displayName"`
	} `json:"user,omitempty"`
}

// ExtractEvent 는 Jira webhook payload + headers 의 normalized WebhookEvent 추출.
// invalid payload (JSON parse error) → error.
// event type 의 priority: payload.webhookEvent > X-Atlassian-Webhook-Identifier
// 와 같은 provider-native header > X-Integration-Event (DevHub-native).
func (a *JiraWebhookAdapter) ExtractEvent(providerKey string, payload []byte, headers map[string]string) (WebhookEvent, error) {
	var env jiraEnvelope
	if err := json.Unmarshal(payload, &env); err != nil {
		return WebhookEvent{}, err
	}
	// event type: 4-tier priority (payload.webhookEvent > provider-native header > DevHub-native header > empty)
	eventType := env.WebhookEvent
	if eventType == "" {
		eventType = firstHeader(headers, "X-Atlassian-Webhook-Identifier", "X-Integration-Event")
	}
	deliveryID := firstHeader(headers, "X-Atlassian-Webhook-Identifier", "X-Integration-Delivery")
	actorLogin := ""
	if env.User != nil {
		actorLogin = env.User.EmailAddress
		if actorLogin == "" {
			actorLogin = env.User.DisplayName
		}
	}
	externalRef := ""
	if env.Issue != nil {
		externalRef = env.Issue.Key
	}
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

// jiraEventTypeAllowed 는 Jira 의 webhook event type whitelist (silent skip for unknown).
// 본 함수 는 invalid payload (unexpected event type) 의 silent drop 검증에 사용.
func jiraEventTypeAllowed(eventType string) bool {
	switch {
	case strings.HasPrefix(eventType, "jira:issue_"),
		strings.HasPrefix(eventType, "jira:project_"),
		strings.HasPrefix(eventType, "jira:sprint_"),
		strings.HasPrefix(eventType, "jira:comment_"),
		strings.HasPrefix(eventType, "jira:worklog_"),
		eventType == "":
		return true
	}
	return false
}
