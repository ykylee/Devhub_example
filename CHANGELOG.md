# Changelog

All notable changes to DevHub will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0-alpha] - 2026-06-11

**v0.1.0-alpha release. 8 DoD 모두 close (7 ✅ + 1 ✅ skipped, N-6 사용자 결정).**

main HEAD: `d860b7c9` (PR #554 squash, N-6 skip + 4 file). Git tag: `v0.1.0-alpha`.

### 8 DoD (Definition of Done) — 모두 close

| # | DoD | Status |
|---|---|---|
| 1 | 사내 Keycloak realm OIDC 로그인 동작 | ✅ done (M0~M2) |
| 2 | system_admin 9 sub-page CRUD | ✅ done (M2) |
| 3 | Application/Repository/Project CRUD + rollup + 현황 페이지 | ✅ done (PR #104~#110) |
| 4 | HomeLab provider + sync + topology v2 | ✅ done (PR #155) |
| 5 | DREQ 흐름 end-to-end | ✅ done (PR #514 + #515) |
| 6 | e2e Playwright + backend `go test ./...` + frontend `npm run build` | ✅ done (CI 11/12 PASS) |
| 7 | 사내 staging 1주 운영 + 외부 사용자 ≥5 로그인 | ✅ skipped (사용자 결정, 2026-06-11) |
| 8 | UI 디자인 polish 1차 (semantic theme + responsive + a11y) | ✅ done |

### Added

**Auth / Org / Dashboard**
- Keycloak OIDC 통합 (ADR-0019, PR #167~#171) — JWKS cache + stale-while-error fallback (PR #242, 24h MaxStaleDuration)
- RBAC PermissionCache LISTEN/NOTIFY (RM-M4-08)
- Dashboard (developer/manager/admin) + 역할 routing (defaultLandingFor + isSystemAdmin)
- Account admin: Keycloak Admin Client 위임 (`/api/v1/accounts/*`, PR #167 KC-PR-C)
- Audit enrichment (source_ip / request_id / source_type, PR #57) + requireRequestID middleware
- Sign Out endpoint (P1-6, PR for N-8)
- e2e Kratos legacy 제거 + dynamic idp_subject sync (PR #249)

**Application / Repository / Project 도메인**
- API-01~58 activated (PR #104~#110, 7 마이그레이션 000012~000018)
- ADR-0011 accepted + 4 신규 RBAC resource (system_admin 일임)
- Platform 등록/조회 + Repository 연결 + Project CRUD + rollup + 현황 페이지 + SCM provider catalog
- 23 integration test (P1/P2 회귀 guard)

**DREQ (Development Request) 도메인**
- API-59~68 + API-79 (PATCH allowed_ips) activated (PR #514, ADR-0028)
- ADR-0012/0013/0014/0017
- Intake token + 외부 시스템 → DevHub POST → assignee dashboard → Promote (신규 application/project 1tx)
- 2 신규 table (dev_request_vocs + user_notifications) + dev_requests 4 column 확장 + 5 신규 API
- `(source_system, external_ref)` UNIQUE for idempotency
- TC-DREQ-* 13건 + `dev-requests.spec.ts` 6 step + `test_cases_m5_dreq.md`

**External Integration 도메인**
- API-69~80 activated (PR #135~#157)
- ADR-0015 / 0016 / 0017
- HomeLab pull (file + HTTP) + Provider/Binding CRUD + topology v2 시각화 + Prometheus 통합
- bindings 관리 UI (PR #154) + topology v2 시각화 (PR #155)
- ADR-0017 §6 atomicity 실 구현 (PR #156) + ADR-0015 §6 (1)+(2) (PR #157)
- TC-INT-FRONTEND-* 10건 + TC-INT-FRONTEND-BIND-* 3건 + TC-INT-HOMELAB-03 + TC-INT-FRONTEND-TOPOLOGY-V2-* 2건

**AI Workflow**
- ai-workflow v0.5.0 → v0.5.11 동기화 (PR #545, theirs-only 1 squash, 97 file / 4562줄)
- 2-tier governance (사외/사내 형상관리, PR #531~#537) — `AGENTS.md` §사외/사내 + `docs/governance/worker_division.md` §6
- v1.1 sprint -a follow-up (PR #538~#543) — port interface + saovae_stub (sso-integrations/keycloak)
- D-72 Phase 1/2/3 wiki integration + D-79/D-80 thin wrapper (PR #544, #551, #552)
- Vault ingestion (T-d-72-4, 82 wiki page 신규 ingest, vault commit b1599cc, 2026-06-11)
- T-d-72-2 (D-72 Phase 1 mirror) re-sync 완료 (2026-06-11 01:45:04Z, 83 file, 1.6M)

**UI / Frontend**
- Frontend 80 files / 1033 tests PASS (FE-08 신규)
- e2e Playwright 40 TC 게이트 (PR #86) + GitHub Actions CI 도입
- Design system (semantic theme + responsive + a11y baseline)
- PermissionEditor at /admin/settings/permissions ↔ /api/v1/rbac/policies
- e2e strict mode violation fix (`repositories-ui.spec.ts` L42 `.first()` 추가, commit `82935f8b`)

### Changed

- **워커 분업 전면 취소** (사용자 결정, PR #500, 2026-06-09) — 모든 신규 작업은 어느 에이전트로든 자유 진행
- **2-tier governance (사외/사내 형상관리)** (PR #531~#537, 2026-06-10) — GitHub 사외 = single source-of-truth / 사내 SCM = GitHub read-only pull
- **Kratos 잔재 residual cleanup** (sprint -ad) — 11 파일 삭제 + `identity_resolver.go` 신규 + Kratos 흐름 완전 제거
- `account_password.go` + `kratos_login_client` / `settings_client` / `session_cache` / `admin_client` + `password_auth_types` 모두 삭제
- `/api/v1/account/password` endpoint 폐기 — Keycloak Account Console redirect 위임 (`keycloak_operations.md` §8.5b)
- main flat memory 분리 (state.json 1515 → 150 line, 90% 감소, sprint `maintenance/work_260610-b-v1-pre-release-housekeeping`)
- v0.5.11-beta ai-workflow 표준화 (메모리 3 file reapply 분기 theirs-only 흡수 + 백업 보존)
- 명명 재검토 (PR #532, 2026-06-10) — `DEVHUB_APP_NAME` / `DEVHUB_APP_SHORT_NAME` env var override

### Fixed

- **N-8 sign-out e2e deterministic race hotfix 4차** (issue #501, PR #502 + #503) — 502 → 204 graceful degradation + `X-Keycloak-Likely-Down: true` marker + typed error sentinel `authview.ErrOIDCConfigMissing` + `authview.ErrOIDCNetworkUnreachable`
- **N-11 CI e2e + backend-integration 복원** (issue #419, sprint 260608-a) — `.github/workflows/ci.yml` 의 `&& false` 2건 코드 레벨 복원 + main 첫 PR 의 e2e shard 1/2/3 PASS 재확인
- e2e strict mode violation (`repositories-ui.spec.ts` L42 `.first()` 추가, commit `82935f8b`, post-merge standard)
- **DB 에러 노출 (SEC-5)** mask 5xx errors (PR-A, M0)
- `SetTrustedProxies(nil)` — `audit_logs.source_ip` X-Forwarded-For 위조 방어
- CI 회귀 (2026-06-01) — `frontend/app/(dashboard)/applications/[id]/page.tsx` 중복 import 제거 + `TC-PROJ-UI-04` 환경 독립 검증
- e2e seed 정합 (E2E seed `e2e-repo-a` 1 row 만 insert, 차트 toggle 전후 strict mode 회피)

### Security

- **SEC-1~SEC-4 resolved** (M0, 2026-05-08) — Keycloak OIDC + JWKS cache + resource_access fallback
- **DB 에러 노출 (SEC-5)** mask 5xx errors (PR-A, M0)
- Audit actor enrichment + requireRequestID middleware (PR-D, M1)
- JWKS stale-while-error fallback (PR #242, sprint -l, ADR-0020 sub-carve D) — 24h MaxStaleDuration, `DEVHUB_OIDC_JWKS_MAX_STALE_DURATION` env
- `metric devhub_jwks_stale_while_error_total{result}` + `devhub_jwks_stale_age_seconds` Histogram

### Deprecated

- `infra/idp/hydra.yaml` / `kratos.yaml` + setup README/ENVIRONMENT_NOTES (PR #169) — Keycloak 단일 IdP 전환
- `/api/v1/account/password` endpoint — Keycloak Account Console redirect 위임
- `DEVHUB_HYDRA_*` / `DEVHUB_KRATOS_*` env — PR #167 머지로 제거
- `infra/idp/README.md` — deprecated (PR #169)

### Removed

- **Kratos 전체 흐름** (PR #167 + sprint -ad) — historical Hydra introspection + Kratos 전체 흐름
- `account_password.go` + `kratos_login_client` / `kratos_settings_client` / `kratos_session_cache` / `kratos_admin_client`
- `kratos_admin_client_test` / `kratos_identity_resolver_test` (→ `identity_resolver_test` rename) / `kratos_login_fake_test` / `kratos_session_cache_test` / `kratos_settings_client_test` / `account_password_test`

### v0.1.0-alpha 후속 (잔여 follow-up)

**잔여 3** (T-d-79-2 / T-d-80-2, my_harness 측 SSOT 작성) — 사용자 전달 후 진행 중.
**잔여 5** (T-d-72-5/6 + D-73/74 + X-1~8, v1.1 forward path) — v1.1 milestone (2026-07-31) 진입 시점 별도 sprint.
**vault Gitea remote push** — 사용자 수동.

### 8 DoD 잔여 (skipped)

- **N-6 (v1.0 staging 1주 운영 검증)** — ✅ skipped, 사용자 결정 (2026-06-11). v0.1.0-alpha release blocker 0건.

### 1차 종합 매트릭스 (2026-05-13, PR #89 + #90)

- REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목
- 도메인 그룹 13행 (auth-session, integration-registry, repository-integration 등)
- [`docs/traceability/report.md`](docs/traceability/report.md) 갱신 — REQ → UC → ARCH → API → RM → IMPL → UT → TC 19 row 매트릭스

### 1차 종합 보안 리뷰 (2026-05-08)

- 신규 0건, 기존 1건 재확인 (SEC-5 권고)
- [`ai-workflow/memory/codebase-security-review-2026-05-08.md`](ai-workflow/memory/codebase-security-review-2026-05-08.md)
