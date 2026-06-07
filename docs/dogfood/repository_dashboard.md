# Repository Dashboard 시나리오

- 문서 목적: 원격 `main`에 추가된 저장소 상세 대시보드를 dogfood 환경에서 바로 검증할 수 있게 한다.
- 범위: developer / manager 시점 저장소 대시보드, build logs modal, manager 탭 토글
- 대상 독자: QA, 개발자, AI 에이전트
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [테스트 시나리오](./test_scenarios.md)

## 1. 목적

이 시나리오는 새 `repository dashboard`가 dogfood 환경에서도 그대로 동작하는지 확인한다. 검증 포인트는 다음 두 가지다.

1. developer 시점: build/integration 중심 뷰 + build log modal
2. manager 시점: manager & governance 뷰 + contributor distribution 토글

## 2. 실행 명령

```sh
./scripts/dogfood.sh test-repository-dashboard
```

이 명령은 내부적으로 다음을 수행한다.

1. `smoke`
2. Playwright spec `frontend/tests/e2e/repository-dashboard.spec.ts`

## 3. 검증 포인트

- `e2e-repo-a` 저장소 상세 진입 가능
- developer 시점에서 `Build & Integration Runs`, `Static Analysis (SonarQube)` 노출
- failed build 기준 `View Logs` modal 동작
- manager 시점에서 `Manager & Governance` 탭 동작
- `Contributor Distribution` 숨김/복원 토글 동작

## 4. 현재 검증 메모

- 이 스펙은 build-runs/logs API를 deterministic mock으로 감싼다.
- 저장소 seed 는 `frontend/tests/e2e/global-setup.ts` 기준 fixture 를 사용한다.
