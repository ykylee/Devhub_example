# DevHub 시스템 Usecase 카탈로그

- 문서 목적: 코드베이스 전체 모듈 기준으로 Usecase를 정의하고, 요구사항(REQ)과 설계(ARCH/API) 사이의 중간 추적 단위를 제공한다.
- 범위: backend-core 주요 모듈(auth/account/org/rbac/gitea ingest/command/audit/realtime/store) + 신규 Project 도메인.
- 대상 독자: 프로젝트 리드, 시스템 관리자, Backend/Frontend 설계 담당, 추적성 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-13
- 관련 문서: [요구사항](../requirements.md), [아키텍처](../architecture.md), [API 계약](../backend_api_contract.md), [ERD 카탈로그](./system_erd.md), [통합 로드맵](../development_roadmap.md)

## 1. 모듈 기준

| 모듈 | 코드 기준 |
| --- | --- |
| Auth/OIDC | `backend-core/internal/auth`, `backend-core/internal/httpapi/auth*.go` |
| Account | `backend-core/internal/httpapi/accounts_admin.go`, `account_password.go` |
| Organization | `backend-core/internal/httpapi/organization.go`, `backend-core/internal/store/users_units.go` |
| RBAC | `backend-core/internal/httpapi/rbac.go`, `permissions.go`, `internal/store/postgres_rbac.go` |
| Gitea Ingest/Snapshot | `backend-core/internal/gitea`, `normalize`, `store/postgres.go` |
| Command | `backend-core/internal/httpapi/commands.go`, `internal/commandworker`, `internal/serviceaction` |
| Audit | `backend-core/internal/httpapi/audit.go`, `internal/store/audit_logs.go` |
| Realtime | `backend-core/internal/httpapi/realtime.go`, `snapshot.go` |
| Application/Project (신규) | `docs/requirements.md` §5.4 기준 (설계/구현 예정) |

## 2. Usecase (모듈별)

### 2.1 Auth/OIDC

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-AUTH-01` | 로그인(code flow) 수행 | 인증 성공 시 actor context 확정 + 토큰 교환 경로 완료 | FR-19, FR-21..24 |
| `UC-AUTH-02` | 로그아웃/세션 종료 | Hydra/Kratos 세션 무효화 + 재진입 시 재인증 | FR-22, FR-23 |
| `UC-AUTH-03` | 인증 실패 처리 | invalid 토큰/권한 부족 시 표준 에러 및 감사 추적 | NFR-3, NFR-18 |

### 2.2 Account

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-ACCOUNT-01` | 관리자 계정 발급/회수 | 사용자-계정 1:1 정책 유지 + 상태 전이 기록 | FR-15..18, FR-61..67 |
| `UC-ACCOUNT-02` | 관리자 비밀번호 재설정 | 강제 재설정 정책 적용 + audit 기록 | FR-20, FR-22 |
| `UC-ACCOUNT-03` | 본인 비밀번호 변경 | current_password 검증 후 변경 완료 | FR-23, FR-26 |

### 2.3 Organization

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-ORG-01` | 사용자 CRUD | 사용자 등록/조회/수정/삭제 일관 처리 | FR-68..72 |
| `UC-ORG-02` | 조직 단위 CRUD | 계층 단위 생성/수정/삭제 + cycle 방지 | FR-73..76 |
| `UC-ORG-03` | 조직 멤버 배정 | unit appointment 정합성 유지 | FR-77..79 |
| `UC-ORG-04` | 조직 계층 조회 | hierarchy + total_count 반환 | FR-80, NFR-21 |

### 2.4 RBAC

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-RBAC-01` | 정책 조회/편집 | 정책 CRUD + 시스템 role 제약 유지 | FR-27, FR-86 |
| `UC-RBAC-02` | 사용자 role 할당 | user-role FK 정합성 보장 | FR-27 |
| `UC-RBAC-03` | 라우트 권한 강제 | deny-by-default + permission cache 반영 | NFR-26 |

### 2.5 Gitea Ingest/Snapshot

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-GITEA-01` | webhook 수집/검증 | dedupe + 상태 추적 + raw payload 보관 | FR-49..55 |
| `UC-GITEA-02` | 도메인 정규화 | repo/issue/pr/ci/risk snapshot 최신화 | FR-49..55 |
| `UC-GITEA-03` | 조회 API 제공 | dashboard/pipeline API 일관 응답 | FR-1..13, FR-28..36 |

### 2.6 Command

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-CMD-01` | 서비스 액션 명령 생성 | command lifecycle 시작 + idempotency 준수 | FR-58, FR-84, FR-95 |
| `UC-CMD-02` | 리스크 대응 명령 생성 | mitigation command 생성/조회 가능 | FR-59, FR-100 |
| `UC-CMD-03` | 명령 실행/상태 전이 | pending→running→종료 상태 일관성 | FR-101 |

