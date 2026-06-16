# ADR-0036: Keycloak Group staging-prod 적용 (X-6, P1-3, #214)

- 문서 목적: X-6 (Keycloak group staging-prod 적용, issue #214) 의 architecture 결정. group 4종 + composite realm role 1:1 매핑 + 2-PR 분할 정공법.
- 범위: staging/prod Keycloak admin console 적용 SOP + setup script 정공법 (idempotent + dry-run + check-only) + 2-PR 분할 (사외 docs / 사내 script).
- 상태: Accepted
- 작성일: 2026-06-14
- 결정 근거 sprint: `docs/x6-keycloak-groups-public` (사외) / `infra/x6-keycloak-groups-internal` (사내)
- 결정 근거 commit: (본 PR squash 예정)
- **Tier**: 공용 (이 ADR 본문) + 사내 (사내 한정 부분 — 별도 append, 사내 Gitea PR)
- 관련 문서: [issue #214](https://github.com/ykylee/Devhub_example/issues/214), [release_v0-1_roadmap.md §3.5 X-6](../planning/release_v0-1_roadmap.md), [ADR-0019 Keycloak-only IdP §5.3 잔여 carve](./0019-keycloak-only-idp.md), [ADR-0026 Keycloak role excluded decision](./0026-keycloak-role-excluded-decision.md), [`docs/setup/keycloak_operations.md` §4.3 + §4.4](../setup/keycloak_operations.md) (이미 정합 — 본 ADR 의 cross-ref), [`docs/domain/rbac-permissions/keycloak_groups_mapping.md`](../domain/rbac-permissions/keycloak_groups_mapping.md) (이미 존재), [`scripts/verify-keycloak-groups.sh`](../../scripts/verify-keycloak-groups.sh) (이미 존재 — read-only 검증).

---

## 1. 컨텍스트

### 1.1 잔여 carve

ADR-0019 §5.3 (Keycloak-only IdP) 의 잔여 carve:
- group 4종 (`devhub-developers` / `devhub-managers` / `devhub-pmo-managers` / `devhub-system-admins`) + composite realm role 1:1 매핑 결정
- Default Groups 미설정 (multi-role order-dependency 회피)
- backend 변경 없음 (sub-carve C event listener 가 P1-1 완료)

[`docs/setup/keycloak_operations.md` §4.3 + §4.4](../setup/keycloak_operations.md) 가 design + verify 자동화 결정. [`scripts/verify-keycloak-groups.sh`](../../scripts/verify-keycloak-groups.sh) 가 read-only 검증.

**잔여 3건**:
1. staging Keycloak admin console 적용
2. prod Keycloak 동일 적용 (staging 1주 관찰 후)
3. 운영 SOP 보강 + ADR-0036 (본 문서)

### 1.2 Issue #214 검증 기준

| 검증 항목 | 기준 |
|---|---|
| staging Keycloak group 4 존재 | `GET /admin/realms/devhub/groups` → 4 발견 |
| composite role 1:1 매핑 | `GET /admin/realms/devhub/groups/{id}/role-mappings/realm` → expected role 1개 |
| Default Groups 비어 있음 | `GET /admin/realms/devhub/default-groups` → `[]` |
| staging test user token role emit | token `realm_access.roles` 가 group 매핑 따라 정확히 emit |
| backend role 추출 정상 | `extractKeycloakRole` priority filter (keycloak_verifier.go:260-285) |

## 2. 검토 옵션

### 2.1 group 4 ↔ role 4 매핑 방식

| 옵션 | 채택 | 이유 |
|---|---|---|
| **옵션 B: composite realm role 1:1 매핑** ([keycloak_groups_mapping §3.2](../domain/rbac-permissions/keycloak_groups_mapping.md) 결정) | ✅ | user 가 group 1개 가입 → Keycloak 이 token 에 role 자동 포함 → backend `extractKeycloakRole` 단일 source-of-truth. group ↔ role 1:1 → multi-role order-dependency 회피 |
| 옵션 A: groups claim mapper | ❌ | mapper 추가 + backend 변경 (group → role 매핑 별도) |
| 옵션 C: SCIM bridge / LDAP federation | ❌ | 1차 release scope 외. 별도 carve |

### 2.2 Default Groups 설정

| 옵션 | 채택 | 이유 |
|---|---|---|
| **Default Groups 미설정** (codex review #9 정공법) | ✅ | Default Group 적용 시 신규 manager/pmo/admin 도 자동 `devhub-developers` 가입 → token multi-role → `extractKeycloakRole` order-dependency 위험. 명시 group 1개 가입 강제 (§8.1 step 3) |
| Default Group = `devhub-developers` | ❌ | multi-role order-dependency 위험 |
| Default Group = role 4 모두 | ❌ | 모든 user 가 4-role token → `extractKeycloakRole` 첫 번째 매칭 = developer 강제 |

### 2.3 2-PR 분할 (2026-06-10 결정 정합)

[AGENTS.md §사외/사내 2-tier 형상관리 분리](../../AGENTS.md) 결정 정합:

| PR | Tier | Push | 내용 |
|---|---|---|---|
| 사외 PR (`docs/x6-keycloak-groups-public`) | 공용 | GitHub | ADR-0036 + roadmap status + traceability + memory |
| 사내 PR (`infra/x6-keycloak-groups-internal`) | **사내** | 사내 Gitea | `scripts/setup-keycloak-groups.sh` + runbook + ADR 사내 한정 append |

## 3. 결정

### 3.1 group 4 + composite realm role 1:1 매핑

| Keycloak Group | Composite Realm Role | DevHub `users.role` |
| --- | --- | --- |
| `devhub-developers` | `developer` | `developer` |
| `devhub-managers` | `manager` | `manager` |
| `devhub-pmo-managers` | `team_manager` | `team_manager` |
| `devhub-system-admins` | `system_admin` | `system_admin` |

### 3.2 Default Groups 미설정

`keycloak_groups_mapping.md §3.2` 결정 정합. 신규 user 는 명시 group 1개 가입 강제 (`docs/setup/keycloak_operations.md §8.1 step 3`).

### 3.3 backend 변경 없음

`keycloak_verifier.go:260-285` 의 `extractKeycloakRole` priority filter 가 group composite role 도 동일하게 token 에서 추출. **추가 변경 불요**.

### 3.4 staging → prod 순서

1. **staging 1주 관찰** (사용자 follow-up)
   - staging Keycloak group 4 적용 (admin console)
   - staging test user 의 token `realm_access.roles` 검증
   - backend role 추출 정상 검증
   - verify-keycloak-groups.sh 1회 실행 (PASS 확인)
2. **prod 적용** (staging 1주 후)
   - prod Keycloak group 4 적용 (admin console)
   - prod smoke test (4 role × 1 user × 1 endpoint)
   - 운영 SOP 확정

### 3.5 2-PR 분할 결정 (AGENTS.md §6 정합)

**사외 PR** (GitHub, `docs/x6-keycloak-groups-public`):
- ADR-0036 (이 문서) — Tier: 공용 (사내 한정 정보 미포함)
- release_v0-1_roadmap.md §3.5 X-6 row status update
- traceability/report.md §6 row
- memory 4 file (state.json + session_handoff + work_backlog + backlog)

**사내 PR** (사내 Gitea, `infra/x6-keycloak-groups-internal`):
- `scripts/setup-keycloak-groups.sh` (NEW) — Idempotent + dry-run + check-only + remove
- runbook (`infra/operations/keycloak-groups-staging-prod-runbook.md`) — staging/prod instance URL + admin client_id/secret rotation SOP
- ADR-0036 사내 한정 append (Keycloak instance URL + client secret rotation)

## 4. trade-off

### 4.1 manual console vs script

| 옵션 | 채택 | 이유 |
|---|---|---|
| **admin console 수동 + script 도구화 (정공법)** | ✅ | admin console 이 source-of-truth (UI audit 가능), script 는 idempotent 도구 (재현성 + 검증 자동화). 둘 다 보존 |
| script only | ❌ | UI audit trail 부재. 사내 IdP 팀 SOP 와 불일치 |
| console only | ❌ | 재현성 + 검증 자동화 부재. 1주 관찰 후 동일 작업 반복 시 human error 위험 |

### 4.2 staging 1주 vs 즉시 prod

| 옵션 | 채택 | 이유 |
|---|---|---|
| **staging 1주 관찰 후 prod** | ✅ | user token role 변경 = RBAC 권한 즉시 변동. staging 검증 후 prod 적용 (안전 우선) |
| 즉시 prod | ❌ | user permission mismatch 시 prod incident 위험 |

### 4.3 dry-run / check-only / remove flag

setup script 의 3 flag (CRITICAL 한 운영 안전성):
- `--dry-run`: 실제 변경 없이 plan 출력. 운영자가 변경 사항 미리 검토
- `--check-only`: read-only 검증 (verify-keycloak-groups.sh 와 동일 효과). staging 적용 직전 PASS 확인
- `--remove`: group 4종 + composite role 모두 제거 (rollback 용). staging reset / cleanup 시 사용

## 5. cross-tier (사외 / 사내 정합)

| 영역 | Tier | 비고 |
|---|---|---|
| ADR-0036 본문 | 공용 | 사내 한정 정보 미포함 |
| release_v0-1_roadmap §3.5 X-6 row | 공용 | status 만 update |
| traceability/report.md §6 row | 공용 | row 만 추가 |
| memory 4 file | 공용 | docs only |
| sprint plan | 공용 | docs only |
| `scripts/setup-keycloak-groups.sh` | **사내** | Keycloak admin API secret 사용 |
| runbook (`infra/operations/keycloak-groups-staging-prod-runbook.md`) | **사내** | staging/prod instance URL + admin client_id/secret |
| ADR-0036 사내 한정 append | **사내** | Keycloak instance URL + client secret rotation |
| staging Keycloak admin console 적용 | **사내** (사용자 수동) | N/A (admin console) |
| prod Keycloak admin console 적용 | **사내** (사용자 수동) | N/A (admin console) |

## 6. 검증

### 6.1 사외 PR

- docs lint = N/A (governance 문서)
- CI 4/4 PASS (path-detect + workflow-lint + migration-prefix + openapi lint; backend/frontend/e2e skip)
- tier self-check: 사내 한정 패턴 (`DEVHUB_KEYCLOAK_*` / `GITEA_URL` / `internal-registry.example.com` / `kc.internal.example.com` / `devhub.example.com` / `172.16.0.0/12`) 미포함 PASS

### 6.2 사내 PR

- `bash scripts/setup-keycloak-groups.sh --check-only --realm=devhub` → 4/4 group + composite role 1:1 검증
- staging Keycloak dry-run → apply → verify-keycloak-groups.sh PASS
- prod Keycloak 동일 (staging 1주 관찰 후)

## 7. supersession

- ADR-0019 §5.3 (Keycloak-only IdP 잔여 carve) 와 cross-ref. 본 ADR = §5.3 의 staging-prod 적용 정공법.
- ADR-0026 (Keycloak role excluded decision) 와 cross-ref. ADR-0026 = Keycloak realm role/group membership 이 DevHub RBAC 에 직접 영향 없음. 본 ADR = composite role 을 통한 token role 추출만 사용 (backend 변경 없음).
- [`docs/setup/keycloak_operations.md` §4.3 + §4.4](../setup/keycloak_operations.md) 와 cross-ref. 본 ADR = §4.3/§4.4 의 staging-prod 적용 + setup script 정공법.

## 8. 변경 이력

| 일자 | 변경 | sprint |
|---|---|---|
| 2026-06-14 | 1차 발행 (Accepted). X-6 Keycloak group staging-prod 적용 + 2-PR 분할 (사외 docs / 사내 script) + setup script idempotent + dry-run + check-only + remove flag. | `docs/x6-keycloak-groups-public` / `infra/x6-keycloak-groups-internal` |
