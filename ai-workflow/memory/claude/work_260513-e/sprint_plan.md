# Sprint Plan — claude/work_260513-e (M1 PR-D 정합 마무리)

- 문서 목적: A 묶음 (caller-supplied X-Request-ID validation + ctx 표준 request_id 전파 + writeRBACServerError 통합) 의 작업 계획.
- 범위: backend-core 만. 문서 변경은 매트릭스 row 갱신 + state.json 머지 후 sync.
- 진입 base: main HEAD `ea8df91` (PR #90 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

| 항목 | 위치 | 규모 |
| --- | --- | --- |
| **A3** writeRBACServerError 통합 | `backend-core/internal/httpapi/rbac.go` (11곳 호출 + helper 24줄) | S |
| **A1** caller-supplied X-Request-ID validation | `backend-core/internal/httpapi/request_context.go` (requireRequestID) | S |
| **A2** ctx 표준 request_id 전파 | `request_context.go` (ctx-aware helper) + `kratos_login_client.go:145/153` + `kratos_identity_resolver.go:52` | M |

작업 순서: A3 (격리됨, 0 위험) → A1 (request_context 단독) → A2 (request_context + 2 클라이언트 변경).

## 2. 세부 설계

### A3 — writeRBACServerError → writeServerError 통합

현황 (`errors.go:19` vs `rbac.go:22` 가 사실상 동일 코드):

```go
// errors.go
func writeServerError(c *gin.Context, err error, op string) {
    requestID := requestIDFrom(c)
    if requestID == "" { requestID = "-" }
    log.Printf("server error: op=%s request_id=%s err=%v", op, requestID, err)
    c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "internal error"})
}

// rbac.go (TODO: replace)
func writeRBACServerError(c *gin.Context, err error, op string) {
    log.Printf("server error: op=%s request_id=%s err=%v", op, requestIDFrom(c), err)
    c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": "internal error"})
}
```

조치: `rbac.go:16-24` 의 주석 + 함수 삭제. 11개 호출 (117/179/249/306/328/351/392/405/446/496/505) 을 `writeServerError` 로 일괄 치환. 동작 차이는 빈 request_id 시 "-" vs 빈 문자열 — 로그 가독성 측면에서 후자 (writeServerError) 가 더 명시적이라 회귀 없음.

### A1 — caller-supplied X-Request-ID validation

현황 (`request_context.go:42-50`):
```go
id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
if id == "" { id = generateRequestID() }
```

조치: `validateCallerRequestID(s string) string` 추가. 유효 문자: `A-Za-z0-9_-`, 길이 1..128. 실패 시 빈 문자열 반환 → generateRequestID() 폴백.

근거: `work_260512-j` 발견. net/http 가 CRLF 차단으로 보안 회귀는 아니지만 audit_logs.request_id 위생 보강. printf verb 회귀는 PR-D 후속에서 이미 처리됨 (`logRequest` %s 렌더링).

테스트:
1. empty → generated
2. valid (e.g. `req_abcdef1234`) → preserved
3. control char (`\n`, `\r`, `\t`) → rejected → generated
4. invalid char (`/`, `$`, `=`, space) → rejected → generated
5. too long (> 128) → rejected → generated
6. 응답 헤더가 server-supplied 값과 일치

### A2 — ctx 표준 request_id 전파

현황: `request_context.go::requireRequestID` 가 `c.Set(ctxKeyRequestID, id)` 만 함. background ctx 만 받는 client/helper (kratos_login_client.go, kratos_identity_resolver.go) 는 request_id 를 찾을 길이 없어 untraced log.Printf 만.

조치:
1. `requireRequestID` 에서 `c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKeyRequestIDCtx, id))` 도 호출. 표준 ctx key 추가.
2. ctx-aware helper 두 개 추가:
   - `requestIDFromContext(ctx context.Context) string`
   - `logRequestCtx(ctx context.Context, format string, args ...any)`
3. `requestIDFrom(c *gin.Context)` 도 `c.Request.Context()` fallback 추가 (양방향, 작은 helper 1개로 통일).
4. `kratos_login_client.go::SubmitLogin` 의 `log.Printf("[KratosClient] SubmitLogin 400: ...")` 2건 → `logRequestCtx(ctx, ...)`.
5. `kratos_identity_resolver.go::resolveKratosIdentityID` 의 `log.Printf("[kratos-cache] backfill ...")` → `logRequestCtx(ctx, ...)`.

테스트:
1. ctx 표준 키에 stash 된 request_id 가 `requestIDFromContext` 로 readable
2. `logRequestCtx` 가 ctx 의 request_id 를 prefix 로 사용 (logRequest 와 동일 포맷)
3. ctx 가 nil/empty 일 때 plain log.Printf fallback
4. caller-supplied X-Request-ID 가 gin.Context + ctx 양쪽에 stash 되는지 e2e 미들웨어 테스트

## 3. 검증

- `go test ./backend-core/...` PASS
- 기존 테스트는 모두 회귀 없이 통과해야 함 (특히 audit_test.go, commands_test.go, request_context_test.go).
- 추적성: 매트릭스 §6 변경 이력 + IMPL-audit-request-id 행 보강. 본 sprint 는 IMPL-rbac-error-helper 행 제거 + IMPL-audit-request-id-validation 행 추가 후보.

## 4. 미진입 / 다음 sprint 후보

- E2E 신규 TC (TC-CMD-*, TC-INFRA-*)
- frontend 컴포넌트 Vitest
- RBAC API §12 IMPL 정밀 매핑
- X-Devhub-Actor 폐기 ADR
- document-standards §8 우선순위 3 (본문 ID 명기)
- M3/M4 진입
