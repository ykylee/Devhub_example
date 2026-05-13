# Session Handoff — claude/work_260513-b (2026-05-13 진행 중)

- 문서 목적: PR #87 머지 직후 FU-CI-1 (no-docker policy 정합) 처리 sprint 인계.
- 범위: ADR-0003 작성 + ci.yml 의 services: postgres 제거 + native PG 15.
- 최종 수정일: 2026-05-13
- 상태: IN PROGRESS. 브랜치 분기 + ADR + ci.yml 수정 commit 준비 단계.

## 1. 진입 컨텍스트

- 직전 sprint = `claude/work_260513-a` (PR #87, `e86f38f` 로 squash merge). FU-CI-2/3/4 처리 + main flat memory sync.
- 사용자 결정 (2026-05-13): no-docker 정책을 CI 까지 확장. `services: postgres:15` 를 native PG 15 로 교체 (Option A — feedback_no_docker 메모리의 "모든 외부 컴포넌트 포함" 텍스트와 일치).

## 2. 작업 목록

### 2-1. ADR-0003 작성

- `docs/adr/0003-no-docker-policy-ci-scope.md` 신규 — 정책 텍스트 인용, 위반 지점 (`services: postgres:15`), 결정 동인 5개, 검토한 옵션 3개 (A/B/C), Option A 선정, 결과 (즉시 변경 + 후속 영향 + 검증), 미해결 항목.
- `docs/setup/test-server-deployment.md` 의 관련 문서 링크와 §0 정책 줄에 ADR-0003 인용 추가.

### 2-2. ci.yml 수정

- `services: postgres:15` 블록 제거 (sidecar 컨테이너 사용 중단).
- 새 step `Setup PostgreSQL (native)` 추가:
  - pgdg apt repo (`apt.postgresql.org`) signing key + sources.list 등록.
  - `apt-get install -y postgresql-15`.
  - `pg_ctlcluster 15 main start` (GH runner 에는 systemctl 의 PostgreSQL unit 이 없을 수 있어 pg_ctlcluster 사용).
  - `runner` SUPERUSER + `devhub` OWNER 생성.
  - `pg_hba.conf` 의 localhost host entries 를 `trust` 로 변경 (CI 전용, loopback 만 영향).
  - `pg_ctlcluster 15 main restart` + `pg_isready` 30s 대기.

### 2-3. main flat state.json 갱신

- `next_actions.ci_followups` 의 FU-CI-1 항목 제거 + 새 `ci_decisions_2026_05_13` 섹션에 ADR-0003 결정 기록.
- `milestones.CI.status` 를 `done (1차)` → `done` 으로, notes 에 PR #87 + ADR-0003 + 본 PR 추가.

## 3. 검증 계획

- 로컬: ci.yml 의 shell 부분은 ubuntu-specific 이라 직접 검증 불가. 문법 검증은 actionlint 가 없어 git push 후 GHA 실행에 위임.
- CI: PR 생성 → 3 잡 (backend-unit / frontend-unit / e2e) 모두 SUCCESS 재확인. e2e 잡의 새 Setup PostgreSQL (native) step 이 ~30-60 s 추가 시간 소비 예상.

## 4. 다음 행동

1. commit + push + PR 생성.
2. CI 그린 확인 → 리뷰어 모드 2-pass → 사용자 머지 지시 대기.
3. 머지 후 본 슬롯 close 커밋.
