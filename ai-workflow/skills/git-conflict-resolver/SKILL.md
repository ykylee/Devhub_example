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
