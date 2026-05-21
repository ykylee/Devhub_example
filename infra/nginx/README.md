# DevHub nginx reverse-proxy 자산

- 문서 목적: DevHub 의 단일 외부 포트 reverse-proxy (ADR-0018) 의 nginx 설정 자산을 정리한다.
- 범위: compose 운영 (envsubst template) 과 native + host nginx 의 두 deploy mode, Keycloak 로컬/외부 모드 분기.
- 대상 독자: 운영자, 신규 합류자, ADR-0018 / ADR-0019 후속 carve 담당자.
- 상태: active
- 최종 수정일: 2026-05-20
- 관련 문서:
  - [ADR-0018 — 단일 외부 포트 역프록시 정책](../../docs/adr/0018-single-port-reverse-proxy-policy.md)
  - [ADR-0019 — Keycloak 단일 IdP](../../docs/adr/0019-keycloak-only-idp.md)
  - [docs/setup/single_port_deployment.md](../../docs/setup/single_port_deployment.md)
  - [docs/setup/keycloak_operations.md](../../docs/setup/keycloak_operations.md)
  - [infra/idp/README.md](../idp/README.md) — Keycloak 로컬 / 외부 모드 분기

## 1. 파일 인벤토리

| 파일 | 용도 | mount 위치 |
| --- | --- | --- |
| `devhub.deploy.conf.template` | docker-compose.deploy.yml 의 nginx 컨테이너용. nginx 공식 이미지의 envsubst entrypoint 가 처리. `${KEYCLOAK_UPSTREAM}` / `${KEYCLOAK_ADMIN_ALLOW_CIDR}` 치환. | `/etc/nginx/templates/devhub.deploy.conf.template` (→ `/etc/nginx/conf.d/devhub.deploy.conf`) |
| `devhub.native.conf` | host nginx + 모든 서비스 native loopback (`127.0.0.1:*`). dev / staging 의 단일 포트 정합성 검증 용도. | `/etc/nginx/sites-available/devhub.native.conf` (host 직접 설치) |

## 2. compose 운영 mode (`devhub.deploy.conf.template`)

### 2.1 Keycloak 로컬 모드 (compose `local-idp` profile)
```bash
KEYCLOAK_UPSTREAM=keycloak:8080
KEYCLOAK_ADMIN_ALLOW_CIDR=10.0.0.0/8   # 사내 운영망만 admin 허용
KEYCLOAK_HOSTNAME=devhub.example.com
# ... 기타 환경변수 (docker-compose.deploy.yml 참조)

docker-compose -f docker-compose.deploy.yml \
  --profile local-idp --profile local-db \
  up -d
```
nginx 의 `${KEYCLOAK_UPSTREAM}` 가 compose 내부 service name 으로 해석되어 같은 docker network 의 keycloak 컨테이너로 reverse proxy.

### 2.2 Keycloak 외부 모드 (사내 운영 Keycloak)
```bash
# compose 의 keycloak / db / db-init profile 활성화 안 함
KEYCLOAK_UPSTREAM=kc.internal.example.com:8443   # 외부 Keycloak host:port
KEYCLOAK_ADMIN_ALLOW_CIDR=10.0.0.0/8
# ... 기타 환경변수

docker-compose -f docker-compose.deploy.yml up -d
# (--profile 옵션 미지정 → keycloak 컨테이너 미가동)
```
nginx 가 외부 Keycloak 으로 reverse proxy. 사용자 브라우저는 항상 `https://<devhub-host>/devhub/auth/keycloak/*` 단일 origin 만 알면 됨 — 단일 포트 컨셉 유지.

> **외부 Keycloak 의 hostname 별칭**: 외부 Keycloak 이 `/devhub/auth/keycloak/*` path prefix 가 아닌 다른 prefix (예: `/auth/*`) 로 서빙된다면 nginx 설정의 `proxy_pass http://devhub_keycloak;` 를 `proxy_pass http://devhub_keycloak/auth/;` 등으로 사내 환경에 맞춰 수정 필요. 사내 운영팀의 별도 nginx mount override 권장.

### 2.3 Keycloak Admin API 접근 제어
`${KEYCLOAK_ADMIN_ALLOW_CIDR}` 환경변수가 `/devhub/auth/keycloak/admin/*` path 의 외부 노출을 제어한다.
- 빈 값 또는 미설정: deny all (외부 admin API 호출 차단). backend 의 `DEVHUB_KEYCLOAK_ADMIN_URL` 는 internal direct (예: 로컬 모드의 `http://keycloak:8080/devhub/auth/keycloak` 또는 외부 모드의 internal-only Keycloak URL) 사용.
- 운영 CIDR (예: `10.0.0.0/8`): 사내 운영망에서만 admin REST API 접근 허용.

## 3. native + host nginx mode (`devhub.native.conf`)

ADR-0018 §3.3 의 단일 포트 dev 검증 용도. host 에 native nginx 가 설치되어 있고 backend / frontend / Keycloak 모두 host loopback (`127.0.0.1:*`) 에 native 로 가동 중일 때 사용.

```bash
# 1. dev-up.sh 로 backend (8080) + frontend (3000) + Keycloak (8180) 모두 native 가동
./dev-up.sh

# 2. host nginx 에 devhub.native.conf 적용
sudo cp infra/nginx/devhub.native.conf /etc/nginx/sites-available/
sudo ln -sf /etc/nginx/sites-available/devhub.native.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# 3. 브라우저로 http://localhost/devhub/ 진입 → 단일 포트 (80) 컨셉 검증
```

이 모드는 단일 포트 컨셉의 회귀 가드 SOP 로 활용한다 — PR 시점에 reverse-proxy 정합성을 native 환경에서 사전 검증.

## 4. SSL/TLS 인증서

- compose 모드: TLS 미사용. `docker-compose.deploy.yml` 은 HTTP `:80` 만 노출한다.
- native 모드: TLS 미사용. `devhub.native.conf` 역시 HTTP `:80` 만 사용한다.

## 5. 단일 포트 컨셉 정합 가드

본 nginx 자산이 외부에서 별도 포트로의 redirect 를 발생시키지 않는지 PR 시점에 확인:

```bash
# 모든 `return 30*` 은 path-relative 또는 same-host 여야 함 (절대 URL 의 host:port 점프 금지)
grep -n "return 30" infra/nginx/*.conf*
```

자세한 정합성 리뷰: [docs/reports/2026-05-20-network-docker-single-port-review.md](../../docs/reports/2026-05-20-network-docker-single-port-review.md).
