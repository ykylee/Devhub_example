# Beta v0.6.3 — P4 harness overlay consistency + memory/log.md (v0.6.x series final)

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **커밋**: `6b2bf00` (v0.6.0 첫 커밋 이후 총 7 commits, ~5,500+ lines)
- **상태**: ✅ v0.6.x series complete (P1~P4). breaking change 없음.

## 1. v0.6.x Series 개요

v0.6.x 라인은 Karpathy LLM Wiki 패턴을 도입하고 기존 memory layer 를 3-state lifecycle 로 진화시키는 4단계 마일스톤 시리즈.

| 마일스톤 | 버전 | 핵심 |
|---|---|---|
| **P1** (LLM Wiki Layer) | v0.6.0 | `ai-workflow/wiki/` 신설 + SCHEMA + 2 concepts + lint V-1·V-4 + ADR-004 |
| **P1.5** (memory/active/ rename) | v0.6.0.1 | bootstrap `--enable-wiki` + 6 harness wiki/ stub + 50+ path updates |
| **P2** (Freeze) | v0.6.1 | memory-freeze skill + freeze lint + memory lint 4종 + R7 merge ext |
| **P2.5** (Source Rule) | v0.6.1.5 | R9 wiki-ingest source = archive/ only (V-R9) |
| **P3** (Query+Ingest) | v0.6.2 | `ingest_session_atomic()` + work_backlog anchor |
| **P4** (Sync) | v0.6.3 | memory/log.md + harness overlay consistency check |

## 2. v0.6.3 변경 파일

| 파일 | 변경 종류 | 라인 |
|---|---|---|
| `workflow_kit/common/ingest.py` | 신규 (P3) | +100 |
| `ai-workflow/memory/log.md` | 신규 (P4) | +12 |
| `workflow-source/tests/check_ingest_atomicity.py` | 신규 (T2+T3) | +120 |
| `workflow-source/tests/check_harness_overlay_consistency.py` | 신규 (P4) | +35 |
| `workflow-source/MEMORY_GOVERNANCE.md` | §4 freeze 규칙 추가 | +12 |
| `workflow-source/tests/check_memory_lint.py` | 신규 (T1) | +140 |
| `workflow-source/tests/check_memory_freeze_lint.py` | 신규 (R10) | +80 |
| `workflow-source/skills/memory-freeze/SKILL.md` | 신규 (R8) | +50 |
| `workflow-source/skills/memory-freeze/scripts/run_memory_freeze.py` | 신규 (R8) | +90 |
| `workflow-source/tests/check_wiki_source_rule.py` | 신규 (R9) | +50 |

## 3. 최종 아키텍처

```
ai-workflow/
├── memory/                  # gitignore — workflow state
│   ├── active/              # mutable per-session (state.json, PROJECT_PROFILE, work_backlog...)
│   ├── archive/YYYY-MM-DD/  # immutable freeze (R8, .frozen marker)
│   ├── release/v0.5.X/      # per-release deep freeze
│   ├── log.md               # append-only operation log
│   ├── plans/               # TASK-XXX plans (preserved)
│   ├── codex/ gemini/       # per-harness phase dirs (preserved)
│   └── archive/             # misc old reports (preserved)
├── wiki/                    # git-tracked — knowledge
│   ├── SCHEMA.md            # operating constitution (5 page types, 3 workflows)
│   ├── index.md             # R4 anchor-based catalog
│   ├── log.md               # append-only ingest/query log
│   ├── .gitignore           # .ingest_lock exclusion
│   └── concepts/            # 2 concept pages (mcp-transport, orchestrator-subagent-pattern)
└── bootstrap_lib/           # --enable-wiki flag
    └── wiki.py              # wiki skeleton emitter (208 lines)
```

## 4. 최종 검증

```
```
$ python3 workflow-source/tests/check_wiki_lint.py          [V-1+V-4] PASS
$ python3 workflow-source/tests/check_wiki_location.py      [V-1]    PASS
$ python3 workflow-source/tests/check_wiki_index_structure.py [V-4]  PASS
$ python3 workflow-source/tests/check_memory_freeze_lint.py  [R8+R10] PASS
$ python3 workflow-source/tests/check_memory_lint.py         [T1]    PASS
$ python3 workflow-source/tests/check_wiki_source_rule.py    [R9]    PASS
$ python3 workflow-source/tests/check_ingest_atomicity.py    [T2+T3] PASS
$ python3 workflow-source/tests/check_harness_overlay_consistency.py [P4] PASS
```

## 5. 다음 단계

- **v0.7**: contract v2 streaming, 추가 하네스, federated sync 평가
- **v0.7+**: wiki-ingest skill 자동화 (session-end trigger), freeze → wiki-ingest 자동 연동
