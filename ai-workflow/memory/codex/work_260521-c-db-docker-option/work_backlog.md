# Work Backlog — codex/work_260521-c-db-docker-option

- 문서 목적: 현재 브랜치의 workflow 작업 항목과 상태를 관리한다.
- 범위: DB docker 옵션 추가 및 배포 wrapper 정합
- 대상 독자: 구현 담당자, 리뷰어
- 상태: done
- 최종 수정일: 2026-05-21

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| DB-DOCKER-01 | `deploy-from-env.sh` 에 `DB_MODE=docker` 추가 | done | local-db profile 활성화 |
| DB-DOCKER-02 | env 파일 shell-safe 출력으로 수정 | done | 공백 포함 값 대응 |
| DB-DOCKER-03 | preflight/up 에 compose profiles 전달 | done | `COMPOSE_PROFILES=local-db` |
| DB-DOCKER-04 | 브랜치 memory 기록/커밋/PR | done | `c3cb37d` |
