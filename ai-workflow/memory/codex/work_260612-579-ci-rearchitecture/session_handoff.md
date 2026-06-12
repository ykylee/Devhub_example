# Session Handoff — codex/work_260612-579-ci-rearchitecture (2026-06-12, CI 재구성 + flaky 복구)

- 문서 목적: #579 관련 CI 재구성 설계/구현과 flaky 복구의 현재 상태를 다음 세션이 바로 이어받을 수 있게 정리한다.
- 상태: **in progress** — 구현과 로컬 정합 검증은 완료, draft PR 생성 단계.
- 최종 수정일: 2026-06-12

## 1. 이번 세션에서 완료한 것

- `docs/planning/2026-06-12-ci-rearchitecture-design.md`
  - required / regression / quarantine 분리 원칙, runner 활용 방향, branch protection 후보 정리
- `.github/workflows/ci.yml`
  - 빠른 required 경로와 smoke E2E 중심으로 정리
- `.github/workflows/e2e-regression.yml`
  - quarantine 제외 전체 회귀를 shard 기반으로 분리
- `.github/workflows/e2e-quarantine.yml`
  - flaky spec 전용 회귀 분리
- `frontend/tests/e2e-manifests/smoke.txt`
- `frontend/tests/e2e-manifests/quarantine.txt`
- `scripts/select-playwright-specs.sh`
  - smoke / quarantine / regression selection SSOT
- `frontend/tests/e2e/fixtures.ts`
  - `/login` loop 감지, OIDC restart helper, CI timeout 완화, sign-in form 재시도 강화
- `frontend/tests/e2e/signout.spec.ts`
  - signout/user-switch timeout 상향

## 2. 검증 결과

- workflow YAML parse OK
- `scripts/ci-e2e-sync-check.sh` 3 workflow 정합 OK
- `scripts/select-playwright-specs.sh smoke|quarantine|regression` 출력 OK
- focused type check OK
  - `cd frontend && npx tsc --noEmit tests/e2e/fixtures.ts tests/e2e/signout.spec.ts --skipLibCheck`
- full `frontend` `npx tsc --noEmit` 실패
  - 사유: repo 전반의 기존 선행 타입 오류
  - 판단: 이번 변경의 회귀 신호로 보기 어려움

## 3. 핵심 판단

- flaky 는 테스트 자체보다도 OIDC 재진입과 `/login` 체류 회복이 약한 데서 많이 발생했다.
- 전체 E2E 를 단일 required workflow 로 운영하면 flaky 1건이 merge path 전체를 막는다.
- 따라서 required path 는 빠르게 유지하고, regression/quarantine 은 별도 workflow 로 runner 병렬성을 활용하는 편이 낫다.

## 4. 다음 세션 바로 할 일

1. 변경 파일 review 후 draft PR 생성
2. PR 본문에 required check 후보와 quarantine 운영 원칙 명시
3. GitHub Actions 상에서 regression/quarantine 실제 동작 확인
4. branch protection 에서 어떤 check 를 required 로 둘지 최종안 확정
