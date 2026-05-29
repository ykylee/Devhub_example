# application-lifecycle 도메인 API

- 문서 목적: Application/Project 관리 + Application 개발 대시보드(APPDASH) + SCM Provider Catalog API 계약을 정의한다.
- 범위: API-41..50, 55, 56, 56A, 56B, 57, 58, 93. Repository 운영 지표 조회(API-51..54)와 Draft→Publish(API-91/92) 는 `docs/domain/repository-integration/api.md` 참조. 외부 SCM provider 원격 import/create(API-88/89/90)는 `docs/domain/integration-registry/api.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master §13 본문 이관)
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [architecture.md](./architecture.md), [master API](../../backend_api_contract.md), [project_concept](./project_concept.md), [dashboard_concept](./dashboard_concept.md), [ADR-0011](../../adr/0011-rbac-row-scoping.md), [ADR-0014](../../adr/0014-application-project-lifecycle.md)

## 개요

본 도메인 API 는 프로젝트 관리 도메인의 API 계약이다. historical 모델은 `Application > Repository > Project` 였고, 현재는 `Application > Project > Repository(N:M)` 로 전환 중이다. `DEVHUB_PROJECT_MODEL=legacy|hybrid|v2` 토글을 사용하며 기본은 `hybrid`.

## 1. API ID 인덱스

| API ID | endpoint | 본문 위치 | 상태 |
| --- | --- | --- | --- |
| `API-41` | `GET /api/v1/scm/providers` | §3.1 | activated (sprint claude/work_260514-b) |
| `API-42` | `PATCH /api/v1/scm/providers/{provider_key}` | §3.1 | activated |
| `API-43` | `GET /api/v1/applications` | §4 | activated |
| `API-44` | `POST /api/v1/applications` | §4 | activated |
| `API-45` | `GET /api/v1/applications/{application_id}` | §4 | activated |
| `API-46` | `PATCH /api/v1/applications/{application_id}` | §4 | activated |
| `API-47` | `DELETE /api/v1/applications/{application_id}` (archive) | §4 | activated |
| `API-48` | `GET /api/v1/applications/{application_id}/repositories` | §5 | activated |
| `API-49` | `POST /api/v1/applications/{application_id}/repositories` | §5 | activated |
| `API-50` | `DELETE /api/v1/applications/{application_id}/repositories/{repo_key}` | §5 | activated (path: gin catch-all, `provider:org/repo` 콜론 컨벤션) |
| `API-55` | `GET /api/v1/repositories/{repository_id}/projects` + `POST` | §6 | activated (legacy/hybrid) |
| `API-56` | `GET /api/v1/projects/{project_id}` + `PATCH` + `DELETE` | §6 | activated |
| `API-56A` | `GET /api/v1/applications/{application_id}/projects` + `POST` | §6 | activated (v2/hybrid) |
| `API-56B` | `GET/POST/DELETE /api/v1/projects/{project_id}/repositories` | §6 | activated (v2/hybrid) |
| `API-57` | `GET /api/v1/applications/{application_id}/rollup` | §7 | activated (concept §13.4 normalize 실 구현 + critical 가드 흡수) |
| `API-58` | `GET /api/v1/integrations` + CRUD | §8 | activated (scope polymorphism application/project) |
| `API-93` | `GET /api/v1/applications/{application_id}/dashboard` | §9 | planned (sprint gemini/application-dashboard-concept) |

**activated 단계 정의 (sprint claude/work_260514-b)**: gin v1 group route + RBAC matrix + handler body + store body + 요청 validation + 상태 전이 가드 + audit emit. RBAC 매트릭스에서 system_admin 만 4 신규 resource (`applications` / `application_repositories` / `projects` / `scm_providers`) 의 모든 axis true (migration 000018, ADR-0011 §4.1).

**가드 1차 (concept §13.2.1 의 부분 흡수)**:
- `planning → active`: 활성 (sync_status=active) Repository ≥1
- `active → on_hold`: `hold_reason` 필수
- `on_hold → active`: `resume_reason` 필수
- `* → archived`: `archived_reason` 필수
- `closed → planning` 같은 invalid transition: `422 invalid_status_transition`
- 미흡수 (carve out): `active → closed` 의 critical 롤업 0건 검증 (롤업 store 의존, 후속 sprint)

**audit 발급 (`application.*` / `application_repository.*` / `scm_provider.*` namespace)**:
- `application.create` / `application.update` / `application.archive`
- `application_repository.link` / `application_repository.unlink`
- `scm_provider.update`
- read endpoint (list / get) 는 audit 발급하지 않음 (운영 노이즈 회피).

## 2. 공통 규칙

- 쓰기 권한: 기본 `system_admin` 전용.
- `pmo_manager`는 정책 확정 전 `disabled`, 쓰기 요청은 `403` + `error=role_not_enabled`.
- archive는 soft-delete로 처리한다 (`archived_at` 기록 + 기본 조회 목록 제외, `include_archived=true` 토글로 노출).
- `Application.key`는 시스템 전역 unique 식별자다.
- `Application.key`는 **immutable** — 발급 후 변경 불가. PATCH 본문에 `key` 가 포함되면 `422 application_key_immutable` 로 거절한다.
- `Application.key` 입력 정책: 영문/숫자 10자 (`^[A-Za-z0-9]{10}$`).
- DB 컬럼 길이는 정책 변경 대비 여유 길이로 유지하고, 길이/패턴 제한은 애플리케이션 검증에서 강제한다.
- provider 라우팅은 `repo_provider`를 기준으로 동작하며, backend는 내부 `SCM Adapter Registry`에서 provider별 어댑터를 선택한다.
- 미등록 provider 요청은 `422 unsupported_repo_provider`로 거절한다.
- **상태 전이 가드 표 SoT (Single Source of Truth)**: `./project_concept.md` §13.2.1 — 권한/검증/실패 코드 매트릭스. 본 §4 PATCH 의 규칙은 그 요약이다.
- **visibility 별 데이터 공개 범위 (초안)**:
  - `public`: 메타(key/name/owner) + 진행 요약 (집계만). 멤버 목록/원본 지표는 비공개.
  - `internal`: 조직 내 사용자에게 메타 + 멤버 목록 + 롤업 요약. 원본 PR/Build 본문은 RBAC 별도.
  - `restricted`: 멤버 본인 + system_admin 만 조회. 외부 노출 없음.

## 3. SCM Provider Catalog (planned)

### 3.1 `GET /api/v1/scm/providers` (API-41 (planned))

- 설명: 사용 가능한 SCM provider 목록 조회 (`enabled`, `display_name`, `adapter_version` 포함).
- `adapter_version`: provider 어댑터 모듈의 semver 문자열 (예: `1.4.0`). 갱신 주체는 어댑터 배포 파이프라인 (배포 후 마이그레이션/관리 API 로 등록). 운영 중 임의 수정 금지.

### 3.2 `PATCH /api/v1/scm/providers/{provider_key}` (API-42 (planned))

- 설명: provider 활성/비활성 및 운영 설정 변경 (system_admin 전용).
- 허용 필드: `enabled`, `display_name`. `adapter_version` 은 배포 파이프라인 외 수정 불가.

## 4. Application

### 4.1 `GET /api/v1/applications` (API-43 (planned))

- 설명: Application 목록 조회.
- Query:
  - `status` (optional): `planning|active|on_hold|closed|archived`
  - `include_archived` (optional, default `false`)
  - `q` (optional): `key`, `name` 검색

### 4.2 `POST /api/v1/applications` (API-44 (planned))

- 설명: Application 신규 생성.
- 요청 body 필드:
  - `key` (required): `^[A-Za-z0-9]{10}$`
  - `name` (required)
  - `description` (optional)
  - `owner_user_id` (required)
  - `start_date`, `due_date` (optional)
  - `visibility` (required): `public|internal|restricted`
  - `status` (required): `planning|active|on_hold|closed|archived`
- 실패:
  - `409 application_key_conflict`
  - `422 invalid_application_key`

요청 예시:

```json
{
  "key": "A1B2C3D4E5",
  "name": "DevHub Platform 2026",
  "description": "Cross-repo delivery governance",
  "owner_user_id": "u-1001",
  "start_date": "2026-01-01",
  "due_date": "2026-12-31",
  "visibility": "internal",
  "status": "planning"
}
```

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "id": "1a2b3c4d-1111-2222-3333-444455556666",
    "key": "A1B2C3D4E5",
    "name": "DevHub Platform 2026",
    "owner_user_id": "u-1001",
    "visibility": "internal",
    "status": "planning",
    "created_at": "2026-05-14T01:00:00Z",
    "updated_at": "2026-05-14T01:00:00Z"
  }
}
```

