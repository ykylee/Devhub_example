# 중간 개발 보고 분석 베이스라인

- 문서 목적: 2026-06-02 기준 DevHub 프로젝트의 현재 개발 현황, SDLC 상태, 테스트 상태, AI agent 활용 현황, 활동 통계를 중간 개발 보고용으로 정리한다.
- 범위: 기능 진척, 문서 체계, 추적성, 테스트, agent 활용 추정, 저장소 통계
- 대상 독자: 발표 자료 작성자, 프로젝트 리드, 후속 AI 에이전트
- 상태: in_progress
- 최종 수정일: 2026-06-03
- 기준 브랜치: `main`
- 기준 HEAD: `9f68fb1`
- 관련 문서: [`ai-workflow/memory/state.json`](../../ai-workflow/memory/state.json), [`docs/traceability/report.md`](../traceability/report.md), [`docs/planning/integrated_test_report_20260601.md`](../planning/integrated_test_report_20260601.md), [`docs/governance/worker_division.md`](../governance/worker_division.md), [`docs/planning/role-access-concept.md`](../planning/role-access-concept.md), [`docs/adr/0026-keycloak-role-excluded-decision.md`](../adr/0026-keycloak-role-excluded-decision.md), [`docs/analysis/2026-05-27-codebase-snapshot/README.md`](./2026-05-27-codebase-snapshot/README.md)

## 1. 분석 기준

### 1.1 기준 시점

- 분석일: 2026-06-02
- 현재 브랜치: `main`
- 현재 HEAD: `9f68fb1`
- 최근 상태 요약: CI 회귀 복구 이후, CI Run API/P1 RBAC 후속이 main 에 반영된 상태

### 1.2 주요 근거 문서

| 구분 | 근거 |
| --- | --- |
| 프로젝트 상태 | `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md` |
| 릴리즈/기능 범위 | `docs/planning/release_v1_roadmap.md` |
| 코드/문서 스냅샷 | `docs/analysis/2026-05-27-codebase-snapshot/` |
| SDLC/추적성 | `docs/traceability/report.md`, `docs/domain/` |
| 테스트 현황 | `docs/tests/e2e_testing_strategy.md`, `docs/tests/reports/`, `docs/reports/2026-05-29-test-coverage-sprint.md`, `docs/tests/test_coverage_carve_out_plan.md`, `docs/planning/integrated_test_report_20260601.md` |
| AI agent 역할 | `docs/governance/worker_division.md`, `ai-workflow/memory/<agent>/...`, `git log` |

## 2. 한 줄 상태 평가

DevHub 는 초기 기획/설계 단계를 넘어, 인증·온보딩·관리 설정·Application/Project/Repository·SCM 연동·DREQ·운영 가시화 일부까지 실제 사용 가능한 수준으로 확장되었고, 이를 뒷받침하는 SDLC 문서 체계와 추적성 매트릭스, 테스트 자산이 함께 누적된 상태다.

## 3. 프로젝트 개요와 개발 흐름

### 3.1 현재 보고 가능한 프로젝트 성격

- 통합 관리 플랫폼
- 주요 사용자군: `developer`, `team_manager`, `system_admin`
- 역할 해석:
  - `developer`: 개인 작업, Dev Request, Application/Project/Repository 조회 중심 사용자
  - `team_manager`: 내부 system role 기준 PMO/관리 운영 사용자이며 UI 표시명은 `Manager`, 기본 landing 은 `/manager`
  - `system_admin`: `/admin` 및 `/admin/settings/*` 기반의 관리 운영 사용자
- 최신 정책 기준 legacy `manager` alias 는 `team_manager` 로 정규화되며, Keycloak realm role 은 권한 source-of-truth 가 아니라 DevHub 내부 `users.role` 이 최종 권한 기준이다.
- 신규 사용자는 onboarding 제출과 관리자 review 이후 위 역할 체계에 편입된다.
- 핵심 도메인:
  - auth-session
  - onboarding
  - organization-management
  - rbac-permissions
  - application-lifecycle
  - repository-integration
  - dev-request
  - integration-registry
  - realtime
  - audit-ops

### 3.2 개발 흐름 요약

1. 초반에는 인증, RBAC, 조직 관리, 기본 대시보드 골격을 우선 구축했다.
2. 이후 Application/Project/Repository 도메인과 DREQ, External Integration 으로 범위를 확장했다.
3. v1.0 릴리즈 로드맵이 정리된 뒤에는 온보딩, Keycloak 전환, UI 정비, 테스트 보강, SCM 연동 고도화, CI hardening 이 병행되었다.
4. 최근에는 통합 테스트, 커버리지 remediation, CI Run API, row-level RBAC 심화까지 이어졌다.

