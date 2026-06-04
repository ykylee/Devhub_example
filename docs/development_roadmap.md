# DevHub 통합 개발 로드맵

> **2026-05-20 이후 진입 자산**: 본 문서는 M0~M6 의 historical 마일스톤 (done 항목 사후 명문화) source-of-truth. **v1.0 릴리즈 + 후속 작업의 신규 source-of-truth 는 [`docs/planning/release_v1_roadmap.md`](./planning/release_v1_roadmap.md)** — 잔여 carve 통합 인벤토리 (P0~P3) + 신규 마일스톤 (M-v1.0, M-v1.1, M-v2) + 워커 분업 (Claude/Codex/Gemini) 매트릭스를 포함한다.

- 문서 목적: DevHub 프로젝트의 전체 개발 방향을 단일 진입점에서 정리한다. 백엔드·프론트엔드·인증/IdP·운영 트랙이 동일 마일스톤 체계 위에서 진행되도록 하는 1차 참조 문서.
- 범위: 머지된 PR #12 이후 시점부터 다음 단계 작업의 마일스톤·우선순위·의존 관계. 트랙별 *세부* 작업은 각 트랙의 세부 로드맵에서 관리.
- 대상 독자: 프로젝트 리드, 백엔드/프론트엔드 개발자, 운영 담당자, 후속 작업자
- 상태: draft
- 최종 수정일: 2026-05-29 (SDLC 재정비 sprint #408~#416 — code-taxonomy SoT 의 10 core 도메인 + Shared + Infrastructure 4 계층 구조 반영. §4 트랙별 매핑을 **도메인별 트랙** 으로 재구성. M2~M7 historical 결과는 그대로 보존.)
- 관련 문서:
  - 백엔드 세부 로드맵: [`docs/backend_development_roadmap.md`](../docs/backend_development_roadmap.md)
  - 프론트엔드 세부 로드맵: [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md)
  - 요구사항 (master index): [`./requirements.md`](./requirements.md), 도메인별 SoT: [`./domain/`](./domain/README.md)
  - 시스템 설계 (master index): [`./architecture.md`](./architecture.md)
  - API 계약 (master index): [`./backend_api_contract.md`](./backend_api_contract.md), 공통 규약: [`./api/conventions.md`](./api/conventions.md)
  - 거버넌스 — code taxonomy SoT: [`./governance/code-taxonomy.md`](./governance/code-taxonomy.md), 문서 표준: [`./governance/document-standards.md`](./governance/document-standards.md)
  - 추적성 매트릭스: [`./traceability/report.md`](./traceability/report.md)
  - Shared / Infrastructure 진입점: [`./shared/README.md`](./shared/README.md), [`./infrastructure/README.md`](./infrastructure/README.md)
  - 인증 ADR: [`./adr/0019-keycloak-only-idp.md`](./adr/0019-keycloak-only-idp.md) (현재 결정), [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) (Hydra+Kratos, superseded)
  - 보안 리뷰: [`../ai-workflow/memory/codebase-security-review-2026-05-08.md`](../ai-workflow/memory/codebase-security-review-2026-05-08.md)

---

## 0. 사용 가이드

본 문서는 **모든 개발자가 작업 전 가장 먼저 읽는** 단일 진입점이다.

1. 본인 트랙(백엔드/프론트엔드)의 다음 마일스톤이 무엇인지 §3 마일스톤 표에서 확인한다.
2. 마일스톤 안의 작업 항목은 §4 트랙별 세부 표에서 출처·의존을 확인한다.
3. *그 작업의 구현 디테일* 은 백엔드/프론트엔드 세부 로드맵에서 관리한다 — 본 문서는 일정·우선순위·교차 의존만 다룬다.
4. 본 문서가 다른 문서(요구사항, 설계, ADR)와 충돌하면 §6 충돌 해소 표를 source-of-truth 로 본다.
5. 신규 결정이 생기면 §7 변경 이력에 한 줄 추가한다.

세부 로드맵과 본 통합 로드맵의 역할 분담:

| 본 문서 (통합) | 세부 로드맵 (트랙별) |
| --- | --- |
| 마일스톤 정의·일정·의존 | 작업 단위 세부, 코드 위치, 검증 절차 |
| 트랙 간 contract / 충돌 해소 | 트랙 내부 phase 진행 |
| 우선순위(P0~P3)·완료 정의 | DoD 의 구현 디테일 |

---

## 1. 트랙 정의

| 트랙 | 책임 영역 | 세부 로드맵 |
| --- | --- | --- |
| **B / Backend** | Go Core API, store, normalize, command worker, realtime hub | [`docs/backend_development_roadmap.md`](../docs/backend_development_roadmap.md) |
| **F / Frontend** | Next.js (역할별 기본 진입 우선순위 대시보드, 조직, 인증 UI, 실시간 통합, RBAC UI) | [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md) |
| **A / Auth & IdP** | Keycloak (단일 IdP), 토큰 검증, 권한 가드. ADR-0019 (현재) / ADR-0001 (Hydra+Kratos, superseded) | [`./adr/0019-keycloak-only-idp.md`](./adr/0019-keycloak-only-idp.md) (current), [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) (superseded) |
| **X / Cross / Contract** | API 계약, 메시지 envelope, role wire format, 데이터 모델 | [`./backend_api_contract.md`](./backend_api_contract.md) |

## 3. 기능 단위별 마일스톤 (Milestones by Functional Units)

### M2 — 인증 및 계정 기반 완성 (DONE, 2026-05-12)

인증/계정 및 사용자 관리의 핵심 흐름이 모두 완성되었다. 1차 완성 sprint (`claude/login_usermanagement_finish`)를 통해 UX 결함과 audit 정합성, 그리고 CI 자동화 파이프라인까지 구축하여 운영 진입 게이트를 통과했다.

- **로그인 / 로그아웃 흐름**:
  - ✅ **B**: Keycloak OIDC 토큰 검증 경계 및 actor 매핑 정착.
  - ✅ **F**: `/auth/login`, `/auth/callback` OIDC code flow 기반 세션 연동 (PR-LOGIN-1~4, PR #33·#34·#45·#51).
- **사용자 관리 기능 (User Management)**:
  - ✅ **B**: `/api/v1/users` CRUD + 조직원 연동 (Phase 12).
  - ✅ **B**: 시스템 관리자용 `/api/v1/accounts` 발급/잠금/재설정/회수 4 endpoint (PR #54).
  - ✅ **F**: `/account` 개인 비밀번호 변경 (PR #50).
  - ✅ **F**: `/admin/settings` shell + users/organization/permissions sub-routes (PR #52·#53).
- **조직 관리 기능 (Org Management)**:
  - ✅ **B·F**: 부서 계층 구조, 멤버 배정, 드래그/리더 변경 (Phase 12 + PR #55).
- **RBAC enforcement**:
  - ✅ **B·F**: per-resource 4-boolean matrix + `requirePermission` enforcement (M1 RBAC track, PR #20·21·22·23·27·29·30·31).

#### 1차 완성 sprint 완료 (`claude/login_usermanagement_finish`)

- ✅ **PR-UX1**: `/admin/settings/users` SearchInput 실 필터링 (DONE).
- ✅ **PR-UX2**: `/account` 비밀번호 변경/재인증 UX 안내 (DONE).
- ✅ **PR-UX3**: Header Switch View 한계 안내 (DONE).
- ✅ **PR-M2-AUDIT**: 인증 이벤트 → DevHub `audit_logs` 통합 (DONE).
- ✅ **PR-T4**: GitHub Actions CI 구축 및 35종 E2E 테스트 자동화 (DONE).

세부는 [sprint_plan](../ai-workflow/memory/claude/login_usermanagement_finish/sprint_plan.md) 참조.

#### M2 명시적 out-of-scope (별도 sprint)

- OIDC JWKS/introspection verifier 실구현 → 보안 강화 sprint.
- Sign Up (셀프 가입, 인사 DB 연동) → M3.
- MFA / Two-Factor → M4.

### M3: 사용자 및 조직 관리 (User & Org Management) — In Progress (대부분 M2 1차 완성 흡수)

> **drift 정합 (2026-05-13, sprint `claude/work_260513-k`)**: 본 §3 가 M3/M4 정의의 source-of-truth. 매트릭스 §2.3 + state.json + backend_roadmap §5 모두 본 절 기준으로 정합화.

**M2 1차 완성 sprint (`claude/login_usermanagement_finish`, PR #85) 가 흡수한 항목** (이전에는 M3 으로 분류):
- ✅ 사용자 관리 (유저 CRUD UI 고도화 + 권한 할당 정교화) — `/admin/settings/users` + RBAC PermissionEditor.
- ✅ 조직 관리 1차 완성 (부서 CRUD + drag&drop + 리더 변경 영속화) — work_26_05_11 트랙 S (PR #52~#55).
- ✅ CI/CD GHA (Unit + E2E + actionlint) — PR #86~#88 + ADR-0005.

**M3 잔여**:
- ⏳ **Sign Up (셀프 가입)**: 인사 DB 연동 기반 사용자 셀프 등록 (IdP admin API 연계 hrdb lookup arm).
    - 대상: 이름, 사내 ID, 사번, 부서명이 인사 DB에 존재하는 인원.
- ⏳ **인사 DB 스키마 (초기)**: `name`, `system_id`, `employee_id`, `department_name`. `internal/hrdb/` 모듈 활용.
- ⏳ **조직 polish**: 본 sprint 시리즈가 carve out 한 `backend_api_contract.md` §10.4 의 자세한 schema, `parent_id` 검증, primary_dept 자동 판정 등 (§5 백로그 항목).

### M4: 실시간 대시보드 및 운영 고도화 (Realtime & Ops)
- **실시간 데이터 (WebSocket 확장 + replay)**: `infra.node.updated`, `ci.run.updated`, `risk.updated` event publish + 리소스 필터링 + last event replay. backend_roadmap §2 Phase 8 잔여 항목.
- **command status WebSocket UI** (frontend Phase 4 마무리): command lifecycle 상태 변화의 UI 실시간 반영.
- **과제 추적**: Gitea PR/Commit 기반 추적 화면 + Hourly Reconciliation (backend_roadmap §2 Phase 10).
- **시스템 관리자 대시보드 (System Admin Dashboard)**:
  - ⏳ **B·F**: Gitea Runner 상태, 시스템 설정 관리 UI 및 API.
- **권한 관리 고도화 (RBAC)**:
  - ✅ **B·F**: 권한 매트릭스 및 역할 할당 기능 완료 (M1).
  - ⏳ **B**: ADR-0007 PostgreSQL `LISTEN/NOTIFY` 기반 PermissionCache 다중 인스턴스 일관성 구현.
  - ⏳ **A**: 외부 SSO 통합 (Gitea 연동 등).
- **역할별 UX 제공 방식 정렬**:
  - ⏳ **F·X**: 역할별 UX는 기본 진입 페이지 우선순위로 제공하고, 시스템 영역은 `system_admin` 권한 전용 노출 정책으로 유지.

### M5: 개발 의뢰 (Dev Request, DREQ) — Concept staged (sprint `claude/work_260515-f`)

외부 시스템 → DevHub → application/project 으로 이어지는 upstream intake 흐름. 컨셉/요구사항/Usecase/설계/API contract 1차 stage 완료 (본 sprint). 후속 sprint hook:

- ✅ **A (ADR)**: 외부 수신 endpoint 인증 정책 — **[ADR-0012](./adr/0012-dreq-external-intake-auth.md) (sprint `claude/work_260515-g`, accepted 2026-05-15)** 가 옵션 A (API 토큰 + IP allowlist) 채택. B (HMAC) / C (OAuth) 는 후속 마이그레이션 경로.
- ✅ **B**: backend 1차 — `domain.DevRequest` / store / handler / migration 000022 dev_requests + 000023 dev_request_intake_tokens + `requireIntakeToken` middleware + API-59..65 활성화 (sprint `claude/work_260515-i`).
- ✅ **F**: 담당자 dashboard 의 "내 대기 의뢰" 위젯 + `/admin/settings/dev-requests` 페이지 + Promote-to-Platform/Project 연계 1차 (sprint `claude/work_260515-j`).
- ✅ **A (ADR)**: PMO Manager / 담당자 위양 정책 — **[ADR-0013](./adr/0013-dreq-rbac-row-scoping.md) (sprint `claude/work_260515-m`, accepted 2026-05-15)** 가 ADR-0011 §4.2 helper 의 dev_requests resource 적용 사례 사후 명문화. handler wire-up 은 PR #124 에 도입 완료.
- ✅ **B (Promote-Tx)**: API-62 promote 의 단일 트랜잭션 (신규 application/project 생성 + dev_request 상태 갱신 + audit) — sprint `claude/work_260515-m` (REQ-FR-DREQ-005 정합 완성, ADR-0013 §5).
- ✅ **B (Admin-UI backend)**: intake token 발급/revoke/list admin endpoint (API-66..68) + ADR-0014 + migration 000026 RBAC seed — sprint `claude/work_260515-o` (carve 2/4 part 1).
- ✅ **F (Admin-UI frontend)**: `/admin/settings/dev-request-tokens` 페이지 + IntakeTokenTable + IssueIntakeTokenModal (plain-1회-노출 reveal phase) + dev_request_token service/types — sprint `claude/work_260515-p` (carve 2/4 part 2).
- ✅ **B·F·X (E2E)**: TC-DREQ-* 13건 정식 발급 + `dev-requests.spec.ts` 6 step + 신규 2 test — sprint `claude/work_260518-d` (PR #144). 추가 회귀: dev-requests.spec PATCH page.evaluate fetch 정합 (sprint -m hotfix #5, PR #154).
- ✅ **B (Atomicity)**: `UpdateDevRequestIntakeTokenIPs` 단일 CTE + `FOR UPDATE` row lock + concurrent race test — sprint `claude/work_260518-o` (PR #156, [ADR-0017 §6 atomicity](./adr/0017-dreq-intake-token-operational-hardening.md) resolved).
- ✅ **B (Cron + Metric)**: 자동 만료 token revoke + 만료/staleness Prometheus metric (devhub_intake_token_expiring_soon/_stale/_auto_revoked_total) — sprint `claude/work_260518-t` (PR #161, [ADR-0017 §6](./adr/0017-dreq-intake-token-operational-hardening.md) (a)+(c)+(d) resolved).
- ✅ **B·F**: PATCH expires_at + admin UI 편집 modal — [ADR-0017 §6](./adr/0017-dreq-intake-token-operational-hardening.md) (b) **resolved** (PR #137 `EditIntakeTokenModal` + backend `intakeTokenAdminUpdateRequest.ExpiresAt`, issue #219 closed 2026-05-21).
- ✅ **F**: DREQ → notification 연계 (Header Bell 배지 + Promote-to-Project 프리필) — PR #323 (sprint `codex/work_260526-b`), TC-DREQ-NOTI-01..03.
- ⏳ **B (carve)**: 외부 시스템 callback (webhook 송신) — MVP 안정화 후 (v1.1).

> **M5 DREQ closing 확정 (2026-05-27)**: intake auth + promote-tx + token admin(발급/revoke/PATCH/cron) + RBAC row-scoping + frontend(목록/상세/위젯/token admin) + notification 연계 + TC-DREQ-* 모두 완료. 잔여 = 외부 callback(webhook 송신) v1.1 carve.

문서 hub: [`docs/domain/dev-request/concept.md`](./domain/dev-request/concept.md), 추적성 [`docs/traceability/report.md §2/§3 DREQ`](./traceability/report.md).

### M6: External Integration — 1차 종합 closing (2026-05-15 ~ 18)

외부 ALM/SCM/CI-CD/문서/HomeLab 시스템과의 provider/binding/snapshot 통합. concept staged → backend 1차 → ADR 3 신규 → frontend 진입점 → API-80 DELETE → bindings UI → topology v2 → 운영 자산 (Alertmanager 가이드 + Grafana JSON).

- ✅ **B (Concept)**: External Integration 컨셉 (REQ-FR-INT + REQ-NFR-INT + UC-INT + ARCH-INT-01..06 + API-69..78 spec) — sprint `codex/memory-next-step-20260515` (PR #135).
- ✅ **B (Backend 1차)**: HomeLab pull adapter (file + HTTP) + Prometheus `/metrics` + `integration_registry` + `infra_service_snapshots` + API-73..78 activated — sprint `codex/next-step-20260516` (PR #139).
- ✅ **A (ADR)**: [ADR-0015 HomeLab pull strategy](./adr/0015-homelab-adapter-pull-strategy.md) + [ADR-0016 Prometheus alerts policy](./adr/0016-prometheus-alerts-policy.md) — sprint `claude/work_260518-c` (PR #143).
- ✅ **F (Provider UI)**: `/admin/settings/integrations` 페이지 + ProviderTable + ProviderModal + API-69~72 service — sprint `claude/work_260518-g` (PR #148).
- ✅ **B·F (API-80 DELETE)**: DELETE provider endpoint + FK guard + Delete UI — sprint `claude/work_260518-j` (PR #151).
- ✅ **F (Bindings UI)**: `/admin/settings/integration-bindings` 페이지 + BindingsTable + CreateBindingModal + API-74/75 service — sprint `claude/work_260518-m` (PR #154).
- ✅ **F (Topology v2)**: `/admin/topology-v2` 페이지 + React Flow + nodes/services/edges + degraded providers banner + snapshot_at 메타 — sprint `claude/work_260518-n` (PR #155).
- ✅ **B (Size limit + token rotation SOP)**: HomeLabFilePuller/HTTPPuller MaxBytes + streaming decode + agent token rotation SOP — sprint `claude/work_260518-p` (PR #157, [ADR-0015 §6](./adr/0015-homelab-adapter-pull-strategy.md) (1)+(2) resolved).
- ✅ **X (Alertmanager + Grafana)**: `docs/setup/prometheus_alertmanager_setup.md` + `docs/setup/grafana/homelab_dashboard.json` — sprint `claude/work_260518-s` (PR #160, [ADR-0016 §6](./adr/0016-prometheus-alerts-policy.md) (1)+(2) resolved).
- ⏳ **B (carve)**: dedicated worker binary — M4 진입 시 재평가 ([ADR-0015 §6](./adr/0015-homelab-adapter-pull-strategy.md) (3)).
- ⏳ **B (carve)**: push/pull dedup 정책 — 별도 ADR ([ADR-0015 §6](./adr/0015-homelab-adapter-pull-strategy.md) (4)).
- ⏳ **X (carve)**: pull latency p95 alert / push 경로 webhook 알림 / stage→prod 임계 확정 — baseline 1주 관찰 후 ([ADR-0016 §6](./adr/0016-prometheus-alerts-policy.md) (3)+(4)+(5)).
- ⏳ **F (carve)**: React Flow group sub-node (services as node children) + WebSocket 실시간 갱신 (`infra.node.updated` / `infra.service.updated`) + v2 node click action.
- ⏳ **F (carve)**: bindings UI 강화 — scope_id lookup combobox + Edit/Delete binding + pagination.

API 인벤토리: **API-69..78 + API-80** 모두 activated (API-79 는 DREQ allowed_ips PATCH).
TC 인벤토리: **TC-INT-FRONTEND-* 12건** (LIST/CREATE/EDIT/SYNC/RBAC/DELETE/DELETE-NEG + BIND-{LIST,CREATE,RBAC} + TOPOLOGY-V2-{NAV,RBAC}) + **TC-INT-HOMELAB-03** active.

문서 hub: [`docs/domain/integration-registry/external_system_concept.md`](./domain/integration-registry/external_system_concept.md), [`docs/setup/homelab_agent_token_rotation.md`](./setup/homelab_agent_token_rotation.md), [`docs/setup/prometheus_alertmanager_setup.md`](./setup/prometheus_alertmanager_setup.md), 추적성 [`docs/traceability/report.md §3 External Integration`](./traceability/report.md).

> **M6 깊이 확장 (2026-05-26~27)**: 1차 종합 closing 이후 외부 연동 깊이가 대폭 확장됐다 — **Gitea SCM 동기화 워커**(pull, `internal/gitea/`, `integration_sync_jobs` 큐, RM-M4-06 1차, PR #341) + **provider 등록 UX 고도화**(vendor 템플릿 7종 + 가이드 자격증명 + base_url + 연결 테스트 API-87, PR #352) + **auth_mode full 모델**(token/basic/app_password/oauth2/agent + write-only auth_secret, migration 000041, PR #358) + **api_token write-only 슬롯**(000040, PR #355) + **webhook 헤더 alias**(X-Gitea/X-Gogs fallback) + **SCM↔시스템 repository 양방향 연동**(소유권 분리 000042 + import API-89 + create API-90 gitea + provider_id 단일화 000045, PR #363/#366/#373) + **repository draft→publish lifecycle**(000043, API-91/92, PR #368) + **admin catalog UI**(PR #357/#361). 향후 방향은 [v1.0 릴리즈 로드맵](./planning/release_v1_roadmap.md) §3 + [코드베이스 스냅샷 §06 향후 방향](./analysis/2026-05-27-codebase-snapshot/06_future_direction.md) 참조.

### M7: 사용자 초기 등록 (Onboarding) — Concept/Requirements/Design/ADR closing (2026-05-21)

Keycloak 인증 통과 + DevHub 프로필 미완료 사용자의 self-service 초기 등록 흐름. 컨셉/요구사항/Usecase/설계/API contract/ADR 1차 stage 완료 (2026-05-21 sprint 5건 누적). IMPL carve 4건은 후속 (RM-ONBOARD-01..04, M-v1.1 진입).

- ✅ **A (Concept)**: 컨셉 1차 + skip-and-resume (§5.9, §8 #7 결정) — sprint `claude/keycloak-user-onboarding-concept` (PR #260) + `claude/keycloak-onboarding-concept-2026-05-21` (PR #265).
- ✅ **A (Requirements)**: REQ-FR-ONBOARD-001..012 + REQ-NFR-ONBOARD-001..008 (`docs/requirements.md §5.7`) — sprint `claude/onboarding-requirements-2026-05-21` (PR #266).
- ✅ **A (Design)**: UC-ONBOARD-01..11 (`system_usecases.md §2.13`) + ARCH-ONBOARD-01..06 (`architecture.md §9`) + API-83..86 + API-32/33 확장 (`backend_api_contract.md §16`) — sprint `claude/onboarding-arch-2026-05-21` (PR #267).
- ✅ **A (ADR)**: [ADR-0021 Onboarding self-service unit selection + lazy auto-create supersession](./adr/0021-onboarding-self-service-unit-selection.md) — sprint `claude/onboarding-adr-2026-05-21` (PR #269). ADR-0020 partial supersession (5 위치).
- ✅ **A (Plan)**: IMPL carve 4건 분할 plan ([`docs/domain/onboarding/impl_plan.md`](./domain/onboarding/impl_plan.md)) + RM-ONBOARD-01..04 발급 — 본 sprint `claude/onboarding-impl-carve-plan-2026-05-21`.
- ✅ **B (Backend)**: RM-ONBOARD-01 — migration 000033 + `onboardingGate` middleware + 5 handler (API-83/84/85/86 + API-32/33 확장) + audit event const. **Carve A 완료 (PR #278)**. lazy_auto_create.go 폐기는 Carve D 후 #290.
- ✅ **F (Frontend)**: RM-ONBOARD-02 — `/onboarding` page + OrganizationPicker + skip flag + dismissible banner + `(dashboard)/layout` 3-branch gating + `/account` self-service unit edit. **Carve B/C 완료 (PR #288)**.
- ✅ **F (Admin UI)**: RM-ONBOARD-03 — `/admin/settings/users` 의 "Confirm Review" 액션 + `ConfirmReviewModal` + pending_review filter. **Carve B/C 완료 (PR #288)**.
- ✅ **T (Tests)**: RM-ONBOARD-04 — UT-onboarding-* (backend) + TC-ONBOARD-* (E2E `onboarding-first-login.spec.ts`) + 6 test seed. **Carve D 완료 (PR #289)** + feature flag default ON flip & `lazy_auto_create.go`/`onboarding_feature_flag.go` 삭제 (PR #290) + codex hotfix #3 AuthGuard whitelist→blocklist (PR #291).

API 인벤토리: **API-83..86 activated** + **API-32 / API-33** 확장.
TC 인벤토리: **TC-ONBOARD-* active** (`onboarding-first-login.spec.ts`).

> **M7 Onboarding 풀스택 closing 확정 (2026-05-27)**: Carve A(backend) → B/C(frontend+admin) → D(tests) 전부 머지 + feature flag default ON + lazy_auto_create 폐기(ADR-0021 §3.3 정공법). 사내 잔여 = staging 1주 monitoring (운영 검증).

문서 hub: [`docs/domain/onboarding/concept.md`](./domain/onboarding/concept.md), [`docs/domain/onboarding/impl_plan.md`](./domain/onboarding/impl_plan.md), [ADR-0021](./adr/0021-onboarding-self-service-unit-selection.md), 추적성 [`docs/traceability/report.md §3 Onboarding`](./traceability/report.md).

### Design 검토 (Phase 1 planning, 결정 후 ADR 승격 예정)

- 📋 **[`docs/infrastructure/deployment-automation/single_port_reverse_proxy.md`](./infrastructure/deployment-automation/single_port_reverse_proxy.md)** — 외부 단일 포트 reverse proxy (nginx + `/devhub` prefix + backend/Hydra/Kratos sub-path 매핑). sprint `claude/work_260518-u` (PR #162). 결정 후 **ADR-0018** 승격 + Phase 2 staging.
- 📋 **[`docs/infrastructure/keycloak-idp/sso_federation.md`](./infrastructure/keycloak-idp/sso_federation.md)** — Keycloak 을 Kratos upstream OIDC provider 로 federation. sprint `claude/work_260518-v` (PR #163). HRDB user mapping = employee_id strict link. RM-M4-09 구체화. 결정 후 **ADR-0019** 승격 + Phase 2 staging.

---

## 4. 트랙별 세부 작업 매핑 — 도메인별 (code-taxonomy SoT 정합)

> **2026-05-29 재구성**: 이전 (4.1/4.2) Backend·Frontend 평면 트랙은 폐기. [`docs/governance/code-taxonomy.md`](./governance/code-taxonomy.md) 의 **3대 레이어 (Domain / Shared / Infrastructure)** + **도메인 내부 4대 계층 (view/service/repository/schema)** 을 따른 도메인별 트랙으로 재정렬. 각 도메인 row 의 RM ID 는 [`./traceability/report.md`](./traceability/report.md) §2.3 (Roadmap 트랙) 에 대응.

### 4.1 Domain — 10 core 도메인 (비즈니스 핵심)

| # | 도메인 | 진입점 | 핵심 작업 (완성/잔여) | 마일스톤 | 상태 |
| --- | --- | --- | --- | --- | --- |
| 1 | `auth-session` | [`./domain/auth-session/`](./domain/auth-session/README.md) | Keycloak OIDC PKCE + JWKS verifier + stale-while-error fallback (ADR-0019, ADR-0020) + 토큰 refresh + AuthGuard | M2 + M4 | done |
| 2 | `audit-ops` | [`./domain/audit-ops/`](./domain/audit-ops/README.md) | `audit_logs` + request_id + source enrichment + Keycloak event polling (ADR-0020 sub-carve E, PR #189~#193) | M1 + M4 | done |
| 3 | `rbac-permissions` | [`./domain/rbac-permissions/`](./domain/rbac-permissions/README.md) | per-resource 4-boolean matrix + `requirePermission` + PermissionCache LISTEN/NOTIFY (ADR-0002, ADR-0007, ADR-0011 row-scoping) | M1 | done |
| 4 | `organization-management` | [`./domain/organization-management/`](./domain/organization-management/README.md) | users/org_units CRUD + single-leader invariant + appointments + HRDB lookup (ADR-0008/0009/0010) | M2 + M3 | done |
| 5 | `onboarding` | [`./domain/onboarding/`](./domain/onboarding/README.md) | gate middleware + submit/search/admin review (API-83..86) + 상태머신 (ADR-0021) + lazy_auto_create 폐기 | M7 | done |
| 6 | `platform-lifecycle` | [`./domain/platform-lifecycle/`](./domain/platform-lifecycle/README.md) | Platform/Project CRUD + 상태머신 + rollup + RBAC row-scoping (ADR-0011, ADR-0014) | M3 + M-v1.0 | done. 잔여 carve: PlatformRepository decouple / ApplicationStore slim |
| 7 | `repository-integration` | [`./domain/repository-integration/`](./domain/repository-integration/README.md) | Repository CRUD + draft→publish lifecycle (#368) + SCM 양방향 import/create (API-89/90, #363/#366/#373) + provider_id 단일화 | M6 | done. 잔여: #368 무테스트 보강 (N-2) |
| 8 | `dev-request` | [`./domain/dev-request/`](./domain/dev-request/README.md) | intake auth (ADR-0012) + promote-tx + token admin (ADR-0014) + 만료 cron (ADR-0017) + RBAC row-scoping (ADR-0013) | M5 | done. 잔여: 외부 callback (webhook 송신, v1.1) |
| 9 | `integration-registry` | [`./domain/integration-registry/`](./domain/integration-registry/README.md) | provider/binding registry + auth_mode full (token/basic/app_password/oauth2/agent) + base_url + write-only api_token + 연결테스트 (API-87) + Gitea sync worker + HomeLab pull (ADR-0015) + Task ingestion (REQ-FR-TASK) | M6 | done. 잔여: webhook header alias 강화 / Task ingestion 구현 |
| 10 | `realtime` | [`./domain/realtime/`](./domain/realtime/README.md) | WebSocket Hub + ticket 인증 (ADR-0024, #344/#348) + command.status.updated publish | M4 (부분) | 부분. 잔여: **RM-M4-01** infra/ci/risk event publish + **RM-M4-02** replay + resource scope filter |

### 4.2 Shared — 공통 기반

| 모듈 | 코드 | 진입점 | 작업 | 상태 |
| --- | --- | --- | --- | --- |
| `config` | `backend-core/internal/shared/config/` + `frontend/shared/config/` | [`./shared/`](./shared/README.md) | 전역 환경 설정 로더 + endpoints | done |
| `logger` | (system 표준) | [`./shared/`](./shared/README.md) | 로그 수집 어댑터 | done |
| `utils` | `backend-core/internal/shared/httphelp/` + `frontend/shared/utils*` | [`./shared/`](./shared/README.md) | 공통 유틸리티 helper | done |
| `ui-foundation` | `frontend/shared/ui-foundation/{components,layout}/` | [`./shared/`](./shared/README.md) | Modal/Badge/Toast/PageState/FilterBar/ComboBox/DestructiveConfirmModal + Header/Sidebar/AuthGuard | done (PR #248 polish 1차) |
| `integrationcaps` | `backend-core/internal/shared/integrationcaps/` (PR #409) | [`./shared/`](./shared/README.md) | provider capability gate OR semantics 공용 helper (3 카피 통합 + 11 unit test) | done (2026-05-29) |

### 4.3 Infrastructure — 외부 기술 구현체

| 모듈 | 코드 | 진입점 | 작업 | 마일스톤 | 상태 |
| --- | --- | --- | --- | --- | --- |
| `keycloak-idp` | `internal/auth/keycloak_verifier.go`, `infra/idp/keycloak-event-listener-spi/`, `infra/idp/sql/` | [`./infrastructure/`](./infrastructure/README.md) | JWKS + admin client + event listener SPI (ADR-0019, ADR-0020, ADR-0022, ADR-0023) | M4 | done |
| `gitea-scm` | `internal/infrastructure/gitea/`, `internal/normalize/gitea/` | [`./infrastructure/`](./infrastructure/README.md) | API 클라이언트 + 백그라운드 sync worker (#341) + webhook signature 검증 + JSON 정규화 | M4 | done. 잔여: **RM-M4-06** Hourly Pull 정밀화 (issue #231) |
| `hrdb` | `internal/infrastructure/hrdb/{postgres,mock}.go` | [`./infrastructure/`](./infrastructure/README.md) | 실 PG / mock 어댑터 (ADR-0008/0010) | M3 | done |
| `commandworker` | `internal/infrastructure/commandworker/{worker,live_worker}.go`, `serviceaction/executor.go` | [`./infrastructure/`](./infrastructure/README.md) | 명령어 폴링/실행 에이전트 + sandbox (mock/compose/k8s) | M1 + M4 | done |
| `infra-topology` | `internal/realtime/`, `frontend/app/admin/topology-v2/` | [`./infrastructure/`](./infrastructure/README.md) | React Flow + nodes/services/edges + degraded providers banner (PR #155) | M6 | done. 잔여: React Flow group sub-node + WebSocket 실시간 갱신 |
| `database-migration` | `backend-core/migrations/000001~000046_*.sql` | [`./infrastructure/`](./infrastructure/README.md) | golang-migrate SQL 마이그레이션 | continuous | done. 잔여: **N-5** prefix uniqueness CI guard 강화 (000042 동시 발급 충돌 이력) |
| `deployment-automation` | `scripts/`, `infra/nginx/`, `docker-compose.{deploy,}.yml` | [`./infrastructure/`](./infrastructure/README.md) | 배포 전처리 + Nginx 역프록시 + compose 구성 (ADR-0018 single-port) | M4 | done |

### 4.4 Cross-cutting (X)

| 작업 | 마일스톤 | 상태 |
| --- | --- | --- |
| API 계약 envelope/enum 통일 | M1 | done (`docs/api/conventions.md`) |
| 추적성 매트릭스 (REQ → UC → ARCH → API → RM → IMPL → UT → TC) | continuous | done (19 row, PR #414) |
| 거버넌스 — code-taxonomy SoT + document-standards | M-v1.0 | done (PR #406/#415) |
| 4 계층 view 카운트 보강 (단위테스트) | M-v1.0 | 부분. rbac/repo/dreq/intreg/org 도메인 90%+ (PR #412). 잔여: app-lifecycle 큰 modal coverage 70% (carve) |

### 4.5 보안 부채 (Cross-cutting)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| `credentials_ref` / `api_token` / `auth_secret` 평문 저장 → DEK + KMS 봉투 암호화 (#6) | M-v1.1 | P1 | release_v1 §3.3 |
| Keycloak group staging → prod 적용 (#214) | M-v1.0 (사내 운영자) | P1 | release_v1 §3.2 |
| Keycloak SPI realm events push 전환 (polling 30s → <1s) | M-v1.1 | P3 | ADR-0020 |

### 4.6 AI (v2)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Python AI gRPC 서버 1차 (AnalysisService 실 구현) | v2 | P3 | backend_roadmap §2 Phase 9 |
| AI Gardener suggestion 모델 + Go Core 연동 | v2 | P3 | backend_roadmap §5 P3 |
| Weekly report 생성 worker | v2 | P3 | frontend_integration §3.4 |
| AI 알림 중재 (집중 시간 보호) 모델 | v2 | P3 | requirements §4-3·§5.3-2 |

### 4.7 후속 carve out (별도 sprint, 2026-05-29 SDLC 재정비 sprint 결과)

| 항목 | 우선순위 | 사유 |
| --- | --- | --- |
| CI e2e + backend-integration job 복원 (`&& false` 제거) | P1 | refactor 정리 stabilize 후 |
| view 컴포넌트 큰 modal coverage 70% (app-lifecycle) | P1 | ApplicationCreationModal (57%) + ProjectCreationModal (39%) edit-mode + member CRUD |
| PlatformRepository cross-domain decouple | P2 | `*IntegrationRepository` embed 제거 (review agent P1) |
| ApplicationStore interface slim | P2 | 13+ integration 메서드 → integration-registry 도메인 이관 |
| §2 인덱스 도메인 분류 정합 | P2 | `traceability/report.md` §2 의 cross-cutting row → 새 도메인 row 정합 (Phase 4 scope 외) |

---

## 5. 백로그 (마일스톤 미배정 / 결정 필요)

다음 항목은 명세가 부족하거나 책임자 미정이라 마일스톤에 배정되지 않았다. 진입 시점에 ADR 또는 spec 작성 후 마일스톤으로 흡수.

| 항목 | 미정 부분 | 출처 |
| --- | --- | --- |
| `GET /api/v1/team/load`, `GET /api/v1/dashboard/velocity` | 데이터 source / 산출 기준 / 오너 | frontend_integration §6.3 |
| `GET /api/v1/me`, focus mode/notification settings 영속화 | 모델 / 저장 위치 | frontend_integration §3.1·P3 |
| Weekly report 생성 worker 실행 매체 | cron vs scheduled command | frontend_integration §3.4, api_contract §10 |
| 조직 도메인 — `parent_id` 검증, primary_dept 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수), 파견/겸임 1:N 테이블, `total_count` Materialized View | spec / 마이그레이션 | backend_requirements_org_hierarchy §1·2, organizational_hierarchy_spec §3 |
| 알림 등급화 (Info / Action Required) 모델 | 모델 / 라우팅 정책 | requirements §5.2-7 |
| 기술 태깅 Kudos 가시성 | RBAC matrix와의 매핑 | requirements §5.1-3 |
| 외부 부서 의존성 수동 등록 | UI / 모델 | requirements §5.2-6 |
| `architecture/README.md`, `planning/README.md` TBD 스텁 | 본 통합 로드맵 채택 후 산출물로 채움 | 양자 |
| **Platform/Project 도메인 (총괄 + 기간성 운영)** — 시스템 관리자 등록·관리 vs 일반 사용자 조회 분리 | REQ-FR 발급 완료 + 모듈별 Usecase/ERD 분리 카탈로그 완료. 다음: ARCH/API/마이그레이션 설계 진입 | [`domain/platform-lifecycle/project_concept.md`](./domain/platform-lifecycle/project_concept.md), [`planning/system_usecases.md`](./planning/system_usecases.md), [`planning/system_erd.md`](./planning/system_erd.md) (2026-05-13) |

---

## 6. 충돌 해소 (source-of-truth 정리)

`requirements.md`, `architecture.md`, `backend/requirements.md`, `backend_api_contract.md` 사이에 시점 차이로 인한 표현이 남아 있다. 본 표가 통합 로드맵의 source-of-truth 다.

| 주제 | 폐기된 표현 | 채택된 표현 | 결정 출처 |
| --- | --- | --- | --- |
| 인증/계정 구현 | 자체 `accounts` 테이블, 자체 7 endpoint (`requirements §2.5`, `architecture §6.2`, `backend/requirements §5`, `api_contract §11` historical) | 정책 invariant 만 보존, 구현은 **Keycloak (단일 IdP) OIDC** (DevHub `users` 는 organizational metadata master + `idp_subject` 캐시) | ADR-0001 (2026-05-07 Hydra+Kratos, superseded) → ADR-0019 (2026-05-19 Keycloak 단일화) |
| 브라우저↔서버 실시간 | gRPC stream (`backend/requirements §1`) | **REST snapshot + WebSocket** | requirements_review §3.1, frontend_integration §2.1 |
| 역할 wire 형식 | `DEVELOPER\|MANAGER\|ADMIN` (`backend/requirements §4`) | **`developer\|manager\|system_admin`** | api_contract §2, requirements_review §3.3 |
| 명령성 액션 응답 | boolean `ActionResponse` (`backend/requirements §2`) | **`command_id` + `command_status` lifecycle** | api_contract §9 |
| 환경 default | docker-compose default (`tech_stack §2`, 일부 `architecture`, `PROJECT_PROFILE`) | **native default**, docker 는 환경별 자산 (git 추적 외부) | PR #12 BLK-1 (2026-05-08), environment-setup §0 |
| Phase 8 상태 | "프론트 done" 만으로 완료 표기 (frontend_roadmap) | **백엔드 in_progress 가 source** — 인증/필터/replay 미완 | backend_roadmap §2 Phase 8 |
| RBAC 모델 | 1차원 `none\|read\|write\|admin` rank (`rbac.go defaultRBACPolicy`, `api_contract §6` legacy) | **per-resource 4-boolean** (`{view, create, edit, delete}`) — 5 resources (`infrastructure`, `pipelines`, `organization`, `security`, `audit`) | ADR-0002 + api_contract §12 (2026-05-08) |
| RBAC enforcement | `requireMinRole` 라우트별 정적 rank 비교 | **`requirePermission(resource, action)`** — 라우트-매핑 표 + DB-backed matrix + deny-by-default | ADR-0002 §4.3, api_contract §12.8·§12.9 (PR-G5 머지 시 발효) |

위 폐기 표현이 본문에 그대로 남은 위치(`requirements.md`, `architecture.md`, `backend/requirements.md`, `tech_stack.md`, `PROJECT_PROFILE.md`)는 M0~M1 의 문서 정리 작업으로 *재설계 박스* 또는 *링크 참조* 형태로 정리한다.

---

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-08 | 초판 작성. M0~M4 정의, 트랙 매핑, 충돌 해소 표 정리. | PR #12, #13 머지 직후. claude/merge_roadmap 브랜치. |
| 2026-05-08 | §6 충돌 해소 표에 RBAC 모델/enforcement 결정 2행 추가. | M1 PR-G1, ADR-0002 채택 반영. claude/m1-pr-g1-rbac-contract 브랜치. |
| 2026-05-12 | §3 M2 갱신 — 핵심 흐름(로그인/로그아웃/계정/RBAC) done 표기 + 1차 완성 sprint 잔여 5 PR 명시 + out-of-scope 분리. | claude/login_usermanagement_finish 진입. |
| 2026-05-13 | §5 백로그에 "Platform/Project 도메인 (총괄 + 기간성 운영)" 1행 추가. 컨셉 문서 staged 상태로 안내. | sprint `claude/work_260513-p`. |
| 2026-05-13 | Platform/Project 요구사항 고도화 반영 — REQ-FR-APP/REQ-FR-PROJ + 모듈별 UC/ERD 카탈로그(`planning/system_usecases.md`, `planning/system_erd.md`) 연결. 다음 단계(ARCH/API) 전환 기준 명시. | sprint `claude/work_260513-p` 외 |
| 2026-05-14 | Application 도메인 backend 1차 완성 — PR #104~#110. API-41..58 activated + 마이그레이션 000012~000018 + ADR-0011 (RBAC row-scoping) + CI backend-integration job 신설. 23 integration test. | sprint `claude/work_260514-*` |
| 2026-05-15 | M5 DREQ 도메인 1차 — concept (REQ-FR-DREQ + UC-DREQ + ARCH-DREQ + API-59..68) + ADR-0012 (intake auth) + backend (PR #124) + frontend (PR #125) + Promote-Tx + ADR-0013 + Admin-UI (ADR-0014). 외부 8 PR (#133~#140) 으로 docker packaging + 대시보드 + token expires_at + IP mutation. | sprint `claude/work_260515-*` (15 PR) |
| 2026-05-18 | **M5 DREQ closing** — TC-DREQ-* 13건 정식 발급 + ADR-0017 §6 atomicity + cron revoke + 만료/staleness metric. **M6 External Integration 1차 종합 closing** — provider lifecycle + bindings UI + topology v2 + API-80 DELETE + ADR-0015/0016/0017 신규 + 운영 자산 (Alertmanager + Grafana). codex hotfix #5/#6/#7/#8 cycle. | sprint `claude/work_260518-*` (24 PR 누적, EOD #1 12건 + post-EOD #1 6건 + post-EOD #2 6건) |
| 2026-05-18 | **Design 검토 2건 staged** — single port reverse proxy (ADR-0018 후보, sprint -u PR #162) + Keycloak SSO federation (ADR-0019 후보, sprint -v PR #163). RM-M4-09 구체화. 결정 후 Phase 2 staging 진입. | post-EOD #2 |
| 2026-05-27 | **M5/M6/M7 모두 closing 정합 (코드베이스 스냅샷)** — M7 Onboarding 풀스택 완료(Carve A/B/C/D PR #278/#288/#289/#290/#291 + lazy_auto_create 폐기) + M5 DREQ closing 확정(PATCH expires_at #137 + notification 연계 #323) + M6 External Integration 깊이 확장(Gitea SCM sync #341 / 등록 UX #352 / auth_mode full #358 / api_token #355 / SCM 양방향 #363/#366/#373 / repository draft·publish #368 / admin catalog #357/#361). §3 의 M5/M7 ⏳ → ✅ 정정 + M6 깊이 확장 note. 분석 근거 = [코드베이스 스냅샷](./analysis/2026-05-27-codebase-snapshot/README.md). | sprint `claude/work_260527-codebase-review-roadmap-refresh` |
| 2026-05-21 | **M7 Onboarding 도메인 1차 (Concept + Requirements + Design + ADR + IMPL plan closing)** — concept §5.9 skip-and-resume (PR #265) + REQ-FR-ONBOARD-001..012 / REQ-NFR-ONBOARD-001..008 (PR #266) + UC-ONBOARD-01..11 + ARCH-ONBOARD-01..06 + API-83..86 + API-32/33 확장 (PR #267) + ADR-0021 (PR #269, ADR-0020 partial supersession 5 위치) + codex hotfix PR #270 + IMPL carve plan (본 sprint, RM-ONBOARD-01..04). IMPL carve 4건 (backend / frontend / admin UI / tests) 은 M-v1.1 진입 (별도 sprint). | sprint `claude/onboarding-impl-carve-plan-2026-05-21` |
| 2026-05-29 | **SDLC 재정비 sprint (8 PR #408~#416)** — code-taxonomy SoT (10 core 도메인 + Shared + Infrastructure 4 계층) 정합 + §4 트랙 매핑을 도메인별로 재구성. (1) `integrationcaps` 통합 (PR #409, 3 카피 → 공용 helper + 11 unit test) (2) SDLC Phase 1~5: 도메인 디렉터리 골격 (PR #410) → planning/tests 도메인별 이관 (PR #411) → REQ/ARCH/API split (PR #413, 34 신규 + 3 master index 전환) → traceability §3 매트릭스 10 도메인 SoT 재구성 (PR #414, 19 row) → document-standards §4 위치 가이드 갱신 (PR #415) (3) view 컴포넌트 단위테스트 +210 (PR #412, rbac/repo/dreq/intreg/org 90%+ — app-lifecycle 모달 carve). 후속 carve 5건 (§4.7). main HEAD `273d9d4` (housekeeping #416 후 갱신). | sprint `claude/work_260529-k` |

---

## 8. 다음 단계

1. 본 문서를 PR 로 머지하면 `docs/README.md`, `docs/DOCUMENT_INDEX.md`, 트랙별 세부 로드맵 상단에 진입점 링크가 추가된다.
2. M0 sprint 진입 직전, 본 문서 §3.1 의 DoD 항목별로 backlog 항목을 분해한다 (`ai-workflow/memory/claude/<branch>/backlog/<date>.md`).
3. 트랙 별 세부 로드맵은 본 문서의 마일스톤·우선순위 분류를 따르도록 갱신한다 — phase 표가 본 문서의 M0~M4 와 어떻게 매핑되는지 표 1개를 상단에 둔다.
4. `architecture/README.md`, `planning/README.md` 는 본 통합 로드맵 채택 후 산출물로 채운다.
