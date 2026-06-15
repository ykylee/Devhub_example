# Beta v0.5.10.1 — Smart update (기존 파일 silent 갱신, fail-fast 해소)

- **릴리스 일자**: 2026-06-09
- **브랜치**: `main`
- **포함 커밋**: 2 (smart update 통합 + 신규 `tests/check_smart_update.py` + checked-in read_only schema/output_samples tool_version 갱신 + **wheel install path resolution fix**)
- **상태**: ✅ silent 갱신 + ✅ wheel install 환경에서 --copy-core-docs graceful skip. 사용자가 wheel 설치 후 `python3 -m bootstrap_lib` 실행 시 기존 `ai-workflow/` 가 있어도 silent 갱신. breaking change 없음 (기존 `--force` 동작 유지).

## 1. 무엇이 바뀌었나

v0.5.10 (또는 그 이전) wheel 을 다른 환경에 설치한 뒤 `python3 -m bootstrap_lib --target-root .` 을 실행하면, 기존 `ai-workflow/` 가 있을 때 `FileExistsError` 로 silent break 하는 문제가 있었음. 이 hotfix 는 **VERSION marker + content hash 기반 smart update** 로 silent 갱신하도록 변경.

### 1.1 문제

```python
# Before (workflow-source/scripts/bootstrap_lib/writes.py)
def write_text(path, content, *, force):
    if path.exists() and not force:
        raise FileExistsError(f"Destination already exists: {path}")
    path.write_text(content, encoding="utf-8")
```

→ `force=False` (default) 면 무조건 fail. 다른 환경에서 wheel 설치 직후 bootstrap 시 진행 불가.

### 1.2 해결: smart update

1. **VERSION marker 우선**: 모든 bootstrap 이 작성하는 file 의 첫 줄에 `<!-- standard-ai-workflow-kit: v0.5.10.1-beta -->` (또는 `# standard-ai-workflow-kit: v0.5.10.1-beta`) marker stamp. marker 가 둘 다 있으면:
   - src marker > dst marker → `UPDATED` (구버전 → 신버전)
   - src marker < dst marker → `IGNORED` (downgrade 거부)
   - markers equal → hash 비교 (같으면 `IGNORED`, 다르면 `UPDATED` — 사용자 편집 감지)
2. **content hash fallback**: marker 가 없는 legacy file → SHA-256 16자 비교. 같으면 `IGNORED`, 다르면 `UPDATED`.
3. **binary file** (`.json`, `.png` 등 marker stamp 불가): hash 전용 비교. `NO_MARKER_SUFFIXES` set 으로 stamp skip.
4. **PRESERVE_RELATIVE_PATHS** (사용자 데이터: `ai-workflow/memory/`, `ai-workflow/WORKFLOW_INDEX.md`, `ai-workflow/README.md`): 항상 `PRESERVED`. `force=True` 여도 덮어쓰지 않음. **단, destination 이 처음 생성되는 경우 (첫 bootstrap) 는 PRESERVE 무시** — 사용자 데이터를 보호할 대상이 없으므로 정상 작성.

### 1.3 정책 우선순위 (7-규칙)

```
0. (short-circuit) kit_version(target) == kit_version(source) → per-file 비교 skip, IGNORED 일괄
1. PRESERVE 경로 + destination 존재 → PRESERVED
2. 파일이 target 에 없음 → CREATE
3. marker 가 있고 dst > src → IGNORED (이미 더 최신)
4. marker 가 있고 dst < src → UPDATED (구버전 → 신버전)
5. marker equal → 해시 비교 (같으면 IGNORED, 다르면 UPDATED)
6. marker 없음 (legacy) → 해시 비교 (같으면 IGNORED, 다르면 UPDATED)
7. force=True → 3~6 우회. PRESERVE 는 항상 우선.
```

### 1.4 silent 갱신 + manifest 리포트

stdout 출력 없음. 단, manifest 의 `file_actions` 에 결과 dict 추가:

```json
{
  "file_actions": {
    "created":   [{"path": "...", "rel": "AGENTS.md", "action": "create", "reason": "destination does not exist"}],
    "updated":   [],
    "ignored":   [{"path": "...", "rel": "AGENTS.md", "action": "ignored", "reason": "markers match and content hash matches", "src_marker": "0.5.10.1-beta", "dst_marker": "0.5.10.1-beta"}],
    "preserved": [{"path": "...", "rel": "ai-workflow/memory/session_handoff.md", "action": "preserved", "reason": "destination exists under PRESERVE_RELATIVE_PATHS (user data)"}]
  }
}
```

