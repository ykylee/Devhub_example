# Session Handoff — main (2026-05-13 EOD)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-13 세션 종료. 머지 17건 (PR #86~#102) — M3 마일스톤 1차 closing + 후속 1-4 + M4 전 잔여 + Project 도메인 컨셉 1차 staged.
- 대상 독자: 후속 에이전트, 프로젝트 리드, 다음 세션 진입자.
- 상태: M1/M2/M3 모두 1차 closing. ADR 5건 신규 (0006~0010). M4 진입 준비 완료 — RM-M4-01..09 (9 항목) + M3 carve out (6 항목) + Project 도메인 후속 (Req → Usecase → Design → Implementation) 이 다음 세션 진입 후보.
- 최종 수정일: 2026-05-13 (세션 종료, PR #102 + housekeeping)
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [거버넌스](../../docs/governance/README.md), [추적성 매트릭스](../../docs/traceability/report.md), [Project 도메인 컨셉](../../docs/planning/project_management_concept.md).
- 브랜치: `main` (HEAD `244f6b1`, PR #102 squash 직후. 본 housekeeping 머지 후 추가 갱신).

## 0. 2026-05-13 머지 흐름

```
244f6b1 PR #102 — docs(planning): Project 도메인 컨셉 1차 — CRUD + 등록 + 조회 MVP scope (claude/work_260513-p)
118899b PR #101 — docs(memory): 2026-05-13 세션 종료 housekeeping sync (claude/work_260513-o)
4134b37 PR #100 — feat(domain,migrations),docs(adr,traceability),scripts: M4 전 잔여 일괄 (claude/work_260513-n)
c7c2f35 PR #99 — feat(hrdb,org,signup),docs(adr,traceability),test: M3 후속 1-4 일괄 (claude/work_260513-m)
b1268ce PR #98 — feat(auth),docs(adr,api,traceability),test: M3 진입 1차 — RM-M3-01..03 (claude/work_260513-l)
3d7d5a2 PR #97 — docs(roadmap,traceability): M3/M4 drift 정합 (claude/work_260513-k)
f551e6a PR #96 — feat(auth),docs(adr,api,traceability),test: M3 진입 전 잔여 후속 일괄 (claude/work_260513-j)
cb9e6d5 PR #95 — docs(traceability,adr,ci),test(frontend): 대형 묶음 B1~D5 (claude/work_260513-i)
ceb0f6f PR #94 — docs(adr): ADR-0004 X-Devhub-Actor 폐기 완료 선언 (B4) (claude/work_260513-h)
594be74 PR #93 — docs(traceability,api): B1 auth 도메인 2차 — §11 본문 ID 노출 + IMPL-auth-01..07 책임 정의 (claude/work_260513-g)
a73dba1 PR #92 — docs(traceability,api): B 묶음 — RBAC API §12 IMPL 정밀 매핑 + 본문 ID 노출 (claude/work_260513-f)
ae8b459 PR #91 — feat(backend): A 묶음 — M1 PR-D 정합 마무리 (writeRBACServerError + writeAuthLoginServerError 통합 + X-Request-ID validation + ctx 표준 request_id 전파) (claude/work_260513-e)
ea8df91 PR #90 — docs(governance,traceability): 갭 분석 정리 + 메타 헤더 표준화 + main flat sync (claude/work_260513-d)
7fac5bf PR #89 — docs: governance + traceability 체계 도입 + 1차 종합 매트릭스 (claude/work_260513-c)
9268227 PR #88 — docs(adr): ADR-0003 no-docker policy CI scope — drop services: postgres + native PG 15 (claude/work_260513-b)
e86f38f PR #87 — ci: FU-CI-2/3/4 (playwright scope, GHA cache, frontend readiness) + main flat memory sync (claude/work_260513-a)
450cc24 PR #86 — ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축 (gemini/prepare-github-action)
```

세부:
- **PR #86** — GitHub Actions CI 1차 도입. backend-unit + frontend-unit + e2e 3잡. 리뷰어 모드 2-pass 로 5 blocker + 5 follow-on fix.
- **PR #87** — FU-CI-2/3/4 처리. playwright install scope (chromium 단일), `actions/cache@v4`, frontend readiness 120s, install split (cache-hit 시 browser install skip).
- **PR #88** — ADR-0003 (no-docker 정책 CI 범위 명문화). `services: postgres:15` sidecar 제거 + pgdg native PG 15 native 설치 step. PG-14 dropcluster + `--port=5432` 강제.
- **PR #89** — `docs/governance/` (README + document-standards.md) + `docs/traceability/` (README + conventions + sync-checklist + report) + PR template + AGENTS/GEMINI 진입점. 1차 종합 매트릭스: REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목.
- **PR #90** — 매트릭스 §5 갭 표 형식 통일 + §5.2 auth.spec.ts TC 미흡수 + §5.3 ADR-0001 vs §3.8 모두 closed. document-standards §2 메타 헤더 9 문서 (누락 4 + 변형 4 + 부분 1) 일괄 적용. PR #87/#88/#89 누적 main flat memory sync.
- **PR #91** — A 묶음 (M1 PR-D 정합 마무리). `writeRBACServerError` (11 호출) + `writeAuthLoginServerError` (5 호출, Pass 1 review 발견) → `writeServerError` 일괄 통합. `requireRequestID` 미들웨어에 `validateCallerRequestID` (정규식 + 길이 + control char) 추가. request_id 를 `requestIDCtxKey{}` typed ctx key 에도 stash + `requestIDFromContext`/`logRequestCtx` ctx-aware helper. kratos_login_client.go 2건 + kratos_identity_resolver.go 1건 ctx-aware 치환. 단위테스트 11건 추가.
- **PR #92** — B 묶음 RBAC 1차. `backend_api_contract.md` §12.2~§12.10 의 9 헤더에 `(API-26..31, 38..40)` 본문 ID 노출. 매트릭스 §2.2 RBAC API + §2.4 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) 서브 표 도입. §5.2 RBAC IMPL 정밀 매핑 항목 closed. Pass 1 review 보강으로 §3 RBAC 행을 ID 범위 + §2 서브 표 참조 패턴으로 정리 ("표 가독성 정책" 명문화).
- **PR #93** — B1 auth 도메인 2차. `backend_api_contract.md` §11.3 `(API-19)` + §11.5 표에 API ID 컬럼 (`API-20..24, 35`) + §11.5.1 `(API-35)` 본문 ID 노출. 매트릭스 §2.2 Auth API + §2.4 IMPL-auth-01..07 책임 정의 (verifier / actor / 5 endpoint handler) 서브 표 도입. §3 인증/회원가입/계정 관리 행 정리 (cross-cut API-23 / API-35 명시).
- **PR #101** — 2026-05-13 세션 종료 housekeeping sync 1차. main flat memory 의 PR #100 흡수 + 다음 세션 진입 후보 명문화 (RM-M4-01..09 + M3 carve out 6 항목).
- **PR #102** — Project 도메인 컨셉 1차. `docs/planning/project_management_concept.md` 신규 (10 절: 도메인 정의 / 행위자 × usecase / MVP scope / 데이터 모델 초안 / 다른 도메인 연계 / UI 컨셉 / 미해결 항목 / 후속 4 sprint hook). 일반 사용자 = 조회 중심, 시스템 관리자 = 등록·관리 전용 분리. RBAC row-scoping (ADR-0011 후보) 는 Design sprint 보류. `docs/planning/README.md §5.1` 도메인 컨셉 인덱스 신설 + `docs/development_roadmap.md §5` 백로그 1행 추가. 추적성 ID 미발급 (컨셉 단계).

## 1. 세션 종료 — 다음 세션 진입 후보

본 세션 (2026-05-13) 종료. 다음 세션 진입자는 본 §1 + §2 를 기준으로 작업 선택.

### 1.1 RM-M4 진입 후보 (매트릭스 §2.3.2 참조)

| RM ID | 항목 | 후보 sprint plan 단위 |
| --- | --- | --- |
| `RM-M4-01` | WebSocket 확장 — infra/ci/risk event publish | M4-WS sprint |
| `RM-M4-02` | WebSocket replay + 리소스 필터 | M4-WS (RM-M4-01 과 묶음 가능) |
| `RM-M4-03` | frontend command status WebSocket UI | M4-WS-UI |
| `RM-M4-04` | AI Gardener gRPC (Python AnalysisService + Go Core client) | M4-AI |
| `RM-M4-05` | AI Suggestion Feed 실데이터 바인딩 | M4-AI (RM-M4-04 직속 후속) |
| `RM-M4-06` | Gitea Hourly Pull worker (Phase 10) | M4-task |
| `RM-M4-07` | System Admin 대시보드 | M4-admin |
| `RM-M4-08` | RBAC PermissionCache LISTEN/NOTIFY ([ADR-0007](../../docs/adr/0007-rbac-cache-multi-instance.md)) | M4-infra (격리된 backend 작업) |
| `RM-M4-09` | 외부 SSO 통합 (Gitea 등) | M4-SSO |

### 1.1.b Project 도메인 후속 진입 후보 (PR #102 컨셉 1차 머지 후, 본 세션 신규)

본 세션에서 [`docs/planning/project_management_concept.md`](../../docs/planning/project_management_concept.md) 컨셉 1차 머지 완료. 후속은 컨셉 §9 의 4 sprint hook 을 따른다.

| 후속 sprint | 산출물 | 진입 조건 |
| --- | --- | --- |
| **Project-Req** | `docs/requirements.md §5.7` 확장 또는 `§5.8 Project 도메인` 신설, REQ-FR-* 일괄 발급 (개별 usecase 단위), NFR (응답시간/페이지네이션) 정의, 매트릭스 row 추가 | 컨셉 머지 직후 (즉시 가능) |
| **Project-Usecase** | 행위자 × usecase 매트릭스 + 핵심 시퀀스 (등록·조회·멤버 변경 3종) + RBAC 매트릭스 확장 후보 → ADR-0011 초안 | Project-Req 머지 직후 |
| **Project-Design (backend)** | `architecture.md` Project 컴포넌트 추가, `backend_api_contract.md` 신규 § `/api/v1/projects/*`, 마이그레이션 (`000012_projects.sql`) | Project-Usecase 머지 직후 |
| **Project-Design (frontend)** | `frontend_development_roadmap.md` 새 phase, 진입 경로 / 컴포넌트 / store 모델 초안 | Backend design 와 병행 |
| **Project-Impl** | IMPL-project-* / UT-project-* / TC-PROJ-* 발급 + 실 구현 | 모든 design 머지 후 |

### 1.2 M3 carve out (M4 와 병행 가능)

- `getHierarchy` MV join 코드 변경 ([ADR-0009](../../docs/adr/0009-org-secondary-memberships-and-total-count-mv.md) §4.2)
- ETL daily cron 운영 entry ([ADR-0008](../../docs/adr/0008-hrdb-production-adapter.md) §6)
- `primary_dept` backfill worker (signup 직후 + admin trigger, [ADR-0010](../../docs/adr/0010-primary-dept-resolution.md) §4.3)
- 파견 종료 (`is_seconded` 자동 갱신) trigger
- 본격 Vitest (AuthGuard / Header / Sidebar mock-heavy)
- TC-CMD/INFRA 인터랙션 spec ts (`test_cases_m3_command_infra.md` §4)

### 1.3 본 세션 머지 영역 요약 (PR #86~#102, 17건)

- **거버넌스/추적성** 체계 도입 + 1차 종합 매트릭스 + 갭 정리 + 메타 헤더 표준화 + M3/M4 drift 정합.
- **Project 도메인 컨셉 staged** (PR #102): 신규 1차 도메인의 컨셉 단계 문서 신규 작성, 후속 4 sprint (Req → Usecase → Design backend/frontend → Impl) 진입 hook 명문화. 추적성 ID 미발급 (컨셉 단계).
- **M1 PR-D 정합** (A 묶음): writeRBACServerError + writeAuthLoginServerError → writeServerError 통합, X-Request-ID validation, ctx 표준 request_id 전파.
- **X-Devhub-Actor 폐기**: ADR-0004 + ADR-0006 (inbound 400 거부) + 회귀 4 테스트 의도 갱신.
- **5 도메인 본문 ID 노출**: RBAC §12 / auth §11 / account §10.2 / users CRUD §10.3 / organization §10.4 / accounts admin / signup §11.5.2.
- **CI / lint**: GitHub Actions 4 잡 (workflow-lint + backend + frontend + e2e) + ADR-0003 no-docker + ADR-0005 actionlint.
- **M3 1차 closing**: Sign Up audit emit + 단위테스트 + §11.5.2 본격 spec / hrdb.MockClient → PostgresClient (ADR-0008) + migration 000010 / parent_id cycle 검증 + ResolvePrimaryUnit (ADR-0010) + 5 단위테스트 / total_count MV migration 000011 (ADR-0009) / scripts/hrdb_etl_seed.sql / frontend signup form alias.

### 1.4 ADR 인덱스 (본 세션 신규 5건)

| ADR | 제목 |
| --- | --- |
| ADR-0006 | inbound `X-Devhub-Actor` 헤더 명시 거부 (400) |
| ADR-0007 | RBAC PermissionCache 다중 인스턴스 일관성 (PG LISTEN/NOTIFY, 구현 carve) |
| ADR-0008 | HR DB production 어댑터 (PostgreSQL `hrdb` schema) |
| ADR-0009 | 파견/겸임 모델 + `total_count` Materialized View |
| ADR-0010 | `users.primary_unit_id` 자동 판정 알고리즘 |

### 1.5 다음 세션 진입 안내

다음 세션 첫 액션 권장:

1. 본 `session_handoff.md` + `state.json` + `work_backlog.md` 읽고 main HEAD 가 본 housekeeping 직후 commit 인지 확인.
2. 본 §1.1 RM-M4 / §1.1.b Project 도메인 후속 / §1.2 M3 carve out 중 우선순위 결정.
3. 새 sprint branch (`claude/work_260514-a` 또는 `claude/project-req-1` 등) + sprint memory 초기화 → 진행.

## 2. 다음 진입점 — 우선순위 후보

`state.json#next_actions` 참조. 본 sprint 머지 후 후속:

| 후보 | 주요 작업 | 규모 |
| --- | --- | --- |
| document-standards §8 우선순위 3 | 본문 ID 명기 (요구사항/설계 문서에 REQ-FR/ARCH/API 옆 backtick ID) | M |
| document-standards §8 우선순위 4 | deprecated 문서 식별 + 마킹 | S |
| E2E 신규 TC 작성 | TC-CMD-*, TC-INFRA-* — 매트릭스 §5.1 의 후보 구현 | M |
| frontend 컴포넌트 Vitest | Header, Sidebar, AuthGuard 등 | S |
| RBAC API §12 IMPL 정밀 매핑 | endpoint 별 IMPL ID 명시 | S |
| X-Devhub-Actor 폐기 ADR | architecture.md §6.2.3 의 완전 제거 trigger | S |
| RBAC cache 다중 인스턴스 일관성 | M1-DEFER-E | M-L |
| actionlint / workflow lint | ADR-0003 §6 의 후속 ADR | S |
| M3/M4 진입 | command status WebSocket UI, WebSocket 확장, AI Gardener gRPC, Gitea Hourly Pull worker | L |

## 3. 환경 / 운영 메모

- **CI 환경**: GitHub Actions ubuntu-24.04. native PostgreSQL 15 (pgdg) + native Ory Hydra/Kratos v26.2.0. `services:` 컨테이너 사용 안 함 (ADR-0003).
- **5 프로세스 native 기동** (prod / dev-server): PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄.

## 4. 잔여 결정 대기

- 본 sprint 후속 항목 (위 §2 표) 의 우선순위 결정.
- 운영 진입 전 hygiene: PoC `test/test` 시드 제거 (test-server-deployment.md §10).

## 5. 거버넌스 / 추적성 진입점

- `docs/governance/README.md` — 두 축 (문서 관리 + 추적성) 인덱스.
- `docs/governance/document-standards.md` — 문서 표준 (메타 헤더, lifecycle, 단계별 유형).
- `docs/traceability/conventions.md` — ID 컨벤션.
- `docs/traceability/sync-checklist.md` — 매 PR 동기화 절차.
- `docs/traceability/report.md` — 1차 종합 매트릭스.
- `.github/pull_request_template.md` — PR body 의 "추적성 영향" 섹션 자동 안내.
