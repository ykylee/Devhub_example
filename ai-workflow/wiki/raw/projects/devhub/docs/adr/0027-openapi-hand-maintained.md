# ADR-0027: OpenAPI hand-maintained spec + 정적 swagger-ui (CDN)

- **문서 목적**: DevHub v0.1.0 출시 직전 swagger UI 1차 bootstrap 의 OpenAPI spec 운영 방식 (hand-maintained yaml + 정적 HTML + CDN) 결정을 명문화한다.
- **범위**: `docs/openapi.yaml` 운영 정책 + `backend-core/internal/httpapi/swaggerui/asset/index.html` (CDN swagger-ui-dist@5.17.14) + `RouterConfig.SwaggerEnabled` / `RouterConfig.OpenAPISpecPath` (default false) + `/swagger/index.html` + `/swagger/openapi.yaml` 라우트 + `swaggerAssetFS embed.FS` + `backend-core/internal/shared/config/config.go` `SwaggerEnabled` env + `docs/backend_api_contract.md` §0/§1/§3 cross-link + `IMPL-swagger-01` 책임.
- **대상 독자**: Backend / 프론트엔드 개발자, AI agent, API consumer, QA, 운영자.
- **상태**: accepted
- **최종 수정일**: 2026-06-10
- **결정 근거 sprint**: v0.1.0 출시 직전 swagger UI 1차 bootstrap (코드/문서 정합 본 sprint, 별도 task 완료).
- **관련 문서**: [docs/openapi.yaml](../openapi.yaml) (1차 hand-maintained spec, 527 lines, 4 path + 7 schema), [docs/backend_api_contract.md §0/§1/§3](../backend_api_contract.md) (master index + envelope cross-link), [docs/traceability/report.md §2.4 + §4 + §6](../traceability/report.md) (IMPL-swagger-01 + ADR-0027 row + 변경 이력), [docs/governance/sync-checklist.md §3.2](../governance/sync-checklist.md) (backend_api_contract.md ↔ openapi.yaml cross-link step).

## 1. 배경

v0.1.0 출시 직전 시점에 DevHub 백엔드 API 가 100+ endpoint 에 도달했으나, 외부/내부 consumer 가 endpoint shape 를 확인할 수 있는 swagger UI 가 부재한 상태였다. 운영자/QA/외부 consumer 가 실시간으로 API contract 를 조회 가능하도록 swagger UI 1차 bootstrap 이 필요했다.

검토 요구:

- (a) swagger UI 즉시 노출 가능 (envelope + health + me + logout + metrics 1차 4 path + 7 schema, 도메인 endpoint 100+ 은 별도 sprint)
- (b) Go 의존성 추가 최소화 (v0.1.0 출시 직전 회귀 위험 차단)
- (c) Gin v1.12.0 (현재 main 의 Go web framework) 과의 호환 검증 부담 회피
- (d) 향후 v0.1.1 에서 도메인 endpoint 100+ 를 spec 으로 흡수할 확장성 보장

## 2. 후보 옵션 (4종)

| # | 옵션 | Go 의존성 | 핸드메인트 부담 | 호환성 위험 | 결정 |
| --- | --- | --- | --- | --- | --- |
| **1** | `gin-swagger` v1.6.0 (tagged release) | 1개 (`github.com/swaggo/gin-swagger` + transitive `swaggo/files`) | 자동 (swag init + 주석) | **Gin v1.12.0 호환 미검증** (v1.6.0 의 go.mod 가 Gin v1.9.x 까지만 검증) | ❌ |
| **2** | `gin-swagger` master | 1개 (master commit SHA) | 자동 (swag init + 주석) | **Go 1.25 호환 미검증** + tagged release 없음 (운영 pin 불가) | ❌ |
| **3** | **정적 HTML + CDN (swagger-ui-dist@5.17.14 unpkg) + embed.FS** | **0개** | hand-maintained yaml | 낮음 (Go 표준 `embed` + 정적 file serve) | ⭐ **채택** |
| 4 | redocly (Redoc CLI) + 정적 HTML 생성 | 0개 (Node CLI 별도) | hand-maintained yaml | 중간 (별도 Node toolchain + 빌드 step) | ❌ |

