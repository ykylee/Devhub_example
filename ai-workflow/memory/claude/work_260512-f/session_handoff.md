# 세션 인계 문서 (2026-05-12 work_260512-f — PR-T3.5 hardening)

## 세션 목표

PR-T3.5 (a18e785, PR #62) 의 follow-up. e2e globalSetup 의 명시된 한계 — "이미 비밀번호가 회전된 상태라면 자동 시드는 변경하지 않는다" — 를 제거한다. password-change.spec 의 finally rollback 이 fatal 실패 시에도 다음 `npm run e2e` 가 시드 비밀번호로 자동 원복하도록 한다.

## 진입 시 상태

- base: `main` HEAD `7e56aad`
- 기존 globalSetup (`frontend/tests/e2e/global-setup.ts`) 의 한계: identity 가 존재하면 password 는 건드리지 않음. password-change 시나리오 fatal 실패 시 stale rotation 잔존 → 다음 실행도 401.

## 픽스

`frontend/tests/e2e/global-setup.ts`:
- `listExistingIdentityEmails(): Set<string>` → `listExistingIdentities(): Map<email, KratosIdentityFull>`.
- 신규 `KratosIdentityFull` 인터페이스 + `resetKratosPassword(identity, seed)`.
- PUT `/admin/identities/{id}` 로 traits/state/metadata_{public,admin} 그대로 echo + `credentials.password.config.password = seed.password`. Kratos 가 plaintext 를 hash. Kratos v26 의 PUT 의미상 명시되지 않은 credential method 는 보존되므로 password 외 method (totp 등) 가 등록되어도 안전.
- `seedKratos()` 분기: 누락 → POST (기존), 존재 → force-reset.

`docs/setup/e2e-test-guide.md`:
- §2.0 자동 시드 단계 설명 — POST / PUT 분기 명시, 기존 "변경하지 않는다" 한계 문구 삭제.
- §6 트러블슈팅 401 / current password incorrect 항목 — "kratos admin identity 로 재설정" → "재실행 1번 자동 복구".
- §8 변경 이력 +1줄.

## 검증

- `cd frontend && npx tsc --noEmit` exit=0.
- 실제 e2e 실행은 사용자 환경 (5 프로세스 native) 에서 별도 검증.

## 결과

stale rotation 한계 제거. password-change.spec 의 finally rollback 이 fatal 실패해도 다음 `npm run e2e` 한 번이면 시드 비밀번호로 자동 복구.

## 다음 슬롯 출발점 후보

1. **PR-T3.5 follow-up #2** — globalTeardown 추가 검토. 본 hardening 으로 거의 불필요해진 듯.
2. **M2 hygiene** — `users.kratos_identity_id` 칼럼 (FindIdentityByUserID O(1)). globalSetup identity_id 매핑도 함께 단순화 가능.
3. **PR-D follow-up** — store/postgres.go commands audit INSERT 3 곳 actor context, 모든 log 라인 request_id, `DEVHUB_TRUSTED_PROXIES` env.
4. **M4 진입** — command status WebSocket UI / 확장 publish+replay / AI Gardener gRPC / Gitea Hourly Pull.
