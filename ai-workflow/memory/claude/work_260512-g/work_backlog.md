# 작업 백로그

`claude/work_260512-g` 슬롯 — PR-T3.5 follow-up #2 (password-change.spec finally best-effort 화).

## [Planned]

- [ ] PR 생성 + 본인 리뷰 모드 → squash 머지 → main 동기화 → close PR

## [In Progress]

(없음)

## [Done — 이번 세션]

- [x] spec 실행 순서 검증 — audit / auth / password-change / signout (alphabetical). signout 이 alice 재사용 → finally rollback 은 같은 run 안에서 필요.
- [x] `password-change.spec.ts` finally 를 try-catch 로 감싸 best-effort 화. 실패 시 `console.warn` + globalSetup backstop 안내. 헤더 코멘트도 의도 명시로 갱신.
- [x] `cd frontend && npx tsc --noEmit` exit=0.

## [Carried over]

- 단발 일자 (2026-05-12) 의 다음 sprint: M2 hygiene (kratos_identity_id), PR-D follow-up.
