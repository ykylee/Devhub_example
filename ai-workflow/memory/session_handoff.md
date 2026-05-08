# Session Handoff — main (M2 login_action sprint in_review)

- 브랜치: `main` (HEAD `dbff50f`, M1 RBAC track 종료 직후)
- 최종 수정일: 2026-05-08
- 상태: M1 RBAC track 머지 완료. M2 login_action sprint 진입 — PR-LOGIN-1/2 push 머지 대기, PR-LOGIN-3/4 미진입.
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [M2 login_action backlog](./claude/login_action/backlog/2026-05-08.md), [ADR-0001 IdP](../../docs/adr/0001-idp-selection.md), [M1 sprint backlog](./claude/m1-sprint-plan/backlog/2026-05-08.md), [M1 PR 리뷰 actions](./M1-PR-review-actions.md), [ADR-0002 RBAC](../../docs/adr/0002-rbac-policy-edit-api.md), [상태 스냅샷](./state.json), [상위 backlog](./work_backlog.md)

## 0. 직전 활동 요약 (2026-05-08, M2 login_action sprint 진입)

| Phase | PR | 결과 |
| --- | --- | --- |
| 로그인 페이지/기능 검토 | — | 진단: `/login` UI 만 존재, OIDC code flow 의 frontend 4단계 (login UI · callback · token exchange · token storage) + backend Kratos/Hydra proxy 모두 부재. |
| **PR-LOGIN-1** backend `/api/v1/auth/login` proxy | #33 | pushed, OPEN. Kratos public flow + Hydra admin login accept. 9 handler tests. |
| **PR-LOGIN-2** frontend `/auth/login` password form + redirect_uri 통일 | #34 | pushed, OPEN, base = `claude/login_action`. 13 static pages (`/auth/login`, `/auth/error` 추가). |
| PR-LOGIN-3·4 | — | 미진입 |

stack 머지 절차: PR #33 → main → #34 base 변경 → 머지.

## 1. 본 세션 활동 요약 (2026-05-08, M1 RBAC track)

| Sub-phase | PR | 결과 |
| --- | --- | --- |
| M1 sprint planning baseline + SEC-5 mask (PR-A) | #20 | merged |
| ADR-0002 RBAC policy edit API — Option A 채택 (PR-F) | #21 | merged |
| API contract §12 rewrite + 라우트 매핑 표 (PR-G1) | #22 | merged |
| Domain `internal/domain/rbac.go` + `rbac_policies` 마이그레이션 (PR-G2) | #23 | merged |
| Postgres store + `users.role` FK 마이그레이션 (PR-G3) | #29 | merged (FIX-A) |
| RBAC 핸들러 + main.go bootstrap (PR-G4) | #30 | merged (FIX-B + FIX-C) |
| Permission cache + enforceRoutePermission (PR-G5) | #31 | merged |
| Frontend PermissionEditor ↔ backend (PR-G6) | #27 | merged (FIX-D) |
| PR 리뷰 actions 트래커 | #28 | merged |

원본 PR #24·#25·#26 은 stack base 자동 삭제로 close 후 #29·#30·#31 로 main 위 재등록.

main 회귀: `go build / vet / test ./...` PASS.

## 2. M1 진행 상태

- ✅ **RBAC track**: T-M1-01 (SEC-5), T-M1-06 (RBAC 결정), G1~G6 완료. ADR-0002 채택, 4-boolean 모델 통일, deny-by-default + audit invariant + write API + cache 모두 main.
- ⏳ **잔여**: T-M1-02 (envelope/role wire), T-M1-03 (cmd lifecycle), T-M1-04 (audit actor 보강), T-M1-05 (auth_test 매트릭스), T-M1-07 (frontend types 분리), T-M1-08 (WS envelope) → **PR-B/C/D** 로 진입.

## 3. 다음 세션 진입점

### 3.0 우선순위 0 — M2 login_action 잔여 (PR-LOGIN-3/4)

[backlog](./claude/login_action/backlog/2026-05-08.md) §1 의 PR-LOGIN-3·4 spec 참조.

| PR | 작업 | 결정 대기 |
| --- | --- | --- |
| **PR-LOGIN-3** | `/auth/callback` 라우트 + Hydra `/oauth2/token` PKCE 교환 + token 저장 + httpClient 추상화 + Authorization 헤더 자동 부착. 머지 시 첫 e2e 로그인 성공 가능. | **Token 저장 방식**: sessionStorage (SPA 표준, ADR-0001 §6 일치) vs httpOnly cookie via BFF (보안 ↑, backend 가 token 보유). 진입 시 사용자 결정. |
| **PR-LOGIN-4** | logout (Hydra revoke + token 삭제) + `/account` 비밀번호 변경 (Kratos public `/self-service/settings`). | account.service.ts 의 mock 제거. |

머지 순서 권장: #33 → #34 → PR-LOGIN-3 → PR-LOGIN-4.

### 3.1 우선순위 1 — M1 잔여 (PR-B/C/D)

