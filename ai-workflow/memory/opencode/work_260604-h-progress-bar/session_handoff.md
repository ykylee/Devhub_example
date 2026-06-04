# Session Handoff — opencode/work_260604-h-progress-bar

- Branch: `opencode/work_260604-h-progress-bar`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Sprint: project milestones 페이지 progress bar 실제 task 완료율 반영

## 🎯 Current Focus

`frontend/app/(dashboard)/projects/page.tsx:172-177` 의 progress bar 가 status 기반 hardcoded width (active=66.7%, closed=100%, else=25%) — "using status as a proxy" 주석 명시. 실제 project 진행 상태와 불일치.

**수정 방안** (사용자 확정):
- `getProjectTasks(projectId, ["todo", "in_progress", "review", "done"])` 를 `Promise.all` 로 병렬 fetch
- `done / total * 100` 계산
- task 0개 → neutral state (빈 바 + "No tasks" text)
- loading state 분리 (project list render 블로킹 X)

## 📊 Work Status

- [WB-01] 브랜치 + memory 디렉터리 set up: done
- [WB-02] projects/page.tsx 수정 (taskProgress state + Promise.all + 동적 렌더): planned
- [WB-03] vitest 추가 (progress 계산 unit test): planned
- [WB-04] tsc + vitest 검증: planned
- [WB-05] 커밋 + push + PR: planned

## ⏭️ Next Actions

- 본 sprint 종료 후 (PR 머지 후):
  1. **#1 Gitea issue sync (backend)** — Claude 별도 sprint
  2. **Codex P2 잔여** — `scm_providers` ↔ `integration_providers` catalog 정합
  3. **v1.0 N-6 staging 운영** + P1-6/P2 carry-over
  4. **Application progress_percent** — backend data_gap 별도 처리 (story-point 기반)

## ⚠️ Risks & Blockers

- N+1 API call: project 20개면 20번 task fetch. `Promise.all` 로 병렬화해도 backend 부하. 향후 backend batch endpoint 추가 시 migrate.
- task 가 많으면 (100+) 응답 payload 큼. 현재 schema 에 pagination 없음 — 별도 최적화.
- `getProjectTasks` 의 default status filter 가 `["todo", "in_progress", "review"]` (done 제외) — 본 sprint 는 명시적으로 all 4 status 전달.

## 📁 Key Files

- `frontend/app/(dashboard)/projects/page.tsx` (변경 대상)
- `frontend/domain/application-lifecycle/service/project.service.ts:176` (`getProjectTasks` API)
- `frontend/domain/application-lifecycle/schema/project.types.ts:110` (`ProjectTaskItem` 타입)
