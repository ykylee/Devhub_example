# Beta v0.5.7.1 — Hotfix: package_data 누락으로 `python -m bootstrap_lib` 가 죽던 문제

- **릴리스 일자**: 2026-06-08
- **브랜치**: `main` (hotfix — 별도 브랜치 안 갈음)
- **포함 커밋**: v0.5.7 이후 1 squash (pyproject.toml)
- **상태**: ✅ packaging hotfix — 신규 기능 없음, 다운스트림 복구만

## 1. 무엇이 고쳐졌나

v0.5.7 wheel 을 fresh venv 에 `pip install` 하면 `workflow_kit.common.state.builder`, `workflow_kit.common.contracts.*`, `workflow_kit.common.schemas.*` 모듈이 포함되지 않음. 결과:

- `from workflow_kit.common.state import builder` → `ModuleNotFoundError`
- `python -m bootstrap_lib ...` → `ModuleNotFoundError: No module named 'workflow_kit.common.state.builder'`
- `bootstrap_lib` CLI 자체가 import 단계에서 죽어서, downstream (`Devhub_example`, `my_harness` 등) 이 `pip install` 후 harness / skills / templates bootstrap 불가

`pyproject.toml` 의 `[tool.setuptools.packages]` 와 `[tool.setuptools.package-dir]` 에 `workflow_kit.common.state`, `workflow_kit.common.contracts`, `workflow_kit.common.schemas` 3개 sub-package 가 빠져 있었던 게 원인. `git log` 상 v0.5.5 (PR #18, §4.2/§5.2 추가) 에서 추가됐는데 packaging 박스 갱신이 누락된 채 v0.5.6 / v0.5.7 까지 진행됨.

## 2. 변경 내용 (diff 요약)

`workflow-source/pyproject.toml` — 두 군데 3줄씩 추가:

```diff
 [tool.setuptools]
 packages = [
     "bootstrap_lib",
     "bootstrap_lib.harnesses",
     "workflow_kit",
     "workflow_kit.common",
     "workflow_kit.common.modes",
+    "workflow_kit.common.state",
+    "workflow_kit.common.contracts",
+    "workflow_kit.common.schemas",
     "workflow_kit.contract_v1",
     "workflow_kit.server",
     "workflow_kit.harness",
 ]

 [tool.setuptools.package-dir]
 "workflow_kit" = "workflow_kit"
 "workflow_kit.common" = "workflow_kit/common"
 "workflow_kit.common.modes" = "workflow_kit/common/modes"
+"workflow_kit.common.state" = "workflow_kit/common/state"
+"workflow_kit.common.contracts" = "workflow_kit/common/contracts"
+"workflow_kit.common.schemas" = "workflow_kit/common/schemas"
 "workflow_kit.server" = "workflow_kit/server"
 "workflow_kit.harness" = "workflow_kit/harness"
```

소스 코드 / 스키마 / API 변경 **없음**. v0.5.7 의 모든 기능 (contract v1 §4.2/§5.2 fan-out/in, §6.3 cross-ref row, `recommend_model_tier`, `artifacts[].action`) 은 그대로.

## 3. 검증

fresh venv (`/tmp/saw-0571-test`, Python 3.14) 에 v0.5.7.1 wheel 설치 후:

| 항목 | v0.5.7 | v0.5.7.1 |
| --- | --- | --- |
| `twine check dist/*` | PASSED | PASSED |
| `pip install ...whl` (no deps missing) | OK | OK |
| `python -m bootstrap_lib --help` | ❌ `ModuleNotFoundError: workflow_kit.common.state.builder` | ✅ help 출력 정상 |
| `from workflow_kit.common.state import builder, cache` | ❌ | ✅ |
| `from workflow_kit.common.contracts import base, high_value` | ❌ | ✅ |
| `from workflow_kit.common.schemas import base, worker, validation` | ❌ | ✅ |
| `from workflow_kit.contract_v1 import choose_role, choose_roles, validate_output, validate_fanin_output, recommend_model_tier` | ✅ | ✅ |
| `python -m bootstrap_lib --project-slug X ... --dry-run` (target dir 존재) | ❌ | ✅ plan JSON 정상 출력 |

## 4. 다운스트림 영향

### 영향 받았던 곳 (v0.5.7 install 후 broken)

- `Devhub_example` — `pip install standard-ai-workflow` 후 `python -m bootstrap_lib` 단계가 필요했던 신규 프로젝트 init 시나리오
- `my_harness` — harness overlay 적용 시 `bootstrap_lib` 경유
- 기타 downstream 신규 consumer

### 영향 없었던 곳 (v0.5.7 정상 동작)

- `from workflow_kit.contract_v1 import ...` 만 쓰는 Mavis 측 wire (이미 v0.5.7 의 회귀 테스트 `check_contract_v1_multi_component.py` 가 fresh venv 에서 PASS — 따라서 위 3개 sub-package 는 **이름만** import 안 되었고 실제 fan-out/in 동작 자체는 영향 없었음)

## 5. 업그레이드 절차

```bash
pip install --upgrade https://github.com/ykylee/standard_ai_workflow/releases/download/v0.5.7.1-beta/standard_ai_workflow-0.5.7.1b0-py3-none-any.whl
```

dev / local install:

```bash
cd workflow-source
pip install -U .
```

## 6. 알려진 한계 (v0.5.7.1 에서 미해결)

- **아직 패키징 박스에 누락 가능성 있음** — 이번 hotfix 는 v0.5.7 검증 중 `python -m bootstrap_lib --help` 가 죽어서 발견된 3개 sub-package 만 다룸. 다른 sub-package 가 발견되면 v0.5.7.2 로 후속 hotfix.
- **자동 누락 검증 부재** — wheel 빌드 후 `python -c "import workflow_kit, bootstrap_lib"` 같은 import smoke 가 `twine check` 와 함께 CI 에 자동화되지 않음. v0.5.8 백로그에 `release/checklist` 추가 검토.
- **mcp extra (`pip install standard-ai-workflow[mcp-sdk]`) 는 fresh venv 검증 안 함** — `mcp[cli]>=1.0` 의존성은 변경 없음.

## 7. 다음 릴리스

- **v0.5.8-beta** (예정, 변경 큼): Mavis 측 `choose_role`/`choose_roles`/`validate_fanin_output` 자동 wire (mavis-team `delegate_to_subagent` hook), Phase 11 case study, `task.required_model_tier` keyword 사전 외부화, contract v2 streaming/observability, packaging smoke 자동화 (위 6번 한계 해결).
