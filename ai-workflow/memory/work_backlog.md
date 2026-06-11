# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: **2026-06-11 update — main HEAD `9f95e8bd` (PR #545 squash `ac52d1ca`, 2026-06-10 22:38 UTC) + PR #546 (N-10 housekeeping, mergeCommit `9ed8c25`) + PR #547 (N-13 housekeeping, mergeCommit `44df9883`) + PR #549 (T-d-72-2 metadata, mergeCommit `43feccfb`) + PR #550 (E2E spec timing, mergeCommit `9f95e8bd`) 머지. 06-09~06-11 사이 PR #514~#550 (36 PR) 정합. T-d-72-2 (D-72 Phase 1 wiki mirror) re-sync 완료 (2026-06-11 01:45:04Z, 83 file, 1.6M). wiki-ingest-from-raw skill (D-72 Phase 3) **본 저장소 측 wrapper** 작성 (`scripts/wiki-ingest-from-raw.sh`, my_harness 측 skill 호출). dry-run PASS: 83 source 식별, 0 errors. 사용자 trigger 시 `--apply` 로 실제 ingest. N-10 P1 follow-up partial verified (E2E 4 TC PR #509 + backend IT 3 TC PR #515 + status ⏳ per PR #512). N-13 ID slot 9 row 발급 (PR #547, REQ-FR-113/UC-DEV-REQ-15/ARCH-23/API-103/RM-DEV-REQ-15/IMPL-inbound-source-01/IMPL-platform-patch-02/UT-inbound-source-01/TC-INBOUND-SRC-01, 모두 planned, v1.1 진입 시점 코드 변경). N-13 backend foundation (PR #548) OPEN. v1.0 출시 직전, 자유 에이전트 정책. 다음 sprint: N-6 staging 1주 운영 (사용자 결정) 외 housekeeping.**
- 최종 수정일: 2026-06-11 (T-d-72-2 re-sync + E2E spec timing PR #550 + wiki-ingest-from-raw skill wrapper + 메모리 drift 정합)
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md), [ADR-0025](../../docs/adr/0025-envelope-encryption-key-management.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | ✅ done | 2026-05-11 | PR-B+C (#56) + PR-D (#57) 머지로 envelope / types / WS / CommandStatus / audit actor enrichment / request_id 완결. PR-D 후속 (#80/#82) 으로 commands enrichment + DEVHUB_TRUSTED_PROXIES 보강. |
| **M2** — 사용자 경험 정합 | ✅ done (1차 완성) | 2026-05-12 | login_action + work_26_05_11 + work_26_05_11-d 완료 후 PR #85 (`claude/login_usermanagement_finish`) 가 1차 완성 sprint 로 닫음: 로드맵 정합 + UX hygiene (PR-UX1+2+3) + Kratos audit (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| **CI** — GitHub Actions 도입 | ✅ done (1차) | 2026-05-13 | PR #86 (`gemini/prepare-github-action`) 가 backend-unit + frontend-unit + e2e (Playwright 40 TC) 3잡 도입. 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개로 그린 도달. PR-T5 의 핵심 잡 묶음은 본 PR 으로 1차 완료. 후속: FU-CI-1 (no-docker policy 정합), FU-CI-2/3/4 는 `claude/work_260513-a` 처리 중. |
| **M3** — Realtime 확장 + 외부 연동 1차 | ✅ done (1차) | 2026-05-13 | RM-M3-01..03 (Sign Up + 인사 DB + 조직 polish) 완료. ADR-0008/0009/0010 신규. |
| **Application Domain (backend 1차)** | ✅ done | 2026-05-14 | API-01~58 전체 activated (본 세션 #104~#110). 마이그레이션 000012~000018 (7), ADR-0011 accepted, RBAC 4 신규 resource (system_admin 일임), CI 5 job (Backend Integration 신설). 23 integration test (P1/P2 회귀 guard 포함) 가 CI 에서 실 실행. |
| **DREQ Domain (closing 1차 + TC-DREQ-* 13건)** | ✅ closing | 2026-05-18 | API-59..68 + API-79 (PATCH allowed_ips) activated + ADR-0012/0013/0014/**0017**. P2 carve out 6/6 모두 해소 + codex hotfix #5 (store revoked guard + 회귀 가드 3 unit test). **TC-DREQ-* 13건 정식 발급** (sprint d) + `dev-requests.spec.ts` 6 step + `test_cases_m5_dreq.md` 신규. 잔여 carve out: ADR-0017 §6 atomicity 실제 구현 (CTE refactor) + 자동 cron revoke + last_used staleness alert. |
| **External Integration Domain (1차 종합 closing)** | ✅ closing | 2026-05-18 | concept staged (PR #135) → backend 1차 (PR #139) → **ADR-0015/0016/0017** (sprint c) → provider frontend (sprint g/h) + API-80 DELETE (sprint j) → **bindings 관리 UI** (sprint m, PR #154) + **topology v2 시각화** (sprint n, PR #155) + **ADR-0017 §6 atomicity** 실 구현 (sprint o, PR #156) + **ADR-0015 §6 (1)+(2)** (sprint p, PR #157). API-69..80 모두 activated. TC-INT-FRONTEND-* 10건 + TC-INT-FRONTEND-BIND-* 3건 + TC-INT-HOMELAB-03 + TC-INT-FRONTEND-TOPOLOGY-V2-* 2건. ADR carve out 추가 종결 (atomicity + size limit + token rotation SOP). 잔여 carve: ADR-0016 §6 (Alertmanager YAML / Grafana JSON / p95 alert / push 알림 / 임계), ADR-0017 §6 잔여 (cron revoke / PATCH expires_at / 만료 alert), ADR-0015 §6 (3)+(4) (dedicated worker / push-pull dedup), React Flow group sub-node + WebSocket 실시간. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. RM-M4-01..09 (9 항목). |

## 2. 최근 머지 (M1 RBAC track, 2026-05-08)

| PR | 제목 | 머지 |
| --- | --- | --- |
| #20 | feat(http): SEC-5 mask 5xx errors + M1 sprint plan (PR-A) | `ae8aca1` |
| #21 | docs(adr): ADR-0002 RBAC policy edit API (PR-F) | `950a11f` |
| #22 | docs(api): RBAC §12 rewrite + route mapping (PR-G1) | `1a090a3` |
| #23 | feat(rbac): domain + rbac_policies migration (PR-G2) | `5239a87` |
| #29 | feat(rbac): postgres store + users.role FK (PR-G3 + FIX-A) | `24815b8` |
| #30 | feat(rbac): RBAC handlers (PR-G4 + FIX-B + FIX-C) | `27b6817` |
| #31 | feat(rbac): permission cache + enforcement (PR-G5) | `02eef35` |
| #27 | feat(frontend): RBAC PermissionEditor ↔ backend (PR-G6 + FIX-D) | `e02ba67` |
| #28 | docs(memory): M1 PR review actions tracker | `9bc30c9` |

원본 PR #24, #25, #26 은 stack base 자동 삭제로 close 후 main 위에서 #29/#30/#31 로 재등록.

## 3. M1 잔여 + DEFER 후속 작업

### 3.1 M1 잔여 (P1, 진입 시 분해)

- **T-M1-02 (PR-B)** — API envelope/role wire/필드 정합 (`backend_api_contract.md` ↔ 코드 1:1 강제, 통합 테스트 매트릭스).
- **T-M1-03 (PR-C)** — command lifecycle 6 상태 일관 적용 + dry-run/live 경계 테스트.
- **T-M1-04 (PR-D)** — Audit actor 보강 (`source_ip`, `request_id`, `source_type` + 마이그레이션 + 미들웨어 + 응답 헤더 `X-Request-ID`).
- **T-M1-05** — `auth_test.go` prod 가드 + role 가드 통합 테스트 매트릭스. PR-C 또는 PR-D 에 흡수 가능.
- **T-M1-07 (frontend)** — `frontend/lib/services/types.ts` UI vs wire 분리, 표시 포맷 프론트 이전. PR-B 에 묶거나 단독.
- **T-M1-08** — WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합. PR-B 에 묶거나 단독.

### 3.2 DEFER (M1 PR 리뷰 후속, [상세](./M1-PR-review-actions.md#3-다음-개발로-넘김--defer))

- **M1-DEFER-A** — `rbac_policies` 의 `is_system` ↔ `role_id` 일관성 CHECK 제약 (P2 방어선)
- **M1-DEFER-B** — `requireMinRole`/`roleMeetsMin`/`roleRank` deadcode 정리 + 단위 테스트 제거
- **M1-DEFER-C** — `writeRBACServerError` 임시 helper → `writeServerError` 통합 (PR-G4 의 TODO)
- **M1-DEFER-D** — `DeleteRBACRole` row-lock (다중 인스턴스 race 강화)
- **M1-DEFER-E** — `PermissionCache` 다중 인스턴스 일관성 (pub/sub 또는 polling)
- **M1-DEFER-F** — API contract §12.4 / §12.5 응답 예시 추가
- **M1-DEFER-G** — `MemberTable` role display 회귀 사용자 환경 검증

### 3.2 P1~P2 should-have

- **frontend `/auth/callback`**: Hydra authorization_code → `/oauth2/token` 교환 후 세션 저장 흐름. PR-D 의 자연 후속.
- **frontend `account.service.ts`**: Kratos public flow 호출 helper.
- **frontend `types.ts` 분리**: UI 표시명 vs API wire 타입 분리, 표시 포맷팅을 프론트로 이전.
- **WebSocket envelope 표준화**: `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합.
- **RBAC policy 편집 API**: 12.x — write/audit 경계 + persistence 또는 *static-default 유지* 결정.

### 3.3 P3 nice-to-have

- ADR-0002 (Gitea SSO 통합) 작성.
- ADR (commits 정규화 테이블 도입 여부).
- ADR (OS service wrapper 결정).

## 4. 운영 / 환경 메모

- Hydra/Kratos PoC: 본 세션에서 e2e 검증 완료. binary 위치 `$env:USERPROFILE\go\bin\` (Windows). 다음 세션에서 재가동은 `hydra serve all --config infra/idp/hydra.yaml --dev` + `kratos serve --config infra/idp/kratos.yaml`.
- 검증용 임시 OIDC client (client_credentials, id `43aa4b74-...`) 는 Hydra 안에 잔존. 보안 위험은 없으나 cleanup 가능.
- backend-core `go test ./...` 는 사내 GoProxy mirror 환경에서 PASS 검증됨.

## 5. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-09 | **N-8 sign-out e2e deterministic race hotfix 4차 (issue #501)** — 3 commit / 2 PR (PR #502 + #503, 4 신규 TC). **PR #502** (`fix/work_260609-a-N8-logout-graceful-degrade`): backend logout handler 502 → 204 graceful degradation + audit `revoke_status=unreachable` + hotfix 식별자. **PR #503** (`fix/work_260609-b-N8-codex-p1-distinguishable`, 2 commit): (1) codex P1 응답 — response header `X-Keycloak-Likely-Down: true` + `X-Logout-Hotfix` marker; (2) codex P1 follow-up 응답 — typed error sentinel `authview.ErrOIDCConfigMissing` (config error, marker 미부착) + `authview.ErrOIDCNetworkUnreachable` (네트워크/5xx, marker 부착). **검증**: e2e shard 1..3 PASS (N-11 잔여 DoD 해소), backend 35 packages PASS, frontend 80 files / 1033 tests PASS (FE-08 신규 +1). **신규 ID 없음** (N-8 follow-up hotfix). main HEAD `897953c`. issue #501 close. |
| 2026-06-09 | **워커 분업 전면 취소 (사용자 결정, PR #500)** — Claude/Codex 자유 이용 불가. `worker_division.md` §0 + §1~§4 historical 標記 + §2.5 branch prefix 자유화 + §5 Owner 권한 명시. `AGENTS.md` 워커 일반 메모 갱신 + 워커별 전용 메모 4종 historical 標記. 유지 정책: §4.2 ADR reversal supersession 정공법 + §5 Owner 권한 + 우선순위 P0~P3 강제. main `f99fef7`. |
| 2026-06-09 | **N-11 메모리 sync (PR #499)** — 5 파일 (+36/-15) main `da7d57e` 머지. 충돌 해소 (release_v1_roadmap.md N-11 row + 워커 분업 row 둘 다 보존). release_v1_roadmap.md 의 PR #500 row 는 §9 변경 이력에 별도 정합. 신규 ID 없음 (cross-cutting infrastructure 운영 정합). |
| 2026-06-08 | **N-11 CI e2e + backend-integration 복원 운영 정합 (sprint 260608-a, issue #419)** — `&& false` 2건은 PR #407 cleanup-recovery 후속 4 squash merge (4a1942e / 5f5fdba / 9395cd9 / ce8ce7c) 로 코드 레벨 복원 완료. 본 sprint 의 1차 PR #498 은 ci.yml 코멘트만 갱신 (코드 변경 0줄), 2차 PR (메모리 4종 + traceability report.md §6 + release_v1_roadmap §3.5/§4.1/§9) 정합. **잔여 DoD**: main 첫 PR 에서 두 job 실 실행 PASS (state.json head + §3 잔여 표 N-11 잔여 row 추가). 신규 ID 없음 (cross-cutting infrastructure 운영 정합). |
| 2026-06-06 | **sprint -h 신규 carve 의 ID 발급 + 매트릭스 cross-ref** — §3.1 의 auth-session / integration-registry / repository-integration row 3 row cross-ref 갱신. **신규 ID**: REQ-FR-106/107/108 + ARCH-18/19/20 + API-98/99/100 + IMPL/UT/TC 관련 ID 발급. Codex 리뷰에 따라 `docs/traceability/report.md`의 `integration-registry` 도메인에 `IMPL-ci-runs-01`, `UT-ci-runs-01`, `TC-CI-RUN-01` 추가 보완 완료. PR #490 머지 완료. |
| 2026-06-02 | 중간 개발 보고 자료 준비 착수. `docs/presentations/2026-06-02-midterm-report-plan.md` 에 슬라이드 구조/데이터 수집 축/디자인 방향을 정리했고, `docs/analysis/2026-06-02-midterm-report-baseline.md` 에 현재 기능 범위, SDLC/추적성, 테스트, AI agent 활용 현황, 활동 통계를 베이스라인으로 기록. 다음 단계는 HTML/CSS/JS 슬라이드 초안 구현. |
| 2026-06-01 | CI 회귀 복구: (1) `frontend/app/(dashboard)/applications/[id]/page.tsx` 중복 import 제거로 `Build App` 타입 에러 해소, (2) `frontend/tests/e2e/admin-projects.spec.ts` TC-PROJ-UI-04를 환경 독립 검증(ComboBox/input 공용)으로 보강. 로컬 `npm run test`/`npm run build` 통과, CI run `26738464130` 성공. |
| 2026-06-10 | **v1.0 출시 직전 housekeeping + 메모리 비대한 문서 분리/간소화** (sprint `maintenance/work_260610-b-v1-pre-release-housekeeping`) — (1) `release_v1_roadmap.md` §3.5 N-8 race row status `✅ resolved` + N-11 row "잔여 DoD" → "✅ 해소" 명시 + §4.1 sprint -k 의 N-11 잔여 DoD 완료 마킹 + §3.5 N-10 follow-up 6 TC 명시 + §9 변경 이력 1 row. (2) `docs/traceability/report.md` §4 ADR-0025 (envelope encryption) + ADR-0026 (Keycloak role excluded) row 보강 + §6 변경 이력 1 row. (3) **메모리 비대한 문서 분리**: `ai-workflow/memory/state.json` 1515 → 150 line (90% 감소). 21 top-level key (merged_prs_2026_05_11~2026_05_20 + 5 progress 객체 + 1 milestone 객체) + next_actions 5 closed key → `_archive/state-2026-05-pr-tracker/` (state.json 85KB + next_actions_closed.json 3.5KB + work_backlog_2026-05_archive.md 80KB). 본 work_backlog.md §5 의 5월 135행 archive summary 1 line 으로 축약. 5월 136건 변경 이력의 canonical source 는 `git log --since=2026-05-01 --until=2026-06-01 --merges`. (4) PR 머지 후 main flat memory finalize. **신규 ID 발급 없음** (governance housekeeping). main HEAD `ad8d481` (PR #505). |
| 2026-06-10 | **v1.1 sprint -a follow-up PR #539** (sprint `feat/work_260610-v1-1-sprint-a-followup`) — sprint -a 본 PR #538 (port interface) 의 후속. (1) `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (NEW, 105 lines) — 4 port stub (BearerTokenVerifier + IdentityAdmin + OIDCLogoutClient + KeycloakEventPort) + webhook handler. 사외 build 시 default wiring, Keycloak 인프라 의존성 0. (2) `backend-core/main.go` — `DEVHUB_BUILD_TIER` env var 분기 추가. default (사외) = saovae_stub; `=internal` 시 real KeycloakAdminClient (현 path). event listener 의 type assertion (`*httpapi.KeycloakAdminClient`) 정리는 다음 PR (real adapter 이전) 에서. (3) `backend-core/internal/domain/auth-session/integration/ports.go` — `KeycloakUserEvent` / `KeycloakAdminEvent` mirror struct → `type X = httpapi.X` alias. `*KeycloakAdminClient` 가 `KeycloakEventPort` 충족 위해 필수. v1.0 의 mirror struct 자체 제거는 별도 PR. (4) `backend-core/internal/domain/auth-session/view/{auth,handler}.go` — `BearerTokenVerifier` + `IdentityAdmin` + `OIDCLogoutClient` 3 interface deprecation comment. canonical = `integration/`. **Tier**: ports.go / view/ = 공용, sso-integrations/keycloak/ = 사외 (saovae_stub), main.go 변경 = 공용 (runtime injection branch 만, 사내 한정 패턴 미도입). **CI 7/7 PASS** (Backend Integration 1m15s, Backend Unit 1m8s, E2E Build 1m49s, E2E Playwright 1/3 4m0s, 2/3 4m2s, 3/3 5m30s, Detect Changed Paths 11s, Migration Prefix 7s, OpenAPI Lint 9s, Workflow Lint 13s, Frontend Unit skip). **commit `a00793bc`** (5 file, +238 -97). **PR #539 squash merge** → main HEAD `87e6c1f5` (branch delete). **신규 ID**: IMPL-30 (saovae_stub), IMPL-31 (main.go runtime injection), IMPL-32 (view/ deprecation + ports.go alias) — `docs/traceability/report.md` 갱신은 후속 PR (C-g). |
| 2026-06-11 | **잔여 follow-up sprint (`chore/work_260611-b-residual-housekeeping-close`)** — N-10 housekeeping close (status ⏳ verified (partial) → ✅ resolved (full)) + lint 8 errors / 62 warns follow-up 사실상 close (D-74 L03 skip patch + skip config 적용, 3개 lint report 모두 0 error / 0 warn / 0 info 정합) + N-13 housekeeping 정공법 (status ⏳ planned 유지, 구현은 별도 sprint) + memory 4 file 갱신. 4 file 변경 (`docs/planning/release_v1_roadmap.md` §3.5 N-10 row close + `docs/validation/N-10-manager-rbac.md` §0/§3.1/§6 갱신 + `state.json` M-v1.0 notes 갱신 + `work_backlog.md` 변경 이력 row 추가). 신규 ID 없음 (housekeeping 정공법). |