### 2.7 Audit

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-AUD-01` | 감사 로그 기록 | actor/request/source context 포함 저장 | FR-18, FR-26, NFR-4 |
| `UC-AUD-02` | 감사 로그 조회 | target/action/time 필터 조회 가능 | FR-102 |

### 2.8 Realtime

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-RT-01` | 실시간 이벤트 구독 | websocket 연결/권한 검증/이벤트 전달 | FR-56, FR-57, FR-60 |
| `UC-RT-02` | 상태 변경 전파 | command/ci/risk 변화가 채널에 반영 | FR-82, FR-83, FR-104, FR-105 |

### 2.9 Application (총괄 단위)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-APP-01` | Application 등록 | Application 생성 + owner/KPI/기간 저장 | REQ-FR-APP-001 |
| `UC-APP-02` | Application-Repository 연결 | 1 Application : N Repository 매핑 저장 | REQ-FR-APP-002 |
| `UC-APP-03` | Application 상태 전이/보관 | status/archived 정책 적용 | REQ-FR-APP-001 |
| `UC-APP-04` | Application 상세 조회 | 메타+연결 repo+상위 마일스톤 조회 | REQ-FR-APP-001,002 |
| `UC-APP-05` | Application 관리 권한 검증 | `system_admin`만 쓰기 허용, `pmo_manager` 비활성 시 403 반환 | REQ-FR-PROJ-000 |
| `UC-APP-06` | Repository 운영 스냅샷 조회 | 작업현황/PR/빌드/품질 지표를 repo 단위로 조회 | REQ-FR-APP-005,006,007,008 |
| `UC-APP-07` | 외부 도구 동기화/재동기화 | webhook+pull 기반 동기화, 중복/누락 보정 처리 | REQ-FR-APP-004, REQ-NFR-PROJ-004 |
| `UC-APP-08` | SCM provider 어댑터 라우팅 | repo_provider 기준으로 적절한 어댑터를 선택해 동일 도메인 계약으로 수집/조회 처리 | REQ-FR-APP-009, REQ-NFR-PROJ-005 |
| `UC-APP-09` | Application 상태 전이 관리 | 상태 머신 규칙에 따라 유효 전이만 허용하고 invalid 전이는 거절 | REQ-FR-APP-010 |
| `UC-APP-10` | Application 롤업 메타 조회 | 롤업 지표와 함께 period/filter/weight/data_gap 메타 제공 | REQ-FR-APP-012, REQ-NFR-PROJ-006 |

### 2.10 Project (Repository 하위 운영 단위)

| UC ID | Usecase | 성공 조건 | 관련 REQ |
| --- | --- | --- | --- |
| `UC-PROJ-01` | Project 등록 | Project 생성 + owner/KPI/기간 저장 | REQ-FR-PROJ-001 |
| `UC-PROJ-02` | Project 멤버/책임자 관리 | 멤버 역할/owner 변경 + 감사 추적 | REQ-FR-PROJ-003 |
| `UC-PROJ-03` | 상/하위 마일스톤 매핑 | child->parent 매핑 저장/롤업 가능 | REQ-FR-PROJ-004 |
| `UC-PROJ-04` | Jira/Confluence 연동 정책 관리 | scope/policy 기준 integration 관리 | REQ-FR-PROJ-005,006 |
| `UC-PROJ-05` | Project 상태 전이/보관 | status/archived 정책 적용 | REQ-FR-PROJ-001,008 |
| `UC-PROJ-06` | 내 Project 조회 | 멤버십 기반 목록 + archived 토글 | REQ-FR-PROJ-002 |
| `UC-PROJ-07` | 공개 Project 조회 | visibility=public 목록 조회 | REQ-FR-PROJ-002 |
| `UC-PROJ-08` | Project 상세 조회 | 메타+마일스톤 롤업 조회 | REQ-FR-PROJ-002,004 |
| `UC-PROJ-09` | Project cadence 리포트 조회 | repo sprint 결과를 상위 롤업 | REQ-FR-PROJ-007 |
| `UC-PROJ-10` | Project 관리 권한 검증 | 관리 쓰기 권한의 RBAC/feature-flag 검증 + 감사 추적 | REQ-FR-PROJ-000,010 |

## 3. 설계/구현 반영 규칙

1. 신규/변경 API는 최소 1개 UC를 참조해야 한다.
2. 신규 마이그레이션은 대응 UC와 연결되어야 한다.
3. 추적성 매트릭스는 REQ→UC→ARCH/API→IMPL→UT/TC 순으로 유지한다.
