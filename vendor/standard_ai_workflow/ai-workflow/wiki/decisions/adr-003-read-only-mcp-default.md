---
type: decision
status: accepted
adr_id: ADR-003
decided_at: 2026-05-05
accepted_in: v0.5.5
alternatives_considered: [rw-by-default, opt-in-readonly, capability-declaration, scaffold-only]
related_pages: [concepts/mcp-transport, concepts/harness-distribution, entities/mcp-read-only-bundle]
created: 2026-06-12
updated: 2026-06-12
---

# ADR-003: Read-only MCP 우선 정책

## Status

Accepted (v0.5.5, 2026-05-05). 6+1 MCP 서버 + 두 transport (`jsonrpc-bridge` default, `stdio-sdk` experimental) 대상. 전체 명세: [ADR-003 원본](../../docs/architecture/ADR-003-read-only-mcp-default-policy.md).

## Context

`standard_ai_workflow` 의 MCP 서버는 두 방향으로 발전 가능했다.

| 방향 | 자유도 | 위험 |
|---|---|---|
| 양방향 (read-write) | 높음 | sub-agent / 외부 시스템에 write 명령 전달 가능 |
| 읽기 전용 (read-only) | 낮음 | 외부 상태 조회만 가능. write 는 orchestrator 측 정책 |

v0.5.0 부터 read-only MCP bundle 이 사실상 default 였지만, 후속 릴리스에서 write 권한을 추가해야 하는지에 대한 정식 정책 결정이 없었다. 핵심 질문은 **agent 자율성 vs 안전성** 사이의 균형이었다.

write 가능한 MCP 서버는 세 문제를 동반한다.

1. **Blast radius 증가**: sub-agent 가 손상된 MCP 응답을 보내면 사용자 데이터에 직접 영향. read-only 면 정보 제공만 가능하므로 변형 위험이 orchestrator 측에 격리됨.
2. **Rollback 어려움**: write 가 orchestrator 가 아닌 MCP 안에서 일어나면 audit / undo 추적성이 약해짐.
3. **Security review 비용 증가**: write surface 가 늘수록 권한 모델 / 영향 범위 / sandbox 정책 검토가 가중됨.

## Decision

**모든 MCP 서버는 default read-only** 이다. write 작업은 orchestrator 측 정책으로 명시적으로 처리한다.

| # | 결정 | 근거 |
|---|---|---|
| 1 | MCP 서버 6+1 종 default = read-only | `latest_backlog`, `check_doc_links`, `check_doc_metadata`, `suggest_impacted_docs`, `check_quickstart_stale_links`, `create_backlog_entry` (v0.5.0~) + `read_only_mcp_sdk` (v0.5.7 candidate) |
| 2 | tool descriptor 에 `readOnlyHint=true` (MCP 2025-06-18 spec) 또는 equivalent 명시 | 하네스 측이 read-only 도구임을 descriptor 단계에서 인지 |
| 3 | `create_backlog_entry` 는 의도적 예외 ("draft 생성기") | "backlog entry draft" 만 생성. 실제 `ai-workflow/memory/backlog/` 에 write 안 함. orchestrator 자동 commit 금지. 사용자 직접 검토 후 commit |
| 4 | Transport 우선순위: `jsonrpc-bridge` (default, stable) > `stdio-sdk` (experimental) | `jsonrpc-bridge` 는 v0.5.0 부터 `tools/list` + `tools/call` round-trip 안정. `stdio-sdk` 는 정식 `mcp[cli]>=1.0` 호환이나 `Connection closed` 회귀 존재 |
| 5 | Bootstrap 시 MCP 자동 emit: `--enable-mcp` 플래그로 하네스별 config snippet 생성 | `.codex/mcp.toml`, `mcp.opencode.json`, `.gemini/mcp.json`, `.antigravity.mcp.json`, `.MiniMax/mcp.json` |
| 6 | `transport_ready=false` 면 manual review only (자동 적용 금지) | v0.5.10 정책. 자동 적용 차단 |

Transport 정식 default 전환 기준: [`workflow-source/core/read_only_mcp_transport_promotion.md`](../../workflow-source/core/read_only_mcp_transport_promotion.md) 참조. 두 transport 의 상세 비교 / failure mode / wiki layer 와의 상호작용은 [[concepts/mcp-transport]] 에 정리. 하네스별 bootstrap 시 MCP emit 흐름은 [[concepts/harness-distribution]] 참조. 노출되는 read-only 도구 4종의 실체는 [[entities/mcp-read-only-bundle]] 에 entity 로 기록.

### §3.1 노출 read-only 도구 (7 종)

| Tool | 역할 | 등장 버전 |
|---|---|---|
| `latest_backlog` | 최근 백로그 인덱스 조회 | v0.5.0 |
| `check_doc_links` | 문서 내부 link 검증 | v0.5.0 |
| `check_doc_metadata` | frontmatter / 메타 검증 | v0.5.0 |
| `suggest_impacted_docs` | 변경 영향 문서 후보 추천 | v0.5.0 |
| `check_quickstart_stale_links` | quickstart 문서 stale link 검사 | v0.5.0 |
| `create_backlog_entry` | backlog draft 생성 (write 예외) | v0.5.0 |
| `read_only_mcp_sdk` | 정식 `mcp[cli]>=1.0` 호환 entrypoint | v0.5.7 candidate |

### §3.2 Transport 우선순위

