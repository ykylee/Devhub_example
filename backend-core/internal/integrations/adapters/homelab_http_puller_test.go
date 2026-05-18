package adapters

import (
	"context"
	"errors"
	"fmt"
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

// ADR-0015 §6 (1) — Content-Length 사전 검사 회귀 가드 (sprint claude/work_260518-p).
// MaxBytes 보다 큰 Content-Length 면 body 다운로드 전에 reject.
func TestHomeLabHTTPPullerRejectsOversizedContentLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 명시 Content-Length 로 사전 reject 경로 검증.
		body := []byte(`{"agent_id":"homelab-agent-a","snapshot_at":"2026-05-18T00:00:00Z","nodes":[{"node_id":"n1"}]}`)
		w.Header().Set("Content-Length", "1048576") // 1 MB 거짓 advertise
		_, _ = w.Write(body)
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL, MaxBytes: 100}
	_, err := puller.PullSnapshot(context.Background())
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

// Content-Length 미제공 (또는 0) 경우 LimitReader 가 cap. body 가 limit 초과 시
// json decoder 가 unexpected EOF → ErrInvalidHomeLabSnapshot.
func TestHomeLabHTTPPullerRejectsBodyOverLimit(t *testing.T) {
	// 큰 padding payload 를 chunked transfer 로 전송 (Content-Length 미제공).
	padding := make([]byte, 4096)
	for i := range padding {
		padding[i] = 'x'
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Content-Length 안 설정 — chunked transfer 로 알려져 LimitReader 가 streaming cap.
		_, _ = fmt.Fprintf(w, `{"agent_id":"homelab-agent-a","snapshot_at":"2026-05-18T00:00:00Z","nodes":[{"node_id":"n1","_padding":"%s"}]}`, string(padding))
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL, MaxBytes: 200}
	_, err := puller.PullSnapshot(context.Background())
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

// MaxBytes = 0 은 unlimited (legacy behavior).
func TestHomeLabHTTPPullerUnlimitedWhenMaxBytesZero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"agent_id":"homelab-agent-a",
			"snapshot_at":"2026-05-18T00:00:00Z",
			"nodes":[{"node_id":"n1"}],
			"services":[{"service_id":"svc-1","health_status":"healthy"}]
		}`))
	}))
	defer server.Close()

	puller := HomeLabHTTPPuller{URL: server.URL, MaxBytes: 0}
	raw, err := puller.PullSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if raw.AgentID != "homelab-agent-a" {
		t.Fatalf("unexpected agent_id: %q", raw.AgentID)
	}
}
