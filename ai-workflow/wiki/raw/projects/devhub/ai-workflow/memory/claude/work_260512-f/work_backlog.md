# 작업 백로그

`claude/work_260512-f` 슬롯 — PR-T3.5 hardening (e2e seed 자동화의 stale rotation 한계 제거). **CLOSED, PR #76**.

## [Planned]

(없음 — sprint close)

## [In Progress]

(없음)

## [Done — 이번 세션]

### 구현
- [x] `frontend/tests/e2e/global-setup.ts` 보강
  - `listExistingIdentityEmails(): Set<string>` → `listExistingIdentities(): Map<email, KratosIdentityFull>` 시그니처 변경.
  - `KratosIdentityFull` 인터페이스 추가 (id/schema_id/state/traits/metadata_public/metadata_admin).
  - 신규 `resetKratosPassword(identity, seed)` — PUT `/admin/identities/{id}` 로 traits/metadata 그대로 echo + `state="active"` 강제 + credentials.password 만 plaintext 로 갱신 (Kratos 가 hash). 다른 credential method 는 PUT 시 보존.
  - `seedKratos()` 분기: 누락 identity → POST (기존), 존재하는 identity → force-reset.

### 자체 리뷰 패스 (reviewer mode)
- [x] PUT payload 의 `state` echo → `"active"` 강제 (보강 commit `df7b099`). inactive 회귀 안전망.
- [x] 노트 3건 — Kratos PUT method preservation 미래 회귀 가능성, password-change.spec finally redundant 화, 실 e2e 실행 미수행. 모두 PR 코멘트로 박제.

### 문서
- [x] `docs/setup/e2e-test-guide.md` §2.0 / §6 / §8 갱신 — 자동 force-reset 동작 설명 + 트러블슈팅 2건 단순화 + 변경 이력 +1줄.

### 검증
- [x] `cd frontend && npx tsc --noEmit` exit=0 (보강 commit 포함)
- [ ] 실 e2e 실행 (`cd frontend && npm run e2e`) — 사용자 환경 검증 미수행 (5 프로세스 native 기동 필요). PR 코멘트 / test plan 에 unchecked.

### 머지
- [x] PR #76 squash merge → main HEAD `e2393a3`.
- [x] 원격 브랜치 자동 삭제 (`gh pr merge --delete-branch`).

## [Carried over / Deferred — 다음 슬롯 진입 후보]

- **PR-T3.5 follow-up #2**: globalTeardown 추가 검토 — 본 hardening 으로 globalSetup 측이 충분히 회복하므로 사실상 불필요해진 듯. spec finally rollback 단순화 (`password-change.spec.ts` 의 finally 블록 제거 + 코멘트로 의도 명시) 와 묶을 수 있음.
- **M2 hygiene**: `users.kratos_identity_id` 칼럼 추가 (FindIdentityByUserID O(1)) — globalSetup identity_id 매핑도 함께 단순화 가능.
- **PR-D follow-up**: store/postgres.go commands audit INSERT 3 곳 actor context, 모든 log 라인 request_id, `DEVHUB_TRUSTED_PROXIES` env.
- **M4 진입**: command status WebSocket UI / 확장 publish+replay / AI Gardener gRPC / Gitea Hourly Pull.
