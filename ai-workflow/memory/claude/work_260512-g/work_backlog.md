# 작업 백로그

`claude/work_260512-g` 슬롯 — PR-T3.5 follow-up #2 (password-change.spec finally best-effort 화). **CLOSED, PR #78**.

## [Planned]

(없음 — sprint close)

## [In Progress]

(없음)

## [Done — 이번 세션]

- [x] spec 실행 순서 검증 (alphabetical) — audit / auth / password-change / signout. signout 이 alice 재사용 → finally 단순 제거 불가.
- [x] `password-change.spec.ts` 의 finally 를 try-catch 로 감싸 best-effort 화. `console.warn` + 헤더 코멘트로 globalSetup backstop 명시.
- [x] `cd frontend && npx tsc --noEmit` exit=0.
- [x] PR #78 squash merge → main HEAD `6d2274c`.

## [Carried over]

- work_260512-h: M2 hygiene (`users.kratos_identity_id` 컬럼 + 마이그레이션 + backend O(1) lookup).
- work_260512-i: PR-D follow-up (commands audit actor context / log request_id / `DEVHUB_TRUSTED_PROXIES`).
