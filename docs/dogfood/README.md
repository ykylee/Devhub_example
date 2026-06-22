# Dogfood 환경 문서

- 문서 목적: DevHub 의 dogfood 전용 로컬 시뮬레이션 환경 문서를 한곳에 모아 진입점을 제공한다.
- 범위: 컨테이너 분리 원칙, 환경 셋업 가이드, 테스트 시나리오 문서 링크
- 대상 독자: 개발자, QA, AI 에이전트, 로컬 시뮬레이션 운영자
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [환경 셋업 가이드](./environment_setup.md), [테스트 시나리오](./test_scenarios.md), [온보딩 Smoke 실행 가이드](./onboarding_smoke.md), [E2E Test Guide](../setup/e2e-test-guide.md), [통합 테스트 시나리오 카탈로그](../planning/integrated_test_scenarios.md)

## 1. 개요

`docs/dogfood/` 는 기존 개발 환경과 충돌하지 않도록 분리된 **dogfood 전용 로컬 시뮬레이션 환경**을 다룬다. 이 문서 세트는 다음 두 가지를 목표로 한다.

1. `colima` 위에 별도 `PostgreSQL + Keycloak` 스택을 구성해 개발용 컨테이너와 병행 운영할 수 있게 한다.
2. 외부 `Gitea` 를 연결한 상태로 실제 사용자 흐름에 가까운 기능 검증 시나리오를 재현할 수 있게 한다.

## 2. 문서 구성

- [environment_setup.md](./environment_setup.md)
  - dogfood 전용 포트 체계
  - 저장소의 `docker-compose.colima.yml` + 로컬 `.env.dogfood` 기반 기동 절차
  - `backend-core` / `frontend` native 실행 연결 방법 (2026-06-22 M-v0.2.2 backend-ai 폐기 반영)
- [test_scenarios.md](./test_scenarios.md)
  - smoke / 운영 확인 / 통합 시나리오
  - 역할별 로그인, Gitea provider 연결, 플랫폼/프로젝트/저장소 생성, CI 대시보드 검증
- [onboarding_smoke.md](./onboarding_smoke.md)
  - 신규 Keycloak 계정 생성
  - `/onboarding` 실제 제출
  - Playwright smoke 재현 명령과 정리 방법
- [self_dogfood_admin.md](./self_dogfood_admin.md)
  - `system_admin` 기준 self dogfooding 시나리오
  - 플랫폼, 저장소 draft, 프로젝트 생성 검증
- [gitea_integration_admin.md](./gitea_integration_admin.md)
  - `system_admin` 기준 외부 Gitea provider 등록과 sync 검증
  - 원격 저장소 목록 확인과 정리 기준
- [organization_admin.md](./organization_admin.md)
  - `system_admin` 기준 조직 단위 생성, 수정, 리더/멤버 변경, 삭제 검증
  - `/admin/settings/organization` 운영 흐름 정리
- [repository_dashboard.md](./repository_dashboard.md)
  - 저장소 상세 대시보드의 developer / manager 시점 검증
  - build logs modal 과 manager view 토글 검증
- [self_dogfood_dashboard.md](./self_dogfood_dashboard.md)
  - self dogfooding 결과를 플랫폼/프로젝트 대시보드에서 확인
  - 현재 구현된 위젯과 검증 포인트 정리

## 3. 분리 원칙

- **Compose project 분리**: dogfood 스택은 `devhub-dogfood` compose project 로 실행한다.
- **포트 분리**: dogfood 는 개발 기본 포트(`3000/8080/8000/8180/5433`)를 사용하지 않는다.
- **비밀값 분리**: Gitea PAT, Keycloak admin client secret, webhook secret 은 `.env.dogfood` 에만 보관하고 git 에 커밋하지 않는다.
- **운영 자산 분리**: dogfood 컨테이너 이름은 `devhub-dogfood-*` 형태로 생성되어 기존 `devhub-*` 컨테이너와 공존한다.

## 4. 현재 기준 포트

| 구성 요소 | dogfood 포트 |
| --- | --- |
| Frontend | `13000` |
| Backend Core | `18080` |
| Backend AI | `18000` |
| Keycloak | `18180` |
| PostgreSQL | `15433` |

## 5. 빠른 진입

### 5.1 한 번에 켜기

```sh
./scripts/dogfood.sh up
```

이미지 재빌드가 정말 필요할 때만:

```sh
./scripts/dogfood.sh up --build
```

### 5.2 상태 확인

```sh
./scripts/dogfood.sh status
```

### 5.3 한 번에 끄기

```sh
./scripts/dogfood.sh down
```

### 5.4 처음부터 다시 시작하기

```sh
./scripts/dogfood.sh reset-db
./scripts/dogfood.sh up
```

런타임 로그와 PID 흔적까지 포함해 완전히 비우려면:

```sh
./scripts/dogfood.sh reset-all
```

### 5.5 로그 확인

```sh
./scripts/dogfood.sh logs
./scripts/dogfood.sh logs backend
```

### 5.6 빠른 검증

헬스 + 외부 Gitea 연결성만 빠르게 확인:

```sh
./scripts/dogfood.sh smoke
```

신규 계정 온보딩 Playwright smoke까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-onboarding
```

관리자 Gitea provider 등록 + sync 시나리오까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-integration-admin
```

관리자 조직 관리 시나리오까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-organization-admin
```

저장소 상세 대시보드 시나리오까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-repository-dashboard
```

관리자 self dogfooding 시나리오까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-self-dogfood
```

self dogfooding 결과를 대시보드에서 확인하는 시나리오까지 한 번에 실행:

```sh
./scripts/dogfood.sh test-self-dogfood-dashboard
```

### 5.7 다음 단계

컨테이너 기동 후에는 [환경 셋업 가이드](./environment_setup.md) 의 §4 를 따라 native 앱 3종을 붙이고, [테스트 시나리오](./test_scenarios.md) 의 순서대로 검증을 진행한다.
