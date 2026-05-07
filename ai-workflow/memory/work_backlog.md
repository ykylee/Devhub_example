# 작업 백로그 인덱스

- 문서 목적: 프로젝트의 모든 작업 항목과 날짜별 백로그 링크를 관리한다.
- 범위: 전체 태스크 목록, 우선순위, 진행 상태, 날짜별 기록 연결
- 대상 독자: 개발자, AI 에이전트, 프로젝트 매니저
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [세션 인계](./session_handoff.md), [프로젝트 프로파일](./PROJECT_PROFILE.md)

> 이 flat backlog 인덱스는 legacy fallback 및 공용 색인이다. 현재 브랜치의 source of truth는 [codex/service-action-command/work_backlog.md](./codex/service-action-command/work_backlog.md)다.

## 1. 운영 원칙

1. 세션 시작 시 인덱스와 최신 백로그 확인
2. 세션 종료 전 인덱스 및 Handoff 갱신
3. 모든 작업 상태는 날짜별 백로그에 기록

## 2. 날짜별 백로그

- [2026-05-01](./backlog/2026-05-01.md)
- [2026-05-01 v2](./backlog/2026-05-01_v2.md)
- [2026-05-02](./backlog/2026-05-02.md)
- [2026-05-03](./backlog/2026-05-03.md)
- [2026-05-04](./backlog/2026-05-04.md)
- [2026-05-05](./backlog/2026-05-05.md)
- [2026-05-06](./backlog/2026-05-06.md)
- [Codex service-action-command 2026-05-07](./codex/service-action-command/backlog/2026-05-07.md)
- [Claude init 2026-05-07](./claude/init/backlog/2026-05-07.md)
- [Claude phase13 2026-05-07](./claude/phase13/backlog/2026-05-07.md)

## 3. 전체 작업 상태 요약

- [x] TASK-002: 백엔드 초기 구조 정리
- [x] TASK-003: 프론트엔드 REST API 및 WebSocket 실시간 연동
- [ ] TASK-007: AI Gardener Suggestions 및 Admin Service Actions 실체화
- [x] TASK-013: 프론트 연동 계약 안정화
- [x] TASK-014: 프론트 snapshot API 1차
- [x] TASK-015: 도메인 정규화 1차
- [x] TASK-016: DB-backed domain 조회 API
- [x] TASK-017: runtime snapshot provider
- [x] TASK-018: risk DB-backed 조회 및 CI 실패 risk 정규화
- [x] TASK-019: command/audit 최소 schema 및 risk mitigation command API
- [x] TASK-020: service action command API
- [x] TASK-021: Workflow kit v0.5.0-beta 적용 및 existing project onboarding 갱신
- [x] TASK-022: Command lifecycle 안전화 및 Admin service action 실연동
- [x] TASK-023: Command status transition worker 및 WebSocket publish 경계 구현
- [x] Phase 12: 조직/사용자 CRUD API 및 Organization UI
- [x] Phase 13: Ory Hydra/Kratos IdP PoC scaffold

## 4. 다음 작업 후보

- 프론트 `RealtimeService` command status event UI 반영
- Gitea Runner 세부 상태 adapter 또는 Gitea REST client 연동 범위 확정
- AI Gardener suggestion API/UI 연결 범위 확정
- 실제 executor/approval boundary 설계
- production command API actor verification을 JWT/session 기반으로 전환
- Phase 13 IdP PoC 실행 검증 및 DevHub OIDC callback 연동
