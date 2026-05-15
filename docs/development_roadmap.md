# DevHub 통합 개발 로드맵

- 문서 목적: DevHub 프로젝트의 전체 개발 방향을 단일 진입점에서 정리한다. 백엔드·프론트엔드·인증/IdP·운영 트랙이 동일 마일스톤 체계 위에서 진행되도록 하는 1차 참조 문서.
- 범위: 머지된 PR #12 이후 시점부터 다음 단계 작업의 마일스톤·우선순위·의존 관계. 트랙별 *세부* 작업은 각 트랙의 세부 로드맵에서 관리.
- 대상 독자: 프로젝트 리드, 백엔드/프론트엔드 개발자, 운영 담당자, 후속 작업자
- 상태: draft (2026-05-08 신규 작성)
- 최종 수정일: 2026-05-13 (Application/Project 도메인 요구사항 고도화 — Usecase/ERD 산출물 반영)
- 관련 문서:
  - 백엔드 세부 로드맵: [`ai-workflow/memory/backend_development_roadmap.md`](../ai-workflow/memory/backend_development_roadmap.md)
  - 프론트엔드 세부 로드맵: [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md)
  - 요구사항: [`./requirements.md`](./requirements.md)
  - 시스템 설계: [`./architecture.md`](./architecture.md)
  - API 계약: [`./backend_api_contract.md`](./backend_api_contract.md)
  - 인증 ADR: [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md)
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
| **B / Backend** | Go Core API, store, normalize, command worker, realtime hub | [`ai-workflow/memory/backend_development_roadmap.md`](../ai-workflow/memory/backend_development_roadmap.md) |
| **F / Frontend** | Next.js (역할별 기본 진입 우선순위 대시보드, 조직, 인증 UI, 실시간 통합, RBAC UI) | [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md) |
| **A / Auth & IdP** | Ory Hydra + Kratos, 토큰 검증, 권한 가드. ADR-0001 결정 사항 | [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) |
| **X / Cross / Contract** | API 계약, 메시지 envelope, role wire format, 데이터 모델 | [`./backend_api_contract.md`](./backend_api_contract.md) |

## 3. 기능 단위별 마일스톤 (Milestones by Functional Units)

### M2 — 인증 및 계정 기반 완성 (DONE, 2026-05-12)

인증/계정 및 사용자 관리의 핵심 흐름이 모두 완성되었다. 1차 완성 sprint (`claude/login_usermanagement_finish`)를 통해 UX 결함과 audit 정합성, 그리고 CI 자동화 파이프라인까지 구축하여 운영 진입 게이트를 통과했다.

