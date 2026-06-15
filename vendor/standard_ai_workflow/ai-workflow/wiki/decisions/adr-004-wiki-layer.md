---
type: decision
status: accepted
adr_id: ADR-004
decided_at: 2026-06-12
accepted_in: v0.6.0
alternatives_considered: [memory-replacement, federated-wiki, crdt, docs-wiki, source-wiki]
related_pages: [concepts/project-architecture, patterns/r4-anchor-index, concepts/wiki-source-rule-r9, concepts/memory-3-state-lifecycle, patterns/frozen-archive-immutability]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# ADR-004: LLM Wiki Layer 도입

## Status

Accepted (v0.6.0, P1 implemented). Full spec at [ADR-004 file](../../docs/architecture/ADR-004-llm-wiki-layer.md).

## Context

Karpathy LLM Wiki (2026-04 gist) analysis showed that `ai-workflow/memory/` was excellent for workflow state but lacked knowledge accumulation/synthesis/reuse capability. A separate layer was needed.

## Decision

`ai-workflow/wiki/` created as a parallel Runtime layer (ADR-001 §3). Git-tracked. Ingest-first + release-freeze model. 7 rules (R1~R7) from the distributed rules plan.

## Consequences

Positive: knowledge accumulates across sessions, ~70x token efficiency vs RAG (within 500KB), existing infra reusable (bootstrap_lib, contract_v1, 6 harness overlays).
Negative: new components needed (SCHEMA, lint, ingest skills), `.gitignore` carve-out required.

## Alternatives Rejected

| Option | Reason |
|---|---|
| memory/ 대체 (방향 B) | 11 skills + 6 harnesses redesign needed |
| Federated Wiki (Ward Cunningham) | federation discovery didn't scale |
| CRDT (Yjs, Automerge) | tombstone overhead + semantic conflict blind |
| `docs/wiki/` (Project Layer) | ADR-001 layer boundary violation |
| `workflow-source/wiki/` (Source Layer) | Source is for templates only, not runtime state |

## v0.6.1+ Evolution — R8 Freeze + R9 Ingest Rule

ADR-004 는 v0.6.0 에서 `accepted` 되었으며, 그 후 도입된 **R8 (Memory Raw Freeze, v0.6.1)** 과 **R9 (Raw Source of Truth, v0.6.1.5)** 은 본 ADR 의 결정 자체를 supersede 하지 않는다. 두 규칙은 모두 ADR-004 가 세운 wiki layer 의 "source 가 어디인가" 와 "source 의 immutability 를 어떻게 보장하는가" 에 대한 **연속적인 refinement (extension)** 다. 따라서 본 ADR 의 `status: accepted` 는 유지된다.

**R8 (Memory Raw Freeze, v0.6.1)** — ADR-004 가 wiki 를 별도 Runtime layer 로 정의했지만, v0.6.0 시점의 wiki ingest 는 `memory/active/` 의 mutable raw 를 source 로 가정했다. R8 은 이 mutable raw 를 session-end 마다 `memory/archive/YYYY-MM-DD/` 로 copy + `.frozen` marker 부착하는 freeze protocol 을 도입했다. 핵심은 **COPY (NOT move)** — `active/` 는 다음 세션을 위해 보존되고, freeze 결과물인 `archive/` 만이 immutable raw 가 된다. 이 정책은 [[concepts/memory-3-state-lifecycle]] §2.2 와 §4 의 freeze mechanism 에 정리되어 있으며, `workflow-source/skills/memory-freeze/SKILL.md` 가 구현을 담당한다.

**R9 (Raw Source of Truth, v0.6.1.5)** — R8 이 immutability 경계를 만들었다면, R9 는 그 경계를 wiki layer 까지 강제한다. "wiki-ingest 는 `archive/` 만 source 로 사용한다. `active/` 는 절대 ingest 하지 않는다" 가 R9 의 한 줄 정의다 ([[concepts/wiki-source-rule-r9]] §Rule Definition). R8 freeze 가 완료된 `archive/YYYY-MM-DD/` 만이 wiki 의 raw source 자격을 얻고, mutable `active/` 는 ingest source 로 절대 사용되지 않는다. lint 는 `workflow-source/tests/check_wiki_source_rule.py` (V-R9) 가 담당하며, 위반 시 `AssertionError` 로 종료 코드 1 을 반환한다. codebase self-ingest (`workflow-source/`, `docs/`, `ai-workflow/wiki/`) 는 R9 면제 대상이다 ([[INGEST_GUIDE]] §1 참조).

이 두 규칙은 ADR-004 의 §Decision ("Git-tracked. Ingest-first + release-freeze model. 7 rules (R1~R7)") 가 명시한 "release-freeze model" 의 후속 구현 정밀화로 읽힌다. R1~R7 은 wiki 운영 layer 의 구조 규칙 (location, atomicity, sync, index, merge, topic branch, conflict resolution) 이었고, R8~R9 는 source layer (memory) 의 immutability 와 wiki-source binding 을 추가했다. 즉 ADR-004 의 layer separation 정신은 그대로 유지되면서, source-of-truth 의 품질 보장이 한 단계 더 강화된 것이다. ADR-005 (Memory as Raw Layer, proposed) 가 이 R8/R9 의 정식 채택을 기록할 예정이며, 본 ADR-004 는 그 후속 결정의 상위 spec 으로 남는다.

### §E.1 R8 Freeze — Detail  {#sE-1-r8-detail}

