# DevHub Project Wiki

Welcome to the DevHub Project Wiki. This directory contains all development-related documentation, including requirements, architecture, and technical specifications.

## 📖 핵심 문서 (Core Documents)

- **[요구사항 정의서 (Requirements)](./requirements.md)**: 프로젝트의 목적, 사용자 역할별 요구사항(기본 진입 우선순위 기반 UX) 및 핵심 기획 아젠다를 정의합니다.
- **[시스템 아키텍처 (Architecture)](./architecture.md)**: 시스템 구성도, 서비스 간 통신 방식(gRPC, WebSocket), 데이터 전략 및 UI 시각화 전략을 다룹니다.
- **[기술 스택 및 환경 (Tech Stack & Env)](./tech_stack.md)**: 프론트엔드(Next.js), 백엔드(Go, Python), 데이터베이스(PostgreSQL) 등 확정된 기술 스택 정보를 제공합니다.
- **[저장소 분석 리포트 (Repository Assessment)](./assessment.md)**: 초기 프로젝트 온보딩 시 수행된 코드베이스 분석 및 개선 권고 사항입니다.
- **[개발 환경 구성 가이드 (Environment Setup)](./setup/environment-setup.md)**: docker / native(no-docker) 모드별 환경 구성 절차. 컨테이너 자산은 git 추적 외부에서 관리.
- **[Docker 패키징/배포 가이드](./setup/docker-packaging-deployment-guide.md)**: 이미지 중심 배포와 compose 기반 배포를 비교하고 권장 운영 방식을 정리.
- **[통합 개발 로드맵 (Integrated Development Roadmap)](./development_roadmap.md)**: 백엔드/프론트엔드/인증/운영 트랙을 단일 마일스톤(M0~M4)·우선순위(P0~P3) 체계로 묶은 1차 진입점. 작업 시작 전 가장 먼저 확인.
- **[Application/Repository/Project 운영 컨셉](./planning/project_management_concept.md)**: 최상위 Application, 실행 단위 Repository, 기간성 운영 단위 Project 계층 모델의 기준 문서.
- **[시스템 Usecase 카탈로그](./planning/system_usecases.md)**: 모듈별 UC와 REQ↔설계 사이 추적 기준.
- **[시스템 ERD 카탈로그](./planning/system_erd.md)**: 모듈별 데이터 모델과 통합 ERD.

## 🛠 워크플로우 및 운영 (Workflow & Operations)

AI 워크플로우와 관련된 메타 데이터 및 세션 관리 문서는 `ai-workflow/` 디렉토리에서 관리됩니다.

- **[프로젝트 워크플로우 프로파일](../ai-workflow/memory/PROJECT_PROFILE.md)**: 프로젝트 특화 규칙 및 기본 명령 정의.
- **[작업 백로그](../ai-workflow/memory/work_backlog.md)**: 현재 진행 중인 작업과 향후 계획.
- **[세션 인계 문서](../ai-workflow/memory/session_handoff.md)**: 세션 간 작업 상태 공유를 위한 문서.

---
*Last updated: 2026-05-13*
