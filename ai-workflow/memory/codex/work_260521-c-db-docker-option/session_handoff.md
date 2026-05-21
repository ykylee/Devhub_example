# Session Handoff — codex/work_260521-c-db-docker-option

- 문서 목적: host build + runtime-only Docker packaging 전환 상태와 로컬 전용 deploy 스크립트 정리 결과를 인계한다.
- 범위: `scripts/build-artifacts.sh`, `deploy-from-env.sh`, runtime-only Dockerfiles, nginx 23000 정합, Keycloak dev/master realm `sslRequired`, backend-core `ca-certificates`, frontend `public`
- 대상 독자: 후속 에이전트, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-05-21

## 이번 세션 요약

- `scripts/build-artifacts.sh` 를 추가해 backend-core, backend-ai, frontend 를 도커 밖에서 먼저 빌드하도록 바꿨다.
- `deploy-from-env.sh` 가 host build 결과물을 만든 뒤 runtime-only Docker image 를 패키징하도록 연결했다.
- `backend-core`, `backend-ai`, `frontend` Dockerfile 을 빌드 스테이지 없이 결과물만 복사하는 방식으로 전환했다.
- nginx 80 서버 블록의 HTTP → HTTPS redirect 를 제거하고 `/devhub/*` 를 직접 프록시하도록 정리했다.
- compose/native nginx 에서 443/TLS 포트와 인증서 마운트를 제거해 HTTP only 로 맞췄다.
- `absolute_redirect off` 를 넣어 `/devhub` 리다이렉트가 포트를 잃지 않도록 조정했다.
- `infra/idp/keycloak-realm.dev.json` 와 `infra/idp/keycloak-realm.prod.json` 의 `sslRequired` 를 `none` 으로 완화했다.
- Keycloak 컨테이너 내부에서 `kcadm.sh` 로 `master` realm 도 `sslRequired=none` 으로 갱신했다.
- `backend-core/go.mod` 의 Go 기준 버전을 `1.25.9` 로 맞췄다.
- `backend-core/Dockerfile` 에서 `apk add --no-cache ca-certificates` 를 제거해 사내 프록시망 / Alpine 패키지 미러 의존을 없앴다.
- `frontend/public` 가 비어 있어 `frontend/Dockerfile` 의 `COPY public ./public` 를 제거했다.
- `deploy-from-env.sh` 에서 `ACTION=push` 분기를 제거해 로컬 전용 흐름을 `build|deploy|all` 로 단순화했다.
- `setup-keycloak.sh` 를 다시 돌려 `devhub-e2e-seeder` client 와 secret 을 재발급했다.
- Playwright e2e 는 Keycloak 시드 단계까지 통과했지만, `idp-apply-schemas` 가 host 에서 `db` 호스트명을 해석하지 못해 실패했다. 이건 SSL/TLS 문제와 별개의 host-run DB 접근 문제다.

## 다음 세션 첫 작업

1. 이 변경분을 커밋하고 브랜치 상태를 갱신한다.
2. 필요하면 HTTPS outbound 가 필요한 경우에만 별도 CA 번들 전략을 도입한다.
3. 필요하면 host-run e2e 가 DB 에 접근할 수 있도록 DSN/포트 노출 방식을 분리한다.
