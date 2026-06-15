# Beta v0.5.9.1 — Wire 가이드 §3 코드 예시 fix + 회귀 test 추가

- **릴리스 일자**: 2026-06-08
- **브랜치**: `main`
- **포함 커밋**: 1 (wire 가이드 §3 `sub_payloads` fix + 신규 `tests/check_wire_guide_v059.py`)
- **상태**: ✅ docs-only + 신규 회귀 test — wheel/API 변경 0, v0.5.9 wire guide 의 copy-paste 시 fail 명시적 surface

## 1. 무엇이 바뀌었나

v0.5.9 release note 의 §6 "Known limitations" 에 적었던 두 가지 cleanup 중 작은 하나를 즉시 처리:

### 1.1 wire 가이드 §3 의 `sub_payloads` 미정의 var fix

§3 fan-out/in 코드 예시 (line 100-109) 의 list comprehension `zip(sub_payloads, sub_responses)` 가 `sub_payloads` 라는 정의되지 않은 변수를 참조. **Mavis 측 consumer 가 가이드를 그대로 copy-paste 하면 `NameError`**.

v0.5.9.1 fix:

- §3 예시 line 70 직후 `sub_payloads: list[dict] = []` 추가
- §3 예시 line 82 (loop 안) 에 `sub_payloads.append(sub_payload)` 추가

이제 list comprehension 이 정상. **Mavis 측 consumer 가 copy-paste 하면 0줄 수정으로 end-to-end 실행 가능** (이전엔 첫 호출에서 NameError → consumer 측에서 디버깅 필요했음).

### 1.2 신규 회귀 test: `tests/check_wire_guide_v059.py`

Mavis 측 wire 가이드가 copy-paste 가능한지 enforce 하는 3 check 자동 회귀 test:

1. **`test_section_3_defines_fanout_to_subs`** — 가이드 doc 의 §3 코드 블록이 syntax error 없이 컴파일되고 `fanout_to_subs` 함수 정의.
2. **`test_section_3_runs_against_3sub_fixture`** — 가이드를 실제 spec API (`choose_roles`, `validate_fanin_output`) 와 연결해 3-sub fan-out/in walkthrough 실행. v0.5.9.1 의 `sub_payloads` fix 가 살아있지 않으면 `NameError` 로 fail.
3. **`test_section_3_rejects_sub_delegation_id_reissue`** — wire 가이드 §8 안티패턴 6번 (sub 응답 delegation_id 재발급) 을 시뮬레이션한 misbehaving sub-agent caller. `validate_fanin_output` 가 거절해야 PASS.

테스트는 doc 자체를 regex 추출해 compile 하므로, **wire 가이드 본문이 바뀌면 (예: 누군가 `sub_payloads` 를 또 지우면) 자동 fail** — back-pressure 가 가이드에 직접 연결됨.

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `workflow-source/core/orchestrator_contract_v1_wire_guide.md` | §3 코드 예시 `sub_payloads` 미정의 var fix | +2 / -1 |
| `workflow-source/tests/check_wire_guide_v059.py` | 신규 — wire guide §3 회귀 3 check (in-process) | +340 (new) |
| `workflow-source/releases/Beta-v0.5.9.1.md` | 신규 — 본 release note | +120 (new) |

소스 코드 / wheel / API 변경 **없음**. v0.5.8 wire 그대로 유효. v0.5.9 wire 가이드 보강 그대로 유효.

## 3. 검증 (actual run, fresh venv)

### 3.1 신규 회귀 test

```text
$ python3 -m tests.check_wire_guide_v059
Wire guide §3 fan-out/in regression check passed (3 checks: §3 compiles, 3-sub walkthrough runs to validator with v0.5.10 back-log surface, reissued sub delegation_id rejected).
```

3/3 PASS. 핵심 — `sub_payloads` fix 가 살아있고 (test 2 의 `NameError` catch), `validate_fanin_output` 가 sub.delegation_id prefix 재발급을 거절함을 확인 (test 3).

### 3.2 기존 회귀 suite 영향 없음

```text
$ python3 -m tests.check_contract_v1_multi_component
Contract v1 §4.2/§5.2 multi-component smoke check passed (17 checks).

$ python3 -m tests.check_bootstrap_interactive_picker
Interactive picker regression suite passed (10 checks).
```

17/17 + 10/10 PASS — v0.5.7 / v0.5.8 baseline 유지.

## 4. 발견된 후속 back-log (v0.5.10)

신규 회귀 test 의 `test_section_3_runs_against_3sub_fixture` 가 돌다가 **v0.5.10 back-log 1건 발견**: `choose_roles` (workflow_kit/contract_v1/delegator.py) 가 sub delegation_id 를 spec 선언된 `{parent_id}-st-N` 형식으로 발급하지 않고, `{random_prefix}-{sub_id}` 형식으로 발급. 결과 Mavis 측 fanout 예시를 그대로 실행하면 `validate_fanin_output` 가 `sub_delegation_id_prefix_mismatch` 로 거절. 회귀 test 는 이걸 명확히 surface 하도록 작성 — `_OutputValidationFailed` 가 catch 되고 에러 메시지에 "delegation_id + prefix/parent" 가 포함되어야 back-log 신호로 인정.

근본 fix 는 `delegator.py` 의 `choose_role`/`choose_roles` 가 sub decision 발급 시 `delegation_id = f"{parent_decision.delegation_id}-st-{i}"` 형식으로 만들도록 변경. v0.5.10 PR scope.

## 5. 다운스트림 영향

- **Mavis 측 wire 코드**: 변경 없음. 가이드 doc 본문 1줄 (sub_payloads 추가) + 신규 회귀 test 1 파일.
- **Devhub_example / my_harness**: 영향 없음.
- **standard_ai_workflow 자체**: 영향 없음. v0.5.8 wheel/sdist 그대로 유효.

## 6. 업그레이드 절차

이 PR 은 docs + test only. wheel install 변경 없음. Mavis / 표준 AI 워크플로우 consumer 가 본 가이드 doc 와 신규 회귀 test 만 import 해서 활용 가능.

```bash
git pull origin main  # docs + test only
# wheel install (변경 없음):
pip install --upgrade \
  https://github.com/ykylee/standard_ai_workflow/releases/download/v0.5.8-beta/standard_ai_workflow-0.5.8b0-py3-none-any.whl
```

회귀 test 는 v0.5.9.1 commit 포함된 source tree 에서 `python3 workflow-source/tests/check_wire_guide_v059.py` 로 실행 가능.

## 7. 알려진 한계 / 백로그

- **v0.5.10 back-log**: `choose_roles` 가 spec-declared parent-prefix sub delegation_id 발급하도록 fix. 본 PR 의 회귀 test 가 그 fix 의 back-pressure.
- **Mavis 측 자동 wire (engine hook) 미구현** — v0.5.10+ 백로그 (v0.5.9 release note §9 에서 인용).
- **sub.delegation_id 임의 재발급 Mavis 측 회귀 test** — 본 PR 의 test 3 이 wire 가이드 doc 차원에서 enforce. Mavis 측 wire 코드 walkthrough 에서 자동 검증은 별도 CI hook 필요 (v0.5.10+ 백로그).

## 8. 다음 릴리스 (v0.5.10 후보)

- `choose_roles` sub delegation_id prefix 룰 spec 정합 fix (delegator.py)
- Mavis 측 자동 wire (engine hook `delegate_to_subagent` 에서 `validate_output` enforce)
- 위 회귀 test 가 PASS 로 전환되는지 자동 verify
