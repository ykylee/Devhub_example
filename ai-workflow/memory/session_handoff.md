# Session Handoff — main (2026-05-13 EOD)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-13 머지 11건 (PR #86~#96) 정리 + 진행 중 sprint + M3/M4 진입 후보.
- 대상 독자: 후속 에이전트, 프로젝트 리드.
- 브랜치: `main` (HEAD `f551e6a`, PR #96 squash 직후).
- 최종 수정일: 2026-05-13
- 상태: M1/M2 100% done. CI 그린 + 거버넌스 + 5 도메인 ID 노출 + X-Devhub-Actor 폐기 + actionlint + frontend Vitest 1차 + TC 카탈로그 + ADR-0006/0007 + §10.2~§10.4 spec 절 완료. 진행 중 sprint `claude/work_260513-k` 가 M3/M4 drift 정합 (development_roadmap §3 source-of-truth 명시 + 매트릭스 §2.3.1/§2.3.2 정의 표 + §3 도메인 행 정합 + state.json m3/m4 분리) 처리 중.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [거버넌스](../../docs/governance/README.md), [추적성 매트릭스](../../docs/traceability/report.md).

## 0. 2026-05-13 머지 흐름

```
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

## 1. 진행 중 sprint — `claude/work_260513-k` (M3/M4 drift 정합)

M3/M4 마일스톤 정의의 3 source (development_roadmap.md, 매트릭스, backend_roadmap + state.json) 사이 drift 해소:

- **source-of-truth 명시**: `docs/development_roadmap.md` §3 가 single source-of-truth. 매트릭스 §2.3 + state.json + backend_roadmap §5 모두 이 정의 기준 정합.
- **development_roadmap.md §3 갱신**: M3 헤더 중복 정리 + M2 흡수된 사용자/조직 관리 항목 ✅ 표기 + M3 잔여 (Sign Up + 인사 DB + 조직 polish) 명확화 + M4 정의 (실시간 + AI + Task + Admin) 보강.
- **매트릭스 §2.3 + §2.3.1 RM-M3 정의 표** (3 항목): Sign Up + 인사 DB + 조직 polish.
- **매트릭스 §2.3.2 RM-M4 정의 표** (9 항목 신규 발급): WebSocket 확장 (RM-M4-01), replay (RM-M4-02), frontend WS UI (RM-M4-03), AI Gardener gRPC (RM-M4-04), Suggestion Feed (RM-M4-05), Gitea Pull (RM-M4-06), System Admin 대시보드 (RM-M4-07), RBAC cache LISTEN/NOTIFY (RM-M4-08), 외부 SSO (RM-M4-09).
- **매트릭스 §3 도메인 행 정합**: 회원가입 (M3-01/02 유지), 명령 lifecycle/실시간/인프라 (M3-XX → M4-XX), M4 마지막 row 정의 갱신.
- **state.json m3/m4 분리**: `m3_entry_candidates` 신규 (Sign Up + 인사 DB + 조직 polish), `m4_entry_candidates` 갱신 (RM-M4-01..09 인용).
- **backend_development_roadmap.md** §5: 이미 정합 (변경 minor).

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
