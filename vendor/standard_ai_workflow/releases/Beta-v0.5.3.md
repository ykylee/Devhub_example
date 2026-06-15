# Beta v0.5.3 — antigravity MCP 표준화 + cross-language stack 표시

- **릴리스 일자**: 2026-06-07
- **브랜치**: `release/v0.5.3`
- **포함 커밋**: v0.5.2 (#16) 이후 1 squash
- **상태**: ✅ 2개 TASK 모두 완료

## 1. 하이라이트

v0.5.2 Beta 노트의 "다음 단계" 에서 미룬 2개 항목을 묶어서 처리:

- **TASK-V053-001**: antigravity MCP config 경로를 dot-dir 컨벤션으로 통일
- **TASK-V053-002**: cross-language 프로젝트 (Python+Go+Node 같은 다중 언어) 도 모든 감지 스택이 manifest / handoff / assessment 에 표시

## 2. TASK-V053-001: antigravity MCP config 경로 표준화

### 동기

5개 하네스 중 4개가 dot-dir 패턴 (`.codex/`, `.gemini/`, `.MiniMax/`, `.antigravity/`)을 따르는 게 자연스러운데, antigravity 만 `<root>/antigravity.mcp.json` (suffixed, project root) 로 발행 중. 일관성 회복.

### 변경

| Before | After |
|--------|-------|
| `<root>/antigravity.mcp.json` | `<root>/.antigravity/mcp.json` |

### 영향받은 파일

- `bootstrap_lib/mcp.py` — `write_mcp_config_files` 의 antigravity 경로 + docstring
- `harnesses/antigravity/apply_guide.md` — 4개 참조
- `core/mcp_installation_by_harness.md` — 2개 참조
- `tests/check_bootstrap.py` — expected suffix 단언문

### 사용자 영향

- 이미 `antigravity.mcp.json` 을 사용 중이라면 수동으로 `.antigravity/mcp.json` 으로 마이그레이션 필요
- 글로벌 `~/.antigravity/mcp.json` copy 패턴은 동일

## 3. TASK-V053-002: cross-language stack 표시

### 동기

Devhub_example pilot 에서 cpp+go+node+python 4개 언어가 감지됐는데, manifest 와 README 에는 `primary_stack: cpp` 한 줄만 노출. 다중 언어 프로젝트에서 다른 스택들이 보이지 않아 단일 언어로 오해 가능.

### 변경

#### 3.1 manifest (`build_manifest`)

신규 필드:
```json
{
  "primary_stack": "go",
  "stack_labels": ["go", "node", "python"],
  "multi_stack": true
}
```

#### 3.2 handoff (`render_session_handoff`)

기존 (단일 스택):
> Existing codebase onboarding completed; inferred primary stack: go.

신규 (다중 스택):
> Existing codebase onboarding completed; inferred primary stack: go; all detected stacks: go, node, python.

#### 3.3 assessment (`render_assessment`)

기존부터 `stack_labels` 섹션이 있었지만 강조 부족. 이제 manifest 와 일관되게 노출됨.

### 신규 테스트

`check_multi_stack_detection`:
- 3개 언어 fixture (package.json + pyproject.toml + go.mod) 로 multi_stack=True 검증
- manifest 의 `stack_labels` 에 3개 모두 포함 검증
- assessment 에 stack 라벨들이 표시되는지 검증

## 4. 회귀 테스트

| 테스트 | 결과 |
|--------|------|
| `check_bootstrap.py` (8 모드) | ✅ 8/8 (기존 7 + 신규 multi_stack) |
| `check_bootstrap_mcp_roundtrip.py` (5 하네스) | ✅ 5/5 |
| `check_docs.py` | ✅ 99 markdown PASS |
| Linter | ✅ status ok, 0 issues |

## 5. 메모리 layer

v0.5.3 의 모든 작업은 `ai-workflow/memory/release/v0.5.3/` 에 기록:
- `PROJECT_PROFILE.md` — v0.5.3 프로젝트 프로파일
- `session_handoff.md` — 세션 인계 (TASK-V053-001/002 done)
- `backlog/2026-06-07.md` — TASK 상세
- `state.json` — workflow 상태 캐시

## 6. 다음 단계 (v0.5.4 후보)

- [ ] mini-coder-max / general agent 통합 검증 (외부 contract 명시 필요)
- [ ] Phase 12 / 실전 튜닝 (Devhub_example 결과 기반 워커 프롬프트 정밀도)
- [ ] PyPI 배포 (보류 중)

## 7. v0.5.0 → v0.5.3 변경 통계

```
v0.5.0 (PR #13): 42/42 smoke + MiniMax Code harness overlay
v0.5.1 (PR #14): per-harness MCP install + auto-emit + guide
v0.5.1 (PR #15): check_bootstrap_mcp_roundtrip + 5 harnesses round-trip
v0.5.2 (PR #16): bootstrap 풀 리팩터 + 패키지화 + pilot validation
v0.5.3 (PR #17): antigravity MCP 표준화 + cross-language stack 표시
```

총 5 PR.
