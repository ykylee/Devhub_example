# Skill: git-conflict-resolver

## Overview
Automated analysis and resolution of git merge conflicts. This skill helps the agent resolve complex 3-way merge conflicts by classifying conflict types and applying predefined or custom strategies.

## Key Features
- **Conflict Marker Analysis**: Detects and parses `<<<<<<<`, `=======`, and `>>>>>>>`.
- **Classification**: Identifies if a conflict is in a configuration file, documentation, or source code.
- **Resolution Strategies**:
    - `ours`: Keep changes from the current branch.
    - `theirs`: Use changes from the incoming branch.
    - `smart`: Attempt to combine non-overlapping changes (for docs/lists).
- **Validation**: Automatically runs syntax checks or tests after resolution.

## Input Contract
```json
{
  "repo_path": "string",
  "target_file": "string",
  "strategy": "ours|theirs|smart|analyze_only"
}
```

## Output Contract
```json
{
  "status": "ok|warning|error",
  "conflict_count": "number",
  "resolved_count": "number",
  "resolution_summary": "string",
  "unresolved_conflicts": "list[object]"
}
```


## v0.6.5 Stage Completion

본 skill 의 출력은 v0.6.5 부터 v0.6.4 의 [Stage Gate Pattern](../../../core/stage_gate_pattern.md) 의 `stage_completion` 필드를 포함한다.

| Field | 값 |
|---|---|
| `stage_name` | `git-conflict-resolver` |
| `next_stage` | `(workflow end)` |
| `approval_actor` | `user` mandatory (state 문서 영향) |
| `approval_timestamp` | ISO 8601 |

자세한 spec: [`core/stage_gate_pattern.md`](../../../core/stage_gate_pattern.md), [`core/output_schema_guide.md §3.4`](../../../core/output_schema_guide.md).