### 4.3 `GET /api/v1/applications/{application_id}` (API-45 (planned))

- 설명: Application 상세 조회.
- 응답 포함:
  - Application 메타
  - 연결 Repository 목록
  - 하위 Project 롤업 요약

### 4.4 `PATCH /api/v1/applications/{application_id}` (API-46 (planned))

- 설명: Application 메타/상태 수정.
- 허용 필드: `name`, `description`, `owner_user_id`, `start_date`, `due_date`, `visibility`, `status` (+ 전이별 보조 필드 `hold_reason`/`resume_reason`/`archived_reason`).
- 금지 필드: `key` (immutable — 요청 body 에 포함 시 `422 application_key_immutable`).
- 상태 전이 정책 (2026-05-28 자유화):
  - **모든 전이 허용** — 5종 (planning/active/on_hold/closed/archived) 끼리 어떤 방향이든 가능 (archived → planning 같은 "unarchive" 포함).
  - 이전엔 forward-only matrix + reason 필수 + planning→active 의 active repo ≥1 + active→closed 의 critical 0건 가드가 있었으나, 운영 유연성 우선으로 모두 제거.
  - `hold_reason` / `resume_reason` / `archived_reason` 필드는 **audit details 기록용 optional** 메타. 비어 있어도 200.
  - `key` 만 immutable 유지 (`422 application_key_immutable`).
