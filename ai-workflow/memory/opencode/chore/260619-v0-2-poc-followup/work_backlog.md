# Work Backlog — chore/260619-v0-2-poc-followup

- Branch: `chore/260619-v0-2-poc-followup`
- Agent: opencode (Sisyphus, MiniMax-M3)
- Updated: 2026-06-19 (T01:15+09:00)
- Status: in_progress

## In Progress (P0)

- [ ] **WB-01**: 2606* 원격 브랜치 27개 일괄 정리 (`git push origin --delete`)
- [ ] **WB-02**: 로컬 chore/260618-v0-2-* 3개 삭제 (`git branch -D`)
- [ ] **WB-03**: `git fetch origin --prune` 으로 stale ref 정리
- [ ] **WB-04**: 최종 state 검증 (branch list / work tree / memory 일관성)
- [ ] **WB-05**: 세션 handoff 갱신 (P0 완료 후)

## Pending (P1) — M-v0.2.0 PoC 운영 + sprint 진입

- [ ] **WB-10**: §3 NFR (Non-functional requirements) 작성
  - performance: sync p95 < 5s, query p95 < 200ms
  - security: 봉투 암호화 / audit log
  - availability: RTO 5 / RPO 1h
  - observability: 28 metrics (M-v0.2.3+ production)
- [ ] **WB-11**: §4 유스케이스 작성 (actor × 시나리오)
- [ ] **WB-12**: §5 추적성 매트릭스 (REQ + FR → IMPL → UT → TC)
- [ ] **WB-13**: §2.5 Tier 분리 정공법 (사외/사내 2-tier)
- [ ] **WB-14**: backend-knowledge/ 디렉터리 skeleton 별도 PR (umbrella doc §2.1 + §3.5.3 + §3.8 + §10.1 + §11.1)
- [ ] **WB-15**: §17.2 known gaps 4 row 자연 해소 (PoC +1주 이내)

## Pending (P2) — M-v0.2.0 release 직후

- [ ] **WB-20**: M-v0.2.0 release announcement (사내/사외 배포)
- [ ] **WB-21**: §3.5.6 cross-link reverse index PoC 운영 검증
- [ ] **WB-22**: §3.7.6 data normalization pipeline 자동 검증 운영
- [ ] **WB-23**: §3.6.6 governance audit log 운영 검증

## Pending (P3) — M-v0.2.0 release +2주 ~ +1개월

- [ ] **WB-30**: §15 ADR supersession (M-v0.2.3+ 부터)
- [ ] **WB-31**: §16 API versioning (M-v0.3.0+ 부터 v0-3 도입)
- [ ] **WB-32**: §3.5.7 Pi LLM cross-link 자동 resolution (M-v0.2.3+)
- [ ] **WB-33**: §3.5.8 Pi LLM resolution false positive rollback (M-v0.2.3+)

## Done (2026-06-19)

- [x] **DB-01**: PR #652 Codex P2 review 3건 해소 (Sisyphus/opencode)
- [x] **DB-02**: PR #652 squash merge (`36cac1d6`)
- [x] **DB-03**: chore/260619-v0-2-poc-followup 브랜치 생성 (main @ 36cac1d6 base)
- [x] **DB-04**: branch-specific memory 초기화 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-19.md)
- [x] **DB-05**: §3 NFR 작성 (docs/requirements/v0.2.0-non-functional-requirements.md, 1249 line)
  - 4 NFR category (Performance / Security / Availability / Observability) × 32 cell 매핑
  - 28 metrics 정의 (umbrella §17.5 1:1 정합)
  - 6 incident type runbook (RTO 5분/15분/30분/1시간)
  - 20 endpoint × 4 NFR cross-mapping
- [x] **DB-06**: umbrella doc §13.4 정합 검증 row 추가 (§3 NFR + §2 FR 12/8 → 9/11 정정)
- [x] **DB-07**: §1 도메인 분류 §5 + §2 FR §8 다음 단계 §3 NFR DONE 표시
- [x] **DB-08**: PR #653 생성 (Codex review 자동 trigger 대기)
- [x] **DB-09**: §4 UC 작성 (docs/requirements/v0.2.0-usecases.md, 967 line)
  - 4 actor (End-User / Curator / Operator / System Admin) × 28 시나리오
  - cell × usecase 매트릭스 (32 cell × 4 actor) + endpoint × usecase cross-mapping (20 × 4)
  - Path Y caller-provided user context 흐름 (gateway 3-step orchestration)
  - 5 curator_type 권한 + §3.6.6 audit log 7 event actor attribution
- [x] **DB-10**: PR #653 body 갱신 (§3 NFR + §4 UC 통합본)
