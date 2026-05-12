# Frontend Development Roadmap (Phase 2+)

> ⚠ **먼저 [통합 개발 로드맵](./development_roadmap.md)을 확인하세요.** 본 문서는 그 통합 로드맵의 Frontend 트랙 세부입니다. 마일스톤(M0~M4) / 우선순위(P0~P3) / 트랙 간 의존은 통합 로드맵의 §3·§4.2 가 source-of-truth.

- 문서 목적: 백엔드 API 연동 및 프론트엔드 기능 고도화 로드맵을 정의한다.
- 기준일: 2026-05-04
- 최종 수정: 2026-05-12 (M2 1차 완성 sprint 반영 — Phase 5.2/6/6.1 상태 갱신)
- 상태: in_progress
- 관련 문서: [`./development_roadmap.md`](./development_roadmap.md) (통합), `docs/backend_api_contract.md`, `ai-workflow/memory/backend_development_roadmap.md`

## 1. 개요

프론트엔드 Phase 1에서 구축된 UI와 서비스 레이어 구조를 바탕으로, Phase 2 이후에는 백엔드 API와의 실시간 데이터 연동, 조직/계정 관리, AI 어드바이저와 관리자 액션을 순차적으로 프로덕션 수준으로 끌어올린다. 역할별 UX는 전용 화면 완전 분리보다 역할별 기본 진입 우선순위로 간접 제공한다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 작업 |
| --- | --- | --- | --- |
| **Phase 1** | **done** | 대시보드 UI & Mock 서비스 | 레이아웃 구축, Glassmorphism 적용, Singleton Service 패턴 도입 |
| **Phase 2** | **done** | 핵심 API 통합 | Infra Topology, Risk List, Command/Audit 연동, role wire mapping |
| **Phase 3** | **done** | 실시간성 및 CI/CD 가시화 | WebSocket 통합, CI Run/Logs 연동, 실시간 알림 피드 |
| **Phase 4** | **in_progress** | AI 어드바이저 & 어드민 액션 | AI Gardener 추천 연동, 시스템 관리자 서비스 제어 액션 실체화, command status UI |
| **Phase 5** | **done** | 사용자 및 조직 관리 UI | 사용자 프로필, 팀/조직 단위(Org Units) 관리 UI, 멤버 할당 모달 |
| **Phase 5.1** | **done** | 조직 관리 API 통합 | 백엔드 조직 CRUD 및 멤버 할당 API 연동 |
| **Phase 5.2** | **done** (1차 완성 sprint 진행 중) | 계정 인증 및 IdP 도입 | Ory Hydra/Kratos OIDC code flow + PKCE end-to-end, `/auth/{login,callback}`, `/account`, `/admin/settings/{users,organization,permissions}`. 잔여 UX hygiene/audit 정합은 `claude/login_usermanagement_finish` sprint 에서 종료. |
| **Phase 6** | **done** | 권한 관리(RBAC) UI 고도화 | PermissionEditor `/admin/settings/permissions` 완료 (PR-G6, PR #27). |
| **Phase 6.1** | **done** | RBAC API 통합 | `/api/v1/rbac/policies` 조회/편집 연동, `requirePermission` 라우트 가드 (M1 RBAC track, PR #20·21·22·23·27·29·30·31). |
| **Phase 7** | **planned** | 통합 검증 및 고도화 | AI Gardener 추천 연동 고도화, 전역 감사 로그 연동 |


## 3. Phase 2 상세 계획 (Core API Integration)

### 3.1 Infra Topology 연동

- **목표**: 인프라 그래프가 실제 런타임 상태를 반영하도록 함.
- **API**: `GET /api/v1/infra/topology`
- **핵심 로직**:
    - 백엔드의 `RuntimeSnapshotProvider`에서 제공하는 Node 상태(`stable`, `warning`, `down`)를 UI 색상 및 애니메이션으로 변환.
    - `cpu_percent`, `memory_bytes` 등 원시 데이터를 프론트엔드 유틸리티를 통해 포맷팅.

### 3.2 Risk Management 연동

- **목표**: 리스크 목록을 DB에서 가져오고, 리스크 완화 조치를 감사 가능한 형태로 처리.
- **API**: `GET /api/v1/risks`, `POST /api/v1/risks/{id}/mitigations`
- **핵심 로직**:
    - 리스크 완화 버튼 클릭 시 `idempotency_key`를 생성하여 중복 요청 방지.
    - 비동기 `command_id`를 수신하여 처리 중 상태를 UI에 표시.

### 3.3 Dashboard Metrics 연동

- **목표**: 역할별 기본 진입 페이지에서 보여줄 메트릭 카드를 실제 통계 데이터로 업데이트.
- **API**: `GET /api/v1/dashboard/metrics?role={role}`

## 4. 기술적 고려 사항

- **Error Handling**: API 호출 실패 시 Toast 메시지 노출 및 Graceful Degradation 적용.
- **Loading States**: `Framer Motion`을 활용한 스켈레톤 UI 및 로딩 애니메이션 유지.
- **Environment**: `NEXT_PUBLIC_API_URL`, `NEXT_PUBLIC_WS_URL` 환경 변수로 백엔드 주소 관리.

## 5. Phase 4 상세 계획 (AI 어드바이저 & 어드민 액션)

- `infra.service.ts`, `risk.service.ts` 실제 API 연동은 완료.
- System Admin service action command 생성 연동은 완료.
- 백엔드 `/api/v1/realtime/ws`와 `command.status.updated` publish 경계가 구현됨.
- 다음 프론트 작업은 기존 `RealtimeService` 구독을 command toast/status UI에 연결하는 것이다.
- AI Gardener suggestion API/UI 연결 범위는 아직 확정 필요.

## 6. Phase 5 상세 계획 (사용자/조직 관리 + 자체 계정 인증)

DevHub 자체 사용자 계정(Account) 1:1 컨셉 도입 (`docs/requirements.md` 2.5, `docs/architecture.md` 6.2 참조)에 따른 프론트 작업.

### 6.1 로그인 / 인증 흐름

- **목표:** `/login` 진입점 + 인증 가드 + `must_change_password=true` 라우팅.
- **API:** Hydra/Kratos login/logout flow, `GET /api/v1/me` (구현됨).
- **핵심 로직:**
    - 로그인 폼은 `login_id` + `password`만 받는다.
    - 응답 `must_change_password=true`면 즉시 `/account/password`로 강제 라우팅한다.
    - 세션 토큰은 백엔드 결정에 맞춰 secure cookie 또는 메모리/스토리지 중 하나로 보관한다.

### 6.2 내 계정 화면 (`/account`)

- **목표:** 본인이 자신의 로그인 ID 확인 및 비밀번호 변경.
- **API:** `GET /api/v1/accounts/{user_id}`, `PUT /api/v1/accounts/{user_id}/password` (본인 변경 페이로드).
- **핵심 로직:**
    - login ID는 읽기 전용 표시.
    - 비밀번호 변경 폼은 `current_password` + `new_password` + `confirm_new_password` 3 필드.
    - 어떤 화면도 비밀번호 평문을 표시/저장하지 않는다.

### 6.3 시스템 관리자 계정 관리 패널

- **목표:** 사용자 row 옆에서 계정 발급/회수/잠금 해제/강제 재설정 수행.
- **API:** `POST /api/v1/accounts`, `PATCH /api/v1/accounts/{user_id}`, `PUT /api/v1/accounts/{user_id}/password` (`force=true`), `DELETE /api/v1/accounts/{user_id}`.
- **핵심 로직:**
    - "계정 발급" 다이얼로그는 `login_id`, 임시 비밀번호, `force_change_on_first_login` 토글을 포함한다.
    - 임시 비밀번호와 강제 재설정 비밀번호는 발급 시점에 1회 표시 후 화면에서 사라진다.
    - 회수(`disabled`)는 확인 다이얼로그와 audit reason 입력을 요구한다.

### 6.4 조직 관리 후속

- Organization UI와 read/write API는 main에 반영됨.
- 멤버 변경 외에 리더 변경, 사용자 신규 등록과 계정 발급을 묶는 합성 액션은 백엔드 정책 확정 후 추가 검토한다.

## 7. 다음 작업 큐

### 7.1 M2 1차 완성 sprint (`claude/login_usermanagement_finish`, 진입 중)

- [ ] PR-UX1 — `/admin/settings/users` SearchInput 실 필터링 (placeholder 제거)
- [ ] PR-UX2 — `/account` Kratos privileged session 안내 추가
- [ ] PR-UX3 — Header Switch View 한계 안내 (서버 RBAC 우회 못함)

세부는 [sprint_plan](../ai-workflow/memory/claude/login_usermanagement_finish/sprint_plan.md) 참조. 백엔드 짝(PR-M2-AUDIT) 은 백엔드 로드맵 §6 참조.

### 7.2 완료 (참고)

- [x] 역할별 기본 진입 우선순위 라우팅 (PR #52, `defaultLandingFor`)
- [x] 로그인 직후 역할 기반 기본 진입 + Header Switch View (PR #52)
- [x] `account.service.ts` 신설 (PR #50)
- [x] `/auth/login` 페이지 + 인증 가드 (PR-LOGIN-2, PR #34)
- [x] `/auth/callback` + tokenStore 영속화 (PR-LOGIN-3)
- [x] `/account` 본인 화면 + 비밀번호 변경 폼 (PR #50)
- [x] Header Sign Out → Hydra/Kratos 세션 종료 (PR-LOGIN-4, PR #45·#51)

### 7.3 후속 (별도 sprint)

- [ ] command status transition WebSocket UI (M4, Phase 4 잔여)
- [ ] AI Gardener suggestion API/UI 연결 (M4)
- [ ] Organization 페이지의 계정 관리 action 통합 — 현재는 `/admin/settings/users` 로 분리 운영 (재정합 필요 시 별도 sprint)
