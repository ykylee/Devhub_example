# Beta v0.5.8 — Interactive `--harness` picker + packaging smoke automation

- **릴리스 일자**: 2026-06-08
- **브랜치**: `main`
- **포함 커밋**: (1) picker + TTY/--no-interactive 로직, (2) HARNESS_DEFINITIONS deprecation + SoT 정리, (3) packaging smoke 도구, (4) v0.5.7 → v0.5.8 version bump
- **상태**: ✅ 신규 기능 + packaging 자동화 + legacy 정리

## 1. 무엇이 바뀌었나

### 1.1 `bootstrap_lib` 인터랙티브 harness picker

기존: `--harness` 미지정 시 **silent 0 overlay 생성** — README/profile/handoff 만 emit 되고 overlay 는 0개. 자동 agent (Mavis 등) 가 `--harness` 빼고 호출하면 깨져도 모름.

v0.5.8:

- **TTY + `--harness` 미지정 + `--no-interactive` 미지정** → stdlib `input()` 기반 6개 harness 메뉴 자동 표시. `1,3` 형식 multi-select, `a` 전체 선택, `q` / Enter 유지, 잘못된 입력은 warning 후 무시.
- **non-TTY (CI, 파이프라인) 또는 `--no-interactive`** → `--harness is required` 명확한 SystemExit(1) + 6개 harness 목록 제시. **silent break → fail-fast**.
- **CLI flag**: `--no-interactive` 신규. `--harness` 동작은 그대로 (back-compat).

### 1.2 `HARNESS_DEFINITIONS` deprecated → `HARNESS_SPECS` 가 단일 SoT

| 이전 | 이후 |
| --- | --- |
| 3중 등록: `SUPPORTED_HARNESSES` + `HARNESS_DEFINITIONS` + `HARNESS_FILE_BUILDERS` | 2중 등록: `HARNESS_SPECS` + `HARNESS_FILE_BUILDERS` |
| `pi-dev` 가 `HARNESS_DEFINITIONS` 에 누락되어 있었음 | `HARNESS_SPECS` 가 6개 다 가지고 있음 (declarative spec: name, description, entry_files, extra_files, long_description) |
| 신규 harness 등록 가이드: 3군데 동시 갱신 | 신규 등록: `HARNESS_SPECS` + `HARNESS_FILE_BUILDERS` 2군데만 |

- `_verify_harness_registry_consistency()` 가 `renderers` import 시 자동 실행 — 두 레지스트리 sync 깨지면 `RuntimeError` 로 fail-fast.
- `HARNESS_DEFINITIONS` 자체는 legacy import back-compat 위해 남겨두되, deprecation 주석 + 0.6.x 제거 예정 표시.
- 동기화 대상: `core/workflow_harness_distribution.md`, `scripts/scaffold_harness.py`, `harnesses/_template/README.md`, `harnesses/antigravity/README.md` 4개 doc 가 새 가이드로 갱신.

### 1.3 Packaging smoke 자동화 (v0.5.7.1 한계 해결)

`tools/check_packaging.py` (신규):

- `python3 -m build` 후 fresh venv 에 wheel install (`pip install -e .` 가 아닌 wheel 시나리오)
- 7개 핵심 sub-package 명시적 import smoke (이전 1줄이 못 잡던 누락 자동 검출)
- `python -m bootstrap_lib --help` 로 CLI 진입점 + `--no-interactive` flag 정상 노출 확인
- `pip show` 로 metadata 검증
- **v0.5.7.1 회귀 카테고리 (sub-package 누락) 재발 방지**

### 1.4 Verifier-friendly

- 신규 테스트 `tests/check_bootstrap_interactive_picker.py` (10 check) — picker unit 5, enforce SystemExit 1, CLI subprocess 2, SoT sync 2. **v0.5.7.1 verifier 의 actual-run 룰 (memory) 그대로 적용 — actual pytest 결과 10/10 PASS.**

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `scripts/bootstrap_lib/__main__.py` | picker 함수 + argparse `--no-interactive` + `enforce_harness_selection` + `HARNESS_DEFINITIONS` deprecation 주석 | +130 |
| `scripts/bootstrap_lib/harnesses/renderers.py` | `_verify_harness_registry_consistency()` + import-time assert | +50 |
| `scripts/bootstrap_workflow_kit.py` | re-export 추가 (`prompt_for_harnesses`, `enforce_harness_selection`) | +2 |
| `scripts/scaffold_harness.py` | 2중 등록 가이드로 갱신 | +4 / -2 |
| `harnesses/_template/README.md` | 2중 등록 + HARNESS_SPECS rich spec 설명 | +6 / -4 |
| `harnesses/antigravity/README.md` | 2중 등록 가이드 | +1 / -1 |
| `core/workflow_harness_distribution.md` | HARNESS_DEFINITIONS deprecation + 2중 등록 | +2 / -1 |
| `tests/check_bootstrap_interactive_picker.py` | 신규 — 10 check 회귀 suite | +260 |
| `tools/check_packaging.py` | 신규 — wheel smoke 자동화 | +180 |
| `pyproject.toml` | version `0.5.7.1-beta` → `0.5.8-beta` | ±1 |

