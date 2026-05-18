package adapters

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestHomeLabFilePullerPullSnapshot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")
	content := `{
		"agent_id":"homelab-agent-a",
		"snapshot_at":"2026-05-16T15:00:00Z",
		"trace_id":"trc_01",
		"nodes":[{"node_id":"node-1"}],
		"services":[{"service_id":"svc-1","health_status":"healthy"}]
	}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	puller := HomeLabFilePuller{Path: path}
	raw, err := puller.PullSnapshot(context.Background())
	if err != nil {
		t.Fatalf("pull snapshot: %v", err)
	}
	if raw.AgentID != "homelab-agent-a" || raw.SnapshotAt != "2026-05-16T15:00:00Z" {
		t.Fatalf("unexpected raw snapshot: %+v", raw)
	}
}

func TestHomeLabFilePullerRejectsInvalidPayload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{"agent_id":"","snapshot_at":""}`), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	puller := HomeLabFilePuller{Path: path}
	_, err := puller.PullSnapshot(context.Background())
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

func TestHomeLabFilePullerRejectsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.json")
	if err := os.WriteFile(path, []byte(`{not-json}`), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	puller := HomeLabFilePuller{Path: path}
	_, err := puller.PullSnapshot(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed json")
	}
}

// ADR-0015 §6 (1) — size limit + streaming decode 회귀 가드 (sprint claude/work_260518-p).
// MaxBytes 보다 큰 fixture 는 ErrInvalidHomeLabSnapshot — 운영 가드 (DoS / memory 압박 후보).
func TestHomeLabFilePullerRejectsOversizedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oversize.json")
	// 1 KB padding 으로 200 byte limit 을 명확히 초과.
	padding := make([]byte, 1024)
	for i := range padding {
		padding[i] = 'x'
	}
	content := `{"agent_id":"homelab-agent-a","snapshot_at":"2026-05-18T00:00:00Z","nodes":[{"node_id":"n1","_padding":"` + string(padding) + `"}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	puller := HomeLabFilePuller{Path: path, MaxBytes: 200}
	_, err := puller.PullSnapshot(context.Background())
	if !errors.Is(err, ErrInvalidHomeLabSnapshot) {
		t.Fatalf("err=%v; want ErrInvalidHomeLabSnapshot", err)
	}
}

// MaxBytes = 0 은 unlimited (legacy behavior) — large fixture 도 정상 처리.
func TestHomeLabFilePullerUnlimitedWhenMaxBytesZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ok.json")
	content := `{
		"agent_id":"homelab-agent-a",
		"snapshot_at":"2026-05-18T00:00:00Z",
		"nodes":[{"node_id":"n1"}],
		"services":[{"service_id":"svc-1","health_status":"healthy"}]
	}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	puller := HomeLabFilePuller{Path: path, MaxBytes: 0}
	raw, err := puller.PullSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if raw.AgentID != "homelab-agent-a" {
		t.Fatalf("unexpected agent_id: %q", raw.AgentID)
	}
}