### 3.3 활동 밀도가 높았던 구간

`git log --since=2026-05-01` 기준 일별 commit 빈도에서 다음 날짜에 활동이 집중되었다.

- 2026-05-08: 71 commit
- 2026-05-18: 50 commit
- 2026-05-19: 46 commit
- 2026-05-20: 40 commit
- 2026-05-26: 49 commit
- 2026-05-27: 42 commit
- 2026-05-29: 36 commit
- 2026-06-01: 26 commit

## 4. 현재 구현 범위

### 4.1 인증/온보딩/권한

- Keycloak OIDC 로그인 플로우 활성
- 신규 사용자 온보딩 게이트 및 조직 선택 플로우 활성
- 관리자 승인(review) 플로우 활성
- RBAC 정책 편집 및 권한 게이트 활성
- 최근에는 row filter, 조직 subtree scope 등 RBAC 고도화가 추가 반영됨

### 4.2 운영/관리 설정

- `/admin/settings/*` 기반 관리 화면군 구축
- users, organization, permissions, audit, integrations, integration-bindings, dev-requests, dev-request-tokens 등의 관리 UI 및 API 연결

### 4.3 Application / Project / Repository

- Application CRUD 및 상태 관리
- Project CRUD, 멤버 관리, lifecycle 관리
- Repository 연결, draft/publish flow, SCM create/import 흐름
- Application rollup 및 대시보드성 지표 일부 구현

### 4.4 External Integration / SCM

- Integration provider CRUD
- Binding CRUD
- Gitea 중심 SCM repository 조회/import/create
- topology v2 시각화
- 일부 sync worker 및 webhook 관련 기반 구현

### 4.5 DREQ

- Intake token 발급
- 외부 개발 요청 수신
- 사용자별 조회/필터
- Application/Project promote 트랜잭션
- idempotency, validation, row-level filter 검증 완료

### 4.6 CI / 운영 가시화

- CI run 목록 조회 API 동작
- 최신 HEAD 에서 CI Run API 추가 구현
- repository build run 범위와 RBAC 심화가 최근 반영됨

## 5. 현재 사용 가능한 기능 관점 요약

중간 보고에서는 "개발 중인 기능"보다 "지금 당장 시연 가능한 기능"을 구분해 보여줄 필요가 있다.

현재 시연 가능성이 높은 기능은 다음과 같다.

1. 로그인 후 역할별 landing 진입 (`/developer`, `/manager`, `/admin`)
2. 신규 사용자 온보딩 및 관리자 승인
3. `system_admin` 기준 관리자 설정 화면에서 사용자/조직/권한 관리
4. `developer`/`team_manager` 기준 Application 조회 및 상태 확인
5. Project 생성, Repository 연결, 멤버 편집
6. Integration provider 등록 및 SCM repository 생성/가져오기
7. Dev Request 수신 후 Application/Project 로 promote
8. CI run 목록 조회 및 일부 개발 상태 확인

## 6. SDLC 체계 현황

### 6.1 구조

현재 SDLC 문서는 `docs/domain/` 기반 10개 도메인 구조로 정리되어 있다.

- 도메인 수: 10
- 도메인 requirements 문서: 10
- 도메인 architecture 문서: 10
- 도메인 API 문서: 10
- 도메인 test_cases 문서: 7
- 도메인 concept 계열 문서: 6

### 6.2 추적성 체계

`docs/traceability/report.md` 는 요구사항, 유스케이스, 설계, API, 로드맵, 구현, 테스트를 단일 매트릭스로 연결하는 중심 문서다.

보고 포인트:

- traceability 체계가 문서상 선언 수준이 아니라 실제 PR/문서 갱신 절차에 포함되어 있다.
- 도메인 분할 이후에도 `docs/domain/*` 와 추적성 매트릭스가 연결된다.
- 중간 보고에서는 "SDLC 문서가 있다"보다 "문서들이 서로 링크되고 구현 및 테스트로 이어진다"를 보여줘야 한다.

### 6.3 SDLC 보고 메시지 초안

- 컨셉 → 요구사항 → 아키텍처 → API → 구현 → 단위테스트 → E2E 의 체인을 이미 운영 중이다.
- 일부 도메인은 test_cases 문서까지 정비가 완료됐고, 일부는 추가 보강이 필요하다.
- 이 프로젝트는 기능 개발과 함께 문서 분해/정합 작업이 병행된 것이 특징이다.

## 7. 테스트 현황

### 7.1 전략

`docs/tests/e2e_testing_strategy.md` 기준으로 다음 원칙이 운영된다.

