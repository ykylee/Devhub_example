# Session Handoff — sprint `claude/codebase-cleanup-2026-05-20`

- 문서 목적: 본 sprint 의 진척 상황과 다음 세션 진입점을 인계한다.
- 범위: main HEAD `3fdcf33` 기준 새 브랜치에서 코드베이스 전체 self-review → dead code 제거 + outdated docs 갱신 + lint 노이즈 감축. PR 미발행 상태로 push 까지 완료.
- 대상 독자: 후속 에이전트, 사용자, 본 sprint reviewer
- 상태: 작업 완료, PR 보류 (사용자 지시)
- 최종 수정일: 2026-05-20
- 관련 문서: [본 sprint state](./state.json), [main 핸드오프](../../session_handoff.md), [PROJECT_PROFILE.md](../../PROJECT_PROFILE.md)

## 1. 진척 요약

| 단계 | 결과 | 커밋 |
| --- | --- | --- |
| Phase A — dead code 제거 (backend + frontend + stray dirs) | ✅ done | `dfda392` |
| Phase B — outdated docs 갱신 (7 files) | ✅ done | `d0f136b` |
| Phase C — lint 노이즈 감축 + 미사용 imports 정리 | ✅ done | `eabcba1` |
| Phase D — workflow memory + auto-memory 갱신 | 🔄 in progress | (본 commit) |
| Phase E — 최종 push + PR (보류) | 🔄 next | — |

## 2. 자가 검토 (Self-review)

### Dead code 제거 항목 (12건)

Backend:
- `backend-core/devhub-backend` — 36.9 MB Mach-O ARM64 binary, commit 218be81 부터 잔존. `.gitignore` 의 `dev-bin/` 으로는 커버 안 됨. git rm.
- `backend-core/internal/httpapi/router.go::RouterConfig.KratosWebhookToken` — route 등록 없이 read 도 안 되는 field. main.go 의 빈 string 할당 + 관련 dead comments 함께 정리.
- `backend-core/internal/httpapi/keycloak_events_webhook.go` 의 KratosWebhookToken 패턴 참조 comment.

Frontend:
- `frontend/public/file.svg` + `globe.svg` + `next.svg` + `vercel.svg` + `window.svg` — Next.js boilerplate, 어디서도 import 안 됨.
- `frontend/lib/services/identity.service.ts::getTeams() + Team interface` — mock 만 self-reference, 외부 사용처 zero.
- `frontend/lib/services/types.ts::BuildLog` interface — 참조처 zero.
- `frontend/app/signup/page.tsx` — stale comment 정리. 기존 코멘트가 `four e2e specs (TC-SIGNUP-01..04)` + 'richer form' 을 언급했지만 현재 active spec 은 placeholder 검증 1건 뿐.
- `frontend/components/integration/EditBindingModal.tsx` — 미사용 React imports (`useEffect`, `useMemo`) + dead `payload` cast.

기타:
- `tests/mock_backlog.md` (TASK-999 generic placeholder, 2026-05-01 Antigravity owner) + `tests/repro_validation.py` (broken stub — `SOURCE_ROOT` undefined). `tests/` 디렉터리 자체가 비어 정리.
- empty 디렉터리 (`ai-workflowmemoryclaudework_260518-lbacklog` typo + `backend-core/ai-workflow/memory/claude/work_260520-{k,l}-*` 2개) — 빈 디렉터리.

### Docs 갱신 항목 (7건)

1. `docs/PROJECT_PROFILE.md` — mirror; codex/service-action-command + `make run`/`make build` + docker desktop 정책 모순 정리 + ADR-0019 명시
2. `ai-workflow/memory/PROJECT_PROFILE.md` — source; gemini/phase6 + codex/service-action-command 참조 제거 + native default 정합 + Keycloak/Hydra+Kratos supersession 명시
3. `docs/tech_stack.md` — 메타 2026-05-13 → 2026-05-20; Next.js 15 → 16, React 19 → 19.2.x; React Flow active 명시; Section 2 native default + docker optional
4. `docs/setup/environment-setup.md` — 상태 draft → stable; 메타 2026-05-08 → 2026-05-20; `/login` → `/auth/login`; PR #18 historical 제거
5. `.github/workflows/docker-image-publish.yml` — Hydra port 4444 build-arg 제거 + runtime override 코멘트
6. `Makefile` — test target placeholder 코멘트 갱신 (Vitest/Playwright 이미 live)
7. `README.md` — 메타 2026-05-01 → 2026-05-20 + 핵심 link 추가

### Lint 노이즈 감축

