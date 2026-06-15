# Beta v0.5.10 — choose_roles sub delegation_id spec 정합 fix

- **릴리스 일자**: 2026-06-08
- **브랜치**: `main`
- **포함 커밋**: 1 (delegator.py spec 정합 + 회귀 test 강화 + version bump)
- **상태**: ✅ spec 정합 — fanout → fanin end-to-end walkthrough 가 검증된 wire contract 와 정합

## 1. 무엇이 바뀌었나

### 1.1 발견된 v0.5.7 spec 위반

v0.5.9.1 회귀 test (3-sub fan-out walkthrough) 가 다음 bug surface:

- `choose_roles` (workflow_kit/contract_v1/delegator.py) 가 sub delegation_id 를 spec 선언된 `{parent_delegation_id}-st-N` 형식이 아닌 `{random_base}-{sub_id}` 형식으로 발급 (v0.5.7.1 commit 부터 의도된 코드 경로).
- `validate_fanin_output` (workflow_kit/contract_v1/output_validator.py) 의 prefix check 는 `sub.delegation_id` 가 `parent.delegation_id` 로 시작해야 한다고 enforce.
- 결과: Mavis 측 fanout 예시 (wire guide §3) 를 그대로 실행하면 `validate_fanin_output` 가 `sub_delegation_id_prefix_mismatch` 로 거절.
- 회귀 test (check_contract_v1_multi_component.py 17 check) 가 fanout → fanin end-to-end 가 아닌 **fanin payload fixture 만 검증**해 PASS — 가짜 PASS (claim-only PASS vs actual verification).

### 1.2 spec 정합 fix

delegator.py 의 `choose_roles` 가 sub delegation_id 를 `f"{parent_delegation_id}-st-{i}"` (1-based) 형식으로 직접 발급. `_generate_delegation_id_with_suffix(sub_id)` 호출은 제거 (deprecated).

- `i` = 1-based sub_tasks index
- 형식이 spec 본체 (orchestrator_subagent_contract_v1.md §4.2 line 495-509) 와 회귀 test 정답 builder (`_ok_payload` line 46) 와 정합
- sub_id 가 `"a"`, `"st-1"`, `"core-build"` 등 어떤 user-defined 값이든 일관된 `st-N` suffix 사용 (sub_id 자체는 consumer 식별자, delegation_id 는 wire identifier — spec 이 둘 분리 권장)

### 1.3 회귀 test 강화 (3건)

| 회귀 test | 변경 |
| --- | --- |
| `check_contract_v1_multi_component.py::check_choose_roles_fanout_three_subs` | v0.5.10 spec 정합 check 추가: sub delegation_id 가 반드시 parent delegation_id 의 prefix 여야 함. v0.5.7.1 까지는 약한 check (`"st-1" in delegation_id`)만 있었음. |
| `check_wire_guide_v059.py::test_section_3_runs_against_3sub_fixture` | v0.5.10 fix 후 success path 로 전환 (이전엔 `_OutputValidationFailed` catch + v0.5.10 back-log 신호). 3-sub walkthrough 가 end-to-end PASS. |
| `check_wire_guide_v059.py::test_section_3_rejects_sub_delegation_id_reissue` | misbehaving sub-agent scenario 재정의: wire guide §3 가 sub 응답의 delegation_id 가 아닌 sub 입력의 `sub_decision.delegation_id` 를 사용함을 verify. v0.5.10 spec 정합 wire 의 의도된 동작. |

### 1.4 doc 정리

- `delegator.py` line 341-343 의 3줄 주석 ("parent prefix 안 씀 — parent delegation_id 가 바뀌어도 sub 가 추적 가능하도록") 가 spec 위반 의도였음. v0.5.10 fix 와 함께 spec 정합 설명으로 교체 (15줄 주석).
- `_generate_delegation_id_with_suffix` 함수 docstring 에 v0.5.10 deprecation 표시.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `workflow-source/workflow_kit/contract_v1/delegator.py` | choose_roles spec 정합 fix + deprecation 표시 | +20 / -3 |
| `workflow-source/tests/check_contract_v1_multi_component.py` | v0.5.10 spec 정합 check 추가 (parent prefix) | +9 / -1 |
| `workflow-source/tests/check_wire_guide_v059.py` | test_section_3_rejects_sub_delegation_id_reissue 재정의 | +25 / -22 |
| `workflow-source/pyproject.toml` | version 0.5.9.1-beta → 0.5.10-beta | +1 / -1 |
| `workflow-source/releases/Beta-v0.5.10.md` | 신규 release note | +120 (new) |

wheel/sdist v0.5.10b0 빌드.

## 3. 검증 (actual run, fresh venv)

### 3.1 회귀 test suite

```text
$ python3 -m tests.check_contract_v1_multi_component
Contract v1 §4.2/§5.2 multi-component smoke check passed
(17 checks: choose_roles single/3-sub + parent-prefix sub.delegation_id
/parent-reject/strict/sub-reject/schema-invalid/type-invalid;
validate_fanin_output ok/partial/failed aggregation, status mismatch,
parent mismatch, sub delegation_id prefix mismatch, missing sub field,
missing parent, artifact action+stats, invalid action, invalid stat type).
```

