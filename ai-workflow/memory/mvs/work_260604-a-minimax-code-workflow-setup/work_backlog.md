# Work Backlog — mvs/work_260604-a-minimax-code-workflow-setup

- Branch: `mvs/work_260604-a-minimax-code-workflow-setup`
- Agent: mvs (Mavis, model `MiniMax-M3`)
- Status: in_progress
- Updated: 2026-06-04

## In Progress

- [WB-07] 본 branch memory 4종 파일 작성 — **in_progress** (state.json/handoff/backlog/daily)
- [WB-08] Root session scratchpad init — **pending**
- [WB-09] 사용자 검토 + commit/PH 결정 — **pending**

## Done

- [WB-01] 브랜치 생성 + checkout — done
- [WB-02] `ai-workflow/minimax_code_workflow.md` 신규 — done
- [WB-03] `AGENTS.md` Mavis 전용 메모 — done
- [WB-04] `ai-workflow/MEMORY_GOVERNANCE.md` §0.5/§3.1 — done
- [WB-05] `ai-workflow/README.md` §2/§8 — done
- [WB-06] `docs/governance/worker_division.md` §6 변경 이력 — done

## Backlog (후속 sprint 후보)

### P1 — 본 sprint 직후 (사용자 confirm 후)

- **Mavis agent memory init** — 본 sprint 에서 학습한 cross-project 노트 append (예: "3-layer 메모리는 layer 1 project memory 안에 외부 worker branch memory 도 포함")
- **Root session scratchpad init** — 본 sprint 결과 요약 + 후속 todo
- **사용자 commit/PH 흐름 결정** — squash merge vs additional review

### P2 — Mavis 추가 tuning

- **`mvs/work_260604-b-minimax-mcp-tuning`** — Mavis MCP 서버 (matrix/cu/playwright/trash) 의 본 저장소 default 활성/비활성 정책 + OpenCode 글로벌 스니펫 (`ai-workflow/global-snippets/opencode/opencode.global.jsonc`) 와의 정합
- **`mvs/work_260604-c-minimax-skills-tuning`** — 본 저장소에서 Mavis 가 자주 load 할 skill 의 `~/.mavis/skills/` override

### P3 — Mavis-assisted 외부 worker 검증

- **5-워커 lint/script 의 `mvs/*` prefix 인식 검증** — opencode branch 명명 lint 가 `mvs/` 를 unknown prefix 로 표시하지 않는지 확인
- **5-워커 ↔ Mavis scratchpad cross-check** — 외부 worker 의 session_handoff.md 가 Mavis 의 cross-cut 정합 노트를 어떻게 consume 하는지 SOP

## 참고

- 단일 entry-point: [`../../../minimax_code_workflow.md`](../../../minimax_code_workflow.md)
- 메모리 정책: [`../../../MEMORY_GOVERNANCE.md`](../../../MEMORY_GOVERNANCE.md) §0.5/§3.1
- daily index: [`./backlog/2026-06-04.md`](./backlog/2026-06-04.md)
- root scratchpad: `$MAVIS_SCRATCHPAD`