| 우선순위 | Transport | 상태 | 선택 기준 |
|---|---|---|---|
| 1 (default) | `jsonrpc-bridge` | stable (v0.5.0) | 일반 도입 / smoke test. `tools/list` + `tools/call` round-trip 안정 |
| 2 (opt-in) | `stdio-sdk` | experimental | 정식 MCP SDK 호환이 필요하고 회귀를 감수할 수 있는 경우 |

### §3.3 `create_backlog_entry` 예외 정책

v0.5.0 부터 read-only 우선 정책의 **유일한 의도적 write tool** 이지만, 동작은 "write 도구" 가 아니라 **"draft 생성기"** 다.

- 출력: `BacklogDraft` 객체 (제목, 본문, 메타) 만 반환
- 실제 write 대상: `ai-workflow/memory/backlog/` 에 **저장하지 않음**
- 자동 commit: **orchestrator 가 수행하지 않음**
- 사용자 흐름: draft 를 받아 검토 → 수동으로 backlog 파일 생성 → 직접 commit

이 분리는 ADR-001 의 3-layer 정신 (state 는 read 가능, write 는 명시적 사용자 결정) 과 일치한다.

## Consequences

### Positive

- **Blast radius 축소**: read-only 면 sub-agent 가 손상된 MCP 응답을 보내도 사용자 데이터에 직접 영향 X. 변형 결정은 orchestrator 측에 격리됨.
- **Audit / undo 용이성**: write 가 orchestrator 정책으로 명시 → 사용자 / operator 가 쉽게 추적 / 취소.
- **승격 경로 명확**: read-only 가 default 라 stdio-sdk 의 `Connection closed` 회귀가 전체 시스템에 영향 X. opt-in 으로 격리.
- **검토 비용 절감**: write surface 가 작아 security review / 권한 모델 검토 부담 감소.
- **Onboarding 속도**: read-only bundle 만 검증하면 되므로 신규 하네스 / 신규 프로젝트 도입 시 smoke test 빠르게 통과.
- **Side-effect free testing**: read-only 도구는 멱등 → dry-run / 회귀 test 작성 용이.

### Negative / Trade-offs

- **자유도 감소**: 일부 use case (e.g. autonomous task scheduling) 는 read-only 부족. 완화: orchestrator 측 write 정책으로 동일 효과 달성 가능.
- **`create_backlog_entry` 의 예외가 정책 복잡도**: 단일 write tool 이 있어 ADR 본문이 모순처럼 보일 수 있음. 완화: "draft 생성기" 로 위치 재명시. 실제 write 는 사용자.
- **Opt-in 문서화 부담**: write 가 필요한 orchestrator / sub-agent 는 명시적 opt-in 정책과 문서를 별도로 유지해야 함.
- **stdio-sdk 회귀 격리**: 회귀 발견 시 stdio-sdk 가 promotion 기준을 통과하지 못해 default 가 안 됨. v0.6+ 후속 작업으로 default 전환 가능.

### 후속 결정의 인용

| 버전 | 변경 |
|---|---|
| v0.5.7 | `read_only_mcp_sdk` 가 정식 SDK 호환 candidate 로 추가. `Connection closed` 회귀 fix 시 default 전환 가능 |
| v0.5.8 | bootstrap 시 MCP config 자동 emit + read_only tool descriptor 자동 생성 |
| v0.5.10 | `transport_ready=false` 시 manual review only 정책. 자동 적용 금지 |

## Alternatives Rejected

| Option | Reason |
|---|---|
| **rw-by-default** (write 가능 default) | blast radius 증가 / rollback 어려움 / security review 비용 가중. 손상된 MCP 응답이 사용자 데이터에 직접 영향 |
| **opt-in-readonly** (default write, 사용자 opt-in 으로 read-only 강제) | default 가 자유로워 안전성 확보 실패. write 권한 보유 자체가 위험. opt-in 절차가 운영 부담 |
| **capability-declaration** (도구별 capability manifest + 서명 / verify) | manifest 작성 + 검증 overhead. 정책 코드 분리. read-only 기본 정책이면 capability 자체가 단순 |
| **scaffold-only** (코드/문서 스캐폴딩만 노출, 상태 정보 차단) | read-only 의 핵심 가치인 "외부 상태 조회" 가 사라짐. `latest_backlog` 같은 workflow state 조회 불가. ADR-001 의 3-layer 분리 정신 (state 는 read 가능) 과 충돌 |

## References

- 공식 ADR-003: [docs/architecture/ADR-003-read-only-mcp-default-policy.md](../../../docs/architecture/ADR-003-read-only-mcp-default-policy.md)
- Transport 승격 기준: [workflow-source/core/read_only_mcp_transport_promotion.md](../../../workflow-source/core/read_only_mcp_transport_promotion.md)
- 하네스별 MCP config: [workflow-source/examples/mcp_config_examples/](../../../workflow-source/examples/mcp_config_examples/) (5종)
- `workflow_kit/server/` — read-only MCP server 구현 (`python3 -m workflow_kit.server` 진입점)
- `mcp[cli]>=1.0` Python SDK (v0.5.7 candidate) — 회귀 fix 시 default 전환
- [[concepts/mcp-transport]] — 두 transport 비교 / failure mode
- [[concepts/harness-distribution]] — 하네스별 MCP config bootstrap
- [[entities/mcp-read-only-bundle]] — 6+1 read-only 도구 entity
- ADR-001 (3-layer 분리) — state 레이어 read 가능 정신의 근거
- Beta v0.5.7 release note — read-only SDK candidate 추가
- Beta v0.5.8 release note — bootstrap 시 MCP config 자동 emit
- Beta v0.5.10 release note — transport_ready false 시 manual review only 정책
