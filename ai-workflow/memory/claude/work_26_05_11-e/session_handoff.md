# Session Handoff — claude/work_26_05_11-e (PR-L4 + PR-T3.5)

- 문서 목적: 진행 중 sprint 상태와 다음 진입점 인계
- 범위: PR-L4 (Kratos session 정합, backend proxy) + PR-T3.5 (e2e seed 자동화)
- 대상 독자: 후속 에이전트
- 브랜치: `claude/work_26_05_11-e` (PR #60 발행, 리뷰 대기) + stacked `claude/work_26_05_11-e-t35` (PR-T3.5 작업 중)
- 상태: PR-L4 PR #60 review_pending. PR-T3.5 코드 완료, e2e 사용자 검증 + PR 발행 대기.
- 최종 수정일: 2026-05-11

## 0. 진입 컨텍스트

- 직전 sprint `work_26_05_11-d` (TDD foundation, PR #58) 가 `password-change.spec.ts` 를 `test.skip` 으로 남김. 본 sprint 가 unblocker.
- main HEAD = `adce7ec`, M1/M2 100% done, M3 partial done.
- 사용자 결정 (2026-05-11):
  - PR-L4 = A안 (backend proxy)
  - PR 두 개 분할
  - `users.kratos_identity_id` 칼럼 흡수

## 1. 작업 순서

1. **L4-A** — `000009_add_kratos_identity_id_to_users.up/down.sql` + OrganizationStore.SetKratosIdentityID + KratosAdminClient 캐싱
2. **L4-B** — KratosClient.SubmitLogin session_token 반환 + auth_login handler 의 cache 저장 (DEC-D 확정 필요)
3. **L4-C** — KratosClient.SettingsFlow (api-mode)
4. **L4-D** — `POST /api/v1/account/password` handler + audit + tests
5. **L4-E** — RBAC self-only
6. **L4-F** — frontend account.service.ts 교체 + Vitest
7. **L4-G** — docs
8. **L4-H** — PR 발행
9. **T3.5-A~E** — globalSetup → unskip → PR

## 2. 결정 (확정)

- DEC-A=A안 / DEC-B=두 PR / DEC-C=kratos_identity_id 흡수
- DEC-D (session_token cache) 잠정 α (in-memory map), L4-B 진입 직전 확정

## 3. 중요한 사전 발견

- `KratosAdminClient.FindIdentityByUserID` 주석이 이미 "production should add ... users.kratos_identity_id column for O(1) lookup" 를 명시 — L4-A 와 정확히 일치
- `KratosClient.SubmitLogin` 의 `parseLoginSuccess` 가 응답 본문의 `session_token` 을 무시 중 → L4-B 에서 노출 필요 (kratosLoginSuccessResponse.SessionToken 은 이미 매핑되어 있으나 KratosIdentity 시그니처가 전달하지 않음)
- frontend `account.service.ts.updateMyPassword` 의 `currentPass` 가 이미 `void` 처리 — backend 로 옮기는 데 호출자 코드 변경 최소

## 4. 다음 세션 시작 시 체크

- `git status` + `git log --oneline -5`
- `ai-workflow/memory/claude/work_26_05_11-e/state.json` (이 파일과 함께 source-of-truth)
- 진행 중 task: TaskList 또는 `backlog/2026-05-11.md` 의 작업 분해 표
