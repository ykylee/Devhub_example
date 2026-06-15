---
type: decision
status: proposed
adr_id: ADR-005
decided_at: 2026-06-12
accepted_in: v0.6.1.5
alternatives_considered: [active-as-source, both-active-and-archive, manual-selection, no-restriction]
related_pages: [concepts/wiki-source-rule-r9, concepts/memory-3-state-lifecycle, patterns/frozen-archive-immutability, decisions/adr-004-wiki-layer]
created: 2026-06-12
updated: 2026-06-12
r9_skip: true
---

# ADR-005: R9 — wiki-ingest source = `archive/` only

## Status

Proposed (R9 implemented in v0.6.1.5 PATCH on 2026-06-12, this ADR is the first formal record). 기존 release note ([`Beta-v0.6.1.5`](../../../workflow-source/releases/Beta-v0.6.1.5.md) 동등) 한 줄 note 로 등재됐던 R9 규칙을 ADR 형식으로 승격. 채택 확정 시 status 를 `accepted` 로 전환.

## Context

ADR-004 ([[decisions/adr-004-wiki-layer]]) 에서 `ai-workflow/wiki/` 를 knowledge accumulation layer 로 채택하고, ingest-first + release-freeze 모델을 도입했다. 그러나 v0.6.1.5 직전까지 wiki-ingest 의 source 는 명시적 제한이 없었다. [[concepts/memory-3-state-lifecycle]] 의 3-state 구조에서 mutable layer 인 `memory/active/` 도 ingest 후보였고, 그 결과 3가지 drift 가 관측됐다.

| Drift 종류 | 증상 | Root cause |
|---|---|---|
| **active hot-write 누출** | 같은 session 안에서 `active/` 에 쓴 hot patch 가 wiki 페이지에 반영돼, 다음 session 의 갱신이 wiki 와 어긋남 | mutable layer 가 immutable wiki 의 source |
| **미성숙 freeze 누출** | session 도중 R8 freeze 가 강제 발동돼 incomplete state 가 archive 로 들어가고, 그게 곧바로 wiki-ingest 에 사용됨 | freeze 완료 보장 없이 ingest trigger |
| **ingest 시점 비결정성** | active/ 가 free-form 으로 위키에 들어와, wiki 페이지의 `last_ingested_from` 메타가 매번 다르고 비교 불가 | ingest 의 source 가 무엇인지 lint 로 강제 불가 |

해결 방향: wiki 의 ingest source 를 R8 freeze 완료된 immutable layer (`memory/archive/YYYY-MM-DD/`) 로 한정. mutable → immutable transition (R8 freeze, [[concepts/memory-3-state-lifecycle]] §4) 의 결과물만 wiki layer 의 raw source 자격을 가지도록 강제한다. [[patterns/frozen-archive-immutability]] 가 freeze immutability 의 메커니즘이고, 본 ADR 은 그 immutability 가 wiki layer 까지 연장되도록 보장한다.

## Decision

R9 (Raw Source of Truth) 를 정식 규칙으로 채택한다. 원문 ([[workflow-source/MEMORY_GOVERNANCE.md]] §4 verbatim):

> **wiki ingest**: wiki-ingest 는 `archive/YYYY-MM-DD/` 만 source 로 사용 (R9). `active/` ingest 금지.

원문 ([[workflow-source/MEMORY_GOVERNANCE.md]] §4 immutability 항):

> **immutability**: freeze 후 archive/ 는 읽기 전용 raw source (R9). freeze 가 wiki-ingest 의 source 가 됨.

원문 ([[workflow-source/skills/memory-freeze/SKILL.md]] §1):

> freeze 후 archive/ 는 읽기 전용 raw source (R9). freeze 가 wiki-ingest 의 source 가 됨.

