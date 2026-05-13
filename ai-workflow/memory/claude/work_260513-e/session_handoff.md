# Session Handoff — claude/work_260513-e (2026-05-13)

- 문서 목적: A 묶음 (PR-D 정합 마무리) sprint 인계.
- 범위: backend-core 의 request_id 위생/전파 + writeRBACServerError 통합.
- 진입 base: main HEAD `ea8df91` (PR #90 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 흐름

| 단계 | 결과 |
| --- | --- |
| 0. branch + sprint memory 초기화 | DONE |
| 1. A3 — writeRBACServerError 통합 | pending |
| 2. A1 — caller-supplied X-Request-ID validation | pending |
| 3. A2 — ctx 표준 request_id 전파 | pending |
| 4. go test ./backend-core/... | pending |
| 5. PR open + 2-pass + squash merge | pending |

## 2. 핵심 파일

- `backend-core/internal/httpapi/rbac.go` — A3
- `backend-core/internal/httpapi/errors.go` — A3 destination (writeServerError)
- `backend-core/internal/httpapi/request_context.go` — A1 + A2
- `backend-core/internal/httpapi/request_context_test.go` — A1/A2 단위테스트 추가
- `backend-core/internal/httpapi/kratos_login_client.go` — A2 (라인 145, 153)
- `backend-core/internal/httpapi/kratos_identity_resolver.go` — A2 (라인 52)

## 3. 회귀 위험

- A3: 0 (ID 시그니처 동일, 동작 차이 = 빈 request_id 시 "-" 출력 vs 빈 문자열 → 가독성만 다름).
- A1: 매우 낮음. 기존 단위테스트는 caller-supplied 가 percentage 같은 invalid char 안 쓰므로 비영향.
- A2: 낮음. requireRequestID 가 c.Request.Context() 에 추가 stash → 기존 코드는 c.Get/c.Request.Context() 어느 쪽도 동등 동작.

## 4. 다음 행동

A3 부터 시작. `rbac.go` 의 helper 제거 후 모든 호출을 writeServerError 로 치환.
