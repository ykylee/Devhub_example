# Session Handoff — mvs/work_260604-a-minimax-code-workflow-setup

- Branch: `mvs/work_260604-a-minimax-code-workflow-setup`
- Agent: mvs (Mavis, model `MiniMax-M3`)
- Runtime: MiniMax Code (root session `mvs_8952f7f57f9749a68171434a78f89960`, current `mvs_d882729c69274dd287d0933213d08e3b`)
- Updated: 2026-06-04

## 🎯 Current Focus

Mavis (MiniMax Code) 가 본 저장소에서 움직이는 운영 패턴을 단일 source-of-truth 로 set up. 본 sprint 의 산출물 = (1) Mavis 운영 entry-point 문서 신규, (2) 3-layer 메모리 매핑 정책, (3) cross-ref 4 위치, (4) 본 branch memory 4종 + root scratchpad init. 후속 sprint 는 Mavis cross-cut 정합 또는 외부 worker 분담 영역의 Mavis-assisted 검증.

## 📊 Work Status

- [WB-01] 브랜치 생성 (`mvs/work_260604-a-minimax-code-workflow-setup`): done
- [WB-02] `ai-workflow/minimax_code_workflow.md` 신규 (~250 lines, 10 sections): done
- [WB-03] `AGENTS.md` "Mavis 전용 메모" 섹션 신설: done
- [WB-04] `ai-workflow/MEMORY_GOVERNANCE.md` §0.5 + §3.1 추가: done
- [WB-05] `ai-workflow/README.md` §2 mvs/ branch memory 링크 + §8 신설: done
- [WB-06] `docs/governance/worker_division.md` §6 변경 이력 row 추가: done
- [WB-07] 본 branch memory 4종 파일 (state.json / handoff / backlog / daily): in_progress
- [WB-08] Root session scratchpad init (`$MAVIS_SCRATCHPAD`): pending
- [WB-09] 사용자 검토 + commit/PH 결정 대기: pending

### 적용된 변경 (5 files, +350+ lines)

- `ai-workflow/minimax_code_workflow.md` — **신규**. §0 핵심 정의 / §1 세션 모델 / §2 작업 라우팅 / §3 메모리 3-layer / §4 communication+cron / §5 도구(skills/agents/MCP + hard limits) / §6 브랜치·PR / §7 인계 SOP / §8 day-1 baseline / §9 5-워커 워크플로우와의 관계 / §10 변경 이력
- `ai-workflow/MEMORY_GOVERNANCE.md` — §0 prefix 예시 `mvs/` 행 추가, **§0.5 신규** (3-layer 매핑 표 + 중복 방지 규칙 + session-start/end 순서 + 외부 worker 경계), **§3.1 신규** (Mavis 행동 지침)
- `ai-workflow/README.md` — §2 mvs/ branch memory 행 추가, §8 신규, "다음에 읽을 문서"에 Mavis 운영 entry-point 1행 추가
- `AGENTS.md` — "Mavis (MiniMax Code / MiniMax-M3) 전용 메모" 섹션 신설 (OpenCode 섹션 직후, 동일 패턴)
- `docs/governance/worker_division.md` — §6 변경 이력 row 추가 (2026-06-04 / Mavis 오케스트레이션 레이어 / `mvs/work_260604-a-...`)

## ⏭️ Next Actions

- 본 sprint 종료 직후:
  1. **사용자 commit/PH 결정** — 본 branch 6 files 변경에 대한 squash merge vs additional review 분기
  2. **cross-worker 검증** (선택) — 외부 worker (opencode 등) 의 branch memory 가 본 sprint 의 `mvs/*` prefix 와 충돌하지 않는지 점검
  3. **Mavis agent memory init** (선택) — 본 sprint 에서 학습한 cross-project 노트 (예: "3-layer 메모리는 layer 1 project memory 안에 외부 worker branch memory 도 포함한다") 를 `~/.mavis/memory/agents/mavis.md` 에 append
- 후속 sprint 후보 (우선순위 사용자 결정 대기):
  1. **`mvs/work_260604-b-minimax-mcp-tuning`** — OpenCode 글로벌 스니펫 + Mavis MCP 매핑 (matrix/cu/playwright/trash 의 Mavis context 내 default 활성/비활성 정책)
  2. **`mvs/work_260604-c-minimax-skills-tuning`** — Mavis 가 본 저장소에서 자주 load 할 skill 의 `~/.mavis/skills/` override (있으면)

## ⚠️ Risks & Blockers

- **5-워커 워크플로우와의 경계**: §7.4 cross-worker PR 패턴이 외부 worker 의 자율성을 침범하지 않도록 *권장* 수준으로 유지 (강제 X). 본 sprint 의 §6 변경 이력 row 에 "본 문서 자체는 Mavis 와 직접 무관" 명시.
- **`mvs/` prefix 인식**: 외부 worker (opencode 등) 의 lint/script 가 `mvs/*` prefix 를 인식 못 할 가능성. 본 sprint 범위 외 — 후속 검증 필요.
- **scratchpad 경로**: `$MAVIS_SCRATCHPAD` 환경변수가 branch 세션에서 자동 주입되는지 미검증. 본 sprint 에서 scratchpad init 시도 후 동작 확인.
- **agent/user memory**: layer 2/3 파일은 본 sprint 에서 신규 작성하지 않음 (학습된 cross-project/user 노트가 없으므로). 사용자 프로파일 (<user_profile_missing>) 도 자연스러운 대화로 채울 예정.
