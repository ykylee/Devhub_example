# Work Backlog — feat/work_260611-a-wiki-ingest-from-raw

- 문서 목적: wiki-ingest-from-raw skill 본 저장소 측 wrapper 작성 sprint 의 백로그.
- 범위: scripts/ + docs/llm-wiki/ + memory/. **코드 + docs 신규**. **신규 ID 0 row**.
- 상태: in_progress (PR #552 OPEN + 6 commit push + T-d-72-4 + T-d-79-1 + T-d-80-1 done + housekeeping + handoff message 작성, PR 머지 pending)
- 최종 수정일: 2026-06-11

## 1. 태스크 (sprint)

- [x] WB-01: my_harness 측 SSOT 작성 (wiki_ingest_skill_spec.md + SKILL.md + run_wiki_ingest.py)
- [x] WB-02: 본 저장소 측 wrapper 작성 (scripts/wiki-ingest-from-raw.sh)
- [x] WB-03: 본 저장소 측 사용법 가이드 (docs/llm-wiki/ingest-skill.md)
- [x] WB-04: dry-run 검증 (83 source 식별, 0 errors)
- [x] WB-05: main flat memory 3 file 갱신 (state.json M-v1.0 notes + work_backlog.md status + date)
- [x] WB-06: branch memory directory 작성 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-11.md)
- [x] WB-07: PR #552 발행 (push + gh pr create)
- [x] WB-08a: 3 commit 모순점 정정 (fix(wiki) publication-matrix + README / chore(wiki) 3 columns accepted / fix(docs) 5 broken link) → PR #552 update
- [x] WB-08b: T-d-72-4 --apply (82 wiki page 신규 ingest) — vault git commit b1599cc
- [x] WB-09a: T-d-79-1 done — wiki-query thin wrapper (scripts/wiki-query.sh 192 lines + docs/llm-wiki/query-skill.md 230+ lines, commit 15ca106f)
- [x] WB-09b: T-d-80-1 done — wiki-pr-update thin wrapper (scripts/wiki-pr-update.sh 150 lines + docs/llm-wiki/pr-update-skill.md 230+ lines, commit 15ca106f)
- [x] WB-10: housekeeping — docs/llm-wiki/README.md 5 → 8 file root index 갱신 + handoff-to-my-harness.md 작성 (T-d-79-2/T-d-80-2 의뢰서)
- [x] WB-11: 5 background agent 결과 회수 + 활용 (vault 운영 정책 + D-72 verbatim template + Obsidian query best practice + PR hook pattern)
- [ ] WB-12: PR #552 머지 (사용자 confirm 후)
- [ ] WB-13: PR #551 (T-d-72-2 re-sync) 머지 (사용자 confirm 후)

## 2. 잔여 (carry-over, 별도 sprint)

- T-d-79-2 (my_harness 측 wiki_query_skill_spec.md + SKILL.md + scripts/run_wiki_query.py 작성) — `handoff-to-my-harness.md` 참고
- T-d-80-2 (my_harness 측 wiki_pr_update_skill_spec.md + SKILL.md + scripts/run_wiki_pr_update.py 작성) — `handoff-to-my-harness.md` 참고
- T-d-79-3, T-d-79-4 (본 저장소 dry-run + --file 검증) — T-d-79-2 완료 후
- T-d-80-3, T-d-80-4, T-d-80-5, T-d-80-6 (본 저장소 dry-run + apply + idempotency + reingest 검증) — T-d-80-2 완료 후
- T-d-79-5, T-d-79-6, T-d-80-7, T-d-80-8, T-d-80-9 (wiki-lint 통합 / v2.0 RAG / CI integration / mergeCommitSha 자동 fill)
- wiki-lint 8 errors / 62 warns follow-up (Phase 3 ingest 결과, 별도 sprint)
- T-d-72-5 (wiki/cross/ cross-project 종합)
- T-d-72-6 (wiki-lint CI integration, D-74+)
- D-73 (my_harness 측 wiki-lint --project 옵션)
- D-74 (`_lint/devhub/` per-project lint report 디렉터리)
- workflow_skill_catalog.md 갱신 (my_harness 측, D-79/D-80 row 추가)
- PR #548 (N-13 backend foundation) 머지 결정
- PR A-2 (routing/auto_route.go + voc_handler 통합 + openapi.yaml)
- N-6 staging 1주 운영 (사용자 결정)

## 3. 관련 PR

- **PR #544** (D-72 Phase 1, 2026-06-10 머지): `docs/llm-wiki/` 5 file + `scripts/wiki-sync-devhub.sh` — 본 sprint 의 raw mirror tool
- **PR #545** (2026-06-10 머지): ai-workflow v0.5.11 sync
- **PR #546** (2026-06-11 머지): N-10 housekeeping
- **PR #547** (2026-06-11 머지): N-13 housekeeping
- **PR #549** (2026-06-11 머지): T-d-72-2 metadata 정합
- **PR #550** (2026-06-11 머지): E2E spec timing 안정화
- **PR #551** (OPEN, chore T-d-72-2 re-sync 5 file 갭 해소)
- **PR #552** (OPEN, 본 sprint — wiki-ingest-from-raw skill wrapper + 3 commit 모순점 정정 + T-d-79-1 + T-d-80-1 thin wrapper + housekeeping + handoff message)

## 4. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | sprint 시작 + 4 file 변경 + 4 file memory + dry-run PASS + PR #552 발행 pending |
| 2026-06-11 | 3 commit 모순점 정정 (commit 9124e4f0 fix(wiki) publication-matrix + README / ea930f24 chore(wiki) 3 columns accepted / f03f491f fix(docs) 5 broken link) → PR #552 update + T-d-72-4 --apply (82 wiki page ingest, vault commit b1599cc) + memory 4 file 갱신 |
| 2026-06-11 | T-d-79-1 + T-d-80-1 done — wiki-query + wiki-pr-update thin wrapper 4 file (commit 15ca106f) + 5 background agent 결과 회수 + housekeeping (docs/llm-wiki/README.md 5 → 8 file + handoff-to-my-harness.md 작성) + memory 4 file 갱신 |
