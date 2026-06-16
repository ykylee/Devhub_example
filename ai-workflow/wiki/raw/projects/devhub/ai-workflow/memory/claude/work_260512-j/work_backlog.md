# 작업 백로그

`claude/work_260512-j` 슬롯 — PR #80 second-pass review 에서 발견한 SetTrustedProxies err 무시 fix. **CLOSED, PR #82**.

## [Planned]

(없음 — sprint close)

## [Done — 이번 세션]

### 진단
- gin v1.12.0 의 `Engine.parseTrustedProxies` → `prepareTrustedCIDRs` 가 invalid CIDR/IP 만나면 그 시점까지의 valid set 과 error 반환 (gin.go:434-436).
- 우리 코드 `_ = router.SetTrustedProxies(...)` 가 err 무시 → operator 의 typo 가 partial-trust silent 회귀.

### 픽스 (commit `93ae758`)
- `internal/httpapi/router.go` — err 검출 시 log.Printf 경고 + `SetTrustedProxies(nil)` 폴백.
- `internal/httpapi/audit_test.go::TestAuditEnrichment_InvalidTrustedProxiesFallsBackToNil` — env=`"203.0.113.42,not-an-ip"` 폴백 동작 확인.

### 검증
- `go test ./... -count=1` PASS.

### Self-review 패스
- 폴백의 두번째 SetTrustedProxies(nil) — gin 이 nil 입력 정상 처리, err 없음.
- wildcard `*` 영향 없음 — valid CIDR.
- log prefix convention 일관 (component-prefixed).

### 머지
- PR #82 squash → main HEAD `6121828`.

## [Carried over — 다음 슬롯 후보]

- caller-supplied X-Request-ID 콘텐츠 validation
- ctx 표준 request_id 전파 (kratos_login_client / kratos_identity_resolver)
- writeRBACServerError → writeServerError 통합
- M4 진입
