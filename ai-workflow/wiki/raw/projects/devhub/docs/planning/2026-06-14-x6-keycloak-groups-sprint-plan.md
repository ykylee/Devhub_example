# X-6 Keycloak Group staging-prod 적용 (P1-3, #214) — Sprint Plan (2026-06-14)

- 문서 목적: X-6 (Keycloak group staging-prod 적용, issue #214) 의 sprint plan. **2-PR 분할** (사외 docs / 사내 script) 정공법.
- 범위:
  - **사외 PR (GitHub)**: ADR-0036 + roadmap status + traceability + memory
  - **사내 PR (사내 Gitea)**: `scripts/setup-keycloak-groups.sh` (group 4 생성 + composite role assign idempotent)
- 상태: draft (X-3 PR #591 + X-5 PR #592 + X-4 PR #593 머지 후 본 sprint 진입)
- 결정 근거 sprint: `docs/x6-keycloak-groups-public` (사외) / `infra/x6-keycloak-groups-internal` (사내)
- 관련 문서: [issue #214](https://github.com/ykylee/Devhub_example/issues/214), [release_v0-1_roadmap.md §3.5 X-6](../release_v0-1_roadmap.md), [ADR-0019 Keycloak-only IdP §5.3 잔여 carve](../adr/0019-keycloak-only-idp.md), [ADR-0026 Keycloak role excluded decision](../adr/0026-keycloak-role-excluded-decision.md), [`docs/setup/keycloak_operations.md` §4.3 + §4.4](../../setup/keycloak_operations.md) (이미 정합 — 본 sprint = status row update), [`docs/domain/rbac-permissions/keycloak_groups_mapping.md`](../../domain/rbac-permissions/keycloak_groups_mapping.md) (이미 존재), [`scripts/verify-keycloak-groups.sh`](../../../scripts/verify-keycloak-groups.sh) (이미 존재 — read-only 검증).

## 0. 컨텍스트

### 0.1 v1.0 의 잔여 carve 정합

[`docs/setup/keycloak_operations.md` §4.3](../../setup/keycloak_operations.md) (이미 merge) 가 다음을 결정:
- group 4종 (`devhub-developers` / `devhub-managers` / `devhub-pmo-managers` / `devhub-system-admins`) + composite realm role 1:1 매핑
- Default Groups 미설정 (codex review #9 정정)
- backend 변경 없음 (sub-carve C event listener 가 P1-1 완료)

[`scripts/verify-keycloak-groups.sh`](../../../scripts/verify-keycloak-groups.sh) (이미 merge) 가 read-only 검증 자동화.

### 0.2 잔여 = staging/prod 실제 적용 + 운영 SOP 보강

`docs/setup/keycloak_operations.md` 의 §4.3/§4.4 가 design + verify 만 다룸. **잔여 3건**:
1. **staging Keycloak** 에 group 4 + composite role 적용 (사내 Keycloak admin console 수동 작업, 사용자 follow-up)
2. **prod Keycloak** 동일 적용 (사용자 follow-up, staging 1주 관찰 후)
3. **운영 SOP 보강** (사외 docs only) — X-6 의 stage 결정 + ADR 신규 + status row

### 0.3 2-tier 분할 (2026-06-10 결정)

[`AGENTS.md` §사외/사내 2-tier 형상관리 분리](../../AGENTS.md) 정합:

| 영역 | Tier | Push 대상 |
|---|---|---|
| ADR-0036 (X-6 Architecture) | **공용** (docs only, 사내 한정 정보 미포함) | GitHub (synchronization) |
| release_v0-1_roadmap §3.5 X-6 row status | **공용** | GitHub |
| traceability/report.md §6 row | **공용** | GitHub |
| memory 4 file | **공용** | GitHub |
| `scripts/setup-keycloak-groups.sh` (sabo 한정 정보 포함 — Keycloak admin URL/secret 등) | **사내** | 사내 SCM (GitHub 에서 pull 만) |
| staging/prod Keycloak admin console 적용 | **사내** (사용자 수동) | N/A (admin console) |

## 1. 결정

### 1.1 opt-in (default off)

| 옵션 | 채택 | 이유 |
|---|---|---|
| **사내 script = opt-in (default 미적용)** | ✅ | production 의 모든 user token 의 role 이 즉시 영향. staging 1주 관찰 후 prod 적용 |
| auto-apply (default on) | ❌ | user token role 변경 = RBAC 권한 즉시 변동. staging 검증 없이 prod 적용 위험 |

### 1.2 script 정공법

- **Idempotent** (N회 실행해도 안전) — group 4종 이미 존재하면 skip
- **read-only fallback** 가능 (verify-keycloak-groups.sh 와 짝) — `--check-only` flag
- **rollback 지원** — `--remove` flag (group 4종 + composite role 모두 제거)
- **audit** — group 생성/role assign 시각 + actor 기록
- **dry-run** — `--dry-run` flag (실제 변경 없이 plan 출력)

### 1.3 토큰 정책 (사내 한정)

`scripts/setup-keycloak-groups.sh` 의 `--client-secret` env (`DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET`) — **이 script 자체에는 하드코드 안 함**. env var 또는 macOS keychain lookup. **사내 한정 정보** (Keycloak admin URL + client secret) 는 **사내 Gitea private repo 의 `.env.deploy.local`** 에 보관.

## 2. 변경 범위

### 2.1 사외 PR (GitHub, `docs/x6-keycloak-groups-public` branch)

1. `docs/adr/0036-keycloak-groups-staging-prod.md` (NEW, ~10KB, 9 section)
2. `docs/planning/release_v0-1_roadmap.md` §3.5 X-6 row status `✅ resolved (design accepted, X-6 PR #594, staging-prod 적용은 사용자 follow-up)`
3. `docs/traceability/report.md` §6 row 추가
4. `ai-workflow/memory/docs/x6-keycloak-groups-public/{state.json, session_handoff.md, work_backlog.md, backlog/2026-06-14.md}` (NEW, 4 file)

**신규 ID (4 row)**:
- `REQ-FR-KEYCLOAK-GROUPS-01` (group 4 + composite role 정공법)
- `ARCH-KEYCLOAK-GROUPS-01` (1:1 composite mapping + Default Groups 미설정 결정)
- `RM-KEYCLOAK-GROUPS-01` (staging-prod 운영 SOP — 1주 staging 관찰 후 prod)
- `IMPL-KEYCLOAK-GROUPS-01` (`scripts/setup-keycloak-groups.sh` — 사내 PR, 사외 PR 은 reference 만)

**Tier: 공용** (docs only, 사내 한정 정보 미포함)

### 2.2 사내 PR (사내 Gitea, `infra/x6-keycloak-groups-internal` branch)

1. `scripts/setup-keycloak-groups.sh` (NEW, ~200 line, bash) — Keycloak admin API 호출:
   - `POST /admin/realms/{realm}/groups` 4회 (Idempotent: 이미 존재 시 skip)
   - `POST /admin/realms/{realm}/groups/{id}/role-mappings/realm` 4회 (composite role assign)
   - `--dry-run` / `--check-only` / `--remove` flag
   - audit: `group_*.created` / `group_*.role_assigned` 4 log
2. `docs/adr/0036-keycloak-groups-staging-prod.md` **사내 한정 부분 append** (Keycloak instance URL + staging/prod admin console URL + client secret rotation SOP)
3. `infra/operations/keycloak-groups-staging-prod-runbook.md` (NEW, 사내 한정 runbook)

**Tier: 사내** (Keycloak admin API secret, staging/prod instance URL 포함)

## 3. 검증

### 3.1 사외 PR

- `bash scripts/wiki-sync-devhub.sh` 정합 (lint-config.toml)
- CI 4/4 PASS (docs only)
- tier self-check PASS (사내 한정 패턴 미포함)

### 3.2 사내 PR

- `bash scripts/setup-keycloak-groups.sh --check-only --realm=devhub` → 4/4 group + composite role 1:1 검증
- staging Keycloak 에서 1회 dry-run → 1회 apply → verify-keycloak-groups.sh PASS
- prod Keycloak 동일 (staging 1주 관찰 후)

## 4. 잔여 follow-up

- **staging Keycloak admin console 적용** (사용자 follow-up, 사내 staging instance)
- **prod Keycloak admin console 적용** (사용자 follow-up, staging 1주 관찰 후)
- **staging test user 의 token `realm_access.roles` 검증** (사용자 follow-up, [issue #214] 검증 기준 §3)
- **prod smoke test** (staging 1주 후)

## 5. 다음 세션 directive

- 본 sprint 진입 시 X-3 PR #591 + X-5 PR #592 + X-4 PR #593 머지 확인 + main 정합
- 사외 PR commit + push + PR 발행 + 머지
- 사내 PR commit (GitHub 에 push 안 함, 사내 Gitea 에만 push — 사용자 follow-up)
- 다음 sprint: X-8 (Keycloak SPI realm events, BE/사내)

## 6. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 본 sprint plan 초안 (X-6 Keycloak group staging-prod 적용, 2-PR 분할) | `docs/x6-keycloak-groups-public` / `infra/x6-keycloak-groups-internal` |
