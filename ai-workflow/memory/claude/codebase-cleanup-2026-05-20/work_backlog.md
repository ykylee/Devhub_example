# Work Backlog — sprint `claude/codebase-cleanup-2026-05-20`

- 문서 목적: 본 sprint 의 in-scope / out-of-scope 작업 분류와 후속 carve out 인계.
- 범위: main HEAD `3fdcf33` 기준 코드베이스 self-review + 정리. PR 미발행.
- 대상 독자: 후속 에이전트, 사용자
- 상태: 작업 완료 (PR 보류)
- 최종 수정일: 2026-05-20
- 관련 문서: [본 sprint state](./state.json), [본 sprint 핸드오프](./session_handoff.md)

## 1. In-scope (완료)

### 1.1 Dead code 제거 ✅
- backend: `RouterConfig.KratosWebhookToken` field + main.go 할당 + dead comments + `devhub-backend` 36.9 MB binary
- frontend: `public/*.svg` 5 boilerplate + `identity.service.ts::getTeams() + Team` + `types.ts::BuildLog` + `signup/page.tsx` stale comment + EditBindingModal dead imports
- 기타: `tests/mock_backlog.md` + `tests/repro_validation.py` (broken stub) + empty typo 디렉터리

### 1.2 Outdated docs 갱신 ✅
- `docs/PROJECT_PROFILE.md` + `ai-workflow/memory/PROJECT_PROFILE.md` — gemini/phase6 / codex/service-action-command long-closed sprint 참조 제거 + `make run`/`make build` 정책 모순 정리 + Keycloak (ADR-0019) 명시
- `docs/tech_stack.md` — Next.js 15 → 16, React 19 → 19.2; native default + docker optional 정합
- `docs/setup/environment-setup.md` — draft → stable + 메타 갱신 + `/login` → `/auth/login`
- `.github/workflows/docker-image-publish.yml` — Hydra port 4444 build-arg 제거
- `Makefile` + `README.md` — placeholder 코멘트 + 메타 갱신

### 1.3 Lint 노이즈 감축 ✅
- `eslint.config.mjs` — `coverage/**`, `playwright-report/**`, `test-results/**` globalIgnores 추가
- `EditBindingModal.tsx` — unused imports

## 2. Out-of-scope (carve out 인계)

### 2.1 Pre-existing ESLint errors (P2)
- `app/(dashboard)/admin/topology-v2/page.tsx:86,96` — `any` types (gemini PR #252)
- `app/(dashboard)/admin/settings/integration-bindings/page.tsx:52` — setState in useEffect (gemini PR #251)
- `components/ui/ComboBox.tsx:58` — setSearch in useEffect (gemini PR #251 정합)

Refactor 필요 — feature 동작 회귀 가드 (e2e shard) 동반 필수.

### 2.2 Test 갭 (8건)
| # | 영역 | 우선순위 |
| --- | --- | --- |
| 1 | DREQ intake token admin E2E spec | P0 |
| 2 | JWKS metric emission assertion | P0 |
| 3 | Keycloak 실 OIDC e2e (login → callback → /me) | P1 |
| 4 | Integration bindings PATCH/DELETE handler 테스트 | P1 |
| 5 | Single-port nginx e2e | P1 |
| 6 | frontend service unit (auth/api-client/websocket) | P2 |
| 7 | backend-ai pytest (gRPC 구현 동반) | P3 |
| 8 | dashboard page snapshot 테스트 | P3 |

### 2.3 Docker 자산 모순 (운영팀 검수)
`docker-compose.deploy.yml` ↔ Dockerfile gitignore 정책 — `docker-image-publish.yml` 워크플로우는 환경별 Dockerfile 자산이 있어야 동작. 본 sprint 는 Hydra 4444 build-arg 만 안전 정리.

### 2.4 check_docs.py pre-existing 실패 (gemini 영역)
gemini sprint memory 일부 파일의 메타 헤더 mojibake (cp949 인코딩 잔재) + `docs/development_roadmap.md` 등 broken links. gemini 영역으로 인계.

## 3. 다음 directive

1. 사용자 검토 → PR 발행 → 본인 4단계 self-review (diff 재검토 → gh pr comment → 보강 commit → squash merge)
2. PR 머지 후 main flat memory (`ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`) 갱신 별도 sprint
3. 잔여 P0 test 갭 2건 (DREQ intake admin E2E + JWKS metric) carve out
