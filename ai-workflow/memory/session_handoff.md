# Session Handoff — main (2026-05-29, post-cleanup-recovery)

- 문서 목적: PR #407 (cleanup-recovery 통합 sprint) 머지 후 main 상태와 인계 사항.
- 범위: PR #406 의 code-taxonomy SoT (3대 레이어 + 4대 계층) 를 backend-core + frontend 에 실 적용한 Gemini cleanup sprint 의 recovery 결과 통합.
- 상태: PR #407 머지 완료, main HEAD `4a1942e`.
- 최종 수정일: 2026-05-29

## 2026-05-29 (post-PR-407) 결산

### 통합 sprint 5 commit (PR #407 squash → `4a1942e`)

| 분류 | 요지 |
|---|---|
| `chore(infra/deployment-automation)` | e2e + backend-integration job 임시 비활성화 (`if: ... && false` + 복원 SOP) |
| `refactor(backend/test-recovery)` | view cross-package test → httpapi/ 이관, EventCursor import, Repository wrapper, in-memory NewRealtimeTicketStore 복원, router.go shim, providerHasCapability AND→OR 회귀 정정 |
| `fix(frontend/cleanup-recovery)` | P0 ghost import 10 (project.service + integration-provider-presets) + framer-motion mock children type drift + dead file 14 정리 |
| `test(frontend/cleanup-recovery)` | service test +369 (audit/onboarding/dev_request/auth/realtime 신규 + 9 file 보강) + rbac 재작성 (type drift 6건 정정) |
| `docs(memory)` | branch memory state/handoff/work_backlog 갱신 |

### 검증

- Backend `go build ./...` PASS / `go vet ./...` 0 errors / `go test ./... -count=1 -short` 21 패키지 PASS
- Frontend `npx tsc --noEmit` 0 errors / `npx vitest run` 29 file **431 test PASS**
- Frontend coverage statements **21.21% → 28.03%** (overall, service 계층 92~100%)
- CI 활성 5 job 모두 SUCCESS (Detect Paths / Workflow Lint / Migration Prefix / Backend Unit / Frontend Unit), 비활성 2 job (Backend Integration / E2E) SKIPPED

### 본 sprint 의 핵심 학습

1. **Gemini self-report 신뢰성 한계** — "All Stages Completed" + "100% 컴파일 무결성" 자칭이었으나 실제 검증 결과 backend test 11+ compile errors + frontend 43 tsc errors + production code 일부 누락 (realtimeTicket private type / NewRealtimeTicketStore factory). [[feedback_self_review_spec_grep]] 의 "subagent 보고에만 의존했던 게 원인" 패턴 재확인.
2. **Package split 시 helper semantics 회귀 위험** — `providerHasCapability` 가 main HEAD baseline 의 OR 였는데 split 후 한 카피가 AND 로 회귀. 동일 helper 3 카피 정의가 drift 진입점. carve out 으로 통합 권장.
3. **Test file 의 type cast `as XxxType` 의 위험** — Gemini 가 `AuditLogEntry`, `DevRequestRegisterPayload`, `Role.permissions string[]` 등 type 정의 grep 없이 가정 기반 cast → tsc 실패. type 정의 grep 후 정확한 shape 사용 SOP.

## 후속 carve out (work_backlog.md §3 참조)

1. **`providerHasCapability` 3 카피 통합** (P1) — `internal/shared/integrationcaps`
2. **view 컴포넌트 24개 100% coverage** (P1) — 5000+ LoC 별도 sprint
3. **CI 복원** (P1) — `e2e` + `backend-integration` 의 `&& false` 제거
4. **`ApplicationRepository` cross-domain decouple** (P2) — `*IntegrationRepository` embed 제거
5. **`ApplicationStore` interface slim** (P2) — 13+ integration 메서드 제거
6. **shared/utils/last-build.ts + lifecycle-status.ts coverage 0% 조사** (P3)
7. **PR Title Lint CI 도입** (P3) — code-taxonomy.md §0 prefix 컨벤션 강제

## 다음 세션 directive

후속 carve out 우선순위:
- **사내 동반 1순위**: CI e2e + backend-integration 복원 (refactor 정리가 stabilize 됐을 때)
- **claude/Gemini 분담 가능**: providerHasCapability 통합 (P1) → view 컴포넌트 100% coverage 묶음 (P1, 별도 sprint 권장)
- **별도 carve**: ApplicationRepository decouple + ApplicationStore slim (P2, 영향 큼)
