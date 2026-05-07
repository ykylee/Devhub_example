# Frontend Development Roadmap (Phase 2 & 3)

- 문서 목적: 백엔드 API 연동 및 프론트엔드 기능 고도화 로드맵을 정의한다.
- 기준일: 2026-05-04
- 최종 수정: 2026-05-07 (Phase 5 계정/로그인 UI 상세화)
- 상태: in_progress
- 관련 문서: `docs/backend_api_contract.md`, `ai-workflow/project/implementation_plan.md`

## 1. 개요

프론트엔드 Phase 1에서 구축된 현대적인 UI와 서비스 레이어 구조를 바탕으로, Phase 2에서는 백엔드 API와의 실시간 데이터 연동을 완료하고 프로덕션 수준의 안정성을 확보하는 것을 목표로 한다.

## 2. Phase 로드맵

| Phase | 상태 | 목표 | 주요 작업 |
| --- | --- | --- | --- |
| **Phase 1** | **done** | 대시보드 UI & Mock 서비스 | 레이아웃 구축, Glassmorphism 적용, Singleton Service 패턴 도입 |
| **Phase 2** | **done** | 핵심 API 통합 | Infra Topology, Risk List, Command/Audit 연동, Role mapping |
| **Phase 3** | **done** | 실시간성 및 CI/CD 가시화 | WebSocket 통합, CI Run/Logs 연동, 실시간 알림 피드 |
| **Phase 4** | **planned** | AI 어드바이저 & 어드민 액션 | AI Gardener 추천 연동, 시스템 관리자 서비스 제어 액션 실체화 |
| **Phase 5** | **planned** | 사용자/조직 관리 + 자체 계정 인증 | 사용자 프로필, 조직 관리 UI(이미 일부 구현됨), 로그인/로그아웃, 계정(Account) 발급·회수·비밀번호 변경, 권한(RBAC) 관리 |

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
    - 비동기 `command_id`를 수신하여 처리 중 상태(pending)를 UI에 표시.

### 3.3 Dashboard Metrics 연동
- **목표**: 역할별 메트릭 카드를 실제 통계 데이터로 업데이트.
- **API**: `GET /api/v1/dashboard/metrics?role={role}`

## 4. 기술적 고려 사항

- **Error Handling**: API 호출 실패 시 Toast 메시지 노출 및 Graceful Degradation(Mock 데이터 또는 이전 캐시 활용) 적용.
- **Loading States**: `Framer Motion`을 활용한 스켈레톤 UI 및 로딩 애니메이션 유지.
- **Environment**: `NEXT_PUBLIC_API_URL` 환경 변수를 통해 백엔드 주소 관리.

## 5. Phase 5 상세 계획 (사용자/조직 관리 + 자체 계정 인증)

DevHub 자체 사용자 계정(Account) 1:1 컨셉 도입 (`docs/requirements.md` 2.5, `docs/architecture.md` 6.2 참조)에 따른 프론트 작업.

### 5.1 로그인 / 인증 흐름
- **목표:** `/login` 진입점 + 인증 가드 + `must_change_password=true` 라우팅.
- **API:** `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`, `GET /api/v1/me` (예정).
- **핵심 로직:**
    - 로그인 폼은 `login_id` + `password` 만 받는다 (이메일/SSO 분리).
    - 응답 `must_change_password=true` 면 즉시 `/account/password` 로 강제 라우팅, 다른 화면 접근 차단.
    - 세션 토큰은 secure cookie 또는 메모리/스토리지 중 하나로 보관 — 백엔드 결정에 맞춰 선택.
    - `401 unauthenticated` / `423 locked` / `403 disabled` 응답에 대해 사용자 안내 문구 분기.

### 5.2 내 계정 화면 (`/account`)
- **목표:** 본인이 자신의 로그인 ID 확인 및 비밀번호 변경.
- **API:** `GET /api/v1/accounts/{user_id}`, `PUT /api/v1/accounts/{user_id}/password` (본인 변경 페이로드).
- **핵심 로직:**
    - login ID 는 읽기 전용 표시 (본인 변경 미지원, 시스템 관리자만 변경).
    - 비밀번호 변경 폼은 `current_password` + `new_password` + `confirm_new_password` 3 필드.
    - 어떤 화면도 비밀번호 평문을 표시/저장하지 않는다.

### 5.3 시스템 관리자 — 계정 관리 패널
- **목표:** 사용자 row 옆에서 계정 발급/회수/잠금 해제/강제 재설정 수행.
- **API:** `POST /api/v1/accounts`, `PATCH /api/v1/accounts/{user_id}`, `PUT /api/v1/accounts/{user_id}/password` (`force=true`), `DELETE /api/v1/accounts/{user_id}`.
- **핵심 로직:**
    - "계정 발급" 다이얼로그 — `login_id` + 임시 비밀번호 + `force_change_on_first_login` 토글. 임시 비밀번호는 발급 시점에 1회 표시 후 화면에서 사라진다 (별도 저장 X, 시스템 관리자가 사용자에게 전달).
    - "강제 재설정" 도 동일하게 1회 표시.
    - 회수(`disabled`)는 확인 다이얼로그 + audit reason 입력.

### 5.4 조직 관리 후속
- 멤버 변경 외에 "리더 변경", "사용자 신규 등록 → 동시에 계정 발급" 같은 합성 액션을 백엔드 정책 확정 후 추가 검토.

## 6. 다음 작업 큐
- [ ] `infra.service.ts` 실제 API 연동
- [ ] `risk.service.ts` 실제 API 및 Command flow 연동
- [ ] WebSocket 클라이언트 초기 스캐폴딩
- [ ] Phase 5 — `account.service.ts` 신설 (Account/Auth API 래핑)
- [ ] Phase 5 — `/login` 페이지 + 인증 가드 미들웨어 / layout
- [ ] Phase 5 — `/account` 본인 화면 + 비밀번호 변경 폼
- [ ] Phase 5 — Organization 페이지에 계정 관리 action 통합 (시스템 관리자 한정)
