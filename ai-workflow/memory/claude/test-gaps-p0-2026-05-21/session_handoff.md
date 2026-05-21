# Session Handoff — sprint `claude/test-gaps-p0-2026-05-21`

- 문서 목적: P0 test gap carve out sprint 의 진척 + 다음 sprint 인계.
- 범위: main `ca0b9ec` 기준. 직전 sprint claude/codebase-cleanup-2026-05-20 (PR #261) 가 인계한 P0 갭 2건 처리 + 사용자 결정 DREQ 보강.
- 상태: 작업 완료, PR 발행 예정
- 최종 수정일: 2026-05-21
- 관련 문서: [본 sprint state](./state.json), [직전 sprint](../codebase-cleanup-2026-05-20/session_handoff.md)

## 1. 진척 요약

| 단계 | 결과 | 커밋 |
| --- | --- | --- |
| P0-1 (DREQ intake admin E2E) — false-positive 정정 | ✅ 직접 검증으로 cover 확인 → DREQ 보강으로 대체 | — |
| P0-2 (JWKS metric assertion) | ✅ done | `116d585` |
| DREQ 보강 1 — IP allowlist UT 보강 | ✅ done | `6fc432f` |
| DREQ 보강 2 — revoke cancel E2E | ✅ done | `6735e92` |
| 본 핸드오프 + push + PR | 🔄 IN PROGRESS | (본 commit) |

## 2. P0-1 false-positive 분석 (학습 1건)

직전 sprint 의 gap 평가 ("DREQ intake admin E2E Playwright spec MISSING") 가 실제로는 `dev-requests.spec.ts` mega lifecycle test 에 8 TC 로 cover 되어 있었음:

- TC-DREQ-ADMIN-TOKEN-01 — issue + plain-1회 modal
- TC-DREQ-INTAKE-AUTH-01 — external Bearer POST 성공
- TC-DREQ-WIDGET-FLOW-01 — assignee 위젯 진입
- TC-DREQ-PROMOTE-TX-01 — ADR-0017 §6 atomicity
- TC-DREQ-ADMIN-TOKEN-REVOKE-01 — revoke
- TC-DREQ-INTAKE-AUTH-NEG-03 — revoked 즉시 401
- TC-DREQ-INTAKE-AUTH-NEG-01 — invalid bearer 401
- TC-DREQ-ADMIN-TOKEN-PATCH-01 — PATCH allowed_ips + flaky 회피 패턴

**원인**: 직전 sprint 의 gap 분석이 subagent 보고에만 의존. 본 sprint 에서 직접 spec 파일 read 로 8 TC 확인.

**학습**: future test-gap audit 시 subagent 보고는 spot-check 필수 (해당 spec 파일 직접 read 1회).

## 3. P0-2 — JWKS metric assertion (commit `116d585`)

### 헬퍼 추가
- `counterVecValue` — delta 패턴 CounterVec reader (audit/metrics_test.go 정합)
- `histogramSampleCount` — Histogram sample-count reader (auth 가 처음 도입)

### 5 stale-while-error test 보강
| Test | result label delta | histogram sample delta |
| --- | --- | --- |
| StaleWhileError_KeycloakUnreachable | `ok` +1 | +1 |
| StaleExpired_Fails401 | `fail` +1 | 0 (fail 분기는 age 미관측) |
| FreshCache_NoStaleFallback | `ok` 0, `fail` 0 | 0 (fresh-path 회귀 가드) |
| StaleCutoff_BasedOnTTLExpiry | `ok` +1 | +1 (PR #242 codex P1 회귀) |
| StaleFallback_DefaultMaxStale | `ok` +1 | +1 (default 24h path) |

## 4. DREQ 보강 — IP allowlist UT (commit `6fc432f`)

### `TestClientIPAllowed` (table test 12 케이스)
empty allowlist / exact IP match·mismatch / CIDR /24 contains·excludes / open CIDR / multiple entries (first-match-wins / no-match) / invalid caller IP / malformed CIDR (no panic) / IPv6 loopback exact / IPv6 CIDR contains.

### `TestIntakeAuth_IPDeniedWhenNotInAllowlist` (middleware-level)
RFC5737 TEST-NET-2/3 (`198.51.100.0/24` + `203.0.113.5`) allowlist + `req.RemoteAddr = "10.99.99.99:54321"` 강제 → 401 `auth_intake_ip_denied` 검증. 기존 empty-allowlist test 가 cover 못한 production 케이스.

### 시행착오 학습 2건
- **gin httptest default RemoteAddr `192.0.2.1:1234`** (TEST-NET-1) — allowlist 에 `192.0.2.0/24` 들어가면 우연 매치. mismatch 보장하려면 (1) AllowedIPs 에 TEST-NET-1 안 넣기 또는 (2) req.RemoteAddr 명시.
- **pure-function unit test 의 가치** — middleware test 가 `c.ClientIP()` 환경 의존 (TEST-NET-1 vs loopback) 이라 edge case 흡수 위해 `clientIPAllowed` 직접 검증이 안정적.

## 5. DREQ 보강 — revoke cancel E2E (commit `6735e92`)

`TC-DREQ-ADMIN-TOKEN-REVOKE-CANCEL-01`:
1. fresh login as system_admin
2. token 발급 (IssueIntakeTokenModal 흐름)
3. row Revoke 버튼 → DestructiveConfirmModal 열기
4. Cancel 버튼 → modal close
5. row 가 active 유지 + revoked badge `toHaveCount(0)`
6. cleanup: `page.evaluate` fetch hard-revoke (OIDC session propagation flake 회피, sprint claude/work_260518-m 패턴)

기존 mega test 가 confirm path 만 cover — cancel path 회귀 가드 추가.

## 6. 검증 (Validation)

| 항목 | 결과 |
| --- | --- |
| `go test ./internal/auth/ -run TestKeycloakJWKSVerifier -v` | ✅ 11 tests (6 기존 + 5 stale-while-error with new assertions) |
| `go test ./internal/httpapi/ -run 'TestIntakeAuth_IP\|TestClientIPAllowed'` | ✅ 13 subtests + 1 middleware test |
| `go test ./...` | ✅ 14 packages |
| `npm run test` (vitest) | ✅ 8 files / 34 tests |
| `npm run lint` | ⚠ 18 problems (4 errors + 14 warnings) — 모두 pre-existing |
| E2E (Playwright) | ⏸ 로컬 skip, CI 검증 |

## 7. 잔여 test gap (다음 sprint 인계)

| # | 영역 | P |
| --- | --- | --- |
| 1 | Keycloak 실 OIDC e2e flow (login → callback → /me → dashboard) | P1 |
| 2 | Integration bindings PATCH/DELETE backend handler 테스트 | P1 |
| 3 | Single-port nginx e2e (ADR-0018) | P1 |
| 4 | frontend service unit (auth.service / api-client / websocket.service) | P2 |
| 5 | backend-ai pytest (gRPC server 구현 동반) | P3 |
| 6 | dashboard page snapshot 테스트 | P3 |

## 8. Pre-existing 4 ESLint errors (cleanup 범위 외)
- `topology-v2/page.tsx:86,96` — `any` types (gemini PR #252)
- `integration-bindings/page.tsx:52` — setState in useEffect (gemini PR #251)
- `ComboBox.tsx:58` — setSearch in useEffect

별도 refactor sprint 진입 시 e2e 회귀 가드 동반.
