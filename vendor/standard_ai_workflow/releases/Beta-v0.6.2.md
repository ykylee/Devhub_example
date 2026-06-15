# Beta v0.6.2 — T2 work_backlog anchor + T3 ingest atomicity

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **상태**: ✅ P3 DoD (T2 work_backlog anchor, T3 ingest atomicity). PATCH release. breaking change 없음.

## 1. 무엇이 바뀌었나

### T2: work_backlog anchor 구조 (Query)

`ai-workflow/memory/active/work_backlog.md` 를 R4-style anchor 기반 index 로 격상:
- 각 release entry → `### [[release/.../backlog/...]] {#release-v0-5-X}` format
- AI agent 가 anchor ID 로 직접 retrieval 가능
- T2 검증: `check_ingest_atomicity.py` 내 test_worklog_anchor_structure()

### T3: ingest_session_atomic() (Ingest)

`workflow_kit/common/ingest.py` 신규:
- `ingest_session_atomic()` — multi-file atomic write (handoff + backlog + state.json + worklog)
- `.ingest_lock` 으로 동시 ingest 방지
- `.partial` suffix + rename 으로 file-level atomicity

## 2. 변경 파일

| 파일 | 변경 종류 | 라인 |
|---|---|---|
| `workflow_kit/common/ingest.py` | 신규 | +100 |
| `ai-workflow/memory/active/work_backlog.md` | anchor 구조 격상 | -12 / +15 |
| `workflow-source/tests/check_ingest_atomicity.py` | 신규 (T2+T3 검증) | +120 |

## 3. 검증

```text
$ PYTHONPATH=workflow-source python3 workflow-source/tests/check_ingest_atomicity.py
T2+T3 smoke check passed (atomic write, lock contention, work_log anchor structure).
```

## 4. 다음 단계

- **v0.6.3** (P4): harness overlay memory/ sync + memory/log.md
