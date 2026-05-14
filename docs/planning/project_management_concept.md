# 프로젝트 관리 도메인 컨셉

- 문서 목적: DevHub 의 신규 1차 도메인 **Application/Repository/Project 운영 모델** 의 컨셉 — 무엇을 묶는 entity 인지, 어떤 usecase 가 핵심인지, MVP 경계 — 를 정의해 후속 요구사항/usecase/설계 단계의 공통 기준으로 삼는다.
- 범위: 컨셉 단계. 도메인 정의 / 행위자 × 핵심 usecase 분리 / 운영 계층 모델 / Jira 연계 정책 / 로드맵·마일스톤 롤업 구조 / MVP scope / 데이터 모델 초안 / 템플릿·예시 운영안 진입점 / out-of-scope / 후속 단계 진입 hook. 본 문서는 **REQ-FR / ARCH / API ID 를 아직 발급하지 않는다** — 후속 sprint 에서 발급.
- 대상 독자: 프로젝트 리드, Backend / Frontend / Auth 트랙 담당자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-13
- 관련 문서: [요구사항 정의서 §5.3, §5.4](../requirements.md), [통합 개발 로드맵](../development_roadmap.md), [아키텍처](../architecture.md), [백엔드 API 계약](../backend_api_contract.md), [추적성 컨벤션](../traceability/conventions.md), [planning 진입점](./README.md), [운영 템플릿](./project_operating_model_template.md), [예시 운영안](./project_operating_model_example_2026.md).

## 1. 컨셉 정리 배경

- 기존 [`requirements.md §5.7 프로젝트 및 저장소 관리`](../requirements.md) 는 *Gitea 저장소 자동 생성 + 브랜치 보호 + 멤버 초대* 같은 **Gitea 연동 자동화 시나리오** 만 다루고, DevHub 자체의 **Project entity 자체의 lifecycle / CRUD / 조회 usecase** 는 정의가 없다.
- 다른 모듈 (조직 / 사용자 / RBAC / 인증) 은 §M2~§M3 으로 1차 안정화되었으나, 그 위에 얹히는 **상위 컨테이너 — "어떤 프로젝트에 무엇이 매여 있는가" — 가 비어 있다.**
- 또한 팀의 운영 관성은 **기간성 운영 과제를 명시적으로 발의**하는 방식이고, 실행 레벨은 repo 단위로 분화되어 있다.
- 따라서 본 컨셉은:
  - Application/Repository/Project 가 **어떤 역할을 갖는 계층 entity 인가** 를 정의하고,
  - **두 핵심 usecase** (일반 사용자 조회 / 시스템 관리자 등록·관리) 를 분리하고,
  - **Application > Repository > Project > GitHub 실행 단위**의 운영 계층을 공식화하고,
  - 기존 §5.7 (Gitea 자동화), M4 (실시간), v2 (AI), 향후 RBAC / 알림 등을 **out-of-scope 후속** 으로 분리한다.
- 본 컨셉이 머지되면 후속 sprint 에서 단계적으로 요구사항(REQ-FR) → usecase → 설계(ARCH/API) 로 발전시킨다.

## 2. 운영 계층 모델 (합의안)

### 2.1 기본 구조

- **최상단 Application** 을 명시적으로 정의한다.
- Application 하위에 **N개의 Repository(실행 단위)** 를 둔다.
- 각 Repository 하위에 **기간성 Project(운영 단위)** 를 둔다.
- 저장소부터는 **GitHub 방식**을 따른다.
  - 이슈: Issue
  - 계획/진행판: Project (Board)
  - 일정 기준점: Milestone
  - 기술/운영 문서: Wiki/Docs

요약하면, **총괄 계층은 Application 중심, 실행 계층은 Repository 중심, 운영 계층은 Project 중심**이다.

### 2.2 대형 과제 구조

- 하나의 제품/서비스는 Application으로 선언한다.
- 하위 repo 각각을 실행 단위로 운영하고, repo 아래 Project를 기간 단위로 갱신한다.
- 필요 시 repo를 도메인별(backend/frontend/data/ops)로 나누고 각 repo에 독립 cadence 를 둔다.

