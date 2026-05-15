# ADR-0012: Dev Request 외부 수신 endpoint 인증 정책

- 문서 목적: Dev Request (DREQ) 도메인의 외부 수신 endpoint (`POST /api/v1/dev-requests`) 가 일반 OIDC 사용자 흐름이 아닌 별도 인증 정책을 사용해야 하는 이유와 채택 옵션을 결정한다.
- 범위: 외부 시스템(ops portal / ITSM / Jira / 사내 워크플로우 도구) 가 DevHub 에 의뢰를 POST 하는 흐름의 인증 + 권한 + 감사. 본 endpoint 외 DREQ 도메인의 다른 endpoint (GET 목록/상세, Promote, Reject, Reassign, Close) 는 일반 OIDC + RBAC 으로 보호되며 본 ADR 의 범위 밖이다.
- 대상 독자: Backend 개발자, Auth 트랙 담당자, 외부 시스템 통합 담당자, AI agent, 운영 진입 감사자.
- 상태: accepted
- 작성일: 2026-05-15
- 결정일: 2026-05-15 (sprint `claude/work_260515-g`)
- 결정 근거 sprint: `claude/work_260515-g` — DREQ-AuthADR
- 관련 문서: [`docs/planning/development_request_concept.md` §7](../planning/development_request_concept.md), [`docs/requirements.md` REQ-NFR-DREQ-001](../requirements.md), [`docs/architecture.md` ARCH-DREQ-03](../architecture.md), [`docs/backend_api_contract.md` §14.1](../backend_api_contract.md), [ADR-0001 IdP 선정](./0001-idp-selection.md), [ADR-0011 RBAC row-scoping](./0011-rbac-row-scoping.md).

## 1. 컨텍스트

DevHub 의 사용자 인증은 OIDC (Hydra + Kratos) 기반이며, 모든 `/api/v1/*` route 는 `enforceRoutePermission` middleware 가 Bearer access token + RBAC matrix 로 보호한다. 그러나 **외부 시스템에서 들어오는 의뢰 수신 endpoint** 는 다음 두 가지 이유로 일반 OIDC 흐름을 그대로 적용할 수 없다:

1. **외부 시스템은 DevHub 의 user 가 아니다.** ops portal / ITSM / 사내 워크플로우 도구 같은 외부 시스템은 사람 사용자가 아닌 service 단위로 호출하므로 Kratos 의 identity 모델에 직접 매핑되지 않는다.
2. **외부 시스템마다 인증 모델 도입 비용이 다르다.** 일부 외부 시스템은 OAuth flow 를 구현할 수 있고, 일부는 단순한 정적 토큰만 지원한다. 1차 도입 시 진입 장벽이 낮은 메커니즘이 운영 친화적이다.

따라서 `POST /api/v1/dev-requests` 는 일반 OIDC bypass + 별도 인증 정책으로 처리해야 하며, 그 정책의 선택이 본 ADR 의 결정 대상이다.

REQ-NFR-DREQ-001 ([requirements.md §5.5.2](../requirements.md)) 가 본 ADR 의 결과를 참조하며, ARCH-DREQ-03 ([architecture.md §7.3](../architecture.md)) 의 "외부 수신 인증 경계" 가 ADR 결정 후 구체화된다.

## 2. 결정 동인

- **외부 시스템 on-boarding 비용**: 외부 시스템이 SDK 작성 / OAuth flow 구현 없이 즉시 호출할 수 있어야 한다.
- **운영 단순성**: 토큰 발급 / 회전 / revoke 절차가 운영자가 이해하고 관리할 수 있어야 한다.
- **감사 가능성**: 외부 호출은 어느 클라이언트(`source_system`)가 언제 호출했는지가 audit 으로 추적되어야 한다.
- **보안 수준**: 토큰 노출 / replay / IP spoofing 등 위협에 대한 1차 방어선이 있어야 한다. 운영 진입 후 보안 강도 강화는 후속 ADR 에서 가능해야 한다.
- **현행 인증 모델과의 호환**: DevHub 내부 사용자 인증 모델(OIDC) 을 깨지 않고 추가되어야 한다. 본 endpoint 만 별도 middleware 로 분기.