| Track | 작업 | DoD | 우선 |
| --- | --- | --- | --- |
| B·X·F | **PR-B** — API envelope/role wire 일관, frontend types.ts 분리, WebSocket envelope (T-M1-02·07·08) | M1 DoD #1·#7·#8 | 큰 PR. 분할 가능 |
| B | **PR-C** — command lifecycle 6 상태 + dry-run/live + auth_test 보강 (T-M1-03·05) | M1 DoD #2·#4 | 중 |
| B | **PR-D** — audit actor 보강 + request_id 미들웨어 + 마이그레이션 (T-M1-04) | M1 DoD #3 | 중. M1-DEFER-C (`writeRBACServerError` 통합) 흡수 가능 |

### 3.2 우선순위 2 — DEFER A~G (M1 PR 리뷰 후속)

[`M1-PR-review-actions.md` §3](./M1-PR-review-actions.md#3-다음-개발로-넘김--defer):

- **A**: `rbac_policies` is_system ↔ role_id CHECK (P2 방어선)
- **B**: `requireMinRole`/`roleMeetsMin`/`roleRank` deadcode 정리
- **C**: `writeRBACServerError` → `writeServerError` 통합 (1줄 PR — PR-D 흡수 권장)
- **D**: `DeleteRBACRole` row-lock (다중 인스턴스 race)
- **E**: PermissionCache 다중 인스턴스 일관성 (M3+)
- **F**: API contract §12.4 / §12.5 응답 예시 추가
- **G**: MemberTable role display 회귀 사용자 환경 검증

### 3.3 우선순위 3 — 운영 검증 (사용자 환경 의존)

- `make migrate-up` 후 `rbac_policies` 시스템 role 3개 seed 확인
- `DEVHUB_TEST_DB_URL=...` 셋 후 `cd backend-core && go test ./internal/store/...` 의 8 RBAC 통합 케이스 PASS
- frontend Organization > Permissions 탭 e2e: list / matrix toggle / Save / Create custom / Delete custom / 시스템 role 삭제 거부 / role-in-use 거부

### 3.4 우선순위 4 — M2 login_action 운영 검증 (PR-LOGIN-1/2 머지 후)

- Hydra/Kratos 실행 (binary 설치 후 `hydra serve all --config infra/idp/hydra.yaml --dev` + `kratos serve --config infra/idp/kratos.yaml`)
- `infra/idp/scripts/register-devhub-client.ps1` 실행 (devhub-frontend OIDC client 등록 — redirect_uri = `/auth/callback`)
- Kratos identity 1건 생성 + `metadata_public.user_id` 셋 (DevHub `users.user_id` 와 매핑)
- backend `DEVHUB_HYDRA_ADMIN_URL` + `DEVHUB_KRATOS_PUBLIC_URL` 셋 후 재시작
- curl e2e:
  ```
  curl -X POST http://localhost:8080/api/v1/auth/login \
       -H "Content-Type: application/json" \
       -d '{"login_challenge":"<from Hydra>","identifier":"<email>","password":"<pw>"}'
  ```
- frontend e2e (PR-LOGIN-3 머지 후): `/login` → 폼 → `/developer` 진입까지

## 4. 머지된 PR 흐름 (M1 RBAC track)

```
e02ba67 Merge pull request #27 (PR-G6 + FIX-D)
aa14882 fix(frontend): clone defaultRoles for rolesBaseline (M1-FIX-D)
604b578 feat(frontend): wire PermissionEditor to RBAC backend (M1 PR-G6)
02eef35 Merge pull request #31 (PR-G5)
579318f feat(rbac): permission cache + route mapping enforcement (M1 PR-G5)
27b6817 Merge pull request #30 (PR-G4 + FIX-B + FIX-C)
106272d fix(rbac): wire RBACStore in main.go + atomic bulk PUT (M1-FIX-B, M1-FIX-C)
5d8c3bb feat(rbac): RBAC policy + subject role handlers (M1 PR-G4)
24815b8 Merge pull request #29 (PR-G3 + FIX-A)
52538fa fix(rbac): map FK violation to ErrRoleInUse in DeleteRBACRole (M1-FIX-A)
44f5b9c feat(rbac): postgres RBAC store + users.role FK migration (M1 PR-G3)
5239a87 Merge pull request #23 (PR-G2)
1a090a3 Merge pull request #22 (PR-G1)
950a11f Merge pull request #21 (PR-F ADR-0002)
ae8aca1 Merge pull request #20 (PR-A SEC-5)
9bc30c9 Merge pull request #28 (review actions tracker)
```

## 5. 잔여 환경 제약 / 결정 대기

- **사내 GoProxy mirror 의존**: backend-core `go test ./...` 가 `proxy.golang.org` 도달 가능한 환경에서만 PASS. 사용자 사내 환경에서는 확인됨.
- **마이그레이션 적용 필요**: `000005_create_rbac_policies` + `000006_users_role_fkey` 가 처음 적용되는 환경에서는 `make migrate-up` 필요. system role seed 3건 자동.
- **frontend 동작 검증**: 사용자 환경에서 Permission 탭의 e2e 시나리오 (§3.3) 확인 필요.
- **`/auth/callback` 부재**: M2 후속 (변동 없음).
- **검증용 임시 OIDC client**: M0 세션의 잔존 — 그대로.

## 6. M0 결과 (참고용 보존)

| SEC | 상태 |
| --- | --- |
| SEC-1~4 | ✅ resolved (M0 PR #15·16·17·18·19) |
| SEC-5 (DB 에러 노출) | ✅ resolved (M1 PR #20, `writeServerError` 헬퍼) |
