# 2026-06-13 CI 재구성 결과 및 레슨런

- 문서 목적: 2026-06-12 CI 재구성 및 flaky 복구 작업의 실제 결과, 운영 효과, 재사용 가능한 레슨런을 정리한다.
- 범위: PR #580 머지 결과, PR #579 재평가 결과, workflow 구조 효과, 다음번 CI 복구 절차.
- 대상 독자: DevHub contributor, 리뷰어, CI 운영 담당자, AI agent.
- 상태: active
- 최종 수정일: 2026-06-13
- Tier: 공용
- 관련 문서: [2026-06-12 CI 재구성 설계안](./2026-06-12-ci-rearchitecture-design.md), [문서 작성·관리 표준](../governance/document-standards.md)

## 1. 요약

2026-06-12에 적용한 CI 재구성은 실제 운영에서 유효했다.

- `CI` / `E2E Regression` / `E2E Quarantine` 3계층 분리가 GitHub Actions 상에서 정상 동작했다.
- 재구성 PR #580은 새 구조로 전 체크를 통과한 뒤 main에 머지되었다.
- 기존에 flaky와 기능 실패가 섞여 있던 PR #579는 최신 `main`을 반영해 새 SHA로 다시 평가한 뒤 새 CI 구조에서 통과했다.

즉, 이번 작업은 단순한 설계안이 아니라 **운영에서 재현된 복구 패턴**으로 검증되었다.

## 2. 무엇을 개선했는가

### 2.1 required path와 regression path를 분리했다

기존에는 heavy Playwright 회귀와 빠른 품질 게이트가 같은 PR required lane에 묶여 있었다. 이번에는 아래처럼 역할을 분리했다.

- `CI`: lint / unit / integration / smoke E2E
- `E2E Regression`: non-quarantine full regression
- `E2E Quarantine`: flaky 또는 고비용 spec

이 구조 덕분에 merge gate는 가볍고 예측 가능해졌고, 전체 회귀는 runner 병렬성을 더 잘 활용할 수 있게 되었다.

### 2.2 spec selection의 source-of-truth를 만들었다

workflow YAML 안에 spec 경계를 하드코딩하지 않고 아래 3개를 기준으로 삼았다.

- `frontend/tests/e2e-manifests/smoke.txt`
- `frontend/tests/e2e-manifests/quarantine.txt`
- `scripts/select-playwright-specs.sh`

이 방식은 smoke / regression / quarantine 구성을 바꿀 때 workflow drift를 줄이고, 어떤 spec이 어느 lane에 속하는지 문서 없이도 추적 가능하게 만든다.

### 2.3 signout/login flaky를 테스트 코드에서 직접 완화했다

이번 복구의 핵심은 workflow 분리만이 아니었다. `frontend/tests/e2e/fixtures.ts` 와 `frontend/tests/e2e/signout.spec.ts` 에서 flaky 특성을 직접 흡수했다.

- app `/login` page 체류 감지
- OIDC restart helper 추가
- redirect chain 정체 시 reload 재시도
- CI 환경 timeout 완화

이 조합이 있었기 때문에 단순 rerun이 아니라 **재평가 가능한 상태**로 테스트를 올릴 수 있었다.

## 3. 실제 운영 결과

### 3.1 PR #580 결과

- PR #580: CI 재구성 및 flaky 복구 PR
- 결과: **merged**
- merge commit: `b1fa5c27698403620b63ef09d7e32e2235592d59`

확인된 점:

- `CI` workflow 정상 동작
- `E2E Regression` workflow 정상 동작
- `E2E Quarantine` workflow 정상 동작
- shard fan-out 및 artifact build 분리 정상 동작

### 3.2 PR #579 재평가 결과

PR #579는 원래 old CI에서 shard 3/3 failure가 merge blocker였다. 새 구조가 main에 적용된 뒤, PR 브랜치에 최신 main을 반영하고 새 SHA로 다시 평가했다.

- 기존 실패 run: old single-lane CI 기준 실패
- 조치: PR #579 브랜치에 최신 `main` merge
- 결과: `CI`, `E2E Regression`, `E2E Quarantine` 모두 성공
- 상태: 2026-06-12 기준 GitHub PR status `CLEAN`