상세 비교:

- **옵션 1 (gin-swagger v1.6.0)**: swaggo/gin-swagger 의 tagged release v1.6.0 의 go.mod 가 Gin v1.9.x 까지만 검증. main 의 Gin v1.12.0 에서 internal API 변경 가능성 → vendoring / fork 부담.
- **옵션 2 (gin-swagger master)**: tagged release 부재, master commit SHA pin 필요. Go 1.25 호환 미검증. 운영 환경에서 SHA 재검증 부담.
- **옵션 3 (정적 HTML + CDN, 본 결정)**: `backend-core/internal/httpapi/swaggerui/asset/index.html` 정적 HTML 한 파일 + `swaggerAssetFS embed.FS` + 외부 `/swagger/openapi.yaml` disk serve. CDN unpkg `swagger-ui-dist@5.17.14` pin. Go 의존성 0개.
- **옵션 4 (Redocly)**: Redoc 정적 HTML 생성을 Node CLI 로 빌드. 빌드 step 추가 + 사내 Node toolchain 일관성 부담. 장점 = redoc 의 3-pane UI 가 깔끔하나 v0.1.0 bootstrap 의 4 path spec 에는 과한 투자.

## 3. 결정

**옵션 3: 정적 HTML + CDN (swagger-ui-dist@5.17.14 unpkg)** 채택.

### 3.1 결정 근거 4 항목

1. **Gin v1.12.0 호환 미검증 회피**: swaggo/gin-swagger v1.6.0 의 go.mod 가 Gin v1.9.x 까지만 검증되어, main 의 Gin v1.12.0 과 internal API 변경 가능성. vendoring / fork 부하 회피.
2. **Go 1.25 호환 미검증 회피**: gin-swagger master 는 Go 1.25 호환 미검증 + tagged release 부재로 운영 pin 어려움. SHA 재검증 부담 회피.
3. **의존성 0**: 옵션 3 은 Go module 에 신규 의존성 0개 추가. v0.1.0 출시 직전 회귀 위험 차단.
4. **v0.1.0 출시 직전 위험 최소**: 정적 HTML 한 파일 + CDN pin + embed.FS + disk serve = 단일 PR 검토 부담. 회귀 시 옵션 1/2 의 vendor 해제 + 코드 회수 부담 없음.

### 3.2 trade-off 인정

- CDN 의존성 (unpkg 가 다운되면 swagger UI 미동작) → §5.3 (b) self-host carve
- 외부 노출 위험 (prod 환경에서 swagger UI 가 public) → §5.3 (a) default false (opt-in)
- 인증 미요구 → §5.3 (c) v0.1.1 system_admin 가드 carve
- 도메인 endpoint 100+ 미포함 → 별도 sprint (`feat/openapi-domain-extend` 후보)

## 4. 결과

### 4.1 코드 / 자산 변경 요약

- `docs/openapi.yaml` (527 lines, 4 path + 7 schema). hand-maintained OpenAPI 3.0.3 spec. envelope (`components.schemas.Envelope` / `EnvelopeError`) + health + `/me` + `/auth/logout` + `/metrics` 4 path.
- `backend-core/internal/httpapi/swaggerui/asset/index.html`. CDN unpkg `swagger-ui-dist@5.17.14` pin swagger UI bootstrap HTML.
- `backend-core/internal/httpapi/router.go`. `swaggerMount` 추가 + `swaggerAssetFS embed.FS` (정적 HTML) + `/swagger/openapi.yaml` disk serve.
- `backend-core/internal/httpapi/router.go::RouterConfig`. `SwaggerEnabled bool` (default false) + `OpenAPISpecPath string` 필드.
- `backend-core/internal/shared/config/config.go`. `SwaggerEnabled bool` env 필드 + `DEVHUB_SWAGGER_ENABLED` env loader.

