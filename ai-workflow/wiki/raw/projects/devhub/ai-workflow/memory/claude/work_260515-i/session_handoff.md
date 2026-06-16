# Session Handoff — claude/work_260515-i (DREQ-Backend)

- 브랜치: `claude/work_260515-i`
- Base: `main` @ `1d24acf`
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: DREQ 도메인 backend 1차 활성화. ADR-0012 의 옵션 A 인증 + API-59..65 7 endpoint.

## 작업 분해

1. (in progress) 메모리 초기화
2. (planned) `domain/dev_request.go` — DevRequest struct + 6-status enum + target enum + 상태 전이 표
3. (planned) migration 000022/000023/000024 — dev_requests / dev_request_intake_tokens / RBAC dev_requests resource seed
4. (planned) `store/dev_requests.go` + `store/dev_request_intake_tokens.go` — interface + Postgres 구현
5. (planned) `httpapi/dev_request_intake_auth.go` — `requireIntakeToken` middleware
6. (planned) `httpapi/dev_requests.go` — 7 handler (API-59..65) + audit + Promote 단일 트랜잭션
7. (planned) router.go + permissions.go + domain/rbac.go + main.go 와이어링
8. (planned) unit tests (memoryDevRequestStore + handler + middleware)
9. (planned) go build + go test + PR + CI + self-review + merge

## 핵심 패턴 reference

- `applications.go` + `projects.go` — handler 패턴 (storeOrUnavailable / audit / row guard)
- `permissions.go` enforceRowOwnership — ADR-0011 §4.2
- `enforceRoutePermission` + routePermissionTable — RBAC gate
- `accounts_admin.go` 의 password issuance — token plain 1회 노출 패턴