- 변경 history: 자유화 sprint `claude/work_260528-archived-hard-delete` (2026-05-28). 자유화 이전 가드 (active repo ≥1, critical 0건 등) 는 운영자 권한/감사 기반으로 제어.

요청 예시:

```json
{
  "status": "on_hold",
  "hold_reason": "External dependency delay"
}
```

### 4.5 `DELETE /api/v1/applications/{application_id}` (API-47 (planned))

- 설명: archive (soft-delete) **+ archived 상태에서 `?hard=true` 시 영구 삭제** (sprint `claude/work_260528-archived-hard-delete` 활성화).
- 동작:
  - `?hard=true` + `status=archived` → hard-delete (`DELETE FROM applications`, FK cascade: application_repositories 자동 삭제, projects.application_id 는 ON DELETE SET NULL 로 보존). audit `application.deleted`.
  - `?hard=true` + `status≠archived` → **400 `application_not_archived`** (archive 후 다시 호출 요구).
  - 그 외 → archive (`status=archived`, `archived_at` 기록). audit `application.archived`. `include_archived=true` 토글로만 list 노출.
- Project 측 동일 패턴: `DELETE /api/v1/projects/{project_id}?hard=true` (이미 활성화됨).

## 5. Application-Repository 연결

### 5.1 `GET /api/v1/applications/{application_id}/repositories` (API-48 (planned))

