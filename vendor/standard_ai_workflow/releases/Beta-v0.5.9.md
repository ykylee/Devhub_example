# Beta v0.5.9 — Wire 가이드 보강: sub.delegation_id parent prefix 룰 명시

- **릴리스 일자**: 2026-06-08
- **브랜치**: `main`
- **포함 커밋**: 1 (docs-only, wire 가이드 §7/§8/§9 보강)
- **상태**: ✅ docs-only — 신규 기능 0, 회귀 0, v0.5.8 wire contract 정합 보강

## 1. 무엇이 바뀌었나

v0.5.7.1 hotfix 검증 중 발견한 side note: **`sub_results[].delegation_id` 가 parent `delegation_id` 의 prefix 여야 spec 정합** 한데, Mavis 측 wire 가이드의 어느 섹션에도 명시되지 않았음. spec 본체 (`contract_v1` §4.2/§5.2) 와 회귀 테스트 (`check_contract_v1_multi_component.py` `check_validate_fanin_sub_delegation_id_prefix_mismatch`) 에는 박혀 있어도, **Mavis 측 wire 가이드에는 없었음**. 결과:

- Mavis 가 sub-agent 응답을 fan-in 보고서로 옮길 때 임의로 `delegation_id` 를 재발급하면 (UUID 새로 박거나 prefix 를 떼거나) `validate_fanin_output` 가 `sub_delegation_id_prefix_mismatch` 로 거절.
- 근데 가이드에 룰이 없어서 Mavis 측에서 그 행위를 사전에 방지할 단서 부재.

v0.5.9 (이 PR) 에서 가이드 doc 만 보강:

- `## 7. v0.5.7 신규 wire 포인트 요약` 표에 **sub.delegation_id prefix rule (v0.5.9 보강)** 1행 추가. 형식 명시: `{parent_delegation_id}-st-{i}` (e.g. `del-2026-06-08-c6cc8da7-st-1`).
- `## 8. 안티패턴` 에 6번 항목 추가 — "sub 응답의 `delegation_id` 를 orchestrator 가 임의 재발급" 하면 spec 정합 실패.
- `## 9. 다음 단계` 의 stale note 정리 — "v0.5.8 에서 PyPI 배포" → "v0.5.8 (Beta) — GitHub Releases 만 배포 (PyPI 미배포 확정)" (너 정책 memory 참조).

## 2. 변경 diff 요약

`workflow-source/core/orchestrator_contract_v1_wire_guide.md` — 1 file / +5 / -2

- §7 표: `choose_roles` 행에 parent-prefix 형식 명시 + 신규 1행 (sub.delegation_id prefix rule)
- §8 안티패턴: 6번 항목 추가 (sub 응답 delegation_id 재가공 금지)
- §9 다음 단계: v0.5.8 PyPI stale note → GitHub Releases 정책 + v0.5.9 (본 PR) 항목 추가

소스 코드 / API / wheel 변경 **없음**. v0.5.8 의 모든 동작 그대로.

## 3. 검증

코드 변경 0 이라 회귀 test 재실행 불필요. docs 자체 cross-check:

| 항목 | 상태 |
| --- | --- |
| `choose_roles` 형식 표기 (가이드 vs `_ok_payload` 정답 builder) | 일치 — `{parent_id}-st-{i}` (`check_contract_v1_multi_component.py:46` 의 `f"{parent_id}-st-{i}"`) |
| `validate_fanin_output` 에러 메시지 표기 | 일치 — `sub_delegation_id_prefix_mismatch` (`output_validator.py` 의 에러 코드) |
| `docs/RELEASE.md` reference (root 경로) | 유효 — v0.5.7.1 hotfix commit (8e5cd7b) 에서 추가된 root `docs/RELEASE.md` 가 PyPI 미배포 정책 명시 |
| v0.5.8 wheel sdist 동작 영향 | 없음 — docs 만 변경 |

## 4. 다운스트림 영향

- **Mavis 측 wire 코드**: 변경 없음. Mavis 가 본 가이드 doc 만 다시 읽으면 됨 (v0.5.8 wire 가이드 baseline 에 2줄 추가 — spec 정합 룰).
- **Devhub_example / my_harness**: 영향 없음. 둘 다 v0.5.7.1 의 회귀 test 가 정답 builder (`_ok_payload`) 를 갖고 있어, prefix 룰 자동 enforce.
- **standard_ai_workflow 자체**: 영향 없음. v0.5.8 wheel/sdist 그대로 유효.

## 5. 업그레이드 절차

이 PR 은 docs-only. wheel install 변경 없음. Mavis / 표준 AI 워크플로우 consumer 가 본 가이드 doc 만 새로 import 해서 읽으면 됨.

```bash
git pull origin main  # docs only
# 또는 wheel install (변경 없음):
pip install --upgrade \
  https://github.com/ykylee/standard_ai_workflow/releases/download/v0.5.8-beta/standard_ai_workflow-0.5.8b0-py3-none-any.whl
```

## 6. 알려진 한계 / 백로그

- **Mavis 측 자동 wire (`choose_role` / `validate_output` hook) 미구현** — §9 다음 단계로 유지. mavis-team engine 의 `delegate_to_subagent` hook 에서 `validate_output` enforce 하는 작업.
- **Mavis 측 sub.delegation_id 임의 재발급 방지를 test 로 enforce** — 본 가이드 doc 의 룰이 consumer 행동에 enforce 되도록 `tests/check_pilot_phase11_contract_v1.py` 에 회귀 test 추가 검토. v0.5.10 백로그.
- **§3 의 fan-out/in 코드 예시 (line 53-125) 의 `sub_payloads` reference 가 stale** — line 108 의 `zip(sub_payloads, sub_responses)` 가 위에서 정의되지 않은 이름 참조. 별도 cleanup PR 검토 (v0.5.10 백로그).

## 7. 다음 릴리스 (v0.5.10 후보)

- Mavis 측 자동 wire (engine hook 에서 `choose_role` / `validate_output` enforce)
- §3 코드 예시 stale variable 정리
- sub.delegation_id 임의 재발급 Mavis 측 회귀 test 추가
- (long-term) v0.5.8 의 `bootstrap_lib` picker 가 `minimax-code` 외 harness 들에도 multi-step UX 적용 검토 (v0.5.8 은 single-pass 만)
