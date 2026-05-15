# 작업 백로그 인덱스

- 문서 목적: `codex/docker-packaging-guide` 브랜치의 작업 항목과 백로그 기록을 관리한다.
- 범위: 태스크 상태 요약, 날짜별 백로그 링크
- 대상 독자: 개발자, AI 에이전트, 프로젝트 매니저
- 상태: active
- 최종 수정일: 2026-05-15
- 관련 문서: [세션 인계](./session_handoff.md), [프로젝트 프로파일](../../PROJECT_PROFILE.md)

## 1. 운영 원칙

1. 세션 시작 시 현재 브랜치와 이 디렉터리 문서를 먼저 확인한다.
2. 세션 종료 전 `state.json`, `session_handoff.md`, 최신 backlog를 갱신한다.
3. flat memory 경로는 legacy fallback 및 공용 색인으로만 사용한다.

## 2. 날짜별 백로그

- [2026-05-15](./backlog/2026-05-15.md)

## 3. 작업 상태 요약

- [x] `TASK-DOCKER-PACKAGING-STRATEGY`: Docker 패키징/배포 전략 정리 및 권장안 확정
- [x] `TASK-DOCKER-DEPLOY-COMPOSE-TEMPLATE`: image pinning 기반 `docker-compose.deploy.yml` 추가
- [x] `TASK-CI-DOCKER-IMAGE-PUBLISH`: 이미지 3종 빌드/푸시 GitHub Actions workflow 추가
- [x] `TASK-ENV-SCHEMA-PUBLIC-INTERNAL-DB`: public/internal/db 환경변수 스키마 정리
- [x] `TASK-DEPLOY-COMPOSE-LOCALDB-PROFILE`: `local-db` profile 추가 및 외부 DB 모드 분리
- [x] `TASK-E2E-PERMISSIONS-FLAKY-FIX`: admin-permissions e2e 실패 2건 안정화

## 4. 다음 작업 후보

- 외부 DB 모드에서 전체 e2e 재검증
- `.env` 템플릿/시크릿 주입 가이드 분리
- 릴리즈 노트 digest 자동 기록