| 항목 | 값 |
|---|---|
| 도입 버전 | v0.6.1 (2026-06-12) |
| Trigger | session-end (D8) |
| Mechanism | `memory-freeze` skill (`scripts/run_memory_freeze.py`) |
| Source | `memory/active/` 의 5 파일 (`session_handoff.md`, `work_backlog.md`, `backlog/`, `state.json`, `log.md`) |
| Target | `memory/archive/YYYY-MM-DD/` (immutable copy) |
| Marker | `.frozen` YAML (`frozen_at`, `source`, `files` 키) |
| Lint | `check_memory_freeze_lint.py` (V-R10) — 5 파일 무결성 자동 검증 |
| Backward compat | **호환**. v0.6.0 까지 wiki 는 `active/` 를 source 로 가정했지만, R8 이후 wiki 는 `archive/` 로 자동 전환. 기존 ingest 사이클은 마이그레이션 불필요. |

### §E.2 R9 Ingest Rule — Detail  {#sE-2-r9-detail}

| 항목 | 값 |
|---|---|
| 도입 버전 | v0.6.1.5 (PATCH, 2026-06-12) |
| Scope | wiki-ingest 의 memory snapshot source 만 (codebase self-ingest 면제) |
| Source 한정 | `memory/archive/YYYY-MM-DD/`, `memory/release/v0.X.Y/` |
| Source 금지 | `memory/active/` (mutable, freeze 미완) |
| Lint | `check_wiki_source_rule.py` (V-R9) — 2 pattern 검사 |
| Lint 심각도 | error (위반 시 종료 코드 1) |
| Backward compat | **호환**. v0.6.0 wiki 가 `active/` ingest 한 산출물은 이미 stale 가능. R9 적용 시점부터 archive/ 로 재-ingest 필요할 수 있음. 자동 마이그레이션 스크립트 없음 (수동 재-ingest) |

### §E.3 What did NOT change  {#sE-3-not-changed}

R8/R9 extension 이 추가되면서도 ADR-004 가 정의한 다음 4 가지는 변하지 않는다.

| ADR-004 결정 | R8/R9 이후 상태 |
|---|---|
| `ai-workflow/wiki/` 는 Runtime layer (ADR-001 §3) | 변함 없음. R8/R9 는 source layer 만 정밀화 |
| Git-tracked | 변함 없음. wiki/ 의 `git log` audit 가능성 유지 |
| Ingest-first + release-freeze model | **강화됨**. freeze 가 source immutability 를, ingest 가 wiki update 를 담당하는 layering 이 더 명확해짐 |
| 7 rules (R1~R7) | R1~R7 그대로. R8~R9 는 별도 group (source immutability + wiki-source binding) 으로 추가. rule ID 보존 정책 (§1.2) 준수 |

### §E.4 Timeline  {#sE-4-timeline}

| Date | Version | Event | 영향 |
|---|---|---|---|
| 2026-06-12 | v0.6.0 | ADR-004 accepted (P1) | wiki/ 디렉토리 + R1~R7 + ingest-first model |
| 2026-06-12 | v0.6.1 | R8 (Memory Raw Freeze) 도입 | active → archive 트랜지션 + `.frozen` marker. ADR-004 의 "release-freeze" 가 session granularity 로 정밀화 |
| 2026-06-12 | v0.6.1.5 | R9 (Raw Source of Truth) 도입 | wiki-ingest source = `archive/` only. ADR-004 의 wiki 가 source immutability 를 엄격히 요구하도록 강화 |
| (planned) | v0.6.2+ | ADR-005 (Memory as Raw Layer) accepted 예정 | R8/R9 의 정식 채택 ADR. 본 ADR-004 의 후속 spec |

### §E.5 Cross-References in this Section  {#sE-5-cross-refs}

이 evolution 섹션에서 사용한 wiki 내부 cross-reference:

- [[concepts/wiki-source-rule-r9]] — R9 의 정의·lint·예외 (§Rule Definition, §Lint Enforcement, §Exceptions)
- [[concepts/memory-3-state-lifecycle]] — R8 freeze 가 active/archive/release 라이프사이클에 어떻게 매핑되는지 (§2.2, §3, §4)
- [[patterns/frozen-archive-immutability]] — R8 freeze 가 archive immutability 를 어떻게 보장하는지 (placeholder; v0.6.1+ 에서 별도 page 예정)

## References

- Full spec: [`../../docs/architecture/ADR-004-llm-wiki-layer.md`](../../docs/architecture/ADR-004-llm-wiki-layer.md) — 원본 결정의 전체 spec (v0.6.0 시점, P1 구현 완료)
- Wiki 운영 헌법: [`../SCHEMA.md`](../SCHEMA.md) §1.1 (decision frontmatter), §5.1 (R1~R7), §5.4 (P1~P4)
- R9 Ingest Rule 개념 페이지: [`../concepts/wiki-source-rule-r9.md`](../concepts/wiki-source-rule-r9.md) — R9 의 정의, lint, 예외 범위
- Memory 3-State Lifecycle: [`../concepts/memory-3-state-lifecycle.md`](../concepts/memory-3-state-lifecycle.md) — R8 freeze mechanism + active/archive/release 전이
- Frozen Archive Immutability 패턴: `[[patterns/frozen-archive-immutability]]` — R8 freeze 의 immutability 보장 패턴 (placeholder)
- Memory-Freeze skill: [`../../../workflow-source/skills/memory-freeze/SKILL.md`](../../../workflow-source/skills/memory-freeze/SKILL.md) — R8 freeze 구현 spec
- v0.6.1.5 release note: [`../../../workflow-source/releases/Beta-v0.6.1.5.md`](../../../workflow-source/releases/Beta-v0.6.1.5.md) — R9 도입 patch release
- 상위 design plan: [`.omo/plans/v0.6.1-plus-memory-raw-ops-design.md`](../../../.omo/plans/v0.6.1-plus-memory-raw-ops-design.md) §4 (R8/R9), §10 (ADR-005 draft)
