# 세션 인계 문서 (Session Handoff)

- 문서 목적: 세션 간 작업 상태 인계 및 다음 단계 제안
- 범위: 최근 작업 완료 사항 및 환경 제약, 차기 권장 사항
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [작업 백로그](./work_backlog.md), [프로젝트 프로파일](../../../../docs/PROJECT_PROFILE.md)

- 작성자: Claude Code
- 현재 브랜치: `claude/init`

## 현재 세션 요약 (Phase 12 풀스택 + Account 컨셉 → Phase 13 IdP 도입 결정)
이번 세션은 (앞부분 — Phase 12 풀스택 검증, CRUD 정비, Account 컨셉 7개 문서 반영) 이후 신규 요구 — **DevHub 의 계정 서비스를 다른 앱에도 제공** — 가 추가됨에 따라 Phase 13 구현 방식을 재검토하고 결정을 ADR 로 기록했다. 자체 `accounts` 테이블 구현 → **Ory Hydra + Kratos IdP 도입** 으로 전환한다 (1순위 채택). 본 결정은 [ADR-0001](../../../../docs/adr/0001-idp-selection.md) 으로 정리됐고, `architecture.md §6.2.3 / §6.3`, `backend_api_contract.md §11`, `backend/requirements.md §5`, `backend/frontend_integration_requirements.md §3.8`, `requirements.md §2.5`, `backend_development_roadmap.md` Phase 13 entry + P1 큐에 모두 반영했다.

## 완료된 사항
1. **Claude Code 환경 검증**:
   - Claude Code `2.1.132` 확인.
   - `.claude/agents/` 4종 (`workflow-orchestrator`, `workflow-doc-worker`, `workflow-code-worker`, `workflow-validation-worker`) 정의 및 sub-agent 위임 패턴 확인.
2. **`.claude/settings.json` 적용**:
   - 사용자 측에서 `! Copy-Item`을 통해 `.claude/settings.json.example` → `.claude/settings.json` 복사 (self-modification 정책 우회).
   - Read 및 git status·diff·log 자동 허용 권한 반영.
   - `/doctor` 보고된 `hooks.$comment` 알 수 없는 hook event 경고 → 빈 `hooks: {}` 로 정정.
3. **브랜치별 메모리 분리**:
   - 신규 디렉터리 `ai-workflow/memory/claude/init/` 생성.
   - `state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/2026-05-07.md` 4종을 브랜치 기준 baseline으로 작성.
   - flat 경로의 stale 항목(TASK-003 = `9553755`로 머지 완료) done 반영, TASK-007은 carried-forward로 분리.

## 진행 중
없음. 이전 in_progress 항목은 모두 검증 통과 후 done 으로 정리됨.

## Phase 12 검증 결과
- **컴파일**: `go run ./backend-core` 통과 — 이전 세션의 미검증 우려 해소.
- **DB 통합**: 마이그레이션 `000004` 적용 후 seed 7 units / 3 users / 4 appointments. 모든 API 응답이 seed 와 일치.
  - `GET /api/v1/users` → 3 users + appointments JOIN 정상
  - `GET /api/v1/organization/hierarchy` → 7 units / 6 edges, 재귀 CTE 의 `direct_count`/`total_count` 정확 (org-root direct=1 total=3)
  - `GET /api/v1/organization/units/team-infra/members` → Sam Jones (예상치 일치)
- **frontend 연동**: Next.js rewrite + `identity.service.ts` 의 backend 응답 매핑 (`unit_id`↔`dept_id`, `system_admin`→`System Admin` 등) 정상. backend gin 로그에서 frontend 호출 확인.

## 다음 세션 작업 제안

Phase 13 은 [ADR-0001](../../../../docs/adr/0001-idp-selection.md) 결정으로 **Ory Hydra + Kratos IdP 도입** 방향이다. 코드 진입 전 ADR-0001 §8 미해결 항목 5종을 먼저 결정해야 한다:

1. **ADR-0001 §8 결정 7종 (이번 세션에 모두 완료, ADR 인라인 반영)**:
   - DB 분리: 기존 `devhub` DB 안에 `hydra`, `kratos` schema 분리 (`?search_path=...`).
   - 외부 앱 신뢰 경계: 단계적 — 1차 사내 first-party only (silent consent, `skip_consent=true`), 외부 SaaS 는 후속 phase.
   - MFA: 1차 미도입 (Kratos schema 만 확장 가능 상태로 유지).
   - `X-Devhub-Actor` 폐기: Phase 13 P1 7단계에서 deprecation warning 시작, 완전 제거는 별도 phase.
   - Gitea SSO: 본 ADR 범위 밖, Phase 13 완료 후 별도 ADR-0002 예정.
   - Binary 획득: `go install github.com/ory/hydra/v2/cmd/hydra@vX.Y.Z` / `…/kratos/cmd/kratos@vX.Y.Z` 를 **사용자 터미널(샌드박스 외) 수동 실행**. 사내 GoProxy 미러 재사용. AI 자동화는 binary 설치 미수행.
   - 호스트 서비스 등록: PoC 1차 직접 실행, 운영 진입 시점 별도 결정.

