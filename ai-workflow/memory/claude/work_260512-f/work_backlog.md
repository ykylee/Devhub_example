# 작업 백로그

`claude/work_260512-f` 슬롯 — PR-T3.5 hardening (e2e seed 자동화의 stale rotation 한계 제거).

## [Planned]

- [ ] PR 생성 + 본인 리뷰 모드 (squash 머지) — `claude/work_260512-f` → `main`

## [In Progress]

(없음)

## [Done — 이번 세션]

### 구현
- [x] `frontend/tests/e2e/global-setup.ts` 보강
  - `listExistingIdentityEmails(): Set<string>` → `listExistingIdentities(): Map<email, KratosIdentityFull>` 로 시그니처 변경.
  - `KratosIdentityFull` 인터페이스 추가 (id/schema_id/state/traits/metadata_public/metadata_admin).
  - 신규 `resetKratosPassword(identity, seed)` — PUT `/admin/identities/{id}` 로 traits/state/metadata 그대로 echo + credentials.password 만 plaintext 로 갱신 (Kratos 가 hash). 다른 credential method 는 PUT 시 보존.
  - `seedKratos()` 분기 변경: 누락 identity → POST (기존), 존재하는 identity → force-reset.
- [x] `docs/setup/e2e-test-guide.md` 갱신
  - §2.0 자동 시드 동작 설명 — POST/PUT 분기 명시, "이미 비밀번호가 회전된 상태라면 변경하지 않는다" 한계 문구 제거.
  - §6 트러블슈팅 2건 (401 invalid credentials, /account current password incorrect) — operator 수동 절차 대신 "재실행 1번이면 자동 복구" 로 갱신.
  - §8 변경 이력에 2026-05-12 hardening 1줄 추가.

### 검증
- [x] frontend `npx tsc --noEmit` exit=0
- [ ] 실제 e2e 실행 (`cd frontend && npm run e2e`) — 사용자 환경 (5 프로세스 native 기동 전제) 에서 별도 검증 필요. 본 sprint 의 코드 변경은 globalSetup 1 파일 + 가이드 1 파일이고, Kratos admin PUT 의 의미 (credentials.password 만 갱신, 다른 method 보존) 는 Kratos v26 source 기준 검증. 실행 검증은 PR 코멘트에 결과 첨부 권장.

## [Carried over / Deferred]

- PR-T3.5 follow-up #2: globalTeardown 추가 검토 — 본 hardening 으로 globalSetup 측에서 충분하면 불필요해질 가능성 크다 (재실행만으로 회복).
- M2 hygiene: `users.kratos_identity_id` 칼럼 (FindIdentityByUserID O(1)) — globalSetup identity_id 매핑도 함께 단순화 가능.
- PR-D follow-up: store/postgres.go commands audit INSERT 3 곳 actor context, 모든 log 라인 request_id, `DEVHUB_TRUSTED_PROXIES` env.
