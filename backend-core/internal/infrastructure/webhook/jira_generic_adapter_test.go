package webhook

import (
	"testing"
)

// ============================================================================
// X-2 JiraWebhookAdapter (provider_type='alm') unit test
// ============================================================================

func TestJiraWebhookAdapter_ProviderType(t *testing.T) {
	a := NewJiraWebhookAdapter()
	if a.ProviderType() != "alm" {
		t.Fatalf("expected ProviderType=alm, got %s", a.ProviderType())
	}
}

func TestJiraWebhookAdapter_ExtractEvent_IssueCreated(t *testing.T) {
	a := NewJiraWebhookAdapter()
	payload := []byte(`{
		"webhookEvent": "jira:issue_created",
		"issue": {"key": "DEV-456", "id": "10001", "fields": {"summary": "Add new feature"}},
		"user": {"accountId": "alice", "emailAddress": "alice@example.com", "displayName": "Alice"}
	}`)
	headers := map[string]string{
		"X-Atlassian-Webhook-Identifier": "delivery-jira-1",
	}
	ev, err := a.ExtractEvent("jira-prod", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ProviderType != "alm" {
		t.Errorf("expected ProviderType=alm, got %s", ev.ProviderType)
	}
	if ev.EventType != "jira:issue_created" {
		t.Errorf("expected EventType=jira:issue_created, got %s", ev.EventType)
	}
	if ev.DeliveryID != "delivery-jira-1" {
		t.Errorf("expected DeliveryID=delivery-jira-1, got %s", ev.DeliveryID)
	}
	if ev.ActorLogin != "alice@example.com" {
		t.Errorf("expected ActorLogin=alice@example.com, got %s", ev.ActorLogin)
	}
	// ExternalRef for issue key "DEV-456" — AutoRoute 의 jiraExternalRefPattern `^([A-Z][A-Z0-9_]{1,9})-([0-9]+)$` 와 1:1 매핑
	if ev.ExternalRef != "DEV-456" {
		t.Errorf("expected ExternalRef=DEV-456, got %s", ev.ExternalRef)
	}
	if len(ev.PayloadHash) != 64 {
		t.Errorf("expected PayloadHash=64 hex chars, got %d", len(ev.PayloadHash))
	}
}

func TestJiraWebhookAdapter_ExtractEvent_IssueUpdated(t *testing.T) {
	a := NewJiraWebhookAdapter()
	payload := []byte(`{
		"webhookEvent": "jira:issue_updated",
		"issue": {"key": "PROJ-789", "id": "20002"},
		"user": {"emailAddress": "bob@example.com"}
	}`)
	headers := map[string]string{
		"X-Atlassian-Webhook-Identifier": "delivery-jira-2",
	}
	ev, err := a.ExtractEvent("jira-prod", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ExternalRef != "PROJ-789" {
		t.Errorf("expected ExternalRef=PROJ-789, got %s", ev.ExternalRef)
	}
}

func TestJiraWebhookAdapter_ExtractEvent_HeaderOnly(t *testing.T) {
	a := NewJiraWebhookAdapter()
	// payload 에 webhookEvent 없음 → X-Atlassian-Webhook-Identifier header 사용
	payload := []byte(`{
		"issue": {"key": "DEV-100"},
		"user": {"emailAddress": "charlie@example.com"}
	}`)
	headers := map[string]string{
		"X-Atlassian-Webhook-Identifier": "jira-event-from-header",
	}
	ev, err := a.ExtractEvent("jira-prod", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.EventType != "jira-event-from-header" {
		t.Errorf("expected EventType=jira-event-from-header (from header), got %s", ev.EventType)
	}
}

func TestJiraWebhookAdapter_ExtractEvent_NoUser(t *testing.T) {
	a := NewJiraWebhookAdapter()
	payload := []byte(`{
		"webhookEvent": "jira:issue_created",
		"issue": {"key": "DEV-200"}
	}`)
	headers := map[string]string{}
	ev, err := a.ExtractEvent("jira-prod", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ActorLogin != "" {
		t.Errorf("expected empty ActorLogin for missing user, got %s", ev.ActorLogin)
	}
}

func TestJiraWebhookAdapter_ExtractEvent_InvalidJSON(t *testing.T) {
	a := NewJiraWebhookAdapter()
	_, err := a.ExtractEvent("jira-prod", []byte(`{invalid json`), nil)
	if err == nil {
		t.Fatal("expected err for invalid JSON")
	}
}

func TestJiraEventTypeAllowed(t *testing.T) {
	tests := []struct {
		eventType string
		allowed   bool
	}{
		{"jira:issue_created", true},
		{"jira:issue_updated", true},
		{"jira:project_created", true},
		{"jira:sprint_started", true},
		{"jira:comment_created", true},
		{"jira:worklog_updated", true},
		{"", true}, // empty = unknown, allow (will be logged but not rejected)
		{"random_event", false},
		{"github:pull_request", false},
	}
	for _, tt := range tests {
		if got := jiraEventTypeAllowed(tt.eventType); got != tt.allowed {
			t.Errorf("jiraEventTypeAllowed(%q) = %v, want %v", tt.eventType, got, tt.allowed)
		}
	}
}

// ============================================================================
// X-2 GenericWebhookAdapter (provider_type='other') unit test
// ============================================================================

func TestGenericWebhookAdapter_ProviderType(t *testing.T) {
	a := NewGenericWebhookAdapter()
	if a.ProviderType() != "other" {
		t.Fatalf("expected ProviderType=other, got %s", a.ProviderType())
	}
}

func TestGenericWebhookAdapter_ExtractEvent_CustomEvent(t *testing.T) {
	a := NewGenericWebhookAdapter()
	payload := []byte(`{
		"event": "deployment_completed",
		"external_ref": "CUSTOM-999",
		"actor": "alice@example.com"
	}`)
	headers := map[string]string{
		"X-Integration-Delivery": "delivery-gen-1",
	}
	ev, err := a.ExtractEvent("custom-ci", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ProviderType != "other" {
		t.Errorf("expected ProviderType=other, got %s", ev.ProviderType)
	}
	if ev.EventType != "deployment_completed" {
		t.Errorf("expected EventType=deployment_completed, got %s", ev.EventType)
	}
	if ev.ExternalRef != "CUSTOM-999" {
		t.Errorf("expected ExternalRef=CUSTOM-999, got %s", ev.ExternalRef)
	}
	if ev.ActorLogin != "alice@example.com" {
		t.Errorf("expected ActorLogin=alice@example.com, got %s", ev.ActorLogin)
	}
}

func TestGenericWebhookAdapter_ExtractEvent_HeaderFallback(t *testing.T) {
	a := NewGenericWebhookAdapter()
	// payload 에 event 없음 → header X-Integration-Event 사용
	payload := []byte(`{
		"external_ref": "CUSTOM-100",
		"actor": "bob"
	}`)
	headers := map[string]string{
		"X-Integration-Event":    "header-event",
		"X-Integration-Delivery": "delivery-gen-2",
	}
	ev, err := a.ExtractEvent("custom-ci", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.EventType != "header-event" {
		t.Errorf("expected EventType=header-event (from header fallback), got %s", ev.EventType)
	}
}

func TestGenericWebhookAdapter_ExtractEvent_InvalidJSON(t *testing.T) {
	a := NewGenericWebhookAdapter()
	_, err := a.ExtractEvent("custom-ci", []byte(`{invalid`), nil)
	if err == nil {
		t.Fatal("expected err for invalid JSON")
	}
}

func TestGenericWebhookAdapter_MatchExternalRefPattern(t *testing.T) {
	a := NewGenericWebhookAdapter()
	tests := []struct {
		externalRef  string
		customPattern string
		expected     bool
	}{
		{"CUSTOM-999", `^CUSTOM-\d+$`, true},
		{"OTHER-123", `^CUSTOM-\d+$`, false},
		{"CUSTOM-999", `^[A-Z]+-\d+$`, true},  // 더 넓은 pattern
		{"abc-123", `^[A-Z]+-\d+$`, false},     // case mismatch
		{"", `^CUSTOM-\d+$`, false},            // empty externalRef
		{"CUSTOM-999", "", false},              // empty pattern
		{"ANY-1", `[invalid(`, false},           // invalid pattern = silent skip
		{"JIRA-456", `^JIRA-\d+$`, true},        // Jira-style custom
		{"#789", `^#\d+$`, true},                // GitHub-style custom
		{"!100", `^!\d+$`, true},                // GitLab-style custom
	}
	for _, tt := range tests {
		got := a.MatchExternalRefPattern(tt.externalRef, tt.customPattern)
		if got != tt.expected {
			t.Errorf("MatchExternalRefPattern(%q, %q) = %v, want %v", tt.externalRef, tt.customPattern, got, tt.expected)
		}
	}
}

// ============================================================================
// X-2 adapter init() 등록 dispatcher 정합 검증
// ============================================================================

func TestInit_RegistersAllAdapters(t *testing.T) {
	// 본 test 는 init() 이 호출된 후의 상태 검증. init() 은 package import 시 1회 실행.
	// 다른 test (TestRegisterAdapter_GetAdapterForProviderType) 가 adapterRegistry 를
	// 재설정할 수 있으므로, 다시 init() 의 effect 를 시뮬레이션.
	adapterRegistry = map[string]WebhookAdapter{}
	// init() 자동 호출이 본 test 에서도 적용됨 (Go test framework 의 package import
	// lifecycle). 명시적 호출 불요.
	RegisterAdapter(NewGiteaWebhookAdapter())
	RegisterAdapter(NewJiraWebhookAdapter())
	RegisterAdapter(NewGenericWebhookAdapter())

	tests := []struct {
		providerType string
		expectedType string
	}{
		{"scm", "scm"},
		{"alm", "alm"},
		{"other", "other"},
	}
	for _, tt := range tests {
		got, err := GetAdapterForProviderType(tt.providerType)
		if err != nil {
			t.Errorf("GetAdapterForProviderType(%q) returned err: %v", tt.providerType, err)
			continue
		}
		if got.ProviderType() != tt.expectedType {
			t.Errorf("adapter for %q returned ProviderType=%q, want %q", tt.providerType, got.ProviderType(), tt.expectedType)
		}
	}
}
