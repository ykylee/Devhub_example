package httpapi

import (
	"context"
	"fmt"
)

// MockIdentityAdmin is a test-only fake that simulates IdentityAdmin and
// tracks the calls so handler tests can assert on them. Previously lived in
// kratos_admin_client.go alongside the (now removed) Kratos production
// client; relocated here in sprint claude/work_260519-ad as part of the
// Kratos residual cleanup (ADR-0019).
type MockIdentityAdmin struct {
	CreatedIDs     []string
	PasswordResets []string
	StateChanges   map[string]bool
	DeletedIDs     []string
	FindIDOverride map[string]string
	// FindCalls counts how many times FindIdentityByUserID was invoked. The
	// resolver cache-hit test asserts this stays at zero when the DevHub
	// users row already carries an idp_subject.
	FindCalls       int
	FindError       error
	UpdatePassError error
	SetStateError   error
	DeleteError     error
}

// compile-time interface check — sprint claude/work_260519-ad Stage 3
// (self-review P1-3). main_test.go 의 idpAdminFake 와 동일 패턴.
var _ IdentityAdmin = (*MockIdentityAdmin)(nil)
func (m *MockIdentityAdmin) CreateIdentity(_ context.Context, email, name, userID, password string) (string, error) {
	fakeID := fmt.Sprintf("mock-k-id-%s", userID)
	m.CreatedIDs = append(m.CreatedIDs, fakeID)
	_ = email
	_ = name
	_ = password
	return fakeID, nil
}

func (m *MockIdentityAdmin) FindIdentityByUserID(_ context.Context, userID string) (string, error) {
	m.FindCalls++
	if m.FindError != nil {
		return "", m.FindError
	}
	if id, ok := m.FindIDOverride[userID]; ok {
		return id, nil
	}
	return fmt.Sprintf("mock-k-id-%s", userID), nil
}

func (m *MockIdentityAdmin) UpdateIdentityPassword(_ context.Context, identityID, password string) error {
	if m.UpdatePassError != nil {
		return m.UpdatePassError
	}
	m.PasswordResets = append(m.PasswordResets, identityID)
	_ = password
	return nil
}

func (m *MockIdentityAdmin) SetIdentityState(_ context.Context, identityID string, active bool) error {
	if m.SetStateError != nil {
		return m.SetStateError
	}
	if m.StateChanges == nil {
		m.StateChanges = map[string]bool{}
	}
	m.StateChanges[identityID] = active
	return nil
}

func (m *MockIdentityAdmin) DeleteIdentity(_ context.Context, identityID string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	m.DeletedIDs = append(m.DeletedIDs, identityID)
	return nil
}
