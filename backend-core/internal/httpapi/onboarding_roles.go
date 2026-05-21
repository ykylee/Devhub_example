package httpapi

import "github.com/devhub/backend-core/internal/domain"

// onboardingValidRole checks whether the role string maps to a known
// rbac_policies seed role (developer / manager / pmo_manager / system_admin).
//
// 본 helper 는 me_onboarding.go (POST /me/onboarding) 의 fallback role
// validation 에서 사용. 2026-05-21 lazy 폐기 sprint (issue #284) 이전에는
// lazy_auto_create.go 의 lazy provisioned default fallback 과 공유했으나
// lazy_auto_create.go deletion 이후 본 helper 는 onboarding 흐름 전용.
func onboardingValidRole(role domain.AppRole) bool {
	switch role {
	case domain.AppRoleDeveloper, domain.AppRoleManager, domain.AppRoleSystemAdmin, "pmo_manager":
		return true
	}
	return false
}
