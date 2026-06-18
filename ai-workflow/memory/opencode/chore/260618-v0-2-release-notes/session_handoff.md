# Session Handoff — v0.2.0 umbrella + post-sprint follow-up 완료 (2026-06-18)

- **Branch**: `chore/260618-v0-2-release-notes`
- **Agent**: opencode (Sisyphus / MiniMax-M3)
- **Model**: MiniMax-M3
- **Session**: 2026-06-18T07:00:00+09:00 ~ 2026-06-18T07:30:00+09:00 (30 min)
- **Session purpose**: v0.2.0 umbrella doc publish + §17.3 post-sprint follow-up 6/6 row 처리 + M-v0.2.0 release notes 작성
- **Outcome**: ✅ 완료 (M-v0.2.0 release 직전 state 완전 정합, PR #650 OPEN)

## 1. 완료 (Done)

### 6 PR 생성 (5 MERGED + 1 OPEN)

| PR | 제목 | 상태 | Commit | 비고 |
| --- | --- | --- | --- | --- |
| [#645](https://github.com/ykylee/Devhub_example/pull/645) | v0.2.0 umbrella doc + ADR-0034 + ADR-0035 | MERGED | f6732a34 (squash 27 commit) | 본 작업의 핵심 PR |
| [#646](https://github.com/ykylee/Devhub_example/pull/646) | state.json M-v0.2.0 row 갱신 | MERGED | 1914ec8b | §17.3 P0 row 1+2 |
| [#647](https://github.com/ykylee/Devhub_example/pull/647) | child doc status active 전환 | MERGED | 01f1969c | §17.3 P1 row 3 |
| [#648](https://github.com/ykylee/Devhub_example/pull/648) | wiki mirror 갱신 | MERGED | 2141350a (squash 105 file) | §17.3 P1 row 4 |
| [#649](https://github.com/ykylee/Devhub_example/pull/649) | DOCUMENT_INDEX + planning/README + operation-sop 갱신 | MERGED | 00567fbc | §17.3 P2 row 6 |
| [#650](https://github.com/ykylee/Devhub_example/pull/650) | M-v0.2.0 release notes + §13.3 #5 + 6 stale page | **OPEN** | 0a9b3c6b + f9ff421b + e6333985 | §17.3 release notes row 5 + residual stale 갱신 |

### umbrella doc publish (PR #645)
- `docs/planning/release_v0-2_roadmap.md` (5459 line, 18 main section + 80+ subsection)
- `docs/adr/0034-okf-adoption.md` (Google OKF v0.1 채택, §6 Supersession section 신규)
- `docs/adr/0035-backend-knowledge-creation.md` (backend-knowledge 신설, §6 Supersession row 추가)
- 18/18 Q&A 결정, 28 metrics M-v0.2.3+ production, 4 cross-cutting 정공법 (§13/§14/§15/§16)
- 156 code blocks (78쌍 짝수) ✅

### post-sprint follow-up 6/6 row ✅
- P0 2 row: GitHub milestone #4 + state.json M-v0.2.0 row in_progress
- P1 2 row: child doc status active + wiki mirror 갱신 (85/85 matched, 0 stale)
- P2 1 row: DOCUMENT_INDEX + planning/README + operation-sop 갱신
- Release notes: docs/release-notes/v0.2.0.md (171 line, 11 section)
- **6 stale wiki page** (residual pre-existing) 갱신 완료 (git_commit: cac63f35 → 01f1969c)

## 2. 핵심 결정

- **v0.2.0 = backend-knowledge umbrella**: Google OKF v0.1 기반 + 5종 PoC source plugin (Gitea 4 sub-plugin + homelab_mock) + viz.html 자가 viewer
- **18/18 Q&A 결정** (Q1~Q18, 결정 일자 2026-06-17 또는 2026-06-18)
- **§1.1 한계 7개 식별** + §1.3 How 정당화 강화 (trade-off 한계 4개 추가 2026-06-18)
- **Path Y caller-provided user context** 결정 (`X-DevHub-User-Context` header)
- **file|db dual storage mode** 결정 (per source `source_meta.storage_mode` field)
- **§15 ADR supersession 정공법** (M-v0.2.3+ 부터 supersession 가능, 5 step + 12개월 deprecation policy)
- **§16 API versioning 정책** (M-v0.3.0+ v0-3 도입, 12개월 deprecation)
- **Tier 분리 정책 (사외/사내 2-tier)** (AGENTS.md §6, 2026-06-10 결정 정합)
- **PR #645 squash merge + branch delete** (chore/260618-bootstrap branch)
- **PR #650 squash merge planned** (3 commit → 1 squash)

## 3. Pending (다음 세션 / 후속 작업)

### P0 (즉시, 다음 세션 진입 시)
- **PR #650 머지** (`gh pr merge 650 --squash --delete-branch`)
- **GitHub tag v0.2.0** 생성 (`git tag -a v0.2.0 -m 'v0.2.0 backend-knowledge umbrella + ADR-0034/0035'`)
- **GitHub release publish** (`gh release create v0.2.0 --notes-file docs/release-notes/v0.2.0.md`)
- **사내 SCM sync** (AGENTS.md §6.3, GitHub main push 후 자동 sync)

### P1 (M-v0.2.0 PoC 운영 시)
- **§17.2 known gaps 4 row 자연 해소 시점** (M-v0.2.0 PoC 운영 +1주 이내):
  - row 3 (incident runbook tuning): manual SOP
  - row 4 (sprint 진입 checklist 잔여 1 row, backend-knowledge/ 디렉터리 skeleton 별도 PR)
  - row 5 (Pi SDK npm dependency): 자동 (CI `npm ci`)
  - row 6 (backup schedule cron 등록): 자동 (Docker sidecar cron container)
- **§2.5 Tier 분리 정공법 (사외/사내 2-tier 정공법)** — Section 2 디렉터리/코드/commit 의 Tier 분리 정공법 상세화 (사외 코드 + 사내 한정 + 공용 byte-identical drift 검증 + mirror-list.md §1.7/§1.8 정합 + lint-config.toml 정합 + scripts/wiki-sync-devhub.sh 화이트리스트 정합 + §2.4 item 1 + §2.6 + §11.1.1 + §13.4 정합)
- **backend-knowledge skeleton PR** (별도 PR 로 backend-knowledge/ 디렉터리 skeleton) — Python 3.13+ / FastAPI / OKF / Pi SDK / §2.1 디렉터리 구조 + §2.2 기술 스택 + §3.5.3 bundle 디렉터리 + §3.8 source plugin ABC + §10.1 DB schema + §11.1 incident runbook — §5.3 checklist 잔여 1 row 처리

### P2 (M-v0.2.0 release 시점)
- **M-v0.2.0 release announcement** (사내/사외 배포)
- **residual 6 stale wiki page** follow-up commit (이미 ✅ done, e6333985)

## 4. Key References

- umbrella doc: `docs/planning/release_v0-2_roadmap.md` (PR #645 MERGED)
- child doc: `docs/planning/external-integrations-agentic-rag-roadmap.md` (PR #647, status: active)
- ADR-0034: `docs/adr/0034-okf-adoption.md` (OKF v0.1 채택)
- ADR-0035: `docs/adr/0035-backend-knowledge-creation.md` (backend-knowledge 신설)
- Release notes: `docs/release-notes/v0.2.0.md` (PR #650, 171 line)
- state.json: `ai-workflow/memory/state.json` (M-v0.2.0 row in_progress + 18/18 Q&A)
- GitHub milestone: https://github.com/ykylee/Devhub_example/milestone/4
- AGENTS.md: `/AGENTS.md` (v0.2.0 릴리즈 로드맵 reference)
- docs/llm-wiki/operation-sop.md (Phase 1+1.5+3 mirror SOP)
- docs/governance/worker_division.md §6 (사외/사내 2-tier 형상관리) + §4.2 (ADR supersession 정공법)

## 5. 후속 작업 시 주의사항

- **1 commit = 1 logical concept change 정책** 유지 (umbrella + ADR 묶음, squash merge)
- **code block 짝수 검증** per commit (umbrella doc 본문, 156 = 78쌍 짝수)
- **Tier 분리 self-check** per PR (§6.5, 사내 한정 패턴 / 경로 / env var / file 경로 / reference 모두 0 row)
- **PR description 의 Tier 필드** 명시 (사외/사내/공용)
- **squash merge + branch delete** (PR 머지 시 chore/260618-* branch 일괄 delete)
- **PR 머지 후 `bash scripts/wiki-sync-devhub.sh` 1회 실행** (real mode, AGENTS.md §문서 작업 기준)
- **Cross-reference 정합** per commit (umbrella + ADR + frontmatter + §9 + §13.4)

## 6. Session Stats

- **PRs created**: 6 (5 MERGED + 1 OPEN)
- **Commits**: 35 (27 in PR #645 + 1 in #646 + 1 in #647 + 1 in #648 + 1 in #649 + 3 in #650 + 1 §17.8 stats 정정)
- **Files changed**: ~150 (umbrella doc + ADR-0034 + ADR-0035 + state.json + child doc + wiki mirror 105 file + DOCUMENT_INDEX + planning/README + operation-sop + release notes + 6 stale page)
- **Lines added**: ~60,000 (umbrella doc 5459 + ADR-0034 + ADR-0035 + wiki mirror 11M = ~5300 unique lines + 105 wiki file mirror)
- **Cross-section 정합 fix**: 9+ 위치 per commit
- **Code block 짝수**: 156 = 78쌍 짝수 ✅
- **Tier**: 사외 (vendor-neutral)
- **M-v0.2.0 release 직전 state**: 완전 정합 ✅
