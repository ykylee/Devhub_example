# Work Backlog — claude/work_260515-o (DREQ-Admin-UI backend)

- 상태: in_progress (PR open 대기)
- 최종 수정일: 2026-05-15
- 위치: DREQ carve out 2/4 part 1

## items

- [done] sprint 메모리 초기화
- [done] ADR-0014 — DREQ intake token admin 정책 ADR 신규
- [done] domain.Resource 확장 (ResourceDevRequestIntakeTokens) + DefaultPermissionMatrix 4 role 갱신
- [done] IntakeTokenStore interface 확장 (Create / List / Revoke)
- [done] PostgresStore admin 메서드 + scanIntakeToken helper
- [done] handler 3개 (createDevRequestIntakeToken / list / revoke) + validateAllowedIPs + generatePlainIntakeToken
- [done] router 3 endpoint + routePermissionTable 3 row
- [done] migration 000026 RBAC seed (up/down)
- [done] fakeIntakeTokenStore 확장 + 8 신규 unit test (go test ./... PASS)
- [done] backend_api_contract §14.9 + §14.10 신규 + 에러 카탈로그 갱신
- [done] traceability §3 DREQ row + §4 ADR-0014 row
- [done] development_roadmap §3 M5 DREQ entry 갱신
- [planned] commit + push + PR + 4단계 self-review + squash merge

## 다음 carve (사용자 지시 진행 중)

- **2/4 part 2 (sprint p)**: frontend `/admin/settings/dev-request-tokens` 페이지
- **3/4 (sprint q)**: DREQ-E2E (Playwright + TC-DREQ-*)
