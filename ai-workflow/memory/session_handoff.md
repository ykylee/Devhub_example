# Session Handoff — main (2026-05-13 진행 중)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-12 EOD 이후 2026-05-13 머지 + 진행 중 sprint
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 브랜치: `main` (HEAD `450cc24`, PR #86 squash 직후)
- 최종 수정일: 2026-05-13
- 상태: M1/M2 100% done + M2 1차 완성 sprint 닫힘 (PR #85). GitHub Actions CI 그린 (PR #86). 진행 중 sprint `claude/work_260513-a` 가 본 sync + FU-CI-2/3/4 처리.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [상태 스냅샷](./state.json), [E2E 가이드](../../docs/setup/e2e-test-guide.md), [배포 가이드](../../docs/setup/test-server-deployment.md)

## 0. 2026-05-13 진행 상태

- main HEAD `450cc24` (PR #86 squash, gemini/prepare-github-action 슬롯 닫힘).
- 2026-05-12 EOD 이후 추가 머지 3건:
  - **PR #84** (`531ff8e`) — `docs(memory): main flat state/handoff EOD update for 2026-05-12` — 어제 EOD 정리. (2026-05-12T08:00Z 머지였지만 #76–83 와 같은 sprint 라인.)
  - **PR #85** (`84b2dde`) — `feat(M2): 1차 완성 sprint — 로드맵 정합 + UX hygiene + Kratos audit (claude/login_usermanagement_finish)`. 4 sub-task 묶음:
    1. 로드맵에 M2 1차 완성 sprint 등재.
    2. UX hygiene (PR-UX1: /admin/settings/users SearchInput 실제 필터링, PR-UX2: /account current_password 라벨 보강, PR-UX3: Header Switch View 한계 안내).
    3. Kratos settings/password/after webhook → audit_logs 통합 (PR-M2-AUDIT).
    4. M2 1차 완성 e2e 게이트 30 TC 묶음 (`account.spec` / `signup.spec` / `signout.spec` / `header-switch-view.spec` / `kratos-audit-webhook.spec` / `admin-users-search.spec` / `admin-permissions.spec` / `rbac-routes.spec` 등 추가).
  - **PR #86** (`450cc24`) — `ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축` (gemini/prepare-github-action). 머지 직전 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개:
    - 5 blockers (`206c16c`): idp-apply-schemas `-dsn` flag, Build App `./` 경로, Start Backend `DB_URL` env 이름, IdP URL 4종 주입, kratos cipher 32 bytes.
    - Follow-on 5: Ory v26.2.0 정합 (`f580feb`), v26 tar nested layout 추출 (`69fe2a0`), GOBIN /usr/local/bin (`4330d87`), DSN/webhook hardcode in yaml (`65353b0`), TC-NAV-02 race fix (`1c94746` → `661af42`).
    - 최종 그린 run 25770711520 (head `1f59a05`).

## 1. 진행 중 sprint — `claude/work_260513-a`

PR #86 머지 직후 후속 sprint. 두 묶음:

1. **main flat memory sync** — 본 문서 + `state.json` + `work_backlog.md` 에 PR #84/#85/#86 반영.
2. **FU-CI-2/3/4** — PR #86 의 work_backlog 에서 인계된 GHA 최적화 3건:
   - FU-CI-2: `playwright install --with-deps chromium` 으로 범위 축소 (Firefox/WebKit 다운로드 제거).
   - FU-CI-3: `actions/cache@v4` 추가 — `~/go/pkg/mod`, `~/.cache/go-build`, `~/.cache/ms-playwright`.
   - FU-CI-4: `Wait for App Readiness` frontend timeout 60s → 120s 분리 상향.
- FU-CI-1 (no-docker policy 정합) 은 정책 결정 선행 필요로 본 sprint 에서 제외, `state.json#next_actions.ci_followups` 에 인계.

## 2. 다음 세션 진입점 — 우선순위 후보

`state.json` 의 `next_actions` 가 단일 source-of-truth. 사용자 결정에 따라 한 축 선택.

| 후보 | 주요 작업 | 규모 | 우선 사유 |
| --- | --- | --- | --- |
| **caller-supplied X-Request-ID validation** | `^req_[A-Za-z0-9_-]{8,64}$` 정규식 강제. control char / newline / 길이 제한. | S | audit_logs.request_id 위생. work_260512-j 발견. |
| **ctx 표준 request_id 전파** | gin Context 의 request_id 를 `context.Value` 로도 흘려서 client/helper 도 자동 tagged. | M | logRequest 의 untraced fallback 해소. PR-D 후속의 자연 마무리. |
| **writeRBACServerError → writeServerError 통합** | rbac.go:22 의 TODO. | S | 코드 정리. |
| **FU-CI-1 no-docker policy** | `.github/workflows/ci.yml` 의 `services: postgres:15` 정합 결정 (예외 명시 vs native PostgreSQL 으로 재작성). | S–M | 정책 일관성. 본 sprint 와 같은 라인. |
| **M4 진입** | command status WebSocket UI, WebSocket 확장 (publish + replay), AI Gardener gRPC, Gitea Hourly Pull worker. | L | 새 기능 트랙. |
| **PR-D follow-up (정합 보강)** | DB integration test 도입 — production INSERT 의 새 enrichment 컬럼이 실제 매칭되는지. | M | 외부 DB 의존. |
| **UX hygiene 후속** | PR #85 의 PR-UX1–3 묶음으로 1차 완성. 후속 UX 발견 시 별도 sprint. | S | UX 보강. |

## 3. 직전 sprint 인계 (2026-05-12 → 2026-05-13)

- [`claude/work_260512-f/g/h/i/j`](./claude/) — 2026-05-12 머지 (PR #76–83). work_260512-h 는 폐기.
- `claude/login_usermanagement_finish` (CLOSED, PR #85) — M2 1차 완성 sprint. 4 sub-task 묶음.
- `gemini/prepare-github-action` (CLOSED, PR #86) — GHA CI 그린 + 리뷰어 모드 2-pass. branch 삭제됨 (`gh pr merge --squash --delete-branch`).
- `claude/work_260513-a` (IN PROGRESS) — 본 문서 sync + FU-CI-2/3/4.

## 4. 환경 / 운영 메모

- **5 프로세스 native 기동**: PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend. `test-server-deployment.md` §5.
- **CI 환경**: GitHub Actions ubuntu-latest. PostgreSQL 은 ephemeral service container (no-docker policy 와 형식적 충돌 — FU-CI-1 에서 의사결정 예정). Hydra/Kratos 는 v26.2.0 바이너리 (prod 와 정합) 를 `scripts/ci-setup.sh` 가 native 설치.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄. globalSetup 이 Kratos identity 3건 + DevHub users 3행 idempotent 시드 + 매 실행마다 시드 비밀번호 force-reset.
- **DEVHUB_TRUSTED_PROXIES**: PR #80 도입 + PR #82 hardening.
- **사내 corp 환경 메모**: `infra/idp/ENVIRONMENT_NOTES.md`.

## 5. 잔여 결정 대기

- FU-CI-1 (no-docker policy 정합) 의사결정 — `services: postgres:` 예외 명시 vs apt postgresql 재작성.
- Hydra JWKS / introspection verifier 실구현 (M2 후속).
- PoC OIDC client 의 secrets 운영 진입 시 교체 절차 (`test-server-deployment.md` §10).
- PoC `test/test` 시드 제거 시점 (운영 진입 직전).
- M4 진입 시점.

## 6. 머지 흐름 요약 (2026-05-12 EOD 이후)

```
450cc24 PR #86 — ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축 (gemini/prepare-github-action)
84b2dde PR #85 — feat(M2): 1차 완성 sprint — 로드맵 정합 + UX hygiene + Kratos audit (claude/login_usermanagement_finish)
531ff8e PR #84 — docs(memory): main flat state/handoff EOD update for 2026-05-12
```
