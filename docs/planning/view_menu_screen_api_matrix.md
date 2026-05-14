# DevHub 뷰별 메뉴/화면/API 매트릭스

- 문서 목적: 역할별 기본 진입 우선순위 정책에 따라 DevHub 화면 구성과 API 연동 목록을 단일 매트릭스로 정리한다.
- 범위: Developer/Manager/System Admin 뷰, 공통 화면, Application > Repository > Project 운영 모델, M4/v2 범위 구분.
- 대상 독자: 기획, Backend/Frontend 개발자, QA, 운영 담당자.
- 상태: draft
- 최종 수정일: 2026-05-13
- 관련 문서: [요구사항](../requirements.md), [아키텍처](../architecture.md), [백엔드 API 계약](../backend_api_contract.md), [프론트 연동 요구사항](../backend/frontend_integration_requirements.md), [통합 로드맵](../development_roadmap.md), [Project 운영 컨셉](./project_management_concept.md)

## 1. 기본 원칙

- 기본 진입 우선순위:
  - `developer` -> 개발 대시보드
  - `manager` -> 관리 대시보드
  - `system_admin` -> 시스템 대시보드 + 시스템 설정
- 관리 권한 정책:
  - `system_admin`: Application/Repository/Project 관리 쓰기 전 권한 허용
  - `pmo_manager`(tentative): 후보 role. 정책 확정 전 `disabled` 유지
- 노출 정책:
  - 시스템 대시보드/시스템 설정은 `system_admin`만 접근 가능
- 운영 계층:
  - `Application > Repository > Project` (Project는 Repository 하위 기간성 운영 단위)
- AI 가드너:
  - 본 문서의 MVP/M4 화면에서 제외, `v2`로 이관

## 2. 전역 메뉴 IA (권장)

| 메뉴 그룹 | Developer | Manager | System Admin | 비고 |
| --- | --- | --- | --- | --- |
| Dashboard | 개발 대시보드 | 관리 대시보드 | 시스템 대시보드 | 역할별 기본 진입 |
| Applications | 조회 | 조회/필터 강화 | 등록/수정/보관 | Application 총괄 |
| Repositories | 내 repo 중심 | 전체 repo 롤업 | 연동/정책 관리 | 실행 단위 |
| Projects | 내 Project | 팀 Project | Project 생성/보관/멤버 | repo 하위 기간성 운영 |
| Milestones | 조회 | 롤업/경고 | 매핑 관리 | 상/하위 매핑 |
| Integrations | 조회(읽기) | 상태 조회 | Jira/Confluence 설정 | 정책 검증 포함 |
| Organization | 읽기 | 읽기 | CRUD/멤버 관리 | RBAC 적용 |
| Audit | 제한 조회 | 제한 조회 | 전체 조회 | 보안/운영 감사 |
| Settings | 개인 설정 | 개인 설정 | 시스템 설정 | system_admin 전용 항목 포함 |

`pmo_manager` 활성화 전:
- Manager 화면에서 관리 쓰기 액션 버튼은 노출하지 않는다.
- API 응답은 `403 role_not_enabled`를 반환한다.

## 3. 뷰별 화면 구성

### 3.1 Developer 뷰

| 화면 | 핵심 위젯 | 주요 액션 | API |
| --- | --- | --- | --- |
| 개발 대시보드 | 내 작업 스트림, CI 요약, 위험 알림, 내 Project 카드 | 이슈/PR 이동, 리스크 상세 조회 | `GET /api/v1/dashboard/metrics`, `GET /api/v1/issues`, `GET /api/v1/pull-requests`, `GET /api/v1/ci-runs`, `GET /api/v1/risks`, `GET /api/v1/me` |
| Repository 상세 | 파이프라인 상태, 최근 PR/이슈, 빌드 로그 | 로그 조회, 필터링 | `GET /api/v1/repositories`, `GET /api/v1/ci-runs`, `GET /api/v1/ci-runs/{ci_run_id}/logs`, `GET /api/v1/issues`, `GET /api/v1/pull-requests` |
| Repository 운영 지표 | 작업량, PR activity, 빌드 성공률, 품질 점수 추이 | 기간별 비교, 상세 drill-down | `GET /api/v1/repositories/{id}/activity` (planned), `GET /api/v1/repositories/{id}/pull-requests` (planned), `GET /api/v1/repositories/{id}/build-runs` (planned), `GET /api/v1/repositories/{id}/quality-snapshots` (planned) |
| Project 상세(읽기) | 기간/목표/상태, 멤버, 마일스톤 매핑 요약 | 상태/문서 조회 | `GET /api/v1/projects/{project_id}` (planned), `GET /api/v1/projects/{project_id}/milestones` (planned) |
| 알림 센터 | Info/Action Required 알림 | 읽음 처리, 상세 이동 | `GET /api/v1/realtime/ws`, `GET /api/v1/events` |

### 3.2 Manager 뷰

