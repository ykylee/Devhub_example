# Work Backlog — codex/work_260521-c-db-docker-option

- 문서 목적: 현재 브랜치의 workflow 작업 항목과 상태를 관리한다.
- 범위: host build packaging + nginx ingress 정합
- 대상 독자: 구현 담당자, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-05-21

## 백로그

| ID | 작업 | 상태 | 비고 |
| --- | --- | --- | --- |
| BUILD-ARTIFACT-01 | host build 스크립트 추가 | done | `scripts/build-artifacts.sh` |
| BUILD-ARTIFACT-02 | runtime-only Dockerfile 로 전환 | done | backend-core/backend-ai/frontend |
| BUILD-ARTIFACT-03 | deploy wrapper 에 host build 연결 | done | `deploy-from-env.sh` |
| BUILD-ARTIFACT-04 | nginx 23000 HTTP redirect 제거 | done | `absolute_redirect off` 포함 |
| BUILD-ARTIFACT-05 | e2e Keycloak HTTPS 요구 원인 정리 | done | `sslRequired=none` + master realm 갱신 |
| BUILD-ARTIFACT-06 | compose nginx TLS 제거 | done | 443 포트/인증서 마운트 삭제 |
| BUILD-ARTIFACT-07 | backend-core Go 버전 1.25.9 고정 | done | `go.mod` + docs 정합 |
| BUILD-ARTIFACT-08 | host-run e2e 의 DB 접근 경로 정리 | in_progress | `db` 호스트명 host 해석 실패 |
| BUILD-ARTIFACT-09 | backend-core Alpine ca-certificates 제거 | in_progress | HTTP-only 기준 프록시 의존 제거 |
| BUILD-ARTIFACT-10 | frontend public copy 제거 | in_progress | 빈 public 디렉터리 정리 |
| BUILD-ARTIFACT-11 | deploy-from-env.sh push 액션 제거 | done | 로컬 전용 build/deploy 흐름 단순화 |
| BUILD-ARTIFACT-12 | 외부 접속 기준 주소/포트 분리 | in_progress | `PUBLIC_ACCESS_*` + `NGINX_HTTP_PORT=3000` |
