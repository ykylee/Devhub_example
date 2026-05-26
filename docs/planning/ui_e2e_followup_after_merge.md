# UI E2E 후속 정비 메모

- 문서 목적: `application / project / repository` UI 고도화 PR 이후 남는 E2E 정비 범위를 정리한다.
- 범위: `/devhub` 운영형 basePath 기준 Playwright E2E 후속 작업
- 대상 독자: Frontend, QA, 유지보수 담당자
- 상태: active
- 최종 수정일: 2026-05-26
- 관련 문서: `docs/planning/ui_app_project_repo_upgrade_plan.md`, `docs/setup/e2e-test-guide.md`

## 1. 이번 PR에서 완료된 E2E 범위

- admin applications smoke
- admin projects smoke
- project model route mode gating
- repositories list/detail smoke
- repositories list error/empty state
- repositories detail activity failure
- applications detail rollup failure
- projects detail partial widget failure

모든 검증은 `/devhub` basePath 도커 환경에서 수행했다.

- ingress: `http://localhost:13000/devhub`
- stack: `frontend/backend-core/backend-ai/nginx/keycloak` docker + host postgres

## 2. 머지 전 필수 추가 작업

없음.

현재 PR은 UI 운영 고도화 범위에 필요한 핵심 happy-path / negative-path E2E를 포함한다.

## 3. 머지 후 follow-up 후보

### 3.1 상세 화면 시각 지표 정밀화

- application detail:
  - `pull_request_distribution` 시각화 축/empty 상태 전용 검증
- project detail:
  - task distribution chart 가 실데이터 기반으로 완전히 전환되면 해당 분기 E2E 추가
- repository detail:
  - contributor empty state / activity window 표기 포맷 검증 추가

### 3.2 경계/권한 시나리오 확장

- application detail 404/403 경로
- project detail 404/403 경로
- repository detail 404/403 경로
- role 별 접근 제한 검증 (`developer`, `manager`, `system_admin`)

### 3.3 운영 안정화

- `/devhub` 도커 E2E 부트스트랩을 스크립트화
  - linux target artifact rebuild
  - keycloak setup wrapper (`timeout`) 우회
  - local-idp compose startup
- CI 또는 로컬 helper command로 재사용 가능하게 정리

## 4. 머지 판단 메모

- 이번 PR의 남은 E2E 작업은 모두 **후속 개선 항목**이다.
- 현재 남은 항목은 머지 블로커가 아니라 커버리지 확장/운영 편의성 강화 성격이다.
- 따라서 CI status check 가 통과하면 머지 검토 가능 상태로 본다.
