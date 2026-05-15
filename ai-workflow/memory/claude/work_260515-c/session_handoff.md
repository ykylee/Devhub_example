# Session Handoff — claude/work_260515-c

- 브랜치: `claude/work_260515-c`
- Base: `main` @ `68f031e` (PR #117 머지 직후)
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: ADR-0011 §4.2 의 `enforceRowOwnership` helper 도입 — Application/Project Owner 위양 (2차 단계) entry point. REQ-FR-PROJ-009 활성화.

## 배경

ADR-0011 (sprint claude/work_260514-a, accepted 2026-05-14) 가 옵션 C (handler/service 코드 검증) 를 1차 채택. §4.2 가 2차 단계 (Owner 위양) 진입 시 `enforceRowOwnership(c, ownerUserID, allowedRoles...) bool` 시그니처 + audit `auth.row_denied` action 명세. 본 sprint 는 helper + 단위 테스트만 도입 (handler 적용은 별도 sprint).

## 시그니처

```go
// enforceRowOwnership returns true when the caller may write to a row whose
// owner is ownerUserID. Allow if any of:
//   1. actor.role == AppRoleSystemAdmin
//   2. actor.role ∈ allowedRoles (예: "pmo_manager")
//   3. actor.login == ownerUserID (owner-self)
// On deny: emits auth.row_denied audit + 403 + AbortWithStatusJSON.
// ADR-0011 §4.2.
func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool
```

## 작업 순서

1. (done) ADR-0011 § 4.2 + § 6 시그니처 / audit 패턴 확인
2. (done) 기존 permissions.go enforceRoutePermission 패턴 + recordAuditBestEffort + actor extraction 키 (devhub_actor_login/role) 검증
3. (in progress) permissions.go 에 helper 추가
4. (planned) permissions_test.go 에 allow/deny 매트릭스 테스트 추가
5. (planned) docs 갱신 — ADR-0011 §5/§7, requirements.md §5.4.1, traceability report.md
6. (planned) go build + go test 검증
7. (planned) commit + push + PR + CI green + 본인 4단계 리뷰 + 머지

## 다음 sprint (본 sprint 머지 후)

- Application/Project handler 의 update / archive / link-repo / project create 등에 `h.enforceRowOwnership(c, app.OwnerUserID, "pmo_manager")` 호출 도입 (pmo_manager RBAC matrix seed 결정 후)
- critical 임계치 외부화 / Repository commit activity / Project 가드 / M4 RM-M4-XX
