# Beta v0.6.1 — R8 Freeze + R10 Freeze Lint + T1 Memory Lint 4종 + R7 merge-res ext

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **상태**: ✅ P2 DoD (memory freeze + freeze lint + memory lint + merge-doc-reconcile wiki conflict type extension). PATCH release (v0.6.0.1 → v0.6.1). breaking change 없음.

## 1. 무엇이 바뀌었나

### 1.1 R8 Memory Raw Freeze (memory-freeze skill)

`ai-workflow/skills/memory-freeze/` 신규. session 종료 시 `active/` 상태를 `archive/YYYY-MM-DD/` 로 freeze. COPY 방식 (active/ 파일 유지). `.frozen` YAML 마커 포함.

**사용 예:**
```bash
python3 workflow-source/skills/memory-freeze/scripts/run_memory_freeze.py
```

**freeze 결과 예시:**
```
archive/2026-06-12/
  PROJECT_PROFILE.md
  project_status_assessment.md
  repository_assessment.md
  state.json
  state.json.template
  work_backlog.md
  .frozen              # YAML: frozen_at, source, files
```

### 1.2 R10 Freeze Lint (`check_memory_freeze_lint.py`)

`workflow-source/tests/check_memory_freeze_lint.py` 신규. freeze 무결성 검증:
- V-R8: `.frozen` 마커 존재 + `frozen_at` 필드
- V-R10: 5개 required 파일 전부 archive/ 존재
- `active/` 파일 존재 여부

### 1.3 T1 Memory Lint 4종 (`check_memory_lint.py`)

`workflow-source/tests/check_memory_lint.py` 신규 (umbrella + 4개 sub-check):
- T1a contradiction: handoff next_actions vs backlog done 불일치
- T1b stale: 90일+ 미갱신 파일 감지
- T1c orphan: inbound link 없는 backlog/task 파일
- T1d missing: TASK-XXX 언급 but 파일 없음

### 1.4 R7 merge-doc-reconcile wiki conflict type 확장

`workflow-source/skills/merge-doc-reconcile/SKILL.md` §7 신규. wiki merge 시 4가지 conflict type 분류 (line/section/semantic/index).

### 1.5 MEMORY_GOVERNANCE.md §4 Freeze 규칙

`workflow-source/MEMORY_GOVERNANCE.md` §4 신규. freeze 시점·단위·내용·동작·immutability·marker·중복·wiki ingest source 규칙 명시.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
|---|---|---|
| `workflow-source/skills/memory-freeze/SKILL.md` | 신규 | +50 |
| `workflow-source/skills/memory-freeze/scripts/run_memory_freeze.py` | 신규 | +90 |
| `workflow-source/tests/check_memory_freeze_lint.py` | 신규 | +80 |
| `workflow-source/tests/check_memory_lint.py` | 신규 | +140 |
| `workflow-source/MEMORY_GOVERNANCE.md` | §4 freeze 규칙 추가 | +12 |
| `workflow-source/skills/merge-doc-reconcile/SKILL.md` | §7 wiki conflict type | +14 |

## 3. 검증 (actual run)

```text
$ python3 workflow-source/tests/check_memory_freeze_lint.py
Freeze lint check passed. 1 archive(s) found, all required files present.

$ python3 workflow-source/tests/check_memory_lint.py
[PASS] contradiction
[PASS] stale
[PASS] orphan
[PASS] missing
Memory lint smoke check passed (all 4 checks).

$ python3 workflow-source/skills/memory-freeze/scripts/run_memory_freeze.py
{"status": "success", "archive_path": ".../archive/2026-06-12", "file_count": 6}
```

## 4. 의도적 비-변경

- freeze = COPY, NOT move (active/ 유지). 추후 R9 source rule 에서 archive/ 만 ingest
- memory-freeze 는 독립 스킬. session-end 통합은 P3 (v0.6.2) 에서 wiki-ingest + freeze 통합
- memory lint 4종은 umbrella 에 통합 (개별 파일 4개로 분리하지 않음)

## 5. Known limitations

### 5.1 v0.6.1.5 (P2.5) — 다음 단기

- R9 wiki-ingest source rule (`archive/` 만 ingest)
- freeze → wiki-ingest 자동 연동

### 5.2 v0.6.2+ (P3~P4) — 중장기

- T2 Query (work_backlog anchor) + T3 Ingest (multi-file atomic)
- 6 harness overlay memory/ sync + memory/log.md
- session-end 자동 freeze (backlog-update --apply + freeze 연동)

## 6. 다음 단계

- **v0.6.1.5** (P2.5): R9 wiki-ingest source rule
- **v0.6.2** (P3): T2 query + T3 ingest
- **v0.6.3** (P4): harness overlay memory/ sync + memory/log.md

## 7. 관련 plan / ADR

| 종류 | 경로 |
|---|---|
| Memory Layer 진화 | `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` |
| 분산 위키 규칙 | `.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md` (v0.2.1) |
| ADR-004 | `docs/architecture/ADR-004-llm-wiki-layer.md` (accepted) |
