# Work Backlog — feat/work_260610-v1-1-sprint-a-real-adapter

- **문서 목적**: 본 sprint 의 todo 목록을 planned → in_progress → done 으로 추적한다.
- **sprint branch**: `feat/work_260610-v1-1-sprint-a-real-adapter` (코드 작업 완료, commit + PR 발행 대기)
- **시작일**: 2026-06-10
- **종료 시점**: 2026-06-10 (검증 완료)
- **관련 문서**: [`session_handoff.md`](./session_handoff.md), [`backlog/2026-06-10.md`](./backlog/2026-06-10.md), [`state.json`](./state.json)

## 상태 정의

- `planned` — 미착수
- `in_progress` — 작업 중
- `blocked` — 의존성/외부 입력 대기
- `done` — 검증 + 완료 확정

## 본 sprint (PR1, C-a~C-f 풀번들) task

| ID | 상태 | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- | --- |
| T-c1a02e8c | done | P0 | C-a: real adapter 작성 (verifier + admin_client) | `sso-integrations/keycloak/{verifier,admin_client}.go` 신규 + metrics + 4 test files 이전 |
| T-c2f1a7b3 | done | P0 | C-b: main.go event listener type assertion 정리 | `*httpapi.KeycloakAdminClient` → `*keycloakadapter.KeycloakAdminClient` + type assertion 제거 + keycloakAdminEventLister adapter 제거 |
| T-c3a9e4d2 | done | P0 | C-c: `_ = keycloakEventPort` placeholder 제거 | keycloakEventPort 가 event listener 의 lister 로 직접 주입 |
| T-c4d2f8a1 | done | P1 | C-d: v1.0 mirror struct 제거 (httpapi → integration struct 직접 정의) | ports.go 의 alias → struct 직접 정의 (flat shape) |
| T-c5e8b6c3 | done | P1 | C-e: audit-ops mirror 통합 | mirror struct → `type X = integration.X` alias + KeycloakEventLister interface 제거 + `RunKeycloakEventPuller` 가 `integration.KeycloakEventPort` 직접 받음 |
| T-c6f7a1d2 | done | P1 | C-f: infra/idp/_archive_2026-06-10/ immutable archive | `identity.schema.json` archive + README 갱신 |
| T-c7e9d3f1 | done | P1 | 검증: go build + go test + backend-integration | `go build ./...` PASS + `go test ./...` (전 패키지) PASS + `go vet` clean (영향 패키지) + saovae/internal build 둘 다 PASS |
| T-c8a1b5d7 | pending | P1 | commit + push + PR 발행 | 사용자 confirm 대기 (정책: explicit instruction only) |

## carry-over (다음 PR2)

| ID | 우선순위 | 제목 | 비고 |
| --- | --- | --- | --- |
| C-g | P2 | `docs/traceability/report.md` IMPL-30/31/32 row 갱신 | 본 PR1 의 PR 본문에서 traceability 영향 명시 + 별도 row 갱신은 PR2 |
| C-h | P2 | ADR-0030 §5 timeline 갱신 (real adapter 완료 후) | PR1 머지 후 |
| C-i | P2 | E2E test 추가: saovae_stub path + real adapter path (CI matrix) | DEVHUB_BUILD_TIER=internal matrix |
| C-j | P3 | build tag 정책 재검토 (현재 runtime injection, build tag 전환 시 trade-off) | 별도 ADR 후보 |

## 사이드 점검 (PR1 진행 시 발견 시)

- `httpapi.KeycloakUserDetails` / `httpapi.KeycloakGroup` (admin event handler 가 사용) — 본 PR1 에서 sso-integrations/keycloak/ 로 함께 이전 필요. callers grep 후 결정.
- `httpapi/identity_resolver_test.go` + `identity_admin_mock_test.go` — sso-integrations/keycloak/ 와 무관 (Kratos 시절 잔재 정리 후), 본 PR1 scope 외.
