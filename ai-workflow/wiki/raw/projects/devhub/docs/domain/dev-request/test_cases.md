# M5 Dev Request (DREQ) 테스트 케이스 카탈로그

- 문서 목적: DREQ (Dev Request) 도메인의 테스트 범위와 우선순위 TC 카탈로그를 정의해 구현 단계의 품질 기준선 + traceability 매핑을 제공한다.
- 범위: intake auth (API-59 외부 수신) + widget/list/detail flow + promote 단일 트랜잭션 (API-62) + admin token 발급/revoke/PATCH (API-66/67/68/79) + reject/reassign/close (API-63/64/65).
- 대상 독자: Backend/Frontend 개발자, QA, AI 에이전트, 운영자.
- 상태: draft
- 최종 수정일: 2026-06-02
- 관련 문서: [`requirements.md`](../requirements.md) §5.5, [`domain/dev-request/concept.md`](../domain/dev-request/concept.md), [`planning/system_usecases.md`](../planning/system_usecases.md), [`architecture.md`](../architecture.md) §7, [`backend_api_contract.md`](../backend_api_contract.md) §14, [`e2e_testing_strategy.md`](./e2e_testing_strategy.md), [`docs/adr/0012-dreq-external-intake-auth.md`](../adr/0012-dreq-external-intake-auth.md), [`docs/adr/0013-dreq-rbac-row-scoping.md`](../adr/0013-dreq-rbac-row-scoping.md), [`docs/adr/0014-dreq-intake-token-admin.md`](../adr/0014-dreq-intake-token-admin.md), [`docs/adr/0017-dreq-intake-token-operational-hardening.md`](../adr/0017-dreq-intake-token-operational-hardening.md).

## 1. 기능 맵 (REQ/UC 기준)

| 기능 ID | 설명 | REQ | UC |
| --- | --- | --- | --- |
| F-DREQ-INTAKE | 외부 시스템 의뢰 수신 (Bearer + IP allowlist + token expires_at) | REQ-FR-DREQ-001..003 | UC-DREQ-01,02 |
| F-DREQ-LIST | 담당자 본인 의뢰 dashboard widget + 일반 사용자 페이지 + system_admin 관리 페이지 | REQ-FR-DREQ-004,006,008 | UC-DREQ-03,04,07 |
| F-DREQ-PROMOTE | Promote (신규 application/project 생성) — 단일 트랜잭션 + RBAC system_admin | REQ-FR-DREQ-005 | UC-DREQ-05 |
| F-DREQ-LIFECYCLE | reject / reassign / close — system_admin 정책 | REQ-FR-DREQ-007,009,010 | UC-DREQ-06,08,09 |
| F-DREQ-ADMIN-TOKEN | intake token 발급/조회/revoke (plain 1회) + PATCH allowed_ips (API-79) + expires_at | REQ-FR-DREQ-011, REQ-NFR-DREQ-001..003 | UC-DREQ-10 |

## 2. 테스트 계층 전략

1. **단위 테스트 (UT)** — store / handler / middleware 단위. intake auth 의 hash → lookup → IP CIDR → revoke → expired 5단 정합 + promote-tx race 가드 + admin token validation.
2. **통합 테스트 (IT)** — DB + handler 조합. row-level RBAC (ADR-0013 enforceRowOwnership) + promote 트랜잭션 rollback.
3. **E2E 테스트 (TC-DREQ-*)** — 본 카탈로그. UI flow + intake auth round-trip + admin token issuance/revoke 의 운영 시나리오. **본 카탈로그의 TC 는 `frontend/tests/e2e/dev-requests.spec.ts` 가 source-of-truth**.

## 3. 우선 테스트 케이스 (P0/P1)

