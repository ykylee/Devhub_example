# Beta v0.6.1.5 — R9 wiki-ingest source rule (archive/ only)

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **상태**: ✅ P2.5 DoD (R9 source rule + V-R9 lint). PATCH release. breaking change 없음.

## 1. 무엇이 바뀌었나

### R9 wiki-ingest source rule (P2.5)

wiki-ingest/memory-ingest 는 `memory/archive/` 만 source 로 사용. `memory/active/` 는 절대 ingest 안 함 (R9). V-R9 lint `check_wiki_source_rule.py` 신규.

**변경 파일**:
- `workflow-source/tests/check_wiki_source_rule.py` 신규 (+50 lines)

## 2. 검증

```text
$ python3 workflow-source/tests/check_wiki_source_rule.py
V-R9 check passed: no active/ references as wiki-ingest source.
```

## 3. 다음 단계

- **v0.6.2** (P3): T2 Query + T3 Ingest
