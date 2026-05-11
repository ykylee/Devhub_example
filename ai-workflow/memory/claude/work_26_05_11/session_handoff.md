# Session Handoff — claude/work_26_05_11

- 문서 목적: `claude/work_26_05_11` 브랜치 sprint 의 세션 간 상태 인계
- 범위: M2 login 마무리(Track L) + 시스템 설정/사용자·조직 관리(Track S)
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: in_progress (결정 3건 확정 2026-05-11. PR-L1 진입 가능)
- 브랜치: `claude/work_26_05_11` (HEAD `0620689`, main 기준 fast-forward)
- 최종 수정일: 2026-05-11
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

## 3. 다음 세션 진입점

### 3.0 결정 확정 (2026-05-11)

| ID | 채택 | 의미 |
| --- | --- | --- |
| **DEC-1** | **B 분담** | backend = Hydra logout accept + revoke / frontend = Kratos `/self-service/logout/browser` 직접 |
| **DEC-2** | **B frontend 직접** | account.service 가 Kratos public settings flow 3단계 호출. backend audit 미통합 1차 |
| **DEC-3** | **C `/admin/settings` 하위** | `/admin` 인프라 대시보드 유지, `/admin/settings/{users,organization,permissions}` 신설. Sidebar 의 system_admin 그룹 = Dashboard + Settings |

세부 흐름과 영향 위치는 [`./backlog/2026-05-11.md`](./backlog/2026-05-11.md) §2·§3 참조.

### 3.1 진입 순서

1. **PR-L1** backend `/api/v1/auth/logout` (Hydra accept + revoke, Kratos 호출 없음)
2. **PR-L2** frontend `/auth/logout` + Header 재배선 + Kratos browser logout
3. **PR-L3** `/account` 비밀번호 변경 실연동 (Kratos settings flow 직접)
4. **PR-S1** Sidebar 가드 + 역할별 기본 진입 + Sys Admin 그룹 (Dashboard + Settings 분리)
5. **PR-S2** `/admin/settings` 골격 + 사용자/조직/권한 sub-routes 이전
6. **PR-S3** 사용자 관리 강화 (계정 발급/리셋/disable backend 실연동, L3 의존)
7. **PR-S4** 조직 관리 강화 (부서 좌표 영속화 + 리더 변경)

## 4. 위험 / 운영 의존

- 결정 3건 모두 확정. PR 진입 가능.
- 로그인 e2e 는 Hydra/Kratos native + DevHub OIDC client 등록 + Kratos identity metadata 셋 필요 (이전 sprint 와 동일 운영 의존).
- PR-L2/L3 는 frontend 가 Kratos public 에 직접 fetch — `NEXT_PUBLIC_KRATOS_PUBLIC_URL` 환경변수 추가 필요. CORS 는 `kratos.yaml:23-39` 에서 이미 `http://localhost:3000` 허용됨. 운영 배포 시 origin 화이트리스트 갱신 필요.
- PR-S2 머지 시 `/organization` URL → `/admin/settings/users` redirect. docs/wiki 링크 (`docs/wiki/_Sidebar.md` 등) 갱신 필요.
- main 의 `ai-workflow/memory/state.json` 이 `dbff50f` 시점에서 stale — 본 sprint 종료 시 일괄 갱신 PR 별도.

## 5. 다음에 읽을 문서

- [개발 계획 + 진척](./backlog/2026-05-11.md)
- [통합 로드맵 §3 M2/M3](../../../../docs/development_roadmap.md)
- [frontend 로드맵 §6 사용자/조직](../../../../docs/frontend_development_roadmap.md)
- [backend 로드맵 §5 [P0] M2 / [P1] M3](../../../backend_development_roadmap.md)
- [ADR-0001 IdP §6 sequence](../../../../docs/adr/0001-idp-selection.md)
