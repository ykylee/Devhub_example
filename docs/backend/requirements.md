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
- **AccountStatus**: `enum { ACTIVE, DISABLED, LOCKED, PASSWORD_RESET_REQUIRED }`

## 5. 사용자 계정 및 인증 (User Account & Authentication)

DevHub 자체 사용자 계정(Account) 도입에 따라 다음 백엔드 API/도메인 작업이 필요하다. 정책 기반은 [요구사항 정의서 2.5](../requirements.md#25-사용자-계정-관리-user-account-management) 와 [architecture.md 6.2](../architecture.md#62-사용자user--계정account-도메인-분리), 계약은 [backend_api_contract.md §11](../backend_api_contract.md#11-계정-및-인증-account--auth) 을 참조한다.

### 5.1 도메인 모델
- `accounts` 테이블: `user_id` UNIQUE FK + `login_id` UNIQUE + 비밀번호 해시 + 상태 + 실패 카운터.
- 사용자 1명 ↔ 계정 1개 invariant 를 DB UNIQUE 제약과 도메인 레이어 동시 검사로 보장.
- 비밀번호 해시 알고리즘: bcrypt cost ≥ 12 또는 argon2id. 알고리즘은 `password_algo` 컬럼에 저장해 회전 가능하도록 함.

### 5.2 API
- `POST /api/v1/accounts` — 시스템 관리자 발급, 1:1 invariant 검사.
- `GET /api/v1/accounts/{user_id}` — 본인 또는 시스템 관리자.
- `PATCH /api/v1/accounts/{user_id}` — login_id/status 변경 (시스템 관리자).
- `PUT /api/v1/accounts/{user_id}/password` — 본인 변경 또는 시스템 관리자 강제 재설정.
- `DELETE /api/v1/accounts/{user_id}` — 회수 (시스템 관리자).
- `POST /api/v1/auth/login`, `POST /api/v1/auth/logout` — 인증 lifecycle.

### 5.3 비기능 요구
- 비밀번호 평문은 어떤 응답/audit/log 에도 포함하지 않는다.
- 로그인 실패 카운터 임계치 초과 시 자동 `locked` 전환. 임계치/잠금 해제 정책은 운영 정책 문서에서 별도 관리.
- 계정 회수 시 활성 세션 즉시 무효화.
- 계정 lifecycle action 은 모두 audit log 기록.

---
*Last Updated: 2026-05-07 (Account/Auth 추가)*
