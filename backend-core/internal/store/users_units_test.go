package store_test

import (
	"context"
	"os"
	"sort"
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
	if len(hierarchy.Edges) < 6 {
		t.Fatalf("expected at least 6 hierarchy edges, got %d", len(hierarchy.Edges))
	}

	unitsByID := make(map[string]domain.OrgUnit, len(hierarchy.Units))
	for _, unit := range hierarchy.Units {
		unitsByID[unit.UnitID] = unit
	}

	orgRoot, ok := unitsByID["org-root"]
	if !ok {
		t.Fatalf("expected org-root in hierarchy")
	}
	if orgRoot.UnitType != domain.UnitTypeCompany {
		t.Fatalf("expected org-root unit_type=company, got %q", orgRoot.UnitType)
	}
	if orgRoot.TotalCount < 3 {
		t.Fatalf("expected org-root total_count >= 3 (3 seeded users), got %d", orgRoot.TotalCount)
	}

	// dept-eng has u1 directly assigned plus descendants (team-infra has u3).
	deptEng := unitsByID["dept-eng"]
	if deptEng.DirectCount < 1 {
		t.Fatalf("expected dept-eng direct_count >= 1, got %d", deptEng.DirectCount)
	}
	if deptEng.TotalCount < 2 {
		t.Fatalf("expected dept-eng total_count >= 2 (u1 + u3 in team-infra), got %d", deptEng.TotalCount)
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

	// Snapshot initial member appointments so we can restore them.
	initialMembers, err := snapshotUnitMembers(ctx, pool, unitID)
	if err != nil {
		t.Fatalf("snapshot initial members: %v", err)
	}
	defer func() {
		if err := restoreUnitMembers(ctx, pool, unitID, initialMembers); err != nil {
			t.Fatalf("restore initial members: %v", err)
		}
	}()

	if err := pgStore.ReplaceUnitMembers(ctx, unitID, []string{"u3"}); err != nil {
		t.Fatalf("replace unit members: %v", err)
	}

	members, err := pgStore.ListUnitMembers(ctx, unitID)
	if err != nil {
		t.Fatalf("list unit members: %v", err)
	}

	memberIDs := make([]string, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.UserID)
	}
	sort.Strings(memberIDs)

	// At minimum u3 should be a member. The unit's leader user (u3 per seed)
	// may or may not have a leader appointment row, but if it does it should
	// be preserved.
	foundU3 := false
	for _, id := range memberIDs {
		if id == "u3" {
			foundU3 = true
			break
		}
	}
	if !foundU3 {
		t.Fatalf("expected u3 to be a member of %s after replace, got %v", unitID, memberIDs)
	}

	// Replace with empty list — u3 should disappear from member rows but any
	// leader appointments preserved.
	if err := pgStore.ReplaceUnitMembers(ctx, unitID, nil); err != nil {
		t.Fatalf("replace unit members with empty list: %v", err)
	}
	membersAfter, err := pgStore.ListUnitMembers(ctx, unitID)
	if err != nil {
		t.Fatalf("list unit members after empty replace: %v", err)
	}
	for _, member := range membersAfter {
		// Any remaining row must come from a leader appointment.
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

type appointmentSnapshot struct {
	userID          string
	appointmentRole string
}

func snapshotUnitMembers(ctx context.Context, pool *pgxpool.Pool, unitID string) ([]appointmentSnapshot, error) {
	rows, err := pool.Query(ctx,
		`SELECT user_id, appointment_role FROM unit_appointments WHERE unit_id = $1 ORDER BY user_id ASC`,
		unitID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]appointmentSnapshot, 0)
	for rows.Next() {
		var item appointmentSnapshot
		if err := rows.Scan(&item.userID, &item.appointmentRole); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func restoreUnitMembers(ctx context.Context, pool *pgxpool.Pool, unitID string, snapshot []appointmentSnapshot) error {
	if _, err := pool.Exec(ctx, `DELETE FROM unit_appointments WHERE unit_id = $1`, unitID); err != nil {
		return err
	}
	for _, item := range snapshot {
		if _, err := pool.Exec(ctx,
			`INSERT INTO unit_appointments (user_id, unit_id, appointment_role) VALUES ($1, $2, $3)`,
			item.userID, unitID, item.appointmentRole,
		); err != nil {
			return err
		}
	}
	return nil
}
