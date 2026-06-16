# L1 Wiki Page Format SSOT (DevHub)

## 목적

본 문서는 **DevHub 의 in-repo wiki L1 page format 의 단일 source-of-truth (SSOT)**.
5 dir (`concepts/`, `decisions/`, `entities/`, `patterns/`, `topics/`) 의 L1 page 가 따라야 할
frontmatter / type / status / dates 의 정확한 제약.

`tests/check_wiki_drift_devhub.py` 의 `_devhub_self_validate_l1_pages` 가 본 SSOT 의
enforce 도구. format 변경 시 본 doc + test 동시 update 필수.

## File location

| Dir | Purpose | 허용 type |
|---|---|---|
| `ai-workflow/wiki/concepts/` | 추상 concept 문서 | `concept` |
| `ai-workflow/wiki/decisions/` | ADR-style 결정 | `decision` |
| `ai-workflow/wiki/entities/` | 외부 시스템/엔티티 | `entity` |
| `ai-workflow/wiki/patterns/` | reusable pattern | `pattern` |
| `ai-workflow/wiki/topics/` | 일반 topic/주제 | `topic` |

위 dir 외 위치 (예: `ai-workflow/wiki/sources/`, `ai-workflow/wiki/raw/`) 의 file 은
본 SSOT 의 validation 대상 ❌ (sources/ 는 L2 derived view, raw/ 는 1:1 mirror).

## Frontmatter (YAML, 4 required + 2 optional 권장)

```yaml
---
type: <5종 enum 중 1>
status: <active|draft>
created: <YYYY-MM-DD>
updated: <YYYY-MM-DD>
last_ingested_from: <space-separated path list>  # optional, 모든 운영 page 에 존재
active_since: <YYYY-MM-DD>                       # optional, status=active 일 때 권장
---
```

## Required fields

| Field | Format | Validation |
|---|---|---|
| `type` | enum: `concept` \| `decision` \| `entity` \| `pattern` \| `topic` | 다른 값 → fail |
| `status` | enum: `active` \| `draft` | 다른 값 → fail |
| `created` | ISO 8601 date `YYYY-MM-DD` | 임의의 string → fail (이전 PR #608 의 P0 fix 와 동일 category) |
| `updated` | ISO 8601 date `YYYY-MM-DD` | 임의의 string → fail |

## Optional fields (권장)

| Field | Format | 비고 |
|---|---|---|
| `last_ingested_from` | space-separated path list | 모든 5/5 page 가 보유. `tests/test_ingested_from_paths_exist` 가 path 존재 검증 |
| `active_since` | YYYY-MM-DD | `status=active` 인 page 의 "when this became active" — 운영 안정성 추적 용도 |
| `active_reason` | free text (한 줄) | `active` 상태 의 정당성 (없어도 OK) |

## Format difference vs vendor ADR-format

| | DevHub (본 SSOT) | vendor (standard_ai_workflow) |
|---|---|---|
| `type` enum | concept/decision/entity/pattern/topic | "concept" only |
| `status` enum | active/draft | accepted/proposed/deprecated |
| required | created/updated | adr_id (decision 한정) |
| optional | last_ingested_from, active_since | superseded_by (decision 한정) |

vendor 의 `test_l1_wiki_pages_format` 가 ADR-format 을 기대하지만, DevHub 의 page 는
위 SSOT 를 따르므로 vendor test 를 skip + DevHub 자체 test (`_devhub_self_validate_l1_pages`)
로 대체.

## DevHub vs vendor validation 분기

- **vendor test 가 DevHub 에서 동작** → format mismatch 로 false positive. **DevHub 는
  skip** 처리 (`SKIP test_l1_wiki_pages_format (DevHub 자체 format 검증 ...)`).
- **DevHub 자체 test** = `_devhub_self_validate_l1_pages`. 본 SSOT 의 변경 시
  `_devhub_self_validate_l1_pages` 의 `VALID_TYPES` / `VALID_STATUS` / `REQUIRED_FIELDS`
  / date format regex 동시 update 필수.

## Validation flow (mtime diagram)

```
[L1 page 작성/수정]
        |
        v
[frontmatter `---` delimiter present?] ----no----> fail: missing frontmatter
        |
       yes
        |
        v
[required 4 fields present? type, status, created, updated] ----missing----> fail
        |
       all 4
        |
        v
[type in {concept, decision, entity, pattern, topic}?] ----no----> fail
        |
       yes
        |
        v
[status in {active, draft}?] ----no----> fail
        |
       yes
        |
        v
[created/updated format YYYY-MM-DD?] ----no----> fail
        |
       yes
        |
        v
        PASS
```

## PR impact

본 SSOT 변경 시:
1. `docs/governance/l1-format.md` update
2. `tests/check_wiki_drift_devhub.py` 의 `_devhub_self_validate_l1_pages` 의 assertion
   update (VALID_TYPES / VALID_STATUS / REQUIRED_FIELDS / date regex)
3. 기존 5 L1 page 의 byte danno (필드 추가/수정 시)
4. PR description 의 §"Format constraint change" 섹션 명시

## Reference

- `tests/check_wiki_drift_devhub.py` (vendor 의 test 의 DevHub adapter, 자체 validation)
- `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/*.md` (5 page 의 sample)
- `docs/governance/document-standards.md` (문서 작성 일반 원칙)
- `docs/llm-wiki/mirror-list.md` (raw mirror 1:1 byte-identical 운영)
- `scripts/emit_wiki_l2_devhub.py` (`L1_DIR_TO_TYPE` 의 type derive, 본 SSOT 와 동일 enum 사용)
