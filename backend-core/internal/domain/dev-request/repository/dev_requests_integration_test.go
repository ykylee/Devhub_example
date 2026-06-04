package repository_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	devreqrep "github.com/devhub/backend-core/internal/domain/dev-request/repository"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newDevRequestTestStore(t *testing.T) (*store.PostgresStore, *pgxpool.Pool, context.Context) {
	t.Helper()
	dbURL := os.Getenv("DEVHUB_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("DEVHUB_TEST_DB_URL is not set")
	}
	ctx := context.Background()
	pgStore, err := store.NewPostgresStore(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect postgres store: %v", err)
	}
	t.Cleanup(pgStore.Close)

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect raw pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pgStore, pool, ctx
}

func TestIntegration_DevRequests_CRUD(t *testing.T) {
	pgStore, pool, ctx := newDevRequestTestStore(t)
	dreq := devreqrep.NewDevRequestRepository(pgStore)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	externalRef := "ref-" + suffix
	sourceSystem := "system-" + suffix

	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM dev_requests WHERE external_ref = $1`, externalRef)
	}()

	// 1. Get non-existent
	ghostUUID := "00000000-0000-0000-0000-000000000000"
	if _, err := dreq.GetDevRequest(ctx, ghostUUID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for ghost UUID, got %v", err)
	}
	if _, err := dreq.GetDevRequestByExternalRef(ctx, sourceSystem, externalRef); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent external ref, got %v", err)
	}

	dr := domain.DevRequest{
		Title:          "Request Title " + suffix,
		Details:        "Request Details " + suffix,
		Requester:      "alice",
		AssigneeUserID: "u1",
		SourceSystem:   sourceSystem,
		ExternalRef:    externalRef,
		Status:         domain.DevRequestStatusPending,
		ReceivedAt:     time.Now(),
	}

	// 2. CreateDevRequest
	created, err := dreq.CreateDevRequest(ctx, dr)
	if err != nil {
		t.Fatalf("CreateDevRequest failed: %v", err)
	}
	if created.ExternalRef != externalRef || created.Title != dr.Title {
		t.Errorf("unexpected created dev_request details: %+v", created)
	}

	// Duplicate check
	_, err = dreq.CreateDevRequest(ctx, dr)
	if !errors.Is(err, store.ErrConflict) {
		t.Errorf("expected ErrConflict for duplicate external_ref, got %v", err)
	}

	// 3. Get by UUID and ExternalRef
	loaded, err := dreq.GetDevRequest(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetDevRequest failed: %v", err)
	}
	if loaded.ID != created.ID {
		t.Errorf("loaded ID mismatch: %s vs %s", loaded.ID, created.ID)
	}

	loadedRef, err := dreq.GetDevRequestByExternalRef(ctx, sourceSystem, externalRef)
	if err != nil {
		t.Fatalf("GetDevRequestByExternalRef failed: %v", err)
	}
	if loadedRef.ID != created.ID {
		t.Errorf("loaded by ref ID mismatch: %s vs %s", loadedRef.ID, created.ID)
	}

	// 4. ListDevRequests
	opts := store.DevRequestListOptions{
		Statuses:     []domain.DevRequestStatus{domain.DevRequestStatusPending},
		SourceSystem: sourceSystem,
	}
	list, total, err := dreq.ListDevRequests(ctx, opts)
	if err != nil {
		t.Fatalf("ListDevRequests failed: %v", err)
	}
	if len(list) != 1 || total != 1 {
		t.Errorf("expected 1 dev request in list, got len=%d total=%d", len(list), total)
	}

	// 5. TransitionDevRequestStatus
	transitioned, err := dreq.TransitionDevRequestStatus(ctx, created.ID, domain.DevRequestStatusInReview, "")
	if err != nil {
		t.Fatalf("TransitionDevRequestStatus failed: %v", err)
	}
	if transitioned.Status != domain.DevRequestStatusInReview {
		t.Errorf("expected status InReview, got %s", transitioned.Status)
	}

	// Reject transition
	rejected, err := dreq.TransitionDevRequestStatus(ctx, created.ID, domain.DevRequestStatusRejected, "some rejection reason")
	if err != nil {
		t.Fatalf("TransitionDevRequestStatus (reject) failed: %v", err)
	}
	if rejected.Status != domain.DevRequestStatusRejected || rejected.RejectedReason != "some rejection reason" {
		t.Errorf("unexpected rejected details: %+v", rejected)
	}

	// ReassignDevRequest
	reassigned, err := dreq.ReassignDevRequest(ctx, created.ID, "u1") // u1 is seeded admin user
	if err != nil {
		t.Fatalf("ReassignDevRequest failed: %v", err)
	}
	if reassigned.AssigneeUserID != "u1" {
		t.Errorf("expected assignee u1, got %s", reassigned.AssigneeUserID)
	}

	// 6. Transition back to pending for promotion
	_, _ = dreq.TransitionDevRequestStatus(ctx, created.ID, domain.DevRequestStatusPending, "")

	// 7. RegisterDevRequestWithNewPlatform
	appKey := fmt.Sprintf("AP%08d", time.Now().UnixNano()%100000000)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM platforms WHERE key = $1`, appKey)
	}()

	app := domain.Platform{
		Key:         appKey,
		Name:        "App from DREQ",
		Status:      domain.PlatformStatusActive,
		Visibility:  domain.PlatformVisibilityInternal,
		OwnerUserID: "u1",
	}

	primaryRepo := domain.PlatformRepository{
		RepoProvider: "gitea",
		RepoFullName: "team/devhub-core",
		Role:         domain.PlatformRepositoryRolePrimary,
		SyncStatus:   domain.SyncStatusRequested,
	}

	promotedDreq, promotedApp, err := dreq.RegisterDevRequestWithNewPlatform(ctx, created.ID, app, &primaryRepo)
	if err != nil {
		t.Fatalf("RegisterDevRequestWithNewPlatform failed: %v", err)
	}
	if promotedDreq.Status != domain.DevRequestStatusRegistered || promotedDreq.RegisteredTargetID != promotedApp.ID {
		t.Errorf("unexpected promoted devrequest details: %+v", promotedDreq)
	}
	if promotedApp.Key != appKey {
		t.Errorf("unexpected promoted application key: %s", promotedApp.Key)
	}

	// 8. RegisterDevRequestWithNewProject
	// Create another dev_request for project promotion
	drProj := domain.DevRequest{
		Title:          "Project Request Title " + suffix,
		Details:        "Project Request Details " + suffix,
		Requester:      "alice",
		AssigneeUserID: "u1",
		SourceSystem:   sourceSystem,
		ExternalRef:    externalRef + "-prj",
		Status:         domain.DevRequestStatusPending,
		ReceivedAt:     time.Now(),
	}
	createdProjDreq, err := dreq.CreateDevRequest(ctx, drProj)
	if err != nil {
		t.Fatalf("CreateProjectDreq failed: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM dev_requests WHERE external_ref = $1`, drProj.ExternalRef)
	}()

	projKey := fmt.Sprintf("PJ%08d", time.Now().UnixNano()%100000000)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM projects WHERE key = $1`, projKey)
	}()

	proj := domain.Project{
		Key:         projKey,
		Name:        "Project from DREQ",
		Status:      domain.PlatformStatusActive,
		Visibility:  domain.PlatformVisibilityInternal,
		OwnerUserID: "u1",
	}

	promotedProjDreq, promotedProj, err := dreq.RegisterDevRequestWithNewProject(ctx, createdProjDreq.ID, proj)
	if err != nil {
		t.Fatalf("RegisterDevRequestWithNewProject failed: %v", err)
	}
	if promotedProjDreq.Status != domain.DevRequestStatusRegistered || promotedProjDreq.RegisteredTargetID != promotedProj.ID {
		t.Errorf("unexpected promoted proj dreq details: %+v", promotedProjDreq)
	}
	if promotedProj.Key != projKey {
		t.Errorf("unexpected promoted project key: %s", promotedProj.Key)
	}
}
