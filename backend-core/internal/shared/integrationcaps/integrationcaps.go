// Package integrationcaps centralizes integration provider capability checks.
//
// Background: prior to this consolidation, `providerHasCapability` was defined
// (with OR semantics) in three separate packages — `domain/integration-registry/
// view`, `domain/repository-integration/view`, and `internal/httpapi` (as a
// router-level shim). The Gemini code-cleanup split (PR #406 SoT applied)
// briefly drifted two of the three copies to AND semantics, breaking
// `TestSyncIntegrationProvider_Happy`. PR #407 patched the AND→OR regression
// in both drifted copies but left three live definitions, keeping the drift
// surface open. This package collapses them to a single source of truth.
package integrationcaps

import "github.com/devhub/backend-core/internal/domain"

// ProviderHasCapability reports whether the provider declares any of the given
// capabilities (OR semantics).
//
// Used as a capability gate by integration registry sync, SCM repository
// import / mirror, and outbound SCM repository creation. The OR semantics
// matches the pre-cleanup main-HEAD baseline behavior: e.g. a provider with
// capabilities `["pull"]` passes `ProviderHasCapability(p, "pull", "sync")`.
//
// Capability vocabulary (sprint scm-repo-sync):
//   - pull    — SCM 으로부터 repo 조회 / import / mirror 허용
//   - sync    — 주기 mirror 허용
//   - push    — outbound repo 생성 허용 (Phase C)
//   - webhook — inbound webhook 수신
func ProviderHasCapability(p domain.IntegrationProvider, caps ...string) bool {
	for _, have := range p.Capabilities {
		for _, want := range caps {
			if have == want {
				return true
			}
		}
	}
	return false
}
