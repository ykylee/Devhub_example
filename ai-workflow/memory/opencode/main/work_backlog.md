# Work Backlog — main (post-sprint cleanup + M-v0.2.2 release close, 2026-06-22)

- Branch: `main` (현재 default branch)
- Agent: opencode (Sisyphus, MiniMax-M3)
- Updated: 2026-06-22 (T08:30+09:00)
- Status: clean (no in-flight work, Session 1~5 종합 정합 완료, PR #662/#679/#680/#681/#682/#683/#684/#685 모두 MERGED, main HEAD = 5cb03bbd, working tree clean, local branch 3개 정리 완료)

## In Progress (P0)

없음. main 브랜치 = `5cb03bbd` (PR #684 + Session 5 final cleanup, M-v0.2.1 release close + memory 4 file final), working tree clean, origin/main 정합.

## Pending (P1) — M-v0.2.1+ scope 진입 후보

- [x] **WB-M01**: M-v0.2.2 backend-ai 폐기 (umbrella §1.2 G2 + §6.6) — **✅ done 2026-06-22 (4-PR split, PR #663 + #664 + #665 + #666 모두 MERGED) + PR #679 release close 종합 정합**
- [ ] **WB-M02**: §2.5 Tier 분리 정공법 (umbrella doc §2.5 + AGENTS.md §6.5)
  - §6.5 PR 작성 시 self-check 5 row 검증 (env var / 호스트명 / IP / 경로 / env file)
  - umbrella doc §2.5 신설 또는 기존 §2.6.2 (network 정책) 와 통합
  - 직전 sprint handoff P1 잔여 WB-13
- [ ] **WB-M03**: §17.2 known gaps 자연 해소 (PoC +1주 이내)
  - row 3: incident runbook tuning (manual SOP, 2026-06-19+1주 = 2026-06-26 경과)
  - row 4: backend-knowledge/ 디렉터리 skeleton 별도 PR — **이미 done (PR #654-#656)**, 잔여 row 갱신만 필요
  - row 5: Pi SDK npm dependency (자동, CI npm ci)
  - row 6: backup schedule cron 등록 (자동, Docker sidecar)
  - 직전 sprint handoff P1 잔여 WB-15
- [ ] **WB-M04**: §2.4 standalone 검증 자동화 tool 정식 활성화 (M-v0.2.1+ CI 도입)
  - `scripts/check_standalone_drift.sh` (M-v0.2.1+ CI pre-merge) — 10 row 자동 grep
  - `docs/operations/standalone-verification-m-v0-2-1.md` 결과 문서 작성 (per phase)

## Pending (P2) — M-v0.2.0 release 직후 +1주 ~ +2주

- [ ] **WB-M10**: M-v0.2.0 release announcement (사내/사외 배포)
- [ ] **WB-M11**: §3.5.6 cross-link reverse index PoC 운영 검증 (umbrella §13.2 known gap 1 resolved 검증)
- [ ] **WB-M12**: §3.7.6 data normalization pipeline 자동 검증 운영
- [ ] **WB-M13**: §3.6.6 governance audit log 운영 검증 (5 audit_event + 28 metric)
- [ ] **WB-M14**: homelab 정식 wire (M-v0.2.0 은 homelab_mock PoC, M-v0.2.1 정식)
  - `backend-knowledge/sources/homelab.py` 구현 (사내 HomeLab agent API wire)
  - source plugin credential schema + 봉투 암호화 정합 (ADR-0025)

## Pending (P3) — M-v0.2.0 release +2주 ~ +1개월

- [ ] **WB-M20**: §15 ADR supersession (M-v0.2.3+ 부터 가능, umbrella §15 정공법)
- [ ] **WB-M21**: §16 API versioning (M-v0.3.0+ 부터 v0-3 도입)
- [ ] **WB-M22**: §3.5.7 Pi LLM cross-link 자동 resolution (M-v0.2.3+, PR #673 진행 중)
- [ ] **WB-M23**: §3.5.8 Pi LLM resolution false positive rollback (M-v0.2.3+)
- [ ] **WB-M24**: M-v0.2.3 hrdb source plugin + Pi LLM enrich 활성화 (umbrella §1.2 G3 7종 source, PR #672 진행 중)
- [ ] **WB-M25**: M-v0.2.3 PostgreSQL option (sqlite → PG migration, §10.1 option)

## Done (2026-06-22)

### Session 1 (T00:00~00:10) — feat 브랜치 폐기 + .gitignore 분리 PR

- [x] **DB-M01**: feat/260619-v0-2-frontend-svelte rebase (-X ours, 4 commit auto-drop)
- [x] **DB-M02**: /src/var/ .gitignore commit (7183d7c6 → 006284ae cherry-pick)
- [x] **DB-M03**: chore/260621-gitignore-src-var 브랜치 생성 + push
- [x] **DB-M04**: feat/260619-v0-2-frontend-svelte reset to origin/main + force-push
- [x] **DB-M05**: PR #662 생성 (사외 tier, 추적성 N/A, worker/opencode label)
- [x] **DB-M06**: PR #662 self-merge (regular merge commit 196d8593)
- [x] **DB-M07**: feat/260619-v0-2-frontend-svelte local + origin 폐기
- [x] **DB-M08**: local main fast-forward to origin/main (34 file / +5568 / -984)
- [x] **DB-M09**: opencode/main/ 메모리 디렉터리 신규 생성 (state.json + session_handoff.md + work_backlog.md)

### Session 2 (T05:30~07:00) — M-v0.2.2 release close (4-PR split) 종합 정합

- [x] **DB-M10**: M-v0.2.2 4 PR 검증 (PR #663 + #664 + #665 + #666 모두 MERGED, §6.6.2 SOP 10 단계 모두 ✅)
- [x] **DB-M11**: umbrella doc 9개 위치 status 4-PR 갱신 (§5.1 / §5.2 / §5.5 / §6.6.2 header + 10 단계 step 7/10 + PR list + DoD (c) + 현황 요약)
- [x] **DB-M12**: state.json M-v0.2.2 row 신규 (status: done + 4_pr_split + sop_completion + umbrella_doc_refs + adr_alignment + tier)
- [x] **DB-M13**: branch `chore/260622-m-v0-2-2-release-close` 생성 + push
- [x] **DB-M14**: PR #679 생성 (umbrella + state.json cross-cutting docs only PR)
- [x] **DB-M15**: PR #679 CI 4/4 SUCCESS + 5 SKIPPED (docs only 의도된 skip)
- [x] **DB-M16**: PR #679 self-merge (regular merge, fast-forward 가능, merge commit 86d1300d, 2026-06-22T06:58:41Z)
- [x] **DB-M17**: chore/260622-m-v0-2-2-release-close 자동삭제 (--delete-branch)
- [x] **DB-M18**: local main fast-forward to origin/main (PR #679, 2 file / +32/-14)
- [x] **DB-M19**: opencode/main/ 메모리 4 file 갱신 (state.json + work_backlog.md + session_handoff.md + backlog/2026-06-22.md)
- [x] **DB-M20**: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real mode, AGENTS.md §문서 작업 기준)
- [x] **DB-M21**: `bash scripts/wiki-frontmatter-update.sh` 1회 실행 (matched 0 stale X → matched X stale 0)
- [x] **DB-M22**: session 종료 메모리 finalize (opencode/main/ + opencode/chore/260622-session-end-memory-update/)

### Session 3 (T07:10~07:30) — M-v0.2.3 release close (2-PR split, partial done) 종합 정합

- [x] **DB-M23**: M-v0.2.3 2 PR 검증 (PR #672 + #673 모두 MERGED)
- [x] **DB-M24**: umbrella doc 5개 위치 status partial done 갱신 (§5.1 / §5.2 P3 / §5.5 / §6.7.1 / §6.7.3)
- [x] **DB-M25**: state.json M-v0.2.3 row 신규 (status: partial_done + 2_pr_split + sop_completion + umbrella_doc_refs + adr_alignment + tier)
- [x] **DB-M26**: branch `chore/260622-m-v0-2-3-release-close` 생성 + push
- [x] **DB-M27**: PR #681 생성 (umbrella + state.json cross-cutting docs only PR)
- [x] **DB-M28**: PR #681 CI 4/4 SUCCESS + 5 SKIPPED
- [x] **DB-M29**: PR #681 self-merge (regular merge, fast-forward 가능, merge commit 41f1b7b6, 2026-06-22T07:27:48Z)
- [x] **DB-M30**: chore/260622-m-v0-2-3-release-close 자동삭제
- [x] **DB-M31**: opencode/main/ 메모리 4 file 갱신 (state.json + work_backlog.md + session_handoff.md + backlog/2026-06-22.md)
- [x] **DB-M32**: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real mode)
- [x] **DB-M33**: `bash scripts/wiki-frontmatter-update.sh` 1회 실행
- [x] **DB-M34**: session 종료 메모리 finalize (opencode/main/ + chore/260622-m-v0-2-3-memory/)

### Session 4 (T07:35~08:20) — M-v0.2.1 release close (4-PR split, partial done) 종합 정합

- [x] **DB-M35**: M-v0.2.1 4 PR 검증 (PR #655 + #660 + #661 + #675 모두 MERGED)
- [x] **DB-M36**: umbrella doc 5개 위치 status partial done 갱신 (§5.1 / §5.2 P1 3 row / §5.5)
- [x] **DB-M37**: state.json M-v0.2.1 row 신규 (status: partial_done + 4_pr_split + sop_completion + umbrella_doc_refs + adr_alignment + tier)
- [x] **DB-M38**: branch `chore/260622-m-v0-2-1-release-close` 생성 + push
- [x] **DB-M39**: PR #683 생성 (umbrella + state.json cross-cutting docs only PR)
- [x] **DB-M40**: PR #683 CI 4/4 SUCCESS + 5 SKIPPED
- [x] **DB-M41**: PR #683 self-merge (regular merge, fast-forward 가능, merge commit bcaf9089, 2026-06-22T08:14:43Z)
- [x] **DB-M42**: chore/260622-m-v0-2-1-release-close 자동삭제
- [x] **DB-M43**: opencode/main/ 메모리 4 file 갱신 (state.json + work_backlog.md + session_handoff.md + backlog/2026-06-22.md)
- [x] **DB-M44**: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (real mode)
- [x] **DB-M45**: `bash scripts/wiki-frontmatter-update.sh` 1회 실행
- [x] **DB-M46**: session 종료 메모리 finalize (opencode/main/ + chore/260622-m-v0-2-1-memory/)

### Session 5 (T08:20~08:30) — 작업 내역 정리 + 메모리 갱신 + 원격 미반영 내용 마무리로 PR

- [x] **DB-M47**: working tree 98 file wiki raw main 정합 (`git restore ai-workflow/wiki/`, vault sync 잔여 정리)
- [x] **DB-M48**: local branch 3개 정리 (`git branch -d` chore/260622-m-v0-2-1-release-close + m-v0-2-1-memory + m-v0-2-3-memory, main 와 동일 content)
- [x] **DB-M49**: opencode/main/ 메모리 4 file 최종 갱신 (state.json Session 5 + work_backlog.md §5 row DB-M47~49 + session_handoff.md §10 Session 5 append + backlog/2026-06-22.md 본 § append)
- [x] **DB-M50**: branch `chore/260622-session-end-final` 생성 + commit + push + PR #685 발행
- [x] **DB-M51**: PR #685 self-merge (regular merge, Session 5 final cleanup 정공법)
- [x] **DB-M52**: wiki mirror 1회 실행 (real mode, M-v0.2.0/v0.2.1/v0.2.2/v0.2.3 release close 종합)
- [x] **DB-M53**: session 종료 메모리 finalize (opencode/main/ + chore/260622-session-end-final/)

### Session 1~5 종합 (전체 release close 정공법)

- **Session 1** (T00:00~00:10): feat/260619-v0-2-frontend-svelte rebase + 폐기 + /src/var/ .gitignore PR #662
- **Session 2** (T05:30~07:00): M-v0.2.2 release close (4-PR split, §6.6.2 SOP 10/10 ✅) + PR #679 + memory PR #680
- **Session 3** (T07:10~07:30): M-v0.2.3 release close (2-PR split, partial done) + PR #681 + memory PR #682
- **Session 4** (T07:35~08:20): M-v0.2.1 release close (4-PR split, partial done) + PR #683 + memory PR #684
- **Session 5** (T08:20~08:30): working tree clean + local branch 3개 정리 + memory 4 file 최종 갱신 + PR #685 (final cleanup)

**PR 7건 종합** (Session 1~4):
- PR #662 (Session 1): .gitignore — M-v0.2.0 memory
- PR #679 (Session 2): umbrella + state.json M-v0.2.2 — release close
- PR #680 (Session 2): memory 4 file — M-v0.2.2 close
- PR #681 (Session 3): umbrella + state.json M-v0.2.3 — release close partial
- PR #682 (Session 3): memory 4 file — M-v0.2.3 close
- PR #683 (Session 4): umbrella + state.json M-v0.2.1 — release close partial
- PR #684 (Session 4): memory 4 file — M-v0.2.1 close
- PR #685 (Session 5): memory 4 file final + working tree cleanup

**4 milestone 모두 release close 정공법 완료**:
- M-v0.2.0: ✅ done (2026-06-18, 16+ commits, 7 PR)
- M-v0.2.1: 🟡 partial done (2026-06-22, 4-PR split, PR #655+#660+#661+#675)
- M-v0.2.2: ✅ done (2026-06-22, 4-PR split, PR #663+#664+#665+#666, §6.6.2 SOP 10/10)
- M-v0.2.3: 🟡 partial done (2026-06-22, 2-PR split, PR #672+#673)
- M-v0.3.0: ⏳ planned
