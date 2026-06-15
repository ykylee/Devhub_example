# Changelog

All notable changes to this project will be documented in this file.

본 파일은 `tools/release_pipeline.py changelog-gen` 으로 자동 생성됩니다 (v0.7.14+).
수동 편집도 가능하나 다음 release 시 자동 갱신 시 충돌 가능.

## [Unreleased] - 2026-06-14

### Added

- feat(wiki): score trend dim 별 alert (--alert --baseline + 4 smoke test) (0224a76)
- feat(wiki): score trend over time (commit 별 추적) + 10 smoke test (99e299f)
- feat(wiki): emit_wiki_l2_body.py --mode=metadata-only + 498 page 본문 emit (c72bdc3)
- feat(wiki): maintainability score metric (6 dim dashboard) + 12 smoke test (49dfc78)
- feat(wiki): emit_wiki_l2_body.py update_l2_full() 추가 + 전체 apply 검증 (7a4dbae)
- feat(v0.7.0 follow-up): packaging + session-start opt-in + evaluate_compliance (71de1b0)
- feat(v0.7.0 step 7): Extension 시스템 일반화 (5 file, +1150 line, 23 test PASS) (0052da1)
- feat(v0.7.0 step 6): Reverse Engineering 9-Artifact (11 file, +925 line, 19 test PASS) (4bbd391)
- feat(v0.7.0 step 8): Security Baseline extension (6 rule + opt-in + 15 test) (dc2c22b)
- feat(v0.7.0 step 9): Unit of Work 3-layer template + smoke test (17 test PASS) (c981cac)
- feat(v0.7.0 step 10): Audit Log 표준화 (1 spec + 1 helper fix + 13 test) (54e96a9)
- feat(v0.7.0 step 1): stage_completion required 격상 + ensure fallback (6e57cf3)
- feat(v0.6.1.5): P2.5 R9 wiki-ingest source rule (archive/ only) (0634dd6)
- feat(v0.6.0.1): P1.5 memory/active/ rename + bootstrap --enable-wiki + 6 harness wiki/ stub (7134c5b)
- feat(harness): add Claude (Claude Code) harness target (578bed6)
- feat(phase9): implement strict pydantic schemas, mcp v1.0 migration, and multi-agent topology (2bb8bc7)
- feat: stabilize phase 8 and prepare for phase 9 evolution (dce9023)
- feat: optimize auto-mode detection with prioritized keywords and reordering (5d20899)
- feat: complete Phase 6/7 residual tasks (robust-patcher MCP, smart-reader unification, auto-mode detection) (e12a3cf)
- feat: complete Phase 8 infrastructure and maturity assessment (a654042)
- feat: implement mode-aware workflow optimization (383a50c)
- feat: improve multi-stack analyzer and update maturity matrix to Phase 6 done (ec0b607)
- feat: bump version to v0.4.1-beta and implement legacy gitignore migration (88e5b12)
- feat: enhance packaging and deployment with version-aware upgrade and gitignore scripts (7a3b932)
- feat(workflow): complete isolation of workflow infrastructure into ai-workflow/ folder (b86256f)
- feat: Implement branch-isolated workspace, doc governance, and AI doc compiler (4470934)
- feat(workflow): implement Phase 6 tools and Event Sourcing backlog for conflict-free collaboration (1b37b86)
- feat: complete transition to antigravity harness and enable package export (e8a6198)
- feat: enhance antigravity harness with artifacts and browser sub-agent guidelines (221e02e)
- feat: migrate workflow state paths to relative workspace-root basis (441d39d)
- ... (10 more)

### Changed