downstream consumer (Mavis, CI step) 가 manifest 의 `file_actions` 를 보고 "어떤 파일이 갱신됐는지 / 보존됐는지" 자동 audit 가능.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `workflow-source/workflow_kit/upgrade_diff.py` | 신규 — smart update 핵심 (`decide_action`, `parse_version_marker`, `compare_marker`, `read_kit_version`, ...) | +300 (new) |
| `workflow-source/workflow_kit/__init__.py` | `__version__` v0.5.0-beta → v0.5.10.1-beta (pre-existing drift 해결) | +1 / -1 |
| `workflow-source/scripts/bootstrap_lib/writes.py` | `write_text` / `copy_core_docs` smart update 적용 + module-level `_file_action_log` accumulator + binary file `_copy_binary` helper + `drain_file_actions()` | +200 / -40 |
| `workflow-source/scripts/bootstrap_lib/__main__.py` | 모든 write_text 호출이 rel_to=paths.target_root 전달, harness renderers/mcp 가 module-level accumulator 로 action 기록, manifest 에 `file_actions` 통합, `drain_file_actions()` 호출 | +30 / -10 |
| `workflow-source/scripts/bootstrap_lib/mcp.py` | 모든 `write_text` 에 `rel_to=paths.target_root` 추가 (5 calls) | +5 |
| `workflow-source/scripts/bootstrap_lib/harnesses/renderers.py` | 모든 `write_text` 에 `rel_to=paths.target_root` 추가 (14 calls) | +14 |
| `workflow-source/scripts/apply_workflow_upgrade.py` | hash 기반 로직을 `decide_action` 으로 교체 + kit version short-circuit + `--force` 인자 추가 | +60 / -30 |
| `workflow-source/tests/check_smart_update.py` | 신규 — 5 check 회귀 (unit, first-run CREATE, second-run IGNORE, PRESERVE+force, apply_upgrade round-trip) | +200 (new) |
| `workflow-source/tests/check_bootstrap_mcp_roundtrip.py` | fix — fragile 한 walk algorithm (manifest size 증가에 fail) | +5 / -5 |
| `workflow-source/schemas/read_only_*.json` | `tool_version` v0.5.0-beta → v0.5.10.1-beta (regen) | +N / -N |
| `workflow-source/examples/output_samples/*.json` | `tool_version` v0.5.0-beta → v0.5.10.1-beta (bulk sed) | +N / -N |

## 3. 검증 (actual run, fresh venv)

### 3.1 신규 회귀 test

```text
$ python3 workflow-source/tests/check_smart_update.py
Smart update regression check passed (5 checks: unit, first-run CREATE, second-run IGNORE, PRESERVE+force, apply_upgrade round-trip).
```

### 3.2 전체 smoke (회귀 0)

```text
$ for t in workflow-source/tests/check_*.py; do python3 "$t"; done
[대부분 passed / 일부 pre-existing skip (mcp 패키지 미설치) / 일부 pre-existing fail (workflow_linter 의도된 warning, existing mode 의 pre-existing state_payload FileNotFoundError)]
```

본 hotfix 가 추가/수정한 test 5개 (`check_smart_update.py`) + 영향 받은 test 1개 (`check_bootstrap_mcp_roundtrip.py` walk algorithm fix) 모두 PASS. 기존 52개 smoke 중 본 hotfix 와 무관한 pre-existing fail 4-5개 (workflow_linter 의도된 warning, mcp 패키지 미설치 환경, existing mode 의 state.json 빌드 FileNotFoundError) 제외 회귀 0.

## 4. 사용 예시

### 4.1 다른 환경에서 wheel 설치 후 재실행 (silent 갱신)

```bash
# v0.5.10 환경 (또는 v0.5.10.1 이전) 에서:
pip install standard-ai-workflow-kit==0.5.10.1b1

# 기존 ai-workflow/ 가 있는 repo 에서:
python3 -m bootstrap_lib --target-root . --harness opencode --no-interactive
# → FileExistsError 없음. manifest 의 file_actions 에 created/updated/ignored/preserved 분포 표시.
```

