---
type: pattern
status: active
used_in: [ai-workflow/wiki/_lint, workflow-source/tests/check_wiki_antipatterns.py]
related_components: [concepts/project-architecture, SCHEMA §4]
created: 2026-06-12
updated: 2026-06-12
---

# 90-Day Staleness Lint (L05)

## TL;DR

| 항목 | 값 |
|---|---|
| Lint ID | L05 (`SCHEMA §4` item 2) |
| Threshold | `updated` 가 90일 이전 |
| Severity | `warning` (5%+ 누적 시), 그 외 `info` |
| Flagged pages | `entities/`, `concepts/`, `decisions/`, `patterns/` 하위 모든 페이지 |
| Excluded | `queries/` (ephemeral), P1~P4 ingest 미완료 stub |
| Automation | `check_wiki_antipatterns.py` (planned) + `ai-workflow/wiki/_lint/` (planned) |
| Source | `[[decisions/adr-004-wiki-layer]]`, [[SCHEMA §4](#)] |

## Problem

지식 페이지의 `updated` 필드는 마지막 ingest 시점의 메타데이터다. 시간 경과 시:

- 소스 코드 / 문서가 바뀌었는데 wiki 가 갱신되지 않으면 reader 가 outdated 정보를 인용한다.
- `last_ingested_from` 으로 추적하던 source 경로가 파일 이동 / 삭제로 깨진다.
- ingest 빈도가 낮을수록 drift 폭이 커지는데, page 단위로 drift 여부를 명시적으로 보는 장치가 없다.
- 자동화 없이는 6개월 이상 된 페이지를 사람이 기억으로 다시 확인하는 방식이 된다.

## Solution

`updated` 가 90일 이전이면 stale 로 표시하고, 누적 비율에 따라 severity 를 결정한다.

| 누적 비율 (stale / total) | Severity | 동작 |
|---|---|---|
| < 5% | info | 페이지별 1줄 노트, lint 출력에 포함 |
| 5% ~ 20% | warning | release gate 에서 non-blocking 경고 |
| > 20% | error (예약) | ingest 1회 의무화. P3 부터 적용 |

`updated` 가 90일 미만이라도 `last_ingested_from` 경로가 filesystem 에 없으면 stale 로 동시 flag (broken source rule, R9 와 결합).

## Threshold Semantics

| 축 | 해석 |
|---|---|
| 왜 90일인가 | release cadence 가 v0.5.x → v0.6.x ≈ 30~60일 주기. 90일은 1.5~3 release 분량의 drift window |
| `updated` 의 의미 | 마지막 ingest commit 시각이 아니라 frontmatter `updated` 필드 (사람이 명시 갱신). 자동화 시 ingest 가 양쪽 동시 갱신 필요 |
| 5% 의 의미 | wiki 전체 페이지가 50개 일 때 2~3개 stale 은 noise, 5개 이상이면 systemic drift |
| 페이지 타입별 차이 | `entities`/`concepts` 는 강한 압력 (코드/문서와 1:1). `decisions` 는 갱신 빈도 낮아 5% 미만 유지 자연스러움. `patterns` 는 사용 이력 기반 |

## When to Use

- 주 1회 정기 lint (CI smoke job 의 일부로 실행)
- release 직전 pre-release gate (`v0.X.Y` tag cut 24h 전)
- 다중 ingest (R2) 후 변화량 sanity check
- 수동 trigger: LLM 이 특정 토픽 답변 전 관련 페이지 anchor 별 freshness 확인

## When Not to Use

- `queries/<date>-<slug>.md` 같은 1회성 답변 페이지 (ephemeral, stale 검사 제외)
- P1~P4 단계별 stub 페이지 (작성 중 + 90일 경과 자연스러움, ingest 완료 후 검사 대상)
- 외부 출처 citation 만 있는 page (날짜 표기 자체가 source-of-truth 일 때)
- ADR `superseded` 상태 페이지 (`updated` 와 무관하게 archive)

## Related

- [[concepts/project-architecture]] — wiki 위치 단일성, 5 page type 의 source-of-truth
- [[decisions/adr-004-wiki-layer]] — wiki layer 자체의 결정. 90일 임계치의 정당성 배경
- [[patterns/r4-anchor-index]] — anchor 기반 index 가 stale 검사의 진입점 (anchor 별 freshness 노출)
- [SCHEMA §4 Lint Checklist](#) — 5항목 lint 의 두 번째 줄
- `check_wiki_antipatterns.py` (planned) — V-8 구현체. SCHEMA §5.3 V-8 참조
