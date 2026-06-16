# Wiki ↔ Code Cross-Reference SSOT (DevHub)

## 목적

본 문서는 **DevHub 의 in-repo wiki (ai-workflow/wiki/) 와 source code (scripts/, tests/, vendor/) 의 양방향 reference 규칙** 의 SSOT.

코드 maintenance 시 **wiki 의 L1 page 가 source code 의 어느 부분과 관련** 되어 있는지 명확히 식별 가능. **wiki 의 변경 사항이 source code 의 어느 부분에 영향** 을 주는지 추적 가능.

## 1. 양방향 reference 의 두 가지 형태

### 1.1 Source → Wiki (코드 → 위키)

Source code 의 **header docstring / comment** 에 `Wiki: ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/<page>.md` 한 줄로 명시. 향후 source code 를 읽는 사람이 **이 코드의 의도/결정/패턴은 어느 L1 page 에 기록** 되어 있는지 즉시 reference.

**Format**:
```python
#!/usr/bin/env python3
"""<module docstring> (기존 본문 유지).

Wiki: ai-workflow/wiki/decisions/v0.7.37-import.md
      ai-workflow/wiki/concepts/devhub-overview.md
"""
```

**규칙**:
- `Wiki:` prefix 사용 (다른 주석과 visually distinct)
- 한 줄에 1 page (`\` line continuation 으로 multiple pages 가능)
- 5 L1 dir 중 하나 (`concepts/`, `decisions/`, `entities/`, `patterns/`, `topics/`)
- index.md, RAW_MIRROR_MANIFEST.md, sources/ 의 L2 는 reference 대상 ❌ (L1 만)

### 1.2 Wiki → Source (위키 → 코드)

L1 page 의 frontmatter `last_ingested_from` field 가 source code path 가리킴. **기존의 1-way ingestion reference** 그대로 유지. L1 page 작성/수정 시 source code 의 정확한 path 명시.

**Format**:
```yaml
---
type: decision
last_ingested_from: scripts/wiki-sync-devhub.sh + docs/llm-wiki/mirror-list.md
related_pages: [concepts/devhub-overview, patterns/in-repo-redirect]
---
```

**규칙**:
- `last_ingested_from` field 가 source path 명시 (기존 convention)
- `related_pages` 가 wiki page 간 cross-reference 명시 (선택적)
- source path 는 **single source of truth** (SSOT) — 변경 시 L1 page 의 `updated` 갱신

## 2. Cross-reference consistency test

`tests/check_wiki_drift_devhub.py` 의 `_devhub_self_validate_l1_pages` 가 wiki format SSOT 검증 (PR #615). 본 SSOT 와 함께 **cross-reference consistency** 도 추가 검증.

**Validation 항목**:
1. **Source code 의 `Wiki:` reference 가 실제 L1 page 에 매칭** (broken link detection)
2. **L1 page 의 `last_ingested_from` 의 source path 가 실제 존재** (기존 test_ingested_from_paths_exist)
3. **L1 page 의 `related_pages` 가 실제 L1 page 에 매칭** (broken cross-page link detection)
4. **`Wiki:` reference 의 count** 가 **L1 page 의 `last_ingested_from` 의 count** 와 **balanced** (1-way 만 있고 1-way 없으면 asymmetry detection)

## 3. Workflow

### 3.1 Code maintenance 시 (source code 변경)

1. **기존 source code 의 `Wiki:` reference 확인** (예: `grep "Wiki:" scripts/emit_wiki_l2_devhub.py`)
2. **변경 사항이 L1 page 의 어떤 결정/패턴과 연관** 인지 확인 (예: 새 in-repo redirect → `patterns/in-repo-redirect.md` 갱신)
3. **L1 page 의 `updated` field 갱신** (frontmatter 의 YYYY-MM-DD)
4. **`docs/llm-wiki/mirror-list.md` 의 Phase breakdown** 갱신 (file count 변경 시)
5. **`bash scripts/wiki-sync-devhub.sh --no-clean`** 1회 실행 (raw mirror 갱신)

### 3.2 Wiki update 시 (L1 page 변경)

1. **L1 page 의 frontmatter `last_ingested_from` 의 source path 검증** (기존 invariant test)
2. **related_pages 의 L1 page 가 실제 존재** 검증 (cross-page link)
3. **L2 emit 도구로 sources/ 갱신** (`python3 scripts/emit_wiki_l2_devhub.py --source <L1> --apply`)
4. **raw mirror 갱신** (`bash scripts/wiki-sync-devhub.sh --no-clean`)

## 4. 현재 cross-reference inventory (2026-06-16, 본 PR)

### 4.1 Source → Wiki

| Source file | Wiki references |
|---|---|
| `scripts/emit_wiki_l2_devhub.py` | `decisions/v0.7.37-import.md`, `concepts/devhub-overview.md`, `decisions/v0.7.17-import.md` |
| `scripts/emit_wiki_l2_devhub_vendor.py` | `decisions/v0.7.37-import.md` (vendor import + in-repo redirect) |
| `scripts/wiki-sync-devhub.sh` | `topics/standard-ai-workflow-vendor.md` (raw mirror 운영) |
| `scripts/wiki-mirror-sources.sh` | `topics/standard-ai-workflow-vendor.md` (mirror list) |
| `scripts/atomic_write.py` | `patterns/in-repo-redirect.md` (POSIX atomic write + DevHub 자체 구현) |
| `scripts/check_vendor_smoke.sh` | `topics/standard-ai-workflow-vendor.md` (vendor smoke gate) |
| `tests/check_wiki_drift_devhub.py` | `concepts/devhub-overview.md` (drift check) |
| `tests/check_l1_format_devhub.py` | `concepts/devhub-overview.md` (L1 format SSOT) |
| `tests/check_mirror_list_devhub.py` | `topics/standard-ai-workflow-vendor.md` (mirror list byte-id) |
| `tests/check_emit_wiki_l2_devhub.py` | `concepts/devhub-overview.md` (emit smoke) |
| `tests/check_wiki_ingest_devhub.py` | `decisions/v0.7.17-import.md` (in-repo redirect) |
| `tests/check_atomic_write_devhub.py` | `patterns/in-repo-redirect.md` (atomic_write helper) |
| `vendor/standard_ai_workflow/tests/check_wiki_drift.py` | (vendor, no Wiki: reference — DevHub 의 PR #608 P0 fix) |

### 4.2 Wiki → Source (L1 page → last_ingested_from)

| L1 page | last_ingested_from (key sources) |
|---|---|
| `concepts/devhub-overview.md` | `AGENTS.md + docs/architecture.md + release_v0-1_roadmap.md` |
| `decisions/v0.7.17-import.md` | `ai-workflow/IMPORT_NOTES.md + vendor/.upstream-url` |
| `decisions/v0.7.37-import.md` | `ai-workflow/IMPORT_NOTES.md + vendor/.upstream-url` |
| `entities/keycloak-iam.md` | `ADR-0019 + keycloak operations + infrastructure/` |
| `patterns/in-repo-redirect.md` | `IMPORT_NOTES.md + vendor emit + isolation test` |
| `topics/standard-ai-workflow-vendor.md` | `vendor README + core workflow files` |

## 5. PR impact (Wiki cross-reference 변경 시)

1. **SSOT 변경** (`docs/governance/wiki-cross-reference.md`) — 본 doc
2. **Source code 변경** (`Wiki:` reference 추가/갱신) — script file header
3. **L1 page 변경** (`last_ingested_from`, `related_pages`) — wiki L1 file frontmatter
4. **Cross-reference consistency test** (`tests/check_wiki_drift_devhub.py` 의 cross-ref test) — broken link detection

## 6. Reference

- `docs/governance/l1-format.md` (PR #615) — L1 page format SSOT (frontmatter constraints, type enum)
- `docs/llm-wiki/mirror-list.md` (PR #604/#618) — raw mirror 의 15 pattern scope
- `ai-workflow/IMPORT_NOTES.md` (PR #603/#619) — vendor import 의 source-of-truth
- `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/*.md` — 6 L1 page 의 운영 record