### 2.3 역할 배치

- **Chief PL (Application 1명)**: 목표/KPI/예산/리스크 총괄, 우선순위 충돌 조정.
- **Delivery PL (repo별 1명)**: 스프린트 목표/백로그/납기 책임.
- 스프린트 실행 책임은 repo 레벨에 두고, Application 레벨은 롤업·조정에 집중한다.

## 3. Jira/Confluence 연결 정책

### 3.1 Jira 연결 옵션 비교 결론

- 단일 Project Jira only: 리포팅은 단순하지만 실무 티켓 혼재 위험이 큼.
- repo별 Jira only: 실행 정합성은 높으나 상위 집계 규칙이 필요.
- 연간+repo 동시 연결: 설계를 잘하면 가장 유연하지만 중복 입력 리스크 존재.

**본 컨셉의 기본 정책은 하이브리드다.**

1. 실행 이슈의 Source of Truth 는 **repo Jira**.
2. Project 이슈 관리는 기본적으로 Repository 도구(Jira/GitHub Projects)에 연결한다.
3. 상위 Jira 에 작업성 Story/Task 를 직접 생성하지 않는다.
4. 하위 Jira 이슈를 상위 Epic 에 링크하여 자동 집계한다.

### 3.2 Confluence 정책

- **Application Confluence**: 방향성 문서, 의사결정, 분기 계획, 경영/관리 보고.
- **repo Confluence/Docs**: 설계서, runbook, RFC, 장애/회고 등 실행 문서.
- 원칙: 문서도 실행 단위는 repo 우선, 상위는 요약·정책 중심.

## 4. 로드맵/마일스톤 계층 운영

### 4.1 계층 원칙

- 상위(Application): **큰 로드맵/마일스톤** (분기·반기 게이트).
- 하위(repo): **세부 로드맵/마일스톤** (스프린트·월 단위).

### 4.2 필수 규칙

1. 하위 마일스톤은 반드시 상위 마일스톤에 매핑한다 (`child -> parent`).
2. 상위 계층은 작업 티켓을 직접 소유하지 않는다.
3. 상위 완료 기준(DoD)은 수치로 명시한다.
4. 매핑 누락 항목은 리포팅에서 제외가 아닌 "경고"로 표시한다.

### 4.3 완료 기준 예시 (상위)

- 연결된 하위 마일스톤 완료율 90% 이상.
- P1 미해결 0건.
- 통합 테스트/릴리즈 게이트 통과.

## 5. 행위자 × 핵심 usecase 분리

### 5.1 일반 사용자 (Developer / Manager / QA) — 조회 중심

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 내가 멤버인 과제/저장소 묶음 보기 | 자신이 owner/member 인 active/on_hold 과제 및 연결 repo 요약 | ✅ |
| 공개(public) 과제 보기 | visibility=`public` 과제 목록 | ✅ |
| 과제 상세 보기 | 메타 + 멤버 + 상태 + 연결 repo + 상/하위 마일스톤 요약 | ✅ |
| 진행 현황/위험 보기 | 진행률, 최근 활동, blocker | 후속 (M4 연계, AI 보강은 v2) |
| 검색/필터 | status, owner, 부서, repo | ✅ (기본 필터) |

### 5.2 시스템 관리자 (System Admin) — 등록·관리 전용

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 전체 과제 목록 보기 | 보관(archived) 포함 옵션 | ✅ |
| Application 신규 등록 | name / description / key / owner / visibility / 기간 / KPI | ✅ |
| 과제 메타 수정 | 위 항목 + status 전환 | ✅ |
| 하위 repo 연결/해제 | 1:N repo 매핑 | ✅ |
| 멤버/책임자 관리 | owner 교체, 멤버 role 변경 | ✅ |
| 과제 보관 (archive) | soft-delete 성격 | ✅ |
| 영구 삭제 | 보관 후 정책 기반 삭제 | 2차 |
| Jira/Confluence 연결 정책 관리 | 상위/하위 연결 규칙 검증 | ✅ |

