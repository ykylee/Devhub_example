# Session Handoff — codex/work_260521-c-db-docker-option

- 문서 목적: host build + runtime-only Docker packaging 전환 상태와 외부 접속/VM ingress 분리 결과를 인계한다.
- 범위: `scripts/build-artifacts.sh`, `deploy-from-env.sh`, runtime-only Dockerfiles, nginx 23000 정합, Keycloak dev/master realm `sslRequired`, backend-core `ca-certificates`, frontend `public`, main rebase
- 대상 독자: 후속 에이전트, 리뷰어
- 상태: in_progress
- 최종 수정일: 2026-05-22

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
- `deploy-from-env.sh` 에 `PUBLIC_ACCESS_SCHEME/HOST/PORT` 를 추가해 외부 접속 기준 주소를 VM ingress 포트와 분리했다. 기본값은 `http://100.90.113.29:13000` + `NGINX_HTTP_PORT=3000` 이다.
- `setup-keycloak.sh` 를 다시 돌려 `devhub-e2e-seeder` client 와 secret 을 재발급했다.
- Playwright e2e 는 Keycloak 시드 단계까지 통과했지만, `idp-apply-schemas` 가 host 에서 `db` 호스트명을 해석하지 못해 실패했다. 이건 SSL/TLS 문제와 별개의 host-run DB 접근 문제다.

## 2026-05-22 rebase 세션

- `origin/main` (`1239f3c`) 위로 rebase 완료.
- PR #282 (`chore(deploy): local-db compose profile 지원 추가`) 가 이 브랜치 초기 3 commit (`6ed6d51`, `ef244a0`, `f08360f`) 을 squash 흡수한 상태였음 → rebase 시 drop.
- 잔여 12 commit replay 완료, 새 HEAD `61e0937` (refactor(deploy): split external access from vm ingress).
- 충돌 surface: memory 파일 4건 (skip 으로 처리) + deploy 스크립트 (replay 시 자동 처리됨).
- `git push --force-with-lease` 로 원격 갱신 완료.
- 백업 ref `backup/pre-rebase-main-260522` (구 HEAD `3385bec`) 로컬 보존.

## 2026-05-22 nginx conf auto-sync 세션 (`3ae46cf`)

- 외부 `:13000` 접근이 `:80/devhub` 로 redirect 되던 회귀 원인 = stale `infra/nginx/devhub.deploy.conf` (template 과 미동기화, `/devhub/*` routing 누락 + `X-Forwarded-Port 23000` stale).
- docker-compose.deploy.yml 은 `.template` 만 mount 하므로 docker deploy 자체는 영향 없으나, 외부 host nginx 등 정적 참조 path 에서 stale 가 적용됐을 가능성.
- 조치:
  - `scripts/nginx-conf-sync.sh` 신규 (envsubst render, `--check`/`--fix` 모드, KEYCLOAK_UPSTREAM/KEYCLOAK_ADMIN_ALLOW_CIDR placeholder 기본값, template SHA-256 + auto-gen banner).
  - `scripts/deploy-preflight.sh` 가 `[0/3]` 단계에서 `--fix` 자동 실행 → drift silent 흡수.
  - `.gitignore` 에 `infra/nginx/devhub.deploy.conf` 추가 (derived artifact).
  - 기존 stale `devhub.deploy.conf` git rm.
- 검증: `absolute_redirect off` + 모든 `return 302` relative + `location = /devhub` → `/devhub/developer` + `/devhub/api/` rewrite 모두 보존.

## 2026-05-22 Keycloak 25.0 pin 세션 (`42b18b1`)

- 운영 image 를 `quay.io/keycloak/keycloak:26.0` → `25.0` 으로 retreat.
- active code 3 위치만 변경: `docker-compose.deploy.yml:106`, `.github/workflows/ci.yml:340`, `dev-up.sh:118`.
- docs (ADR-0019/0020/0021 등) 의 historical 26.0 언급은 immutable 보존.
- ADR-0022 draft 발행 — §3.1 retreat 사유 placeholder, 사용자 finalize 후 Accepted 승격.
- `latest` tag 미사용 정책 재확인 (4 risk — 재현 불가 / silent upgrade / rollback 소실 / Keycloak 호환 risk).

## 다음 세션 첫 작업

1. **사용자 사내 환경 redeploy + :13000 smoke test** — preflight 가 nginx conf 자동 sync + Keycloak 25.0 pull. 아래 명령 참조.
2. **ADR-0022 §3.1 retreat 사유 finalize** + Accepted 승격.
3. **BUILD-ARTIFACT-08** — host-run e2e 의 `db` 호스트명 해석 분리.
4. 이 브랜치 PR 생성 또는 main 직접 머지 결정.

## 사내 환경 redeploy + smoke test 절차

```bash
# 1. latest pull
git pull --ff-only

# 2. env 설정 (사내 값으로 override)
export IMAGE_TAG=$(git rev-parse --short HEAD)
export IMAGE_REPO_PREFIX=local/devhub
export PUBLIC_ACCESS_HOST=<사내 외부 IP>
export PUBLIC_ACCESS_PORT=13000
export NGINX_HTTP_PORT=3000
# 기타 secret 은 env 또는 deploy.env 파일

# 3. preflight (자동으로 nginx conf 재생성 — [0/3] 단계 출력 확인)
bash scripts/deploy-preflight.sh

# 4. 전체 build + deploy
bash scripts/deploy-from-env.sh
# 또는 build 만 다시: ACTION=build bash scripts/deploy-from-env.sh
# 또는 deploy 만: ACTION=deploy bash scripts/deploy-from-env.sh

# 5. smoke test (외부에서 — 브라우저 / curl)
curl -vI -L --max-redirs 3 http://<사내 외부 IP>:13000/
# 기대: 302 /devhub → 302 /devhub/developer → 200 (Next.js)
# 실패 시: Location 헤더 :80 여부 확인 → nginx-conf-sync 재실행

# 6. Keycloak 25.0 정합 검증
docker compose -f docker-compose.deploy.yml ps keycloak
docker compose -f docker-compose.deploy.yml logs keycloak | grep -i "version\|started"
curl -fsS http://<사내 외부 IP>:13000/devhub/auth/keycloak/realms/devhub/.well-known/openid-configuration | jq .issuer
```

## 검증 체크리스트

- [ ] `:13000/` → `:13000/devhub/developer` redirect chain 정상 (포트 13000 보존)
- [ ] Next.js dashboard 로그인 페이지 표시
- [ ] Keycloak realm 정상 import (Keycloak admin console 진입 + `devhub` realm 존재 확인)
- [ ] OIDC discovery endpoint 200 응답
- [ ] backend /health 정상
- [ ] frontend /api/runtime-config 정상
