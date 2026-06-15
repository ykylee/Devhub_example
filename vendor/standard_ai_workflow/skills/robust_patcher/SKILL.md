# Skill: robust-patcher

- 문서 목적: 로컬 LLM의 낮은 편집 정밀도를 보완하기 위해 Search-Replace 블록 기반의 퍼지 패치 기능을 제공한다.
- 범위: SEARCH/REPLACE 블록 파싱, 퍼지 매칭, 문법 검사
- 대상 독자: AI 에이전트, 개발자
- 상태: prototype
- 최종 수정일: 2026-05-03
- 관련 문서: `core/phase6_precision_editing_plan.md`

## 1. 개요
줄 번호나 정확한 공백 일치가 어려운 로컬 LLM 환경에서, 수정하고 싶은 코드의 **원래 모습(SEARCH)**과 **수정된 모습(REPLACE)**만 제공하면 지능적으로 위치를 찾아 수정합니다.

## 2. 사용 방법 (LLM 지침)
코드를 수정할 때 아래 형식을 사용하십시오.

```
<<<<<<< SEARCH
[수정하고 싶은 기존 코드 블록]
=======
[새롭게 변경될 코드 블록]
>>>>>>> REPLACE
```

## 3. 실행 예시
```bash
python3 skills/robust-patcher/scripts/patch_engine.py --file "src/main.py" --patch-file "patch.txt"
```

## 4. 주요 특징
- **Fuzzy Match**: 들여쓰기나 빈 줄 차이가 있어도 80% 이상 유사하면 패치를 적용합니다.
- **Auto-Validation**: 패치 후 Python 문법 에러가 발생하면 파일에 저장하지 않고 실패를 보고합니다.


## v0.6.5 Stage Completion

본 skill 의 출력은 v0.6.5 부터 v0.6.4 의 [Stage Gate Pattern](../../../core/stage_gate_pattern.md) 의 `stage_completion` 필드를 포함한다.

| Field | 값 |
|---|---|
| `stage_name` | `robust-patcher` |
| `next_stage` | `validation-plan` |
| `approval_actor` | `user` mandatory (state 문서 영향) |
| `approval_timestamp` | ISO 8601 |

자세한 spec: [`core/stage_gate_pattern.md`](../../../core/stage_gate_pattern.md), [`core/output_schema_guide.md §3.4`](../../../core/output_schema_guide.md).
