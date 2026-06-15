# Work Backlog Index

- 문서 목적: 일별 작업 백로그에 대한 인덱스로, 최근 작업 흐름을 빠르게 복원한다.
- 범위: 인덱스 항목, 백로그 경로 규약, 갱신 규칙
- 대상 독자: AI agent, 저장소 maintainer
- 상태: stable
- 최종 수정일: 2026-06-09
- 관련 문서: [./PROJECT_PROFILE.md](./PROJECT_PROFILE.md), 브랜치별 daily backlog (각 브랜치 디렉터리 아래 `backlog/YYYY-MM-DD.md`)

## 인덱스 규칙

- 한 줄 = `YYYY-MM-DD 작업 백로그` 링크, 형식 예: `- [2026-06-06 작업 백로그](./release/v0.5.2/backlog/2026-06-06.md) — 요약` (실제 인덱스 항목은 각 브랜치 디렉터리(`BRANCH`) 의 실제 경로로 작성)
- 각 일자 백로그는 TASK-NNN 식별자를 가진 작업 항목 1개 이상 포함
- 같은 일자에 여러 브랜치 작업이 있으면 브랜치별로 별도 백로그 파일

## 최근 작업 백로그

### release/v0.5.10 (2026-06-09)
- [2026-06-09 작업 백로그](./release/v0.5.10/backlog/2026-06-09.md) — v0.5.10 1개 TASK (TASK-V0510-001: 개발자용 설치·사용 가이드 `docs/INSTALLATION_AND_USAGE.md` 신규 작성 + 문서 인덱스/메모리 layer 동기화)

### release/v0.5.6 (2026-06-07)
- [2026-06-07 작업 백로그](./release/v0.5.6/backlog/2026-06-07.md) — v0.5.6 1개 TASK (§5 출력 validator + §6.1 자동 위임 delegator, P0 enforcement)

### release/v0.5.5 (2026-06-07)
- [2026-06-07 작업 백로그](./release/v0.5.5/backlog/2026-06-07.md) — v0.5.5 1개 TASK (S4 live demo + Phase 11 본격 pilot, contract v1 실전 검증)

### release/v0.5.4 (2026-06-07)
- [2026-06-07 작업 백로그](./release/v0.5.4/backlog/2026-06-07.md) — v0.5.4 3개 TASK (orchestrator → sub-agent delegation contract v1 / maturity matrix 동기화 / root baseline 동기화)

### release/v0.5.3 (2026-06-07)
- [2026-06-07 작업 백로그](./release/v0.5.3/backlog/2026-06-07.md) — v0.5.3 2개 TASK (antigravity MCP config 표준화 / cross-language stack 표시)

### release/v0.5.2 (2026-06-06)
- [2026-06-06 작업 백로그](./release/v0.5.2/backlog/2026-06-06.md) — v0.5.2 3개 TASK (bootstrap 풀 리팩터 / workflow_kit 패키지화 / 외부 pilot validation)

### release/v0.5.1 (2026-06-05)
- [2026-06-05 작업 백로그](./release/v0.5.1/backlog/2026-06-05.md) — v0.5.1 = self-dogfooding 부트스트랩 + MCP 설치 가이드 + 5개 하네스 round-trip smoke (TASK-V051-001..006)

### Historical (참고용, 보존)
- [2026-05-01 작업 백로그](./codex/phase6/backlog/2026-05-01.md) — Phase 6 multi-agent delegation pilot (코드x/phase6 스냅샷)
- [2026-04-30 작업 백로그](./gemini/phase10/backlog/2026-04-24.md) — Phase 10 (MCP/JSON-RPC draft) (gemini/phase10 스냅샷)

## 다음에 읽을 문서

- [release/v0.5.6/session_handoff.md](./release/v0.5.6/session_handoff.md)
- [release/v0.5.6/backlog/2026-06-07.md](./release/v0.5.6/backlog/2026-06-07.md)
- [release/v0.5.5/session_handoff.md](./release/v0.5.5/session_handoff.md)
- [release/v0.5.5/backlog/2026-06-07.md](./release/v0.5.5/backlog/2026-06-07.md)
- [release/v0.5.4/session_handoff.md](./release/v0.5.4/session_handoff.md)
- [release/v0.5.4/backlog/2026-06-07.md](./release/v0.5.4/backlog/2026-06-07.md)
- [release/v0.5.3/session_handoff.md](./release/v0.5.3/session_handoff.md)
- [release/v0.5.3/backlog/2026-06-07.md](./release/v0.5.3/backlog/2026-06-07.md)
- [release/v0.5.2/session_handoff.md](./release/v0.5.2/session_handoff.md)
- [release/v0.5.2/backlog/2026-06-06.md](./release/v0.5.2/backlog/2026-06-06.md)
- [Maturity Matrix](../../workflow-source/core/maturity_matrix.json)
