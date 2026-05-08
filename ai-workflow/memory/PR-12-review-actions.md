# PR #12 종합 리뷰 액션 플랜

- 문서 목적: PR #12 (`test/backend-integration` → `main`) 에 대한 self-review, Codex 자동 리뷰, claude code 종합 리뷰를 통합하여 머지 전 처리해야 할 수정 사항을 한 곳에 정리한다.
- 범위: PR #12 코멘트 종합, 우선순위 분류, 검증 체크리스트
- 대상 독자: PR 작성자(ykylee), 리뷰어, 다음 세션 담당자
- 상태: draft
- 최종 수정일: 2026-05-08
- 관련 문서: PR https://github.com/ykylee/Devhub_example/pull/12, `ai-workflow/memory/state.json`, `ai-workflow/memory/session_handoff.md`, `CLAUDE.md`

## 1. 입력 자료 출처

| 출처 | 작성자 | 일시 | 결론 |
| --- | --- | --- | --- |
| Self review (issue comment 4398244076) | ykylee | 2026-05-07 | Approve, Next steps 3건 제안 |
| Codex review (review 4245103190) | chatgpt-codex-connector[bot] | 2026-05-07 | Comment, P1 1건 + P2 1건 |
| Codex line comment 3202430736 | bot | 2026-05-07 | P1 — `organization.go:416` audit 실패 시 500 |
| Codex line comment 3202430740 | bot | 2026-05-07 | P2 — `next.config.ts:3` localhost fallback 제거 |
| 종합 review (issue comment 4401940017) | claude code | 2026-05-08 | Comment, 블로커 3건 외 다수 |

## 2. PR 메타

- 규모: 353 files, +32,087 / -3,187, commits 8
- mergeable: CLEAN, review decision 없음
- CI: `gh pr checks 12` → no checks reported (워크플로우 미가동)

## 3. 액션 항목

### 3.1 블로커 — 머지 전 반드시 처리

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| BLK-1 | 정책 | 사용자 메모리 `feedback_no_docker.md` 와 PR 의 docker 우선 방향이 충돌 | **결정 (2026-05-08): 컨테이너 자산은 git 추적 외부, git 에는 docker/native 분기 가이드만.** 적용 — `.gitignore` DEV ENVIRONMENT 섹션 추가, `docker-compose.yml` 및 3개 `Dockerfile` `git rm --cached`, `docs/setup/environment-setup.md` 신규, README/`docs/README.md` 진입점 추가, `Makefile` `build`/`run` 가이드 위임, `CLAUDE.md` 실행 기본값 갱신, 메모리(`feedback_no_docker.md`) How to apply 보강. | resolved |
| BLK-2 | `frontend/next.config.ts:3` | `BACKEND_API_URL` 미설정 시 기본값이 `http://backend-core:8080` (compose 내부 호스트명) → native 환경 DNS 실패 (Codex P2) | **적용 (2026-05-08): default 를 `http://localhost:8080` 로 복원.** 영향 검토 — `BACKEND_API_URL` 은 next.config.ts 1곳만, client-side 서비스(`rbac`/`realtime`/`websocket`)는 이미 `NEXT_PUBLIC_*` default 가 localhost 라 일관. compose 모드 사용자는 가이드 §3 에 따라 `BACKEND_API_URL` 을 명시 주입하므로 영향 없음. | resolved |
| BLK-3 | `backend-core/internal/httpapi/organization.go` 7곳 (createUser/updateUser/deleteUser/createOrgUnit/updateOrgUnit/deleteOrgUnit/replaceUnitMembers) | 본 mutation 성공 후 `recordAudit` 실패만으로 500 반환 → 클라이언트 재시도 시 중복/충돌 (Codex P1 + 동일 패턴 6곳 전수 점검 결과) | **적용 (2026-05-08): best-effort + 로그.** `audit.go` 에 `recordAuditBestEffort` 헬퍼 추가 (`log.Printf` 사용, backend-core 표준 logger 컨벤션 따름), organization.go 의 7곳 호출을 헬퍼로 교체. 응답 흐름은 `addAuditMeta` 의 기존 `AuditID==""` 가드로 자연 흡수. `gofmt` 통과. organization 다른 핸들러에 동일 패턴 잔존 없음 (grep 검증). 컴파일 검증은 PR 작성자 환경에서 수행 필요(사내 GOPROXY 제약). | resolved |

