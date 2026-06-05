# Dogfood 온보딩 Smoke 실행 가이드

- 문서 목적: dogfood 로컬 시뮬레이션 환경에서 신규 계정 1개를 생성하고 온보딩을 끝까지 검증하는 절차를 기록한다.
- 범위: 계정 생성, 앱 기동, Playwright smoke 실행, 기대 결과, 정리 방법
- 대상 독자: 개발자, QA, AI 에이전트, 기능 검증 수행자
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [환경 셋업 가이드](./environment_setup.md), [테스트 시나리오](./test_scenarios.md), [dogfood-create-user.sh](../../scripts/dogfood-create-user.sh), [dogfood-onboarding-smoke.spec.ts](../../frontend/tests/e2e/dogfood-onboarding-smoke.spec.ts)

## 1. 목적

이 smoke 는 다음 질문에 빠르게 답하기 위한 것이다.

1. `dogfood.sh up` 으로 macOS + `colima` 환경이 정상 기동되는가
2. Keycloak 신규 사용자가 첫 로그인 시 `/onboarding` 으로 진입하는가
3. 조직 선택과 제출이 성공하고, 이후 `/developer` 로 라우팅되는가
4. `/api/v1/me` 에서 `onboarding_required=false`, `review_status=pending_review` 가 확인되는가

## 2. 사전 조건

dogfood 환경이 이미 준비돼 있어야 한다.

```sh
./scripts/dogfood.sh up
```

최소 헬스 체크:

```sh
curl http://localhost:18080/health
curl http://localhost:18000/health
curl -I http://localhost:13000/
curl http://localhost:18180/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration
```

## 3. 수동 계정 생성

브라우저에서 직접 테스트하고 싶다면 Keycloak 계정을 먼저 만든다.

```sh
./scripts/dogfood-create-user.sh dogfood-manual@example.com 'ChangeMe-12345!' 'Dogfood Manual'
```

이 스크립트는 다음만 수행한다.

- Keycloak 사용자 생성 또는 갱신
- 비밀번호 설정

첫 로그인 전에는 DevHub DB `users` row 가 없다. 따라서 로그인 후 `/onboarding` 으로 이동하는 것이 정상이다.

## 4. 자동 온보딩 smoke 실행

자동 검증은 Playwright spec 으로 기록돼 있다.

- 대상 파일: `frontend/tests/e2e/dogfood-onboarding-smoke.spec.ts`
- 시나리오 요약:
  - Keycloak 에 신규 사용자 생성
  - `/login` → Keycloak 로그인
  - `/onboarding` 진입 확인
  - 이름 입력
  - `Engineering (dept-eng)` 검색/선택
  - 제출 후 `/developer` 이동 확인
  - `/api/v1/me` 에서 `pending_review` 확인

실행 명령:

```sh
./scripts/dogfood.sh test-onboarding
```

이 명령은 내부적으로 다음을 수행한다.

- `./scripts/dogfood.sh smoke`
- `frontend/tests/e2e/dogfood-onboarding-smoke.spec.ts` 실행

## 5. 기대 결과

정상 실행 시 다음이 확인돼야 한다.

1. 신규 사용자가 `/onboarding` 으로 이동한다.
2. 이메일 필드는 SSO 값으로 자동 채워지고 수정 불가다.
3. 조직 검색 `Eng` 로 `Engineering / dept-eng` 를 선택할 수 있다.
4. 제출 후 `/developer` 로 이동한다.
5. `/api/v1/me` 응답에서 아래 상태가 보인다.

```json
{
  "onboarding_required": false,
  "primary_unit_id": "dept-eng",
  "review_status": "pending_review"
}
```

## 6. 운영 메모

- 이 smoke spec 은 실행할 때마다 새 Keycloak 사용자를 만든다.
- DevHub DB 에는 `pending_review` 상태의 사용자 row 가 남는다.
- 테스트 흔적을 완전히 지우고 싶으면 아래처럼 초기화한다.

```sh
./scripts/dogfood.sh reset-db
./scripts/dogfood.sh up
```

## 7. 수동 확인 체크리스트

브라우저 수동 검증 시에는 아래 순서로 보면 된다.

1. `http://localhost:13000/login` 접속
2. 방금 만든 계정으로 로그인
3. `/onboarding` 진입 확인
4. 이름 입력
5. `Engineering` 검색 후 선택
6. 제출
7. `/developer` 이동 확인
8. 필요 시 네트워크 탭 또는 `GET /api/v1/me` 로 `pending_review` 확인
