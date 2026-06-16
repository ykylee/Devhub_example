# Session Handoff — opencode/work_260604-d-fix-projects-list-aggregation

- Branch: `opencode/work_260604-d-fix-projects-list-aggregation`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 🎯 Current Focus

Frontend `/projects` 페이지에서 신규 생성 project 가 list 에 표시되지 않는 결함 fix. `listAllProjects` 가 repos 만 순회 + ListProjects 410 GONE 의 이중 결함. `listStandaloneProjects` + per-application 통합 + dedup 으로 재작성.

## 📊 Work Status

- [WB-01] 브랜치 + memory 디렉터리 set up: done
- [WB-02] listAllProjects 재작성: in_progress
- [WB-03] projects/page.tsx 갱신: planned
- [WB-04] listAllProjects 테스트 재작성: planned
- [WB-05] frontend UT 실행: planned
- [WB-06] 커밋 + push + PR: planned

## ⏭️ Next Actions

- 본 sprint 종료 후:
  1. v1.0 D-11 안: P1 follow-up (E2E 6 TC 구현) — 별도 sprint
  2. v1.0 D-11 안: N-6 staging 운영 — 사용자 동반
  3. v1.0 후: Lane 1 housekeeping (role-access §2.1 표 정리 등)

## ⚠️ Risks & Blockers

- 본 fix 는 frontend only. backend 변경 없음.
- listAllProjects 의 시그니처 변경 (`repositoryIds: number[]` 제거) 으로 호출처 (`projects/page.tsx`) 동시 갱신 필수.
- 테스트 4건 모두 갱신 — 기존 의미 (concatenate per-repo) 폐기.
