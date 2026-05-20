# 작업 백로그 인덱스

- 문서 목적: `codex/workflow_refactoring` 브랜치의 작업 항목과 날짜별 백로그 링크를 관리한다.
- 범위: 현재 브랜치 태스크 목록, 진행 상태, 날짜별 기록 연결.
- 대상 독자: 개발자, AI 에이전트, 프로젝트 매니저.
- 상태: in_progress
- 최종 수정일: 2026-05-20
- 관련 문서: [세션 인계](./session_handoff.md), [프로젝트 프로파일](../../PROJECT_PROFILE.md)

## 1. 운영 원칙

1. 세션 시작 시 현재 git 브랜치를 확인하고 이 브랜치 인덱스를 먼저 읽는다.
2. 세션 종료 전 `state.json`, `session_handoff.md`, `work_backlog.md`, 최신 backlog를 갱신한다.
3. flat memory는 legacy fallback 및 공용 색인으로만 사용한다.

## 2. 날짜별 백로그

- [2026-05-20](./backlog/2026-05-20.md)

## 3. 작업 상태 요약

- [ ] TASK-DOC-WORKFLOW-REFINEMENT: 문서 정책 반영 및 기존 문서 정리
- [x] TASK-DOC-MEMORY-BRANCH-INIT: 브랜치별 memory 디렉터리 초기화
