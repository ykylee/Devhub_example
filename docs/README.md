# DevHub Project Wiki

Welcome to the DevHub Project Wiki. This directory contains all development-related documentation, including requirements, architecture, and technical specifications.

## 📖 핵심 문서 (Core Documents)

- **[요구사항 정의서 (Requirements)](./requirements.md)**: 프로젝트의 목적, 사용자 역할별 요구사항 및 핵심 기획 아젠다를 정의합니다.
- **[시스템 아키텍처 (Architecture)](./architecture.md)**: 시스템 구성도, 서비스 간 통신 방식(gRPC, WebSocket), 데이터 전략 및 UI 시각화 전략을 다룹니다.
- **[기술 스택 및 환경 (Tech Stack & Env)](./tech_stack.md)**: 프론트엔드(Next.js), 백엔드(Go, Python), 데이터베이스(PostgreSQL) 등 확정된 기술 스택 정보를 제공합니다.
- **[저장소 분석 리포트 (Repository Assessment)](./assessment.md)**: 초기 프로젝트 온보딩 시 수행된 코드베이스 분석 및 개선 권고 사항입니다.
- **[개발 환경 구성 가이드 (Environment Setup)](./setup/environment-setup.md)**: docker / native(no-docker) 모드별 환경 구성 절차. 컨테이너 자산은 git 추적 외부에서 관리.

## 🛠 워크플로우 및 운영 (Workflow & Operations)

AI 워크플로우와 관련된 메타 데이터 및 세션 관리 문서는 `ai-workflow/` 디렉토리에서 관리됩니다.

- **[프로젝트 워크플로우 프로파일](../ai-workflow/memory/PROJECT_PROFILE.md)**: 프로젝트 특화 규칙 및 기본 명령 정의.
- **[작업 백로그](../ai-workflow/memory/work_backlog.md)**: 현재 진행 중인 작업과 향후 계획.
- **[세션 인계 문서](../ai-workflow/memory/session_handoff.md)**: 세션 간 작업 상태 공유를 위한 문서.

---
*Last updated: 2026-04-29*
