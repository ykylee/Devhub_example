# Beta v0.5.4 — Orchestrator ↔ Sub-agent Delegation Contract v1

- **릴리스 일자**: 2026-06-07
- **브랜치**: `release/v0.5.4`
- **포함 커밋**: v0.5.3 (PR #17) 이후 1 squash
- **상태**: ✅ 3개 TASK 모두 완료 (이슈 #1 영구 해결)

## 1. 하이라이트

v0.5.3 까지 "원칙" 으로만 존재하던 오케스트레이터 ↔ 서브 에이전트 위임을 **외부 contract v1** 로 박아 영구 spec 화.

- **TASK-V054-001**: [`workflow-source/core/orchestrator_subagent_contract_v1.md`](../core/orchestrator_subagent_contract_v1.md) 신규. 입력/출력 JSON 스키마, 4개 역할 경계, 위임 가능/불가 카탈로그, 에러/폴백 정책, 4개 검증 시나리오 명시
- **TASK-V054-002**: [`workflow-source/core/maturity_matrix.json`](../core/maturity_matrix.json) Phase 9/10 → `done`, Phase 11 → `in_progress`
- **TASK-V054-003**: 루트 `ai-workflow/memory/state.json` + `docs/PROJECT_PROFILE.md` v0.5.3 stale → v0.5.4 baseline 동기화

**이슈 #1 영구 해결**: [ykylee/standard_ai_workflow#1](https://github.com/ykylee/standard_ai_workflow/issues/1) ("오케스트레이터가 서브 에이전트를 호출하지 않고 직접 도구를 사용") — 2026-04-24 부터 open 이던 가장 오래된 bug. v0.5.4 PR 머지 후 close.

## 2. TASK-V054-001: Orchestrator ↔ Sub-agent Contract v1

### 동기

[이슈 #1](https://github.com/ykylee/standard_ai_workflow/issues/1) 에서 보고된 현상:

> "오케스트레이터는 과제 전체를 관리하는 역할로 실제 작업은 서브 에이전트에게 일임하도록 정의했는데, 테스트 환경에서 사용해보니 실제로는 직접 모든 작업을 하고 있는 현상을 발견"

[workflow_agent_topology.md §5](../core/workflow_agent_topology.md) 에서도 다음 원칙을 명시:

> "메인 오케스트레이터는 읽기/쓰기 작업을 직접 모두 떠안기보다 서브 에이전트를 적극 활용해 작업을 분담한다. … 대량 파일 탐색, 문서 비교, 로그 확인, 초안 생성처럼 컨텍스트를 빠르게 부풀리는 작업은 가능한 한 서브 에이전트로 분리한다."

그러나 v0.5.3 까지 "원칙" 만 있고 외부 contract 가 없어, 구현체/도구/세션 환경에 따라 orchestrator 가 직접 도구를 호출하는 회귀가 반복됐다. v1 은 이 contract 를 **외부 명세**로 박아, 후속 구현체(메인 오케스트레이터 런타임, 워커 에이전트, 검증 도구) 가 같은 기준을 따르도록 강제한다.

### 변경

#### 2.1 신규 spec doc

[`workflow-source/core/orchestrator_subagent_contract_v1.md`](../core/orchestrator_subagent_contract_v1.md) (11개 섹션, ~310줄)

| 섹션 | 내용 |
| --- | --- |
| §1 | 동기 — 이슈 #1 + topology §5 의 권장 운영 원칙을 외부 spec 으로 격상 |
| §2 | 적용 범위 — v0.5.4 이상, Mavis / mini-coder-max / general 호환 |
| §3 | 4개 역할 정의 — orchestrator / doc-worker / code-worker / validation-worker (+ workflow-worker 임시) |
| §4 | 위임 입력 스키마 (orchestrator → sub-agent) — JSON, 17 필드 |
| §5 | 위임 출력 스키마 (sub-agent → orchestrator) — JSON, 14 필드 |
| §6 | 위임 가능/불가 카탈로그 — §6.1 MUST delegate, §6.3 MUST NOT delegate |
| §7 | 에러/폴백 정책 — failed / partial / timeout / 스키마 위반 / 검증 실패 |
| §8 | 검증 시나리오 — S1~S4 (roundtrip / role mapping / direct-only / live demo) |
| §9 | 구현 가이드 — 오케스트레이터 / sub-agent / 검증 도구 측 |
| §10 | v0.5.3 → v0.5.4 마이그레이션 |
| §11 | v2 후보 (스트리밍, 멀티 fan-out, observability) |

#### 2.2 신규 회귀 테스트 3개

| 파일 | S# | 검증 |
| --- | --- | --- |
| [`tests/check_contract_v1_roundtrip.py`](../tests/check_contract_v1_roundtrip.py) | S1 | 입력/출력 JSON 스키마 + delegation_id 보존 + role 매핑 |
| [`tests/check_contract_v1_role_mapping.py`](../tests/check_contract_v1_role_mapping.py) | S2 | §6.1 8개 MUST-delegate 행이 §3 역할 정의와 일치 |
| [`tests/check_contract_v1_direct_only.py`](../tests/check_contract_v1_direct_only.py) | S3 | §6.3 7개 MUST-NOT-delegate 액션이 orchestrator-only 로 거부됨 + §8.3 negative-test examples 존재 |

#### 2.3 Cross-link 갱신

- `workflow-source/core/workflow_agent_topology.md` — 상단 + §5 운영 원칙에 contract v1 링크 추가, 상태 `draft` → `stable`
- `workflow-source/core/global_workflow_standard.md` — 상단 + §1.2 말미에 contract v1 링크 추가
- `ai-workflow/README.md` — §6 운영 원칙 상단에 contract v1 링크 추가
- `docs/PROJECT_PROFILE.md` — 상단 + §4 검증 + §5 예외 규칙 + "다음에 읽을 문서" 에 contract v1 링크 추가

### 라이브 데모 (S4)

TASK-V054-001 의 spec doc 작성 자체가 contract v1 의 첫 라이브 데모 대상:

- 오케스트레이터 (Mavis / mvs_a96f8eb4990a482ca14e3e5223447bb7) 가 doc-worker 에게 contract spec 의 1차 초안 작성을 위임
- doc-worker 가 초안 작성 + validation
- 오케스트레이터가 통합/리뷰/최종 커밋
- PR 본문에 "S4 라이브 데모" 섹션 추가

## 3. TASK-V054-002: Maturity Matrix 동기화

### 동기

v0.5.3 까지 모든 작업이 머지됐는데 maturity matrix 의 `Phase 9`, `Phase 10` 가 아직 `in_progress`. 또한 v0.5.2 pilot validation 으로 `Phase 11` 도 실질적으로 in_progress 진입.

### 변경

| Phase | 이름 | Before | After |
| --- | --- | --- | --- |
| Phase 9 | System Maturity & Multi-Agent Evolution | `in_progress` | `done` (PR #12 머지로 완료) |
| Phase 10 | Document & Link Hygiene | `in_progress` | `done` (memory layer path 정책 + cross-link 정합성 작업 완료) |
| Phase 11 | Real-world Pilot Validation | `planned` | `in_progress` (v0.5.2 Devhub_example pilot 시작, v0.5.4 부터 본격) |
| `last_updated` | | `2026-05-03` | `2026-06-07` |

## 4. TASK-V054-003: Root Baseline 동기화

### 동기

루트 baseline 문서들이 v0.5.1 / Phase 10 상태로 stale. v0.5.3 release/v0.5.3/ 의 source-of-truth 와 동기화 필요.

### 변경

| 파일 | Before | After |
| --- | --- | --- |
| `ai-workflow/memory/state.json` (루트) | 1줄, v0.5.1 / Phase 10 in_progress | v0.5.4 baseline, TASK-V054-001/002/003 in_progress |
| `docs/PROJECT_PROFILE.md` (루트) | draft, 2026-05-02, antigravity 부트스트랩 | stable, 2026-06-07, codex 부트스트랩, contract v1 회귀 명령, venv 셋업 단계, contract v1 링크 |

## 5. 회귀 테스트 종합

| 테스트 | 결과 |
| --- | --- |
| `check_bootstrap.py` (8 모드) | ✅ 8/8 |
| `check_bootstrap_mcp_roundtrip.py` (5 하네스) | ✅ 5/5 |
| `check_docs.py` | ✅ 102 markdown PASS |
| `check_contract_v1_roundtrip.py` (S1) | ✅ |
| `check_contract_v1_role_mapping.py` (S2) | ✅ |
| `check_contract_v1_direct_only.py` (S3) | ✅ |
| Workflow linter | ✅ 0 broken_links, 0 bloat_warnings, 0 sync_errors |

## 6. 메모리 layer

v0.5.4 의 모든 작업은 `ai-workflow/memory/release/v0.5.4/` 에 기록:
- `PROJECT_PROFILE.md` — v0.5.4 프로젝트 프로파일 (contract v1 링크)
- `session_handoff.md` — 세션 인계 (TASK-V054-001/002/003 work status, next actions, risks)
- `backlog/2026-06-07.md` — TASK 상세
- `state.json` — workflow 상태 캐시 (3개 in_progress + 2개 done)

## 7. 다음 단계 (v0.5.5 후보)

- [ ] contract v2 (스트리밍, 멀티 fan-out, observability) — §11 의 후보
- [ ] Mavis 측 위임 결정 로직에 §6 카탈로그 자동 enforce
- [ ] mini-coder-max / general agent 측 출력 스키마 가드
- [ ] Phase 11 Real-world Pilot 본격화 (Devhub_example + 다른 외부 저장소 + 케이스 스터디)
- [ ] PyPI 배포 (현재는 local pip install only)

## 8. v0.5.0 → v0.5.4 변경 통계

```
v0.5.0 (PR #13): 42/42 smoke + MiniMax Code harness overlay
v0.5.1 (PR #14): per-harness MCP install + auto-emit + guide
v0.5.1 (PR #15): check_bootstrap_mcp_roundtrip + 5 harnesses round-trip
v0.5.2 (PR #16): bootstrap 풀 리팩터 + 패키지화 + pilot validation
v0.5.3 (PR #17): antigravity MCP 표준화 + cross-language stack 표시
v0.5.4 (PR #18): orchestrator ↔ sub-agent contract v1 + maturity sync + baseline sync
```

총 6 PR.

## 9. 이슈 트래커

- ✅ [issue #1](https://github.com/ykylee/standard_ai_workflow/issues/1) 영구 해결 (2026-04-24 부터 open, v0.5.4 PR 머지 후 close)
