package ci

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/infrastructure/gitea"
)

func TestGiteaActionsAdapter_GetRuns(t *testing.T) {
	// 1. Mock Gitea API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/ykylee/e2e-repo-a/actions/runs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := gitea.GiteaActionRunsResponse{
			TotalCount: 1,
			WorkflowRuns: []gitea.GiteaActionRun{
				{
					ID:         456,
					RunNumber:  12,
					Event:      "push",
					Status:     "completed",
					Conclusion: "success",
					HeadBranch: "main",
					HeadSHA:    "abcdef123456",
					HTMLURL:    "http://gitea/run/456",
					CreatedAt:  time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2026, 6, 5, 12, 2, 0, 0, time.UTC),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 2. Setup Client & Adapter
	client := gitea.NewClient(server.URL, "dummy-token")
	adapter := NewGiteaActionsAdapter(client)

	// 3. Test GetRuns
	runs, err := adapter.GetRuns(context.Background(), "ykylee", "e2e-repo-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.ID != "456" {
		t.Errorf("expected run ID '456', got %s", run.ID)
	}
	if run.Status != "success" {
		t.Errorf("expected run status 'success', got %s", run.Status)
	}
	if run.DurationSeconds != 120 {
		t.Errorf("expected duration 120, got %d", run.DurationSeconds)
	}
}

func TestGiteaActionsAdapter_GetRunLogs(t *testing.T) {
	// 1. Mock Gitea API Server for logs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/repos/ykylee/e2e-repo-a/actions/runs/456/logs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("2026-06-05T12:00:00Z checkout started\n2026-06-05T12:01:00Z [error] build failed\n"))
	}))
	defer server.Close()

	// 2. Setup Client & Adapter
	client := gitea.NewClient(server.URL, "dummy-token")
	adapter := NewGiteaActionsAdapter(client)

	// 3. Test GetRunLogs
	logs, err := adapter.GetRunLogs(context.Background(), "ykylee", "e2e-repo-a", "456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logs) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(logs))
	}

	if logs[0].Message != "checkout started" {
		t.Errorf("expected message 'checkout started', got %s", logs[0].Message)
	}
	if logs[1].Level != "error" {
		t.Errorf("expected level 'error', got %s", logs[1].Level)
	}
}
