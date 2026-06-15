# Beta v0.5.7 — Contract v1 §4.2/§5.2 Multi-component Fan-out/In + §6.3 cross-ref Row

- **릴리스 일자**: 2026-06-08
- **브랜치**: `release/v0.5.7`
- **포함 커밋**: v0.5.6 (PR #20) 이후 1 squash
- **상태**: ✅ 1개 TASK 완료 (TASK-V057-001)

## 1. 하이라이트

v0.5.6 §5/§6 P0 enforcement 위에 v0.5.5 pilot P1/P2 priorities 2개 묶어서 enforce:

- **TASK-V057-001-A**: contract v1 §4.2 sub_task fan-out / §5.2 sub_result fan-in (1차 컷) — `delegator.choose_roles(task, strict=False) -> list[DelegationDecision]`, `output_validator.validate_fanin_output(payload, expected_parent_delegation_id=None) -> OutputValidationResult`
- **TASK-V057-001-B**: §6.3 "cross-ref 갱신" / "fan-in 통합 보고" row 추가 — delegator `MUST_NOT_DELEGATE_MARKERS` 7→9
- **TASK-V057-001-C (P2)**: `task.required_model_tier` 필드 + `delegator.recommend_model_tier(task)` 자동 main/small 결정
- **TASK-V057-001-D (P2)**: `result.artifacts[i].action` (created/modified/deleted) + `added/removed/total` 분리 validator
- **신규 회귀 1개**: `check_contract_v1_multi_component.py` (7 choose_roles + 10 validate_fanin = 17 check)
- **Mavis 측 wire 가이드**: [`orchestrator_contract_v1_wire_guide.md`](../core/orchestrator_contract_v1_wire_guide.md) (신규, ~250 줄) — Mavis / Mavis 가 `choose_role`/`choose_roles`/`validate_output`/`validate_fanin_output` 을 실제 위임 결정 경로에 묶는 패턴 + 안티패턴 5종

## 2. TASK-V057-001-A: §4.2/§5.2 Multi-component Fan-out/In

### 동기

v0.5.5 pilot 의 4 시나리오는 모두 **single-task 위임** (PR #490/491/492/493 각자 1 delegation). v0.5.5 P1 도출: **multi-component sub-task** (e.g. Devhub N+1 endpoint 1차 구현 = 3 sub-task fan-out + fan-in 통합) 가 contract v1 spec 에 없음. v0.5.6 P1 priorities 1번: contract v2 fan-out/in — v0.5.7 에서 1차 컷.

### 산출물

[`workflow_kit/contract_v1/delegator.py`](../workflow_kit/contract_v1/delegator.py) (확장, ~340 줄, pure Python)

신규 API:

```python
from workflow_kit.contract_v1 import choose_roles, validate_fanin_output, recommend_model_tier

# §4.2 fan-out: sub_tasks 별 role 결정 + parent-prefix delegation_id
decisions = choose_roles(task, strict=False)  # task.sub_tasks 가 있으면 fan-out
parent = decisions[0]
subs = decisions[1:]  # 각 sub: sub_id / delegation_id / role / model_tier

# §5.2 fan-in: parent 보고 검증 + sub_results 집계 일관성
result = validate_fanin_output(
    fanin_payload,
    expected_parent_delegation_id=parent.delegation_id,
)
if not result.is_valid:
    # aggregated status mismatch / sub delegation_id prefix mismatch / missing sub fields
    for err in result.errors:
        log(f"{err.field}: {err.reason}")

# P2: model_tier 자동 결정
tier = recommend_model_tier(task)  # "small" (default) / "main" (아키텍처, 정책, 5+ 파일 등)
```

검증 항목 (총 27 + 회귀 1개):

| 영역 | 항목 | 신규 |
| --- | --- | --- |
| `choose_roles` | sub_tasks 없으면 단일 결과 (`[choose_role(task)]`) | v0.5.7 |
| `choose_roles` | N sub × parent + 1 = N+1 decisions | v0.5.7 |
| `choose_roles` | parent §6.3 위반 → non-strict: must_not_delegate decision / strict: raise | v0.5.7 |
| `choose_roles` | sub §6.3 위반 → v0.5.7 spec §6.3: sub marker 도 parent reject | v0.5.7 |
| `choose_roles` | sub schema invalid (필수 필드 누락, 잘못된 task_type) → ValueError | v0.5.7 |
| `validate_fanin_output` | sub_results 의 status 집계 → parent status 일관성 | v0.5.7 |
| `validate_fanin_output` | parent_delegation_id 매치 (expected vs declared) | v0.5.7 |
| `validate_fanin_output` | sub_result.delegation_id 가 parent prefix 로 시작 | v0.5.7 |
| `validate_fanin_output` | sub_results 필수 필드 5개 (sub_id, delegation_id, status, summary, written_paths) | v0.5.7 |
| `recommend_model_tier` | main keyword (아키텍처, 정책, 5+ 파일, cross-cutting, 5+ source) 매치 → "main" | v0.5.7 |
| `recommend_model_tier` | 명시 `task.required_model_tier` → override | v0.5.7 |
| `choose_role.recommended_model_tier` | 자동 결정값이 decision 에 박힘 | v0.5.7 |

### 회귀

[`tests/check_contract_v1_multi_component.py`](../tests/check_contract_v1_multi_component.py) (신규, 17 check)

- choose_roles: 7 check (single/3-sub/parent-reject/strict/sub-reject/schema-invalid/type-invalid)
- validate_fanin_output: 10 check (ok/partial/failed aggregation, status mismatch, parent mismatch, sub delegation_id prefix mismatch, missing sub field, missing parent, artifact action+stats, invalid action, invalid stat type)

## 3. TASK-V057-001-B: §6.3 Cross-ref 갱신 / Fan-in 통합 Row 추가

### 동기

v0.5.6 P1 priorities 2번: "§6.1 cross-ref 갱신 명시 row 추가". v0.5.5 pilot 에서 `handoff / state.json / backlog` 3 marker 만으로는 부족. **문서 간 cross-link 수정** (e.g. workflow_agent_topology.md ↔ orchestrator_subagent_contract_v1.md ↔ memory layer) 과 **fan-in 통합 보고 작성** 도 sub-agent 가 처리하면 dead link / inconsistent integration 위험.

### 변경

| §6.3 row | 처리 주체 | 이유 (v0.5.7 추가) |
| --- | --- | --- |
| cross-ref 갱신 (docs/ ↔ memory layer ↔ handoff 간 링크 수정) | orchestrator | cross-ref 는 단일 진실 공급원(orchestrator) 만이 일관성 보장 가능; sub-agent 가 부분 갱신 시 dead link 위험 |
| fan-in 결과 통합 보고 작성 | orchestrator | sub_results 통합은 §6.3 의 "sub-agent 출력 통합/리뷰" 의 멀티 컴포넌트 버전; sub-agent 가 자기 sub 만 보고, 통합은 orchestrator |
| parent_delegation_id 발급 | orchestrator | fan-out 의 부모 ID 는 fan-out 결정의 일부, orchestrator 만 발급 |

### Delegator 변경

`MUST_NOT_DELEGATE_MARKERS` 7→9:

```python
MUST_NOT_DELEGATE_MARKERS = (
    "handoff",
    "backlog",
    "state.json",
    "ask_user",
    "우선순위",
    "통합/리뷰",
    "PR 본문",
    # v0.5.7 신규
    "cross-ref",   # cross-ref 갱신은 orchestrator 직접 처리
    "fan-in 통합", # fan-in 통합 보고는 orchestrator 직접 처리
)
```

매치 검사도 확장: `sub_tasks[].brief` 까지 haystack 에 포함 → fan-out 결정 시 sub 의 §6.3 marker 도 parent reject (sub 가 fan-out 의 일부이므로 정합성).

## 4. TASK-V057-001-C (P2): `required_model_tier` 자동 결정

### 동기

v0.5.5 pilot 의 4 시나리오 중 PR #492 (N:M weight, multi-file) 만 main 승격 (나머지 small). v0.5.5 P2 도출: model_tier 결정을 sub-agent 가 아니라 orchestrator 가 자동 enforce.

### 변경

- §4.1 input schema 에 `task.required_model_tier` (optional, "small" / "main") 추가
- `delegator.recommend_model_tier(task)` 신규 API
- main 후보 keyword 5개 (case-insensitive substring match against `task.brief` / `constraints` / `expected_outputs.must_include`):
  - "아키텍처" (architecture)
  - "정책" (policy)
  - "5+ 파일" (5+ file cross-cutting)
  - "cross-cutting"
  - "5+ source" (bounded_research 5+ source)
- 명시 `task.required_model_tier` 있으면 override
- `choose_role` 결과의 `decision.recommended_model_tier` 필드에 자동 결정값 박힘

## 5. TASK-V057-001-D (P2): `artifacts` Action + Stats 분리

### 동기

v0.5.5 pilot 의 PR #492 (N:M weight) 가 multi-file 변경. 단순 `path:kind` 만으로는 변경 통계 파악 어려움. P2 도출: `action` (created/modified/deleted) + `added/removed/total` 분리.

### 변경

`result.artifacts[i]` (optional, validated if present):

| 필드 | 타입 | enum | 설명 |
| --- | --- | --- | --- |
| `action` | enum | `created` / `modified` / `deleted` | 변경 종류 (v0.5.7 신규) |
| `added` | int | - | 추가된 줄 수 (v0.5.7 신규) |
| `removed` | int | - | 삭제된 줄 수 (v0.5.7 신규) |
| `total` | int | - | 최종 줄 수 (v0.5.7 신규) |

기존 `path` / `kind` / `lines` 는 그대로 (backward compatible).

## 6. Mavis 측 wire 가이드 (v0.5.7 신규)

[`core/orchestrator_contract_v1_wire_guide.md`](../core/orchestrator_contract_v1_wire_guide.md) (신규, ~250 줄)

핵심 패턴:

```python
# 1) Single-task 위임
decision = choose_role(task)
if decision.must_not_delegate:
    raise DelegationRejected(decision)
response = sub_agent_caller(payload)
validate_output(response, expected_delegation_id=decision.delegation_id)

# 2) Fan-out + Fan-in (v0.5.7)
decisions = choose_roles(task, strict=True)
parent = decisions[0]
subs = decisions[1:]
# ... sub-agent 병렬 호출
fanin_payload = build_fanin(parent, sub_responses)
validate_fanin_output(fanin_payload, expected_parent_delegation_id=parent.delegation_id)

# 3) Mavis 측 안티패턴 5종
# - choose_role 호출 없이 sub-agent 직접 spawn
# - sub-agent 가 fan-in 보고서 작성
# - sub_results 무시하고 통째 합치기
# - expected_delegation_id 없이 validate_output 호출
# - model_tier 를 sub-agent 가 자기 결정
```

## 7. 회귀 테스트 종합

| 테스트 | 결과 |
| --- | --- |
| `check_bootstrap.py` (8 모드) | ⚠️ env 의존 (main 에서도 동일하게 fail; v0.5.7 범위 밖) |
| `check_bootstrap_mcp_roundtrip.py` (5 하네스) | ✅ 5/5 |
| `check_docs.py` | ✅ 111 markdown PASS |
| `check_contract_v1_roundtrip.py` (S1) | ✅ |
| `check_contract_v1_role_mapping.py` (S2) | ✅ |
| `check_contract_v1_direct_only.py` (S3) | ✅ (9 MUST-NOT-delegate [7 v0.5.6 + 2 v0.5.7]) |
| `check_contract_v1_output_validator.py` (v0.5.6 신규) | ✅ (4 pilot + 8 violation + 4 pilot doc) |
| `check_contract_v1_delegator.py` (v0.5.6 확장) | ✅ (4 mapping + 9 rejection + strict + 4 model_tier) |
| **`check_contract_v1_multi_component.py` (v0.5.7 신규)** | ✅ (7 choose_roles + 10 validate_fanin) |
| `check_pilot_phase11_contract_v1.py` (v0.5.5) | ✅ |
| `check_workflow_linter.py` | ⚠️ env 의존 (main 에서도 동일) |

**v0.5.7 신규 회귀 1개 + v0.5.6 회귀 2개 (delegator 확장, direct_only 9개로) + v0.5.5/v0.5.4 회귀 그대로 PASS** = 회귀 7/7 + 111/111 docs.

## 8. spec 갭 발견 (v0.5.7)

`multi_component` 회귀 실행 중 1개 spec 갭 empirical 발견:

| 필드 | Before | After |
| --- | --- | --- |
| `sub_result` 의 status (v0.5.7 spec §5.2) | 미정의 | `ok` / `partial` / `failed` (parent status 와 동일 enum) |
| `parent_delegation_id` 위치 (v0.5.7 spec §5) | §4 input 만 | §5 output top-level 에 추가 (sub_results 있으면 필수) |

회귀 자체에는 영향 없음 (enum 동일하므로), 다만 spec doc 에 명시 추가.

## 9. v0.5.0 → v0.5.7 변경 통계

```
v0.5.0 (PR #13): 42/42 smoke + MiniMax Code harness overlay
v0.5.1 (PR #14): per-harness MCP install + auto-emit + guide
v0.5.1 (PR #15): check_bootstrap_mcp_roundtrip + 5 harnesses round-trip
v0.5.2 (PR #16): bootstrap 풀 리팩터 + 패키지화 + pilot validation
v0.5.3 (PR #17): antigravity MCP 표준화 + cross-language stack 표시
v0.5.4 (PR #18): orchestrator ↔ sub-agent contract v1 + maturity sync + baseline sync
v0.5.5 (PR #19): Phase 11 본격 pilot (Devhub Example × Contract v1)
v0.5.6 (PR #20): contract v1 §5/§6 P0 enforcement (validator + delegator)
v0.5.7 (PR #21): contract v1 §4.2/§5.2 multi-component fan-out/in + §6.3 cross-ref row
```

총 9 PR.

## 10. 다음 단계 (v0.5.8 후보)

- [ ] Mavis 측 `choose_role`/`choose_roles` 을 mavis-team engine 의 `delegate_to_subagent` hook 에 자동 wire (가이드 v0.5.7 기반)
- [ ] **배포 채널 정책 변경 (2026-06-08)**: PyPI/TestPyPI 업로드 폐기, **GitHub Releases** 에 wheel/sdist attach 만. `releases/Beta-v0.5.7.md` 가 release notes 본문.
- [ ] Phase 11 case study 추가 — v0.5.7 fan-out/in 시나리오 walkthrough (Devhub N+1 endpoint 4 동시 처리)
- [ ] contract v2 스트리밍 출력 / 양방향 ping / observability (v0.5.7 §11 1차 컷 외 항목)
- [ ] `task.required_model_tier` 의 keyword 사전 외부화 (config 기반)

## 11. 메모리 layer

v0.5.7 의 모든 작업은 `ai-workflow/memory/release/v0.5.7/` 에 기록:
- `PROJECT_PROFILE.md` — v0.5.7 프로젝트 프로파일 (choose_roles/validate_fanin/recommend_model_tier/wire_guide 추가)
- `session_handoff.md` — 세션 인계
- `backlog/2026-06-08.md` — TASK 상세
- `state.json` — workflow 상태 캐시

## 12. 이슈 트래커

- ✅ [issue #1](https://github.com/ykylee/standard_ai_workflow/issues/1) 영구 해결 (v0.5.4)
- 0 open issues