- **로그인 / 로그아웃 흐름**:
  - ✅ **B**: `/api/v1/auth/{login,logout,token,signup,consent}` Kratos/Hydra 프록시.
  - ✅ **F**: `/auth/login` 패스워드 폼, `/auth/callback` 토큰 교환 + 저장, Header Sign Out → Hydra `/oauth2/sessions/logout` (PR-LOGIN-1~4, PR #33·#34·#45·#51).
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
- ✅ **PR-UX2**: `/account` Kratos privileged session 안내 (DONE).
- ✅ **PR-UX3**: Header Switch View 한계 안내 (DONE).
- ✅ **PR-M2-AUDIT**: Kratos webhook → DevHub `audit_logs` 통합 (DONE).
- ✅ **PR-T4**: GitHub Actions CI 구축 및 35종 E2E 테스트 자동화 (DONE).

세부는 [sprint_plan](../ai-workflow/memory/claude/login_usermanagement_finish/sprint_plan.md) 참조.

#### M2 명시적 out-of-scope (별도 sprint)

- Hydra JWKS / introspection verifier 실구현 → 보안 강화 sprint.
- Sign Up (셀프 가입, 인사 DB 연동) → M3.
- MFA / Two-Factor → M4.

### M3: 사용자 및 조직 관리 (User & Org Management) — In Progress (대부분 M2 1차 완성 흡수)

> **drift 정합 (2026-05-13, sprint `claude/work_260513-k`)**: 본 §3 가 M3/M4 정의의 source-of-truth. 매트릭스 §2.3 + state.json + backend_roadmap §5 모두 본 절 기준으로 정합화.

**M2 1차 완성 sprint (`claude/login_usermanagement_finish`, PR #85) 가 흡수한 항목** (이전에는 M3 으로 분류):
- ✅ 사용자 관리 (유저 CRUD UI 고도화 + 권한 할당 정교화) — `/admin/settings/users` + RBAC PermissionEditor.
- ✅ 조직 관리 1차 완성 (부서 CRUD + drag&drop + 리더 변경 영속화) — work_26_05_11 트랙 S (PR #52~#55).
- ✅ CI/CD GHA (Unit + E2E + actionlint) — PR #86~#88 + ADR-0005.

**M3 잔여**:
- ⏳ **Sign Up (셀프 가입)**: 인사 DB 연동 기반 사용자 셀프 등록 (`POST /api/v1/auth/signup` 의 hrdb lookup arm).
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
- ⏳ **B**: backend 구현 (`domain.DevRequest` / store / handler / migration 000022 dev_requests + 000023 dev_request_intake_tokens + `requireIntakeToken` middleware) + API-59..65 활성화. ADR-0012 머지로 진입 조건 충족.
- ⏳ **F**: 담당자 dashboard 의 "내 대기 의뢰" 위젯 + `/admin/settings/dev-requests` 페이지 + Promote-to-Application/Project 연계.
- ⏳ **A (ADR)**: PMO Manager / 담당자 위양 정책 (ADR-0011 §4.2 패턴) — DREQ-RBAC-ADR. backend 구현과 병행.
- ⏳ **B·F·X**: UT-dreq / TC-DREQ-* 발급 + Playwright spec.
- ⏳ **B (carve)**: 외부 시스템 callback (webhook 송신) — MVP 안정화 후.

문서 hub: [`docs/planning/development_request_concept.md`](./planning/development_request_concept.md), 추적성 [`docs/traceability/report.md §2/§3 DREQ`](./traceability/report.md).

---

## 4. 트랙별 세부 작업 매핑

### 4.1 Backend (B)
| 기능 단위 | 작업 | 마일스톤 | 우선순위 |
| :--- | :--- | :--- | :--- |
| **Auth** | OIDC Token Exchange 및 세션 관리 | M2 | P0 |
| **Realtime** | WebSocket Replay 및 리소스 필터링 | M3 | P1 |
| **Task** | Gitea REST 연동 및 데이터 정규화 | M4 | P3 |
| **Admin** | Gitea Runner 상태 어댑터 구현 | M4 | P3 |

### 4.2 Frontend (F)
| 기능 단위 | 작업 | 마일스톤 | 우선순위 |
| :--- | :--- | :--- | :--- |
| **Auth** | `/auth/callback` 및 API 헤더 토큰 주입 | M2 | P0 |
| **Dashboard** | 실시간 로그 스트리밍 UI 구현 | M3 | P1 |
| **Task** | 과제 추적 대시보드 및 상세 페이지 | M4 | P3 |
| **Admin** | 시스템 관리자 설정 및 Runner 모니터링 UI | M4 | P3 |

---

## 5. 변경 이력
| 일자 | 변경 내용 | 비고 |
| :--- | :--- | :--- |
| 2026-05-08 | 시스템 기능 단위(Functional Units) 중심으로 로드맵 구조 재편 | gemini/redesign 세션 |
·§5.3-5 |
| OS 서비스 wrapper 운영 진입 시점 결정 | M4 | P3 | ADR-0001 §8-7 |

### 4.6 AI (v2)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Python AI gRPC 서버 1차 | v2 | P3 | backend_roadmap §2 Phase 9 |
| AI Gardener suggestion 모델 + Go Core 연동 | v2 | P3 | backend_roadmap §5 P3 |
| Weekly report 생성 worker | v2 | P3 | frontend_integration §3.4 |
| AI 알림 중재 (집중 시간 보호) 모델 | v2 | P3 | requirements §4-3·§5.3-2 |

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
| **Application/Project 도메인 (총괄 + 기간성 운영)** — 시스템 관리자 등록·관리 vs 일반 사용자 조회 분리 | REQ-FR 발급 완료 + 모듈별 Usecase/ERD 분리 카탈로그 완료. 다음: ARCH/API/마이그레이션 설계 진입 | [`planning/project_management_concept.md`](./planning/project_management_concept.md), [`planning/system_usecases.md`](./planning/system_usecases.md), [`planning/system_erd.md`](./planning/system_erd.md) (2026-05-13) |

---

## 6. 충돌 해소 (source-of-truth 정리)

`requirements.md`, `architecture.md`, `backend/requirements.md`, `backend_api_contract.md` 사이에 시점 차이로 인한 표현이 남아 있다. 본 표가 통합 로드맵의 source-of-truth 다.

| 주제 | 폐기된 표현 | 채택된 표현 | 결정 출처 |
| --- | --- | --- | --- |
| 인증/계정 구현 | 자체 `accounts` 테이블, 자체 7 endpoint (`requirements §2.5`, `architecture §6.2`, `backend/requirements §5`, `api_contract §11` historical) | 정책 invariant 만 보존, 구현은 **Hydra + Kratos** (Kratos가 credential master, `users` 테이블은 organizational metadata 만) | ADR-0001 (2026-05-07) |
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
| 2026-05-13 | §5 백로그에 "Application/Project 도메인 (총괄 + 기간성 운영)" 1행 추가. 컨셉 문서 staged 상태로 안내. | sprint `claude/work_260513-p`. |
| 2026-05-13 | Application/Project 요구사항 고도화 반영 — REQ-FR-APP/REQ-FR-PROJ + 모듈별 UC/ERD 카탈로그(`planning/system_usecases.md`, `planning/system_erd.md`) 연결. 다음 단계(ARCH/API) 전환 기준 명시. | current session |

---

## 8. 다음 단계

1. 본 문서를 PR 로 머지하면 `docs/README.md`, `docs/DOCUMENT_INDEX.md`, 트랙별 세부 로드맵 상단에 진입점 링크가 추가된다.
2. M0 sprint 진입 직전, 본 문서 §3.1 의 DoD 항목별로 backlog 항목을 분해한다 (`ai-workflow/memory/claude/<branch>/backlog/<date>.md`).
3. 트랙 별 세부 로드맵은 본 문서의 마일스톤·우선순위 분류를 따르도록 갱신한다 — phase 표가 본 문서의 M0~M4 와 어떻게 매핑되는지 표 1개를 상단에 둔다.
4. `architecture/README.md`, `planning/README.md` 는 본 통합 로드맵 채택 후 산출물로 채운다.
