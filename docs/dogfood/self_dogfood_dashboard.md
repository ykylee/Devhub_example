# Self Dogfood Dashboard 시나리오

- 문서 목적: self dogfooding으로 생성한 platform / project / repository 결과를 현재 대시보드 구현에서 확인하는 절차를 정리한다.
- 범위: 플랫폼 대시보드, 프로젝트 대시보드, 위젯 렌더링 확인
- 대상 독자: QA, 개발자, AI 에이전트
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [테스트 시나리오](./test_scenarios.md), [self dogfood admin 시나리오](./self_dogfood_admin.md)

## 1. 목적

이 시나리오는 self dogfooding 생성 흐름이 끝난 뒤, 실제 UI 대시보드에서 결과가 보이는지 확인한다. 현재 구현 기준으로 `/admin` 홈은 아카이브 상태이므로, 확인 대상은 다음 두 상세 페이지다.

1. `/platforms/:id`
2. `/projects/:id`

## 2. 실행 명령

```sh
./scripts/dogfood.sh test-self-dogfood-dashboard
```

이 명령은 내부적으로 다음을 수행한다.

1. `smoke`
2. Playwright spec `frontend/tests/e2e/dogfood-self-dogfood-dashboard.spec.ts`

## 3. 검증 포인트

### 3.1 플랫폼 대시보드

- 플랫폼 heading 노출
- `Build & Quality 7-Day Trend` 위젯 노출
- `Linked Projects Roadmap` 위젯 노출
- `Repositories` 위젯 노출
- self dogfood 로 만든 프로젝트명과 저장소 slug 확인

### 3.2 프로젝트 대시보드

- 프로젝트 heading 노출
- `Connected Repositories` 위젯 노출
- `Recent Activity` 위젯 노출
- `Active Tasks` 위젯 노출
- `Linked Repositories` 위젯 노출
- self dogfood 로 만든 저장소 slug 확인

## 4. 해석 기준

- 이 시나리오는 대시보드 데이터가 풍부한지를 검증하는 것이 아니라, 현재 구현된 위젯이 오류 없이 렌더링되고 self dogfood 생성 결과를 참조할 수 있는지를 검증한다.
- build trend, recent activity, active task 는 초기 데이터가 적어도 섹션 자체가 보이면 정상으로 본다.
