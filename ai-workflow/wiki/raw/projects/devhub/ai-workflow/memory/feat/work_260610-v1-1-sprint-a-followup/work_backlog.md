# Work Backlog — feat/work_260610-v1-1-sprint-a-followup

- **문서 목적**: 본 sprint 의 todo 목록을 planned → in_progress → done 으로 추적한다.
- **sprint branch**: `feat/work_260610-v1-1-sprint-a-followup` (PR #539, MERGED main @ `87e6c1f5`)
- **최종 갱신일**: 2026-06-10 (session 완료 — merge 완료)

## 상태 정의

- `planned` — 미착수
- `in_progress` — 작업 중
- `blocked` — 의존성/외부 입력 대기
- `done` — 검증 + 완료 확정

## 본 sprint (PR #539) task

| ID | 상태 | 제목 | 비고 |
| --- | --- | --- | --- |
| T-b1a02e8c | done | saovae_stub 작성 (4 port + webhook handler) | `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` 105 lines NEW. `go build ./...` PASS |
| T-d2f1a7b3 | done | main.go DEVHUB_BUILD_TIER env var 분기 + 3 port wiring | L148-235 분기 + 2 import 추가. internal = real KeycloakAdminClient, default = saovae_stub |
| T-c3a9e4d2 | done | ports.go KeycloakUserEvent/KeycloakAdminEvent alias 통합 | mirror struct → `type X = httpapi.X` alias. `*KeycloakAdminClient` 가 `KeycloakEventPort` 충족 위해 필수 |
| T-9f8b6c1d | done | view/ deprecation comment 3 interface | `view/auth.go:59` + `view/handler.go:27` + `view/handler.go:197`. canonical = `integration/` |
| T-7e4d2f8a | done | commit + push + PR #539 + CI 7/7 PASS + merge | commit `a00793bc`, PR https://github.com/ykylee/Devhub_example/pull/539, main HEAD `87e6c1f5` |

## carry-over (다음 sprint 후보)

| ID | 우선순위 | 제목 | 출처 |
| --- | --- | --- | --- |
| C-a | P0 | `sso-integrations/keycloak/verifier.go` + `admin_client.go` real adapter 작성 | sprint -a follow-up 다음 PR |
| C-b | P0 | main.go event listener type assertion 정리 (`*httpapi.KeycloakAdminClient` → `KeycloakEventPort` interface) | sprint -a follow-up 다음 PR |
| C-c | P0 | `_ = keycloakEventPort` placeholder 제거 (C-b 완료 시) | sprint -a follow-up 다음 PR |
| C-d | P1 | v1.0 mirror struct 제거: `httpapi.KeycloakUserEvent` / `httpapi.KeycloakAdminEvent` → `integration/` 의 alias 만 유지 | sprint -a follow-up 다음 PR |
| C-e | P1 | audit-ops 의 mirror 와 통합 (cross-package) | sprint -a follow-up 다음 PR |
| C-f | P1 | `infra/idp/_archive_2026-06-10/` immutable archive (real adapter 이전 후) | sprint -a follow-up 다음 PR |
| C-g | P2 | `docs/traceability/report.md` IMPL-30/31/32 row 갱신 | 본 follow-up 본 PR 의 traceability 갱신 |
| C-h | P2 | ADR-0030 §5 timeline 갱신 (real adapter 완료 후) | ADR-0030 |
| C-i | P2 | E2E test 추가: saovae_stub path + real adapter path (CI matrix) | sprint -a follow-up 다음 PR |
| C-j | P3 | build tag 정책 재검토 (현재 runtime injection, build tag 전환 시 trade-off) | ADR-0030 §2.3 |