### 3.1 Happy path (lifecycle 전체)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-DREQ-ADMIN-TOKEN-01` | P0 | E2E | system_admin 이 `/admin/settings/dev-request-tokens` 페이지에서 token 발급 → 2-phase modal (form → reveal) → plain token 1회 노출 → `저장 완료` 클릭 | 201 + plain token 응답 + 목록에 `Active` row 노출 | `dev-requests.spec.ts` step 1 |
| `TC-DREQ-INTAKE-AUTH-01` | P0 | E2E | 발급된 plain token 으로 외부 시스템이 `POST /api/v1/dev-requests` 호출 (Bearer + IP allowlist 통과) | 201 + `data.id` 반환 + dev_requests row 생성 | `dev-requests.spec.ts` step 2 |
| `TC-DREQ-WIDGET-FLOW-01` | P0 | E2E | assignee (developer) 가 `/developer` dashboard 진입 → "내 대기 의뢰" widget 에 의뢰 노출 → click → `/dev-requests` list → row click → DevRequestDetailModal 표시 | widget → list → detail modal 전체 flow 진행 | `dev-requests.spec.ts` step 3 |
| `TC-DREQ-PROMOTE-TX-01` | P0 | E2E | system_admin 이 `/admin/settings/dev-requests` 에서 의뢰 detail modal 진입 → "Register as Application" → platform_id 입력 → confirm | dev_requests row status `registered` + dev_request_target_id 매핑 + audit `dev_request.promoted` | `dev-requests.spec.ts` step 4 |
| `TC-DREQ-ADMIN-TOKEN-REVOKE-01` | P0 | E2E | system_admin 이 token row 의 `revoke` → DestructiveConfirmModal → 확인 | row badge `Revoked` + 동일 token 으로 재호출 시 401 | `dev-requests.spec.ts` step 5 (mega test 의 마지막 단계) |

### 3.2 Token operational hardening (ADR-0017)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-DREQ-ADMIN-TOKEN-PATCH-01` | P1 | E2E | system_admin 이 `PATCH /api/v1/dev-request-tokens/:token_id` 로 allowed_ips 갱신 | 200 + `allowed_ips` 변경 응답 + audit `dev_request_intake_token.updated` | `dev-requests.spec.ts` (신규 test) |
| `TC-DREQ-ADMIN-TOKEN-PATCH-NEG-01` | P2 | UT/IT | 빈 `allowed_ips` 로 PATCH | 400 `invalid_allowed_ips` | unit test 영역 |
| `TC-DREQ-INTAKE-AUTH-EXPIRED-01` | P1 | UT/IT | `expires_at < NOW()` 인 token 으로 intake 호출 | 401 `auth_intake_token_expired` + audit `dev_request_intake.expired` | unit test (E2E 는 시간 의존으로 부적합) |

### 3.3 Negative path (intake auth)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-DREQ-INTAKE-AUTH-NEG-01` | P0 | E2E | random/invalid bearer 로 `POST /api/v1/dev-requests` | 401 (token hash 미발견) | `dev-requests.spec.ts` (신규 test) |
| `TC-DREQ-INTAKE-AUTH-NEG-02` | P1 | UT/IT | 정상 token + `allowed_ips` 외부 IP 호출 | 401 (CIDR 불일치) | unit test 영역 |
| `TC-DREQ-INTAKE-AUTH-NEG-03` | P0 | E2E | revoke 된 token 으로 intake 호출 | 401 (revoked_at IS NOT NULL) | `dev-requests.spec.ts` mega test 마지막 부분이 cover |

### 3.4 RBAC negative (ADR-0013)

| TC ID | 우선순위 | 계층 | 시나리오 | 기대 결과 | spec ts 위치 |
| --- | --- | --- | --- | --- | --- |
| `TC-DREQ-RBAC-NEG-01` | P1 | UT/IT | non-system_admin 이 `POST /api/v1/dev-request-tokens` 호출 | 403 `auth_role_denied` | unit test (RBAC route gate) |
| `TC-DREQ-RBAC-NEG-02` | P1 | UT/IT | non-system_admin 이 reassign / close 호출 | 403 | unit test |
| `TC-DREQ-RBAC-ROW-01` | P1 | UT/IT | team_manager 가 본인 assignee 의뢰만 promote 가능, 타인 의뢰 시도 → 403 | row-level scoping (enforceRowOwnership) | unit test |

