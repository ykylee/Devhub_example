# Document Index

- 문서 목적: 프로젝트의 영구 지식 자산(Knowledge Base)을 체계적으로 정리하여 개발자와 AI 에이전트의 온보딩 및 분석을 돕는다.
- 범위: 프로젝트 설계, 개발 및 표준, 분석 및 계획
- 대상 독자: 개발자, AI 에이전트, 프로젝트 온보딩 담당자
- 상태: stable
- 최종 수정일: 2026-05-10
- 관련 문서: [./README.md](./README.md), [../README.md](../README.md)

이 인덱스는 프로젝트의 영구 지식 자산(Knowledge Base)을 체계적으로 정리하여 개발자와 AI 에이전트의 온보딩 및 분석을 돕습니다.

> [!NOTE]
> 세션 상태, 백로그 등 워크플로우 운영 문서는 `ai-workflow/` 폴더를 참조하십시오. 이 인덱스는 프로젝트 자체의 설계와 개발 지침만을 포함합니다.

## 1. 프로젝트 설계 (Architecture)
*시스템 구조 및 핵심 설계 원칙을 다룹니다.*

- **[Project Profile](./PROJECT_PROFILE.md)**: 프로젝트 개요, 목적, 설치/실행 가이드 및 워크플로우 기본 규칙.
- **[Architecture (본문)](./architecture.md)**: 시스템 컴포넌트, 통신, 데이터 전략, 보안, UI/UX 시각화 전략의 narrative source-of-truth.
- **[Architecture 디렉터리 진입점](./architecture/README.md)**: 시스템 본문 / ADR 인덱스 / 컴포넌트별 설계 자료의 디렉터리 안내. ADR-0001(IdP 선택) 등 결정 기록 진입점.

## 2. 개발 및 표준 (Development)
*코드 작성 및 문서 관리 표준을 다룹니다.*

- **[Documentation Governance](./README.md)**: 문서 분류 체계 및 PR 리뷰 프로세스.
- **[Code Index](./CODE_INDEX.md)**: 코드베이스 구조 및 핵심 컴포넌트 안내.
- **[개발 환경 구성](./setup/environment-setup.md)**: 로컬 개발 환경 (native default + docker option) 셋업.
- **[테스트 서버 배포 가이드](./setup/test-server-deployment.md)**: 단일 테스트 서버에 native binary 로 PostgreSQL/Hydra/Kratos/backend/frontend 5종 빌드·기동.

## 3. 분석 및 계획 (Analysis & Planning)
*요구사항 분석 및 상위 수준의 로드맵을 다룹니다.*

- **[통합 개발 로드맵 (Integrated Development Roadmap)](./development_roadmap.md)**: 백엔드/프론트엔드/인증/운영 트랙을 마일스톤(M0~M4)과 우선순위(P0~P3) 체계로 묶은 1차 진입점. 모든 트랙이 작업 시작 전 가장 먼저 확인.
- **[Backend Development Roadmap](../ai-workflow/memory/backend_development_roadmap.md)**: 백엔드 트랙 세부 (Phase 1~13).
- **[Frontend Development Roadmap](./frontend_development_roadmap.md)**: 프론트엔드 트랙 세부 (Phase 1~7, 역할별 기본 진입 우선순위 UX 기준 포함).
- **[Planning 디렉터리 진입점](./planning/README.md)**: 마일스톤·트랙·PR/보안 트래커 인덱스. backlog 위치 정책과 향후 추가 예정 자료(sprint plan, release plan 등) 안내.

---
*모든 신규 문서는 `docs/` 하위의 적절한 카테고리에 생성되어야 하며, PR 리뷰를 통해 승인되어야 합니다.*
