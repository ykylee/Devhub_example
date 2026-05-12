# Work Backlog — claude/login_usermanagement_finish

- 문서 목적: 로그인 / 로그아웃 / 사용자 관리 1차 완성 sprint 의 PR 단위 backlog.
- 범위: 잔여 UX hygiene 3건 + M2 follow-up 1건.
- 대상 독자: sprint 담당자, 후속 에이전트.
- 최종 수정일: 2026-05-12
- 상태: OPEN. sprint_plan + 로드맵 3종 정합 완료. 결정 3건 후 진입.
- 관련 문서: [세션 인계](./session_handoff.md), [sprint plan](./sprint_plan.md), [상태](./state.json)

## 1. 진입 우선순위

| 순서 | PR | 의존 | 상태 | 비고 |
| --- | --- | --- | --- | --- |
| 1 | PR-DOCS (로드맵 3종 + sprint_plan + 메모리) | 없음 | 변경 완료, 머지 대기 | 본 sprint 진입의 정합 기준 확정. |
| 2 | PR-UX (묶음) | PR-DOCS 무관 | 진입 대기 | UX-1/2/3 단일 PR 로 확정. |
| 3 | PR-M2-AUDIT | PR-UX 머지 무관 | 진입 대기 | shared secret + 최소 password hook 만 확정. |

## 2. PR 정의

### PR-UX1 — `/admin/settings/users` SearchInput 실 필터링

- 목적: placeholder 였던 SearchInput 을 실 클라이언트 필터로 활성화.
- 파일: `frontend/app/(dashboard)/admin/settings/users/page.tsx`
- 변경 요지:
  - `query` state 추가 + `<input>` 에 `value`/`onChange` 바인딩
  - `MemberTable` 에 넘기는 `members` 를 `query` 로 필터 (login / email / role 부분일치, 대소문자 무시)
  - Filter 버튼: 현재 미구현임을 명시 (tooltip 또는 비활성화). 또는 본 PR 에서는 제거.
- 검증: `cd frontend && npm run build` PASS + 수동으로 사용자 검색이 즉시 반영되는지 확인.
- DoD: 빌드 통과 + 검색어 입력 시 행 수가 즉시 변함.

### PR-UX2 — `/account` Kratos privileged session 안내

- 목적: `current_password` 입력이 필요한 이유와 `REAUTH_REQUIRED` 흐름을 사전에 안내.
- 파일: `frontend/app/(dashboard)/account/page.tsx`
- 변경 요지:
  - Current Password 라벨 아래 1줄 도움말: "Kratos 보안 정책으로, 최근 로그인 직후가 아닌 경우 현재 비밀번호가 필요합니다. 만료 시 Sign In Again 으로 재인증해주세요." 류.
  - 필요 시 라벨에 `*` 표시 또는 small 텍스트.
- 검증: `cd frontend && npm run build` PASS + 시각 확인.
- DoD: 빌드 통과 + 라벨 옆/아래 안내 노출.

### PR-UX3 — Header Switch View 한계 안내

- 목적: dropdown 의 "Switch View" 가 서버 RBAC 권한 변경이 아니라 메뉴 미리보기 시뮬레이션임을 명시.
- 파일: `frontend/components/layout/Header.tsx`
- 변경 요지:
  - "Switch View" 라벨 아래 1줄 보조 텍스트: "메뉴 미리보기 (실제 권한은 서버 actor.role 기준)" 류.
  - 또는 dropdown 헤더에 info 아이콘 + tooltip.
- 검증: `cd frontend && npm run build` PASS + 시각 확인.
- DoD: 빌드 통과 + 한계 안내 노출.

### PR-M2-AUDIT — Kratos self-service webhook → `audit_logs` 통합

- 목적: Kratos 의 password 변경, 로그인 성공/실패, identity 생성 이벤트를 DevHub `audit_logs` 에 기록.
- 사전 조건 (결정 필요):
  - webhook 인증: shared secret (`Authorization: Bearer $DEVHUB_KRATOS_WEBHOOK_TOKEN`) 권장.
  - 어떤 hook 을 등록할지: `settings/password/after`, `login/after` (성공 + 실패), `registration/after`.
- 파일:
  - `backend-core/internal/httpapi/kratos_webhook_handler.go` (신규) — hook payload 파싱 + audit_logs INSERT.
  - `backend-core/internal/httpapi/router.go` — `/api/v1/kratos/hook/{event}` 라우트 등록 + auth middleware.
  - `backend-core/internal/audit/*` — 필요 시 `source_type=kratos` enum 보강.
  - `infra/idp/kratos.yaml` — `selfservice.flows.*.after.hooks` 추가.
  - `docs/setup/test-server-deployment.md` — webhook 운영 안내 §보강.
- 검증:
  - `cd backend-core && go test ./...` PASS (webhook handler 단위 테스트 1+).
  - 수동: kratos password 변경 → `audit_logs` 1행 INSERT 확인.
- DoD:
  - go test PASS + 새 webhook 라우트 단위 테스트 추가.
  - audit_logs 에 source_type=kratos 행이 정상 기록.

## 3. Out-of-scope (이번 sprint 미진입)

- Hydra JWKS / introspection verifier 실구현 → 별도 보안 sprint (`m2_followups.hydra_jwks`).
- PR-T5 CI 도입 → 사용자 보류 결정 (2026-05-12).
- M4 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker).
- main.session_handoff §1 의 PR-D 후속 (caller-supplied X-Request-ID validation / ctx 표준 request_id 전파 / writeRBACServerError 통합) — 본 sprint 의 "로그인/사용자 관리" 범위에서 벗어남.

## 4. 검증 명령 (주의)

- frontend 빌드: `cd frontend && npm run build` (Next.js 16.2.4 기준)
- backend 테스트: `cd backend-core && go test ./...`
- 로컬 e2e: `cd frontend && npm run e2e` (globalSetup 이 시드 자동화)

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-12 | sprint 진입 + 초기 backlog 정의 (PR-UX1/2/3 + PR-M2-AUDIT). |
| 2026-05-12 | sprint_plan.md 신설, 로드맵 3종 정합 갱신, PR-DOCS 단위 추가, 결정 항목 3개로 정리. |
