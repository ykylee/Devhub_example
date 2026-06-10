# DevHub Example

- 문서 목적: DevHub Example 프로젝트의 개요 및 통합 가이드를 제공한다.
- **저장소 이름** (이 GitHub repo) = "DevHub Example" — 사외 공개 시 사용되는 코드 명. 사내 deploy 시에도 동일 코드 베이스가 사용되지만 **product 명칭** (브라우저 tab title / sidebar brand / 페이지 heading) 은 `DEVHUB_APP_NAME` / `DEVHUB_APP_SHORT_NAME` env var 로 override 가능 (사외=DevHub, 사내=운영자 결정). 자세한 정책: [PR #532 commit message + docs/governance/worker_division.md §6.7](docs/governance/worker_division.md#67-명명-재검토-2026-06-10).
- 범위: 전체 시스템 구조, 개발 환경 구축, 서브 시스템 안내, SDLC 도메인 진입점
- 대상 독자: 모든 개발자 및 운영자
- 상태: stable
- 최종 수정일: 2026-05-29 (SDLC 재정비 sprint #408~#416 후속 — code-taxonomy SoT 4 계층 / 10 도메인 + Shared + Infrastructure 구조 반영)
- 관련 문서: [워크플로우 README](ai-workflow/README.md), [프로젝트 프로파일](ai-workflow/memory/PROJECT_PROFILE.md), [환경 구성 가이드](docs/setup/environment-setup.md), [개발 로드맵](docs/development_roadmap.md), [Code Taxonomy](docs/governance/code-taxonomy.md)

# DevHub: Role-Prioritized Entry Team Hub

본 프로젝트는 역할별 UX를 **진입 페이지 우선순위**로 간접 제공하는 통합 개발 허브입니다.

## 🎯 핵심 목적 및 확장성
단순히 고정된 기능을 제공하는 것이 아니라, **역할별 기본 진입 경로를 우선 배정하고 기능 영역은 권한에 따라 노출**하는 구조를 지향합니다.

## 👥 현재 및 향후 지원 사용자
- **개발 대시보드 (Developer Dashboard):** 개발 업무 흐름, 기술 문서, API/CI 상태 중심.
- **관리 대시보드 (Management Dashboard):** 과제 진행률, 리소스 현황, 리스크 모니터링 중심.
- **시스템 대시보드 + 시스템 설정 (System Dashboard + System Settings):** 시스템 운영/보안/인프라 제어 영역. **시스템 관리자 권한 보유자에게만 노출**.
- **확장 가능 (Extensible):**
    - **QA/테스트 담당자:** 테스트 케이스 관리, 결함 현황, 배포 승인 UI.
    - **기획자/디자이너:** 요구사항 정의서, 디자인 에셋 링크, 마일스톤 관리.
    - **운영자:** 시스템 모니터링, 장애 전파, 배포 이력 관리.

## 🏛 코드베이스 구조 (SDLC SoT)

본 저장소는 [`docs/governance/code-taxonomy.md`](./docs/governance/code-taxonomy.md) 를 single source-of-truth 로 하여 **3대 레이어 + 도메인 내부 4대 계층** 구조를 따릅니다.

### 3대 레이어

1. **Domain** (비즈니스 핵심) — 10 core 도메인:
   `auth-session`, `audit-ops`, `rbac-permissions`, `organization-management`, `onboarding`,
   `platform-lifecycle`, `repository-integration`, `dev-request`, `integration-registry`, `realtime`
2. **Shared** (공통 기반) — `config`, `logger`, `utils`, `ui-foundation`, `integrationcaps`
3. **Infrastructure** (외부 기술 구현) — `keycloak-idp`, `gitea-scm`, `hrdb`, `commandworker`, `infra-topology`, `database-migration`, `deployment-automation`

### 도메인 내부 4대 계층

각 Domain 도메인은 `view` (API/UI 접점) / `service` (비즈니스 제어) / `repository` (영속성 추상화) / `schema` (엔티티/DTO) 로 수직 격리됩니다. Backend / Frontend 모두 `<lang>/domain/<도메인>/<계층>/` 구조로 코드와 문서가 1:1 mirror 관계입니다.

- Backend: `backend-core/internal/domain/<도메인>/{view,service,repository,schema}`
- Frontend: `frontend/domain/<도메인>/{view,service,schema}` (repository 는 backend mirror)

## 📚 Documentation (Wiki)

모든 개발 문서는 **[Project Wiki (docs/)](./docs/README.md)** 에서 진입할 수 있습니다. 새 SDLC 구조에서는 **도메인별 진입점** + **master index** + **거버넌스** 3축으로 나뉩니다.

### 도메인별 SDLC 진입점 (10 core 도메인)

각 도메인은 `docs/domain/<도메인>/README.md` 가 진입점이며, 그 아래 `requirements.md` / `architecture.md` / `api.md` / `test_cases.md` 가 도메인별 SoT 입니다.

- [Domain index](./docs/domain/README.md) — 10 도메인 진입점 표
- [Shared layer index](./docs/shared/README.md) — 공통 기반 모듈
- [Infrastructure layer index](./docs/infrastructure/README.md) — 외부 기술 구현체

### Master index (cross-cutting)

- [요구사항 정의서 (master index)](./docs/requirements.md) — §5 도메인 link 표 + cross-cutting 정책
- [시스템 아키텍처 설계 (master index)](./docs/architecture.md) — 3대 레이어 + 호출 규칙 + 도메인 link 표
- [백엔드 API 계약 (master index)](./docs/backend_api_contract.md) — envelope/enum cross-cutting + 도메인 link 표
- [API 공통 규약](./docs/api/conventions.md) — envelope + 공통 enum
- [통합 개발 로드맵](./docs/development_roadmap.md) — 마일스톤 + 도메인 트랙
- [v1.0 릴리즈 로드맵](./docs/planning/release_v1_roadmap.md)

### 거버넌스 + 추적성

- [Code Taxonomy (SoT)](./docs/governance/code-taxonomy.md) — 3대 레이어 + 4대 계층 + 10 도메인
- [Worker Division (Claude 영역)](./docs/governance/worker_division.md)
- [Document Standards](./docs/governance/document-standards.md) — 메타 헤더 / lifecycle / 단계별 위치 가이드
- [Traceability Report](./docs/traceability/report.md) — REQ → UC → ARCH → API → RM → IMPL → UT → TC 19 row 매트릭스

### 환경 / 설정

- [기술 스택](./docs/shared/tech_stack.md) (Shared 레이어 진입점)
- [개발 환경 구성 가이드 (docker / native)](./docs/setup/environment-setup.md)

## 🏗 아키텍처 원칙
- **Role-Prioritized Entry:** 역할별 기본 진입 페이지 우선순위로 UX를 간접 제공.
- **Permission-Gated System Area:** 시스템 대시보드/시스템 설정은 `system_admin` 권한 사용자에게만 노출.
- **Shared Design System:** 대시보드가 달라도 일관된 시각적 경험을 위해 공통 컴포넌트 라이브러리 활용 (`frontend/shared/ui-foundation/`).
- **Layered Domain Isolation:** Domain 도메인은 Shared 만 import 가능. Infrastructure 는 Domain interface 의 기술 구현체로 격리 (`docs/architecture.md §2.2` 호출 규칙).
