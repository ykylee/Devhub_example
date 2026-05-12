# Session Handoff — main (2026-05-12 EOD)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-12 머지 8건 정리 + 다음 sprint 후보 + 알려진 한계
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 브랜치: `main` (HEAD `29a90bd`, PR #83 squash 직후)
- 최종 수정일: 2026-05-12
- 상태: M1/M2 100% done. E2E 자동 시드 hardening 일단락, PR-D audit 정합 완결. 다음 sprint 결정 대기.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [E2E 가이드](../../docs/setup/e2e-test-guide.md), [배포 가이드](../../docs/setup/test-server-deployment.md)

## 0. 2026-05-12 일과 종료 상태

- main HEAD `29a90bd` (PR #83 squash, work_260512-j close).
- 오늘 머지된 PR 8건: #76 #77 #78 #79 #80 #81 #82 #83. 자세한 목록은 `state.json` 의 `merged_prs_2026_05_12`.
- 폐기 sprint 1건: `work_260512-h` — 후보 (M2 hygiene `kratos_identity_id`) 가 PR-L4 (`809f525`, work_26_05_11-e) 에 이미 머지된 것 확인하여 진입 직후 폐기. handoff candidate 표가 outdated 였다는 관찰.
- 오늘의 두 흐름:
  1. **E2E 자동 시드 hardening**: globalSetup 의 PUT 으로 시드 비밀번호 force-reset (PR #76) → password-change.spec finally 의 best-effort 화 (PR #78). 어떤 중단 상태에서도 다음 `npm run e2e` 가 자동 복구.
  2. **PR-D audit 정합 완결**: commands 흐름의 audit_logs INSERT 도 source_ip/request_id/source_type 채우기, handler/middleware log 라인 request_id 부착, `DEVHUB_TRUSTED_PROXIES` env 도입 (PR #80). second-pass review 에서 발견한 invalid env 시 partial-trust silent 회귀를 nil 폴백 + warning log 로 픽스 (PR #82).

## 1. 다음 세션 진입점 — 우선순위 후보

`state.json` 의 `next_actions` 가 단일 source-of-truth. 사용자 결정에 따라 한 축 선택.

| 후보 | 주요 작업 | 규모 | 우선 사유 |
| --- | --- | --- | --- |
| **caller-supplied X-Request-ID validation** | `^req_[A-Za-z0-9_-]{8,64}$` 정규식 강제. control char / newline / 길이 제한. | S | audit_logs.request_id 위생. work_260512-j 발견. |
| **ctx 표준 request_id 전파** | gin Context 의 request_id 를 `context.Value` 로도 흘려서 `kratos_login_client.go:145/153` / `kratos_identity_resolver.go:52` 등 ctx 만 받는 client 도 자동 tagged. | M | logRequest 의 untraced fallback 해소. PR-D 후속의 자연 마무리. |
| **writeRBACServerError → writeServerError 통합** | rbac.go:22 의 TODO. 형식은 work_260512-i 에서 이미 일치 (`op=%s request_id=%s err=%v`). 통합은 호출처 시그니처만 정리. | S | 코드 정리. |
| **M4 진입** | command status WebSocket UI, WebSocket 확장 (publish + replay), AI Gardener gRPC, Gitea Hourly Pull worker. | L | 새 기능 트랙. |
| **PR-T5 CI 도입** | GitHub Actions: backend `go test` + frontend Vitest + e2e matrix/nightly. | M-L | 사용자가 본 시점 진입 보류 결정 (2026-05-12). 다른 후보 정리 후 재논의. |
| **PR-D follow-up (정합 보강)** | DB integration test 도입 — production INSERT 의 새 enrichment 컬럼이 실제 매칭되는지 (현재는 fake store echo + 코드 검토 + 컴파일 검증). | M | 외부 DB 의존. |
| **M2 hygiene** | Kratos webhook → audit_logs 통합, Hydra JWKS 실구현. | M | 운영 진입 전. |
| **UX hygiene** | `/admin/settings/users` SearchInput 필터, `/account` Kratos privileged session 안내, Header Switch View 한계 안내. | S | UX 보강. |

## 2. 직전 sprint 인계 (2026-05-12 흐름)

- [`claude/work_260512-f`](./claude/work_260512-f/session_handoff.md) (CLOSED, PR #76) — PR-T3.5 hardening. globalSetup PUT force-reset + state="active" 강제 + 가이드 §2.0/§6/§8 갱신.
- [`claude/work_260512-g`](./claude/work_260512-g/session_handoff.md) (CLOSED, PR #78) — PR-T3.5 follow-up #2. password-change.spec finally best-effort 화.
- [`claude/work_260512-h`](.) — DISCARDED. M2 hygiene `kratos_identity_id` 가 PR-L4 에 이미 포함된 것 확인.
- [`claude/work_260512-i`](./claude/work_260512-i/session_handoff.md) (CLOSED, PR #80) — PR-D follow-up 3 sub-task: TRUSTED_PROXIES env / commands audit enrichment / handler log request_id + self-review format-safety fix.
- [`claude/work_260512-j`](./claude/work_260512-j/session_handoff.md) (CLOSED, PR #82) — PR #80 second-pass review fix. `SetTrustedProxies` err 무시 → invalid env 시 nil 폴백.

## 3. 환경 / 운영 메모 (변경 없음, 참고용)

- **5 프로세스 native 기동**: PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend. `test-server-deployment.md` §5.
- **DSN env override**: `infra/idp/{hydra,kratos}.yaml` 의 `dsn` 에 credential 없음.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄. globalSetup 이 Kratos identity 3건 + DevHub users 3행 idempotent 시드 + 매 실행마다 시드 비밀번호 force-reset.
- **DEVHUB_TRUSTED_PROXIES**: PR #80 도입 + PR #82 hardening. 사용법은 `router.go::trustedProxiesFromEnv` doc 코멘트 참조.
- **사내 corp 환경 메모**: `infra/idp/ENVIRONMENT_NOTES.md`.

## 4. 잔여 결정 대기

- PR-T5 CI 도입 시점 (사용자 보류 결정 후 재논의).
- Hydra JWKS / introspection verifier 실구현 (M2 후속).
- PoC OIDC client 의 secrets 운영 진입 시 교체 절차 (`test-server-deployment.md` §10).
- PoC `test/test` 시드 제거 시점 (운영 진입 직전).

## 5. 머지 흐름 요약 (2026-05-12)

```
29a90bd PR #83 — docs(memory): close work_260512-j slot
6121828 PR #82 — fix(audit): fall back to nil trust when DEVHUB_TRUSTED_PROXIES invalid
d40b73a PR #81 — docs(memory): close work_260512-i slot
ffcaec6 PR #80 — feat(audit): PR-D follow-up — commands enrichment + log request_id + TRUSTED_PROXIES
9549395 PR #79 — docs(memory): close work_260512-g slot
6d2274c PR #78 — test(e2e): best-effort cleanup in password-change.spec (PR-T3.5 follow-up)
354609b PR #77 — docs(memory): close work_260512-f slot
e2393a3 PR #76 — test(e2e): force-reset seeded passwords (PR-T3.5 hardening)
```
