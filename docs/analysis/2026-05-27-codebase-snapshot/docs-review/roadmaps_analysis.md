# 세부 개발 로드맵 2개 stale 점검 + 갱신 계획 (2026-05-27 스냅샷)

- 문서 목적: 백엔드/프론트엔드 **세부** 개발 로드맵 2개의 stale 항목을 `파일:줄` 근거로 식별하고, main `cf19c94` 현행 상태 기준 갱신 계획을 제시한다.
- 범위: `docs/backend_development_roadmap.md`(최종수정 2026-05-12, **매우 stale**), `docs/frontend_development_roadmap.md`(최종수정 2026-05-20, stale). 코드 수정 없음 — 두 로드맵 + 본 분석 문서만.
- 대상 독자: 로드맵 갱신 작업자, 추적성 동기화 검토자, 리뷰어.
- 상태: snapshot (2026-05-27, main `cf19c94` / 마지막 기능 커밋 `99d6edc`)
- 근거 자료: `../01_codebase_state_analysis.md`, `../03_frontend_summary.md`, `../04_backend_summary.md`, `../05_fe_be_balance.md`, `../06_future_direction.md`, `../code/backend/*`, `../code/frontend/*`.

---

## 0. 통합/릴리즈 로드맵은 메인 처리 (범위 외)

- `docs/development_roadmap.md`(통합) + `docs/planning/release_v1_roadmap.md`(v1.0 릴리즈) 2개는 **메인 에이전트가 별도 처리**한다. 본 분석/갱신 대상 아님.
- 두 세부 로드맵 상단의 "통합 로드맵 우선 확인" 배너는 유지하되, **v1.0/v1.1 신규 source-of-truth = release_v1_roadmap.md** 안내 + 본 코드베이스 스냅샷 링크를 보강한다.

---

## 1. backend_development_roadmap.md — stale 항목 (매우 stale)

기준일이 2026-05-12 로 정지되어 **ADR-0019 Keycloak 단일 IdP 전환(2026-05-19) 이후의 모든 사실이 미반영**이다. 가장 심한 stale 축은 (a) Hydra/Kratos IdP 서술, (b) `/api/v1/auth/*` + `/api/v1/accounts/*` 자체 계정 흐름, (c) Phase 13 "in_progress", (d) Phase 7/8 "in_progress", (e) §5 우선순위 계획의 M2~M4 미반영.