```text
$ python3 -m tests.check_wire_guide_v059
Wire guide §3 fan-out/in regression check passed (3 checks: §3 compiles,
3-sub walkthrough runs end-to-end with v0.5.10 spec 정합,
reissued sub delegation_id back-pressure intact).
```

```text
$ python3 -m tests.check_bootstrap_interactive_picker
Interactive picker regression suite passed (10 checks).
```

핵심 신호: `test_section_3_runs_against_3sub_fixture` 가 v0.5.10 fix 후 success path 로 전환. wire_guide §3 의 3-sub walkthrough 가 end-to-end PASS. v0.5.9.1 의 회귀 test 가 더 이상 `_OutputValidationFailed` catch 후 back-log 신호 안 함.

### 3.2 packaging smoke (v0.5.7.1 회귀 패턴 동일)

```text
$ python3 tools/check_packaging.py
{
  "wheel": "dist/standard_ai_workflow-0.5.10b0-py3-none-any.whl",
  "imported": [7 modules],
  "missing": [],
  "boot_lib_help_has_no_interactive": true,
  "result": "PASS"
}
```

fresh venv + `pip install dist/*.whl` (no -e) + 7 module import smoke + `--no-interactive` flag 노출 확인.

### 3.3 twine check

```text
$ twine check dist/*
Checking dist/standard_ai_workflow-0.5.10b0-py3-none-any.whl: PASSED
Checking dist/standard_ai_workflow-0.5.10b0.tar.gz: PASSED
```

## 4. 다운스트림 영향

### 4.1 Mavis 측 (Mavis 1.x, my_harness, Devhub_example)

- Mavis 측 fanout wire code (mavis-team engine 의 `delegate_to_subagent`) 가 sub-agent 호출 결과에서 delegation_id 를 읽지 않고, sub 입력의 `sub_decision.delegation_id` 를 fanin 에 그대로 옮기는 패턴이 spec 정합. v0.5.10 fix 가 이를 enforce.
- 이전 (v0.5.7.1~v0.5.9.1) Mavis 가 wire 가이드를 copy-paste 했다면 fanin 에서 `sub_delegation_id_prefix_mismatch` 거절. v0.5.10 부터 정상 동작.
- v0.5.7 회귀 test 의 `_ok_payload` (line 46) 형식 `f"{parent_id}-st-{i}"` 와 v0.5.10 fix 의 `choose_roles` 출력이 정확히 정합 — 기존 회귀 test fixture 가 그대로 fanout → fanin end-to-end walkthrough 에 사용 가능.

### 4.2 Devhub_example / my_harness / 기타 downstream

- `workflow_kit.contract_v1.choose_roles` API 시그니처 변경 없음. 반환값 형식만 spec 정합으로 정렬.
- Devhub_example backend / frontend: 영향 없음.
- my_harness: v0.5.10 spec 정합된 `choose_roles` 가 향후 Rust port 의 reference impl 로 사용 가능.

### 4.3 standard_ai_workflow 자체

- `workflow_kit/contract_v1/delegator.py` 본체 변경 (sub_id 발급 형식) 외 spec/스키마 변경 없음.
- v0.5.8 wheel 의 모든 동작 (bootstrap_lib, packaging smoke, picker) 그대로 유효.

## 5. 업그레이드 절차

```bash
pip install --upgrade \
  https://github.com/ykylee/standard_ai_workflow/releases/download/v0.5.10-beta/standard_ai_workflow-0.5.10b0-py3-none-any.whl
```

dev / local install:

```bash
cd workflow-source
pip install -U .
```

## 6. 알려진 한계 / 백로그

- **`_generate_delegation_id_with_suffix` 함수 본체는 살아있음** — deprecation docstring 만 추가. external caller 가 import 했을 수 있어 즉시 제거는 위험. v0.6.x 에서 제거.
- **Mavis 측 자동 wire (engine hook `delegate_to_subagent` 에서 `validate_output` enforce) 미구현** — v0.5.9 release note §9 에서 인용, v0.5.10 에서도 미구현. v0.5.11+ 백로그.
- **회귀 test 의 v0.5.10 back-pressure 가 새 fix 가 들어오면 자동 전환** — `test_section_3_runs_against_3sub_fixture` 의 success-path assertion 이 현재 PASS. 누군가 sub.delegation_id 발급 형식을 parent-prefix 가 아닌 random 으로 되돌리면 즉시 fail.

## 7. 다음 릴리스 (v0.5.11 후보)

- Mavis 측 engine hook `delegate_to_subagent` 에서 `validate_output` enforce
- `sub_delegation_id_prefix_mismatch` 외 spec 정합 자동 회귀 (e.g. `parent_delegation_id` 누락, sub_id unique-violation)
- `_generate_delegation_id_with_suffix` v0.6.x 제거 타임라인 확정
- (long-term) contract v2 streaming/observability
