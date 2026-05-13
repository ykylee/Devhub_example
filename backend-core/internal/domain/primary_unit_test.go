package domain

import "testing"

// ResolvePrimaryUnit unit tests (ADR-0010, sprint claude/work_260513-n).
// Covers the four decision cases from the ADR:
//   1. single leader → that unit
//   2. no leader, multiple members → highest total_count wins
//   3. multiple leaders (invariant violation) → highest total_count wins,
//      lexicographic tie-break
//   4. empty appointments → ("", false)

func TestResolvePrimaryUnit_SingleLeader(t *testing.T) {
	appointments := []UnitAppointment{
		{UnitID: "team-A", UserID: "u1", AppointmentRole: AppointmentRoleMember},
		{UnitID: "team-B", UserID: "u1", AppointmentRole: AppointmentRoleLeader},
	}
	counts := map[string]int{"team-A": 10, "team-B": 3}

	got, ok := ResolvePrimaryUnit(appointments, counts)
	if !ok || got != "team-B" {
		t.Fatalf("expected (team-B, true) — leader trumps higher total_count, got (%q, %v)", got, ok)
	}
}

func TestResolvePrimaryUnit_NoLeader_BiggestTotalCountWins(t *testing.T) {
	appointments := []UnitAppointment{
		{UnitID: "team-A", UserID: "u1", AppointmentRole: AppointmentRoleMember},
		{UnitID: "team-B", UserID: "u1", AppointmentRole: AppointmentRoleMember},
		{UnitID: "team-C", UserID: "u1", AppointmentRole: AppointmentRoleMember},
	}
	counts := map[string]int{"team-A": 4, "team-B": 9, "team-C": 4}

	got, ok := ResolvePrimaryUnit(appointments, counts)
	if !ok || got != "team-B" {
		t.Fatalf("expected (team-B, true), got (%q, %v)", got, ok)
	}
}

func TestResolvePrimaryUnit_LeaderTie_LexicographicWins(t *testing.T) {
	// Invariant violation: u1 is leader of two units. Algorithm picks
	// the bigger total_count; when those tie, the lexicographically
	// smaller unit_id wins (deterministic).
	appointments := []UnitAppointment{
		{UnitID: "team-Z", UserID: "u1", AppointmentRole: AppointmentRoleLeader},
		{UnitID: "team-A", UserID: "u1", AppointmentRole: AppointmentRoleLeader},
	}
	counts := map[string]int{"team-A": 5, "team-Z": 5}

	got, ok := ResolvePrimaryUnit(appointments, counts)
	if !ok || got != "team-A" {
		t.Fatalf("expected (team-A, true) — lex tie-break, got (%q, %v)", got, ok)
	}
}

func TestResolvePrimaryUnit_EmptyAppointmentsReturnsFalse(t *testing.T) {
	got, ok := ResolvePrimaryUnit(nil, nil)
	if ok || got != "" {
		t.Fatalf("expected (\"\", false), got (%q, %v)", got, ok)
	}
}

func TestResolvePrimaryUnit_MissingCountKeyTreatedAsZero(t *testing.T) {
	// total_count map may not contain every unit (stale MV).
	// Missing keys must be treated as zero so the algorithm still
	// terminates deterministically.
	appointments := []UnitAppointment{
		{UnitID: "team-A", UserID: "u1", AppointmentRole: AppointmentRoleMember},
		{UnitID: "team-B", UserID: "u1", AppointmentRole: AppointmentRoleMember},
	}
	counts := map[string]int{"team-A": 3} // team-B missing

	got, ok := ResolvePrimaryUnit(appointments, counts)
	if !ok || got != "team-A" {
		t.Fatalf("expected (team-A, true) — missing key treated as 0, got (%q, %v)", got, ok)
	}
}
