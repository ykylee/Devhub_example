# DevHub Example

- 문서 목적: DevHub Example 프로젝트의 개요 및 통합 가이드를 제공한다.
- 범위: 전체 시스템 구조, 개발 환경 구축, 서브 시스템 안내
- 대상 독자: 모든 개발자 및 운영자
- 상태: stable
- 최종 수정일: 2026-05-01
- 관련 문서: [워크플로우 README](ai-workflow/README.md), [프로젝트 프로파일](ai-workflow/memory/PROJECT_PROFILE.md)

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

## 📚 Documentation (Wiki)

모든 개발 관련 문서는 **[Project Wiki (docs/)](./docs/README.md)**에서 확인할 수 있습니다.

- [요구사항 정의서](./docs/requirements.md)
- [시스템 아키텍처 설계](./docs/architecture.md)
- [기술 스택 및 환경 설정](./docs/tech_stack.md)
- [개발 환경 구성 가이드 (docker / native)](./docs/setup/environment-setup.md)

## 🏗 아키텍처 원칙
- **Role-Prioritized Entry:** 역할별 기본 진입 페이지 우선순위로 UX를 간접 제공.
- **Permission-Gated System Area:** 시스템 대시보드/시스템 설정은 `system_admin` 권한 사용자에게만 노출.
- **Shared Design System:** 대시보드가 달라도 일관된 시각적 경험을 위해 공통 컴포넌트 라이브러리 활용.
