# Apply Guide

- 문서 목적: `standard-ai-workflow-antigravity` `v0.4.1-beta` 패키지를 실제 저장소에 적용하는 최소 절차를 안내한다.
- 범위: 압축 해제, 파일 복사, 하네스별 확인 포인트, 첫 세션 시작 순서
- 대상 독자: 배포 패키지를 다른 환경에서 적용하는 운영자와 개발자
- 상태: draft
- 최종 수정일: 2026-05-01
- 관련 문서: `./PACKAGE_CONTENTS.md`, `./manifest.json`

## 1. 적용 대상

- 기존 저장소에 workflow/skill 기반 온보딩 묶음을 얹고 싶을 때
- `antigravity` 하네스에서 바로 읽을 최소 runtime 파일만 가져가고 싶을 때
- 개발 참고 문서보다 실제 agent 소비 경로를 우선하고 싶을 때

## 2. 적용 절차

1. zip 파일을 풀거나 `bundle/` 디렉터리를 연다.
2. 아래 runtime 파일을 대상 저장소 루트에 복사한다.

- `bundle/ANTIGRAVITY.md -> <repo>/ANTIGRAVITY.md`
- `bundle/ai-workflow -> <repo>/ai-workflow`

3. 기존 저장소에 같은 경로의 파일이 있으면 덮어쓰기 전에 프로젝트 특화 값이 이미 들어 있는지 확인한다.
4. 복사 후 하네스 진입 파일이 `ai-workflow/memory/` 문서를 먼저 읽는지 확인한다.
5. 첫 세션에서 backlog/handoff/profile 을 실제 저장소 기준으로 갱신한다.

## 3. 하네스별 확인 포인트

- 압축을 풀고 가능하면 `ai-workflow/scripts/apply_workflow_upgrade.py` 를 사용하여 `bundle/` 내용을 반영한다. 이 스크립트는 버전 비교, .gitignore 셋업, 스테일 파일 정리를 지원한다.
- 수동 적용 시 `bundle/ANTIGRAVITY.md` 와 `bundle/ai-workflow/` 디렉터리를 대상 저장소 루트에 복사한다.
- `ANTIGRAVITY.md` 가 `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`, `PROJECT_PROFILE.md` 를 먼저 읽도록 유지한다.
- Antigravity 는 루트의 `ANTIGRAVITY.md` 를 시스템 지침에 우선 반영하며, Artifacts 와 Browser sub-agent 를 적극 활용합니다.
- 첫 세션에서는 `state.json`, `session_handoff.md`, `work_backlog.md`, 오늘 날짜 backlog 를 실제 저장소 상태로 갱신한다.

## 4. 첫 세션 권장 읽기 순서

- `ANTIGRAVITY.md`
- `ai-workflow/memory/state.json`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/PROJECT_PROFILE.md`

## 5. 적용 후 바로 수정할 항목

- `ai-workflow/memory/state.json` 의 current_focus 와 next_documents
- `ai-workflow/memory/PROJECT_PROFILE.md` 의 실행/테스트/검증 명령
- `ai-workflow/memory/session_handoff.md` 의 현재 기준선
- `ai-workflow/memory/work_backlog.md` 와 최신 날짜 backlog 의 실제 작업 상태

## 6. 주의 사항

- 이 패키지는 minimal runtime profile 이므로 개발 참고용 source docs 와 전역 snippet 예시는 기본적으로 들어 있지 않다.
- 필요하면 export 원본 저장소에서 `--include-source-docs`, `--include-global-snippets` 옵션으로 다시 패키징한다.
- MCP draft 자료는 이번 릴리즈 기본 적용 경로가 아니므로, runtime 적용 전에 별도 검토 없이는 바로 연결하지 않는다.
