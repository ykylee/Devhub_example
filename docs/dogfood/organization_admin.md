# Organization Admin 시나리오

- 문서 목적: dogfood 환경에서 `system_admin` 이 조직 단위를 생성, 수정, 멤버 변경, 리더 변경, 삭제하는 흐름을 검증한다.
- 범위: `/admin/settings/organization` 리스트 뷰 기준 조직 단위 관리
- 대상 독자: QA, 개발자, AI 에이전트
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [테스트 시나리오](./test_scenarios.md), [환경 셋업 가이드](./environment_setup.md)

## 1. 목적

이 시나리오는 관리자 조직 관리 화면이 실제 운영 흐름에서 필요한 핵심 작업을 끝까지 처리할 수 있는지 확인한다.

1. 조직 단위 생성
2. 조직 단위명 수정
3. 리더 변경
4. 멤버 추가 및 변경
5. 조직 단위 삭제

## 2. 실행 명령

```sh
./scripts/dogfood.sh test-organization-admin
```

이 명령은 내부적으로 다음을 수행한다.

1. `smoke`
2. Playwright spec `frontend/tests/e2e/dogfood-organization-admin.spec.ts`

## 3. 검증 포인트

- `/admin/settings/organization` 진입 가능
- `Create Unit` 모달에서 새 조직 단위 생성
- `Edit` 액션에서 이름과 리더 수정
- `Members` 액션에서 roster 변경과 leader 재지정
- 멤버 수 변경이 리스트에 반영
- `Delete` 액션으로 생성한 조직 단위 삭제

## 4. 현재 기준 시나리오 데이터

- parent unit: `Engineering`
- 최초 leader: `charlie`
- 수정 후 leader: `bob`
- member management leader: `alice`
- member roster 변경: `alice`, `bob` 추가 후 `bob` 제거

## 5. 실패 시 점검 순서

1. `./scripts/dogfood.sh smoke` 통과 여부 확인
2. `/admin/settings/organization` 에서 조직 데이터 초기 로딩 실패 여부 확인
3. `frontend` 로그와 `backend-core` 로그에서 organization API 에러 확인
4. 시드 사용자 `alice`, `bob`, `charlie` 가 정상 존재하는지 확인
