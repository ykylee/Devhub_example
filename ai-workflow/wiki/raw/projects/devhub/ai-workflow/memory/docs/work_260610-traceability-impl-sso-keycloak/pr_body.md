# docs(traceability,adr,memory): v1.1 sprint -a follow-up PR1 (PR #540) 의 carry-over C-g + C-h 정합 — IMPL 5 row + ADR-0030 §5 timeline accepted/done

## 목적

v1.1 sprint -a follow-up PR1 (PR #540, `feat/work_260610-v1-1-sprint-a-real-adapter`, main HEAD `58d163f`) 의 carry-over **C-g + C-h** 의 정공법 PR. **문서만 변경 (코드 0줄)**.

- C-g (P2): `docs/traceability/report.md` IMPL-sso-keycloak-* + auth-session-port-01 + audit-ops-event-mirror-01 row 갱신
- C-h (P2): `docs/adr/0030-sso-integrations-and-auth-session-port.md` §5 timeline 갱신 (1.1a + 1.1b status = accepted/done, C-h row done)

sprint -a follow-up PR1 PR #540 의 body 가 "추가 PR" carry-over 로 명시한 본 PR.

## 변경 요약

### 1. `docs/traceability/report.md` (7 변경)

- **§2.4 IMPL 개요 paragraph** (L224) — `sso-keycloak-01` + `sso-keycloak-stub-01` + `sso-keycloak-metrics-01` + `auth-session-port-01` + `audit-ops-event-mirror-01` 5 row mention.
- **§2.4 새 sub-table** (IMPL-audit-XX 다음 위치) — 5 row 신규 + sub-table header + intro paragraph.
- **§3.1 auth-session row** (L406) IMPL 컬럼에 `auth-session-port-01` cross-ref 추가.
- **§3.1 audit-ops row** (L407) IMPL 컬럼에 `audit-ops-event-mirror-01` cross-ref 추가.
- **§3.3 keycloak-idp row** (L427) — 7 column 전체 갱신: ARCH/API 컬럼에 ADR-0030 link + SOP list 보강 / ROADMAP 컬럼에 ADR-0030 §5 timeline 1.1a + 1.1b accepted/done / IMPL 컬럼에 sso-keycloak-* 3 row + cross-ref auth-session-port-01 + audit-ops-event-mirror-01 / UT 컬럼에 sprint -a follow-up PR1 의 신규 4 test file mention.
- **§4 ADR 인덱스** (L476) ADR-0030 row 신규.
- **§6 변경 이력** (L525) 본 PR row 신규 (2026-06-10).

### 2. `docs/adr/0030-sso-integrations-and-auth-session-port.md` (3 변경)

- **§5 결정 timeline** (L110-115) — 1.1a + 1.1b status = `accepted (P1), done (PR #538 / #539 / #540 머지, 2026-06-10)`.
- **§5 C-h row** (L116) 신규 — 본 PR 의 carry-over 정공법 (`docs/work_260610-traceability-impl-sso-keycloak` PR) status = `done (2026-06-10)`.
- **§9 변경 이력** (L161) row 추가.

### 3. Memory sync (4 file 신규 + 3 file 갱신)

- **sprint memory 신규**: `ai-workflow/memory/docs/work_260610-traceability-impl-sso-keycloak/{state.json,session_handoff.md,work_backlog.md,backlog/2026-06-10.md,pr_body.md}` — 본 PR 의 branch 별 memory 디렉터리 4종.
- **main flat memory sync**: `ai-workflow/memory/{state.json,session_handoff.md,work_backlog.md}` — head_commit = `58d163f` (PR #540 머지 baseline) + status 갱신 + §6 변경 이력 row.

## 추적성 영향

| Stage | ID | Status | 비고 |
| --- | --- | --- | --- |
| IMPL | `IMPL-sso-keycloak-01` | **신규** | real adapter — `backend-core/internal/sso-integrations/keycloak/verifier.go` (`KeycloakJWKSVerifier`) + `admin_client.go` (`KeycloakAdminClient` 3 port 동시 충족) |
| IMPL | `IMPL-sso-keycloak-stub-01` | **신규** | 사외 build stub — `sso-integrations/keycloak/saovae_stub.go` (4 port + webhook handler) |
| IMPL | `IMPL-sso-keycloak-metrics-01` | **신규** | JWKS stale-while-error metric — `sso-integrations/keycloak/metrics.go` |
| IMPL | `IMPL-auth-session-port-01` | **신규** | canonical port interface — `domain/auth-session/integration/ports.go` + view/ deprecation + struct 직접 정의 |
| IMPL | `IMPL-audit-ops-event-mirror-01` | **신규** | mirror 통합 — `audit-ops/service/keycloak_event_puller.go` + `KeycloakEventLister` interface 통폐합 |

**신규 ID 5건 (모두 IMPL, REQ/UC/ARCH/API/RM/UT/TC 신규 발급 0건)** — sprint -a follow-up PR1 PR #540 코드 정합의 문서 cell fill.

## 매트릭스 영향 (cross-ref)

- §3.1 **auth-session** row: `auth-session-port-01` cross-ref 추가
- §3.1 **audit-ops** row: `audit-ops-event-mirror-01` cross-ref 추가
- §3.3 **keycloak-idp** row: 7 column 갱신 (ARCH/API + ROADMAP + IMPL + UT 컬럼 4 변경)
- §4 **ADR 인덱스**: ADR-0030 row 신규
- §6 **변경 이력**: 본 PR row 신규 (2026-06-10)
- ADR-0030 **§5 timeline**: 1.1a + 1.1b accepted/done + C-h row done
- ADR-0030 **§9 변경 이력**: row 추가

## ID 형식 결정 (정책)

- **kebab-case module ID** 채택 (`conventions.md §1` 정합) — `IMPL-sso-keycloak-{01, stub-01, metrics-01}` + `IMPL-auth-session-port-01` + `IMPL-audit-ops-event-mirror-01` (5 row)
- **메모리 출발점 정정**: sprint -a follow-up PR1 의 carry-over 표기 `IMPL-30/31/32` (단순 정수) 는 `conventions.md §1` 위반 → kebab-case module ID 로 정정
- 5 row 분리 (vs coarse 1 row 통합) 채택 — architecture boundary 별 1 row (real adapter / stub / metrics / port / mirror), 코드 위치 + 책임 cell 단위 자연 매핑

## Tier 분류 (self-check)

| 변경 영역 | Tier | 근거 |
| --- | --- | --- |
| `docs/traceability/report.md` | **공용** | 문서 정합, 사내 한정 정보 미포함 |
| `docs/adr/0030-...` | **공용** | ADR 본문 갱신, 사내 한정 정보 미포함 |
| `ai-workflow/memory/...` | **공용** | workflow metadata, 사내 한정 정보 미포함 |

**본 PR 의 모든 변경 = 공용** — `check-tier-separation.sh` 의 `no changes between origin/main and HEAD` 메시지로 정합 확인.

## 검증 (run on this branch)

- `bash scripts/check-tier-separation.sh` — ✅ no changes between origin/main and HEAD
- `bash scripts/check-openapi-yaml-lint.sh` — ✅ passed (openapi.yaml 변경 0, paths=81, schemas=78)
- `bash scripts/check-migration-uniqueness.sh` — ✅ All migration prefixes are valid and unique
- `python3.13 ai-workflow/tests/check_docs.py` — 본 PR 의 4 file 정합 (metadata 6 field + cross-link + 제목 헤더). exit 1 의 원인은 본 PR 영역 외 기존 file 의 historical link (sprint 분업 전면 취소 이전의 stale link)

## Out of scope (별도 PR carry-over)

- **C-i (P2)**: E2E saovae_stub + real adapter CI matrix — DEVHUB_BUILD_TIER=internal env var + e2e shard 양쪽 정합
- **C-j (P3)**: build tag 정책 재검토 — runtime injection (현재) ↔ build tag 전환 trade-off

## Refs

- Sprint -a 본 PR #538 (port interface) — main `20b4bb3`
- Sprint -a follow-up 본 PR #539 (saovae_stub + main wiring) — main `87e6c1f5`
- Sprint -a follow-up PR1 PR #540 (real adapter + v1.0 mirror struct 제거) — main `58d163f` (이 PR 의 carry-over source)
- [ADR-0030 §2.1 port interface canonical 위치 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `domain/auth-session/integration/`
- [ADR-0030 §2.2 real adapter 분리 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `sso-integrations/keycloak/`
- [ADR-0030 §2.3 runtime injection 결정](../adr/0030-sso-integrations-and-auth-session-port.md) — option 2 = `DEVHUB_BUILD_TIER` env var, build tag 미사용
- [release_v1_roadmap.md §3.5 N-13](../planning/release_v1_roadmap.md) — v1.1 sprint -a follow-up carry-over N-13

## Base / target

- **base branch**: `main` (HEAD `58d163f`, PR #540 머지 후)
- **target branch**: `main`
- **merge strategy**: squash
- **branch name**: `docs/work_260610-traceability-impl-sso-keycloak`
