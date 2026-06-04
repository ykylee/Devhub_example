# platform-lifecycle 도메인 요구사항

- 문서 목적: Platform + Project + Repository 계층 운영 모델, 상태 머신, 롤업 요구사항 + Platform 개발 대시보드(APPDASH) 기능 요구사항을 정의한다.
- 범위: REQ-FR-PROJ-*, REQ-FR-APP-*, REQ-NFR-PROJ-*, REQ-FR-APPDASH-*, REQ-NFR-APPDASH-*, REQ-FR-ROLE-*. SCM↔시스템 repository 연동은 `docs/domain/repository-integration/requirements.md` 참조. DREQ promote 흐름은 `docs/domain/dev-request/requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-06-01 (Two-Dimensional RBAC 도입 — §2 PROJ-000/APP-010/PROJ-009~010 갱신, §6 REQ-FR-ROLE-001..016 신규)
- 관련 문서: [도메인 README](./README.md), [project_concept](./project_concept.md), [dashboard_concept](./dashboard_concept.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0014](../../adr/0014-application-project-lifecycle.md)

## 1. 개요

본 도메인은 두 묶음으로 구성된다.

1. **Platform + Project 운영 모델 + Repository 계층**(§2 / §3, 기존 master `docs/requirements.md` §5.4) — DevHub 의 최상위 계층(Platform > Repository > Project > GitHub 실행 단위) + 상태 머신 + 롤업.
2. **Platform 개발 대시보드(APPDASH)**(§4 / §5, 기존 master `docs/requirements.md` §5.9) — DevHub 핵심 단위인 Platform 상세 대시보드의 기능 요구사항.

## 2. Project 운영 모델 + Platform/Repository 계층 (REQ-FR-PROJ / REQ-FR-APP)

- **논의 배경:** 제품 수명 단위의 상위 관리와 연간 운영 단위의 행정 갱신을 함께 지원해야 한다. 이를 위해 `Platform > Repository > Project > GitHub 실행 단위` 계층을 요구사항으로 확정한다.
- **용어 정책 (확정):**
    - **Platform:** 제품 수명 주기와 함께 가는 최상위 총괄 단위.
    - **Repository:** Platform 하위 실행 단위 (1 Platform : N Repository).
    - **Project:** Repository 하위 기간성 운영 단위. Project는 1개 이상의 Repository와 연결될 수 있으며(`Project ↔ Repository` N:M), 연결된 Repository 중 1개를 primary로 지정한다.
    - **Execution Artifact:** Repository 내부의 Issue / Milestone / Project Board / Wiki(또는 Docs).

### 2.1 기능 요구사항 (REQ-FR)

- **REQ-FR-PROJ-000 (MVP, 확정):** `Platform > Repository > Project` 관리 쓰기 권한은 기본적으로 `system_admin`에 한정해야 한다.
    - 대상 기능: Platform 생성/수정/보관, Repository 연결/해제, Project 생성/수정/보관, Project 멤버/owner 관리, Integration 정책 변경, 마일스톤 매핑 관리.
    - 예외 역할: `team_manager`는 team scope 내에서 Project 관리 권한을 가진다 (`team_manager` role `ResourceProjects.{Create, Edit, Delete}` = true). 단, Platform 생성/수정/보관은 `system_admin` 전유.
    - READ scope(목록/상세 조회)는 `developer` 이상 모든 role 에게 row-scoped 로 허용 (matrix `ResourceApplications.View = true`, `ResourceProjects.View = true`).
- **REQ-FR-APP-001 (MVP, 확정):** 시스템 관리자는 Platform을 생성/수정/보관(archive)할 수 있어야 한다.
    - 필수 필드: `key`, `name`, `owner`, `start_date`, `due_date`, `visibility`, `status`.
    - `status` 최소 상태: `planning`, `active`, `on_hold`, `closed`, `archived`.
- **REQ-FR-APP-002 (MVP, 확정):** 하나의 Application은 0개 이상의 Repository를 연결할 수 있어야 한다.
    - 연결 단위 필드: `repo_provider`, `repo_full_name`, `role(primary|sub|shared)`.
    - 동일 Platform 내 서로 다른 provider Repository를 동시에 연결할 수 있어야 한다 (예: bitbucket + gitea 병행).
- **REQ-FR-APP-003 (MVP, 확정):** `Platform.key`는 시스템 전역 고유값(unique)이어야 하며 관리 식별자로 사용해야 한다.
    - 표시명(`name`) 변경과 무관하게 `key`는 안정 식별자로 유지한다.
    - **변경 불가(immutable):** 발급 후 `key`는 변경할 수 없다. rename 이 필요하면 신규 Application 생성 + 기존 Platform archive 절차로만 수행한다 (PATCH 응답 422 `application_key_immutable`).
    - 현재 입력 정책: 영문숫자 조합 10자 (`^[A-Za-z0-9]{10}$`).
    - 데이터베이스 컬럼은 정책 변경 여지를 위해 더 긴 길이(예: VARCHAR(32) 이상)를 허용하고, 실제 길이 제한은 애플리케이션 검증 정책으로 강제한다.
- **REQ-FR-APP-004 (MVP, 확정):** Repository는 외부 형상관리 도구와 연결되는 구조여야 하며, DevHub는 운영/분석용 관리 데이터를 보유해야 한다.
    - 외부 SoT: 코드/PR/빌드 원본.
    - DevHub 보유: 연결 메타데이터, 동기화 상태, 운영 스냅샷.
    - 지원 정책: 특정 SCM 단일 종속이 아니라 provider 추상화(`repo_provider`)를 사용하며, `bitbucket`, `gitea`, `forgejo` 등 복수 provider를 동등하게 지원/확장할 수 있어야 한다.
- **REQ-FR-APP-005 (MVP, 확정):** Repository 작업현황을 수집/조회할 수 있어야 한다.
    - 최소 지표: commit 활동량, active contributor 수, 작업 추이.
- **REQ-FR-APP-006 (MVP, 확정):** PR/PR Activity 정보를 수집/조회할 수 있어야 한다.
    - 최소 정보: PR 상태(open/draft/merged/closed), 생성/리뷰/코멘트/머지 이벤트 타임라인.
- **REQ-FR-APP-007 (MVP, 확정):** 빌드 정보를 수집/조회할 수 있어야 한다.
    - 최소 정보: run status, duration, branch/commit, 시작/종료 시각.
- **REQ-FR-APP-008 (MVP, 확정):** 소스코드 품질 지표(정적분석/스코어링)를 수집/조회할 수 있어야 한다.
    - 최소 정보: tool, quality score, gate pass/fail, metric 상세(coverage, bug/vuln, duplication 등).
- **REQ-FR-APP-009 (MVP, 확정):** 형상관리 도구 연동은 provider별 어댑터 구조를 따라야 한다.
    - 공통 도메인 계약(Repository/PR/Build/Quality 이벤트/스냅샷)과 provider 전용 구현을 분리한다.
    - 신규 provider 추가 시 기존 도메인 API/화면 계약을 깨지 않고 어댑터 추가만으로 확장 가능해야 한다.
    - provider별 인증/웹훅 검증/속도제한/에러 포맷 차이는 어댑터 내부에서 흡수한다.
- **REQ-FR-APP-010 (MVP, 확정):** Platform 상태 전이는 정의된 상태 머신 규칙을 따라야 한다.
    - 상태 집합: `planning`, `active`, `on_hold`, `closed`, `archived`.
    - `archived`는 기본적으로 종료 상태이며 일반 상태 전이로 복구하지 않는다.
    - 상태 전이 권한: 기본적으로 `system_admin`만 허용한다. `team_manager`는 team scope 내 Platform 에 대해 `active → on_hold` / `on_hold → active` 전이만 허용한다. `developer` 이하는 상태 전이 불가.
    - 전이 검증 가드:
      - `planning -> active`: 연결된 활성 Repository 1개 이상 필요.
      - `active -> closed`: `severity=critical` 롤업 경고 0건 + 연결 Repository 1개 이상 필요.
      - `on_hold -> active`: `due_date` 만료 시 재개 사유(`resume_reason`) 기록 필요.
      - `* -> archived`: soft-delete 처리, `archived_reason` 기록 필요.
- **REQ-FR-APP-011 (MVP, 확정):** Platform-Repository 연결은 라이프사이클 상태를 가져야 한다.
    - 최소 상태: `requested`, `verifying`, `active`, `degraded`, `disconnected`.
    - 연결 검증 실패/일시 장애 시 `sync_error_code`를 기록해야 한다.
    - `sync_error_code`는 표준 코드 사전을 사용해야 하며(`provider_unreachable`, `auth_invalid`, `permission_denied`, `rate_limited`, `webhook_signature_invalid`, `payload_schema_mismatch`, `resource_not_found`, `internal_adapter_error`), 임의 문자열 사용을 금지한다.
    - `sync_error_code`에는 재시도 가능 여부(`retryable`)와 최근 발생 시각이 함께 관리되어야 한다.
- **REQ-FR-APP-012 (MVP, 확정):** Platform 롤업은 누락/장애 데이터를 숨기지 않고 `data_gap` 또는 경고 상태로 표시해야 한다.
    - 최소 롤업 대상: PR 분포, 빌드 성공률/평균 시간, 품질 점수, gate 실패 건수.
    - 기본 `weight_policy`는 `equal`(동일 가중)이다.
    - 선택 `weight_policy`는 `repo_role`(primary/sub/shared 가중), `custom`(관리자 정의)를 지원할 수 있어야 한다.
    - `custom` 정책은 가중치 합이 1.0(±허용오차)이어야 하며, 음수 가중치는 허용하지 않는다.
- **REQ-FR-PROJ-001 (MVP, 확정):** 시스템 관리자는 Platform 범위에서 Project를 생성/수정/보관(archive)할 수 있어야 한다.
    - 필수 필드: `key`, `name`, `owner`, `start_date`, `due_date`, `visibility`, `status`.
    - `status` 최소 상태: `planning`, `active`, `on_hold`, `closed`, `archived`.
    - 생성 시 `repository_ids` (1개 이상)와 `primary_repository_id`를 함께 지정해야 하며, `primary_repository_id`는 `repository_ids`에 포함되어야 한다.
- **REQ-FR-PROJ-002 (MVP, 확정):** 일반 사용자는 자신이 멤버인 Project 및 공개 Project를 조회할 수 있어야 한다.
    - `archived` Project는 기본 숨김이며, 명시적 토글로 노출한다.
- **REQ-FR-PROJ-003 (MVP, 확정):** 시스템 관리자는 Project별 멤버/책임자(owner)를 관리할 수 있어야 한다.
- **REQ-FR-PROJ-004 (MVP, 확정):** 상위(Platform) 로드맵/마일스톤과 하위(Repository) 로드맵/마일스톤을 연결(매핑)할 수 있어야 한다.
    - 모든 하위 마일스톤은 상위 마일스톤에 `child -> parent` 매핑 가능해야 한다.
- **REQ-FR-PROJ-005 (MVP, 확정):** Jira 연동은 하이브리드 정책을 지원해야 한다.
    - 실행 이슈 Source of Truth는 Repository Jira.
    - Project는 Repository 하위 기간성 운영 단위로 관리하며, 다중 Repository 연결 시에도 실행 이슈 Source of Truth는 연결된 Repository들의 Jira 정책을 따른다.
    - Project Jira에 작업성 Story/Task 직접 생성은 정책 위반으로 취급.
- **REQ-FR-PROJ-006 (MVP, 확정):** Confluence(또는 문서 체계)는 상/하위 분리 정책을 지원해야 한다.
    - Project 문서: 방향성/의사결정/분기 계획.
    - Repository 문서: 설계/RFC/runbook/회고.
- **REQ-FR-PROJ-007 (MVP, 확정):** 스프린트는 Repository 단위로 운영되어야 하며, Platform 레벨은 주간/월간 cadence로 상태를 롤업해야 한다.
    - 권장 cadence: 주간 Program Sync, 월간 KPI/리스크 리뷰.
- **REQ-FR-PROJ-008 (후속):** Project 영구 삭제는 `archive 후 N일 보존 + 관리자 재확인` 정책을 따라야 한다.
- **REQ-FR-PROJ-009 (활성화, 2026-05-15 sprint `claude/work_260515-c`):** Owner 위양(RBAC row-level)은 ADR-0011 §4.2 의 `enforceRowOwnership(c, ownerUserID, allowedRoles...)` helper 로 활성화한다. allow 규칙: (1) `system_admin`, (2) `allowedRoles` 화이트리스트, (3) `actor.login == ownerUserID`. deny 시 `auth.row_denied` audit + 403 + `code=auth_row_denied`. handler 단위 호출은 별도 sprint (team_manager seed 결정 후).
- **REQ-FR-PROJ-010 (후속):** `team_manager` 역할의 team scope 내 Project 관리 권한 범위를 세분화한다.
    - 기본 허용 범위: `project.manage`(metadata 수정), `project.member.manage`(member role 변경), `milestone.mapping.manage`.
    - 범위 제한: team scope(primary_unit_id 기준 subtree) 밖의 Project 에는 `developer` 와 동일한 row-scoped member 접근만 허용.
    - 금지 범위: 시스템 설정, 계정/조직/RBAC 정책 변경.

### 2.2 비기능/운영 요구사항 (REQ-NFR)

- **REQ-NFR-PROJ-001 (MVP):** Project/Repository 매핑 정보는 감사(audit) 가능해야 하며 생성/수정/해제 이력을 기록해야 한다.
- **REQ-NFR-PROJ-002 (MVP):** 상위 롤업 지표는 매핑 누락 항목을 조용히 제외하지 않고 경고 상태로 표시해야 한다.
- **REQ-NFR-PROJ-003 (후속):** Project 대시보드 응답시간 목표(예: p95 2초 이내)와 페이지네이션 한계는 설계 단계에서 별도 계약한다.
- **REQ-NFR-PROJ-004 (MVP):** 외부 형상관리/CI/품질 도구 연동 데이터는 idempotency key 기반 중복 방지 및 재동기화(reconciliation) 정책을 가져야 한다.
- **REQ-NFR-PROJ-005 (MVP):** 어댑터 장애는 provider 단위로 격리되어야 하며, 특정 provider 장애가 전체 수집 파이프라인 중단으로 전파되지 않아야 한다.
- **REQ-NFR-PROJ-006 (MVP):** Platform 롤업 계산은 동일 요청 조건에서 재현 가능해야 하며, 집계 기준(기간/필터/가중치)을 메타데이터로 함께 제공해야 한다.
    - `weight_policy`와 실사용 가중치 맵(`applied_weights`)을 응답 메타에 포함해야 한다.
    - 가중치 누락 repository는 기본값 fallback(`equal`) 적용 여부를 메타에 명시해야 한다.

### 2.3 Usecase 산출물 (확정)

- 본 도메인의 설계 진입 직전 Usecase 산출물은 [`docs/planning/system_usecases.md`](../../planning/system_usecases.md) 를 source-of-truth 로 사용한다.
- 해당 문서의 `UC-*` 는 REQ와 ARCH/API 사이의 중간 추적 단계다.

### 2.4 ERD 산출물 (확정)

- 데이터 모델 기준 문서는 [`docs/planning/system_erd.md`](../../planning/system_erd.md) 를 사용한다.
- 설계/구현 단계의 신규 엔티티·관계는 ERD 문서와 동기화해야 한다.

### 2.5 범위 경계 (Out of Scope)

- 신규 Project 생성 시 Gitea 저장소 자동 생성/브랜치 보호/멤버 초대 자동화는 별도 sprint에서 진행한다.
- WebSocket 기반 실시간 위험 탐지는 M4 범위에서 다루고, AI 제안 자동화는 v2 범위에서 다룬다.
- MFA 기반 위험 작업 다단계 확인은 운영 진입 직전 정책으로 별도 확정한다.

## 3. Platform 개발 대시보드 (REQ-FR-APPDASH / REQ-NFR-APPDASH)

본 절은 컨셉 문서([`./dashboard_concept.md`](./dashboard_concept.md))에 정의된 DevHub 핵심 단위인 Platform 상세 대시보드의 기능 요구사항을 정의한다.

### 3.1 기능 요구사항 (REQ-FR-APPDASH)

- **REQ-FR-APPDASH-001 (MVP, 확정):** 실시간 타겟 브랜치 빌드 상태(Target Branch Build Status)를 최상단 메트릭 카드를 통해 노출해야 한다.
    - **실시간 실패 빌드 런 표시**: 단순 빌드 성공률(%)보다 실시간 broken/red 상태 빌드 현황을 즉시 표기한다.
    - **리포지토리 슬러그 연계**: 연결된 어떤 리포지토리의 어떤 브랜치에서 실패했는지 `org/repo-slug` 형식으로 표시해야 한다.
    - **빌드 실패 진단 정보 및 로그 연동**: 실패 건에 대해 빌드 번호, 실패 경과 시간, 에러 요약 스니펫을 노출하고 해당 빌드 로그로 즉시 이동하는 **[로그 진입 딥링크]** 액션을 제공해야 한다. (모두 정상일 시 `Healthy 🟢` 표시)
- **REQ-FR-APPDASH-002 (MVP, 확정):** 다차원 코드 품질 지표 및 정적 분석 이슈(Quality & Issues) 카드를 노출해야 한다.
    - **5점 만점 normalized 품질 스코어**: 리포지토리별 SonarQube 품질 데이터를 5.0 만점 스케일로 정규화 및 가중 평균하여 노출한다.
    - **심각도별 미해결 정적분석 이슈 노출**: Blocker, Critical, Major 등 심각도 등급에 따라 미해결된 정적분석 이슈 건수를 집계하여 표시해야 한다.
    - **코딩 룰 검사의 역할 분리**: 세부 코딩 룰 및 가독성 린트 지표 등은 상위 대시보드에서 배제하고, 개별 리포지토리 상세 대시보드로 역할을 격리 분리해야 한다.
- **REQ-FR-APPDASH-003 (MVP, 확정):** 하위 프로젝트 진척도 및 로드맵 관리(Linked Projects Progress & Roadmap) 섹션을 대시보드 최상단 주요 영역에 배치 노출해야 한다.
    - **진척 산정 공식**: 단순 완료 태스크 개수 비례 방식과 **스토리 포인트(SP) 가중치 비례 방식**을 선택하여 실질적 진척율(%)을 계산/표시해야 한다.
    - **지능형 리스크 감지 배지**: 남은 작업 대비 잔여 기간 비율을 산출하여 지연 위험도를 계산하고 D-Day와 함께 위험 알림 라벨(`Healthy 🟢`, `Warning 🟡`, `At Risk 🔴`)을 자동 제공해야 한다.
- **REQ-FR-APPDASH-004 (MVP, 확정):** 연결된 모든 개발 의뢰 관리(All Linked Dev Requests - DREQ Overview) 및 프로젝트 승격 워크플로우를 제공해야 한다.
    - **DREQ 조회 및 필터**: 어플리케이션에 매핑된 모든 개발 의뢰 리스트와 상태(대기 중, 검토 중, 승격 완료 등)를 전용 탭에서 필터링 조회 가능해야 한다.
    - **원클릭 프로젝트 승격 연계**: 대기 중인 DREQ 우측의 **[프로젝트 승격 🚀]** 버튼 클릭 시, DREQ의 메타데이터(Key, Name, Description)를 자동 상속/프리필하는 프로젝트 생성 모달을 팝업하고 단일 트랜잭션으로 연계 생성해야 한다.
- **REQ-FR-APPDASH-005 (MVP, 확정):** SCM 및 CI/CD 빌드 안정성 시계열 트렌드 차트(Area Chart)를 제공해야 한다.
    - 7일 및 30일 간의 평균 빌드 소요 시간 변화 추이와 빌드 성공률 추이를 제공해야 한다.
- **REQ-FR-APPDASH-006 (MVP, 확정):** 가중치 배분 비주얼라이저(Weight Policy Visualizer)를 통해 리포지토리 역할(`primary`/`sub`/`shared`)에 따라 계산 롤업에 적용된 가중치를 도넛 차트로 노출하고 가중치 수정 설정을 제공해야 한다.

### 3.2 비기능 / 운영 요구사항 (REQ-NFR-APPDASH)

- **REQ-NFR-APPDASH-001 (MVP):** UI 레이아웃은 모던 글래스모피즘(Glassmorphism) 스타일을 적용하고, 라이트/다크 모드에 최적화된 curated HSL 색상 팔레트 시스템을 사용해 시인성을 극대화해야 한다.
- **REQ-NFR-APPDASH-002 (MVP):** 대시보드 로딩 성능 보장을 위해 캐싱 및 비동기 병렬 aggregation을 적용하여 첫 진입 시 p95 로딩 속도 1.5초 이하를 달성해야 한다.
- **REQ-NFR-APPDASH-003 (MVP):** 연동 장애로 인한 우아한 성능 저하(Graceful Degradation)를 보장해야 한다. 특정 저장소/CI 연동 장애 시 전체 화면이 깨지지 않고, 장애 대상 리포지토리에 대해 `data_gap` 또는 경고 표시를 하며 가용한 데이터만 롤업 집계해 표시해야 한다.

### 3.3 범위 경계 (Out of Scope)

- 실시간 리포지토리 빌드 실패 시 외부 메신저 알림(Slack 등) 자동 전송 기능 (v2 범위).
- AI 기반 빌드 실패 원인 자동 분석 및 코드 패치 제안 (v2 범위).
- 다차원 코드 품질 스코어 산식의 동적 튜닝 UI (어플리케이션 설정 모달에서 weight matrix 직접 입력 기능은 1차 제외).

## 4. Project 상세 대시보드 (REQ-FR-PROJDASH / REQ-NFR-PROJDASH)

본 절은 컨셉 문서([`./project_dashboard_concept.md`](./project_dashboard_concept.md))에 정의된 프로젝트 상세 대시보드의 기능 요구사항을 정의한다.

### 4.1 기능 요구사항 (REQ-FR-PROJDASH)

- **REQ-FR-PROJDASH-001 (MVP, 확정):** 대시보드는 3대 페르소나(개발자, PL, 조직 관리자) 관점의 스위처(3-Way Persona Switcher)를 제공하고 역할 기반으로 렌더링해야 한다.
    - **역할 기반 매핑**: 유저의 Keycloak Resource Access Role에 따라 기본 뷰를 노출한다 (`contributor` -> 개발자 뷰, `project_leader` -> PL 뷰, `pmo_manager/team_manager` -> 관리자 뷰).
    - **다이내믹 뷰 토글**: 상단 세그먼트 스위치를 통해 모드를 전환할 수 있어야 하며, 권한 외 뷰 접근 시 접근 제한 경고를 표시해야 한다.
- **REQ-FR-PROJDASH-002 (MVP, 확정):** 개인화된 실무 피드(Developer My Work)를 최상단에 노출해야 한다.
    - **할당된 이슈 리스트**: 로그인한 개발자 본인에게 할당된 활성 일감(Active Tasks) 목록을 우선순위별로 노출한다.
    - **리뷰 대기 피드**: 본인이 리뷰어로 지정된 PR 목록을 연계 표시한다.
- **REQ-FR-PROJDASH-003 (MVP, 확정):** 프로젝트 리더 관점의 PR 통합 병목 해소 허브(PL Integration Hub)를 제공해야 한다.
    - **통합 블로커 감지**: 빌드가 실패했거나, 충돌이 났거나, 48시간 이상 방치된(Stale) PR 목록을 하이라이트한다.
    - **협업 촉구 퀵액션**: 각 병목 PR 우측에 담당자 호출 및 리마인드 퀵액션을 제공해야 한다.
- **REQ-FR-PROJDASH-004 (MVP, 확정):** 팀 업무 부하 미터(Team Workload Meter)를 관리자 뷰에 노출해야 한다.
    - **리소스 할당 시각화**: 팀원별 할당된 활성 이슈 수와 PR 건수를 게이지 바로 시각화한다.
    - **과부하 알림 배지**: 활성 작업량이 5개 초과인 멤버에게 "Overloaded" 경고 배지를 적용한다.
- **REQ-FR-PROJDASH-005 (MVP, 확정):** 코드 헬스 및 저장소 CI 빌드 상태 롤업 가시성을 제공해야 한다.
    - **저장소별 빌드 헬스**: 연결된 리포지토리별 최신 빌드 성공/실패 여부를 노출한다.
    - **정적 분석 스코어 및 이슈**: SonarQube 품질 스코어 및 미해결 정적분석 이슈(Blocker/Critical) 현황을 보여준다.
- **REQ-FR-PROJDASH-006 (MVP, 확정):** 속도(Velocity) 기반 마일스톤 번다운 및 리스크 예측 지표를 노출해야 한다.
    - **달성률(%) 및 D-Day**: 마일스톤별 남은 기간(D-Day) 및 완료 수준 추적.
    - **지연 리스크 예측**: 팀의 최근 개발 속도 대비 남은 이슈 양을 계산하여 지연 리스크 레벨(`Healthy`, `Warning`, `At Risk`)을 자동 노출한다.

### 4.2 비기능 / 운영 요구사항 (REQ-NFR-PROJDASH)

- **REQ-NFR-PROJDASH-001 (MVP):** 특정 SCM이나 빌드 API 연동 장애 시 화면이 완전히 깨지지 않고, 해당 저장소 영역에 우아한 에러 폴백 UI(data gap 가이드)를 노출해야 한다.
- **REQ-NFR-PROJDASH-002 (MVP):** UI 레이아웃은 글래스모피즘(Glassmorphism) 테마를 따르고, 빌드 실패/Blocker 항목 등 긴급 경고가 필요한 요소에 HSL 기반 Red Neon 펄스 애니메이션을 적용해야 한다.

## 5. 역할 기반 접근 권한 (REQ-FR-ROLE)

> **Two-Dimensional RBAC**: 3개 system role (developer / team_manager / system_admin) × 4개 resource role (project_member / project_leader / application_leader / org_head).  
> 상세 컨셉은 [`docs/planning/role-access-concept.md`](../../planning/role-access-concept.md) 참조.

### 5.1 멤버십 기반 접근 (Baseline)

- **REQ-FR-ROLE-001 (MVP, 확정):** Developer role 사용자는 자신이 `project_members` 에 포함된 project 만 `ListProjects` 조회할 수 있어야 한다.
    - 대상: `GET /api/v1/projects`
    - 동작: actor.user_id 가 project_members 에 포함된 project 목록만 반환. member 가 아닌 project 는 응답에서 제외.
    - 실패 조건: 없음 (빈 목록 허용).
- **REQ-FR-ROLE-002 (MVP, 확정):** Developer role 사용자가 member 가 아닌 project 의 `GetProject` 상세 조회 시 403 을 반환해야 한다.
    - 응답: `status: 403`, `code: "auth_row_denied"`, `denied_reason: "not_project_member"`.
- **REQ-FR-ROLE-003 (MVP, 확정):** Developer role 사용자에게 `ListPlatforms` 는 자신이 member 인 project 의 부모 application 만 조회되어야 한다.
    - 대상: `GET /api/v1/platforms`
    - 동작: `WHERE id IN (SELECT platform_id FROM projects WHERE id IN (member_project_ids))`.
- **REQ-FR-ROLE-015 (MVP, 확정):** `project_members` 에 없는 사용자(nobody)의 `ListProjects` 호출은 빈 목록(`data: []`)을 반환해야 한다.
    - 대상: `GET /api/v1/projects` (developer role, project_members 0건).
- **REQ-FR-ROLE-016 (MVP, 확정):** `project_members` 에 없는 사용자(nobody)의 project 상세 조회는 403 을 반환해야 한다.
    - 대상: `GET /api/v1/projects/{id}` (developer role, member 아님).

### 5.2 Project Leader 관리 정보 접근

- **REQ-FR-ROLE-004 (P1, 확정):** `project_members.project_role = 'lead'` 인 project leader 는 해당 project 의 management info(rollup/metrics/risks) 에 접근할 수 있어야 한다.
    - 대상: `GET /api/v1/projects/{id}/rollup`
    - 조건: actor.user_id 가 해당 project 의 member 이면서 `project_role = 'lead'` 이어야 함.
- **REQ-FR-ROLE-005 (P1, 확정):** Project leader 가 `project_role = 'lead'` 가 아닌(contributor) project 의 management info 요청 시 403 을 반환해야 한다.
    - 응답: `code: "auth_row_denied"`, `denied_reason: "not_project_leader"`.

### 5.3 Platform Leader 관리 정보 접근

- **REQ-FR-ROLE-006 (P1, 확정):** `Platform.LeaderUserID` 에 지정된 platform leader 는 해당 application 의 dashboard/metrics 에 접근할 수 있어야 한다.
    - 대상: `GET /api/v1/platforms/{id}/dashboard`
    - 조건: `Platform.LeaderUserID == actor.user_id`.
- **REQ-FR-ROLE-007 (P1, 확정):** Platform leader 가 leader 로 지정되지 않은 application 의 dashboard 요청 시 403 을 반환해야 한다.
    - 응답: `code: "auth_row_denied"`.

### 5.4 Org Head 부서 범위 접근

- **REQ-FR-ROLE-008 (P2, 확정):** `org_units.LeaderUserID` 에 지정된 org head 는 소속 org unit subtree 전체의 project 목록을 조회할 수 있어야 한다.
    - 대상: `GET /api/v1/projects`
    - 동작: 재귀 CTE 로 하위 org_unit 전체 조회, `development_unit_id IN (subtree unit_ids)` 조건 추가.
    - member 가 아니어도 subtree 내 project 는 조회 가능.
- **REQ-FR-ROLE-009 (P2, 확정):** Org head 는 member 가 아니어도 subtree 내 project 의 상세를 조회할 수 있어야 한다.
    - 대상: `GET /api/v1/projects/{id}` (subtree 내 project, member 아님).
    - 응답: 200 OK, 정상 project 상세.

### 5.5 Team Manager 팀 범위 관리

- **REQ-FR-ROLE-010 (P2, 확정):** `team_manager` role 사용자는 자신이 속한 org unit(primary_unit_id 기준 subtree) 전체의 project 목록을 조회할 수 있어야 한다.
    - 대상: `GET /api/v1/projects`
    - 동작: `development_unit_id IN (team_subtree unit_ids)`.
    - team scope 밖은 member 인 project 만 조회.
- **REQ-FR-ROLE-011 (P2, 확정):** `team_manager` role 사용자는 team scope 내 project 의 metadata 를 수정할 수 있어야 한다.
    - 대상: `PUT /api/v1/projects/{id}` (team scope 내).
    - 조건: project 의 `development_unit_id` 가 team_manager 의 scope 내.
- **REQ-FR-ROLE-012 (P2, 확정):** `team_manager` role 사용자가 team scope 밖 project 의 수정 요청 시 403 을 반환해야 한다.
    - 대상: `PUT /api/v1/projects/{id}` (team scope 밖).
    - 응답: `code: "auth_row_denied"`.

### 5.6 System Admin Global 접근

- **REQ-FR-ROLE-013 (MVP, 확정):** `system_admin` role 사용자는 모든 project/application 을 member 여부와 무관하게 조회할 수 있어야 한다.
    - 대상: `ListProjects`, `ListPlatforms`, `GetProject`, `GetApplication`.
    - 동작: row filter 미적용, unrestriced access.

### 5.7 Scope 통합

- **REQ-FR-ROLE-014 (P2, 확정):** 여러 scope 조건(member + org_head + team_manager 등)이 중첩될 때는 **합집합(union)** 으로 병합해야 한다.
    - 동작: 가장 넓은 scope 가 적용되며, 서로 다른 차원의 scope 는 OR 조건으로 통합.
    - 예시: developer + org_head 인 경우 org_head subtree scope + member scope 의 합집합.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-06-04 | **§4 REQ-FR-PROJDASH-001..006 신규** — 3대 페르소나별 프로젝트 대시보드 및 스위처 기능/비기능 요구사항 정의. [project_dashboard_concept.md](./project_dashboard_concept.md) 신규 발급 연계. |
| 2026-06-01 | **§5 REQ-FR-ROLE-001..016 신규** — Two-Dimensional RBAC 요구사항 16개 정의. §5.1 멤버십 baseline (ROLE-001/002/003/015/016), §5.2 project leader (ROLE-004/005), §5.3 platform leader (ROLE-006/007), §5.4 org head (ROLE-008/009), §5.5 team manager (ROLE-010/011/012), §5.6 system admin (ROLE-013), §5.7 scope 통합 (ROLE-014). |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.4 + §5.9 본문 그대로 이관. ID(REQ-FR-PROJ-000..010, REQ-FR-APP-001..012, REQ-NFR-PROJ-001..006, REQ-FR-APPDASH-001..006, REQ-NFR-APPDASH-001..003) 보존, 신규 발급/삭제 없음. |
