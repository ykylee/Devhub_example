# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: M1·M2 done. GHA CI 그린 + FU-CI-1..4 모두 처리 (PR #86/#87/#88). 거버넌스 + 추적성 체계 + 1차 종합 매트릭스 도입 (PR #89). 진행 중 sprint `claude/work_260513-d` 가 매트릭스 §5 gap + 문서 메타 헤더 표준화 처리 중.
- 최종 수정일: 2026-05-13
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | ✅ done | 2026-05-11 | PR-B+C (#56) + PR-D (#57) 머지로 envelope / types / WS / CommandStatus / audit actor enrichment / request_id 완결. PR-D 후속 (#80/#82) 으로 commands enrichment + DEVHUB_TRUSTED_PROXIES 보강. |
| **M2** — 사용자 경험 정합 | ✅ done (1차 완성) | 2026-05-12 | login_action + work_26_05_11 + work_26_05_11-d 완료 후 PR #85 (`claude/login_usermanagement_finish`) 가 1차 완성 sprint 로 닫음: 로드맵 정합 + UX hygiene (PR-UX1+2+3) + Kratos audit (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| **CI** — GitHub Actions 도입 | ✅ done (1차) | 2026-05-13 | PR #86 (`gemini/prepare-github-action`) 가 backend-unit + frontend-unit + e2e (Playwright 40 TC) 3잡 도입. 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개로 그린 도달. PR-T5 의 핵심 잡 묶음은 본 PR 으로 1차 완료. 후속: FU-CI-1 (no-docker policy 정합), FU-CI-2/3/4 는 `claude/work_260513-a` 처리 중. |
| **M3** — Realtime 확장 + 외부 연동 1차 | planned | — | 통합 로드맵 §3.4. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. ADR-0002 (Gitea SSO) 등. |

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
| --- | --- |
| 2026-05-07 | PR #12 통합 sprint 종료 (BLK/SEC/HYG 트래커 등록) |
| 2026-05-08 | PR #13~#19 머지. 통합 로드맵 채택 + M0 sprint 종료. flat memory 갱신. |
| 2026-05-08 | M1 sprint 진입. `claude/m1-sprint-plan` 브랜치 + backlog 초안. 진입 PR 5건 분할 결정. |
| 2026-05-08 | M1 RBAC track 8 PR (#20·21·22·23·29·30·31·27) + 리뷰 트래커 (#28) 머지. FIX A~D 적용. DEFER A~G 백로그 인계. |
| 2026-05-08 | M2 login_action sprint 진입. 로그인 검토 후 PR-LOGIN-1 (#33 backend proxy) + PR-LOGIN-2 (#34 frontend form) push 머지 대기. backlog 작성 (`claude/login_action/backlog/2026-05-08.md`). PR-LOGIN-3·4 미진입. |
| 2026-05-12 | 4 sprint × 8 PR 머지 — E2E 자동 시드 hardening (#76, #78) + PR-D audit 정합 완결 (#80) + second-pass review fix (#82). work_260512-h 폐기 (kratos_identity_id 가 PR-L4 에 이미 포함). 자세히는 `session_handoff.md` §0/§5 + `state.json` 의 `merged_prs_2026_05_12`. |
| 2026-05-12 | PR #85 (`claude/login_usermanagement_finish`) — M2 1차 완성 sprint. 로드맵 정합 + UX hygiene 묶음 (PR-UX1+2+3) + Kratos settings/password/after webhook → audit_logs (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| 2026-05-13 | PR #86 (`gemini/prepare-github-action`) — GitHub Actions CI 그린. 리뷰어 모드 2-pass 에서 5 blocker (idp-apply-schemas DSN, build path, DB_URL env, IdP URL 4종, kratos cipher 32 bytes) + follow-on 5 (Ory v26.2.0, tar layout, GOBIN, YAML interpolation, TC-NAV-02 race) → 보강 commit 7개. |
| 2026-05-13 | PR #87 (`claude/work_260513-a`) — FU-CI-2/3/4 처리: playwright install scope (chromium), `actions/cache@v4` (go mod + ms-playwright), frontend readiness 120s, install split. 추가로 리뷰어 모드 2-pass 에서 cache 효율 fix. |
| 2026-05-13 | PR #88 (`claude/work_260513-b`) — ADR-0003 (no-docker 정책 CI 범위 명문화). `services: postgres:15` 제거 + pgdg native PG 15. 두 번의 fix-cycle 후 PG-14 dropcluster + `--port=5432` 강제로 그린. |
| 2026-05-13 | PR #89 (`claude/work_260513-c`) — `docs/governance/` (README + document-standards.md) + `docs/traceability/` (README + conventions + sync-checklist + report) + PR template + AGENTS/GEMINI 진입점. 1차 종합 매트릭스: REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목, 도메인 그룹 13행. 리뷰어 모드 2-pass 에서 ADR-0003 broken link + workflow-memory 상태 enum 정합화. |
