# DevHub 요구사항 정의서 (Master Index)

- 문서 목적: 팀 통합 개발 허브 (DevHub) 의 역할별 상세 기능 (Developer / Manager / System Admin) 과 공통 운영 원칙·기술 결정·도메인별 요구사항 진입점을 제공한다.
- 범위: §1 개요 / §2 역할별 요구사항 / §3 공통 시스템 요구사항 / §4 공통 운영 원칙 + 데이터 운영 기준 / §5 도메인별 요구사항 link 표 / §6 기술 스택 결정 / §7 상세 시스템 아키텍처 진입.
- 대상 독자: 프로젝트 리드, Backend / 프론트엔드 / DevOps 개발자, AI agent, QA, UX 검토자.
- 상태: accepted
- 작성일: 2026-04-28
- 최종 수정일: 2026-05-29 (Phase 3 split — §5 도메인별 요구사항이 sub-document 로 이관, 본 문서는 master index 로 전환)
- 관련 문서: [통합 개발 로드맵](./development_roadmap.md), [아키텍처 (master index)](./architecture.md), [기술 스택](./tech_stack.md), [백엔드 API 계약 (master index)](./backend_api_contract.md), [추적성 매트릭스](./traceability/report.md), [거버넌스 — 문서 표준](./governance/document-standards.md), [코드베이스 스냅샷 (2026-05-27)](./analysis/2026-05-27-codebase-snapshot/README.md).

## 1. 개요

본 문서는 팀 통합 개발 허브(DevHub)의 역할별 상세 기능과 데이터 구조 + 공통 정책을 정의한다. 도메인별 상세 요구사항은 [`docs/domain/<도메인>/requirements.md`](./domain/) 에 sub-document 로 분산되어 있으며, 본 문서는 그 진입점 + cross-cutting 정책을 유지한다.

### 1.1 요구사항 범위 구분

본 문서의 기능 항목은 다음 상태로 구분합니다.

- **확정:** 현재 제품 방향과 초기 구현 기준에 포함되는 요구사항.
- **후보:** 사용자 가치가 있으나 세부 정책, 우선순위, 구현 범위 추가 검토가 필요한 항목.
- **MVP 이후:** 초기 버전 이후 단계적으로 도입할 확장 요구사항.

역할별 요구사항의 체크박스는 기능 후보를 추적하기 위한 목록이며, `핵심 기획 아젠다`의 결정 사항은 제품 방향과 정책 기준으로 우선 적용합니다. 단, 구현 범위는 각 항목의 `확정`, `후보`, `MVP 이후` 상태에 따라 별도로 관리합니다.

## 2. 사용자 역할별 요구사항 (Two-Dimensional RBAC)

> **2026-06-01 갱신**: 기존 3 system role(developer/manager/system_admin) 모델을 **2차원 RBAC** 로 확장. system role 3종(developer/team_manager/system_admin) + resource role 4종(project_member/project_leader/application_leader/org_head). 상세 컨셉은 [`docs/planning/role-access-concept.md`](./planning/role-access-concept.md), 도메인 REQ 는 [`docs/domain/application-lifecycle/requirements.md`](./domain/application-lifecycle/requirements.md) §6 REQ-FR-ROLE 참조.

### 2.1 개발자 (Developer)
- **핵심 니즈:** 정보 탐색 최소화, 개발 몰입도 향상. 자신이 속한 project/application 에 대한 읽기 접근.
- **기본 진입 우선순위:** 개발 대시보드 (Developer Dashboard)
- **System Role:** `developer`
- **Resource Role baseline:** `project_member` (project_members 포함 시 해당 project + 연결 application 조회)
- **View Scope:** 자신이 member 인 project + 연결 application 으로 row-scoped. management 정보(롤업/메트릭/리스크)는 `project_leader` 이상만 접근.
- **주요 기능 (확정):**
    - [x] 멤버십 기반 Project/Application 목록 조회 (row-scoped).
    - [x] Project/Application 상세 조회 (member 인 경우).
- **주요 기능 (후보):**
    - [ ] 기술 스택별 가이드 및 Wiki 통합 검색.
    - [ ] 프로젝트별 환경 설정(Environment Setup) 원클릭 확인.
    - [ ] 팀 내 공통 라이브러리/컴포넌트 카탈로그.
    - [ ] (기타 제안) CI/CD 빌드 결과 및 실시간 에러 로그 요약.