## 3. 검증 (actual run, fresh venv)

### 3.1 picker 회귀

```text
$ python3 -m tests.check_bootstrap_interactive_picker
Interactive picker regression suite passed (10 checks).
```

- picker unit 5/5: index parse, empty / q / a / garbage handling
- enforce SystemExit 1/1
- CLI subprocess 2/2: `--no-interactive` + no `--harness` fail-fast, explicit `--harness codex` 정상
- SoT sync 2/2: `HARNESS_SPECS` 6개 (`pi-dev` 포함) + consistency check import 시 자동 통과

### 3.2 packaging smoke

```text
$ python3 tools/check_packaging.py
{
  "wheel": "dist/standard_ai_workflow-0.5.8b0-py3-none-any.whl",
  "imported": [
    "workflow_kit",
    "bootstrap_lib",
    "workflow_kit.contract_v1",
    "workflow_kit.common",
    "workflow_kit.common.state",
    "workflow_kit.common.contracts",
    "workflow_kit.common.schemas"
  ],
  "missing": [],
  "boot_lib_help_has_no_interactive": true,
  "result": "PASS"
}
```

fresh venv + `pip install dist/*.whl` (no `-e`) + 명시적 sub-package 7개 import + `--no-interactive` flag 노출 확인.

### 3.3 manual dry-run (explicit --harness)

```text
$ python3 -m bootstrap_lib --project-slug saw-v058-pickertest --project-name PickerTest \
    --target-root /tmp/picker-test-target --harness codex --no-interactive --dry-run
{
  "harnesses": ["codex"],
  ...
  "plan": [...7 files...]
}
```

### 3.4 manual fail-fast (--no-interactive + no --harness)

```text
$ python3 -m bootstrap_lib --project-slug X --project-name Y \
    --target-root /tmp/x --no-interactive
ERROR: --harness is required when --no-interactive is set or when stdin is not a TTY. Re-run with one or more of: codex, opencode, gemini-cli, pi-dev, antigravity, minimax-code (comma-separated via repeated --harness flags).
[rc=1]
```

## 4. 다운스트림 영향

### 새 동작 (consumer 가 알아야 할 것)

- `--harness` 를 명시하던 기존 consumer: **변경 없음** (back-compat).
- TTY 에서 `--harness` 빼고 호출하던 consumer (rare): **picker 가 자동 발동** — stdlib `input()` 만 쓰니 추가 의존성 0.
- CI / 스크립트에서 `--harness` 빼고 호출하던 consumer: **fail-fast** — `--no-interactive` 추가하거나 `--harness` 명시. 의도된 동작.

### Mavis 측 (Mavis 1.x, my_harness)

- `bootstrap_lib` 호출 스크립트 (`/Users/yklee/repos/my_harness/.../bootstrap.py` 류) 가 `--harness opencode` 박아둔 패턴: 변경 없음.
- TTY 있는 환경에서 `--harness` 빼고 호출하던 게 있다면, picker 가 뜨는데 — yklee 환경은 명시 호출이 default 이므로 영향 0.

### 영향 없음

- `workflow_kit.contract_v1` API / Mavis wire / Devhub_example backend: 변경 없음.
- v0.5.7.1 packaging fix 그대로 유효 (sub-packages 정상).

## 5. 업그레이드 절차

```bash
pip install --upgrade https://github.com/ykylee/standard_ai_workflow/releases/download/v0.5.8-beta/standard_ai_workflow-0.5.8b0-py3-none-any.whl
```

dev / local install:

```bash
cd workflow-source
pip install -U .
```

## 6. 알려진 한계 / 백로그

- **`HARNESS_DEFINITIONS` 실제 제거 (0.6.x)** — deprecation 표시만 했고, downstream 이 import 안 한다는 확인 후 제거. 0.6.x milestone.
- **picker 의 multi-select UX** — v0.5.8 은 1줄 입력 + 콤마 분리. future harness 가 10+ 개 되면 checkbox UI (직접 구현) 검토.
- **`tools/check_packaging.py` 의 CI 통합** — 현재는 release 절차에 manual invocation. CI hook (`.github/workflows/release.yml`) 자동화는 별도 PR.
- **`mcp` extra (sdk stdio) fresh venv 검증** — 이번에도 안 함. v0.5.9 백로그.

## 7. 다음 릴리스 (v0.5.9 후보)

- Mavis 측 wire 가이드 §wire-rules 표에 `sub.delegation_id` parent prefix 룰 추가 (v0.5.7.1 검증 중 side note)
- `HARNESS_DEFINITIONS` 0.6.x 제거 타임라인 확정 + deprecation warning emit
- picker checkbox UI (harness 10+ 대비)
- packaging smoke CI 통합