| 항목 | 값 |
|---|---|
| Rule ID | **R9** (Raw Source of Truth) |
| 허용 source | `memory/archive/YYYY-MM-DD/` (R8 freeze 완료 + `.frozen` 마커), `memory/release/v0.X.Y/` (deep freeze, archive 동급) |
| 금지 source | `memory/active/` (mutable, R8 freeze 미완) |
| 면제 scope | codebase self-ingest (`workflow-source/`, `docs/`, `ai-workflow/wiki/`) — [[INGEST_GUIDE]] §1 참조. 본 ingest 는 git-tracked + PR-reviewed 라 R9 의 immutability boundary 목적과 무관 |
| Lint | **V-R9** — `workflow-source/tests/check_wiki_source_rule.py` |
| Lint 심각도 | error (위반 시 `AssertionError` + exit 1) |
| 도입 버전 | v0.6.1.5 (PATCH, 2026-06-12) |
| 상위 결정 | [[decisions/adr-004-wiki-layer]] (wiki layer), [[concepts/memory-3-state-lifecycle]] (3-state) |

### V-R9 Lint 검사 패턴

| 검사 함수 | Regex 패턴 | 실패 메시지 |
|---|---|---|
| `check_wiki_specs_no_active_ref` | `ingest.*active[/\s] \| active/.*ingest \| source.*active/` (case-insensitive) | `[V-R9] <wiki-path>: references active/ as ingest source (must use archive/ only)` |
| `check_active_files_not_in_wiki` | `ingest.*memory/active \| memory/active.*ingest \| source.*memory/active \| memory/active.*source` (case-insensitive) | `[V-R9] <wiki-path>: references memory/active/ as ingest source (must use archive/ only)` |

scope: `ai-workflow/wiki/**/*.md` 전체. R8 freeze 자체의 무결성은 `check_memory_freeze_lint.py` (V-R10) 가 별도 검사. V-R9 는 **wiki layer 의 마크다운이 `active/` 를 ingest/source 로 언급하는지** 만 본다.

### Codebase self-ingest 예외 ([[INGEST_GUIDE]] §1)

| Ingest 종류 | source | R9 적용 | 사유 |
|---|---|---|---|
| Memory snapshot ingest | `archive/YYYY-MM-DD/` → wiki | YES | mutable → immutable 전이 ([[concepts/memory-3-state-lifecycle]] §3) |
| Codebase self-ingest | `workflow-source/`, `docs/`, `ai-workflow/wiki/` → wiki | NO (R9 면제) | 모두 git-tracked + PR-reviewed. `active/` 미사용. wiki page 가 다음 ingest 의 raw source 가 됨 |
| Release snapshot ingest | `memory/release/v0.X.Y/` → wiki | YES (archive 동급) | deep freeze, immutability 동급. 별도 위변조 검사 불필요 |

## Consequences

### Positive

| # | 결과 | 근거 |
|---|---|---|
| 1 | **Immutability boundary 형성** | mutable raw (`active/`) 와 immutable knowledge (wiki) 가 R9 + R8 freeze 양쪽에서 단절. mutable 변경이 wiki 를 흔들 수 없음 |
| 2 | **Drift 제거** | hot-write 누출, 미성숙 freeze 누출, ingest 시점 비결정성 3종 모두 차단 ([[concepts/wiki-source-rule-r9]] §Why archive/ only 참조) |
| 3 | **Mature freeze 강제** | wiki-ingest 를 시도하려면 R8 freeze + `.frozen` marker 가 선행돼야 함. freeze 가 lightweight suggestion 이 아닌 hard prerequisite 으로 격상 |
| 4 | **Lint 자동 검증 (V-R9)** | wiki 페이지가 `active/` 를 ingest/source 로 언급하면 `AssertionError` + exit 1. wiki layer 만 보면 R9 준수 여부 즉시 확인 |
| 5 | **Audit 단순화** | wiki 페이지의 `last_ingested_from` 메타가 항상 `archive/YYYY-MM-DD/` 또는 codebase 경로 중 하나. provenance 추적 결정적 |
| 6 | **Release snapshot 재사용** | `memory/release/v0.X.Y/` 도 archive 동급 immutability 로 인정. 별도 위변조 검사 코드 불필요 |

### Negative

