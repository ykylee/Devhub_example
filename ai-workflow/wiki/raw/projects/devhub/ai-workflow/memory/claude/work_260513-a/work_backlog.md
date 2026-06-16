# Work Backlog — claude/work_260513-a

- 문서 목적: 본 sprint 의 작업 목록과 진행 상태.
- 최종 수정일: 2026-05-13

## 진행 중

- [ ] FU-CI-2: `.github/workflows/ci.yml` 의 playwright install → `chromium` 한정.
- [ ] FU-CI-3: backend-unit + e2e 잡에 `actions/cache@v4` (go mod / go build / ms-playwright).
- [ ] FU-CI-4: e2e 잡의 frontend readiness timeout 60s → 120s.
- [ ] main flat `ai-workflow/memory/{state.json, session_handoff.md, work_backlog.md}` 에 PR #85/#86 머지 반영 + FU-CI-1 등재.

## 미진입 (다음 sprint 후보)

- [ ] FU-CI-1: No-Docker policy 정합 — `services: postgres:` 컨테이너 유지 vs ubuntu-latest preinstalled PostgreSQL 14 native 기동. 정책 ADR 1 페이지 선행.
- [ ] caller-supplied X-Request-ID validation (정규식 강제). work_260512-j 발견.
- [ ] ctx 표준 request_id 전파 — `kratos_login_client.go:145/153` 등 ctx 만 받는 client.
- [ ] `writeRBACServerError → writeServerError 통합` (`rbac.go:22` TODO).

## 완료

- (없음)
