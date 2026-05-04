package store_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/normalize"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresStoreProcessesNormalizedWebhookEvent(t *testing.T) {
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}

	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	defer pgStore.Close()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect cleanup pool: %v", err)
	}
	defer pool.Close()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	repositoryName := "devhub-test/api-" + suffix
	senderLogin := "devhub-test-user-" + suffix
	dedupeKey := "integration-" + suffix
	riskKey := "integration-risk-" + suffix
	defer cleanupIntegrationRows(ctx, t, pool, repositoryName, senderLogin, dedupeKey, riskKey)

	payload := []byte(fmt.Sprintf(`{
		"repository": {
			"id": 0,
			"full_name": %q,
			"name": "api-%s",
			"default_branch": "main",
			"owner": {"login": "devhub-test"}
		},
		"sender": {
			"id": 0,
			"login": %q
		},
		"issue": {
			"id": 0,
			"number": 1,
			"title": "Integration issue",
			"state": "open",
			"user": {"login": %q},
			"created_at": "2026-05-03T10:00:00Z"
		}
	}`, repositoryName, suffix, senderLogin, senderLogin))

	now := time.Now().UTC()
	eventID, err := pgStore.SaveWebhookEvent(ctx, store.WebhookEvent{
		EventType:      "issues",
		DedupeKey:      dedupeKey,
		RepositoryName: repositoryName,
		SenderLogin:    senderLogin,
		Payload:        payload,
		Status:         "validated",
		ReceivedAt:     now,
		ValidatedAt:    &now,
	})
	if err != nil {
		t.Fatalf("save webhook event: %v", err)
	}

	event := store.WebhookEvent{
		ID:             eventID,
		EventType:      "issues",
		DedupeKey:      dedupeKey,
		RepositoryName: repositoryName,
		SenderLogin:    senderLogin,
		Payload:        payload,
		Status:         "validated",
		ReceivedAt:     now,
		ValidatedAt:    &now,
	}
	processor := normalize.Processor{Sink: pgStore}
	if err := processor.Process(ctx, event); err != nil {
		t.Fatalf("process event: %v", err)
	}

	events, err := pgStore.ListWebhookEvents(ctx, store.ListWebhookEventsOptions{Limit: 100})
	if err != nil {
		t.Fatalf("list webhook events: %v", err)
	}
	foundEvent := false
	for _, listed := range events {
		if listed.DedupeKey == dedupeKey {
			if listed.Status != "processed" {
				t.Fatalf("expected processed status, got %q", listed.Status)
			}
			foundEvent = true
			break
		}
	}
	if !foundEvent {
		t.Fatalf("saved event with dedupe key %q was not listed", dedupeKey)
	}

	repositories, err := pgStore.ListRepositories(ctx, domain.ListOptions{Limit: 100})
	if err != nil {
		t.Fatalf("list repositories: %v", err)
	}
	if !hasRepository(repositories, repositoryName) {
		t.Fatalf("expected repository %q in list, got %+v", repositoryName, repositories)
	}

	issues, err := pgStore.ListIssues(ctx, domain.ListOptions{RepositoryName: repositoryName, State: "open", Limit: 10})
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(issues) != 1 || issues[0].RepositoryName != repositoryName || issues[0].Number != 1 {
		t.Fatalf("unexpected issues: %+v", issues)
	}

	duration := 134
	startedAt := time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC)
	if err := pgStore.UpsertCIRun(ctx, domain.CIRun{
		ExternalID:      "run-" + suffix,
		RepositoryName:  repositoryName,
		Branch:          "main",
		CommitSHA:       "abc1234",
		Status:          "success",
		Conclusion:      "success",
		StartedAt:       &startedAt,
		DurationSeconds: &duration,
	}); err != nil {
		t.Fatalf("upsert ci run: %v", err)
	}
	runs, err := pgStore.ListCIRuns(ctx, domain.ListOptions{RepositoryName: repositoryName, Status: "success", Limit: 10})
	if err != nil {
		t.Fatalf("list ci runs: %v", err)
	}
	if len(runs) != 1 || runs[0].ExternalID != "run-"+suffix || runs[0].DurationSeconds == nil || *runs[0].DurationSeconds != duration {
		t.Fatalf("unexpected ci runs: %+v", runs)
	}

	if err := pgStore.UpsertRisk(ctx, domain.Risk{
		RiskKey:          riskKey,
		Title:            "Integration risk",
		Reason:           "integration test risk",
		Impact:           "high",
		Status:           "action_required",
		SourceType:       "ci_run",
		SourceID:         "run-" + suffix,
		SuggestedActions: []string{"inspect_logs", "rerun_ci"},
		DetectedAt:       startedAt,
	}); err != nil {
		t.Fatalf("upsert risk: %v", err)
	}
	risks, err := pgStore.ListRisks(ctx, domain.ListOptions{Status: "action_required", Impact: "high", Limit: 10})
	if err != nil {
		t.Fatalf("list risks: %v", err)
	}
	if !hasRisk(risks, riskKey) {
		t.Fatalf("expected risk %q in list, got %+v", riskKey, risks)
	}

	command, auditLog, replayed, err := pgStore.CreateRiskMitigationCommand(ctx, domain.RiskMitigationCommandRequest{
		RiskID:         riskKey,
		ActorLogin:     senderLogin,
		ActionType:     "rerun_ci",
		Reason:         "integration test mitigation",
		DryRun:         true,
		IdempotencyKey: "integration-command-" + suffix,
		RequestPayload: map[string]any{"ci_run_id": "run-" + suffix},
	})
	if err != nil {
		t.Fatalf("create risk mitigation command: %v", err)
	}
	if replayed {
		t.Fatal("expected first command creation, got idempotent replay")
	}
	if command.CommandType != "risk_mitigation" || command.TargetID != riskKey || command.Status != "pending" {
		t.Fatalf("unexpected command: %+v", command)
	}
	if auditLog.AuditID == "" || auditLog.CommandID != command.CommandID {
		t.Fatalf("unexpected audit log: %+v", auditLog)
	}

	replayedCommand, replayedAuditLog, replayed, err := pgStore.CreateRiskMitigationCommand(ctx, domain.RiskMitigationCommandRequest{
		RiskID:         riskKey,
		ActorLogin:     senderLogin,
		ActionType:     "rerun_ci",
		Reason:         "integration test mitigation",
		DryRun:         true,
		IdempotencyKey: "integration-command-" + suffix,
		RequestPayload: map[string]any{"ci_run_id": "run-" + suffix},
	})
	if err != nil {
		t.Fatalf("replay risk mitigation command: %v", err)
	}
	if !replayed || replayedCommand.CommandID != command.CommandID || replayedAuditLog.AuditID != auditLog.AuditID {
		t.Fatalf("unexpected idempotent replay: command=%+v audit=%+v replayed=%v", replayedCommand, replayedAuditLog, replayed)
	}

	loadedCommand, err := pgStore.GetCommand(ctx, command.CommandID)
	if err != nil {
		t.Fatalf("get command: %v", err)
	}
	if loadedCommand.CommandID != command.CommandID || loadedCommand.TargetID != riskKey {
		t.Fatalf("unexpected loaded command: %+v", loadedCommand)
	}
}

