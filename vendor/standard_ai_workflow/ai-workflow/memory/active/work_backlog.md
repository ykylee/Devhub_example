# Work Backlog Index

- 문서 목적: 일별 작업 백로그에 대한 인덱스로, 최근 작업 흐름을 빠르게 복원한다.
- 범위: 인덱스 항목, 백로그 경로 규약, 갱신 규칙
- 대상 독자: AI agent, 저장소 maintainer
- 상태: stable
- 최종 수정일: 2026-06-09
- 관련 문서: [./PROJECT_PROFILE.md](./PROJECT_PROFILE.md), 브랜치별 daily backlog (각 브랜치 디렉터리 아래 `backlog/YYYY-MM-DD.md`)

## 인덱스 규칙

- 각 `### [[release/v0.5.X/backlog/YYYY-MM-DD.md]] {#release-v0-5-X}` 형식의 anchor entry 로 인덱스 표시
- anchor ID 로 직접 retrieval (session-start 의 index-based load)
- 각 일자 백로그는 TASK-NNN 식별자를 가진 작업 항목 1개 이상 포함
- 같은 일자에 여러 브랜치 작업이 있으면 브랜치별로 별도 백로그 파일

## 최근 작업 백로그

### [[release/v0.7.17/backlog/2026-06-15.md]] {#release-v0-7-17}
- 2026-06-15: v0.7.17 1개 TASK (wiki in-repo storage isolation — 5 file redirect + ai-workflow/wiki/sources/ 신규 + 11 smoke)

### [[release/v0.7.16/backlog/2026-06-15.md]] {#release-v0-7-16}
- 2026-06-15: v0.7.16 4개 TASK (config score_alert + memory_alert_mb consumer + linter excluded_paths + linter IndentationError/dict-slice fix + 9 smoke)

### [[release/v0.7.15/backlog/2026-06-15.md]] {#release-v0-7-15}
- 2026-06-15: v0.7.15 2개 TASK (atomic_write helper + changelog-gen --from-tag/--to-tag filter + 5 smoke)

### [[release/v0.7.14/backlog/2026-06-15.md]] {#release-v0-7-14}
- 2026-06-15: v0.7.14 4개 TASK (cmd_version_bump auto-sync + cmd_changelog_gen subcommand + 8 smoke + state sync)

### [[release/v0.7.13/backlog/2026-06-15.md]] {#release-v0-7-13}
- 2026-06-15: v0.7.13 3개 TASK (cmd_release --version flag + 3 smoke + version sync)

### [[release/v0.7.12/backlog/2026-06-15.md]] {#release-v0-7-12}
- 2026-06-15: v0.7.12 3개 TASK (REPO_ROOT auto-detect + 4 smoke + v0.7.5~v0.7.10 backfill)

### [[release/v0.7.11/backlog/2026-06-15.md]] {#release-v0-7-11}
- 2026-06-15: v0.7.11 5개 TASK (dist subcommand + state sync backfill + version string sync + release note + session_handoff)

### [[release/v0.7.10/backlog/2026-06-14.md]] {#release-v0-7-10}
- 2026-06-14: v0.7.10 3개 TASK (release / verify / rollback — release_pipeline Phase 2)

### [[release/v0.7.9/backlog/2026-06-14.md]] {#release-v0-7-9}
- 2026-06-14: v0.7.9 3개 TASK (validate / version-bump / note-draft — release_pipeline Phase 1)

### [[release/v0.7.8/backlog/2026-06-14.md]] {#release-v0-7-8}
- 2026-06-14: v0.7.8 2개 TASK (state-aware evaluate_compliance + config actual apply)

### [[release/v0.7.7/backlog/2026-06-14.md]] {#release-v0-7-7}
- 2026-06-14: v0.7.7 1개 TASK (workflow_kit.cli.doctor × load_config integration)

### [[release/v0.7.6/backlog/2026-06-14.md]] {#release-v0-7-6}
- 2026-06-14: v0.7.6 2개 TASK (run_all_checks 통합 runner + workflow_kit.metadata loader)

### [[release/v0.7.5/backlog/2026-06-14.md]] {#release-v0-7-5}
- 2026-06-14: v0.7.5 2개 TASK (Extension audit 4 dispatcher + refresh_wiki_memory 정식화)