### 5.3 책임자(Owner) 권한 위양 범위

- MVP 1차: 시스템 관리자가 모든 쓰기 작업의 단일 주체.
- 2차: owner 에게 제한적 메타/멤버 수정 위양.
- 3차: RBAC row-scoping (ADR-0011 후보) 채택 시 부서 단위 위양 확장.

## 6. 스프린트/애자일 운영 모델

1. 스프린트는 **repo 단위 독립 운영**이 기본.
2. Application 레벨은 **주간 Program Sync + 월간 KPI 리뷰** cadence 로 동기화.
3. 공통 게이트(Alpha/Beta/GA 등)는 상위 마일스톤으로만 강제.
4. Jira 티켓 생성은 repo 레벨만 허용(상위는 롤업 전용).

## 7. 데이터 모델 초안 (운영 계층 반영)

```text
applications
  id                 UUID PK
  key                TEXT UNIQUE
  name               TEXT NOT NULL
  description        TEXT
  status             TEXT NOT NULL   -- planning | active | on_hold | closed | archived
  visibility         TEXT NOT NULL   -- public | internal | restricted
  owner_user_id      UUID FK users.id
  start_date         DATE NULL
  due_date           DATE NULL
  created_at         TIMESTAMPTZ NOT NULL
  updated_at         TIMESTAMPTZ NOT NULL
  archived_at        TIMESTAMPTZ NULL

application_repositories
  application_id          UUID FK applications.id
  repo_provider           TEXT NOT NULL   -- bitbucket | gitea | forgejo | github | etc
  repo_full_name          TEXT NOT NULL   -- org/repo
  external_repo_id        TEXT NULL        -- provider 내부 식별자 (있는 경우)
  role                    TEXT NOT NULL   -- primary | sub | shared
  sync_status             TEXT NOT NULL   -- requested | verifying | active | degraded | disconnected
  sync_error_code         TEXT NULL        -- §13.3 표준 사전 (link 단위 최신 1건 캐시)
  sync_error_retryable    BOOLEAN NULL
  sync_error_at           TIMESTAMPTZ NULL
  last_sync_at            TIMESTAMPTZ NULL
  linked_at               TIMESTAMPTZ NOT NULL
  PRIMARY KEY (application_id, repo_provider, repo_full_name)
  -- 운영 룰: 동일 link 의 sync_error_code 는 최신 1건만 캐시. event 단위 상세 에러는
  -- `webhook_events` (현행) 또는 후속 `adapter_event_logs` 에 별도 보관 (§13.3 참조).

projects
  id                 UUID PK
  application_id     UUID FK applications.id
  repository_id      BIGINT FK repositories.id
  key                TEXT NOT NULL
  name               TEXT NOT NULL
  status             TEXT NOT NULL
  owner_user_id      UUID FK users.id
  start_date         DATE NULL
  due_date           DATE NULL
  UNIQUE (repository_id, key)

pr_activities
  id                 BIGSERIAL PK
  repository_id      BIGINT FK repositories.id
  external_pr_id     TEXT NOT NULL
  event_type         TEXT NOT NULL
  actor_login        TEXT NOT NULL
  occurred_at        TIMESTAMPTZ NOT NULL
  payload            JSONB NOT NULL

build_runs
  id                 BIGSERIAL PK
  repository_id      BIGINT FK repositories.id
  run_external_id    TEXT UNIQUE NOT NULL
  branch             TEXT NOT NULL
  commit_sha         TEXT NOT NULL
  status             TEXT NOT NULL
  duration_seconds   INT NULL
  started_at         TIMESTAMPTZ NOT NULL
  finished_at        TIMESTAMPTZ NULL

quality_snapshots
  id                 BIGSERIAL PK
  repository_id      BIGINT FK repositories.id
  tool               TEXT NOT NULL
  ref_name           TEXT NOT NULL
  commit_sha         TEXT NULL
  score              NUMERIC NULL
  gate_passed        BOOLEAN NULL
  measured_at        TIMESTAMPTZ NOT NULL

project_integrations
  -- 참고: 본 컨셉 단계에서는 Application/Project 어느 레벨에서도 integration 을
  -- 등록할 수 있도록 단일 테이블에 `scope` 컬럼으로 구분한다.
  -- ERD §2.5 의 `PROJECT_INTEGRATIONS` 와 동일 entity (단일 명칭 채택).
  id                 UUID PK
  project_id         UUID FK projects.id NULL       -- scope=project 인 경우 필수
  application_id     UUID FK applications.id NULL    -- scope=application 인 경우 필수
  scope              TEXT NOT NULL   -- application | project
  integration_type   TEXT NOT NULL   -- jira | confluence
  external_key       TEXT NOT NULL
  url                TEXT NOT NULL
  policy             TEXT NOT NULL   -- summary_only | execution_system
```

