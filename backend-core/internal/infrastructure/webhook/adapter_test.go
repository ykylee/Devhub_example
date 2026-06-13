package webhook

import (
	"testing"
)

func TestGiteaWebhookAdapter_ProviderType(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	if a.ProviderType() != "scm" {
		t.Fatalf("expected ProviderType=scm, got %s", a.ProviderType())
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_PullRequest(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{
		"action": "opened",
		"number": 123,
		"repository": {"id": 1, "full_name": "devhub/core", "name": "core"},
		"sender": {"login": "alice", "username": "alice"}
	}`)
	headers := map[string]string{
		"X-Gitea-Event":    "pull_request",
		"X-Gitea-Delivery": "delivery-1",
	}
	ev, err := a.ExtractEvent("gitea-main", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ProviderType != "scm" {
		t.Errorf("expected ProviderType=scm, got %s", ev.ProviderType)
	}
	if ev.ProviderKey != "gitea-main" {
		t.Errorf("expected ProviderKey=gitea-main, got %s", ev.ProviderKey)
	}
	if ev.EventType != "pull_request" {
		t.Errorf("expected EventType=pull_request, got %s", ev.EventType)
	}
	if ev.DeliveryID != "delivery-1" {
		t.Errorf("expected DeliveryID=delivery-1, got %s", ev.DeliveryID)
	}
	if ev.ActorLogin != "alice" {
		t.Errorf("expected ActorLogin=alice, got %s", ev.ActorLogin)
	}
	// ExternalRef for pull_request with number=123 → "GITEA-123" (AutoRoute giteaExternalRefPattern 정합).
	if ev.ExternalRef != "GITEA-123" {
		t.Errorf("expected ExternalRef=GITEA-123, got %s", ev.ExternalRef)
	}
	// PayloadHash = sha256 hex.
	if len(ev.PayloadHash) != 64 {
		t.Errorf("expected PayloadHash=64 hex chars, got %d chars", len(ev.PayloadHash))
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_Issue(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{
		"action": "opened",
		"number": 456,
		"repository": {"id": 2, "full_name": "devhub/api", "name": "api"},
		"sender": {"login": "bob"}
	}`)
	headers := map[string]string{
		"X-Gitea-Event":    "issues",
		"X-Gitea-Delivery": "delivery-2",
	}
	ev, err := a.ExtractEvent("gitea-main", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ExternalRef != "GITEA-456" {
		t.Errorf("expected ExternalRef=GITEA-456, got %s", ev.ExternalRef)
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_Push(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{
		"ref": "refs/heads/main",
		"repository": {"id": 3, "full_name": "devhub/cli", "name": "cli"},
		"sender": {"login": "charlie"}
	}`)
	headers := map[string]string{
		"X-Gitea-Event":    "push",
		"X-Gitea-Delivery": "delivery-3",
	}
	ev, err := a.ExtractEvent("gitea-main", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// push event 는 number=0 → repository.full_name fallback → "GITEA-devhub/cli"
	if ev.ExternalRef != "GITEA-devhub/cli" {
		t.Errorf("expected ExternalRef=GITEA-devhub/cli, got %s", ev.ExternalRef)
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_AliasHeaders(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{
		"action": "opened",
		"number": 789,
		"repository": {"id": 4, "full_name": "devhub/test", "name": "test"},
		"sender": {"login": "dave"}
	}`)
	// Gogs 의 X-Gogs-Event header 사용 (X-Gitea-* 가 없음).
	headers := map[string]string{
		"X-Gogs-Event":    "issues",
		"X-Gogs-Delivery": "delivery-4",
	}
	ev, err := a.ExtractEvent("gogs-test", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.EventType != "issues" {
		t.Errorf("expected EventType=issues (from X-Gogs-Event alias), got %s", ev.EventType)
	}
	if ev.ExternalRef != "GITEA-789" {
		t.Errorf("expected ExternalRef=GITEA-789, got %s", ev.ExternalRef)
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_InvalidJSON(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{invalid json`)
	headers := map[string]string{"X-Gitea-Event": "push"}
	_, err := a.ExtractEvent("gitea-main", payload, headers)
	if err == nil {
		t.Fatal("expected err for invalid JSON")
	}
}

func TestGiteaWebhookAdapter_ExtractEvent_NoSender(t *testing.T) {
	a := NewGiteaWebhookAdapter()
	payload := []byte(`{
		"action": "opened",
		"number": 100,
		"repository": {"id": 5, "full_name": "devhub/no-sender", "name": "no-sender"}
	}`)
	headers := map[string]string{"X-Gitea-Event": "pull_request", "X-Gitea-Delivery": "delivery-5"}
	ev, err := a.ExtractEvent("gitea-main", payload, headers)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ev.ActorLogin != "" {
		t.Errorf("expected empty ActorLogin for missing sender, got %s", ev.ActorLogin)
	}
	if ev.ExternalRef != "GITEA-100" {
		t.Errorf("expected ExternalRef=GITEA-100, got %s", ev.ExternalRef)
	}
}

func TestRegisterAdapter_GetAdapterForProviderType(t *testing.T) {
	// cleanup after test
	defer func() {
		// reset registry (in case other tests added)
		adapterRegistry = map[string]WebhookAdapter{}
	}()
	// Note: this test manipulates the package-level adapterRegistry. 다른 test 와
	// 병행 실행 시 race condition 가능 — 단위 test 라 race-free (sequential).
	adapterRegistry = map[string]WebhookAdapter{}
	a := NewGiteaWebhookAdapter()
	RegisterAdapter(a)

	got, err := GetAdapterForProviderType("scm")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ProviderType() != "scm" {
		t.Errorf("expected scm, got %s", got.ProviderType())
	}

	// unknown provider_type
	_, err = GetAdapterForProviderType("unknown")
	if err == nil {
		t.Fatal("expected ErrAdapterNotFound")
	}
}

func TestFirstHeader(t *testing.T) {
	headers := map[string]string{
		"X-Gitea-Event":    "pull_request",
		"X-Gitea-Delivery": "delivery-1",
		"X-Empty":          "",
	}
	if firstHeader(headers, "X-Integration-Event", "X-Gitea-Event", "X-Gogs-Event") != "pull_request" {
		t.Errorf("expected pull_request, got %s", firstHeader(headers, "X-Integration-Event", "X-Gitea-Event"))
	}
	if firstHeader(headers, "X-Nonexistent") != "" {
		t.Errorf("expected empty, got %s", firstHeader(headers, "X-Nonexistent"))
	}
	if firstHeader(headers, "X-Empty") != "" {
		t.Errorf("expected empty (skip empty value), got %s", firstHeader(headers, "X-Empty"))
	}
	// nil map
	if firstHeader(nil, "X-Any") != "" {
		t.Errorf("expected empty for nil map, got %s", firstHeader(nil, "X-Any"))
	}
}