### 2.2 팀 관리자 (Team Manager)
- **핵심 니즈:** (기존 `manager` + `team_manager` 통합) 팀 범위 프로젝트 가시성 확보, 리스크 선제 대응.
- **기본 진입 우선순위:** 관리 대시보드 (Management Dashboard)
- **System Role:** `team_manager` (신규. 기존 `manager`/`team_manager` → 통합)
- **View Scope:** 자신이 속한 org unit(primary_unit_id 기준 subtree) 범위 내 전체 Project/Application 접근. team scope 밖은 member 인 project 만 접근.
- **주요 기능 (확정):**
    - [x] Team scope 내 Project/Application 목록·상세 조회.
    - [x] Team scope 내 Project 관리 (metadata 수정, member role 변경).
    - [x] Team scope 내 Application 관리 (metadata 수정).
- **주요 기능 (후보):**
    - [ ] 마일스톤별 진행률 시각화 대시보드.
    - [ ] 팀원별 작업량(Load) 및 할당 현황.
    - [ ] 일정 지연 및 차단(Blocked) 과제 알림.
    - [ ] (기타 제안) 투입 공수(Man-month) 및 예산 추정치 비교.

### 2.3 확장 역할: 테스트 담당자 (QA)
- **핵심 니즈:** 품질 지표 관리, 결함 추적 효율화.
- **주요 기능 (후보):**
    - [ ] 버전별 테스트 커버리지 및 패스율 리포트.
    - [ ] 결함(Bug) 수정 현황 및 재테스트 대기 목록.

### 2.4 시스템 관리자 (System Administrator)
- **핵심 니즈:** Gitea 연동 인프라와 DevHub 운영 설정을 안전하게 관리.
- **기본 진입 우선순위:** 시스템 대시보드 + 시스템 설정 메뉴
- **System Role:** `system_admin`
- **View Scope:** Global unrestricted (모든 row 접근, row filter 미적용).
- **노출 정책 (확정):** 시스템 대시보드/시스템 설정은 `system_admin` 권한 보유자에게만 노출한다.
- **주요 기능 (확정):**
    - [x] Gitea 서버 및 Runner 상태 모니터링.
    - [x] Runner 재시작/설정 등 제한된 인프라 제어.
    - [x] Gitea 계정, 조직, 권한 연동 관리.
    - [x] Keycloak 연계 계정/세션 운영 상태 가시성 및 사용자 메타데이터 연동 관리 — [auth-session 도메인](./domain/auth-session/requirements.md) 참조.
    - [x] 백업 상태와 시스템 알림 임계치 확인.
    - [x] 시스템 관리 작업에 대한 Audit Log 조회.

### 2.5 사용자 계정 관리 (User Account Management)

본 절의 상세 정책 (User/Account 분리, Keycloak 단일 IdP, 운영 책임 분리)은 **[auth-session 도메인 requirements](./domain/auth-session/requirements.md)** 로 이관됐다. 본 master 의 §4.1 데이터 운영 기준 표는 cross-cutting 으로 아래에 유지된다.

## 3. 공통 기능 및 시스템 요구사항
- **역할 확장성:** 새로운 역할이 추가되더라도 역할별 기본 진입 우선순위를 설정해 UX를 간접 제공하고, 메뉴/기능은 권한 기반으로 확장할 수 있어야 함.
- **데이터 일관성:** 개발자가 업데이트한 진행 상황이 관리자 대시보드에 즉시 반영되어야 함.

## 4. 공통 운영 원칙 (Common Operating Principles)

세 system role(developer / team_manager / system_admin) 간의 뷰 공존과 데이터 신뢰를 위해 다음 원칙을 준수합니다.

1. **데이터 주권과 정합성의 조화:**
    - 공식 업무 로그(PR, 이슈, 빌드 등)는 데이터 무결성을 유지하며 관리자 KPI의 기반이 됨.
    - 개인적 회고, 업무 설명, 개인 로드맵 영역에서는 개발자에게 완전한 수정/삭제 권한을 부여함.
    - 보고용 데이터는 특정 시점의 **스냅샷(Snapshot)**을 기반으로 운영하여 데이터 변경으로 인한 혼선을 방지함.
2. **기술 태깅 기반의 전문가 맵:**
    - '스텔스 전문가'라는 용어 대신 **'기술 태깅(Tech Tagging)'** 개념을 사용함.
    - 시스템이 인식한 전문성은 개발자에게 긍정적인 피드백(Kudos)과 함께 투명하게 공유하여 자발적 정보 제공 동기를 부여함.
3. **AI 가드너 기반의 알림 중재 (v2 예정):**
    - 모든 강제 알림은 **AI 가드너**가 사용자의 현재 업무 몰입 상태를 판단하여 전달 시점을 중재함.
    - 관리자의 긴급 요청이라도 개발자의 집중 시간을 보호하는 것을 원칙으로 함.