func hasRepository(repositories []domain.Repository, fullName string) bool {
	for _, repository := range repositories {
		if repository.FullName == fullName {
			return true
		}
	}
	return false
}

func hasRisk(risks []domain.Risk, riskKey string) bool {
	for _, risk := range risks {
		if risk.RiskKey == riskKey {
			return true
		}
	}
	return false
}

func cleanupIntegrationRows(ctx context.Context, t *testing.T, pool *pgxpool.Pool, repositoryName, senderLogin, dedupeKey, riskKey string) {
	t.Helper()
	queries := []string{
		`DELETE FROM webhook_events WHERE dedupe_key = $1`,
		`DELETE FROM audit_logs WHERE target_type = 'risk' AND target_id = $1`,
		`DELETE FROM commands WHERE target_type = 'risk' AND target_id = $1`,
		`DELETE FROM risks WHERE risk_key = $1`,
		`DELETE FROM issues WHERE repository_id IN (SELECT id FROM repositories WHERE full_name = $1)`,
		`DELETE FROM pull_requests WHERE repository_id IN (SELECT id FROM repositories WHERE full_name = $1)`,
		`DELETE FROM ci_runs WHERE repository_name = $1`,
		`DELETE FROM repositories WHERE full_name = $1`,
		`DELETE FROM gitea_users WHERE login = $1`,
	}
	args := [][]any{
		{dedupeKey},
		{riskKey},
		{riskKey},
		{riskKey},
		{repositoryName},
		{repositoryName},
		{repositoryName},
		{repositoryName},
		{senderLogin},
	}
	for i, query := range queries {
		if _, err := pool.Exec(ctx, query, args[i]...); err != nil && !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("cleanup query failed: %v", err)
		}
	}
}
