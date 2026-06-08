# Sprint work backlog — mvs/work_260608-i-frontend-sign-out

## Carve

N-8 (P1-6) v1.0 안정성 — frontend logout status 분기. backend PR #495/#496 의 frontend 잔여.

## 진행 중

- [x] frontend auth.service.ts logout 함수 status 분기 (postBackendLogout + redirectToOIDCEndSession 분리)
- [x] vitest 6 test (UT-frontend-auth-05 FE-01..06) — useStoreAddToast mock 추가
- [x] frontend vitest 80 files / 1030 tests PASS
- [x] e2e auth.spec.ts TC-E2E-LOGOUT-01 추가
- [x] traceability §2.4/§2.5/§2.6/§3.1/§4 cross-ref + 변경이력
- [x] PR body 작성 (pr_body.md)
- [ ] git add + commit + push
- [ ] gh pr create
- [ ] CI wait (type check / lint / vitest / E2E)
- [ ] PR review (codex P1/P2 가능성) + fix
- [ ] PR merge + sprint close

## 후속 (sprint -i 잔여 or 다른 sprint)

- N-8 frontend 잔여 X (본 sprint 으로 종료)
- 다음 우선순위는 `release_v1_roadmap.md` §3.5 N-* 잔여 또는 P1-* 잔여 중 user 지시 시
