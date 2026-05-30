package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresStoreListUsersAndHierarchy(t *testing.T) {
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

	users, total, err := pgStore.ListUsers(ctx, domain.UserListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total < 3 {
		t.Fatalf("expected at least 3 seeded users, got total=%d", total)
	}

	expectedUsers := map[string]struct {
		role          domain.AppRole
		primaryUnitID string
	}{
		"u1": {role: domain.AppRoleSystemAdmin, primaryUnitID: "dept-eng"},
		"u2": {role: domain.AppRoleManager, primaryUnitID: "dept-prod"},
		"u3": {role: domain.AppRoleDeveloper, primaryUnitID: "team-infra"},
	}

	seen := make(map[string]bool)
	for _, user := range users {
		expected, ok := expectedUsers[user.UserID]
		if !ok {
			continue
		}
		seen[user.UserID] = true
		if user.Role != expected.role {
			t.Fatalf("user %s expected role %q, got %q", user.UserID, expected.role, user.Role)
		}
		if user.PrimaryUnitID != expected.primaryUnitID {
			t.Fatalf("user %s expected primary unit %q, got %q", user.UserID, expected.primaryUnitID, user.PrimaryUnitID)
		}
	}
	for userID := range expectedUsers {
		if !seen[userID] {
			t.Fatalf("expected seeded user %s missing from ListUsers result", userID)
		}
	}

	loaded, err := pgStore.GetUser(ctx, "u1")
	if err != nil {
		t.Fatalf("get user u1: %v", err)
	}
	if loaded.UserID != "u1" || loaded.Role != domain.AppRoleSystemAdmin {
		t.Fatalf("unexpected loaded user: %+v", loaded)
	}
	if len(loaded.Appointments) < 2 {
		t.Fatalf("expected at least 2 appointments for u1, got %d", len(loaded.Appointments))
	}
	hasOrgRoot := false
	for _, appointment := range loaded.Appointments {
		if appointment.UnitID == "org-root" && appointment.AppointmentRole == domain.AppointmentRoleLeader {
			hasOrgRoot = true
		}
	}
	if !hasOrgRoot {
		t.Fatalf("expected u1 to be leader of org-root, appointments=%+v", loaded.Appointments)
	}

	hierarchy, err := pgStore.GetHierarchy(ctx)
	if err != nil {
		t.Fatalf("get hierarchy: %v", err)
	}
	if len(hierarchy.Units) < 7 {
		t.Fatalf("expected at least 7 seeded org units, got %d", len(hierarchy.Units))
	}
}

func TestPostgresStoreReplaceUnitMembers(t *testing.T) {
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

	const unitID = "team-frontend"

	// Replace with empty list
	if err := pgStore.ReplaceUnitMembers(ctx, unitID, nil); err != nil {
		t.Fatalf("replace unit members with empty list: %v", err)
	}
	membersAfter, err := pgStore.ListUnitMembers(ctx, unitID)
	if err != nil {
		t.Fatalf("list unit members after empty replace: %v", err)
	}
	for _, member := range membersAfter {
		appointments, err := pgStore.GetUserAppointments(ctx, member.UserID)
		if err != nil {
			t.Fatalf("get appointments for %s: %v", member.UserID, err)
		}
		hasLeader := false
		for _, appointment := range appointments {
			if appointment.UnitID == unitID && appointment.AppointmentRole == domain.AppointmentRoleLeader {
				hasLeader = true
				break
			}
		}
		if !hasLeader {
			t.Fatalf("user %s should not remain attached to %s without a leader appointment", member.UserID, unitID)
		}
	}
}
