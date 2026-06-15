---
type: concept
status: active
last_ingested_from: workflow-source/MEMORY_GOVERNANCE.md §4 + workflow-source/releases/Beta-v0.6.1.5.md
related_pages: [concepts/memory-3-state-lifecycle, concepts/project-architecture, decisions/adr-005-r9-wiki-source-rule, patterns/frozen-archive-immutability, INGEST_GUIDE]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# R9 Rule: wiki-ingest source = `archive/` only

## TL;DR

| # | 항목 | 값 |
|---|---|---|
| 1 | Rule ID | **R9** (Raw Source of Truth) |
| 2 | 도입 버전 | v0.6.1.5 (PATCH, 2026-06-12) |
| 3 | Source 제한 | wiki-ingest/memory-ingest 는 `memory/archive/YYYY-MM-DD/` 만 사용 |
| 4 | 금지 source | `memory/active/` (mutable, freeze 미완) |
| 5 | 면제 범위 | codebase self-ingest (`workflow-source/`, `docs/`, `ai-workflow/wiki/`) — [[INGEST_GUIDE]] §1 참조 |
| 6 | Lint | **V-R9** — `workflow-source/tests/check_wiki_source_rule.py` |
| 7 | Lint 심각도 | error (위반 시 `AssertionError` + 종료 코드 1) |
| 8 | 관련 ADR | [[decisions/adr-005-r9-wiki-source-rule]] (proposed) |

## Rule Definition

R9 (v0.6.1.5): **"wiki-ingest 는 `archive/` 만 source 로 사용한다. `active/` 는 절대 ingest 하지 않는다."**

원문 (`workflow-source/MEMORY_GOVERNANCE.md` §4, R8 freeze 규칙 내 마지막 항목):

> **wiki ingest**: wiki-ingest 는 `archive/YYYY-MM-DD/` 만 source 로 사용 (R9). `active/` ingest 금지.

원문 (`workflow-source/skills/memory-freeze/SKILL.md` §1):

> freeze 후 archive/ 는 읽기 전용 raw source (R9). freeze 가 wiki-ingest 의 source 가 됨.

근거: mutable raw state (`active/`) 가 wiki 의 immutable source 가 되는 것을 방지하기 위한 **immutability boundary**. wiki 가 `active/` 의 hot-write 변경에 따라 흔들리거나, 미완성 freeze 가 wiki 로 흘러드는 것을 차단한다. R8 (Memory Raw Freeze) 가 mutable → immutable 전환을 정의하고, R9 는 그 immutable 결과물만 ingest source 로 인정한다.

## Why archive/ only

| 후보 source | Mutability | R9 적용 | 이유 |
|---|---|---|---|
| `memory/active/` | mutable (session write) | ❌ 금지 | freeze 미완, 다음 세션에서 덮어쓰기 가능 |
| `memory/archive/YYYY-MM-DD/` | immutable (R8 freeze + `.frozen` marker) | ✅ 허용 | R8 freeze 가 완료된 읽기 전용 raw |
| `memory/release/v0.X.Y/` | immutable (release snapshot) | ✅ 허용 | archive 와 동급 immutability |
| `workflow-source/`, `docs/`, `ai-workflow/wiki/` | git-tracked, PR-reviewed | ❌ R9 면제 | codebase self-ingest. 별도 가이드 ([[INGEST_GUIDE]] §1) |

`archive/` 가 wiki-ingest source 가 되려면 다음 두 조건이 모두 충족돼야 한다.

- **R8 freeze 완료**: `archive/YYYY-MM-DD/` 디렉토리가 존재하고 `.frozen` YAML marker 가 작성됨
- **immutability 보존**: 동일 날짜 중복 freeze 시 skip 처리 (기존 freeze 가 수정되지 않음)

이 두 조건이 [[patterns/frozen-archive-immutability]] 의 핵심이고, R9 는 그 immutability 가 wiki layer 까지 유지되도록 강제한다. `active/` 는 mutable 이므로 freeze 가 적용되지 않으며, 따라서 wiki 의 source 자격이 없다. 라이프사이클 전체 흐름은 [[concepts/memory-3-state-lifecycle]] 참조.