### 3.2 보안 — backlog 등록 후 phase 종료 전 해소

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| SEC-1 | `frontend/components/layout/AuthGuard.tsx` | "Basic Mock Auth Check" — `useStore().role` 유무로만 redirect, 실제 인증 아님 | **backlog 등록 (2026-05-08).** 코드에 `// SECURITY (SEC-1)` 주석 부착하여 grep 추적 가능. 위치/영향/DoD/후속 분기는 `ai-workflow/memory/claude/test/backend-integration/backlog/2026-05-08.md` 참조. Phase 5.2 종료 전 실제 토큰 검증으로 교체. | tracked |
| SEC-2 | `backend-core/internal/httpapi/auth.go`, `backend-core/main.go` | `BearerTokenVerifier` 가 nil 이면 인증 우회 + **`main.go` 가 verifier 를 전혀 주입하지 않음** + `auth.go` 의 *empty Authorization → c.Next()* 분기로 헤더 미부착 시도도 통과 → 현재 prod 빌드 무인증 통과 | **backlog 등록 (2026-05-08).** auth.go 의 nil 분기와 main.go 의 RouterConfig 생성부 두 곳에 `// SECURITY (SEC-2)` 주석. **2026-05-08 코드베이스 리뷰로 DoD 보강:** empty Authorization 분기 제거 + prod fail-fast 가드 추가. SEC-1, SEC-3, SEC-4 와 동일 sprint 진행 권장. | tracked (보강) |
| SEC-3 | `backend-core/internal/httpapi/auth.go:73`, `router.go` (`/api/v1` 그룹), 모든 mutating 핸들러 | **role(권한) 미적용** — `devhub_actor_role` 이 context 에 set 되지만 어느 핸들러도 읽지 않음. `defaultRBACPolicy()` 매트릭스가 라우트 결정에 미연결. SEC-2 해소만으로는 권한 차등이 동작하지 않음. | **backlog 등록 (2026-05-08).** `router.go` 에 `// SECURITY (SEC-3)` 주석 부착. backlog DoD: 라우트별 role 가드 미들웨어 + 5개 핵심 라우트 매핑 + 통합 테스트. SEC-2 와 동일 sprint. | tracked (신규) |
| SEC-4 | `backend-core/internal/httpapi/commands.go:271-287` | `requestActor` 의 `X-Devhub-Actor` 헤더 fallback — 인증 컨텍스트 미설정 시 헤더 값을 그대로 actor 로 사용 → 무인증 상태에서 audit/command actor 위조. `Warning: 299` 헤더는 거부 메커니즘 아님. | **backlog 등록 (2026-05-08).** `commands.go` 의 requestActor 정의 위에 `// SECURITY (SEC-4)` 주석 부착. backlog DoD: fallback 코드 제거(옵션 A) 또는 dev-only env 가드(옵션 B). SEC-2 해소 후에도 코드 경로 잔존은 회귀 위험이라 함께 정리. | tracked (신규) |
| SEC-5 | `organization.go`, `commands.go`, `audit.go` 등 다수의 5xx 응답 처리부 | **DB 에러 메시지의 응답 본문 노출** — `pgx` 에러를 그대로 `error: err.Error()` 로 응답 → 스키마/제약/내부 SQL 단편 노출. high-confidence 취약점 임계 미달의 defense-in-depth. | **backlog 등록 (2026-05-08).** 코드 마커는 부착하지 않음(패턴 광범위). backlog DoD: prod 5xx 일반화 + `writeServerError` 헬퍼 도입. SEC-1~4 와 분리된 별도 PR 권장. | tracked (신규, medium) |

