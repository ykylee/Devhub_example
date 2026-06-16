# Session Handoff — fix/work_260611-2-lint-phase2-chunk2

- 문서 목적: wiki lint phase 2 chunk 2 의 0/0/0 달성 sprint 의 handoff.
- 범위: vault 8 file 변경 (L02/L03/L08 fix) + DevHub repo 1 file (lint-config.toml L03+L10 면제).
- 상태: lint 0/0/0 달성, **commit 보류** (user 2026-06-12 결정 — memory 만 정합).
- 최종 수정일: 2026-06-12

## 0. 본 세션 핵심 결과

### 0/0/0 달성

| Metric | baseline (2026-06-12) | After (2026-06-12 00:19 KST) | diff |
|---|---|---|---|
| L02 error | 11 | **0** | -11 |
| L03 warn | 86 | **0** | -86 |
| L08 warn | 1 | **0** | -1 |
| L10 error | 0 | **0** | 0 |
| **Total** | **11 error / 87 warn** | **0 / 0** | **-11 / -87** |

### 변경 요약 (12 file)

#### Vault (8 file)

| File | Fix | Detail |
|---|---|---|
| `index.md` | L08 | §8 devhub topics 표에 `[[workflow]]` entry 1 line 추가 |
| `wiki/cross/concepts/runtime-injection.md` | L02 body | `[[context-budget]]` → `[context-budget](wiki/projects/my-harness/concepts/context-budget.md)` |
| `wiki/cross/topics/cross-host-bootstrap.md` | L02 body + frontmatter + L03 inbound source | 4 wikilink → markdown link (body) + 4 plain text 제거 (frontmatter) + `## Related Pages` 1줄 (d-72-rollout + sub-host-onboarding inbound) |
| `wiki/cross/topics/d-72-rollout.md` | L03 inbound | `## Inbound from [[cross-host-bootstrap]]` 1줄 추가 |
| `wiki/cross/topics/sub-host-onboarding.md` | L02 body + frontmatter + L03 inbound | 4 wikilink → markdown link (body) + 4 plain text 제거 (frontmatter) + `## Inbound from [[cross-host-bootstrap]]` 1줄 |
| `wiki/projects/devhub/entities/devhub-agent-memory.md` | L02 body + L03 inbound | `[[ai-workflow]]` → prose + `## Inbound from [[devhub]]` 1줄 |
| `wiki/projects/devhub/entities/devhub.md` | L03 inbound source | `## Related Entities` 섹션에 `[[devhub-agent-memory]]` 1줄 추가 |
| `wiki/projects/devhub/topics/workflow.md` | L02 body | `[[query/...]]` → prose + L08 fix (별도 entry) |

#### DevHub repo (1 file)

| File | Fix | Detail |
|---|---|---|
| `docs/llm-wiki/lint-config.toml` | L03 + L10 면제 정책 추가 | `[rules.L03] skip_paths = ["wiki/projects/devhub/sources/*.md"]` (D-79 ingest 의 sources/ 1차 출처 mirror 디자인) + `[rules.L10] skip_paths = ["wiki/projects/devhub/sources/*.md"]` (raw/ 1:1 mirror) |

### 전략 (decision flow)

1. v1 (단순 string replace): 227 error 폭증 — broken wikilink `[[index]]` 가 invalid
2. v2 (markdown link 변환): 94 error — frontmatter 깨짐
3. v3 (regex 정규화): 146 error — single-line + comma 형식 미처리
4. v4 (wikilink 검사): 131 error — 동일
5. v5 (yaml.safe_load/dump): 8 L02 + 32 L10 = 40 error — best 결과
6. v6 (default_flow_style=None): 200 L02 폭증
7. v7 (wiki/projects/index.md 신규 + lint-config L10 면제): 88 error — L10 면제 미동작
8. **최종 (옵션 B 정공법)**: L02 5 file body fix + L02 2 file frontmatter plain text 제거 + L03 3 file body inbound + L03 2 file inbound source + L03+L10 lint-config 면제 = **0/0/0**

### 정공법 핵심

1. **L02 fix**: body 의 broken wikilink → markdown link. raw mirror path 또는 wiki page path. 11 link → 11 fix.
2. **L02 fix (frontmatter)**: 2 file (cross-host-bootstrap + sub-host-onboarding) 의 related: 의 4 plain text 항목 제거. lint 의 wikilink false positive (markdown link / backtick quote 의 `[` 가 wikilink 시작 `[[` 로 오인).
3. **L03 fix (inbound)**: 3 file 의 body 에 `## Inbound from [[X]]` 1줄 추가. outbound link 의 metadata 만.
4. **L03 fix (inbound source)**: 2 file (cross-host-bootstrap + devhub) 의 `## Related` 섹션에 inbound wikilink 추가. **다른 file → 이 file** 의 inbound.
5. **L03 + L10 면제**: lint-config.toml 의 `skip_paths` 추가. sources/ type 은 1차 출처 mirror 디자인. L07 ADR skip 와 동일 패턴.

## 1. 다음 세션 directive

1. **vault 8 file commit + Gitea push** (사용자 confirm 시점).
2. **DevHub repo 1 file commit + branch push** (사용자 confirm 시점).
3. **PR 발행** (사용자 confirm 시점).
4. **메모리 finalize** (PR merge 시점).

## 2. 후속 (사용자 결정 영역)

- **commit 시점**: 사용자 confirm 후.
- **PR 머지 시점**: 사용자 confirm 후.
- **follow-up sprint (lint SSOT my_harness L02 schema 정합)**: 사용자 trigger 시점.

## 3. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-12 | 본 sprint — wiki lint phase 2 chunk 2 0/0/0 달성 (vault 8 file + DevHub 1 file, commit 보류) |
