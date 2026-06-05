# Dogfood Self-Dogfood Admin 시나리오

- 문서 목적: `system_admin` 계정으로 접속해 현재 워크스페이스를 대표하는 플랫폼, 저장소, 프로젝트를 하나씩 만드는 self dogfooding 시나리오를 정의한다.
- 범위: 관리자 로그인, 플랫폼 생성, 저장소 draft 생성, 플랫폼-저장소 연결, 프로젝트 생성, 상세 화면 검증
- 대상 독자: 개발자, QA, AI 에이전트, 로컬 기능 검증 수행자
- 상태: draft
- 최종 수정일: 2026-06-05
- 관련 문서: [Dogfood 환경 문서](./README.md), [환경 셋업 가이드](./environment_setup.md), [테스트 시나리오](./test_scenarios.md), [dogfood-self-dogfood-admin.spec.ts](../../frontend/tests/e2e/dogfood-self-dogfood-admin.spec.ts)

## 1. 시나리오 목표

이 시나리오는 관리자 입장에서 현재 DevHub 예제 워크스페이스를 시스템 안에 다시 등록하는 흐름을 검증한다.

1. `system_admin` 으로 로그인한다.
2. 현재 프로젝트를 대표하는 플랫폼을 만든다.
3. `gitea` provider 를 사용하는 DevHub 저장소 draft 를 만든다.
4. 그 저장소를 플랫폼에 연결한다.
5. 연결된 저장소를 참조하는 프로젝트를 만든다.
6. 플랫폼 상세와 프로젝트 상세에서 결과를 확인한다.

## 2. 설계 원칙

- 외부 Gitea 에 실제 새 저장소를 publish 하지 않는다.
- 대신 DevHub 내부의 repository draft 를 생성해 반복 실행에 따른 외부 상태 오염을 피한다.
- 이름과 key 는 매 실행마다 unique suffix 를 붙여 충돌을 피한다.

## 3. 생성되는 자산 형태

예시:

- Platform name: `DevHub Example Codex 123456`
- Repository slug: `yklee/devhub-example-codex-dogfood-123456`
- Project name: `Self Dogfood Project 123456`

## 4. 실행 명령

```sh
./scripts/dogfood.sh test-self-dogfood
```

이 명령은 내부적으로 다음을 수행한다.

- `./scripts/dogfood.sh smoke`
- `frontend/tests/e2e/dogfood-self-dogfood-admin.spec.ts` 실행

## 5. 기대 결과

정상 실행 시 다음을 확인한다.

1. 플랫폼 생성 성공
2. 저장소 draft 생성 성공
3. 플랫폼-저장소 연결 성공
4. 프로젝트 생성 성공
5. 플랫폼 상세 페이지에서 플랫폼 이름 확인
6. 프로젝트 상세 페이지에서 프로젝트 이름과 저장소 slug 확인

## 6. 운영 메모

- 이 시나리오는 DevHub DB 안에 platform / repository / project row 를 남긴다.
- 외부 Gitea 에 새 repository 를 publish 하지는 않는다.
- 완전히 초기 상태로 되돌리고 싶으면 아래 순서를 사용한다.

```sh
./scripts/dogfood.sh reset-db
./scripts/dogfood.sh up
```
