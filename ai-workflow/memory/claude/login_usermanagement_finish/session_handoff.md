# Session Handoff — claude/login_usermanagement_finish (2026-05-12 진입)

- 문서 목적: 로그인 / 로그아웃 / 사용자 관리 1차 완성 sprint 의 진입 시점 상태와 결정 대기 항목을 기록한다.
- 범위: 분기 시점 main HEAD, 잔여 항목 목록, 후보 PR 정의, 결정 대기
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 브랜치: `claude/login_usermanagement_finish` (분기 `main` @ `29a90bd`)
- 최종 수정일: 2026-05-12
- 상태: OPEN. sprint_plan + 로드맵 3종 정합 완료. 결정 3건 확정. PR-DOCS 진입 대기.
- 관련 문서: [브랜치 상태](./state.json), [브랜치 백로그](./work_backlog.md), [sprint plan](./sprint_plan.md), [main 상태](../../state.json), [main 인계](../../session_handoff.md), [통합 로드맵](../../../../docs/development_roadmap.md), [프론트 로드맵](../../../../docs/frontend_development_roadmap.md), [백엔드 로드맵](../../backend_development_roadmap.md)

## 0. 분기 시점

- main HEAD `29a90bd` (PR #83 squash 직후). M1/M2 핵심 100% done 상태.
- 분기 사유: 로그인/로그아웃 + 사용자 관리 트랙의 잔여 UX 결함과 audit follow-up 을 묶어 "1차 완성" 게이트를 통과시키기 위함.

## 1. 1차 완성 범위

| ID | 제목 | 규모 | 파일 |
| --- | --- | --- | --- |
| **PR-UX1** | `/admin/settings/users` SearchInput 실 필터링 | S | `frontend/app/(dashboard)/admin/settings/users/page.tsx` |
| **PR-UX2** | `/account` current_password 라벨 Kratos privileged session 안내 | S | `frontend/app/(dashboard)/account/page.tsx` |
| **PR-UX3** | Header Switch View 한계 안내 (actor.role 우회 못함 명시) | S | `frontend/components/layout/Header.tsx` |
| **PR-M2-AUDIT** | Kratos self-service webhook → DevHub `audit_logs` 통합 | M | `backend-core/internal/httpapi/kratos_webhook_handler.go` 신규, `router.go`, `infra/idp/kratos.yaml`, deploy guide |

명시적 out-of-scope:
- Hydra JWKS verifier 실구현 (현행 introspection 동작 안정, 별도 보안 sprint)
- PR-T5 CI 도입 (사용자 보류 결정)
- M4 트랙 진입

## 2. 코드에서 확인된 현재 상태 (gap 검증)

- `/admin/settings/users` SearchInput: 단순 `<input>` 만 렌더링, `value`/`onChange` 없음. Filter 버튼도 dead. 사용자가 검색해도 `members` 배열에 영향 없음.
- `/account` current_password 라벨: "Current Password" 만 출력. `REAUTH_REQUIRED` 처리는 catch 블록에 있으나 사전 안내가 없어 "왜 현재 비밀번호가 필요한가" 가 첫 사용자에게 불명확.
- Header Switch View: `handleRoleChange` 가 frontend store 의 `role` 만 바꾸고 path 만 강제 이동. 서버 측 `actor.role` 은 변경 불가하며 RBAC enforce 우회 못함. 드롭다운 헤더에는 "Switch View" 만 적혀 있어 사용자가 권한 변경으로 오해할 여지.
- Kratos webhook: `backend-core` 내 webhook handler 없음 (grep 0건). self-service password 변경은 현재 `audit_logs` 에 기록되지 않음.
- Hydra JWKS: 현 verifier 는 admin introspection 만 사용 (`backend-core/internal/auth/hydra_introspection.go`). JWKS 기반 verifier 는 미구현 — 본 sprint scope out.

## 3. 결정 (2026-05-12 확정)

1. **PR-UX 단위**: PR-UX1/2/3 **단일 묶음 PR**.
2. **Kratos webhook 인증**: **shared secret** `Authorization: Bearer $DEVHUB_KRATOS_WEBHOOK_TOKEN`. env 누락/불일치 시 401.
3. **Kratos hook 범위**: 최소 **`settings/password/after`** 만. `login/after`, `registration/after` 는 후속 sprint.

## 4. 즉시 다음 행동

1. **PR-DOCS** — 방금 갱신한 로드맵 3종 + sprint_plan + state/handoff/backlog 의 메모리 변경분을 묶어 머지.
2. **PR-UX 묶음** — `users/page.tsx` + `account/page.tsx` + `Header.tsx` 한 PR.
3. **PR-M2-AUDIT** — Kratos password webhook handler + kratos.yaml hook 등록 + deploy guide 갱신.
4. **sprint close** — state/handoff/backlog 종료 표기 + 후속 인계.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-12 | 브랜치 분기 + state/handoff/backlog 초기화. PR-UX1/2/3 + PR-M2-AUDIT 후보 정의. |
| 2026-05-12 | sprint_plan.md 신설 (1차 완성 정의/게이트/PR 단위/검증/위험 명시). 통합/프론트/백엔드 로드맵 3종 정합 갱신. 결정 항목 3개로 정리. PR-DOCS 단위 추가. |
