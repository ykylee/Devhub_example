package httpapi

import (
	"context"
	"fmt"
)

// MockIdentityAdmin is a test-only fake that simulates IdentityAdmin and
// tracks the calls so handler tests can assert on them. Previously lived in
// kratos_admin_client.go alongside the (now removed) Kratos production
// client; relocated here in sprint claude/work_260519-ad as part of the
// Kratos residual cleanup (ADR-0019). ADR-0020 sub-carve E (sprint -n) 가
// IdentityAdmin write methods 제거 후 mock 도 read-only.
type MockIdentityAdmin struct {
	FindIDOverride map[string]string
	// FindCalls counts how many times FindIdentityByUserID was invoked. The
	// resolver cache-hit test asserts this stays at zero when the DevHub
	// users row already carries an idp_subject.
	FindCalls int
	FindError error
}

// compile-time interface check.
var _ IdentityAdmin = (*MockIdentityAdmin)(nil)

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