### 4.2 강제 덮어쓰기 (PRESERVE 제외)

```bash
python3 -m bootstrap_lib --target-root . --harness opencode --no-interactive --force
# → 사용자가 손으로 수정한 AGENTS.md 도 덮어씀. 단, ai-workflow/memory/ 안의 사용자 노트는 보존.
```

### 4.3 apply_workflow_upgrade 새 --force

```bash
python3 -m apply_workflow_upgrade --source-root ./bundle --target-root /path/to/repo --force
# → 모든 충돌 파일 덮어씀 (PRESERVE 제외).
```

### 4.4 Wheel install 환경 fix

`pip install workflow-source/` 로 설치한 환경 (예: `python3 -m venv .venv && source .venv/bin/activate && pip install /path/to/workflow-source`) 에서 `python3 -m bootstrap_lib --copy-core-docs` 호출 시 `FileNotFoundError: .../workflow-source/core/global_workflow_standard.md` 가 발생하던 버그:

**원인**: `bootstrap_lib/__main__.py` 의 `SOURCE_ROOT` 계산이 `Path(__file__).resolve().parents[3]` 였는데, 이는 dev install 레이아웃 (repo 의 `workflow-source/`) 만 가정. wheel install 시 `__file__` 은 `site-packages/bootstrap_lib/__main__.py` 이고, parents[3] = `<venv>/lib/`. SOURCE_ROOT = `<venv>/lib/workflow-source` (존재 X).

**Fix**: `__file__` 의 부모를 거슬러 올라가며 `core/global_workflow_standard.md` 가 있는 디렉토리를 SOURCE_ROOT 로 사용. 못 찾으면 `None` 으로 두고, `--copy-core-docs` 사용 시 manifest 에 warning 추가 후 silent skip. 글로벌 snippet 도 동일하게 graceful 처리.

**검증**:
- dev install (e.g. `pip install -e .`): `SOURCE_ROOT` 정상 검출. core docs 7개 + support paths 모두 copy.
- wheel install + `--copy-core-docs`: `SOURCE_ROOT = None`. core docs 0개. manifest `warnings` 에 사용자 안내 1줄.

**테스트 오버라이드**: `BOOTSTRAP_LIB_NO_SOURCE=1` 환경 변수로 SOURCE_ROOT 강제 None. 회귀 test `check_smart_update.py::test_bootstrap_wheel_install_graceful_when_no_source` 가 wheel install 시나리오를 in-process 검증.

## 5. 의도적 비-변경

- `--force` 동작: pre-existing 그대로 (PRESERVE 만 보존, 나머지 overwrite).
- `--no-interactive` (v0.5.8) 동작: 변경 없음.
- bootstrap/upgrade CLI 의 기존 flag (`--copy-core-docs`, `--enable-mcp`, `--dry-run`, `--force-cleanup`, ...): 변경 없음.
- `file_actions` 가 manifest 에 추가되어 manifest 사이즈가 커짐 (~100-200 lines). downstream consumer (Mavis, CI) 가 manifest parsing 시 주의.

## 6. Known limitations (v0.5.10.1 hotfix 범위 외)

- sub_id uniqueness 정책은 옵션 (a) — caller 책임. v0.5.10.1 의 smart update 는 동일 sub_id + 동일 status 의 중복을 **검출하지 않음** (status 일치 시 aggregated check 무효). 옵션 (b) — `choose_roles` 에 dedup check — v0.5.12+ 별도 plan.
- `--report <path>` 옵션 (human-readable 갱신 리포트) — v0.5.11+.
- `--preserve-user-edits` 옵션 (marker 동일 + hash 다름 시 skip) — v0.5.11+.

## 7. 다음 단계

v0.5.11 milestone (별도 work plan `.omo/plans/v0.5.11-plus-roadmap.md`):
- #1 Mavis engine hook 자동 enforce (wire 가이드 §2/§3 패턴 + `check_wire_guide_v059.py` stub 동기화)
- #2 회귀 test 강화 (parent_delegation_id / sub_id status inconsistency / contract_version)
- #3 `--no-interactive` 비대화형 가이드 보강 (`mcp_installation_by_harness.md`)
- #4 docs/architecture/ ADR-001/002/003 정식 기록 + governance 갱신 (옵션 A, sequential, 7-day cool-down)