4. **팀 성취 중심의 서비스 톤 정합:**
    - 게이미피케이션(Fun) 요소와 공식 거버넌스(Professional) 로그를 **'팀 마일스톤/업적'**으로 통합함.
    - 관리자의 의사결정이 개발자 뷰에서는 '팀 퀘스트 완료'나 '팀 성취'로 시각화되도록 설계함.

### 4.1 데이터 및 권한 운영 기준

데이터 주권, 조회 권한, 보존 정책, 알림 정책은 다음 기준을 우선 적용합니다.

| 데이터 분류 | 원천 | 수정 권한 | 조회 권한 | 보존 기준 | 알림 기준 |
| --- | --- | --- | --- | --- | --- |
| 공식 업무 로그 | Gitea | Gitea 원천 기준, DevHub 직접 수정 불가 | 역할/프로젝트 권한 기준 | 운영 로그 1개월 | 상태 변경/지연/실패 시 역할별 알림 |
| 개인 업무 연혁 | DevHub | 본인 수정/삭제 가능 | 본인 기본, 공유 선택 시 승인된 대상 | 계정 활성 중 유지, 삭제 후 1개월 | 사용자가 선택한 경우만 공유 제안 |
| 보고 스냅샷 | DevHub | 생성 후 직접 수정 불가, 새 스냅샷으로 대체 | 관리자/승인된 리더 | 정책 기간 별도 정의 필요 | 보고 생성/변경 시 관리자 알림 |
| 기술 태깅/Kudos | DevHub | 시스템 추천 + 사용자 확인/관리자 승인 | PL/GL 이상 등 정책 기준 | 계정 활성 중 유지, 삭제 후 1개월 | 긍정 피드백 중심, 강제 알림 최소화 |
| 시스템 관리 로그 | DevHub/Gitea | 시스템 자동 기록, 직접 수정 불가 | 시스템 관리자 | 운영 로그 1개월 이상 검토 필요 | 보안/운영 이벤트는 즉시 알림 |
| 사용자 계정/자격 | Keycloak + DevHub | 본인은 자기 비밀번호 변경, 계정 발급/회수/강제 재설정은 IdP 운영 절차, DevHub는 사용자 메타데이터/감사 추적 관리 | 본인(자기 자신), 시스템 관리자(전체) | 계정 활성 중 유지, 회수 후 90일 보존 후 삭제 | 비밀번호 변경/잠금/회수 시 본인 + 시스템 관리자 알림 |

## 5. 도메인별 요구사항 (sub-document link 표)

본 절은 도메인별 sub-document 의 진입점이다. ID 본문 (REQ-FR-*/REQ-NFR-*) 은 각 sub-document 가 source-of-truth.

| 도메인 | 요구사항 | 비고 |
| --- | --- | --- |
| auth-session | [`./domain/auth-session/requirements.md`](./domain/auth-session/requirements.md) | User ↔ Account 분리, Keycloak 단일 IdP, historical 비밀번호 정책 |
| audit-ops | [`./domain/audit-ops/requirements.md`](./domain/audit-ops/requirements.md) | Audit log emit, Keycloak event sync, Prometheus metric |
| rbac-permissions | [`./domain/rbac-permissions/requirements.md`](./domain/rbac-permissions/requirements.md) | Role + Resource + Action matrix, row-scoping |
| organization-management | [`./domain/organization-management/requirements.md`](./domain/organization-management/requirements.md) | Users + org_units + appointments, HRDB lookup |
| onboarding | [`./domain/onboarding/requirements.md`](./domain/onboarding/requirements.md) | REQ-FR-ONBOARD-001..012, REQ-NFR-ONBOARD-001..008 |
| application-lifecycle | [`./domain/application-lifecycle/requirements.md`](./domain/application-lifecycle/requirements.md) | REQ-FR-APP-001..012, REQ-FR-PROJ-000..010, REQ-FR-APPDASH-001..006, REQ-FR-ROLE-001..016, REQ-NFR-PROJ/APPDASH |
| repository-integration | [`./domain/repository-integration/requirements.md`](./domain/repository-integration/requirements.md) | REQ-FR-REPO-001..005, REQ-NFR-REPO-001..003 |
| dev-request | [`./domain/dev-request/requirements.md`](./domain/dev-request/requirements.md) | REQ-FR-DREQ-001..013, REQ-NFR-DREQ-001..006 |
| integration-registry | [`./domain/integration-registry/requirements.md`](./domain/integration-registry/requirements.md) + [`task_requirements.md`](./domain/integration-registry/task_requirements.md) | REQ-FR-INT-001..015, REQ-NFR-INT-001..009, REQ-FR-TASK-001..010, REQ-NFR-TASK-001..004 |
| realtime | [`./domain/realtime/requirements.md`](./domain/realtime/requirements.md) | WebSocket ticket 인증, event RBAC 재검사 |