## 8. 템플릿 / 예시 운영안

- 운영 템플릿: [`project_operating_model_template.md`](./project_operating_model_template.md)
- 예시 운영안: [`project_operating_model_example_2026.md`](./project_operating_model_example_2026.md)

## 9. 명시적 out-of-scope (별도 sprint / 마일스톤)

| 항목 | 분류 | 후속 진입 |
| --- | --- | --- |
| Gitea 저장소 자동 생성 / 브랜치 보호 / 멤버 자동 초대 | §5.7 자동화 시나리오 | 별도 sprint, §5.7 source-of-truth 유지 |
| 실시간 진행 상태 (WebSocket) | M4 이벤트 확장 | M4 |
| AI Suggestion / 위험 자동 탐지 | AI 가드너 | v2 |
| MFA / 위험 작업(영구 삭제) 다단계 확인 | 운영 보안 | 운영 진입 직전 |
| CSV 일괄 가져오기 / 내보내기 | 운영 효율화 | 운영 진입 후 |

## 10. 미해결 / 결정 대기

| 항목 | 결정 후보 | 결정 시점 |
| --- | --- | --- |
| ~~Application `key` 형식 (현재 10자 영문숫자, 정책 변경 가능)~~ | **closed** — REQ-FR-APP-003 확정 (immutable, `^[A-Za-z0-9]{10}$`) | Req sprint (closed 2026-05-14) |
| 영구 삭제 정책 | 보관 후 N일 + admin 재확인 | Design sprint |
| Owner 위양 범위 | 2차 sprint | 후속 |
| RBAC row-level enforcement | ADR-0011 후보 (placeholder draft 발급) | Design sprint |
| 다중 SCM provider 운영 시 표준 provider 정책 | provider 동등 지원 + 표준 어댑터 확장 | Integration sprint |
| 비-Jira SCM (bitbucket/gitea/forgejo) 의 Jira 매핑 | REQ-FR-PROJ-005 하이브리드 정책의 cross-cut. 외부 issue tracker scope 결정 | Integration sprint |

## 11. 후속 단계 진입 hook

| 단계 | 산출물 후보 | 진입 조건 |
| --- | --- | --- |
| Req sprint | `docs/requirements.md` 확장 (`Project/하위 repo/Jira 정책`) + REQ-FR 발급 | 본 컨셉 머지 직후 |
| Usecase sprint | 행위자 × usecase 매트릭스, Jira/Confluence 운영 규칙 상세 | Req sprint 머지 직후 |
| Design sprint (backend) | `/api/v1/applications/*`, `/api/v1/repositories/:repo_id/projects/*` 계약 + 마이그레이션 초안 | Usecase sprint 머지 직후 |
| Design sprint (frontend) | `/admin/settings/applications` IA/화면 흐름 | backend design 과 병행 |
| Implementation sprint | IMPL-application-*, UT-application-*, TC-application-* | design 머지 후 |

## 12. SCM Adapter 설계 결정표

### 12.1 어댑터 경계 (Core vs Adapter)