## 3. 검토 옵션 평가

### 3.1 비교 표

| 옵션 | 요약 | on-boarding 비용 | 보안 수준 | 운영 단순성 | 감사 가능성 | 평가 |
| --- | --- | --- | --- | --- | --- | --- |
| **A. API 토큰 + IP allowlist** | 정적 long-lived bearer token + caller IP whitelist. 토큰은 외부 시스템별로 발급, DB 에 hashed 저장, 사용 시점에 audit 기록. | **낮음** — 외부 시스템은 `Authorization: Bearer <token>` 헤더 1개만 추가 | 중 — 토큰 노출 위험 있으나 IP allowlist 가 2차 방어, replay 는 idempotency(REQ-NFR-DREQ-002) 가 흡수 | 좋음 — 운영자 UI/CLI 로 토큰 발급/회수, IP 변경 시 DB row 수정 | 좋음 — token_id 가 audit 의 `source_system` 으로 매핑 | ✅ (1차 채택) |
| **B. HMAC 시그니처** | Request body + timestamp + nonce 를 shared secret 으로 HMAC-SHA256. `X-DREQ-Timestamp` / `X-DREQ-Nonce` / `X-DREQ-Signature` 헤더. | 중 — 외부 시스템에 HMAC 구현(약 20줄) + nonce 저장 + timestamp skew 처리 필요 | **높음** — replay 방지 (nonce + skew window), secret 노출 시에도 timestamp/nonce 검증으로 시간 제한, body tampering 방지 | 중 — secret rotation 시 동시 활성 secret 운영 필요, nonce 저장소 운영 | 좋음 — client_id 기준 audit | 🟡 (2차 강화 후보) |
| **C. OAuth client_credentials (Hydra)** | Hydra 의 `client_credentials` grant 로 access_token 발급. 외부 시스템이 token endpoint 호출 → DevHub 가 introspection 으로 검증. | **높음** — 외부 시스템에 OAuth client 등록 + token 획득/refresh flow 구현, Hydra 호출 path 노출 | 높음 — 토큰 단명 + 자동 회전, Hydra 의 일관된 검증 모델 | 중 — Hydra 의 외부 client 등록/회수 운영 필요, OIDC ops 부담 증가 | 우수 — DevHub 의 OIDC audit pipeline 과 통합 | ❌ (1차 거부, 외부 client 가 OAuth 가능하면 후속 단계에서 재평가) |

### 3.2 단계적 도입 관점

- **1차 (MVP, 본 ADR 시점)**: 외부 시스템 다양성을 모르는 상태. 가장 호환성 높은 옵션이 필요. 보안은 IP allowlist + 토큰 revoke + audit 로 1차 보장.
- **2차 (운영 안정화 + 보안 요구 강화 시)**: 옵션 B (HMAC) 로 마이그레이션. A 의 토큰 인증을 그대로 두고 HMAC 검증을 추가 (additive 변경), 외부 시스템이 자발적으로 전환하도록 grace period 운영. 이 단계에서 별도 ADR.
- **3차 (사내 OIDC 통합 진입 시)**: 옵션 C (OAuth client_credentials) 로 자연 마이그레이션. 사내 외부 시스템이 모두 Hydra client 가 되는 시점에 도입.

## 4. 결정

**옵션 A (API 토큰 + IP allowlist) 를 1차 채택**한다. 보안 강화 마이그레이션 경로 (A → B → C) 는 추가 ADR 로 결정한다.

### 4.1 1차 채택 정책 — 본 결정