> 추가 source — 기존 file 들은 유지 (역사 보존 + detailed body):
> - `docs/backend/requirements.md` — backend-specific 상세 REQ
> - `docs/backend_requirements_org_hierarchy.md` — organization 도메인 detail (organization-management 도메인 README 가 진입)
> - `docs/frontend_integration_requirements.md` — frontend 연동 REQ

## 6. 기술 스택 결정 사항 (Technology Stack Decisions)

기술 스택 상세 계약, 버전, 설치/검증 명령은 **[tech_stack.md](./tech_stack.md)**를 기준으로 관리합니다. 본 요구사항 문서에서는 제품 요구사항과 직접 연결되는 기술 결정 요약만 유지합니다.

- **하이브리드 백엔드:** Gitea 연동, Webhook 수집, 시스템 제어, 권한 관리는 Go Core가 담당합니다. AI 가드너와 분석성 작업(Python AI)은 v2에서 도입합니다.
- **내부 통신:** Go Core와 Python AI의 분석 요청/응답은 gRPC를 기본 계약으로 사용합니다.
- **프론트엔드:** 역할별 진입 우선순위와 실시간 상태 시각화는 Next.js 기반 UI에서 제공합니다.
- **데이터베이스:** Gitea 원본 이벤트, 프로젝트/저장소/사용자/권한 관계, 비정형 분석 결과 저장에는 PostgreSQL을 사용합니다.

## 7. 상세 시스템 아키텍처 설계 (Detailed System Architecture)

상세한 시스템 아키텍처 설계 내용은 별도 문서인 **[architecture.md](./architecture.md)** (master index) 에서 관리하며, 구체적인 기술 스택 및 환경 설정 가이드는 **[tech_stack.md](./tech_stack.md)**를 참조합니다.

### 주요 아키텍처 결정 사항:
- **내부 통신:** Go Core ↔ Python AI 간 gRPC 도입.
- **실시간성:** WebSocket을 통한 프론트엔드 실시간 상태 전송 (ticket 패턴, ADR-0024). SSE는 초기 구현 범위에서 제외하고 운영 환경 제약 발생 시 fallback으로 재검토.
- **시각화:** React Flow를 이용한 인터랙티브 인프라 구성도.

---

## 변경 이력 (요약)

| 일자 | 변경 |
| --- | --- |
| 2026-06-01 | **Two-Dimensional RBAC 도입** — §2 사용자 역할 구조를 system role 3종(developer/team_manager/system_admin) + resource role 4종(project_member/project_leader/application_leader/org_head) 으로 확장. `manager`/`team_manager` → `team_manager` 통합. §2.1 developer matrix 확장(applications/projects view ON, row-scoped). §2.2 manager → team_manager 갱신. §2.4 system_admin view scope 명시화. §5 app-lifecycle link 표에 REQ-FR-ROLE-001..016 추가. 컨셉 문서: `docs/planning/role-access-concept.md`. |
| 2026-05-29 | **Phase 3 split** — 도메인별 본문(§5.4~§5.10)을 10 도메인 sub-document 의 `requirements.md` 로 이관. 본 master 는 §1-4 (cross-cutting) + §5 link 표 + §6-7 (cross-cutting) 로 축소. ID 보존(REQ-FR-APP-001..012, REQ-FR-PROJ-000..010, REQ-FR-DREQ-001..013, REQ-FR-INT-001..015, REQ-FR-ONBOARD-001..012, REQ-FR-REPO-001..005, REQ-FR-APPDASH-001..006, REQ-FR-TASK-001..010, REQ-NFR-* 전체), 신규 발급/삭제 없음. §2.5 사용자 계정 관리 본문은 auth-session 도메인으로 이관. |
| 2026-05-28 | (split 이전) §5.10 (Task Item Ingestion) 신규 — integration-registry/task_requirements.md 로 이관됨. |
| 2026-05-27 | (split 이전) §5.8 (SCM↔시스템 Repository) 신규 — repository-integration/requirements.md 로 이관됨. §5.6 INT 보강 — integration-registry/requirements.md 로 이관. |
