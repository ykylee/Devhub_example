---
type: concept
status: active
last_ingested_from: workflow-source/MEMORY_GOVERNANCE.md + .omo/plans/v0.6.1-plus-memory-raw-ops-design.md
related_pages: [concepts/project-architecture, concepts/wiki-source-rule-r9, patterns/frozen-archive-immutability, decisions/adr-005-r9-wiki-source-rule]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# Memory 3-State Lifecycle (v0.6.1+)

- 문서 목적: `ai-workflow/memory/` 의 3-state lifecycle (active ↔ archive ↔ release) 정책, R8 freeze 메커니즘, R9/R10 부수 규칙을 정리한다.
- 범위: state 정의, lifecycle transition, R8 freeze protocol, related decisions
- 관련 결정: ADR-005 (Memory as Raw Layer, proposed)
- 최종 수정일: 2026-06-12

## §1 TL;DR  {#s1-tldr}

| State | 위치 | Mutability | Lifecycle Position |
|---|---|---|---|
| **active** | `memory/active/` | mutable (session write) | session start → end |
| **archive** | `memory/archive/YYYY-MM-DD/` | immutable (R8 freeze) | session end → release |
| **release** | `memory/release/v0.X.Y/` | immutable (deep freeze) | release time → forever |

R8 freeze = `active` → `archive` 트랜지션. Release 시점 deep freeze = `archive` 전체 → `release/`. Karpathy LLM Wiki 의 3-Layer (raw/wiki/SCHEMA) 와 1:1 매핑.

## §2 State Definitions  {#s2-state-definitions}

### §2.1 active  {#s2-1-active}

| 필드 | 값 |
|---|---|
| 위치 | `ai-workflow/memory/active/` |
| 추적 | gitignore (Runtime layer, ADR-001) |
| Owner | 현재 session 의 AI agent |
| Mutability | mutable. session 진행 중 자유 read/write |
| 내용 | `session_handoff.md`, `work_backlog.md`, `backlog/`, `state.json`, `log.md` |
| Lifecycle | session start ~ end. 종료 시 R8 freeze 로 archive 로 copy |

### §2.2 archive  {#s2-archive}

| 필드 | 값 |
|---|---|
| 위치 | `ai-workflow/memory/archive/YYYY-MM-DD/` |
| 추적 | gitignore. 단 `.frozen` 마커로 freeze 시점 audit 가능 |
| Owner | freeze 완료 후 owner 없음 (immutable raw) |
| Mutability | immutable. R8 freeze 후 read-only |
| 내용 | active/ 의 5개 파일 copy + `.frozen` YAML 마커 |
| Lifecycle | session end 시점 ~ release 시점. release 시 deep freeze 로 release/ 로 이동 |

### §2.3 release  {#s2-3-release}

| 필드 | 값 |
|---|---|
| 위치 | `ai-workflow/memory/release/v0.X.Y/` |
| 추적 | gitignore. 단 MANIFEST.json 으로 release provenance 기록 |
| Owner | release manager / project maintainer |
| Mutability | immutable. deep freeze snapshot. 영구 보관 |
| 내용 | release 시점의 `archive/` 전체 + `wiki/` 전체 + `MANIFEST.json` |
| Lifecycle | release time 부터 영구. 다음 release 시 별도 디렉토리 생성 |

## §3 Lifecycle Transitions  {#s3-lifecycle-transitions}

| From | To | Trigger | Mechanism | Rule |
|---|---|---|---|---|
| **active** | **archive** | session-end (D8) | `memory-freeze` skill (R8) | R8 + R10 |
| **archive** | **release** | release time | release 절차의 deep freeze snapshot | ADR-005 §Release |

Active → archive 는 매 session 마다 발생. Archive → release 는 release cycle 시점에만 발생 (수개월~1년 단위).

## §4 R8 Freeze Mechanism  {#s4-r8-freeze-mechanism}

`workflow-source/skills/memory-freeze/SKILL.md` 가 구현. `scripts/run_memory_freeze.py` 호출.

| Step | 동작 | 비고 |
|---|---|---|
| 1 | `archive/YYYY-MM-DD/` 디렉토리 mkdir | atomic 준비 |
| 2 | active/ 내 `.md` `.json` 파일을 `archive/YYYY-MM-DD/` 로 **COPY** | NEVER move (다음 세션 active/ 보존) |
| 3 | `.frozen` YAML 마커 작성 | `frozen_at`, `source`, `files` 키 |
| 4 | `memory/log.md` 에 `## [YYYY-MM-DD] freeze \| <summary>` append | D5. append-only |
| 5 | 5 파일 무결성 검증 (R10) | 통과 시에만 freeze 확정 |

### §4.1 핵심 정책  {#s4-1-key-policies}

| 정책 | 의미 | 근거 |
|---|---|---|
| **COPY (NOT move)** | freeze 후 active/ 파일 유지. 다음 세션에서 재사용 | D9 (per-session granularity). active 비우면 다음 세션 cold start |
| **`.frozen` YAML 마커** | `frozen_at: <ISO>`, `source: active/`, `files: [<list>]` | R8 audit. 동일 날짜 중복 freeze 감지 |
| **Same-date freeze = skip** | 이미 `archive/YYYY-MM-DD/` 가 존재하면 skip. status = `skipped` | immutability 보존. R10 위반 방지 |
| **Active → archive 만 허용** | archive → active 역방향 금지 (rollback 안 됨) | R7. 잘못된 갱신은 next session 에 active/ 에서 |

## §5 Related Decisions  {#s5-related-decisions}

- [[concepts/project-architecture]] — 3-Layer separation (Source/Runtime/Project Docs). memory 는 Runtime layer.
- [[concepts/wiki-source-rule-r9]] — wiki-ingest source = `archive/...` only. active/ 절대 ingest 안 함.
- [[patterns/frozen-archive-immutability]] — freeze 후 read-only 강제. lint (R10) 가 무결성 자동 검증.
- [[decisions/adr-005-r9-wiki-source-rule]] — ADR-005 (Memory as Raw Layer, proposed). 3-state lifecycle 의 정식 채택 결정.

## §6 References  {#s6-references}

- [Memory Governance 원문](../../../workflow-source/MEMORY_GOVERNANCE.md) — §4 Freeze 규칙
- [Memory-Freeze skill](../../../workflow-source/skills/memory-freeze/SKILL.md) — R8 freeze 구현 spec
- [v0.6.1+ Memory Raw Ops design](../../../.omo/plans/v0.6.1-plus-memory-raw-ops-design.md) — §3 architecture, §4 R8~R10, §10 ADR-005 draft
- [Memory Operation Log](../../memory/log.md) — 실제 freeze 이벤트 기록 예시
- [Project Architecture concept](./project-architecture.md) — 3-Layer + Memory 3-State 요약
