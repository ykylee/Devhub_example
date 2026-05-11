package domain

import "testing"

// TestCommandStatus_TransitionMatrix locks the 6-state lifecycle defined in
// docs/backend_api_contract.md §2 (CommandStatus enum) and the worker /
// approval handler boundaries that consume it. Anything not explicitly
// allowed must be denied — same-state transitions included.
func TestCommandStatus_TransitionMatrix(t *testing.T) {
	all := []CommandStatus{
		CommandStatusPending,
		CommandStatusRunning,
		CommandStatusSucceeded,
		CommandStatusFailed,
		CommandStatusRejected,
		CommandStatusCancelled,
	}

	allowed := map[CommandStatus]map[CommandStatus]bool{
		CommandStatusPending: {
			CommandStatusRunning:   true,
			CommandStatusRejected:  true,
			CommandStatusCancelled: true,
		},
		CommandStatusRunning: {
			CommandStatusSucceeded: true,
			CommandStatusFailed:    true,
			CommandStatusCancelled: true,
		},
	}

	for _, from := range all {
		for _, to := range all {
			expect := allowed[from][to]
			got := from.CanTransitionTo(to)
			if got != expect {
				t.Errorf("CanTransitionTo(%s -> %s) = %t, want %t", from, to, got, expect)
			}
		}
	}
}

// TestCommandStatus_IsTerminal locks the terminal-state set used by store /
// worker logic to decide whether an in-flight update is allowed.
func TestCommandStatus_IsTerminal(t *testing.T) {
	terminal := map[CommandStatus]bool{
		CommandStatusSucceeded: true,
		CommandStatusFailed:    true,
		CommandStatusRejected:  true,
		CommandStatusCancelled: true,
	}
	nonTerminal := []CommandStatus{CommandStatusPending, CommandStatusRunning}

	for s, want := range terminal {
		if got := s.IsTerminal(); got != want {
			t.Errorf("IsTerminal(%s) = %t, want %t", s, got, want)
		}
	}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("IsTerminal(%s) = true, want false", s)
		}
	}
}

// TestCommandStatus_NoSelfTransition reinforces that CanTransitionTo always
// returns false for s -> s, even on non-terminal states. The handler layer
// uses this to reject idempotent re-application of the current status.
func TestCommandStatus_NoSelfTransition(t *testing.T) {
	for _, s := range []CommandStatus{
		CommandStatusPending,
		CommandStatusRunning,
		CommandStatusSucceeded,
		CommandStatusFailed,
		CommandStatusRejected,
		CommandStatusCancelled,
	} {
		if s.CanTransitionTo(s) {
			t.Errorf("CanTransitionTo(%s -> %s) returned true; same-state must be denied", s, s)
		}
	}
}