#### 4.1.1 토큰 발급 / 저장
- 외부 시스템별로 long-lived bearer token 을 발급한다. 토큰 자체는 32+ byte 의 cryptographically random string (base64url 인코딩 권장).
- DB 테이블 `dev_request_intake_tokens` 에 저장. 컬럼: `token_id` (PK, UUID), `client_label` (운영용 식별자, 예: "ops_portal"), `hashed_token` (SHA-256 hex, **plain token 은 절대 저장하지 않음**), `allowed_ips` (jsonb, CIDR 배열), `source_system` (token 매핑되는 source_system 값), `created_at`, `last_used_at` (nullable), `revoked_at` (nullable), `created_by` (system_admin user_id).
- 발급 직후 1회만 plain token 을 admin 에게 노출하고 이후 어디에도 저장하지 않는다 (Kratos password issuance 패턴과 동일 — accounts_admin).

#### 4.1.2 검증 흐름
- 외부 호출은 `Authorization: Bearer <plain-token>` 헤더로 도착.
- middleware 가 `SHA-256(plain-token)` 계산 후 `dev_request_intake_tokens.hashed_token` 으로 lookup.
- 매칭된 row 가 없거나 `revoked_at IS NOT NULL` 이면 401.
- caller IP 가 `allowed_ips` CIDR 범위 안에 없으면 401.
- 검증 성공 시 `c.Set("devhub_dreq_source_system", row.source_system)` + `last_used_at = NOW()` 갱신 + audit `dev_request.intake_auth_succeeded` emit.
- 검증 실패 시 audit `dev_request.intake_auth_failed` emit (헤더 누락 / 토큰 미매칭 / IP 차단 등 사유 분기).

#### 4.1.3 운영 정책
- **회전 주기**: 12개월 (운영 정책 default, REQ-NFR-DREQ-001 의 후속 운영 hygiene).
- **revoke 절차**: `system_admin` 이 admin UI 또는 CLI 로 즉시 revoke. 즉시 `revoked_at = NOW()` 갱신.
- **토큰 노출 시 절차**: 즉시 revoke + 외부 시스템에 통보 + 새 토큰 발급. 의뢰 ingest 중단 위험 있으므로 staged rotation 권장 (구 토큰 revoke 전 신 토큰 발급).
- **IP allowlist 변경**: 운영자가 admin UI 로 CIDR 추가/제거. 변경 즉시 적용.
- **last_used_at 모니터링**: 30일 이상 미사용 토큰은 audit 알림 (자동 revoke 는 아님 — 운영자 판단).

#### 4.1.4 routePermissionTable 처리
- `POST /api/v1/dev-requests` 는 `routePermissionTable` 에 등록하되, 정책 `Bypass: true` 또는 별도 `IntakeAuth: true` 플래그로 일반 OIDC enforce 를 건너뛴다.
- 본 endpoint 의 actual auth 는 별도 middleware (`requireIntakeToken`) 가 처리. 검증 실패 시 즉시 401 + 응답 envelope.

#### 4.1.5 idempotency 와의 상호작용
- REQ-NFR-DREQ-002 의 `(source_system, external_ref)` idempotency 는 본 인증 모델과 독립적으로 작동. `source_system` 은 토큰에서 자동 주입되므로 외부 시스템이 body 에서 임의로 변경 못 함 (ARCH-DREQ-03 의 "spoofing 방지" 정책 정착).

#### 4.1.6 audit 카탈로그 추가 (ARCH-DREQ-06 확장)
- `dev_request.intake_auth_succeeded` — payload: `{token_id, client_label, source_ip}` (token 자체는 절대 audit 에 기록 안 함)
- `dev_request.intake_auth_failed` — payload: `{reason, source_ip, header_present, token_prefix_4chars (디버깅용, full token 금지)}`
- 기존 `dev_request.received` 는 인증 성공 후 emit.

### 4.2 옵션 B (HMAC) 거부 사유 — 1차 한정

