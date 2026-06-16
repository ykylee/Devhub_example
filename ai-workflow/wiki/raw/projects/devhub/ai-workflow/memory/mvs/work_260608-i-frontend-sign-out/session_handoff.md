# Sprint session handoff — mvs/work_260608-i-frontend-sign-out

## Carve

issue #488 spec §"Frontend (Gemini — 후속, sprint -i)" — N-8 (P1-6) v1.0 안정성. backend PR #495/#496 의 frontend 잔여.

## 무엇이 결정됐나

- singleton 내 `postBackendLogout` + `redirectToOIDCEndSession` private method 2개로 분리 (단일 책임 + cleanup sequence 명시화)
- backend status 4 outcome 분기: `ok` (204) / `expired` (401, idempotent) / `unreachable` (502) / `error` (5xx/network)
- 502: addToast(error) + 강제 `/login` (OIDC 단계 skip, 정합 우선 — issue spec 권장)
- 5xx/network: addToast(warning) + OIDC redirect (UX 우선)
- cleanup 순서: tokenStore.clear + clearActor + setIsLoggingOut 은 backend 응답 "이후" 로 이동 (Bearer 헤더 보호 — backend 가 access token 검증하는 동안 frontend 가 먼저 비우면 middleware 가 401 로 단락)

## 다음 세션이 가져가야 할 핵심

- PR # — N-8 frontend logout status 분기 (sprint -i)
  - 변경 4 file: auth.service.ts (logout 분리), auth.service.test.ts (6 test), auth.spec.ts (1 e2e), traceability/report.md (cross-ref)
  - 1030 vitest PASS (1024 기존 + 6 신규)
- backend 영향 0 — singleton private method 추가만
- backend PR #496 머지 후속 (Mavis 가 직접)
- codex 외부 review 가능성 — P1/P2 inline 발견 시 hotfix PR

## 인계 SOP

1. PR open 후 CI 결과 wait (cron self-reminder 권장 — `mavis cron self pr-frontend-sign-out-ci-watch --every 3m --ttl 1h`).
2. CI green 시 merge confirm 요청. squash + delete-branch.
3. merge 후 mvs state.json done 갱신 + handoff 마무리.
4. user 가 sprint -i 잔여 다른 작업 지시하면 진행. 없으면 대기 모드.
