# Session Handoff — codex/work_260521-c-db-docker-option

- 문서 목적: `deploy-from-env.sh` 에 DB docker 옵션을 추가한 작업 상태를 인계한다.
- 범위: DB_MODE=docker, COMPOSE_PROFILES=local-db, env 파일 shell-safe 출력
- 대상 독자: 후속 에이전트, 리뷰어
- 상태: done
- 최종 수정일: 2026-05-21

## 이번 세션 요약

- `scripts/deploy-from-env.sh` 에 `DB_MODE=docker` 를 추가했다.
- docker DB 모드에서는 `db` / `db-init` 를 `local-db` profile 로 함께 다루도록 `deploy-preflight.sh` 와 `deploy-up.sh` 를 정리했다.
- env 파일을 shell-safe 하게 출력하도록 바꿔 공백 포함 값도 안정적으로 source 되게 했다.
- `docker build` 가 실행 위치에 덜 민감하도록 각 서비스의 `Dockerfile` 을 `-f` 로 명시했다.
- `IMAGE_REPO_PREFIX` 기본값을 `local/devhub` 로 바꿔 로컬 전용 빌드가 더 쉽게 되도록 정리했다.
- `DB_MODE=docker` 빌드 경로와 `deploy-preflight.sh` 렌더를 검증했다.
- 커밋 `c3cb37d` 로 정리했다.

## 다음 세션 첫 작업

1. 변경분 푸시/PR 업데이트 확인.
2. 필요하면 `ACTION=deploy DB_MODE=docker` 로 실제 compose up 까지 확인.
