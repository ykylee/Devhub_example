# 2026-06-12 CI 재구성 설계안

- 문서 목적: PR CI 실패율과 재시도 비용을 낮추기 위한 DevHub CI 재구성 목표 상태와 단계별 실행 설계를 정의한다.
- 범위: GitHub Actions workflow 분리, required gate 재정의, E2E 분류, runner 활용 전략, 단계별 전환 계획.
- 대상 독자: DevHub contributor, 리뷰어, CI 운영 담당자, AI agent.
- 상태: draft
- 최종 수정일: 2026-06-12
- Tier: 공용
- 관련 문서: [문서 작성·관리 표준](../governance/document-standards.md), [v1.0 릴리즈 로드맵](./release_v1_roadmap.md), [추적성 체크리스트](../traceability/sync-checklist.md)

## 1. 배경

2026-06-12 기준 `PR #579`의 최신 CI는 `E2E Tests (Playwright, shard 3/3)`만 실패했고, backend/frontend unit, integration, lint 계열은 모두 통과했다. 직전 rerun 역시 같은 shard가 실패했고 실패 패턴도 반복됐다.

관찰된 핵심 징후는 다음과 같다.

- PR CI 실패가 대부분 Playwright shard 3/3에 집중된다.
- `signout.spec.ts`는 이미 CI flaky 성격을 코드 주석과 로그에서 드러낸다.
- `voc-auto-routing.spec.ts`는 `beforeAll`에서 admin login + platform PATCH를 수행하는 무거운 준비 단계에 의존한다.
- 각 shard가 PostgreSQL 설치, `npm ci`, Playwright deps 설치, Keycloak pull/start를 반복해 환경 준비 비용이 크다.
- 현재 workflow는 fast validation과 heavy regression이 같은 required lane에 섞여 있어, flaky failure가 merge blocker가 된다.

## 2. 문제 정의

현재 CI 구조의 문제는 단순히 테스트가 느리다는 점이 아니다. 더 큰 문제는 아래 세 가지다.

### 2.1 실패 원인의 구분이 어렵다

현재 PR check가 실패하면 코드 회귀, 환경 bootstrap failure, flaky timeout이 같은 수준의 blocker로 보인다. 리뷰어와 작성자는 rerun 없이는 실패 종류를 빠르게 구분하기 어렵다.

### 2.2 runner를 병렬로 써도 체감 효율이 낮다

shard를 3개로 나누었지만, 느린 spec과 환경 의존 spec이 특정 shard에 몰리면 전체 wall-clock 이 longest shard에 묶인다. 즉 shard 수 증가가 곧 효율 증가로 이어지지 않는다.

### 2.3 required gate가 과하게 무겁다

모든 PR에서 full Playwright regression을 required로 걸면, 생산성은 flaky test의 안정도에 종속된다. 현재는 heavy regression과 fast quality gate를 분리해야 한다.

## 3. 목표

이번 재구성의 목표는 다음과 같다.

1. PR required CI를 빠르고 안정적으로 만든다.
2. full E2E regression은 여러 runner를 효율적으로 쓰도록 별도 lane으로 분리한다.
3. flaky 테스트를 quarantine 해 merge blocker에서 분리한다.
4. 실패 시 artifact와 로그를 통해 원인 분류 속도를 높인다.

## 4. 설계 원칙

### 4.1 Required gate는 코드 회귀 신호 위주로 유지한다

required lane에는 lint, unit, integration, smoke E2E만 남긴다. full regression은 required에서 분리한다.

### 4.2 환경 준비형 테스트는 일반 shard에 섞지 않는다

`beforeAll`에서 admin 로그인, external system patch, long retry/backoff를 수행하는 테스트는 일반 smoke 또는 stable shard에 포함하지 않는다.

### 4.3 flaky 테스트는 숨기지 말고 quarantine 한다

flaky test를 required에 남겨두고 rerun에 기대는 대신, quarantine lane으로 이동해 pass/fail 추세를 별도로 관찰한다.

### 4.4 build-once, run-many를 유지한다

backend/frontend build artifact 재사용, migrate binary 재사용은 유지한다. 다만 shard별 환경 bootstrap 비용은 계속 줄이는 방향으로 간다.

### 4.5 runner는 성격별로 나눈다

짧은 static validation, 일반 unit/integration, browser-heavy E2E를 같은 runner class에 섞지 않는다.

## 5. 목표 구조

## 5.1 Workflow 계층

### A. `CI` (Fast Required)

PR merge gate 역할을 담당한다.

- changed-paths
- workflow lint
- migration prefix lint
- OpenAPI YAML lint
- backend unit
- frontend unit
- backend integration
- smoke E2E

### B. `E2E Regression`

full Playwright regression을 담당한다.

- PR relevant change 시 실행
- `workflow_dispatch` 수동 실행
- nightly schedule 실행
- non-blocking 또는 branch protection 비필수 check로 운영

### C. `Quarantine E2E`

CI flaky 이력이 있는 테스트를 별도로 운영한다.

- signout 계열
- heavy setup / long timeout 계열
- 반복 flaky 이력 spec

## 5.2 E2E 분류

### Smoke

빠른 로그인/기본 라우팅/핵심 진입 검증 중심.

- 인증 기본 플로우
- 기본 landing / onboarding 진입
- 핵심 목록/상세 한두 개

Smoke는 1개 job에서 serial로 돌려도 좋다. 목적은 breadth가 아니라 stable gate 제공이다.

### Regression