| # | 위치(`파일:줄`) | stale 서술 | 현행 사실(근거) |
| --- | --- | --- | --- |
| B-1 | `backend_development_roadmap.md:10` | 최종수정 2026-05-12 (Phase 13 + P0 M2 갱신) | 메타 날짜 → 2026-05-27 갱신 필요 |
| B-2 | `:11` | 관련 문서에 `frontend_integration_requirements.md` 등 + ADR-0001 "current IdP" 혼재 | ADR-0019 가 current, ADR-0001 superseded. release_v1_roadmap + 스냅샷 링크 추가 |
| B-3 | `:13` | "Phase 13 (IdP PoC) ... 머지 완료" 기준선 | Keycloak 단일 IdP 전환 완료, Hydra/Kratos 전면 제거. 기준선 → main `cf19c94` |
| B-4 | `:23` | "운영 actor 는 ... Hydra/Kratos 기반 세션 또는 JWT claim 에서 도출" | Keycloak OIDC JWT claim(JWKS 검증). Hydra/Kratos 폐기 |
| B-5 | `:24` | "credential/session master 는 Kratos 가 담당" | Keycloak 단일 IdP 가 credential/session master(ADR-0019) |
| B-6 | `:44` (Phase 13 row) | `in_progress (1차 완성 sprint)` — "Ory Hydra/Kratos IdP 도입", `/api/v1/auth/{login,...}`, `/api/v1/accounts/*`, `BearerTokenVerifier(Hydra introspection)`, "Kratos self-service webhook" | **done** — Keycloak OIDC 단일 IdP 전환(ADR-0019). 자체 auth/accounts proxy 폐기. JWKS 검증 + stale-while-error fallback(`internal/auth/keycloak_verifier.go`). Keycloak event polling → audit(`internal/audit/*`) |
| B-7 | `:38` (Phase 7 row) | `in_progress` — command/audit | **done** — command dry-run + live executor + approval(`commandworker/*`·`serviceaction/*`), audit Keycloak event polling 완성(04 §1) |
| B-8 | `:39` (Phase 8 row) | `in_progress` — WS, "replay, infra/ci/risk event publish ... 미완" | **부분(유지)** — ticket auth(ADR-0024) + command publish 완성, infra/ci/risk publish + replay 는 RM-M4-01/02 잔여(04 §6) |
| B-9 | `:80-82` (재검토 §4) | Phase 8/13 전제가 Hydra/Kratos introspection·admin wrapper 기준 | Keycloak 전환 후 무의미. §4 를 현행 잔여(RM-M4 publish/replay) 중심으로 정정 |
| B-10 | `:87-96` (§5 [P0] M2) | "Hydra introspection verifier ✅", "Accounts admin endpoints ✅", "Kratos self-service webhook 🟡 대기", "Hydra JWKS verifier deferred" | M2 인증 기반은 **Keycloak 전환으로 종결**. Hydra/Kratos/accounts 항목 전부 historical. 도메인 완성 현황으로 재작성 |
| B-11 | `:98-111` (§5 [P1]~[P3]) | Sign Up(`POST /auth/signup` + Kratos identity), gRPC, Gitea Reconciliation 만 나열 | Application/Repo/Project·DREQ·External Integration·Onboarding·Gitea sync 완성 반영 + Sign Up cancelled(외부 IdP) 표기. 잔여 = RM-M4 publish/replay + secret 암호화(#6) + backend-ai gRPC(v2) |
| B-12 | `:116-138` (§6 다음 작업 큐) | `[ ] PR-M2-AUDIT (Kratos webhook)`, `[ ] Hydra JWKS verifier`, M4 항목 일부 | Kratos/Hydra 큐 항목 historical 종결. 큐 = #368 draft/publish UT(N-2) + happy-path E2E(N-3) + RM-M4-01/02 + secret 암호화(#6) + backend-ai gRPC(v2) |
| B-13 | `:143` (§7 Blocked) | "Phase 13 round-trip 검증은 Hydra/Kratos native binary ... 필요" | Keycloak E2E CI 실 연동(shard 1/2) 완료. Hydra/Kratos blocker 소멸 |

### 갱신 방침 (backend)
- 기존 `done` 이력(§3 완료 범위, §6 `[x]` 항목)은 **보존**. stale 한 미래·진행 표기와 Hydra/Kratos·accounts 서술만 정정.
- Phase 13 row 는 삭제하지 않고 "done(Keycloak 전환, ADR-0019 — Hydra/Kratos PoC 는 historical)"로 정정(이력 보존).
- §5 [P0] M2 의 Hydra/Kratos/accounts 항목은 inline "historical(Keycloak 전환으로 대체)" 표기로 정정.

---

## 2. frontend_development_roadmap.md — stale 항목

기준일 2026-05-20. §6 의 자체 계정 관리(`/api/v1/accounts/*` + `must_change_password`) 서술과 Phase 4/7 "in_progress" 가 핵심 stale.

| # | 위치(`파일:줄`) | stale 서술 | 현행 사실(근거) |
| --- | --- | --- | --- |
| F-1 | `frontend_development_roadmap.md:9` | 최종수정 2026-05-20 | 메타 날짜 → 2026-05-27 |
| F-2 | `:11` | 관련 문서 — release_v1_roadmap·스냅샷 링크 없음 | source-of-truth 안내 + 스냅샷 링크 추가 |
| F-3 | `:24` (Phase 4 row) | `in_progress` — command status UI | **부분(유지)** — RealtimeService 구독 있으나 command toast/status 연결 미완(03 §1, Phase 4 잔여). 표기 정밀화 |
| F-4 | `:30` (Phase 7 row) | `in_progress` — 조직 1차 완성(부서 CRUD/계층/감사) | **done** — `/admin/settings/{organization,users,permissions,audit}` 완성(03 §1) |
| F-5 | `:70` (§6 제목) | "사용자/조직 관리 + **자체 계정 인증**" | 자체 계정 인증 폐기 → Keycloak Account Console 위임(ADR-0019/0020/0021) |
| F-6 | `:74-81` (§6.1 로그인 흐름) | `must_change_password=true` 라우팅, 로그인 폼 `login_id`+`password` | Keycloak OIDC PKCE 위임(IdP hosted login). `must_change_password` 폐기. `/login` 진입점 + AuthGuard + token refresh(03 §1) |
| F-7 | `:83-90` (§6.2 내 계정) | `GET /api/v1/accounts/{id}` + `PUT .../password`(본인 변경 3 필드) | accounts API 폐기. `/account` 는 ProfileSelfEdit(표시명/온보딩) + 비밀번호는 Keycloak Account Console redirect(`auth.service` Account Console URL) |
| F-8 | `:92-99` (§6.3 관리자 계정 관리) | `POST/PATCH/DELETE /api/v1/accounts` 발급/회수/잠금/강제 재설정 + 임시 비밀번호 1회 표시 | accounts admin endpoints 폐기. 사용자 lifecycle 은 Keycloak + Onboarding review(`/admin/settings/users` + admin review API-86) |
| F-9 | `:108-114` (§7.1 M2 sprint) | PR-UX1/2/3 + PR-M2-AUDIT(Kratos webhook) "진입 중" | 해당 sprint 종결(historical). 잔여 큐 현행화 필요 |
| F-10 | `:120-124` (§7.2 완료) | `/auth/login` 페이지(PR-LOGIN-2), `/account` 비밀번호 변경 폼(PR #50) | 이력 보존 가능하나 `/auth/login` → `/login` canonical(sub-carve F), 비밀번호 변경 폼 → Account Console 위임 정정 주석 |
| F-11 | `:128-130` (§7.3 후속) | command WS UI(M4) + AI Gardener(v2) + org 계정 관리 통합 | command WS UI·AI Gardener 유지. 잔여 큐에 FE 단위테스트 보강(B1) + happy-path E2E(B2) 추가 |

### 갱신 방침 (frontend)
- §6 의 자체 계정 관리 서술은 **삭제 대신 inline 정정 배너 + historical 표시**로 처리(사용자 지시: "삭제 대신 inline 정정 배너 + historical 표시 가능"). 폐기된 endpoint 는 취소선/주석으로 historical 표기 후 Keycloak Account Console 위임으로 정정.
- Phase 표(2~7)는 대부분 done 으로 정정하되 Phase 4 는 "부분(command WS UI 잔여)" 유지.
- §7 다음 작업 큐 = 단위테스트 보강 + command WS UI + happy-path E2E + AI suggestion(v2).

---

## 3. 공통 보강 (두 문서)

1. **source-of-truth 안내**: 두 문서 상단에 "v1.0/v1.1 신규 작업의 source-of-truth = `docs/planning/release_v1_roadmap.md`" 1줄 추가(기존 통합 로드맵 배너는 유지).
2. **코드베이스 스냅샷 링크**: `docs/analysis/2026-05-27-codebase-snapshot/`(특히 04 backend / 03 frontend / 05 balance) 링크를 관련 문서에 추가.
3. **변경 이력 1줄**: 각 문서 메타에 "2026-05-27 — Keycloak 전환(ADR-0019) 반영 + 현행 도메인 완성 정정(스냅샷 기준)" 1줄.

---

## 4. 현행 사실 요약 (반영 기준 — main `cf19c94`)

- **인증**: Keycloak OIDC 단일 IdP(ADR-0019). JWKS 검증 + stale-while-error fallback. WS ticket 인증(ADR-0024). 자체 accounts/auth/password 흐름 전부 폐기.
- **백엔드 완성 도메인**: 인증/조직/RBAC/감사 / Application·Repository·Project(+rollup, row-scoping) / DREQ(intake+promote-tx+token cron) / External Integration(provider/binding+auth_mode full+HomeLab pull) / Onboarding(gate+풀스택) / Gitea SCM pull sync worker(integration_sync_jobs 큐) / Repository draft→publish + SCM 양방향(import/create).
- **백엔드 잔여/부채**: Realtime infra/ci/risk publish + replay(RM-M4-01/02, command publish 만 완성), 평문 secret envelope 암호화(#6), #368 draft/publish 무테스트, backend-ai gRPC 미구현(v2), SPI push 전환(P3-5).
- **프론트 완성**: 인증/온보딩/조직/RBAC/감사/Application·Project/Repository(draft·SCM)/DREQ/Integration(auth_mode 동적 입력)/topology v2/admin catalog/운영 UI 전환(mock 제거+PageState).
- **프론트 잔여/부채**: 단위테스트 밀도 낮음(service 18 중 vitest 2), command status WS UI 미완(Phase 4 잔여), 최신 backend 기능 happy-path E2E 후행, AI suggestion UI(v2).
- **스택**: Next.js 16.2.6 / React 19.2.4 / Zustand 5 / Vitest 4 / Playwright 28 spec. Go + Gin + pgx, 마이그레이션 45, ADR 24.
</content>
</invoke>