| 구분 | Core 책임 | Adapter 책임 |
| --- | --- | --- |
| 도메인 계약 | Repository/PR/Build/Quality/Event 공통 스키마 유지 | provider payload를 공통 스키마로 변환 |
| 보안/권한 | RBAC, 감사로그, API 권한 검증 | provider 인증 방식, webhook 서명 검증 |
| 운영 정책 | 롤업, 경고, 재동기화 정책 결정 | provider rate limit/재시도/백오프 처리 |
| 오류 처리 | 표준 에러 코드/운영 노출 정책 | provider별 에러 포맷 흡수 및 표준 코드 매핑 |

### 12.2 데이터 모델 정밀화

| 항목 | 결정 |
| --- | --- |
| provider 카탈로그 | `scm_providers(provider_key, display_name, enabled, adapter_version, created_at, updated_at)` 관리 |
| Repository 연결 | `application_repositories`는 `repo_provider + repo_full_name` 기반 연결 관리 |
| 외부 식별자 | provider별 `external_repo_id`/`external_project_key`는 설계 단계에서 컬럼 확장 검토 |
| 동기화 상태 | `last_sync_at`, `sync_status`, `sync_error_code`를 연결/수집 엔티티에 단계적 도입 |

### 12.3 운영 시나리오

| 시나리오 | 설계 원칙 |
| --- | --- |
| 단일 provider 운영 | 표준 ingest/pull 파이프라인 적용 |
| 다중 provider 병행 | provider별 어댑터 독립 실행 + 공통 도메인 롤업 |
| provider 전환(cutover) | 기존 provider read-only 전환 -> 신규 provider 병행 검증 -> 기준 provider 전환 |
| 장애 상황 | provider 장애는 해당 파이프라인만 degraded, 전체 수집은 계속 |
| 재동기화 | provider별 reconciliation 작업을 분리 스케줄로 운영 |

### 12.4 API 계약 정밀화

| 항목 | 결정 |
| --- | --- |
| Provider 카탈로그 API | `GET /api/v1/scm/providers`, `PATCH /api/v1/scm/providers/{provider_key}` |
| Repository 연결 검증 | `repo_provider` 유효성 + provider별 repo 참조 검증 |
| 표준 에러 코드 | `unsupported_repo_provider`, `provider_unreachable`, `webhook_signature_invalid`, `repository_link_conflict`, `invalid_repository_reference` |
| 계약 안정성 | 신규 provider 추가 시 기존 화면/API 응답 shape를 유지 |

## 13. Application 상세 설계

### 13.1 엔티티 필드 계약

| 필드 | 타입/제약 | 설명 |
| --- | --- | --- |
| `id` | UUID PK | 내부 식별자 |
| `key` | 전역 unique, 앱 검증(현재 10자 영문숫자) | 관리 안정 식별자 |
| `name` | NOT NULL | 표시명 |
| `description` | NULL 허용 | 개요 |
| `owner_user_id` | FK users | 총괄 책임자 |
| `visibility` | `public|internal|restricted` | 가시성 정책 |
| `status` | `planning|active|on_hold|closed|archived` | 운영 상태 |
| `start_date`,`due_date` | NULL 허용 | 운영 기간 |
| `created_at`,`updated_at`,`archived_at` | timestamp | 생성/수정/보관 시각 |

### 13.2 상태 전이 규칙 (State Machine)

| 현재 상태 | 허용 전이 | 비고 |
| --- | --- | --- |
| `planning` | `active`, `on_hold`, `archived` | 초기 준비 단계 |
| `active` | `on_hold`, `closed`, `archived` | 정상 운영 단계 |
| `on_hold` | `active`, `closed`, `archived` | 일시 중지 단계 |
| `closed` | `archived` | 종료 후 보관만 허용 |
| `archived` | (기본 불가) | 복구는 후속 정책으로 별도 승인 절차 필요 |

상태 전이 가드:
- `active -> closed`: 연결 Repository가 1개 이상이고 주요 롤업 경고가 치명 상태가 아닌 경우만 허용(세부 임계치는 운영 정책으로 분리).
- `* -> archived`: soft-delete로 처리하며 조회 기본 목록에서 제외.

