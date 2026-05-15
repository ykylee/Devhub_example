# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: M1·M2·M3 done (1차). 2026-05-14 sprint 로 **Application Domain (backend 1차) 완성** (API-01~58 + ADR-0011 + RBAC matrix 4 신규 resource + CI 5 job). **2026-05-15 세션 종료** — 5 sprint(claude/work_260515-a/b/c/d/e) + 2 codex/gemini PR + 1 누락 흡수 = 총 7 PR. **ADR-0011 §4.2 enforceRowOwnership helper + pmo_manager seed effective 활성화** 완료 (PR #118 + #119), endpoints 통일 모듈 + theme FOUC + standalone gate + token sweep 일괄. 본인 4단계 리뷰 4회 + codex review cycle 1회 완주. 다음 진입 후보는 session_handoff.md §3 참조.
- 최종 수정일: 2026-05-15
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | ✅ done | 2026-05-11 | PR-B+C (#56) + PR-D (#57) 머지로 envelope / types / WS / CommandStatus / audit actor enrichment / request_id 완결. PR-D 후속 (#80/#82) 으로 commands enrichment + DEVHUB_TRUSTED_PROXIES 보강. |
| **M2** — 사용자 경험 정합 | ✅ done (1차 완성) | 2026-05-12 | login_action + work_26_05_11 + work_26_05_11-d 완료 후 PR #85 (`claude/login_usermanagement_finish`) 가 1차 완성 sprint 로 닫음: 로드맵 정합 + UX hygiene (PR-UX1+2+3) + Kratos audit (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| **CI** — GitHub Actions 도입 | ✅ done (1차) | 2026-05-13 | PR #86 (`gemini/prepare-github-action`) 가 backend-unit + frontend-unit + e2e (Playwright 40 TC) 3잡 도입. 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개로 그린 도달. PR-T5 의 핵심 잡 묶음은 본 PR 으로 1차 완료. 후속: FU-CI-1 (no-docker policy 정합), FU-CI-2/3/4 는 `claude/work_260513-a` 처리 중. |
| **M3** — Realtime 확장 + 외부 연동 1차 | ✅ done (1차) | 2026-05-13 | RM-M3-01..03 (Sign Up + 인사 DB + 조직 polish) 완료. ADR-0008/0009/0010 신규. |
| **Application Domain (backend 1차)** | ✅ done | 2026-05-14 | API-01~58 전체 activated (본 세션 #104~#110). 마이그레이션 000012~000018 (7), ADR-0011 accepted, RBAC 4 신규 resource (system_admin 일임), CI 5 job (Backend Integration 신설). 23 integration test (P1/P2 회귀 guard 포함) 가 CI 에서 실 실행. |
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
| 2026-05-13 | PR #90 (`claude/work_260513-d`) — 매트릭스 §5 갭 표 형식 통일 + §5.2 auth.spec.ts TC 미흡수 + §5.3 ADR-0001 vs §3.8 모두 closed. document-standards §2 메타 헤더 9 문서 (누락 4 + 변형 4 + 부분 1) 일괄 적용. PR #87/#88/#89 누적 main flat memory sync. 매트릭스 ADR-0003 link 정상화. |
| 2026-05-13 | PR #91 (`claude/work_260513-e`) — A 묶음 (M1 PR-D 정합 마무리). `writeRBACServerError` (11 호출) + `writeAuthLoginServerError` (5 호출, Pass 1 review 발견) → `writeServerError` 통합. `requireRequestID` 미들웨어에 `validateCallerRequestID` (`^[A-Za-z0-9_-]{1,128}$`) + ctx 표준 request_id 전파 (typed key `requestIDCtxKey{}` + `requestIDFromContext`/`logRequestCtx`). kratos_login_client + kratos_identity_resolver 의 untraced `log.Printf` 3건 ctx-aware 치환. 단위테스트 11건 추가. |
| 2026-05-13 | PR #92 (`claude/work_260513-f`) — B 묶음 RBAC 1차. `backend_api_contract.md` §12.2~§12.10 의 9 헤더에 `(API-26..31, 38..40)` 본문 ID 노출. 매트릭스 §2.2 RBAC API 매핑 + §2.4 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) 서브 표. §5.2 의 "RBAC API §12 IMPL 정밀 매핑" closed. Pass 1 review 보강으로 §3 RBAC 행을 ID 범위 + §2 서브 표 참조 패턴으로 정리 ("표 가독성 정책" 명문화). |
| 2026-05-13 | PR #93 (`claude/work_260513-g`) — B1 auth 도메인 2차. `backend_api_contract.md` §11.3 `(API-19)` + §11.5 표에 API ID 컬럼 (`API-20..24, 35`) + §11.5.1 `(API-35)` 본문 ID 노출. 매트릭스 §2.2 Auth API 매핑 + §2.4 IMPL-auth-01..07 책임 정의 (verifier / actor / 5 endpoint handler) 서브 표. §3 인증/회원가입/계정 관리 행 정리 (cross-cut API-23 / API-35 명시). Pass 1 review clean. |
| 2026-05-13 | PR #94 (`claude/work_260513-h`) — B4 X-Devhub-Actor 폐기 ADR-0004 발급. M0 SEC-4 + M1 PR-D Bearer token verifier 도입으로 ADR-0001 §8 #4 trigger 가 이미 충족됐음을 ex-post-facto 명문화. architecture.md / ADR-0001 §8 #4 / me.go 주석 잔재 정리. Pass 1 review 보강으로 `backend_api_contract.md` §8/§9.1/§9.2/§11.3 + frontend_integration_requirements §3.5 + environment-setup §2.4 의 spec 잔재 6 위치 함께 정리 + ADR-0004 §5 정정. 매트릭스 §4 ADR-0004 + §5.3 closed + §6 변경 이력. |
| 2026-05-13 | PR #95 (`claude/work_260513-i`) — 대형 묶음 B1~D5. B1 추가 5 도메인 (account / org / command / audit / infra) 본문 ID 노출 + §10.1 carve out 표 (API-25/33/34/36/37) + §2.2 4 신규 서브 표 + §2.4 6 신규 IMPL 서브 표. §3 6 행 정리. B2 archive/AGENTS.md + archive/split_checklist.md deprecated. C1 ThemeToggle Vitest 3 tests. C2 test_cases_m3_command_infra.md (5 TC 카탈로그, spec ts carve). D5 ADR-0005 + workflow-lint 잡 (actionlint). 매트릭스 §5.1/§5.2/§4/§6 갱신. Pass 1 review clean. |
| 2026-05-13 | PR #96 (`claude/work_260513-j`) — M3 진입 전 잔여 후속 일괄. D6 inbound X-Devhub-Actor 거부 (auth.go) + ADR-0006. 회귀 4 테스트 의도 갱신 (ignore → reject). ADR-0007 RBAC cache 다중 인스턴스 결정 (PG LISTEN/NOTIFY, 구현 carve). B2-2 deprecated 추가 마킹 (requirements_review / DOCUMENT_INDEX / assessment). 매트릭스 §2.2 nit 정정. backend_api_contract §10.2 (API-25 accounts admin) + §10.3 (API-33 users CRUD) + §10.4 (API-34 organization) spec 절 신설 — §5.2 "본문 spec 부재 endpoint" closed. AuthGuard smoke Vitest. TC-INFRA-RENDER-01 spec ts. 매트릭스 §4 ADR-0006/0007 + §5.3 closed. |
| 2026-05-13 | PR #97 (`claude/work_260513-k`) — M3/M4 drift 정합. development_roadmap.md §3 single source-of-truth 명시. 매트릭스 §2.3.1 RM-M3 (3 항목) + §2.3.2 RM-M4 (9 항목) 정의 표 신규. §3 도메인 행 RM-M3 → RM-M4 정합 (명령/실시간/인프라). state.json m3/m4 분리. |
| 2026-05-13 | PR #98 (`claude/work_260513-l`) — M3 진입 1차 RM-M3-01..03. auth_signup.go audit emit + auth_signup_test.go 4 case + §11.5.2 본격 spec + ADR-0008 (HRDB PostgreSQL 채택) + §10.4.1~§10.4.4 schema 보강. |
| 2026-05-13 | PR #99 (`claude/work_260513-m`) — M3 후속 1-4. PostgresClient + migration 000010 + cycle 검증 + ADR-0009 + frontend signup form. |
| 2026-05-13 | PR #100 (`claude/work_260513-n`) — M4 전 잔여. ETL seed SQL + ADR-0008 §6 갱신 + MV migration 000011 + ADR-0010 primary_dept + 5 단위테스트 + signup alias 단일화. |
| 2026-05-14 | PR #112 (`codex/frontend_color_review`, 머지 `3f387cd`) — Admin UI 가독성 통일 + ActionMenu 공통 컴포넌트 + iPad 터치 안정화 + 백엔드 조직 단위 트랜잭션 강화. flat memory 흡수는 2026-05-15 sprint 에서. |
| 2026-05-15 | sprint `claude/work_260515-a` — 본인 PR 2건 리뷰어 모드 + housekeeping. PR #115 (`b669bc7`, light theme + dropdown refactor + **endpoints 통일 모듈**) + PR #114 (`25f97ba`, Application leader/dev_unit + search 확장 + search predicate refactor + leader backfill). 본 sprint 핵심: `frontend/lib/config/endpoints.ts` 단일 진실 소스 도입 + theme FOUC inline script 패턴 정착 + `output: standalone` NEXT_OUTPUT gate. PR #115 E2E 2회 실패 → service-logs artifact + raw job log 두 layer 분석 학습. PR #116 (`cbc36b0`) 머지. |
| 2026-05-15 | sprint `claude/work_260515-b` — PR #114 신규 컴포넌트 token 정책 일관 sweep. Application/Project/Repository 모달 3종의 `text-primary-foreground` 50건을 `text-foreground dark:text-primary-foreground` 패턴으로 마이그레이션 (PR #115 의 UserCreationModal reference). light theme 가독성 회귀 차단. PR #117 (`68f031e`) 머지. 본인 4단계 self-review clean (보강 commit 불필요). |
| 2026-05-15 | sprint `claude/work_260515-c` — ADR-0011 §4.2 `enforceRowOwnership` helper 도입. `internal/httpapi/permissions.go` 에 시그니처 `(c, ownerUserID, ...allowedRoles) bool` + audit `auth.row_denied` + 403 `code=auth_row_denied`. 단위 테스트 6건 (system_admin / allowedRoles / owner-self / deny + audit / empty owner / no actor). REQ-FR-PROJ-009 "후속" → "활성화" 표기 갱신. ADR-0011 §5/§7 변경 이력. PR #118 (`519a508`) 머지. 본인 4단계 self-review clean. |
| 2026-05-15 | sprint `claude/work_260515-d` — codex 외부 리뷰 hotfix. P1×3 + P2×1. (1) ApplicationCreationModal + ProjectCreationModal edit 분기 immutable key 제외, (2) ApplicationTable + ProjectTable 의 `new Date(YYYY-MM-DD)` → `parseISO` (timezone shift 회귀), (3) PR #115 P1 은 이미 해소, (4) **pmo_manager seed migration 000021** + handler 4개에 `enforceRowOwnership` wire (`updateApplication/archiveApplication/updateProject/archiveProject`) + `enforceRowOwnership` 의 `devFallbackEnabled` bypass. e2e regex anchor (`admin-users-crud.spec.ts` 의 "Manager" 가 새 "PMO Manager" option 과 strict mode 충돌 → `/^Manager$/` exact match + "PMO Manager" guard 추가). PR #119 (`bca612e`) 머지. 본인 4단계 self-review clean. |
| 2026-05-15 | sprint `claude/work_260515-e` — 세션 종료 housekeeping. 본 세션 7 PR (#112/#114/#115/#116/#117/#118/#119) 흡수 + 다음 세션 진입 후보 9건 정리. main flat state.json/handoff/work_backlog + auto memory 갱신. |