- 설명: Application에 연결된 Repository 조회. 직접 link + 프로젝트 경유 간접 link 의 UNION (#395 활성화, #395/#396 후속 carve).
- 응답 data 의 각 link 객체에 `link_source` 필드 (`direct` | `via_project`) — UI/디버깅/감사용:
  - `direct`: `application_repositories` 의 직접 link (SCM webhook sync metadata 보유)
  - `via_project`: 본 application 의 하위 project 경유 간접 link — sync 메타는 의미적 default (`sync_status=active`, `sync_error_*` 빈/NULL, `last_sync_at=pr.linked_at`)
  - 같은 (repo_provider, repo_full_name) 이 direct + via_project 양쪽 source 로 잡히면 두 row 모두 응답 (UI 가 구분 표기).

### 5.2 `POST /api/v1/applications/{application_id}/repositories` (API-49 (planned))

- 설명: Repository 연결 생성.
- 요청 body 필드:
  - `repo_provider` (required): `bitbucket|gitea|forgejo|github|...` (동등 지원, 특정 provider 우선순위 없음)
  - `repo_full_name` (required): `org/repo`
  - `role` (required): `primary|sub|shared`
- 실패:
  - `409 repository_link_conflict`
  - `422 invalid_repository_reference`
  - `422 unsupported_repo_provider`
  - `503 provider_unreachable`

- 연결 lifecycle:
  - 초기: `requested`
  - 검증중: `verifying`
  - 정상: `active`
  - 부분장애: `degraded` (`sync_error_code` 기록)
  - 해제: `disconnected`
- `sync_error_code` 표준 (link 단위 최신 1건 캐시):
  - `provider_unreachable` (retryable=true)
  - `auth_invalid` (retryable=false)
  - `permission_denied` (retryable=false)
  - `rate_limited` (retryable=true)
  - `webhook_signature_invalid` (retryable=false)
  - `payload_schema_mismatch` (retryable=false)
  - `resource_not_found` (retryable=false)
  - `internal_adapter_error` (retryable=true)
- **scope**: 본 `sync_error_code` 는 **link (application_id, repo_provider, repo_full_name) 단위의 최신 에러 1건만 캐시**한다. 개별 webhook event/payload 단위 상세 에러는 `webhook_events` (현행) 또는 후속 `adapter_event_logs` 테이블에 보관 (예: 동일 link 에서 1시간 동안 발생한 N건의 `webhook_signature_invalid` 는 link.sync_error_code 에 최신 1건 + 카운트는 별도 테이블).

요청 예시:

```json
{
  "repo_provider": "bitbucket",
  "repo_full_name": "team/devhub-core",
  "role": "primary"
}
```

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "application_id": "1a2b3c4d-1111-2222-3333-444455556666",
    "repo_provider": "bitbucket",
    "repo_full_name": "team/devhub-core",
    "role": "primary",
    "sync_status": "verifying",
    "sync_error_code": null,
    "sync_error_retryable": null,
    "linked_at": "2026-05-14T01:10:00Z"
  }
}
```

### 5.3 `DELETE /api/v1/applications/{application_id}/repositories/{repo_key}` (API-50 (planned))

- 설명: Application-Repository 연결 해제.

## 6. Project + Repository 연결

운영 모델은 `Application > Project > Repository(N:M)` 을 기본으로 한다. legacy 경로(`/repositories/{repository_id}/projects`)는 호환을 위해 유지되며 `DEVHUB_PROJECT_MODEL=v2` 에서는 `410 gone` 으로 비활성화된다.

### 6.1 `GET /api/v1/repositories/{repository_id}/projects` (API-55 (planned))

- 설명: Repository 하위 Project 목록 조회.
- Query:
  - `status` (optional)
  - `include_archived` (optional, default `false`)

### 6.2 `POST /api/v1/repositories/{repository_id}/projects` (API-55 (planned))

- 설명: Project 생성.
- 요청 body 필드:
  - `key` (required)
  - `name` (required)
  - `description` (optional)
  - `owner_user_id` (required)
  - `start_date`, `due_date` (optional)
  - `visibility` (required)
  - `status` (required)
- 제약:
  - `UNIQUE (repository_id, key)`

### 6.3 `GET /api/v1/applications/{application_id}/projects` (API-56A)

- 설명: Application 하위 Project 목록 조회 (v2 경로).

### 6.4 `POST /api/v1/applications/{application_id}/projects` (API-56A)

- 설명: Application 하위 Project 생성 (v2 경로).
- 요청 body:
  - `repository_id` (required): legacy primary repository
  - `repository_ids` (optional): N:M 연결할 repository 목록
- 동작: 단일 트랜잭션으로 `projects` row + `project_repositories` link rows 생성.

### 6.5 `GET /api/v1/projects/{project_id}/repositories` (API-56B)

- 설명: Project 와 연결된 repository 링크 조회.

### 6.6 `POST /api/v1/projects/{project_id}/repositories` (API-56B)

- 설명: Project-repository 링크 추가.

### 6.7 `DELETE /api/v1/projects/{project_id}/repositories/{repository_id}` (API-56B)

- 설명: Project-repository 링크 제거.

### 6.8 `GET /api/v1/projects/{project_id}` (API-56 (planned))

- 설명: Project 상세 조회.
- 응답 포함:
  - Project 메타
  - 멤버/owner
  - 상/하위 마일스톤 매핑 요약
  - Integration 상태 요약

### 6.9 `PATCH /api/v1/projects/{project_id}` (API-56 (planned))

- 설명: Project 메타/상태 수정.
- 요청 body 필드 (모두 optional, nil = 변경 안 함):
  - `name`, `description`, `owner_user_id`, `visibility`, `status`, `start_date`, `due_date`
  - `application_id` — application 이전/해제 (#395/#396 후속 carve):
    - `""` (빈 string) → 해제 (NULL, ON DELETE SET NULL 의도 일치)
    - non-empty UUID → 해당 application 으로 이전 (존재하지 않으면 422 `application_id_invalid`)
  - `hold_reason` / `resume_reason` / `archived_reason` — status 전이 필수 payload
- audit: `project.updated` event 의 details 에 `application_id_from` / `application_id_to` 기록 (변경된 경우).

### 6.10 `DELETE /api/v1/projects/{project_id}` (API-56 (planned))

- 설명: Project archive (soft-delete) — Application archive 와 동일 규칙.

## 7. Application 롤업

### 7.1 `GET /api/v1/applications/{application_id}/rollup` (API-57 (planned))

- 설명: Repository 단위 운영 지표를 Application 레벨로 집계해 조회.
- Query:
  - `weight_policy` (optional, default `equal`): `equal|repo_role|custom`
- 최소 집계 항목:
  - PR 상태 분포/open 지속시간
  - 빌드 성공률/평균 소요시간 (시계열 추세 차트 source; 카드/리스트는 `target_branch_build_status` 사용 — REQ-FR-APPDASH-001)
  - `target_branch_build_status`: 연결된 repo 의 **마지막 빌드 결과** 종합 derive (`healthy`/`broken`/`unknown`)
  - 품질 점수 평균/게이트 실패 건수
- `meta` 필드(필수):
  - `period`: 집계 기간
  - `filters`: 적용 필터
  - `weight_policy`: 가중치 정책
  - `applied_weights`: repo별 최종 적용 가중치 맵
  - `fallbacks`: 가중치 누락/정책 불일치 시 적용된 fallback 목록
  - `data_gaps`: 누락/장애 provider 또는 repository 목록
- 검증:
  - `weight_policy=custom` 인 경우 `custom_weights`의 합은 1.0(±허용오차)이어야 한다. **허용오차 기본값 = ±0.001** (1e-3) — 합이 [0.999, 1.001] 범위면 통과.
  - 음수 가중치는 허용하지 않는다 (`422 invalid_weight_policy`).
- **Normalize 규칙 (weight_policy 별)**:
  - `equal`: 모든 연결 Repository 가 `1/N` 가중치. 0개면 `weight_policy=equal` 이라도 가중치 맵은 빈 객체, 결과는 `data_gap`.
  - `repo_role`: 기본 카탈로그 `primary=0.6 / sub=0.3 / shared=0.1`. 단일 카테고리 내 다중 repo 가 있으면 카테고리 가중치를 균등 분할 (예: primary 2개면 각 0.3). 카테고리가 0개면 해당 가중치는 다른 카테고리에 비례 재분배 후 정규화.
  - `custom`: 명시되지 않은 repo 는 `equal` 카테고리 가중치로 fallback 후 합 정규화. `fallbacks` 메타에 `reason="custom_weight_missing"` 으로 기록.

응답 예시:

```json
{
  "status": "ok",
  "data": {
    "pull_request_distribution": {
      "open": 24,
      "draft": 5,
      "merged": 132,
      "closed": 11
    },
    "build_success_rate": 0.94,
    "target_branch_build_status": "broken",
    "build_avg_duration_seconds": 412,
    "quality_score": 83.4,
    "quality_gate_failed_count": 2
  },
  "meta": {
    "period": {
      "from": "2026-05-01T00:00:00Z",
      "to": "2026-05-14T00:00:00Z"
    },
    "filters": {
      "repository_roles": ["primary", "sub", "shared"]
    },
    "weight_policy": "repo_role",
    "applied_weights": {
      "team/devhub-core": 0.6,
      "team/devhub-web": 0.3,
      "shared/devhub-lib": 0.1
    },
    "fallbacks": [],
    "data_gaps": [
      {
        "repo_full_name": "shared/devhub-lib",
        "provider": "forgejo",
        "reason": "provider_unreachable"
      }
    ]
  }
}
```

## 8. Integration

### 8.1 `GET /api/v1/integrations` (API-58 (planned))

- 설명: Application/Project 연계 통합 설정 조회.

### 8.2 `POST /api/v1/integrations` (API-58 (planned))

- 설명: Jira/Confluence 연계 생성.
- 요청 body 필드:
  - `scope`: `application|project`
  - `integration_type`: `jira|confluence`
  - `external_key`, `url`
  - `policy`: `summary_only|execution_system`

### 8.3 `PATCH /api/v1/integrations/{integration_id}` (API-58 (planned))

- 설명: 연계 정책/키 수정.

### 8.4 `DELETE /api/v1/integrations/{integration_id}` (API-58 (planned))

- 설명: 연계 해제.

> **Jira 정책 cross-cut 메모 (`REQ-FR-PROJ-005` 후속)**: REQ-FR-PROJ-005 는 "Repository Jira 가 실행 SoT" 라는 하이브리드 정책을 명시한다. 그러나 `repo_provider` 가 `bitbucket|gitea|forgejo` 인 경우의 Jira 매핑 (= 비-Jira SCM 의 실행 이슈를 어떻게 Jira project 와 묶는가) 은 본 sprint 에서 결정되지 않음. concept §10 미해결 항목으로 이관, Integration sprint 에서 결정.

## 9. Application 개발 대시보드 API

### 9.1 `GET /api/v1/applications/{application_id}/dashboard` (API-93)

- **설명**: Application 상세 대시보드용 실시간 빌드 상태, 다차원 품질 메트릭, 하위 프로젝트 진척율 및 지연 리스크 배지, 매핑된 DREQ 목록, SCM 및 빌드 시계열 트렌드 데이터를 일괄 병렬 집계하여 반환합니다.
- **인증**: OIDC + RBAC `applications:view`.
- **에러**:
  - `404 application_not_found`: 존재하지 않는 Application ID
  - `403 Forbidden`: 권한 부족 또는 onboarding_required 미결 완료 상태

요청 예시:
`GET /api/v1/applications/1a2b3c4d-1111-2222-3333-444455556666/dashboard`

응답 예시 (대시보드 페이로드 — 본문은 master `docs/backend_api_contract.md` §13.10 참조; 본 문서는 endpoint·인증·에러 계약을 SoT 로 보존하고 페이로드 sample 은 master 원본을 인용 형태로 유지한다).

## 10. 공통 에러 코드 (초안)

```text
role_not_enabled
application_key_conflict
invalid_application_key
application_key_immutable
application_activation_precondition_failed
application_close_precondition_failed
invalid_status_transition_payload
invalid_status_transition
repository_link_conflict
invalid_repository_reference
unsupported_repo_provider
provider_unreachable
webhook_signature_invalid
invalid_weight_policy
project_key_conflict
integration_policy_violation
integration_provider_required
integration_provider_not_found
integration_sync_unsupported_provider_type
integration_capability_not_enabled
integration_provider_not_gitea_compatible
integration_base_url_missing
integration_outbound_credentials_missing
integration_scm_create_failed
integration_scm_auth_failed
```

## 11. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §13 (Application/Repository/Project 본문) 을 도메인 sub-document 로 이관. ID(API-41..50, 55..58, 56A/56B, 57, 58, 93) 보존, 신규 발급/삭제 없음. API-51..54 (repository 운영 지표) + API-91/92 (draft→publish) 는 repository-integration api 로 분리. API-88/89/90 (외부 SCM 원격 import/create) 는 integration-registry api 에 위치. §13.10 의 대시보드 응답 페이로드 sample 은 length 우려로 master 인용 유지. |