| # | Trade-off | 완화책 |
|---|---|---|
| 1 | **Ingest latency** | session 종료 시점부터 wiki 반영까지 최소 1일 (다음 session 시작 시 ingest). 실시간 wiki 반영 안 됨 | R8 freeze 즉시 trigger + 다음 session 시작 시 ingest 우선순위 최상위 ([[concepts/memory-3-state-lifecycle]] §4.1 step 5 무결성 검증 후 진행) |
| 2 | **Mid-session knowledge 손실** | session 진행 중 wiki 검색이 필요한데 archive 가 아직 freeze 안 됐으면 새 정보가 wiki 에 없음 | session-start 시 handoff 의 `open questions` 섹션을 우선 query source 로 사용 (volatile but explicit). freeze 후 다음 session 에서 정식 wiki 반영 |
| 3 | **Codebase self-ingest audit 부담** | R9 면제 scope 인 codebase ingest 도 누가 어떤 시점에 어떤 commit 으로 wiki 갱신했는지 추적 필요 | `last_ingested_from` + `related_pages` + ingest commit message 의 page 목록 ([[INGEST_GUIDE]] §6) 으로 line-by-line audit 가능 |
| 4 | **V-R9 lint false positive 가능성** | wiki 페이지 본문에서 `active/` 를 메타 설명용으로 언급하는 경우 (예: "active/ 는 ingest 하지 않는다") 도 패턴에 매칭될 수 있음 | 본 ADR 자체가 R9 정의를 설명하며 `active/ ingest 금지` 를 적시. 추후 lint 를 `code block` 또는 `frontmatter` 로 scope 제한하는 정밀화 검토 |

## Alternatives Rejected

| Option | 이유 |
|---|---|
| **active/ as source** | mutable. session 진행 중 hot-write 가 wiki 에 반영돼 drift 발생. 다음 session 갱신 시 wiki 와 active 가 어긋남. 3종 drift (hot-write, 미성숙 freeze, ingest 비결정성) 모두 재발 |
| **both active/ and archive/** | mixed source. 위키 페이지마다 어떤 snapshot 이 source 인지 모호. `last_ingested_from` 메타 신뢰 불가. lint 하나로 강제 불가 (조건 분기 폭증) |
| **manual selection** (ingest 시점에 사람/agent 가 source 고름) | human error / agent memory 편향. 매 ingest 마다 명시적 선택이 필요해 자동 ingest 가 깨짐. R3 (Pull-Before-Push) 동기화 흐름과 충돌 |
| **no restriction** | chaos. `active/` 의 hot patch 가 wiki 에 들어가 wiki 가 매 ingest 마다 다름. R8 freeze 의 immutability 가 wiki layer 에서 무너짐. lint 부재로 회귀 감지 불가 |

## Related Decisions

- [[concepts/wiki-source-rule-r9]] — R9 정의 / V-R9 lint / 면제 scope 본문 (R9 의 concept 페이지)
- [[concepts/memory-3-state-lifecycle]] — active → archive → release 의 3-state immutability 계층
- [[patterns/frozen-archive-immutability]] — R8 freeze 가 archive immutability 를 어떻게 보장하는가
- [[decisions/adr-004-wiki-layer]] — wiki layer 도입 결정. 본 ADR 은 그 layer 의 source 제한을 정의
- [[INGEST_GUIDE]] — codebase self-ingest 절차 (R9 면제 대상) + §1 "왜 R9 가 본 ingest 에는 적용되지 않는가"

## References

- R9 원본 정의: `workflow-source/MEMORY_GOVERNANCE.md` §4 (Freeze 규칙 마지막 항목)
- R8 freeze skill: `workflow-source/skills/memory-freeze/SKILL.md` §1
- V-R9 lint: `workflow-source/tests/check_wiki_source_rule.py` (v0.6.1.5 신규, +50 lines)
- 릴리스 노트: `workflow-source/releases/Beta-v0.6.1.5.md` (R9 + V-R9 동시 도입)
- 상위 plan: `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` §4 R8/R9
- Wiki 운영 헌법: `ai-workflow/wiki/SCHEMA.md` §5.1 Rules
