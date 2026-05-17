package adapters

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHomeLabHTTPPullerPullSnapshot(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"agent_id":"homelab-agent-a",
			"snapshot_at":"2026-05-17T00:00:00Z",
			"trace_id":"trc_http_01",
			"nodes":[{"node_id":"n1"}],
			"services":[{"service_id":"svc-1","health_status":"healthy"}]
		}`))
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL, Token: "token-123"}
	raw, err := puller.PullSnapshot(context.Background())
	if err != nil {
		t.Fatalf("pull snapshot: %v", err)
	}
	if raw.AgentID != "homelab-agent-a" || raw.SnapshotAt != "2026-05-17T00:00:00Z" {
		t.Fatalf("unexpected raw snapshot: %+v", raw)
	}
	if authHeader != "Bearer token-123" {
		t.Fatalf("unexpected auth header: %q", authHeader)
	}
}

func TestHomeLabHTTPPullerRejectsInvalidPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"agent_id":"","snapshot_at":""}`))
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL}
	_, err := puller.PullSnapshot(context.Background())
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

func TestHomeLabHTTPPullerRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad upstream secret-token", http.StatusBadGateway)
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL}
	_, err := puller.PullSnapshot(context.Background())
	if err == nil {
		t.Fatal("expected error for non-2xx")
	}
	if msg := err.Error(); msg == "" {
		t.Fatalf("unexpected error message: %q", msg)
	}
	if msg := err.Error(); strings.Contains(msg, "secret-token") {
		t.Fatalf("error leaked response body: %q", msg)
	}
}

func TestHomeLabHTTPPullerRespectsContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"agent_id":"homelab-agent-a",
			"snapshot_at":"2026-05-17T00:00:00Z",
			"nodes":[{"node_id":"n1"}]
		}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	puller := HomeLabHTTPPuller{
		URL:    server.URL,
		Client: &http.Client{Timeout: time.Second},
	}
	_, err := puller.PullSnapshot(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestHomeLabHTTPPullerRetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			http.Error(w, "upstream temporary failure", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"agent_id":"homelab-agent-a",
			"snapshot_at":"2026-05-17T00:00:00Z",
			"nodes":[{"node_id":"n1"}]
		}`))
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{
		URL:          server.URL,
		RetryMax:     3,
		RetryBackoff: 10 * time.Millisecond,
	}
	_, err := puller.PullSnapshot(context.Background())
	if err != nil {
		t.Fatalf("expected success after retries: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls=%d; want 3", calls)
	}
}

func TestHomeLabHTTPPullerStopsAfterRetryLimit(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, "still broken", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{
		URL:          server.URL,
		RetryMax:     2,
		RetryBackoff: 10 * time.Millisecond,
	}
	_, err := puller.PullSnapshot(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 3 {
		t.Fatalf("calls=%d; want 3", calls)
	}
	if msg := err.Error(); msg == "" {
		t.Fatal("expected non-empty error message")
	}
}
