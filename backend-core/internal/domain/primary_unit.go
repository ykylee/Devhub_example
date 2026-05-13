package domain

import "sort"

// ResolvePrimaryUnit returns the unit a user should be assigned as their
// primary department when `users.primary_unit_id` is empty or stale
// (ADR-0010, sprint claude/work_260513-n). The algorithm is deterministic:
//
//   1. Filter to appointments where AppointmentRole == leader. If none,
//      fall back to all appointments (member-only case).
//   2. Sort candidates by unitTotalCounts (descending) so the unit with
//      the biggest sub-tree wins when multiple leaders or members tie.
//   3. Break remaining ties lexicographically by UnitID so the result is
//      reproducible across calls.
//
// The second return value is false when appointments is empty — caller
// stores an empty primary_unit_id in that case.
//
// `unitTotalCounts` is whatever the caller has on hand: the cached MV
// from migration 000011 in production, or a fresh count from a recursive
// CTE in tests / batch tools. Missing keys default to 0.
//
// Callers that already trust users.primary_unit_id (admin-set or
// recently backfilled and present in appointments) MUST skip this helper
// — the algorithm is a fallback, not an override.
func ResolvePrimaryUnit(appointments []UnitAppointment, unitTotalCounts map[string]int) (string, bool) {
	if len(appointments) == 0 {
		return "", false
	}

	candidates := make([]UnitAppointment, 0, len(appointments))
	for _, a := range appointments {
		if a.AppointmentRole == AppointmentRoleLeader {
			candidates = append(candidates, a)
		}
	}
	if len(candidates) == 0 {
		candidates = append(candidates, appointments...)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		ci := unitTotalCounts[candidates[i].UnitID]
		cj := unitTotalCounts[candidates[j].UnitID]
		if ci != cj {
			return ci > cj
		}
		return candidates[i].UnitID < candidates[j].UnitID
	})
	return candidates[0].UnitID, true
}
