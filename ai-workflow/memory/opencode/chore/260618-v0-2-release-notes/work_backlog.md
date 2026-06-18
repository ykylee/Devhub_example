# Work Backlog — v0.2.0 branch `chore/260618-v0-2-release-notes` (2026-06-18)

## P0 (즉시, 다음 세션 진입 시)

- [ ] **PR #650 머지** — `gh pr merge 650 --squash --delete-branch` (3 commit → 1 squash commit, chore/260618-v0-2-release-notes branch delete)
- [ ] **GitHub tag v0.2.0 생성** — `git tag -a v0.2.0 -m 'v0.2.0 backend-knowledge umbrella + ADR-0034/0035'` (main 머지 후 또는 PR 머지 시)
- [ ] **GitHub release publish** — `gh release create v0.2.0 --notes-file docs/release-notes/v0.2.0.md` (release notes 자동 첨부)
- [ ] **사내 SCM sync** — AGENTS.md §6.3 (GitHub main push 후 자동 sync, 사외/사내 2-tier)

## P1 (M-v0.2.0 PoC 운영 시)

- [ ] **§17.2 known gaps 4 row 자연 해소 검증** (M-v0.2.0 PoC 운영 +1주 이내):
  - [ ] row 3 (incident runbook tuning): §11.1.1~§11.1.6 trigger condition 의 false positive / false negative 분석 → §11.3 monitoring 5 지표 의 alert threshold 재조정
  - [ ] row 4 (sprint 진입 checklist 잔여 1 row, backend-knowledge/ 디렉터리 skeleton 별도 PR)
  - [ ] row 5 (Pi SDK npm dependency): 자동 (CI `npm ci`), runtime check `node --version + npm list @earendil-works/pi-coding-agent`
  - [ ] row 6 (backup schedule cron 등록): 자동 (Docker sidecar cron container `backend-knowledge-cron` service in docker-compose.yml, §6.5.1)
- [ ] **§2.5 Tier 분리 정공법 (사외/사내 2-tier 정공법)** — Section 2 디렉터리/코드/commit 의 Tier 분리 정공법 상세화 (사외 코드 + 사내 한정 + 공용 byte-identical drift 검증 + mirror-list.md §1.7/§1.8 정합 + lint-config.toml 정합 + scripts/wiki-sync-devhub.sh 화이트리스트 정합 + §2.4 item 1 + §2.6 + §11.1.1 + §13.4 정합)
- [ ] **backend-knowledge skeleton PR** (별도 PR 로 backend-knowledge/ 디렉터리 skeleton) — Python 3.13+ / FastAPI / OKF / Pi SDK / §2.1 디렉터리 구조 + §2.2 기술 스택 + §3.5.3 bundle 디렉터리 + §3.8 source plugin ABC + §10.1 DB schema + §11.1 incident runbook — §5.3 checklist 잔여 1 row 처리
- [ ] **M-v0.2.0 PoC 운영 시작** — 5종 PoC source plugin + viz.html + DB-based raw + Pi periodic ingest + 운영 runbook 6 incident type

## P2 (M-v0.2.0 release 시점)

- [ ] **M-v0.2.0 release announcement** (사내/사외 배포)
- [ ] **6 stale wiki page 갱신 follow-up** (이미 ✅ done via e6333985)

## P3 (M-v0.2.1+ ~ M-v0.3.0+ 향후 결정 row)

- [ ] **M-v0.2.1 frontend 운영 정책** — 5 page (concept list / detail / ingest / bundles / raw inspector) + viz.html incoming edge visualization (§3.5.6) M-v0.2.1+ 정합 (§5.1 M-v0.2.1 DoD / §12.5 cutover 정책)
- [ ] **M-v0.2.2 backend-ai 폐기 + 6종 source wire** — 10 단계 폐기 절차 (§6.6.2) PR 4 분리 (디렉터리 / Dockerfile / docker-compose / docs) + metrics.py 추가 + 6 step e2e 6종 smoke + alert routing 검증 (§5.1 M-v0.2.2 DoD / §7 Q18)
- [ ] **M-v0.2.3 hrdb + Pi RPC mode + PostgreSQL** — hrdb.py 추가 (7종 source = 7종) + Pi `pi-coding-agent` RPC mode option + PostgreSQL option + cross-link 자동 resolution (Pi LLM 추천, §3.5.6.4) (§5.1 M-v0.2.3 DoD / §6.7 / §10.1)
- [ ] **§1.1 한계 4~7 능동적 강화 timing** — HMAC signature / storage_mode CLI tool / transactional backup / CI contract test 의 후속 milestone 별 scope 결정 (§1.1 / §13.2 known gaps 5 row 자연 해소)

## 결정 / 검증 / 운영 메모

- **Tier: 사외 (vendor-neutral)** — 모든 신규 PR 의 Tier 필드 명시
- **1 commit = 1 logical concept change** — umbrella + ADR 묶음, squash merge
- **code block 짝수 검증** per commit (umbrella doc 본문, 156 = 78쌍 짝수)
- **PR 머지 후 `bash scripts/wiki-sync-devhub.sh` 1회 실행** (real mode)
- **Cross-reference 정합** per commit (umbrella + ADR + frontmatter + §9 + §13.4)
- **Contributor: ykylee** (owner, lead AI agent Sisyphus / MiniMax-M3)
