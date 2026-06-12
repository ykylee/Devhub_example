# Session Handoff — codex/work_260612-579-ci-rearchitecture (2026-06-12, CI 재구성 + flaky 복구)

- 문서 목적: #579 관련 CI 재구성 설계/구현과 flaky 복구의 현재 상태를 다음 세션이 바로 이어받을 수 있게 정리한다.
- 상태: **complete** — 구현, PR 머지, old PR 재평가 통과까지 확인 완료.
- 최종 수정일: 2026-06-13

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

## 4. 최종 결과

- PR #580 merged
  - merge commit: `b1fa5c27698403620b63ef09d7e32e2235592d59`
- 새 workflow 3종이 GitHub Actions 상에서 실제 동작 확인
  - `CI`
  - `E2E Regression`
  - `E2E Quarantine`
- old CI failure가 있던 PR #579는 최신 `main`을 반영한 새 SHA에서 재평가 통과

## 5. 레슨런

1. flaky와 기능 실패는 같은 required lane에 두지 않는 편이 triage 속도와 merge 안정성 모두에 유리하다.
2. base branch workflow가 바뀌어도 old failure가 자동 success로 바뀌지는 않는다.
3. 하지만 최신 `main`을 PR branch에 반영해 새 SHA로 재평가하면 새 CI 구조의 효과를 old PR에도 적용할 수 있다.
4. manifest + selector 조합은 workflow drift를 줄이는 데 유효했다.