- `frontend/eslint.config.mjs` 의 globalIgnores 에 `coverage/**`, `playwright-report/**`, `test-results/**` 추가 (이미 gitignored 인 generated 디렉터리들)
- 22 problems → 18 (-4 warnings). 4 errors + 14 warnings 잔존 — 모두 pre-existing feature-code (gemini PRs 의 Topology v2 / integration-bindings / ComboBox useEffect 패턴 + topology-v2 의 `any` types). cleanup 범위 외, 별도 refactor 필요.

## 3. 테스트 커버리지 갭 분석 (사용자 지시 반영)

| # | 영역 | 현재 상태 | 우선순위 | 권장 후속 |
| --- | --- | --- | --- | --- |
| 1 | DREQ intake token admin E2E Playwright spec | MISSING (DREQ E2E carve 3/4 잔여) | P0 | `frontend/tests/e2e/dev-request-intake.spec.ts` 신규 |
| 2 | JWKS metric `devhub_jwks_stale_while_error_total{result}` emission 명시 검증 | PARTIAL (옵션 emitter callback만) | P0 | `keycloak_verifier_test.go` 에 metric assertion 추가 |
| 3 | Keycloak 실 OIDC e2e flow (login → callback → /me → dashboard) | MISSING | P1 | `frontend/tests/e2e/auth-oidc-full.spec.ts` 신규 |
| 4 | Integration bindings PATCH/DELETE backend handler 테스트 | MISSING | P1 | `backend-core/internal/httpapi/integration_registry_test.go` 보강 |
| 5 | Single-port nginx e2e (ADR-0018) | MISSING | P1 | 별도 통합 테스트 환경 또는 운영팀 검수 |
| 6 | frontend service 단위 (auth/api-client/websocket) | MISSING | P2 | Vitest 추가 |
| 7 | backend-ai pytest | MISSING (gRPC server 미구현 반영) | P3 | gRPC 구현 시 동반 |
| 8 | dashboard page-level snapshot 테스트 | MISSING | P3 | @testing-library/react 도입 시 |

본 sprint 에서는 신규 test 추가하지 않음 — cleanup 범위 외. 위 매트릭스를 다음 sprint directive 로 인계.

## 4. 검증 (Validation)

| 항목 | 결과 |
| --- | --- |
| `go build ./...` | ✅ silent |
| `go vet ./...` | ✅ silent |
| `go test ./...` | ✅ 14 packages 모두 green |
| `npm run test` (vitest) | ✅ 8 files / 34 tests green |
| `npm run build` (Next.js 16.2.6) | ✅ 26 routes (static + dynamic 혼합) |
| `npm run lint` | ⚠ 18 problems (4 errors + 14 warnings) — 모두 pre-existing feature-code |
| `pytest ai-workflow/tests/check_docs.py` | ⚠ pre-existing gemini sprint memory mojibake + broken links (cleanup 영역 외) |

## 5. Pre-existing issues — cleanup 범위 외

1. **frontend ESLint 4 errors** — Topology v2 (`any` types L86/L96) + integration-bindings page setState in useEffect (L52) + ComboBox useEffect setSearch pattern (L58). gemini PRs 의 feature code; refactor 시 e2e 회귀 가드 필요.
2. **ai-workflow/tests/check_docs.py 의 gemini sprint memory mojibake** — cp949 인코딩 잔재 + broken `docs/development_roadmap.md` links. gemini 영역.
3. **docker-compose.deploy.yml ↔ Dockerfile 모순** — `ghcr.io/ykylee/devhub_example/{backend-core,backend-ai,frontend}` images 를 expect 하지만 해당 Dockerfile 은 `.gitignore` 로 untracked. `docker-image-publish.yml` 워크플로우는 환경별 Dockerfile 자산이 있어야 동작 — 운영팀 검수 필요. 본 sprint 는 Hydra 4444 build-arg 만 정리 (안전한 범위).

## 6. 다음 directive

1. **본 sprint state finalize + 4번째 commit + push** — 본 핸드오프 + work_backlog + state.json + auto-memory MEMORY.md 갱신.
2. **사용자 검토 받기** — PR 발행 보류 상태. 사용자 ack 후 본인 4단계 self-review (diff 재검토 → gh pr comment → 보강 commit → squash merge).
3. **잔여 carve (test 갭 8건)** — 별도 sprint 진입 시 우선순위 부여. P0 2건 (DREQ intake admin E2E + JWKS metric assertion) 부터.

## 7. 본 sprint 커밋 매트릭스

| 커밋 | 영역 | 변경 |
| --- | --- | --- |
| `dfda392` | dead code | 14 files (8 deleted, 6 modified) — backend Kratos residue + frontend boilerplate + stale tests/ |
| `d0f136b` | docs | 7 files modified — PROJECT_PROFILE x2 + tech_stack + environment-setup + docker-image-publish + Makefile + README |
| `eabcba1` | lint | 2 files modified — eslint config + EditBindingModal dead imports |