### 13.2.1 상태 전이별 권한/검증 가드

| 전이 | 권한 | 검증 가드 | 실패 코드 |
| --- | --- | --- | --- |
| `planning -> active` | `system_admin` | 연결된 `sync_status=active` Repository 1개 이상 | `application_activation_precondition_failed` |
| `active -> on_hold` | `system_admin` | `hold_reason` 필수 | `invalid_status_transition_payload` |
| `on_hold -> active` | `system_admin` | `resume_reason` 필수, `due_date` 경과 시 사유 길이 최소 정책 적용 | `invalid_status_transition_payload` |
| `active -> closed` | `system_admin` | 롤업 `critical` 0건 + 활성 Repository 1개 이상 | `application_close_precondition_failed` |
| `closed -> archived` | `system_admin` | `archived_reason` 필수 | `invalid_status_transition_payload` |
| `* -> archived` | `system_admin` | soft-delete + 감사로그 기록 | `invalid_status_transition` (불허 전이 시) |

운영 메모:
- `pmo_manager` 활성화 이전에는 모든 전이 요청이 `403 role_not_enabled`다.
- 가드 임계치(예: critical 기준)는 운영 정책 테이블로 외부화 가능해야 한다.

### 13.3 Application-Repository 연결 라이프사이클

| 단계 | 설명 | 상태 필드 | "활성" 분류 (가드용) |
| --- | --- | --- | --- |
| `requested` | 연결 요청 생성 | `sync_status=requested` | 아님 |
| `verifying` | provider 어댑터가 repo 접근/권한/참조 검증 | `sync_status=verifying` | 아님 |
| `active` | 연결 활성 + 수집 대상 등록 | `sync_status=active` | **✅ 활성** |
| `degraded` | 부분 실패(일시 장애/제한) | `sync_status=degraded`, `sync_error_code` 기록 | 아님 (1차 정책) |
| `disconnected` | 명시적 해제 또는 provider 비활성 | `sync_status=disconnected` | 아님 |

**활성 정의 (§13.2.1 가드 참조용)**: `planning → active` 전이 가드의 "활성 Repository 1개 이상" 검증은 **`sync_status='active'` 만** 활성으로 카운트한다. `degraded` 는 link 자체는 살아있으나 부분 장애 상태이므로 1차 정책에서 활성에서 제외 — 후속 sprint 에서 운영 정책 재평가 후 변경 가능 (예: degraded 의 retryable=true 인 경우 grace period 내 활성 포함).

운영 규칙:
- provider 전환 시 기존 연결을 즉시 삭제하지 않고 `degraded/read-only` 관찰 단계를 거친다.
- 동일 `application_id + repo_provider + repo_full_name` 중복 연결은 금지한다.

`sync_error_code` 표준 사전:

| code | 의미 | retryable | 운영 액션 |
| --- | --- | --- | --- |
| `provider_unreachable` | provider endpoint 도달 실패 | true | 지수 백오프 재시도, 임계 초과 시 알림 |
| `auth_invalid` | 토큰/자격 증명 오류 | false | 자격 갱신 필요, 관리자 조치 |
| `permission_denied` | provider 권한 부족 | false | 권한 스코프 재설정 |
| `rate_limited` | provider 요청 한도 초과 | true | reset window 이후 재시도 |
| `webhook_signature_invalid` | webhook 서명 검증 실패 | false | 시크릿/서명 설정 점검 |
| `payload_schema_mismatch` | payload 구조 불일치 | false | 어댑터 스키마 업데이트 |
| `resource_not_found` | 대상 repo/project 미존재 | false | 연결 정보 정합성 점검 |
| `internal_adapter_error` | 어댑터 내부 처리 오류 | true | 어댑터 로그 점검 후 재시도 |

### 13.4 Application 롤업 계산 규칙

