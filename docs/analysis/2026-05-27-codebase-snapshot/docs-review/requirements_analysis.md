# docs/requirements.md ↔ 현재 코드 정합 분석 (2026-05-27 스냅샷)

- 문서 목적: `docs/requirements.md`(524 L, 최종수정 2026-05-21) 를 main `cf19c94` 코드 상태와 대조해, 코드에 반영됐으나 요구사항에 누락/stale 인 항목을 근거와 함께 식별하고 갱신 계획을 제시한다.
- 범위: §2.5(계정관리) / §5.4(Project) / §5.6(External Integration) / §5.7(Onboarding) 중심 + 신규 §5.8(SCM↔시스템 Repository) 후보. 코드 수정은 없으며 requirements.md 갱신 계획만 다룬다.
- 대상 독자: requirements 갱신자, 추적성 동기화 작업자, 리뷰어.
- 상태: snapshot (2026-05-27, main `cf19c94`)
- 근거 자료: `../01_codebase_state_analysis.md`, `../02_sdlc_chain_status.md`, `../code/backend/domain.md`, `../code/backend/httpapi.md`, `../code/backend/migrations.md`, `docs/backend_api_contract.md` §15(API-69..90), `docs/adr/0019..0024`.

---

## 1. 현재 코드 구조 요약 (requirements 와 대조 관점)

### 1.1 인증 / 계정 (요구사항 §2.5 대응)
- Keycloak **단일 IdP** 완전 전환. Hydra/Kratos·`/api/v1/accounts/*`·`/api/v1/auth/*` 모두 폐기 (ADR-0019/0020/0021). `users.kratos_identity_id` → `idp_subject` rename(migration 000030). 자체 비밀번호 변경 핸들러(`/api/v1/account/password`)는 dead path 였고 잔재 정리 완료.
- DevHub `users` row 는 프로필/권한 메타데이터만 보유. 계정 발급/회수/강제 재설정/세션 만료는 Keycloak Admin Console 책임. `confirmUserReview` 외 account-admin write 경로 제거 (`resolveIdPSubject` dead).
- Keycloak 버전 pin = ADR-0022(25.0)/ADR-0023(26.0). WS ticket auth = ADR-0024.

### 1.2 Project (요구사항 §5.4 대응)
- Project model v2 활성: `application-project-repo` 관계 + **standalone project**(`projects.repository_id` nullable, migration 000037/000039/000044). `projectModel` 게이트 legacy|hybrid|v2 (default hybrid). repository-centric/application-centric/standalone 라우트 공존.
- → §5.4 의 `Project ↔ Repository N:M`, primary 지정, standalone 은 **이미 REQ-FR-PROJ-001/004** 등에 prose 로 존재. 코드와 큰 drift 없음. (standalone 은 §5.4.5 Out-of-Scope 의 "Gitea 자동생성 후속" 과 분리된 별개 개념 — 충돌 없음.)