전체 Playwright suite를 shard fan-out으로 실행한다.

- 관리 화면 CRUD
- repository / SCM flows
- topology / integrations
- dev request / dogfood
- screenshot capture

### Quarantine

현재 기준 우선 후보:

- `frontend/tests/e2e/signout.spec.ts`
- `frontend/tests/e2e/voc-auto-routing.spec.ts`

이 둘은 회귀 검증 가치가 없다는 뜻이 아니라, 안정성 회복 전까지 required gate에 섞지 않는다는 뜻이다.

현재 저장소에서는 workflow 내부 하드코딩 대신 아래 manifest를 source-of-truth 로 사용한다.

- `frontend/tests/e2e-manifests/smoke.txt`
- `frontend/tests/e2e-manifests/quarantine.txt`

workflow는 `scripts/select-playwright-specs.sh`를 통해 smoke / regression / quarantine 대상을 계산한다.

## 6. Runner 활용 전략

### 6.1 Runner class

- `light`: changed-paths, lint, OpenAPI, migration lint
- `medium`: backend/frontend unit, backend integration
- `heavy`: Playwright smoke, full regression, browser + Keycloak + PostgreSQL 포함 job

현재는 `ubuntu-latest` 단일 label을 유지하더라도, 설계상 역할을 분리해 두고 self-hosted 또는 larger runner 도입 시 그대로 옮길 수 있게 job 경계를 먼저 정리한다.

### 6.2 병렬화 기준

- lint / unit / integration은 independent job으로 병렬화
- regression은 shard fan-out
- smoke는 단일 job 유지

주의할 점은 shard 수보다 shard 구성이다. 느린 spec이 한 shard에 몰리면 runner를 늘려도 total wall-clock은 줄지 않는다. 따라서 fan-out 전, spec 분류가 선행되어야 한다.

## 7. Artifact / 진단 전략

failure triage 속도를 높이기 위해 다음을 기본으로 한다.

- Playwright HTML report 업로드
- trace 업로드
- screenshot 업로드
- backend/frontend log 업로드

현재 CI의 `playwright-report` path warning은 즉시 해소해야 한다. regression lane은 실패 시 rerun보다 artifact 분석이 먼저 가능해야 한다.

## 8. 단계별 실행 계획

### Phase 1. Workflow 분리

- 기존 `CI`를 fast required 중심으로 축소
- full Playwright는 `E2E Regression` workflow로 분리
- smoke E2E를 명시적 spec list로 분리

### Phase 2. Quarantine 도입

- signout, voc-auto-routing 등 flaky / heavy setup spec을 quarantine lane으로 이동
- regression lane은 stable regression과 quarantine regression을 분리

### Phase 3. Runner 최적화

- 필요 시 heavy lane에 larger / self-hosted runner 적용
- shard 수와 spec 배치를 runtime data 기반으로 조정

### Phase 4. 안정화 및 재편입

- flaky root cause 수정
- pass rate가 회복된 spec은 quarantine에서 regression, 이후 required 여부 재평가

## 9. 1차 구현 범위

현재까지 적용한 범위는 다음과 같다.

- `CI` workflow를 fast required gate로 축소
- smoke E2E 전용 job 도입
- `E2E Regression` workflow 신설
- regression workflow에 HTML report artifact 복구
- `Quarantine E2E` workflow 신설
- regression workflow에서 quarantine spec 제외

다음 항목은 후속 PR로 남긴다.

- spec tag/폴더 기준 재편
- self-hosted runner 또는 larger runner 도입
- flaky pass-rate 측정 자동화

## 9.1 운영 기준

branch protection required check 이름은 아래를 기준 후보로 유지한다.

- `Workflow Lint (actionlint)`
- `Migration Prefix Uniqueness`
- `OpenAPI YAML Lint`
- `Backend Unit Tests`
- `Frontend Unit Tests`
- `Backend Integration Tests`
- `E2E Smoke Tests`

`E2E Regression`과 `E2E Quarantine`은 진단 및 추세 관찰 lane이며, 기본적으로 required 대상에서 분리한다.

## 10. 성공 기준

1차 재구성의 성공 기준은 아래와 같다.

- PR required CI가 full regression failure 없이 완료된다.
- smoke lane이 merge gate 역할을 안정적으로 수행한다.
- full regression은 별도 workflow로 병렬 실행된다.
- 실패 시 report/trace/log artifact로 원인 분석이 가능하다.

## 11. 리스크와 대응

### 리스크 1. smoke coverage가 너무 얇아질 수 있다

대응:

- smoke spec은 인증 + 핵심 사용자 여정 최소 2~3개로 구성
- full regression을 PR relevant change와 nightly에 계속 유지

### 리스크 2. branch protection required check 이름이 바뀔 수 있다

대응:

- workflow split 후 required check 이름을 고정
- 운영 측 branch protection 설정은 후속으로 동기화

### 리스크 3. regression이 non-blocking이 되며 품질 감시가 느슨해질 수 있다

대응:

- nightly + manual rerun 경로 유지
- regression failure 추세를 PR 코멘트 또는 workflow summary로 노출하는 후속 개선 검토

## 12. 이번 문서에 기반한 구현 원칙

이번 문서를 기준으로 코드를 수정할 때는 다음 순서를 따른다.

1. required gate와 regression lane을 먼저 분리한다.
2. flaky spec을 required lane에 다시 섞지 않는다.
3. 환경 bootstrap 비용 절감보다 merge gate 안정화에 우선순위를 둔다.