- 외부 시스템의 다양성을 모르는 상태에서 HMAC 구현을 강요하면 on-boarding 지연 발생.
- 1차 운영에서 의뢰량/노출 위험이 낮을 것으로 추정 — 보안 강도 vs 운영 비용 trade 에서 운영 단순성 우선.
- B 의 보안 강점 (replay 방지) 은 idempotency (REQ-NFR-DREQ-002) 가 부분적으로 흡수.
- 2차 강화 sprint 에서 A 위에 HMAC 검증을 *additive* 로 추가하는 것이 자연스러우므로 거부가 아닌 단계적 보강 경로.

### 4.3 옵션 C (OAuth client_credentials) 거부 사유 — 1차 한정

- 외부 시스템이 Hydra client 로 등록 + token endpoint 호출 구현 필요 — 진입 장벽 최대.
- DevHub 의 OIDC pipeline (Hydra) 을 외부 client 에 노출하는 추가 보안 surface.
- 사내 외부 시스템이 모두 OIDC client 화될 시점이 명확하지 않음 — 그 시점에 별도 ADR 로 마이그레이션 검토.

## 5. 결과

- **REQ-NFR-DREQ-001** 의 후보 3종 중 옵션 A 가 결정됨. 본 ADR 머지 후 REQ-NFR-DREQ-001 의 "후보" 표기가 "확정" 으로 갱신.
- **ARCH-DREQ-03** 의 외부 수신 인증 경계 구체화 — `dev_request_intake_tokens` 테이블 + `requireIntakeToken` middleware + Bypass routePermissionTable + audit 2 action.
- **ARCH-DREQ-06** audit 카탈로그에 `dev_request.intake_auth_{succeeded,failed}` 2개 추가.
- **API §14.1** spec 의 "인증" 섹션이 본 ADR 의 결정으로 구체화 — `Authorization: Bearer <token>` 헤더 + IP allowlist.
- **DREQ-Backend sprint 진입 조건 충족** — 본 ADR 머지 후 backend 구현 sprint 가 진입 가능.
- **신규 마이그레이션**: `000023_dev_request_intake_tokens.up.sql` (DREQ-Backend sprint 에서 작성). `000022_dev_requests` 보다 뒤 — token 검증이 dev_requests 수신의 의존 조건이라 같은 PR 에서 함께 처리하거나 token migration 을 먼저 적용한다.
- **추적성 매트릭스 §4**: ADR-0012 인덱스 row 추가. §3 DREQ 도메인 row 의 ADR 컬럼에 ADR-0012 추가.

## 6. 후속 작업

- **(carve out, 2차 보안 강화)** 옵션 B (HMAC) 추가 ADR — 운영 진입 후 의뢰량 / 노출 위험 / 외부 시스템 역량을 평가한 시점에 작성.
- **(carve out, 3차 OIDC 통합)** 옵션 C (OAuth client_credentials) 마이그레이션 ADR — 사내 외부 시스템의 OIDC 통합 완료 시점.
- **(DREQ-Backend sprint)** `dev_request_intake_tokens` 테이블 마이그레이션 + `requireIntakeToken` middleware 구현 + audit 2 action 도입.
- **(DREQ-Admin UI sprint)** 토큰 발급/revoke/IP 관리 admin UI — `/admin/settings/dev-request-tokens`. 발급 직후 plain token 1회 노출 패턴 (accounts_admin 의 password issuance 와 일관).
- **(REQ-NFR-DREQ-006 결정)** rate limiting / RPS 제한 정책 — 본 ADR 의 인증 + IP allowlist 만으로는 DoS 방어 부족. 별도 sprint.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-15 | accepted — 옵션 A (API 토큰 + IP allowlist) 1차 채택, B/C 거부 + 단계적 마이그레이션 경로 명시. `dev_request_intake_tokens` 테이블 스키마 + `requireIntakeToken` middleware 정책 + audit 2 action 정의. | sprint `claude/work_260515-g` |