| 화면 | 핵심 위젯 | 주요 액션 | API |
| --- | --- | --- | --- |
| 관리 대시보드 | KPI 카드, 크리티컬 리스크, 진행률/지연, 의사결정 로그 | 리스크 대응 트리거, 롤업 점검 | `GET /api/v1/dashboard/metrics`, `GET /api/v1/risks/critical`, `GET /api/v1/risks`, `GET /api/v1/audit-logs` |
| Application 상세 | KPI/리스크 롤업, 연결 repo, 하위 Project 현황 | 필터/기간 변경, 보고 스냅샷 확인 | `GET /api/v1/applications/{application_id}` (planned), `GET /api/v1/repositories`, `GET /api/v1/risks` |
| Application 롤업 지표 | Repository별 PR/빌드/품질 집계 | 리스크/품질 저하 repo 탐지 | `GET /api/v1/applications/{application_id}/rollup` (planned) |
| 마일스톤 매핑 보드 | 상위-하위 매핑, 누락/충돌 경고 | 매핑 누락 점검 | `GET /api/v1/applications/{application_id}/milestones` (planned), `GET /api/v1/projects/{project_id}/milestones` (planned) |
| 리스크 대응 | 리스크 상세, 대응 명령 상태 | 완화 명령 요청, 진행 추적 | `POST /api/v1/risks/{risk_id}/mitigations`, `GET /api/v1/commands/{command_id}`, `GET /api/v1/realtime/ws` |

PMO 후보 role 정책:
- `pmo_manager` 비활성 단계: Manager와 동일하게 조회 중심.
- `pmo_manager` 활성 단계(후속): `Project 운영 관리`와 `마일스톤 매핑 관리`의 제한적 쓰기 권한 허용 검토.

### 3.3 System Admin 뷰

| 화면 | 핵심 위젯 | 주요 액션 | API |
| --- | --- | --- | --- |
| 시스템 대시보드 | 인프라 토폴로지, 노드 상태, 서비스 액션 | 서비스 제어 명령 실행 | `GET /api/v1/infra/topology`, `GET /api/v1/infra/nodes`, `GET /api/v1/infra/edges`, `POST /api/v1/admin/service-actions`, `GET /api/v1/commands/{command_id}` |
| Application 관리 | Application 목록/상세, 상태/가시성/기간 | 생성/수정/보관 | `GET /api/v1/applications` (planned), `POST /api/v1/applications` (planned), `PATCH /api/v1/applications/{application_id}` (planned), `DELETE /api/v1/applications/{application_id}` (planned, archive) |
| Repository 연결 관리 | Application-Repository 매핑, 역할(primary/sub/shared) | 연결/해제/역할 변경 | `GET /api/v1/applications/{application_id}/repositories` (planned), `POST /api/v1/applications/{application_id}/repositories` (planned), `DELETE /api/v1/applications/{application_id}/repositories/{repo_key}` (planned) |
| Project 운영 관리 | repo 하위 Project 목록/상세, owner/멤버/기간 | 생성/수정/보관, 멤버 변경 | `GET /api/v1/repositories/{repo_id}/projects` (planned), `POST /api/v1/repositories/{repo_id}/projects` (planned), `PATCH /api/v1/projects/{project_id}` (planned), `DELETE /api/v1/projects/{project_id}` (planned, archive) |
| 통합 설정 | Jira/Confluence 연결 상태/정책 | 연결 등록/검증/해제 | `GET /api/v1/integrations` (planned), `POST /api/v1/integrations` (planned), `PATCH /api/v1/integrations/{integration_id}` (planned), `DELETE /api/v1/integrations/{integration_id}` (planned) |
| 계정/조직/RBAC | 사용자/조직 CRUD, 권한 매트릭스, 역할 할당 | 계정 발급/회수, 조직 편집, RBAC 편집 | `GET/POST/PATCH/DELETE /api/v1/users`, `GET /api/v1/organization/hierarchy`, `POST/PATCH/DELETE /api/v1/organization/units`, `PUT /api/v1/organization/units/{unit_id}/members`, `GET/PUT/POST/DELETE /api/v1/rbac/policies`, `GET/PUT /api/v1/rbac/subjects/{subject_id}/roles`, `POST/PUT/PATCH/DELETE /api/v1/accounts*` |
| 감사로그 | 보안/운영/계정 변경 이력 | 필터 조회, 추적 | `GET /api/v1/audit-logs` |

권한 스코프(요구사항 연계):
- `application.manage`
- `application.repo.link`
- `project.manage`
- `project.member.manage`
- `integration.manage`
- `milestone.mapping.manage`

## 4. 공통 화면 및 API

| 공통 화면 | 설명 | API |
| --- | --- | --- |
| 로그인/콜백 | OIDC 로그인, 토큰 교환 | `POST /api/v1/auth/login`, `POST /api/v1/auth/token`, `POST /api/v1/auth/logout`, `GET /api/v1/auth/consent` |
| 내 계정 | 내 프로필/비밀번호 변경 | `GET /api/v1/me`, `POST /api/v1/account/password` |
| 실시간 채널 | 명령 상태/이벤트 반영 | `GET /api/v1/realtime/ws` |

## 5. M4 vs v2 범위 분리

| 항목 | M4 | v2 |
| --- | --- | --- |
| WebSocket replay/필터링 | 포함 | - |
| command status 실시간 UI | 포함 | - |
| 시스템 대시보드 고도화 | 포함 | - |
| AI 가드너 추천/Suggestion Feed | - | 포함 |
| AI 기반 알림 중재 | - | 포함 |
| AI 분석 gRPC (`AnalysisService`) | - | 포함 |

## 6. 설계/구현 반영 규칙

1. 신규 화면 PR은 본 문서의 화면 ID(섹션 기준)와 API 매핑을 PR 본문에 명시한다.
2. planned API가 구현으로 전환되면 `docs/backend_api_contract.md`에 본문 스펙을 먼저 추가한다.
3. 추적성 매트릭스는 `REQ -> UC -> ARCH/API -> IMPL -> UT/TC` 체인으로 동기화한다.