### 1.3 External Integration (요구사항 §5.6 대응)
- Provider CRUD(API-69..71,80) 에 **auth_mode full 모델**(token/basic/oauth2/app_password/agent) + **base_url**(000038) + **api_token write-only**(000040) + **auth credentials 4 컬럼**(auth_username/auth_client_id/auth_token_url/auth_secret, 000041) 추가. `ResolveOutboundAuth()` 가 mode 별 자격증명 resolve.
- **연결 테스트** API-87(`POST /integration/test-connection`) — pre-save reachability(SSRF 의도 수용).
- **vendor preset 7종** + 가이드 자격증명(frontend `integration-provider-presets.ts`).
- **webhook 헤더 alias** — 범용 ingest(API-73) 가 `X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` fallback.
- write-only secret 응답 패턴 — `api_token_set`/`auth_secret_set` bool 만 노출. **단 `credentials_ref` 는 raw 노출(알려진 #6 보안 gap)**.

### 1.4 Onboarding (요구사항 §5.7 대응)
- Carve A(backend) + B/C(frontend·admin UI) + D(tests) 모두 머지 + feature flag default ON + `lazy_auto_create.go` 삭제 + codex hotfix(#291) 완료. API-83..86 활성, migration 000033.
- → §5.7 REQ-FR-ONBOARD-001..012 는 코드와 정합. drift 는 **로드맵/추적성의 진행상태 표기**(⏳→✅)지 requirements.md 본문이 아니다. requirements 본문은 대체로 정확.

### 1.5 Repository 소유권 / 생애주기 (요구사항에 신규 도메인)
- **소유권 분리**: `repositories.source`(scm|system, 빈값=legacy scm) + `provider_id`(integration_providers FK, 단일 출처) + `description`(system-owned). migration 000042/000045.
- **inbound import**: API-88(원격 목록)/API-89(import) — SCM 재조회 값으로 upsert(`source=scm`), system-owned 메타 보존.
- **outbound create**: API-90(create-repository) — gitea `CreateRepo` → `source=system` 미러. push capability + gitea-compat gate.
- **draft→publish lifecycle**: `repository_status`(draft|active, 000043) + `publish_requested_at`/`published_at`. `POST /api/v1/repositories`(createRepositoryDraft, source=system) → `POST /api/v1/repositories/:id/publish`(requestRepositoryPublish → gitea 생성 → active). SCM sync 는 `active` 직행. (#368/#373)
- **capability gate**: provider `capabilities` 가 기능 gate — pull=import(88/89), sync=mirror sync(72), push=create(90).
- **멱등 upsert + system-owned 보존**: `UpsertRepository` ON CONFLICT 가 SCM mirror 필드만 갱신, `source`/`description` 보존.

---

## 2. 발견한 누락 / 불일치 (requirements.md:줄 근거)

### G-A. §2.5 계정관리 — 자체 accounts/비밀번호 흐름 stale 표현 (정정 배너 대상)
요구사항 §2.5 는 2026-05-19 supersession 배너(`requirements.md:66`)를 이미 갖고 있어 "구현은 Keycloak 단일 IdP" 임을 명시한다. 다만 본문 항목들이 여전히 **DevHub 가 자체적으로 책임지는 것처럼 읽히는** 표현을 남긴다:
- `requirements.md:74` (로그인 ID 정책), `:75` (비밀번호 정책 — bcrypt/argon2id 해시 "저장한다"), `:76` (계정 상태 4종), `:85` (데이터 주권 메모 — "사용자 자신은 자신의 계정 정보(로그인 ID, 비밀번호)를 변경할 수 있다").
- 특히 `:75` 의 "단방향 해시…만 저장한다" 와 `:85` 의 "비밀번호를 변경할 수 있다" 는 자체 credential store 시절 표현. 현행은 Keycloak Account Console 로 redirect (자체 password form 폐기 — ADR-0019 sprint -ad).
- **조치**: §2.5 머리에 `> **갱신(2026-05-27)**` inline 배너를 1건 추가 — 본문 항목은 **삭제·재배열하지 않고** historical 로 표시하고 현행(Keycloak 단일 IdP, DevHub 는 프로필/권한/감사만)을 명시. 기존 `:66` 배너(2026-05-19, 정책 invariant 유지)와 **중복이 아니라 보완**(:66 은 "정책 유지", 신규 배너는 "self-service 비밀번호 흐름 = historical").

### G-B. §5.6 INT — auth_mode/base_url/api_token/연결테스트 요구사항 부재 (추가 항목 대상)
§5.6.1 REQ-FR-INT-001(`requirements.md:365`)은 provider 최소 속성으로 `auth_mode`, `scope` 만 언급하고 `base_url`/`api_token`/outbound 자격증명 모델은 없다. 코드/계약은 이미 full 모델(API-70 contract `backend_api_contract.md:2003-2012`, migration 000038/000040/000041)을 갖췄다.
- 누락 1: **auth_mode full 모델** — token/basic/oauth2/app_password/agent 5종 + mode 별 자격증명(api_token / auth_username+auth_secret / auth_client_id+auth_token_url+auth_secret).
- 누락 2: **base_url** — outbound sync/pull 대상 endpoint.
- 누락 3: **write-only secret** — `api_token`/`auth_secret` 응답 비노출(`*_set` bool). (단 `credentials_ref` raw 노출 = #6 미해소 gap, NFR 로 명시 가치.)
- 누락 4: **연결 테스트** — pre-save reachability(API-87), SSRF 사내 수용.
- 누락 5: **webhook 헤더 alias** — `X-Integration-*`→`X-Gitea-*`→`X-Gogs-*` fallback.
- **조치**: §5.6.1 에 REQ-FR-INT-013..015 **추가**(기존 001..012 유지), §5.6.2 에 REQ-NFR-INT-009 추가(write-only secret + credentials_ref gap 명시).

### G-C. §5.x — SCM↔시스템 Repository 연동 + lifecycle 도메인 전체 부재 (신규 §5.8)
requirements.md 전체에 다음 개념의 요구사항 ID 가 **하나도 없다**:
- repository **소유권 분리**(source=scm|system, provider_id FK) — §5.4 의 "외부 SoT vs DevHub 보유"(REQ-FR-APP-004, `:230`)는 metadata 보유만 언급, 소유권 source 구분/system-owned 필드 보존은 없음.
- 원격 **import**(API-88/89) / **outbound create**(API-90) / **draft→publish**(API 미발급, `repository_status`).
- capability 기능 gate.
- **조치**: §5.7 뒤에 신규 **§5.8 "SCM↔시스템 Repository 연동 + Repository Lifecycle"** 삽입. REQ-FR-REPO-001.. + REQ-NFR-REPO-001.. 신규 발급. (지시서 요구.)

### G-D. (참고, 본 PR 범위 아님) draft→publish 핸들러 무테스트 + 평문 secret
- `httpapi.md` F-1: `createRepositoryDraft`/`requestRepositoryPublish` 무테스트 머지(#368). → §5.8 REQ-NFR 에 "테스트 보강" 을 요구로 박을지 검토했으나, 요구사항은 "무엇을" 만 정의하고 테스트 부채는 추적성/로드맵 carve 영역 → 본 갱신은 기능 요구만, 테스트 부채는 분석 노트로만 남김.
- `migrations.md` 발견 2 + `httpapi.md` F-2: 평문 secret(credentials_ref/api_token/auth_secret). → REQ-NFR-INT-009 에 "평문 저장 금지(REQ-NFR-INT-001) 의 미충족 잔여 + credentials_ref raw 노출 gap" 으로 명시(이미 REQ-NFR-INT-001 이 평문 금지 요구, 현 코드 미충족을 노출).

### G-E. 메타 헤더 stale
- `requirements.md:8` 최종 수정일 = 2026-05-21. `:9` 관련 문서에 코드베이스 스냅샷 링크 없음.
- **조치**: 최종 수정일 → 2026-05-27, 관련 문서에 스냅샷 링크 추가, 문서 끝 변경 이력 1줄.

---

## 3. 갱신 계획 (requirements.md Edit)

| # | 위치 | 변경 | 원칙 |
| --- | --- | --- | --- |
| 1 | `:8`/`:9` 메타 헤더 | 최종 수정일 2026-05-27 + 스냅샷 링크 추가 | 메타 갱신 |
| 2 | §2.5 머리(`:64` 직후) | `> **갱신(2026-05-27)**` inline 배너 1건 (Keycloak 단일 IdP self-service 흐름 = historical) | **삭제·재배열 금지**, inline 정정만 |
| 3 | §5.6.1 끝(`:387` REQ-FR-INT-012 뒤) | REQ-FR-INT-013(auth_mode full) / -014(base_url+연결테스트) / -015(webhook 헤더 alias) **추가** | 기존 001..012 유지, 추가만 |
| 4 | §5.6.2 끝(`:398` REQ-NFR-INT-008 뒤) | REQ-NFR-INT-009(write-only secret + credentials_ref gap) **추가** | 추가만 |
| 5 | §5.7 뒤(`:503` 직후, §6 앞) | 신규 **§5.8** + REQ-FR-REPO-001..005 + REQ-NFR-REPO-001..003 | 신규 섹션 |
| 6 | 문서 끝(`:524`) | 변경 이력 1줄 | 추가만 |

### 신규 ID 범위 (확정)
- **REQ-FR-INT-013, -014, -015** (§5.6.1 보강).
- **REQ-NFR-INT-009** (§5.6.2 보강).
- **REQ-FR-REPO-001 ~ -005** (§5.8.1 신규).
- **REQ-NFR-REPO-001 ~ -003** (§5.8.2 신규).

### §5.8 내용 매핑 (코드 근거)
| 신규 REQ | 코드/계약 근거 |
| --- | --- |
| REQ-FR-REPO-001 소유권 분리 | `domain.go:8` Source/ProviderID/Description, migration 000042/000045 |
| REQ-FR-REPO-002 원격 import | API-88/89 (`backend_api_contract.md:2068-2083`), `integration_scm_repositories.go` |
| REQ-FR-REPO-003 outbound create | API-90 (`:2085-2092`), gitea `CreateRepo` |
| REQ-FR-REPO-004 draft→publish | `repository_status`(000043), `POST /repositories`(+`/publish`), httpapi `domain.go:141-287` |
| REQ-FR-REPO-005 capability gate | `backend_api_contract.md:1989` (pull/sync/push gate) |
| REQ-NFR-REPO-001 멱등 upsert + system-owned 보존 | `UpsertRepository` ON CONFLICT, migrations.md 발견 |
| REQ-NFR-REPO-002 gitea-compat gate | `isGiteaCompatibleProvider`, `:1991` |
| REQ-NFR-REPO-003 audit + 무테스트 부채 명시 | httpapi.md F-1 (테스트 보강 후속), draft→publish 부분 실패 경로 |