## 4. 카버리지 매핑

| TC ID | spec ts test 명 | 상태 |
| --- | --- | --- |
| `TC-DREQ-ADMIN-TOKEN-01` | `Intake to Promote to Revoke lifecycle` (step 1) | ✅ active (PR #136 mega test) |
| `TC-DREQ-INTAKE-AUTH-01` | `Intake to Promote to Revoke lifecycle` (step 2) | ✅ active (PR #136) |
| `TC-DREQ-WIDGET-FLOW-01` | `Intake to Promote to Revoke lifecycle` (step 3) | ✅ active (PR #136) |
| `TC-DREQ-PROMOTE-TX-01` | `Intake to Promote to Revoke lifecycle` (step 4) | ✅ active (PR #136) |
| `TC-DREQ-ADMIN-TOKEN-REVOKE-01` | `Intake to Promote to Revoke lifecycle` (step 5) | ✅ active (PR #136) |
| `TC-DREQ-PROMOTE-PROJ-01` | `TC-DREQ-PROMOTE-PROJ-01 — Intake to Promote to Project lifecycle` | ✅ active (PR #323, cleanup warning 제거 2026-06-02) |
| `TC-DREQ-INTAKE-AUTH-NEG-03` (revoked) | `Intake to Promote to Revoke lifecycle` (revoke 후 재호출 fail) | ✅ active (PR #136) |
| `TC-DREQ-INTAKE-AUTH-NEG-01` (invalid bearer) | `Invalid bearer is rejected` | 🟢 신규 (본 sprint `claude/work_260518-d`) |
| `TC-DREQ-ADMIN-TOKEN-PATCH-01` | `PATCH allowed_ips updates token` | 🟢 신규 (본 sprint) |
| `TC-DREQ-INTAKE-AUTH-EXPIRED-01` | (UT only — `dev_request_intake_auth_test.go`) | ✅ active (PR #137) |
| `TC-DREQ-ADMIN-TOKEN-PATCH-NEG-01` | (UT only — admin handler) | ✅ active (PR #137) |
| `TC-DREQ-RBAC-NEG-01..02` | (UT only — RBAC route gate) | ✅ active (sprint `i` + `m`) |
| `TC-DREQ-RBAC-ROW-01` | (UT only — enforceRowOwnership) | ✅ active (sprint `m` ADR-0013) |
| `TC-DREQ-INTAKE-AUTH-NEG-02` (CIDR 외부) | (UT only) | ✅ active (sprint `i` UT-dreq-intake_auth-XX) |

## 5. 운영 메모

- mega test 는 의도적으로 단일 케이스 + `test.step()` 4 단계로 운영. 각 step 은 TC ID 와 1:1 매핑되며 step 실패 시 보고서에서 어느 단계인지 즉시 식별.
- intake auth path 의 시간 의존 (expires_at) E2E 는 부적합 — UT 가 결정적. backend `requireIntakeToken` middleware test (`dev_request_intake_auth_test.go`) 가 `expires_at` 과거/미래 4 case cover.
- IP allowlist 의 외부 CIDR negative 는 CI runner 의 outbound IP 제어가 어려워 UT 영역. E2E 에서는 `0.0.0.0/0` 으로 우회.
- DestructiveConfirmModal (PR #140) 의 revoke 확인 흐름은 `TC-DREQ-ADMIN-TOKEN-REVOKE-01` 의 일부로 cover.

## 6. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-18 | 1차 draft — TC-DREQ-* 13건 정식 발급. mega test (PR #136) + 신규 2 test (본 sprint) + UT 영역 6 매핑. | sprint `claude/work_260518-d` |
| 2026-06-02 | `TC-DREQ-PROMOTE-PROJ-01` 카탈로그 매핑 보강 + cleanup fetch 경로 정규화. | `dev-requests.spec.ts`의 best-effort cleanup warning 제거, traceability matrix sync |