### 3.3 위생 — 분리 PR 또는 일괄 정리

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| HYG-1 | `.DS_Store` | macOS 부산물 커밋 | **적용 (2026-05-08): `git rm` + `.gitignore` EDITOR/OS NOISE 섹션 추가** (`.DS_Store`, `**/.DS_Store`, `*.bak`, `*.swp`, `Thumbs.db`). | resolved |
| HYG-2 | `ai-workflow/memory/session_handoff.md.bak` | 백업 파일 51줄 포함 | **적용 (2026-05-08): `git rm`.** `.gitignore` 의 `*.bak` 패턴으로 재발 방지. | resolved |
| HYG-3 | 루트 `workflow-source` | **정정 (2026-05-08):** 단순 1줄 텍스트 파일이 아니라 git mode 120000 의 **symlink** (`workflow-source` → `ai-workflow`). Windows 의 git 이 core.symlinks 비활성 상태에서 plain file 로 펼쳐 보여 줘 오인 → `git rm`. 사용자 결정: **symlink 는 사용하지 않는 방향으로 유지**(즉 제거 유지) + 참조 경로는 `ai-workflow/` 로 정리. **후속 처리 (db5bf27 다음 커밋):** `docs/CODE_INDEX.md` 의 12 라인 `workflow-source/` → `ai-workflow/` 일괄 치환, 트리 도식을 단일 `ai-workflow/` 진입점으로 재작성하고 과거 명칭 안내 노트 추가. 보존 대상: `ai-workflow/memory/gemini/phase6/backlog/tasks/2026-05-06_TASK-021.md`(과거 보정 이력), `ai-workflow/releases/Beta-v0.5.0.md`(릴리즈 노트 시점 명령), `docs/archive/AGENTS.md`(archive snapshot), `source-docs/workflow-source/**`(별개 디렉터리 self-참조). | resolved |
| HYG-4 | `tests/devhub_temp_source/standard-ai-workflow-antigravity-v0.4.1-beta.zip` 외 | release artifact 가 git 본체에 포함 | **적용 (2026-05-08): `git rm -r tests/devhub_temp_source/`** (zip + bundle/ + manifest.json + APPLY_GUIDE.md + PACKAGE_CONTENTS.md, 총 11 파일). 향후 release artifact 는 GitHub Releases / LFS 로 분리. | resolved |
| HYG-5 | `ai-workflow/memory/backlog/2026-05-04.md`, `2026-05-06.md` 삭제 | 백로그 이력 손실 위험 | **검증 결과 부분 손실:** TASK-003 (Antigravity) 는 `gemini/phase6/backlog/2026-05-04.md` 에 보존되었으나, **TASK-019 (Codex command/audit schema) 와 TASK-FRONTEND-PHASE5-ORG (Antigravity org UI) 의 plan/act/validation 본문은 새 구조 어디에도 옮겨지지 않았다** (state.json 의 메타와 짧은 implementation log 만 잔존). **적용 (2026-05-08): 원본 두 파일을 `ai-workflow/memory/_archive/legacy-flat-backlog/` 로 복원하고 README.md 로 archive 정책/손실 추정 항목 명시.** | resolved (archive) |

### 3.4 인프라 — 정책 결정에 종속

| ID | 위치 | 문제 | 권장 조치 | 상태 |
| --- | --- | --- | --- | --- |
| INF-1 | `docker-compose.yml`, `frontend/next.config.ts` | docker 의존을 기본 가정으로 만든 변경 | BLK-1 결정으로 `docker-compose.yml` 은 git 추적 해제됨. `frontend/next.config.ts` 만 BLK-2 로 처리. | resolved (compose 측) / open (next.config 측 → BLK-2) |
| INF-2 | DB 포트 5432 → 5433 매핑 | 로컬 PG 와의 충돌 회피 — 정책 유지 시 의미 없음 | `docker-compose.yml` 자체가 git 외부로 이동. 환경별 포트 매핑은 사용자 로컬 자산에서 관리. | resolved |

