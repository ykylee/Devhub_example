# 세션 인계 문서 (2026-05-12 work_260512-j — PR #80 second-pass review fix)

## 세션 목표

PR #80 (PR-D follow-up) 머지 후 main 에 대한 second-pass 깊은 review 패스에서 발견한 한 가지 silent fallback 회귀를 픽스한다.

## 발견 사항 (full audit)

| # | 항목 | 처리 |
| --- | --- | --- |
| 1 | `router.go::NewRouter` 가 `SetTrustedProxies` 의 err 를 무시 → invalid env 시 partial trust silent | **이번 sprint 보강** |
| 2 | caller-supplied X-Request-ID 콘텐츠 validation | noted, 별도 sprint |
| 3 | ctx 표준 request_id 전파 (client/background log untraced) | noted, 별도 sprint |
| 4 | dev-fallback path 의 log 부재 | 의도된 silent path, fix 불필요 |
| 5 | memoryCommandStore 의 production INSERT 검증 한계 | DB integration test 별도 |

## 픽스 요약

`internal/httpapi/router.go`:
```go
if err := router.SetTrustedProxies(trustedProxiesFromEnv()); err != nil {
    log.Printf("[trusted-proxies] DEVHUB_TRUSTED_PROXIES contains an invalid entry (%v); falling back to attribution-grade default (SetTrustedProxies(nil))", err)
    _ = router.SetTrustedProxies(nil)
}
```

`internal/httpapi/audit_test.go::TestAuditEnrichment_InvalidTrustedProxiesFallsBackToNil` — env=`"203.0.113.42,not-an-ip"` 일 때 X-Forwarded-For ignored, RemoteAddr 가 audit row 의 SourceIP 로 기록되는지 검증.

## 검증

- `cd backend-core && go test ./... -count=1` PASS (전 모듈).

## 다음 슬롯

- caller-supplied X-Request-ID validation
- ctx 표준 request_id 전파
- writeRBACServerError → writeServerError 통합
- M4 진입
