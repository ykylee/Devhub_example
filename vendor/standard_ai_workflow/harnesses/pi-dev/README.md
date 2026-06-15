# Harness: pi-dev

- 문서 목적: Pi Coding Agent (pi.dev) 환경에서 Standard AI Workflow를 운영하기 위한 하네스 설정을 관리한다.
- 범위: AGENTS.md, SYSTEM.md, Pi 전용 스킬 정의
- 대상 독자: Pi 에이전트 사용자, 저장소 관리자
- 상태: prototype
- 최종 수정일: 2026-04-27
- 관련 문서: `../../core/global_workflow_standard.md`, `../../README.md`

## 개요
Pi Coding Agent는 "Primitives, not features" 철학을 가진 미니멀한 에이전트입니다. 본 하네스는 Pi가 프로젝트의 `state.json`과 `session_handoff.md`를 엄격하게 관리하도록 지침을 제공합니다.

## 포함 파일
- `AGENTS.md`: Pi 에이전트용 프로젝트 진입 지침.
- `SYSTEM.md`: Pi 에이전트 전용 페르소나 및 운영 원칙 보강.
- `apply_guide.md`: Pi 하네스 적용 및 실행 방법 안내.