- chore: ignore local tooling (.mavis/, .obsidian/) (95b1d2c)
- docs(wiki): v0.7.0 5 concept page + L2 emit helper + drift smoke test (021ec16)
- wiki: v0.7.0 step 9 — Unit of Work 3-layer template (17 test PASS) (b7641e3)
- wiki: v0.7.0 step 10 — Audit Log 표준화 (1 spec + 1 helper fix + 13 test) (2458cf8)
- wiki: v0.7.0 step 1 — stage_completion required 격상 (8 test PASS) (6148c0f)
- wiki: AIDLC benchmark topic §9 v0.6.5 → v0.6.6 정정 (17 commit, ~3,800 line) (58f3ad2)
- wiki: v0.6.6 follow-up #1 — 5 SKILL.md-only skill runtime (12/12 일관성) (8ae9102)
- wiki: v0.6.5 release — AIDLC 패턴 차용 (10 commit, ~2,600 line) (46e4d1f)
- wiki: v0.6.5 batch runtime — 6 spec stage_completion (10/11, +72 line) (0ae8d4a)
- wiki: v0.6.5 pilot runtime — automated-repro-scaffold stage_completion (1/11, +44 line) (fbe9673)
- wiki: v0.6.5 runtime — stage-gate-pattern §12 + log entry (35 test PASS) (fbc8370)
- wiki: v0.6.5 — stage-gate-pattern §8 + log entry (StageCompletion 11종 적용 추적) (0001782)
- wiki: v0.6.4 신규 concept 2종 (Question File Format + Stage Gate Pattern) (d32226b)
- wiki: promote topics/aidlc-benchmark-analysis-2026-06-12 (draft → active) (9946d91)
- wiki-ingest: AIDLC benchmark analysis (1 topic page) (2916d49)
- lint-fix: phase-7-verification — L1 log entry append (a8fc1ad)
- lint-fix: phase-7-verification — 11 broken wikilink 일괄 fix (a5762e5)
- wiki-ingest: phase-5-topics — 3 cross-cutting topic (R2 batch 4) (10bbbd0)
- wiki-ingest: phase-4-patterns-entities — 3 pattern + 12 entity (R2 batch 3, 15 page 한도 매칭) (d95681e)
- wiki-ingest: phase-3-decisions — 4 ADR 신규 + 1 ADR extend (R2 batch 2) (1f60554)
- wiki-ingest: phase-2-concepts — 5 concept pages 신규 (R2 batch 1) (5d7db10)
- wiki-ingest: phase-1-scaffold (INGEST_GUIDE + index anchors + log) (cae89f1)
- self-dogfood: wiki organization — 5 page types with 6 entries + V-R9 fix (73fea00)
- self-dogfood: apply wiki + memory workflow to own project (792fddc)
- docs(v0.5.10-post): full documentation refresh + dev install guide (e927060)
- packaging: include runtime skill/mcp executables and upgrade cleanup review (db63581)
- packaging: ship runtime-only artifacts for v0.5.1-beta (eca63d8)
- docs: add column-style wiki draft structure (9394eda)
- docs: split human vs ai docs guidance across workflow standards (16ffd12)
- docs: align memory path policy to memory/<branch> (f623726)
- ... (97 more)

### Fixed

