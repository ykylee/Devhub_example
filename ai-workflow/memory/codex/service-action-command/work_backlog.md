# 작업 백로그 인덱스

- 문서 목적: codex/service-action-command 브랜치 작업 항목과 날짜별 백로그 링크를 관리한다.
- 범위: 현재 브랜치 태스크 목록, 진행 상태, 날짜별 기록 연결
- 대상 독자: 개발자, AI 에이전트, 프로젝트 매니저
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [세션 인계](./session_handoff.md), [프로젝트 프로파일](../../PROJECT_PROFILE.md)

## 1. 운영 원칙

1. 세션 시작 시 현재 git 브랜치를 확인하고 이 브랜치별 인덱스를 읽는다.
2. 세션 종료 전 이 브랜치의 `state.json`, `session_handoff.md`, 최신 backlog를 갱신한다.
3. flat memory는 legacy fallback 및 공용 색인으로만 사용한다.

## 2. 날짜별 백로그

- [2026-05-07](./backlog/2026-05-07.md)

## 3. 작업 상태 요약

- [x] TASK-020: service action command API
- [x] TASK-021: Workflow kit v0.5.0-beta 적용 및 existing project onboarding 갱신
- [x] TASK-022: Command lifecycle 안전화 및 Admin service action 실연동
- [x] TASK-023: Command status transition worker 및 WebSocket publish 경계 구현
- [x] TASK-CODEX-MEMORY-BRANCH-SPLIT: 브랜치별 ai-workflow/memory 구조 정리
- [x] TASK-CODEX-BACKEND-ROADMAP-REVIEW: main 반영 사항 기준 백엔드 로드맵 재검토
- [x] TASK-BACKEND-P0-AUDIT-ACTOR: audit 조회 API 및 actor fallback deprecation
- [x] TASK-BACKEND-P0-AUTH-BOUNDARY: Hydra/Kratos API 계약 재작성 및 Bearer actor 경계
- [x] TASK-BACKEND-P0-RBAC-POLICY: RBAC policy 조회 API 및 프론트 Permissions 연동 준비
- [ ] TASK-007: AI Gardener Suggestions 및 Admin Service Actions 실체화

## 4. 다음 작업 후보

- RBAC policy persistence/edit API와 audit/approval 경계 설계
- Hydra JWKS/introspection verifier 실제 구현
- `/api/v1/me`와 role/permission lookup 순서 확정
- 프론트 `RealtimeService` command status event UI 반영
- Gitea Runner 세부 상태 adapter 또는 Gitea REST client 연동 범위 확정
- AI Gardener suggestion API/UI 연결 범위 확정
- 실제 executor/approval boundary 설계
- production command API actor verification을 JWT/session 기반으로 전환
