# Gitea Integration Admin 시나리오

- 문서 목적: dogfood 환경에서 `system_admin` 이 외부 Gitea provider 를 등록하고 sync 하는 검증 절차를 정리한다.
- 범위: provider 생성, sync 요청, 원격 저장소 목록 확인, cleanup
- 대상 독자: QA, 개발자, AI 에이전트
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [테스트 시나리오](./test_scenarios.md), [환경 셋업 가이드](./environment_setup.md)

## 1. 목적

이 시나리오는 dogfood 환경에서 외부 `Gitea` 연동이 실제로 살아 있는지 빠르게 확인하는 관리자용 smoke 다. 핵심은 다음 네 가지다.

1. `system_admin` 세션으로 integration provider 를 등록할 수 있다.
2. 등록한 provider 로 sync job 을 요청할 수 있다.
3. provider API 를 통해 원격 저장소 목록을 조회할 수 있다.
4. 테스트가 끝난 뒤 임시 provider 를 정리할 수 있다.

## 2. 실행 명령

```sh
./scripts/dogfood.sh test-integration-admin
```

이 명령은 내부적으로 다음을 수행한다.

1. `smoke`
2. Playwright spec `frontend/tests/e2e/dogfood-gitea-integration-admin.spec.ts`

## 3. 검증 포인트

- `/admin/settings/integrations` 접근 가능
- 신규 provider row 생성 확인
- `POST /api/v1/integration/providers/:provider_id/sync` 가 `accepted` 반환
- `GET /api/v1/integration/providers/:provider_id/scm-repositories` 결과에 최소 1개 이상 저장소 존재
- 현재 기준 기대 원격 저장소 예시: `yklee/devhub-simulation`
- 테스트 종료 후 생성한 provider 삭제

## 4. 실패 시 점검 순서

1. `.env.dogfood` 의 `GITEA_URL`, `GITEA_TOKEN`, `GITEA_WEBHOOK_SECRET` 확인
2. `./scripts/dogfood.sh smoke` 에서 Gitea check 통과 여부 확인
3. 외부 Gitea 서버 상태와 PAT 권한 재확인
4. 필요 시 `/admin/settings/integrations` 화면에서 row 상태와 backend 로그 함께 확인