2. **Phase 13 P1 큐 진입 (결정 완료, 진입 가능)** — backend_development_roadmap.md P1 의 9단계 task:
   - **1단계 PoC** — (a) **사용자 수동 단계**: 샌드박스 외 터미널에서 Ory hydra/kratos `go install`. (b) AI 자동화: `hydra`, `kratos` schema 생성 SQL, `infra/idp/{hydra,kratos}.yaml` 작성, Kratos identity schema 정의, DevHub OIDC client 등록 스크립트, Next.js `/login` round-trip 검증.
   - 2단계 DevHub OIDC client 등록 (PKCE + refresh rotation + skip_consent=true).
   - 3단계 `users.user_id` ↔ Kratos identity 1:1 매핑 검증 어댑터 + 테스트.
   - 4단계 identity ↔ users 동기화 어댑터 (Kratos webhook → Go Core).
   - 5단계 `/api/v1/admin/identities/*` Kratos admin API wrapper.
   - 6단계 Kratos 이벤트 → DevHub audit log 6종 매핑.
   - 7단계 Bearer token 검증 미들웨어 + `X-Devhub-Actor` deprecation warning.
   - 8단계 `backend_api_contract.md §11` 재작성 (Hydra 표준 + admin wrapper + Kratos public flow 3개 절).
   - 9단계 테스트 (round-trip, invariant, audit, admin wrapper).

3. **Phase 12 audit 보강**: 사용자/조직 CRUD HTTP 핸들러에 `X-Devhub-Actor` → `audit_logs` 기록 추가 (Phase 13 token 미들웨어 도입 전 임시 폴백).

4. **Phase 8 WebSocket realtime 엔드포인트**: frontend `realtime.service.ts` 가 호출하는 `/api/v1/realtime/ws` 는 현재 404. gorilla/websocket 또는 nhooyr.io/websocket 도입 검토.

5. **frontend Phase 5 (계정/인증 UI)**: ADR-0001 결정에 따라 화면 책임 재정의 — `/login` 은 Hydra OIDC code flow 시작 + Kratos public flow 호출, `/account` 비밀번호 변경은 Kratos self-service flow, 시스템 관리자 화면은 신규 admin wrapper 호출.

6. **(carried-forward) TASK-007 Gitea Webhook 수신부 상태 재확인** (5일 stale).

## 환경 가동/정리 안내
세션 종료 시점에 다음이 떠있었다 (다음 세션에서 살아있을 수도, 죽었을 수도 있음 — 시작 시 먼저 확인):
- 로컬 PostgreSQL: `localhost:5432`, db `devhub`, user `postgres/postgres`, 마이그레이션 `000001~000004` 적용 완료
- backend-core: `go run .` 백그라운드, `:8080`
- frontend: `npm run dev` 백그라운드, `:3000`
- 종료 명령: 백그라운드 프로세스 PID 를 모르면 `Get-Process node, go | Stop-Process` 또는 PowerShell 창 닫기.
- DB 초기화 필요 시: `psql -U postgres -c "DROP DATABASE devhub"` 후 재생성 + 마이그레이션 재적용.

## 환경 호환 메모
- 사내 SSL inspection 환경: `proxy.golang.org`, `fonts.gstatic.com` 직접 접근 차단. Go 모듈은 GOPROXY 미러 또는 직접 다운로드, npm 은 `cafile`/`NODE_EXTRA_CA_CERTS` 로 사내 CA 등록 필요. Geist 폰트는 system font fallback 으로 교체.
- Next.js rewrite 가 `/api/*` → backend 로 프록시하므로, frontend service 들은 `${API_BASE}` 가 아닌 상대경로 `/api/v1/...` 사용을 권장 (브라우저 CORS 회피).

## 주의 사항
- **legacy flat 경로 보존**: `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/*.md` 등은 legacy 폴백으로 유지한다. 신규 갱신은 모두 `ai-workflow/memory/claude/init/` 아래에서만 수행한다.
- **`.claude/` gitignore**: `.claude/settings.json`은 로컬 전용이며 git 추적 대상이 아니다. 공유가 필요한 정책은 `.claude/settings.json.example`에 반영한다.
- **TASK-007 carried-forward**: 5일간 다른 작업으로 우선순위가 이동했으므로 신규 state.json `in_progress_items`에서 제외했다. 재진입 시 실제 코드/스킬 상태를 다시 검증해야 한다.
- **Phase 12 미검증 커밋**: 빌드 검증을 못한 채 커밋되므로, 다음 세션 첫 행동은 반드시 빌드/테스트 검증으로 시작한다. 컴파일 실패 시 fix-up 커밋으로 처리.

## 다음에 읽을 문서
- [ADR-0001 IdP 선택](../../../../docs/adr/0001-idp-selection.md)
- [backend_development_roadmap.md](../../backend_development_roadmap.md)
- [architecture.md §6.2.3 / §6.3](../../../../docs/architecture.md)
- [backend_api_contract.md §11](../../../../docs/backend_api_contract.md)
- [2026-05-07.md](./backlog/2026-05-07.md)
