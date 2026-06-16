# 작업 백로그

`claude/work_260512-i` 슬롯 — PR-D follow-up 3 항목. **CLOSED, PR #80**.

## [Planned]

(없음 — sprint close)

## [In Progress]

(없음)

## [Done — 이번 세션]

### Sub-task #3 DEVHUB_TRUSTED_PROXIES (commit `277b53b`)
- `trustedProxiesFromEnv()` — "" / "none" → nil, "*" → 듀얼스택 any, csv → trim list.
- 단위 테스트 8 case + 통합 테스트 (env=peer IP → X-Forwarded-For 가 audit row 의 진짜 client IP).

### Sub-task #1 commands-flow audit enrichment (commit `7a0e6ce`)
- domain Request 3개에 SourceIP/RequestID/SourceType.
- postgres.go 3 INSERT 컬럼 + VALUES + RETURNING + Scan 갱신.
- commands.go 3 handler 가 clientIPFrom/requestIDFrom/sourceTypeFrom 으로 채워 전달.
- `TestCommandsForwardActorEnrichment` 4 subtest — 4 흐름 모두 store 까지 전파 확인.

### Sub-task #2 handler log request_id (commit `ac22ade`)
- `logRequest(c, format, args...)` helper.
- 13 log 라인 교체 (handler/middleware).
- server error 계열 (permissions/rbac) errors.go 형식 통일.
- unused log import 5개 정리.

### Self-review fix (commit `47c9264`)
- logRequest format-safety: rid 를 `%s` verb 로 렌더링 (이전 문자열 연접 prefix 가 caller-supplied X-Request-ID 의 percent 에 취약).
- 단위 테스트 2건 (percent in id / no id fallback).

### 머지
- PR #80 squash merge → main HEAD `ffcaec6`.

## [Carried over / Deferred — 다음 슬롯 진입 후보]

- **ctx 표준 request_id 전파**: gin.Context 의 request_id 를 `context.Value` 로도 흘려서 client/background (`kratos_login_client.go:145/153`, `kratos_identity_resolver.go:52`) log 도 자동 tagged. logRequest 의 untraced fallback 해소.
- **caller-supplied X-Request-ID validation**: `^req_[A-Za-z0-9_-]{8,64}$` 같은 정규식 강제. format-safety 는 본 PR 에서 해소, 콘텐츠 sanitize 는 별도.
- **writeRBACServerError → writeServerError 통합**: rbac.go:22 의 TODO.
- **M4 진입**: WebSocket UI / 확장 / AI Gardener / Gitea Pull.

## [Sprint B 폐기 기록]

`work_260512-h` 슬롯은 진입 직후 main 확인에서 `kratos_identity_id` 컬럼 + 마이그레이션 000009 + resolveKratosIdentityID 모두 PR-L4 (commit `809f525`, work_26_05_11-e) 에 이미 머지되어 있음을 발견하고 폐기. handoff candidate 표가 outdated 였다.