### [[release/v0.7.4/backlog/2026-06-13.md]] {#release-v0-7-4}
- 2026-06-13: v0.7.4 3개 TASK (workflow doctor CLI + @graceful_shutdown + optional dep)

### [[release/v0.7.3/backlog/2026-06-13.md]] {#release-v0-7-3}
- 2026-06-13: v0.7.3 2개 TASK (4 Runtime Helper + 7 baseline dispatcher)

### [[release/v0.7.2/backlog/2026-06-13.md]] {#release-v0-7-2}
- 2026-06-13: v0.7.2 2개 TASK (Extension sub-cat lint + 4 baseline 본 구현)

### [[release/v0.7.1/backlog/2026-06-13.md]] {#release-v0-7-1}
- 2026-06-13: v0.7.1 5개 TASK (score dashboard / L2 emit / 위치 lint / drift check / follow-up 묶음)

### [[release/v0.7.0/backlog/2026-06-13.md]] {#release-v0-7-0}
- 2026-06-13: v0.7.0 7개 TASK (AIDLC Extension + Reverse Engineering + UOW + audit log + Stage Completion + Security Baseline + packaging)

### [[release/v0.5.10/backlog/2026-06-09.md]] {#release-v0-5-10}
- 2026-06-09: v0.5.10 1개 TASK (TASK-V0510-001: 개발자용 설치·사용 가이드)

### [[release/v0.5.6/backlog/2026-06-07.md]] {#release-v0-5-6}
- 2026-06-07: v0.5.6 1개 TASK (§5 출력 validator + §6.1 자동 위임 delegator)

### [[release/v0.5.5/backlog/2026-06-07.md]] {#release-v0-5-5}
- 2026-06-07: v0.5.5 1개 TASK (Phase 11 본격 pilot, contract v1 실전 검증)

### [[release/v0.5.4/backlog/2026-06-07.md]] {#release-v0-5-4}
- 2026-06-07: v0.5.4 3개 TASK (orchestrator delegation contract v1 / maturity matrix)

### [[release/v0.5.3/backlog/2026-06-07.md]] {#release-v0-5-3}
- 2026-06-07: v0.5.3 2개 TASK (antigravity MCP config 표준화 / cross-language stack)

### [[release/v0.5.2/backlog/2026-06-06.md]] {#release-v0-5-2}
- 2026-06-06: v0.5.2 3개 TASK (bootstrap 리팩터 / workflow_kit 패키지화)

### [[release/v0.5.1/backlog/2026-06-05.md]] {#release-v0-5-1}
- 2026-06-05: v0.5.1 self-dogfooding bootstrap + MCP 설치 가이드 (TASK-V051-001..006)

### Historical archives {#historical-archives}
### [[codex/phase6/backlog/2026-05-01.md]] {#codex-phase6}
- 2026-05-01: Phase 6 multi-agent delegation pilot
### [[gemini/phase10/backlog/2026-04-24.md]] {#gemini-phase10}
- 2026-04-24: Phase 10 MCP/JSON-RPC draft

## 다음에 읽을 문서

- [release/v0.5.6/session_handoff.md](../release/v0.5.6/session_handoff.md)
- [release/v0.5.6/backlog/2026-06-07.md](../release/v0.5.6/backlog/2026-06-07.md)
- [release/v0.5.5/session_handoff.md](../release/v0.5.5/session_handoff.md)
- [release/v0.5.5/backlog/2026-06-07.md](../release/v0.5.5/backlog/2026-06-07.md)
- [release/v0.5.4/session_handoff.md](../release/v0.5.4/session_handoff.md)
- [release/v0.5.4/backlog/2026-06-07.md](../release/v0.5.4/backlog/2026-06-07.md)
- [release/v0.5.3/session_handoff.md](../release/v0.5.3/session_handoff.md)
- [release/v0.5.3/backlog/2026-06-07.md](../release/v0.5.3/backlog/2026-06-07.md)
- [release/v0.5.2/session_handoff.md](../release/v0.5.2/session_handoff.md)
- [release/v0.5.2/backlog/2026-06-06.md](../release/v0.5.2/backlog/2026-06-06.md)
- [Maturity Matrix](../../workflow-source/core/maturity_matrix.json)
