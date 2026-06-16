# Session Handoff — sprint `claude/test-integration-bindings-handlers-2026-05-21`

- 문서 목적: P1 carve "Integration bindings PATCH/DELETE backend handler 테스트" 처리 sprint 의 진척 + 다음 sprint 인계.
- 범위: main `ddb5591` 기준. 직전 sprint claude/test-gaps-p0-2026-05-21 (PR #262) 이 인계한 P1 갭 중 1번 처리.
- 상태: 작업 완료, PR 발행 예정
- 최종 수정일: 2026-05-21
- 관련 문서: [본 sprint state](./state.json), [직전 sprint](../test-gaps-p0-2026-05-21/session_handoff.md)

## 1. 진척 요약

| 단계 | 결과 | commit |
| --- | --- | --- |
| Audit — 기존 binding test coverage 식별 | ✅ Create 2건만 cover, PATCH/DELETE 분기 zero | — |
| fake `UpdateIntegrationBinding` 보강 | ✅ FK 가드 + 4-tuple unique + Policy/Enabled 갱신 | (단일 commit) |
| PATCH 5 test 추가 (Happy/404/422/409/403) | ✅ | (단일 commit) |
| DELETE 3 test 추가 (Happy/404/403) | ✅ | (단일 commit) |
| sprint memory + push + PR | 🔄 IN PROGRESS | — |

## 2. Gap 분석 (audit 결과)

`backend-core/internal/httpapi/integration_registry_test.go` 의 binding 영역:

| Method | Path | 기존 cover | 추가 |
| --- | --- | --- | --- |
| POST | `/api/v1/integration/bindings` | Happy + Forbidden | — |
| GET | `/api/v1/integration/bindings` | (간접 — DeleteHappy 검증에 사용) | — |
| **PATCH** | `/api/v1/integration/bindings/:binding_id` | **none** | **5 신규** |
| **DELETE** | `/api/v1/integration/bindings/:binding_id` | **none** | **3 신규** |

handler 분기는 모두 구현돼 있었으나 회귀 가드 zero — 본 sprint 가 메우는 영역.

## 3. fake store 보강 — `memoryApplicationStore.UpdateIntegrationBinding`

| 측면 | 변경 전 | 변경 후 |
| --- | --- | --- |
| Provider FK 가드 | 없음 | ProviderID 변경 시 `s.integrationProviders[b.ProviderID]` 존재 검증 → `store.ErrNotFound` |
| Unique 가드 | 없음 | (scope_type, scope_id, provider_id, external_key) 4-tuple 중복 시 `store.ErrConflict` |
| field 갱신 | ScopeID, ExternalKey, UpdatedAt | ScopeID, ProviderID, ExternalKey, Policy, Enabled, UpdatedAt |

production `*store.PostgresStore.UpdateIntegrationBinding` 의 unique index 위반 + FK 가드 mirror — handler 의 `errors.Is(err, store.ErrConflict)` / `store.ErrNotFound` 분기를 의미 있게 test 가능.

기존 test 회귀 zero (`go test ./...` PASS).

## 4. PATCH 5 test 상세

- **Happy** — external_key + policy + enabled 동시 갱신 → 200 + 응답 substring 3건 검증
- **NotFound** — ghost binding_id → 404
- **InvalidPolicy** — `not_a_valid_policy` → 422 + "unsupported policy" message
- **ConflictDuplicate** — 같은 scope_id 의 dup-candidate binding 사전 생성 + PATCH 의 external_key 를 candidate 와 일치 → 4-tuple 충돌 → 409
- **ForbiddenForDeveloperRole** — developer role + Bearer token → 403

## 5. DELETE 3 test 상세

- **Happy** — 200 OK + list 에서 사라짐 검증 + 두 번째 DELETE 가 404 (idempotency 가드)
- **NotFound** — ghost binding_id → 404 + `"binding not found"` error 메시지
- **ForbiddenForDeveloperRole** — developer role + Bearer token → 403

## 6. helper — `seedBindingFixture`

5+ test 가 같은 provider+binding seed 를 반복하던 패턴을 helper 1건으로 통합. fake 의 ID convention `bind-<scope_id>-<external_key>` 활용 (production UUID 와는 다르지만 test 는 fake convention 으로 충분 — handler 가 path `:binding_id` 만 사용).

## 7. 검증

| 항목 | 결과 |
| --- | --- |
| `go test ./internal/httpapi/ -run 'TestUpdateIntegrationBinding\|TestDeleteIntegrationBinding'` | ✅ 8 신규 모두 PASS |
| `go test ./...` | ✅ 14 packages green (기존 test 회귀 zero) |
| frontend | 변경 없음 |

## 8. 잔여 P1/P2/P3 carve (다음 sprint 인계)

| # | 영역 | P |
| --- | --- | --- |
| 1 | Keycloak 실 OIDC e2e flow | P1 |
| 2 | Single-port nginx e2e (ADR-0018) | P1 |
| 3 | frontend service unit (auth/api-client/websocket) | P2 |
| 4 | main flat memory housekeeping (PR #261, #262, #263, 본 PR 머지 후) | P2 |
| 5 | pre-existing ESLint 4 errors refactor | P2 |
| 6 | backend-ai pytest (gRPC server 구현 동반) | P3 |
| 7 | dashboard page snapshot 테스트 | P3 |
