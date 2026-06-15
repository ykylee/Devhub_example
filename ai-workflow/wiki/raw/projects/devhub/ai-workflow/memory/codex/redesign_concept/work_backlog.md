# 작업 백로그 인덱스

- 문서 목적: `codex/redesign_concept` 브랜치 작업 항목과 날짜별 백로그 링크를 관리한다.
- 범위: 현재 브랜치 태스크 목록, 진행 상태, 날짜별 기록 연결
- 대상 독자: 개발자, AI 에이전트, 프로젝트 매니저
- 상태: active
- 최종 수정일: 2026-05-10
- 관련 문서: [세션 인계](./session_handoff.md), [프로젝트 프로파일](../../PROJECT_PROFILE.md)

## 1. 운영 원칙

1. 세션 시작 시 현재 git 브랜치를 확인하고 이 브랜치별 인덱스를 읽는다.
2. 세션 종료 전 이 브랜치의 `state.json`, `session_handoff.md`, 최신 backlog를 갱신한다.
3. flat memory는 legacy fallback 및 공용 색인으로만 사용한다.

## 2. 날짜별 백로그

- [2026-05-10](./backlog/2026-05-10.md)

## 3. 작업 상태 요약

- [x] TASK-BRANCH-SETUP-CODEX-REDESIGN-CONCEPT: 브랜치 및 memory 초기화
- [x] TASK-REDESIGN-CONCEPT-DOC-UPDATE-V1: 역할별 진입 우선순위 기반 UX 컨셉 문서 1차 반영
- [x] TASK-REDESIGN-CONCEPT-ROADMAP-PLAN-UPDATE-V1: 로드맵/개발 계획 문서 전반 정렬
- [ ] TASK-REDESIGN-CONCEPT-REFRAME: 개발 컨셉 재정리

## 4. 다음 작업 후보

- 역할별 진입 우선순위(개발/관리/시스템+설정) 라우팅 규칙 상세화
- 시스템 관리자 전용 메뉴 비노출/접근 차단 정책을 UI + API 계약으로 정합화
- 변경 후 `make build`, `cd backend-core && go test ./...`, `cd frontend && npm run lint` 검증