- 현실성: 실제 OIDC 흐름 기반 검증
- 멱등성: seeded 계정/복구 기반
- 자동화 우선
- CI 연동

### 7.2 테스트 자산 분류

#### 단위테스트/패키지 테스트

- 프론트엔드 Vitest 테스트 파일: 74
- 백엔드 Go 테스트 파일: 104

#### E2E 테스트

- Playwright E2E spec: 27

#### 통합테스트/시나리오 자산

- 통합 테스트 리포트: `docs/planning/integrated_test_report_20260601.md`
- backend integration 및 store integration 관련 보고/계획 문서 존재

#### 참고 자산

- migration up 파일: 47

### 7.3 커버리지 현황

커버리지는 주로 2026-05-29 coverage sprint 문서를 기준으로 설명하는 것이 가장 안전하다.

#### 프론트엔드 단위테스트 커버리지

2026-06-02 실행 기준 `cd frontend && npm run test:coverage`:

- Lines: **84.71%**
- Statements: **83.41%**
- Branches: **77.80%**
- Functions: **81.33%**

보강 포인트 예시:

- `AuthGuard.tsx` self-coverage: **92.68%**
- `GardenerFeed.tsx`: **96.96%**
- `OrgTree.tsx`: **98.56%**

#### 백엔드 테스트 커버리지

2026-06-02 실행 기준 `cd backend-core && go test ./... -count=1 -coverpkg=./... -coverprofile=...`:

- backend total (`-coverpkg=./...`): **56.8%**
- 2026-05-29 보고치(실 DB 기준 54.4%) 이후 추가 상승이 확인됨

도메인/패키지 예시:

- `internal/domain/application-lifecycle/view`: **90.2%**
- `internal/store`: **20.2%**
- `internal/domain/dev-request/repository`: **24.1%**
- `internal/domain/audit-ops/view`: **47.6%**
- `internal/domain/realtime/view`: **27.8%**

### 7.4 최근 테스트 관련 정황

- 2026-06-02 기준 로컬 `frontend npm run test` PASS: **73 files, 969 tests**
- 2026-06-02 기준 로컬 `frontend npm run test:coverage` PASS
- 2026-06-02 기준 로컬 `backend-core go test ./...` PASS
- 2026-06-02 기준 로컬 `backend-core` coverage 재측정: **total 56.8%**
- 2026-06-02 기준 `colima` fresh stack 재구성 후 `frontend npm run e2e` 최신 실행: **77 passed**
- 초기 재실행에서는 `frontend/tests/e2e/dev-requests.spec.ts` 4건과 정적 skip 3건이 남아 있었으나, `/devhub` base path 정합과 프로젝트/카탈로그 시나리오 안정화 후 전체 통과로 수렴했다
- CI 회귀 복구 완료 후 GitHub Actions green 상태 기록 존재
- 통합 테스트 문서에서는 인증/온보딩, 시스템 설정, SCM 연동, DREQ, CI 조회까지 실제 시나리오 검증 흔적이 확인된다
- 2026-05-29 전후로 frontend/backend 커버리지 remediation sprint 가 연속 수행되었다

### 7.5 중간 보고용 테스트 메시지

- 테스트는 "있다" 수준이 아니라 단위테스트, E2E, 통합 시나리오, 커버리지 remediation sprint 로 다층 구조를 갖추고 있다.
- 특히 DREQ 통합 테스트는 문서상 10/10 통과, BUG/ISSUE 0건으로 보고 가능하다.
- 발표에서는 `Vitest 최근 실행 73 files / 969 tests PASS`, `Playwright 최근 실행 77 passed`, `Playwright spec 27`, `frontend Lines 84.71%`, `backend total 56.8%`처럼 분리해 제시하는 편이 가장 명확하다.

## 8. AI agent 활용 현황

### 8.1 역할 분담

`docs/governance/worker_division.md` 기준 역할은 명확하다.

- Claude: backend + design + traceability + workflow memory
- Codex: infra + CI + security + packaging
- Gemini: frontend + UX + E2E + visual polish
- DeepSeek/Reasonix: 2026-05 말부터 제한적으로 참여했고, 최근 테스트/워크플로우/통합 시나리오 영역에서 존재감이 커진 보조 축

### 8.2 사용 시작 시점

- Claude/Codex/Gemini 는 2026-05 초반부터 main 개발 흐름에 등장
- DeepSeek 는 2026-05-26 전후부터 제한적으로 등장했고, 2026-06-01 통합 테스트/워크플로우 작업에서 가시성이 높아졌다

### 8.3 근거 기반 비중 추정

단일 지표로 단정하지 않고 3개 축으로 추정한다.

