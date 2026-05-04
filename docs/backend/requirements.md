# 백엔드 개발 요구사항 명세 (Backend Requirements)

이 문서는 프론트엔드 개발 중 백엔드에 구현이 필요하다고 판단된 API 및 데이터 구조 요구사항을 정리합니다.

> 상세 리뷰와 구현 계약 재정의 제안은 `docs/backend/requirements_review.md`, 현재 프론트 구현 기반 연동 요구사항은 `docs/backend/frontend_integration_requirements.md`를 기준으로 확인합니다.

## 1. gRPC 실시간 데이터 스트리밍
프론트엔드 대시보드의 실시간성 확보를 위해 다음 스트리밍 API가 필요합니다.

### A. 인프라 토폴로지 & 상태 (System Admin)
- **API**: `stream InfrastructureEvents(Empty) returns (stream InfraState)`
- **필요 데이터**:
    - `NodeState`: 서비스 상태(Health), 리소스 사용량(CPU/Mem), 활성 인스턴스 수
    - `EdgeState`: 서비스 간 레이턴시(ms), 처리량(RPS)
- **빈도**: 1~2초 간격 업데이트

### B. 빌드 파이프라인 로그 (Developer)
- **API**: `stream BuildLogs(BuildRequest) returns (stream LogLine)`
- **필요 데이터**: `timestamp`, `level`, `message`, `step_name`

## 2. 관리자 제어 명령 (Admin/Manager Actions)
프론트엔드에서 트리거된 조치 사항을 백엔드에서 실행하기 위한 API입니다.

### A. 서비스 제어 (Admin)
- **API**: `ControlService(ServiceActionRequest) returns (ActionResponse)`
- **Actions**: `RESTART`, `PAUSE`, `STOP`, `CLEAR_CACHE`
- **필요 데이터**: `service_id`, `action_type`, `force_flag`

### B. 리스크 완화 조치 (Manager)
- **API**: `ApplyRiskMitigation(MitigationRequest) returns (ActionResponse)`
- **Actions**:
    - `ROLLBACK`: 특정 SHA로 배포 롤백
    - `SCALE`: 특정 리전의 리소스 증설
    - `SCHEDULE_ADJUST`: 마일스톤 일정 자동 조정 및 공지
- **필요 데이터**: `risk_id`, `plan_type`, `comment`

## 3. 리스크 탐지 및 푸시 알림 (Intelligent Analytics)
- **목적**: 대시보드 미접속 시에도 브라우저 토스트 알림으로 긴급 리스크 전파
- **API**: `stream CriticalAlerts(UserIdentity) returns (stream AlertEvent)`
- **필요 데이터**: `severity`, `title`, `description`, `suggested_action_id`

## 4. 데이터 스키마 상세
- **UserRole**: `enum { DEVELOPER, MANAGER, ADMIN }`
- **ImpactLevel**: `enum { LOW, MEDIUM, HIGH, CRITICAL }`
- **Status**: `enum { STABLE, WARNING, DEGRADED, DOWN }`

---
*Last Updated: 2026-05-02 (UI 고도화 반영)*