## Lint Enforcement

**V-R9** — `workflow-source/tests/check_wiki_source_rule.py` (v0.6.1.5 신규, +50 lines).

| 검사 | 패턴 | 실패 메시지 |
|---|---|---|
| `check_wiki_specs_no_active_ref` | `ingest.*active[/\s]` / `active/.*ingest` / `source.*active/` (대소문자 무시) | `[V-R9] <wiki-path>: references active/ as ingest source (must use archive/ only)` |
| `check_active_files_not_in_wiki` | `ingest.*memory/active` / `memory/active.*ingest` / `source.*memory/active` / `memory/active.*source` | `[V-R9] <wiki-path>: references memory/active/ as ingest source (must use archive/ only)` |

scope: `ai-workflow/wiki/**/*.md` 의 모든 마크다운. R8 freeze 자체는 `check_memory_freeze_lint.py` (V-R10) 가 별도 검사. V-R9 는 **wiki layer 가 active/ 를 언급하지 않는지** 만 본다.

```bash
$ python3 workflow-source/tests/check_wiki_source_rule.py
V-R9 check passed: no active/ references as wiki-ingest source.
```

## Exceptions

R9 는 **memory snapshot 의 wiki-ingest** 에만 적용된다. codebase 자체를 runtime wiki 로 ingest 하는 작업은 R9 scope 밖이다.

| 대상 | R9 적용 | 근거 |
|---|---|---|
| Memory snapshot ingest (`archive/YYYY-MM-DD/` → wiki) | YES | [[concepts/memory-3-state-lifecycle]] 의 mutable → immutable 전이 |
| Codebase self-ingest (`workflow-source/`, `docs/`, `ai-workflow/` → wiki) | NO | 모두 git-tracked + PR-reviewed 또는 manual curation. `active/` 미사용. ingest 산출물이 다음 ingest 의 raw source 가 됨 |
| Release snapshot ingest (`memory/release/v0.X.Y/` → wiki) | YES (R9 준거) | archive 와 동급 immutability. 별도 위변조 검사 없음 |

자세한 사유와 phase 계획: [[INGEST_GUIDE]] §1 ("왜 R9 가 본 ingest 에는 적용되지 않는가").

## Related Decisions

- [[decisions/adr-005-r9-wiki-source-rule]] — R9 도입 ADR (proposed, v0.6.1.5)
- [[concepts/project-architecture]] — 3-Layer + LLM Wiki + Memory 3-State 의 source/raw/wiki 분리 정신
- [[concepts/memory-3-state-lifecycle]] — `active/` → `archive/` → `release/` immutability 계층
- [[patterns/frozen-archive-immutability]] — R8 freeze 가 archive immutability 를 어떻게 보장하는가
- [[INGEST_GUIDE]] — codebase self-ingest 절차 (R9 면제 대상)

## References

- R9 원본 정의: [`../../workflow-source/MEMORY_GOVERNANCE.md`](../../workflow-source/MEMORY_GOVERNANCE.md) §4 (Freeze 규칙)
- R8 freeze skill: [`../../workflow-source/skills/memory-freeze/SKILL.md`](../../workflow-source/skills/memory-freeze/SKILL.md)
- V-R9 lint: [`../../workflow-source/tests/check_wiki_source_rule.py`](../../workflow-source/tests/check_wiki_source_rule.py)
- 릴리스 노트: [`../../workflow-source/releases/Beta-v0.6.1.5.md`](../../workflow-source/releases/Beta-v0.6.1.5.md)
- 상위 메모리 raw ops plan: [`.omo/plans/v0.6.1-plus-memory-raw-ops-design.md`](../../.omo/plans/v0.6.1-plus-memory-raw-ops-design.md) §4 R8/R9
- Wiki 운영 헌법: [`../SCHEMA.md`](../SCHEMA.md) §5.1 Rules