| 근거 축 | Claude | Codex | Gemini | DeepSeek |
| --- | ---: | ---: | ---: | ---: |
| memory 디렉터리 수 | 73.4% | 11.0% | 13.3% | 2.3% |
| merge branch 식별 수 | 60.3% | 24.1% | 13.8% | 1.7% |
| commit subject 내 agent 언급 수 | 61.3% | 27.8% | 8.5% | 2.4% |

### 8.4 발표용 해석

- Claude 는 backend/design/문서/정합 작업의 주도 agent 로 보는 것이 타당하다.
- Codex 는 절대량보다 infra/CI/security 쪽 영향도가 크다.
- Gemini 는 frontend/UI/E2E 축에서 집중적으로 기여했다.
- DeepSeek 는 최근 테스트/통합 검증과 workflow 확장에 제한적이지만 명확한 기여 흔적이 있다.

실무 발표에서는 다음 수준으로 요약하는 것이 적절하다.

- Claude: 약 60~70% 수준의 주도 기여
- Codex: 약 15~25% 수준의 인프라/CI/보안 기여
- Gemini: 약 10~15% 수준의 프론트엔드/UX 기여
- DeepSeek: 최근 2~5% 수준의 테스트/검증 보조 기여

## 9. 산출물 및 개발 활동 통계

아래 수치는 모두 `main @ 9f68fb1` 기준으로 git log 및 현재 저장소 파일시스템을 직접 집계한 값이다.

### 9.1 저장소 활동량

- 전체 commit 수: 683
- PR 연계 commit/merge subject 수: 384
- 현재 HEAD 기준 최근 PR 번호: `#462`

### 9.2 문서/설계 자산

- ADR 문서 수: 26
- domain 문서 파일 수: 66
- planning 문서 파일 수: 22
- analysis 문서 파일 수: 26
- reports 문서 파일 수: 5
- 2026-05-27 코드베이스 스냅샷 산출물 수: 24

### 9.3 workflow memory 자산

- `state.json`: 176
- `session_handoff.md`: 112
- `work_backlog.md`: 115
- 일별 backlog 문서: 92

### 9.4 코드/테스트 자산

- frontend page 수: 32
- frontend component 파일 수: 26
- backend migration up 수: 47
- Vitest 테스트 파일 수: 74
- Playwright E2E spec 수: 27
- backend Go test 파일 수: 104

### 9.5 최신 실행 스냅샷

- `cd frontend && npm ci`: PASS
- `cd frontend && npm run test`: PASS, **73 files / 969 tests**
- `cd frontend && npm run test:coverage`: PASS, **Lines 84.71% / Statements 83.41% / Branches 77.80% / Functions 81.33%**
- `cd backend-core && go test ./...`: PASS
- `cd backend-core && go test ./... -count=1 -coverpkg=./... -coverprofile=...`: PASS, **total 56.8%**
- `cd frontend && npm run e2e` (`colima` fresh stack): **77 passed**
- base-path 환경(`/devhub`) 기준 DREQ API 호출과 프로젝트/카탈로그 시나리오까지 포함해 전체 E2E green 을 확인했다

## 10. 보고자료 작성 시 주의점

1. 숫자는 가능한 한 "기준일 기준"으로 명시한다.
2. AI agent 비중은 추정치임을 명시하고, 근거 축을 같이 제시한다.
3. "완료"와 "사용 가능"을 혼동하지 말고, 통합 테스트/CI/문서 근거가 있는 항목을 우선 소개한다.
4. 미래 계획은 짧게 다루고, 현재까지의 실적과 체계를 전면에 둔다.
5. 테스트 수치는 `Vitest`, `Playwright`, `통합테스트`, `coverage`를 분리해 표기한다.

## 11. 슬라이드로 옮길 핵심 문장 후보

- DevHub 는 단순 화면 프로토타입이 아니라 인증, 조직, 프로젝트, 외부 연동, 개발 의뢰 흐름까지 연결된 운영형 플랫폼으로 진화했다.
- 기능 개발과 동시에 SDLC 문서 체계와 추적성 매트릭스를 구축해, 요구사항에서 테스트까지 이어지는 관리 기반을 확보했다.
- AI agent 는 backend/design, infra/CI/security, frontend/UX, 통합 테스트/검증 축으로 역할을 분담하며 개발 속도와 문서화를 함께 끌어올렸다.

## 12. 추가 확인 필요 항목

- 슬라이드용 현재 스크린샷 대상 화면 선정
- 차트 라이브러리 사용 여부
- agent 기여도를 commit 기준과 memory 기준 중 어느 축으로 메인 그래프화할지 결정
