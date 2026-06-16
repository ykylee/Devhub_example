# Session Handoff — opencode/work_260604-e-rbac-e2e-followup

- Branch: `opencode/work_260604-e-rbac-e2e-followup`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04
- Sprint: N-10 P1 follow-up (release_v1_roadmap §3.5) — RBAC E2E 6 TC 중 frontend E2E 4건 구현

## 🎯 Current Focus

docs/domain/rbac-permissions/test_cases.md 의 7 TC 중 frontend E2E 권장 4건 구현:
- TC-RBAC-LOGOUT-02 (FE signout → logout API)
- TC-RBAC-ROW-READ-01 (List read scope 필터)
- TC-RBAC-ROW-READ-02 (Get read scope 차단)
- TC-RBAC-CODE-01 (거부 코드 표준화)

잔여 3건은 scope out (각주):
- TC-RBAC-LOGOUT-01 — backend IT, 별도 sprint
- TC-RBAC-ROLE-DRIFT-01 — backend IT (Keycloak drift 환경 의존)
- TC-RBAC-TRACE-01 — process/review (본 PR 의 spec 주석으로 입증)

## 📊 Work Status

- [WB-01] 브랜치 + memory 디렉터리 set up: done
- [WB-02] spec 분석 + seed data 확인: done
- [WB-03] rbac-data-scope.spec.ts 작성 (4 TC): in_progress
- [WB-04] vitest + next build 검증: planned
- [WB-05] E2E 로컬 dry-run (Playwright): planned
- [WB-06] 커밋 + push + PR + state/handoff final: planned

## ⏭️ Next Actions

- 본 sprint 종료 후 (PR 머지 후):
  1. N-6 v1.0 staging 1주 운영 (사내 동반, 사용자)
  2. P1-6 Sign-out endpoint backend (Claude)
  3. P2 design tension fix (pathRequiresSystemAdmin ↔ team_manager.organization:edit)
  4. P2 actorCanReadApplication 5000 limit 정공법

## ⚠️ Risks & Blockers

- seed data 가 미니멀 (project 1개, charlie 소유) — alice 의 member project 가 없음. List test 는 "0개 project" 가 기대값. 추후 다중 membership seed 가 필요할 수 있음.
- backend handler 가 200 + 403 의 어느 쪽을 반환하는지에 따라 FE expectation 다름. E2E 실행 결과 보고 후 필요시 조정.
