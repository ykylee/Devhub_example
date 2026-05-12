# 작업 백로그

`claude/work_260512-i` 슬롯 — PR-D follow-up (audit 정합 + log request_id + TRUSTED_PROXIES).

## [Planned]

- [ ] PR 생성 + 본인 리뷰 모드 → squash 머지 → main 동기화 → close PR

## [In Progress]

(없음)

## [Done — 이번 세션]

### Sub-task #3 DEVHUB_TRUSTED_PROXIES (commit `277b53b`)
- `internal/httpapi/router.go` — `trustedProxiesFromEnv()` 추가. env "" / "none" → nil (default), "*" → 듀얼스택 any, "csv" → trim list.
- `internal/httpapi/router_test.go` 신규 — table-driven 8 case.
- `internal/httpapi/audit_test.go::TestAuditEnrichment_SourceIPHonoursTrustedProxy` 추가 — env=peer IP 일 때 X-Forwarded-For 가 audit row 의 SourceIP 로 기록.

### Sub-task #1 commands-flow audit enrichment (commit `7a0e6ce`)
- `internal/domain/domain.go` — RiskMitigation / ServiceAction / CommandApproval Request struct 에 SourceIP / RequestID / SourceType 추가.
- `internal/store/postgres.go` — 3 INSERT INTO audit_logs (587/772/1172) 컬럼 + VALUES + RETURNING + Scan 갱신.
- `internal/httpapi/commands.go` — createServiceAction / createRiskMitigation / reviewCommand 에서 enrichment 채워 전달.
- `internal/httpapi/commands_test.go::TestCommandsForwardActorEnrichment` 신규 (4 subtest) — 모든 흐름이 store 로 enrichment 전파.

### Sub-task #2 handler log request_id 부착 (commit `ac22ade`)
- `internal/httpapi/request_context.go::logRequest(c, format, args...)` 추가 — prefix `request_id=req_xxx ` (id 없으면 plain log.Printf).
- handler/middleware 의 13 log 라인 → logRequest 또는 인라인 request_id.
- server error 계열 (permissions.go / rbac.go) 은 errors.go 형식 통일.
- 미부착: kratos_login_client / kratos_identity_resolver (ctx 만 받음). logRequest doc 에 untraced fallback 명시.
- unused "log" import 5개 정리.

### 검증
- `go build ./... ; go test ./... -count=1` PASS (전 모듈).

## [Carried over]

- ctx 에 request_id 를 표준 context.Value 로도 흘리는 광범위 리팩터 — client/background log 도 자동 tagged. 별도 sprint.
- writeRBACServerError 의 TODO ("M1 PR-A 후 writeServerError 통합") — 본 sprint 에서는 형식만 일치, 통합은 별도.

## [Sprint B 폐기 기록]

`work_260512-h` 슬롯은 진입 직후 main 확인에서 `kratos_identity_id` 컬럼 + 마이그레이션 000009 + resolveKratosIdentityID 모두 PR-L4 (commit `809f525`, work_26_05_11-e) 에 이미 머지되어 있음을 발견하고 폐기. 후보 표 outdated 였다.