### 3.5 PR 분할 — 결정 (2026-05-08)

영역별 분량 측정 결과 `source-docs/workflow-source/**` 가 추가 라인의 61% 를 차지하면서 다른 영역과 결합도 0 이라 단독 분리 가성비가 가장 높았다. 사용자 결정: **source-docs 만 분리**.

| 후보 영역 | 결정 | 근거 |
| --- | --- | --- |
| `source-docs/workflow-source/**` (148 files, +19,317) | **분리 → PR #13** (`chore/source-docs-workflow-import`) | 결합도 0, PR #12 코드 리뷰 분량 즉시 절반 이하로 감소 |
| `ai-workflow/**` (121 files, +9.7k/-4.5k) | 본 PR 유지 | 본 PR 의 작업 기록이라 history 일관성 |
| `backend-core/**` (21 files, +1.8k) | 본 PR 유지 | frontend 와 강결합 (rbac/audit/auth API) |
| `frontend/**` (31 files, +1.8k) | 본 PR 유지 | backend API 호출 |
| `docs/**` (15 files, +841) | 본 PR 유지 | backend/frontend 와 결합 |
| `tests/devhub_temp_source/**` | 제거 (HYG-4) | 임시 자료, release artifact |
| infra (docker-compose, Dockerfile) | git 추적 외부 (BLK-1) | 환경별 자산 |

후속 commit 으로 본 PR 의 `source-docs/workflow-source/**` 를 제거하여 PR #13 과 path 충돌을 정리한다.

### 3.6 self-review 의 follow-up 흡수

PR 작성자가 이미 적은 next steps 도 동일 트래커에 흡수.

| ID | 항목 | 매핑 |
| --- | --- | --- |
| FU-1 | DB 비밀번호 등 `.env` 분리 | INF 영역 — docker 트랙 결정 후 처리 |
| FU-2 | API 통신 오류 UI 피드백 강화 | frontend Phase 5.2 backlog |
| FU-3 | docker 빌드 최적화 (volumes 바인드 마운트) | INF 영역 — BLK-1 결정 후 |

PR description 의 향후 작업도 동일 트래커:

- RBAC write API 연동 → backend backlog
- WebSocket 실시간 명령 상태 스트레스 테스트 → 검증 backlog
- main 머지 → 본 문서의 모든 블로커 해소 조건

## 4. 머지 전 검증 체크리스트

PR 코멘트로 결과를 첨부.

- [ ] BLK-1 의사결정 기록 (정책 유지/갱신, 어디에 기록했는지)
- [ ] BLK-2 next.config fallback 수정 후 native dev 환경에서 `/api/*` 통신 확인
- [ ] BLK-3 organization.go audit 실패 패턴 통일 후 update/delete 핸들러 점검
- [ ] SEC-1, SEC-2 backlog 항목 등록 (위치/ID 첨부)
- [ ] HYG-1~5 처리 또는 별도 cleanup PR 발행
- [ ] `cd backend-core && go test ./...` 결과 첨부
- [ ] `cd frontend && npm run build && npx tsc --noEmit` 결과 첨부
- [ ] `pytest ai-workflow/tests/check_docs.py` 결과 첨부
- [ ] PR 분할 결정 기록 (분할 시 신규 PR 번호, 미분할 시 사유)

## 5. 다음 세션 인계 메모

- 본 문서는 main 브랜치 flat 위치(`ai-workflow/memory/PR-12-review-actions.md`)에 둔다. PR #12 가 들고 오는 `ai-workflow/memory/plans/` 디렉터리와 충돌하지 않기 위함.
- PR #12 머지 시 본 문서 status 를 `done` 으로 갱신하거나, 미해결 항목만 `ai-workflow/memory/work_backlog.md` 로 옮기고 본 문서는 archive.
- 신규 코멘트가 추가되면 §1 표에 row 를 덧붙이고 §3 액션 항목을 갱신한다.