- fix(v0.5.10.1): smart update — VERSION marker + content hash 기반 silent 갱신 (332de1a)
- fix(v0.5.9.1): wire 가이드 §3 sub_payloads fix + 회귀 test (c202634)
- fix(v0.5.7.1): include workflow_kit.common.{state,contracts,schemas} in wheel packaging (8e5cd7b)
- fix: 42/42 smoke tests green + MiniMax Code harness overlay (#13) (fd12017)
- fix(mcp): remove unsupported version param from FastMCP init (fa0eee4)
- fix(ci): update test paths and dependencies in smoke.yml (d6df3d0)
- fix: address PR #10 review feedback (6d89752)
- fix(workflow): stabilize beta 0.4.0 state configuration (49e74a7)
- fix: resolve PR #8 feedback, improve smart-reader formatting, and stabilize state management (1f80787)
- fix: update relative links in project documents to reflect new root structure (a76f34b)
- fix: include runtime scripts in default bootstrap paths (d721cd9)
- fix: resolve empty project detection bug and update docs for team sharing (1c9860e)
- fix: comprehensive smoke_check inference across all harnesses (Gemini, Codex, Antigravity, OpenCode) (7fca126)
- fix: ensure smoke_check inference in all opencode agent templates (100b4de)
- Fix CI-broken README dist links (c7fc67d)

## [3.0.1] - 2026-04-27

### Added

- feat: add Pi Coding Agent (pi.dev) harness support (v3.0.1) (9c4fb1d)

## [3.0] - 2026-04-27

### Added

- feat: Phase 5 official release (v3.0) with unified operations path and updated schemas (3a7e4c1)

## [0.7.9] - 2026-06-14

### Added

- feat(v0.7.9): release_pipeline tool 정식화 (validate / version-bump / note-draft) + 8 smoke test (cb0a892)

### Changed

- docs(v0.7.9): release note backfill + refresh_wiki_memory v0.7.9 tracking (283823e)
- chore(v0.7.9): version bump 0.7.8 → 0.7.9 + release note (d39be44)

## [0.7.8] - 2026-06-14

### Added

- feat(v0.7.8): refresh_wiki_memory 에 v0.7.8 release tracking 추가 (662bead)
- feat(v0.7.8): state-aware evaluate_compliance + config actual apply (d3235ad)

### Changed

- docs(v0.7.8): release note commit hash backfill (b67af83) (f444e68)
- chore(v0.7.8): version bump 0.7.7 → 0.7.8 + release note (b67af83)

## [0.7.7] - 2026-06-14

### Added

- feat(v0.7.7): refresh_wiki_memory 에 v0.7.7 release tracking 추가 (fd18288)
- feat(v0.7.7): workflow_kit.cli.doctor 에 load_config + should_fail integration (022672f)

### Changed

- docs(v0.7.7): release note commit hash backfill (3300e73) (7581dd2)
- chore(v0.7.7): version bump 0.7.6 → 0.7.7 + release note (3300e73)

## [0.7.6] - 2026-06-14

### Added

- feat(v0.7.6): refresh_wiki_memory 에 v0.7.6 release tracking 추가 (1fefdfd)
- feat(v0.7.6): workflow_kit.metadata (pyproject.toml [tool.workflow-doctor] loader) + 10 smoke test (0daf6da)
- feat(v0.7.6): run_all_checks 통합 runner + 10 smoke test (53d5dc8)

### Changed

- docs(v0.7.6): release note commit hash backfill (b9ede19) (7a5c56e)
- chore(v0.7.6): version bump 0.7.5 → 0.7.6 + release note (b9ede19)

## [0.7.5] - 2026-06-14

### Added

- feat(v0.7.5): refresh_wiki_memory 에 v0.7.5 release tracking 추가 (150ee32)
- feat(v0.7.5): refresh_wiki_memory tool 정식화 + 10 smoke test (Wiki 운영 자동화) (0741775)

### Changed

- docs(v0.7.5): release note commit hash backfill (c2a75f8) (51edde5)
- chore(v0.7.5): version bump 0.7.4 → 0.7.5 + release note (c2a75f8)
- test(v0.7.5): 4 sub-cat dispatcher runtime test 보강 (12 → 16) (9e1f206)

## [0.7.4] - 2026-06-13

### Added

- feat(v0.7.4): CLI wrapper (workflow doctor) + @graceful_shutdown + optional dep (hypothesis/objgraph) (22e7750)

### Changed

- chore(v0.7.4): score history v0.7.4 entry (Overall 4.67 A 유지) (cfb09fb)
- docs(v0.7.4): wiki log v0.7.4-beta entry 추가 (1818dd6)

## [0.7.3] - 2026-06-13

### Added

- feat(v0.7.3): 4 runtime helper (auth/testing/profiling/resiliency) + 7 baseline dispatcher (d03348a)

### Changed

- chore(v0.7.3): score history v0.7.2/v0.7.3 entry 추가 (Overall 4.66→4.67 A 유지) (c732c0f)
- docs(v0.7.3): wiki log v0.7.3-beta entry 추가 (be49e0f)

## [0.7.2] - 2026-06-13

### Added

- feat(v0.7.2): Extension sub-cat + 4종 (resiliency) 본 구현 (179 test PASS) (3bffba3)

### Changed

- docs(v0.7.2): wiki log commit hash TBD → 3bffba3 갱신 (7cae496)

## [0.7.13] - 2026-06-15

### Added

- feat(v0.7.13): cmd_release --version flag (staging backfill, pyproject 일시 patch 불필요) (922ebc0)

### Changed

- chore(v0.7.13): state sync (v0.7.12 + v0.7.13 backfill) + 2 daily backlog (afc685a)
- chore(v0.7.13): version bump 0.7.11 → 0.7.13 + __version__ sync + Beta-v0.7.13.md (628bf93)

### Fixed

- fix(v0.7.13): Beta-v0.7.13.md Commit section + 3 commit hash (727c59c)

## [0.7.12] - 2026-06-15

### Added

- feat(v0.7.12): refresh_wiki_memory REPO_ROOT auto-detect (CLI flag > env var > git rev-parse > legacy fallback) + 4 smoke (63080ba)

### Changed

- chore(v0.7.12): v0.7.5~v0.7.10 release backfill (6 wheel/sdist + 6 git tag + 6 GH release) + Beta-v0.7.12.md 갱신 (89b7af5)

### Fixed

- fix(v0.7.12): Beta-v0.7.12.md commit table TBD → 89b7af5 (5b8e730)
- fix(v0.7.12): Beta-v0.7.12.md commit table TBD → 63080ba (0b3e704)

## [0.7.11] - 2026-06-15

### Added

- feat(v0.7.11): release_pipeline Phase 3 (dist subcommand) + state sync + 8 smoke (b2650f5)

### Changed

- chore(v0.7.11): version bump 0.7.10 → 0.7.11 + __version__ sync (ec407f1)

### Fixed

- fix(v0.7.11): cmd_verify --json field names (tag → tagName, createdAt → publishedAt) + release note commit table (aa4e837)

## [0.7.10] - 2026-06-14

### Added

- feat(v0.7.10): release_pipeline Phase 2 (release / verify / rollback) + 8 smoke test (fdf8159)

### Changed

- docs(v0.7.10): release note backfill + refresh_wiki_memory v0.7.10 tracking (fc87fdd)
- chore(v0.7.10): version bump 0.7.9 → 0.7.10 + release note (67d4a37)

## [0.7.1] - 2026-06-13

### Added

- feat(v0.7.1): follow-up 4건 + wiki 개선 4건 묶음 (158 test PASS, GH release) (f09034d)

### Changed

- docs(v0.7.1): wiki log commit hash TBD → 0224a76 갱신 (9935e06)
- docs(v0.7.1): wiki log commit hash TBD → 99e299f 갱신 (d8c981c)
- docs(v0.7.1): wiki log commit hash TBD → f09034d 갱신 (bad14d8)

## [0.7.0] - 2026-06-13

### Changed

- docs(v0.7.0): wiki log commit hash TBD → c72bdc3 갱신 (bdc6ceb)
- docs(v0.7.0): wiki log commit hash TBD → 49dfc78 갱신 (471fee2)
- docs(v0.7.0): wiki log commit hash 갱신 TBD-pending → 7a4dbae (b375951)
- docs(v0.7.0): wiki log commit hash TBD → 021ec16 갱신 (ac75d72)
- docs(v0.7.0): wiki log entry header 에 commit hash 7자 prefix 명시 (3fcd480)
- docs(v0.7.0): release note follow-up section 추가 (Task 3+2+1) (8818cbe)
- chore(v0.7.0): version bump 0.6.3 → 0.7.0 (390a6e0)
- docs(v0.7.0): Release notes + wiki log entry (15 commit, ~3,200 line, 130 test PASS) (dff0aae)

## [0.6.6] - 2026-06-12

### Added

- feat(v0.6.6): 5 SKILL.md-only skill runtime 통합 (12/12 spec+runtime 일관성) (6a9126c)

## [0.6.5] - 2026-06-12

### Added

- feat(v0.6.5): batch stage_completion integration — 6 spec 보유 skill (10/11 완료) (ca7a685)
- feat(v0.6.5): pilot stage_completion integration — automated-repro-scaffold (2fab835)
- feat(v0.6.5): Stage Gate Runtime helper + migration guide (3 file, 13 test PASS) (dd98e69)
- feat(v0.6.5): StageCompletion field 11종 skill spec + catalog 보강 (13 file) (5b16517)

### Changed

- release(v0.6.5): AIDLC 패턴 차용 (Question File Format + Stage Gate) (3897da7)

## [0.6.4] - 2026-06-12

### Added

- feat(v0.6.4): Question Format + Stage Gate 코드 (2 module + 2 smoke test) (bc16d91)
- feat(v0.6.4): Question File Format + Stage Gate 명시화 (4 doc) (25756bb)

### Fixed

- fix(v0.6.4): V-R9 skip marker — naive grep false-positive 17 → 0 (30183c5)

## [0.6.3] - 2026-06-12

### Added

- feat(v0.6.3): P4 memory/log.md + harness overlay consistency check (3261e20)

### Changed

- release(v0.6.3): final v0.6.x series release — all 4 milestones complete (1923705)

### Fixed

- fix(v0.6.3): P6 phase-6 backfill — INGEST_GUIDE path 정정 + log 보강 (1d7ca77)
- fix(v0.6.3): broken relative links after memory/active/ rename + fix bootstrap test leniency (6b2bf00)

## [0.6.2] - 2026-06-12

### Added

- feat(v0.6.2): P3 T2 work_backlog anchor + T3 ingest atomicity (2713059)

## [0.6.1] - 2026-06-12

### Added

- feat(v0.6.1): P2 R8 freeze + R10 freeze lint + T1 memory lint 4종 + R7 merge-res ext (0620373)

## [0.5.9] - 2026-06-08

### Changed

- docs(v0.5.9): wire 가이드 §7/§8/§9 보강 — sub.delegation_id parent prefix 룰 명시 (1006ff0)

## [0.5.8] - 2026-06-08

### Added

- feat(v0.5.8): interactive --harness picker + packaging smoke automation (6213dcc)

## [0.5.7] - 2026-06-08

### Added

- feat(v0.5.7): contract v1 §4.2/§5.2 multi-component fan-out/in + §6.3 cross-ref row (#21) (ebf7e7c)

## [0.5.6] - 2026-06-07

### Added

- feat(v0.5.6): contract v1 §5/§6 P0 enforcement (validator + delegator) (#20) (79f3bec)

### Changed

- docs(v0.5.6): mark TASK-V056-001 done post-merge (731787b)

## [0.5.5] - 2026-06-07

### Added

- feat(v0.5.5): Phase 11 본격 pilot (Devhub Example × Contract v1) (#19) (1f095ec)

### Changed

- docs(v0.5.5): mark TASK-V055-001 done post-merge (75a3fc6)

## [0.5.4] - 2026-06-07

### Added

- feat(v0.5.4): orchestrator ↔ sub-agent delegation contract v1 (closes #1) (#18) (7737e14)

## [0.5.3] - 2026-06-07

### Added

- feat(v0.5.3): antigravity MCP config 표준화 + cross-language stack 표시 (#17) (b6ae73a)

### Changed

- docs(v0.5.3): mark tasks done + release notes for Beta v0.5.3 (a961fa0)

## [0.5.2] - 2026-06-06

### Changed

- refactor(v0.5.2): bootstrap_workflow_kit.py → bootstrap_lib/ 6-module package (#16) (9497a35)

## [0.5.11] - 2026-06-09

### Added

- feat(v0.5.11): Mavis engine hook (§6.5) + 회귀 test 강화 (4ce3635)

### Changed

- docs(v0.5.11): Beta-v0.5.11.md release note — 릴리스 일자 표현 정정 (5bd4c4f)
- chore(v0.5.11): bump version to 0.5.11-beta (cfde435)
- docs(v0.5.11): Beta-v0.5.11 release note (cb0f698)
- docs(v0.5.11): governance 갱신 — 1인 dev 환경, 24h cool-down 명시 (c89971d)
- docs(v0.5.11): --no-interactive 비대화형 가이드 보강 (7329dbe)

### Fixed

- fix(v0.5.11): bootstrap MCP env — wheel install 시 PYTHONPATH omit (e388366)
- fix(v0.5.11): MCP initialize response — protocolVersion 추가 (677bcff)

## [0.5.10] - 2026-06-08

### Fixed

- fix(v0.5.10): choose_roles sub.delegation_id parent-prefix spec 정합 (8359cfc)

## [0.5.1] - 2026-06-06

### Added

- feat(v0.5.1): end-to-end MCP round-trip smoke (#15) (73f8f2f)
- feat(v0.5.1): per-harness MCP install + auto-emit + guide (#14) (c3c9a90)