이 결과는 “base branch에 CI 구조가 머지되기만 하면 과거 실패 PR이 자동으로 초록색으로 바뀐다”는 뜻은 아니다. 대신 다음 운영 규칙이 유효하다는 걸 의미한다.

> **과거 실패 PR도 최신 main을 반영해 새 SHA로 다시 태우면, 새 CI 구조에서 회복될 수 있다.**

## 4. 레슨런

### 4.1 flaky와 기능 실패를 같은 lane에 두지 말아야 한다

예전 구조에서는 signout timeout 같은 flaky와 기능 실패가 같은 shard failure로 보였다. 이 상태에서는 rerun 비용이 크고 triage 우선순위도 흔들린다.

이번 작업으로 확인된 점:

- flaky는 quarantine으로 분리해야 한다.
- 기능 실패는 regression에서 더 명확하게 드러난다.
- required lane은 smoke 수준으로 좁혀야 merge blocker 품질이 올라간다.

### 4.2 old PR의 실패 상태는 자동으로 재해석되지 않는다

GitHub는 required checks를 latest commit SHA 기준으로 판단한다. 그래서 base branch의 workflow가 바뀌어도, 예전 실패 run 자체가 success로 변하지는 않는다.

운영상 필요한 조치는 다음 둘 중 하나다.

1. PR 브랜치에 새 commit을 push한다.
2. 최신 `main`을 PR 브랜치에 반영해 새 merge SHA로 다시 평가받는다.

이번 사례에서는 2번이 효과적이었다.

### 4.3 build-once / run-many 구조는 유지할 가치가 높다

artifact reuse는 이번에도 유지 가치가 있었다. workflow를 쪼개더라도 다음은 계속 가져가는 편이 좋다.

- backend/frontend build artifact 재사용
- migrate binary 재사용
- shard별 app bootstrap 비용 최소화

즉, workflow를 나눈다고 해서 artifact 전략까지 되돌릴 필요는 없었다.

### 4.4 spec manifest는 생각보다 큰 운영 이점을 준다

manifest가 있으면 다음이 쉬워진다.

- smoke에 넣을 spec 교체
- quarantine 대상 추가/제거
- branch에 없는 spec 자동 skip
- PR review 시 selection diff 검토

특히 이번에는 `voc-auto-routing.spec.ts` 가 어떤 baseline에는 있고 어떤 baseline에는 없는 상황이 있었는데, selector가 안전하게 흡수했다.

## 5. 다음번 동일 증상 대응 절차

비슷한 CI failure가 다시 나오면 아래 순서로 대응하는 것이 좋다.

1. 실패가 `required lane`인지 `regression/quarantine lane`인지 먼저 분리한다.
2. flaky 성격이면 spec 자체와 fixture의 retry / redirect / timeout 취약점을 먼저 본다.
3. 이미 base branch에 더 안정적인 CI 구조가 있다면, 실패 PR에 최신 `main`을 반영해 새 SHA로 다시 평가한다.
4. old run을 붙잡고 해석하기보다 새 구조 기준 run을 확보한 뒤 판단한다.
5. 기능 실패와 flaky를 한 번에 고치려 하지 말고, lane 분리와 root cause fix를 분리한다.

## 6. 후속 권장 사항

- branch protection required check 목록을 문서화된 목표 상태와 다시 맞춘다.
- quarantine spec의 pass/fail 추세를 주기적으로 정리한다.
- `full frontend tsc --noEmit` 선행 오류는 별도 정리 과제로 분리한다.
- PR #579는 현재 check clean 상태이므로, 코드 리뷰 결과에 따라 merge 또는 추가 후속 판단만 남는다.

## 7. 결론

이번 CI 재구성은 다음 세 가지를 동시에 달성했다.

- required CI를 더 빠르고 안정적인 구조로 바꿨다.
- flaky signout/login 계열의 복구 가능성을 실제로 높였다.
- 이미 실패했던 PR도 최신 `main` 반영 후 새 구조에서 다시 통과시킬 수 있음을 운영적으로 검증했다.

앞으로는 “old CI failure를 rerun만 반복”하기보다, **stable main의 CI 구조를 PR branch에 적용해 새 SHA 기준으로 재평가**하는 방식이 기본 복구 패턴이 되어야 한다.
