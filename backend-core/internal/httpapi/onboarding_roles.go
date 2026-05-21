package httpapi

import "github.com/devhub/backend-core/internal/domain"

// onboardingValidRole checks whether the role string maps to a known
// rbac_policies seed role (developer / manager / pmo_manager / system_admin).
//
// 본 helper 는 me_onboarding.go (POST /me/onboarding) 의 fallback role
// validation + lazy_auto_create.go 의 lazy provisioned default fallback
// 두 곳에서 공유. lazy_auto_create.go 가 Carve D default ON flip 이후
// deletion 시점에 본 helper 는 onboarding 흐름에서 계속 사용 — 별도
// file 로 분리해 cross-file dependency 제거 (PR #278 self-review P1 #2).
func onboardingValidRole(role domain.AppRole) bool {
	switch role {
	case domain.AppRoleDeveloper, domain.AppRoleManager, domain.AppRoleSystemAdmin, "pmo_manager":
		return true
	}
	return false
}
