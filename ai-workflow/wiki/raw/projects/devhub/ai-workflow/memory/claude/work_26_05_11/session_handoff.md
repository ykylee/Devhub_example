# Session Handoff — claude/work_26_05_11

- 문서 목적: `claude/work_26_05_11` 브랜치 sprint 의 세션 간 상태 인계
- 범위: M2 login 마무리(Track L) + 시스템 설정/사용자·조직 관리(Track S)
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: **CLOSED 2026-05-11**. Track L 4 PR + Track S 4 PR + 배포 가이드 1 PR 모두 main 머지. Codex review fix-up 2건 흡수.
- 브랜치: `claude/work_26_05_11` (HEAD `0620689` baseline; 모든 sprint commit 은 sub-stack 브랜치들에서 main 으로 머지됨)
- 최종 수정일: 2026-05-11 (sprint closure)
- 관련 문서: [개발 계획 + 진척](./backlog/2026-05-11.md), [상위 backlog](./work_backlog.md), [상태 스냅샷](./state.json), [통합 로드맵](../../../../docs/development_roadmap.md)

## 0. 현재 기준선

- main 33 커밋 fast-forward 완료, stale 브랜치 3개 정리됨
- 어제 머지된 PR: #38(auth-m2), #39(token exchange), #40·#44(frontend_260510), #42(redesign-concept-a), #43(wiki)
- M2 PR-LOGIN-1·2·3 main 합류 — `/auth/login` 폼, `/auth/callback` PKCE 교환, token-store, `apiClient` Authorization 자동 부착, `AuthGuard` 401 redirect 모두 동작
- M3 일부 main 합류 — `/auth/signup` (HR DB + Kratos), `/users` CRUD UI 와 `MemberTable` 의 사용자 추가 흐름

## 1. 본 sprint 작업 축

본 sprint 는 어제 갱신된 컨셉을 코드에 반영한다.

- 새 컨셉:
  - 역할별 UX 는 **역할별 기본 진입 페이지 우선순위** 로 제공
  - 시스템 대시보드 + 시스템 설정 = `system_admin` 권한 전용 노출
- 사용자 의도:
  - **로그인 마무리** (logout + 비밀번호 변경 실연동) 먼저
  - 다음으로 **시스템 설정 신설 + 사용자/조직 관리 재배치**

## 2. 진척 관리 방식

- 본 sprint 의 단일 source-of-truth 는 [`./backlog/2026-05-11.md`](./backlog/2026-05-11.md). 모든 PR 진입/완료/블록은 그 문서의 §3 DoD 와 §5 체크리스트로 관리한다.
- 상태 라벨은 `planned`, `in_progress`, `blocked`, `done` 4종.
- 검증되지 않은 작업은 `done` 으로 전환하지 않는다.
- 세션 종료 전 `state.json`, 본 문서, 위 backlog 의 §5 체크리스트를 함께 갱신한다.

## 3. 머지 결과 (2026-05-11 closure)

### 3.0 결정 확정

| ID | 채택 | 의미 |
| --- | --- | --- |
| **DEC-1** | **B 분담** | backend = Hydra logout accept + revoke / frontend = Kratos `/self-service/logout/browser` 직접 |
| **DEC-2** | **B frontend 직접** | account.service 가 Kratos public settings flow 3단계 호출. backend audit 미통합 1차 |
| **DEC-3** | **C `/admin/settings` 하위** | `/admin` 인프라 대시보드 유지, `/admin/settings/{users,organization,permissions}` 신설. Sidebar 의 system_admin 그룹 = Dashboard + Settings |

### 3.1 머지 commit (시간순)

1. PR #45 (PR-L1) backend `/auth/logout` — `61541da`
2. PR #51 (PR-L2) frontend logout flow — `858129f`
3. PR #50 (PR-L3) `/account` password — `205980e`
4. PR #49 (deploy guide) — `92b3459`
5. PR #52 (PR-S1) Sidebar gating + role landing — `a2f707e`
6. PR #53 (PR-S2) `/admin/settings` shell + 4-tab move — `2d0075e`
7. PR #54 (PR-S3) accounts admin endpoints — `b935876` (Codex P1+P2 fix 흡수)
8. PR #55 (PR-S4) org coords + leader — `818d54a` (Codex P2 fix 흡수)

main 최종 HEAD: `818d54a`.

### 3.2 다음 sprint 시드 (PR-S3/S4 의 follow-up hygiene)

- Kratos identity 매핑 캐싱 — `FindIdentityByUserID` 가 매번 page-scan. `users.kratos_identity_id` 칼럼 추가 시 O(1).
- Kratos webhook → DevHub audit_logs 통합 (DEC-2=B 후속 — 비밀번호 변경 등 self-service 이벤트)
- Hydra JWKS / introspection verifier 실구현 (backend roadmap M2 P0 잔여)
- `/admin/settings/users` 의 SearchInput 실제 필터링 미구현 (현재 placeholder)

## 4. 운영 의존 메모

- 로그인 e2e: Hydra/Kratos native + DevHub OIDC client 등록 + Kratos identity `metadata_public.user_id` 셋 (가이드: `docs/setup/test-server-deployment.md`)
- frontend 신규 환경변수: `NEXT_PUBLIC_KRATOS_PUBLIC_URL` (logout/account/admin)
- backend 신규 환경변수: `DEVHUB_HYDRA_PUBLIC_URL` (revoke 용)
- main 의 `ai-workflow/memory/state.json` 갱신 — 본 정리 PR 에서 함께 처리

## 5. 다음에 읽을 문서

- [개발 계획 + 진척](./backlog/2026-05-11.md)
- [테스트 서버 배포 가이드](../../../../docs/setup/test-server-deployment.md)
- [통합 로드맵 §3 M2/M3](../../../../docs/development_roadmap.md)
- [frontend 로드맵 §6 사용자/조직](../../../../docs/frontend_development_roadmap.md)
- [backend 로드맵 §5 [P0] M2 / [P1] M3](../../../backend_development_roadmap.md)
- [ADR-0001 IdP §6 sequence](../../../../docs/adr/0001-idp-selection.md)
