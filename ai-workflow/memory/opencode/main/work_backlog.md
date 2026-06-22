# Work Backlog — main (post-sprint cleanup, 2026-06-22)

- Branch: `main` (현재 default branch)
- Agent: opencode (Sisyphus, MiniMax-M3)
- Updated: 2026-06-22 (T00:10+09:00)
- Status: clean (no in-flight work, main 정합 완료)

## In Progress (P0)

없음. main 브랜치 = `196d8593` (PR #662 MERGED), working tree clean, origin/main 정합.

## Pending (P1) — M-v0.2.1+ scope 진입 후보

- [ ] **WB-M01**: M-v0.2.2 backend-ai 폐기 (umbrella §1.2 G2 + §6.6)
  - `backend-ai/` 디렉터리 제거 (placeholder state, 흡수할 코드 0)
  - root `Makefile` / `docker-compose.{local,test,deploy,colima}.yml` 의 backend-ai reference 일괄 정리
  - `docs/` (release notes / ADR / umbrella doc) 의 backend-ai reference 정리
  - umbrella doc §6.6.2 cross-reference 정합 + §9 변경 이력 row 추가
  - PR template 의 `affects-backend-ai: yes` flag (M-v0.2.1+ 도입 검토) 운영 검증
  - 의존: §6.6.1 backend-ai 운영 log snapshot 보존 정책 (umbrella §11.4 정합)

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
- [ ] **WB-M22**: §3.5.7 Pi LLM cross-link 자동 resolution (M-v0.2.3+)
- [ ] **WB-M23**: §3.5.8 Pi LLM resolution false positive rollback (M-v0.2.3+)
- [ ] **WB-M24**: M-v0.2.3 hrdb source plugin + Pi LLM enrich 활성화 (umbrella §1.2 G3 7종 source)
- [ ] **WB-M25**: M-v0.2.3 PostgreSQL option (sqlite → PG migration, §10.1 option)

## Done (2026-06-22)

- [x] **DB-M01**: feat/260619-v0-2-frontend-svelte rebase (-X ours, 4 commit auto-drop)
- [x] **DB-M02**: /src/var/ .gitignore commit (7183d7c6 → 006284ae cherry-pick)
- [x] **DB-M03**: chore/260621-gitignore-src-var 브랜치 생성 + push
- [x] **DB-M04**: feat/260619-v0-2-frontend-svelte reset to origin/main + force-push
- [x] **DB-M05**: PR #662 생성 (사외 tier, 추적성 N/A, worker/opencode label)
- [x] **DB-M06**: PR #662 self-merge (regular merge commit 196d8593)
- [x] **DB-M07**: feat/260619-v0-2-frontend-svelte local + origin 폐기
- [x] **DB-M08**: local main fast-forward to origin/main (34 file / +5568 / -984)
- [x] **DB-M09**: opencode/main/ 메모리 디렉터리 신규 생성 (state.json + session_handoff.md + work_backlog.md)