| 지표 | 계산 기준 | 누락 데이터 처리 |
| --- | --- | --- |
| PR 분포 | 연결 repo들의 `open/draft/merged/closed` 합산 | provider 미수집 repo는 `data_gap`으로 표시 |
| Build 성공률 | 기간 내 성공 run / 전체 run | run 0건은 `N/A` 처리 |
| 평균 Build 시간 | 성공 run 기준 평균 | 이상치(운영 정책 기준) 제외 옵션 |
| Quality 점수 | repo별 최신 score의 가중 평균(기본 동일 가중) | score 미수집 repo는 경고 |
| Gate 실패 건수 | gate_passed=false 집계 | provider 장애 구간은 별도 주석 |

가중치 정책 결정:
- 기본: `equal` (모든 연결 Repository 동일 가중치).
- 예외 1: `repo_role` (`primary=0.6`, `sub=0.3`, `shared=0.1` 기본안, 조직 정책으로 조정 가능).
- 예외 2: `custom` (Application 단위 관리자 정의, 합계 1.0 ±0.001 허용오차 제약).
- `custom`에서 특정 repo 가중치 미정의 시 `equal` fallback을 적용하고 `fallbacks` 메타에 `reason="custom_weight_missing"` 으로 이유를 남긴다.

Normalize 규칙 (다중 repo / 카테고리 edge case):
- `equal`: N개 repo 면 각 `1/N`. N=0 이면 가중치 맵 빈 객체 + 전체 결과 `data_gap` 표시.
- `repo_role`: 동일 카테고리 다중 repo 가 있으면 카테고리 가중치를 그 카테고리 내 균등 분할 (예: primary 가 2개면 각 0.3). 카테고리가 0개면 그 가중치를 나머지 카테고리에 비례 재분배 후 합 1.0 으로 정규화.
- `custom`: 합 1.0 검증 통과 후 적용. 누락 repo 는 위 fallback. 음수 가중치는 `422 invalid_weight_policy`.

상태 전이 가드와 마찬가지로 본 §13.4 가 롤업 정책의 SoT 이고 `backend_api_contract.md` §13.6 의 표현은 이 SoT 의 요약이다.

### 13.5 Application 설계 오픈 이슈

| 항목 | 현재안 | 후속 |
| --- | --- | --- |
| `archived -> active` 복구 | 기본 비허용 | 운영 승인 워크플로우 도입 시 재논의 |
| 롤업 가중치 | `equal` 기본 + `repo_role/custom` 예외 정책 | 운영 실측 기반 가중치 튜닝 |
| 종료 조건 자동화 | 수동 close | KPI/리스크 기준 자동 추천은 v2 검토 |

## 14. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-13 | 초안 — 컨셉 1차. 도메인 정의 + 핵심 usecase 분리 + MVP scope + 데이터 모델 초안 + RBAC 영향(보류) + 후속 단계 진입 hook. | sprint `claude/work_260513-p`. |
| 2026-05-13 | 리뷰 1차 보강 — archived 가시성/영구삭제 단계/Owner 위양 단계/`code` 제약 미해결 정합화. | 동일 sprint. |
| 2026-05-13 | 컨셉 심화 — Application > Repository > Project 운영 계층, Jira/Confluence 하이브리드 정책, 상/하위 로드맵·마일스톤 롤업 규칙, PL·스프린트 운영 모델, 템플릿/예시 문서 링크 추가. | 사용자 워크숍 합의 반영. |
| 2026-05-13 | 설계 고도화 — SCM 어댑터 결정표 + Application 상태전이/연결 라이프사이클/롤업 계산 규칙 추가. | 현재 세션 반영. |
| 2026-05-14 | 리뷰 보강 — §7 데이터 모델 PK 정합 (application_id+repo_provider+repo_full_name) + 신규 컬럼 동기 + `project_integrations` 명 통일. §10 `key 형식` row close, 비-Jira SCM Jira 매핑 row 신규. §13.4 weight normalize 룰 + 허용오차 ±0.001 명문화. ADR-0011 placeholder 발급. | PR #104 본인 리뷰 보강. |
