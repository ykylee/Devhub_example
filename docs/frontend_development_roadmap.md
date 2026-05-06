# Frontend Development Roadmap (Phase 2 & 3)

- 문서 목적: 백엔드 API 연동 및 프론트엔드 기능 고도화 로드맵을 정의한다.
- 기준일: 2026-05-04
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
| **Phase 5** | **planned** | 사용자 및 조직 관리 | 사용자 프로필, 팀/조직 관리 UI, 권한 설정(RBAC) 관리 인터페이스 |

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

## 5. 다음 작업 큐
- [ ] `infra.service.ts` 실제 API 연동
- [ ] `risk.service.ts` 실제 API 및 Command flow 연동
- [ ] WebSocket 클라이언트 초기 스캐폴딩