### 4.2 라우트

- `GET /swagger/index.html`. swagger UI HTML (embed.FS)
- `GET /swagger/openapi.yaml`. OpenAPI spec (disk, `OpenAPISpecPath`)

`SwaggerEnabled=false` (default) 시 두 라우트 모두 미등록. 운영 환경에서 opt-in.

### 4.3 검증

- 3 회귀 테스트 PASS (`backend-core/internal/httpapi/router_test.go` SwaggerEnabled/Path 분기)
- backend `go test ./...` PASS (35+ packages)
- v0.1.0 출시 직전 회귀 0

## 5. trade-off + carve out

### 5.1 prod 노출 위험

**문제**: swagger UI 가 prod 에서 노출되면 API contract 가 외부에 공개.

**완화**: `DEVHUB_SWAGGER_ENABLED` default false (opt-in). 운영 환경에서는 staging/dev 에서만 활성화, prod 에서는 명시적 enable 필요. 노출 시 reverse proxy 단의 IP allowlist 권장 ([ADR-0018 §4](../adr/0018-single-port-reverse-proxy-policy.md) 의 nginx allowlist 패턴 정합).

### 5.2 CDN 의존성

**문제**: unpkg CDN 가 다운되면 swagger UI 미동작.

**완화**: §6 carve (b). v0.1.1 에서 self-host 로 전환 (정적 파일 vendor). 단기 (v0.1.0) 에서는 CDN 가용성 99%+ 가정 + fallback UI (static "spec 파일 직접 조회 안내" HTML) 는 후속.

### 5.3 인증 미요구

**문제**: swagger UI 페이지 자체는 인증 미요구 (내부 spec 조회).

**완화**: §6 carve (c). v0.1.1 에서 `system_admin` 가드 미들웨어 추가 (Cookie session 검증). 단기 (v0.1.0) 에서는 §5.1 의 opt-in + reverse proxy IP allowlist 로 defense in depth.

### 5.4 도메인 endpoint 100+ 미포함

**문제**: 본 1차 spec 은 envelope + health + me + logout + metrics 4 path 만 포함. application-lifecycle / repository-integration / dev-request / integration-registry / realtime / onboarding / audit-ops / rbac-permissions / organization-management / auth-session 10 도메인의 100+ endpoint 미포함.

**완화**: §6 carve (d). 별도 sprint (`feat/openapi-domain-extend` 후보) 에서 도메인 endpoint 100+ spec 흡수 + envelope / cross-cutting 결정 (§1/§2 의 본문 + `conventions.md`) 와 정합 유지.

## 6. carve out (후속 sprint)

| # | 항목 | 우선순위 | 비고 |
| --- | --- | --- | --- |
| (a) | prod 노출 위험 → reverse proxy IP allowlist 가이드 | P1 | v0.1.0 운영 SOP 에서 처리 |
| (b) | CDN 의존성 → v0.1.1 self-host (정적 파일 vendor) | P2 | v0.1.1 milestone |
| (c) | 인증 미요구 → v0.1.1 system_admin 가드 | P2 | v0.1.1 milestone |
| (d) | 도메인 endpoint 100+ spec 확장 | P1 | 별도 sprint (`feat/openapi-domain-extend` 후보) |
| (e) | CI lint gate (openapi.yaml schema validity check) | P3 | v0.1.1+ |

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-10 | 1차 발행. v0.1.0 출시 직전 swagger UI 1차 bootstrap 결정 명문화. 옵션 3 (정적 HTML + CDN) 채택. 결정 근거 4 항목 (Gin v1.12.0 미검증 / Go 1.25 미검증 / 의존성 0 / 위험 최소). carve out 5 항목 (prod 노출 / CDN 의존 / 인증 미요구 / 도메인 endpoint 100+ / CI lint). 신규 ID: `IMPL-swagger-01`. | 본 sprint (v0.1.0 출시 직전 swagger UI 1차 bootstrap) |
