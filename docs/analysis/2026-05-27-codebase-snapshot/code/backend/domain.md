# backend-core/internal/domain — 도메인 모델 분석

- 문서 목적: `backend-core/internal/domain/` 의 aggregate struct·enum·상태머신·검증 helper 를 정밀 분석한다.
- 범위: 8 Go 파일 (`domain.go`, `application.go`, `dev_request.go`, `primary_unit.go`, `rbac.go` + 테스트 3: `command_test.go`, `primary_unit_test.go`, `rbac_test.go`).
- 대상 독자: 도메인 모델 변경자, 새 aggregate/enum 추가자, store·handler 작업자.
- 상태: snapshot (2026-05-27, main `cf19c94`)
- 관련 문서: `store.md`, `migrations.md`

## 1. 파일 구성

| 파일 | 담당 | 비고 |
|------|------|------|
| `domain.go` (489 L) | SCM mirror(Repository/User/Issue/PullRequest/CIRun/Risk) + Command + AuditLog + 사용자/조직(AppUser/OrgUnit/UnitAppointment) + Sink interface + 다수 enum | `CommandStatus` 6-state 머신 정의처 |
| `application.go` (425 L) | Application/Project aggregate + ApplicationRepository link + ProjectIntegration + IntegrationProvider/Binding + SCMProvider + 롤업(rollup) 타입 | `IsRetryableSyncError`, `ResolveOutboundAuth` 정의처 |
| `dev_request.go` (119 L) | DevRequest aggregate + DevRequestIntakeToken | `DevRequestStatusTransitions` 6-state 머신 정의처 |
| `rbac.go` (284 L) | Resource/Action enum + PermissionMatrix + RBACRole + 기본 매트릭스 | `DefaultPermissionMatrix`, `EnforceAuditInvariant`, `ValidateRoleID` |
| `primary_unit.go` (51 L) | `ResolvePrimaryUnit` 결정적 fallback 알고리즘 | leader 우선 + total_count desc + unit_id 사전순 tiebreak |

도메인 패키지는 의존성이 표준 라이브러리(`time`/`sort`/`regexp`/`fmt`/`context`)뿐 — store/handler 에 의존하지 않는 순수 모델 계층.

## 2. Aggregate 별 struct + 핵심 필드

### Repository (`domain.go:8`) — 소유권 분리(migration 000042/000045)
SCM mirror 필드(GiteaID/FullName/OwnerLogin/Name/CloneURL/HTMLURL/DefaultBranch/Private/Status/Publish*At) + 시스템 메타 분리:
- `Source string` — `"scm"`|`"system"` (빈값=legacy scm 취급). 상수 `RepositorySourceSCM`/`RepositorySourceSystem` (`domain.go:31`).
- `ProviderID string` — integration_providers(scm) FK (단일 출처, 000045 통합).
- `ProviderKey string` — derived (join read-only 표시용).
- `Description string` — system-owned, SCM sync 가 덮어쓰지 않음.

### Command (`domain.go:167`)
`CommandID/CommandType/TargetType/TargetID/ActionType/Status/ActorLogin/Reason/DryRun/RequiresApproval/IdempotencyKey/RequestPayload/ResultPayload`. 요청 DTO 3종: `RiskMitigationCommandRequest`/`ServiceActionCommandRequest`/`CommandApprovalRequest` (모두 SourceIP/RequestID/SourceType audit enrichment 보유).

### AuditLog (`domain.go:202`)
`AuditID/ActorLogin/Action/TargetType/TargetID/CommandID/Payload` + enrichment(`SourceIP/RequestID/SourceType`) + `SourceEventID`(dedup key, migration 000032 partial UNIQUE 매핑, 빈값이면 제약 미적용).

### AppUser (`domain.go:347`)
`UserID/Email/DisplayName/Role(AppRole)/Status(UserStatus)/Type(UserType)/IdPSubject/PrimaryUnitID/CurrentUnitID/IsSeconded/JoinedAt/Appointments` + onboarding 상태:
- `OnboardingCompletedAt *time.Time` — nil=미완료.
- `ReviewStatus string` — `"pending_review"`|`"reviewed"`, 빈값=NULL=미제출.
- bi-implication: `OnboardingCompletedAt` 와 `ReviewStatus` 는 동시 NULL 또는 동시 NOT NULL (CHECK `users_onboarding_review_consistency`).

입력 DTO: `CreateUserInput`/`UpdateUserInput`(pointer 필드 = optional)/`OnboardingSubmitInput`. `UpdateUserInput.ReviewStatus` 변경은 store 가 primary_unit 변경 시 pending_review reset 정책과 연동.

### OrgUnit / UnitAppointment / Hierarchy (`domain.go:386`)
`OrgUnit`(UnitID/ParentUnitID/UnitType/Label/LeaderUserID/Position/DirectCount/TotalCount). `UnitAppointment`(UnitID/UserID/AppointmentRole). `Hierarchy`(Units+Edges).

### Application (`application.go:115`)
`ID(UUID)/Key(immutable)/Name/Description/Status/Visibility/OwnerUserID(legacy)/LeaderUserID/DevelopmentUnitID/StartDate/DueDate/ArchivedAt`.

### ApplicationRepository (`application.go:134`)
composite PK = (ApplicationID, RepoProvider, RepoFullName). `Role/SyncStatus/SyncErrorCode/SyncErrorRetryable *bool/SyncErrorAt/LastSyncAt`.

### Project (`application.go:149`)
`ID(UUID)/ApplicationID(opt)/RepositoryID(legacy primary)/Key/Name/Status(ProjectStatus alias)/Visibility/OwnerUserID/dates`. `ProjectRepository`(N:M link), `ProjectMember`(role).

### IntegrationProvider (`application.go:228`)
`ProviderKey/ProviderType/DisplayName/Enabled/AuthMode/CredentialsRef/Capabilities/SyncStatus/BaseURL/APIToken` + auth_mode 구조화 자격증명: `AuthUsername/AuthClientID/AuthTokenURL/AuthSecret`. `OutboundAuth`(`application.go:254`) 는 active mode 별 자격증명 resolve 결과.

### DevRequest (`dev_request.go:35`)
`ID/Title/Details/Requester/AssigneeUserID/SourceSystem/ExternalRef/Status/RegisteredTargetType/RegisteredTargetID/RejectedReason/ReceivedAt`. (SourceSystem, ExternalRef) UNIQUE = idempotency. `DevRequestIntakeToken`(`dev_request.go:96`): HashedToken(SHA-256 hex, plain 미저장)/AllowedIPs(CIDR)/SourceSystem/RevokedAt/ExpiresAt + `IsActive()` helper.

## 3. enum 카탈로그

| enum (타입) | 값 | 정의 |
|-------------|----|----|
| `CommandStatus` | pending/running/succeeded/failed/rejected/cancelled | `domain.go:114` |
| `AuditSourceType` | oidc/webhook/kratos(legacy)/system/keycloak_event | `domain.go:194` |
| `AppRole` | developer/manager/system_admin/pmo_manager | `domain.go:305` |
| `UserType` | human/system | `domain.go:317` |
| `UserStatus` | active/pending/deactivated | `domain.go:322` |
| `UnitType` | company/division/team/group/part | `domain.go:330` |
| `AppointmentRole` | leader/member | `domain.go:340` |
| `ReviewStatus`(const) | pending_review/reviewed | `domain.go:381` |
| `ApplicationStatus` | planning/active/on_hold/closed/archived | `application.go:8` |
| `ProjectStatus` | = ApplicationStatus alias | `application.go:23` |
| `ApplicationVisibility` | public/internal/restricted | `application.go:28` |
| `ApplicationRepositoryRole` | primary/sub/shared | `application.go:37` |
| `ApplicationRepositorySyncStatus` | requested/verifying/active/degraded/disconnected | `application.go:46` |
| `SyncErrorCode` | provider_unreachable/auth_invalid/permission_denied/rate_limited/webhook_signature_invalid/payload_schema_mismatch/resource_not_found/internal_adapter_error | `application.go:58` |
| `ProjectMemberRole` | lead/contributor/observer | `application.go:84` |
| `IntegrationScope`/`IntegrationScopeType` | application/project | `application.go:93`/`:222` |
| `IntegrationType` | jira/confluence | `application.go:101` |
| `IntegrationPolicy` | summary_only/execution_system | `application.go:109` |
| `IntegrationProviderType` | alm/scm/ci_cd/doc/infra | `application.go:200` |
| `IntegrationAuthMode` | token/basic/oauth2/app_password/agent | `application.go:211` |
| `WeightPolicy` | equal/repo_role/custom | `application.go:365` |
| `DevRequestStatus` | received/pending/in_review/registered/rejected/closed | `dev_request.go:8` |
| `DevRequestTargetType` | application/project | `dev_request.go:29` |
| `Resource`(RBAC) | infrastructure/pipelines/organization/security/audit/applications/application_repositories/projects/scm_providers/dev_requests/dev_request_intake_tokens (11종) | `rbac.go:12` |
| `Action`(RBAC) | view/create/edit/delete | `rbac.go:32` |

## 4. 상태머신

### Command 6-state (`domain.go:114`-165)
- terminal: succeeded/failed/rejected/cancelled (`commandTerminalStates` `domain.go:125`).
- 허용 전이(`commandValidTransitions` `domain.go:136`):
  - `pending → running | rejected | cancelled`
  - `running → succeeded | failed | cancelled`
  - terminal 4종 → (없음)
- helper: `IsTerminal()` / `CanTransitionTo(next)` (same-state 는 false=no-op).

### DevRequest 6-state (`dev_request.go:52`-92)
`DevRequestStatusTransitions` 표:

| from \ to | received | pending | in_review | registered | rejected | closed |
|-----------|:--:|:--:|:--:|:--:|:--:|:--:|
| received | - | ✓ | - | - | ✓(invalid_intake) | - |
| pending | - | - | ✓ | ✓ | ✓ | - |
| in_review | - | - | - | ✓ | ✓ | - |
| rejected | - | ✓(reopen) | - | - | - | ✓ |
| registered | - | - | - | - | - | ✓ |
| closed | - | - | - | - | - | - |

helper: `IsValidDevRequestTransition(from,to)` (`dev_request.go:86`). closed 가 유일 종착.

### Application / Project status (`application.go:8`, `:23`)
planning/active/on_hold/closed/archived 5종. 전이표는 도메인 코드에 없고 handler+DB CHECK(`applications_status_check`)로 강제. `ProjectStatus` 는 `ApplicationStatus` 의 **type alias**(`=`) — 향후 Project 전용 상태(예: cancelled) 분기 시 alias 를 끊을 확장 지점으로 설계됨(`application.go:16` 주석). archived 시 archived_at 필수(CHECK `applications_archived_consistency`/`projects_archived_consistency`) — store 의 CASE 절이 자동 set.

### Repository source / status draft→active (`domain.go:18`, migration 000042/000043)
- `source`: scm|system (빈값=scm).
- `repository_status`: draft|active (CHECK `repositories_status_check`). draft 생성(`CreateRepositoryDraft`, source=system) → `MarkRepositoryDraftPublishRequested`(publish_requested_at set) → publish 시 active. SCM sync upsert 는 `repository_status='active'` 로 직행.

### AppUser onboarding / review_status (`domain.go:347`, migration 000033)
- `onboarding_completed_at` NULL ↔ `review_status` NULL (bi-implication).
- 흐름: admin pre-seed(둘 다 NULL) 또는 미등록 → `SubmitOnboarding`(completed_at=NOW, review_status='pending_review') → system_admin `ConfirmUserReview`(→'reviewed'). primary_unit 변경 시 'pending_review' reset 정책(ADR-0021 §3.2). `pending_review` 사용자는 무소속 취급(ARCH-ONBOARD-02).

## 5. 검증 / resolve helper

- **`IsRetryableSyncError(code)`** (`application.go:71`): provider_unreachable/rate_limited/internal_adapter_error 만 true, 나머지·unknown 은 false. store 의 `UpdateApplicationRepositorySync` 가 retryable 플래그 산출에 사용.
- **`ResolveOutboundAuth()`** (`application.go:265`): `IntegrationProvider` receiver. AuthMode 별로 자격증명 컬럼을 `OutboundAuth` 로 매핑 — basic/app_password→Username+Secret, oauth2→ClientID+TokenURL+Secret, agent→Username, default(token/unset)→APIToken(Mode 를 token 으로 강제). Gitea adapter 의 Authorization 헤더 생성 근거.
- **`ResolvePrimaryUnit(appointments, unitTotalCounts)`** (`primary_unit.go:26`): 결정적 fallback — (1) leader appointment 만 후보(없으면 전체), (2) total_count desc 정렬, (3) unit_id 사전순 tiebreak. 빈 입력 → `("", false)`. 캐시된 MV(000011) 또는 fresh count 를 caller 가 공급. admin-set primary_unit 신뢰 가능 시 호출 금지(override 아님).
- **RBAC helpers** (`rbac.go`): `AllResources()`(11종)/`AllActions()`(4종)/`SystemRoleIDs()`(4종)/`IsSystemRole`/`ValidateRoleID`(system 또는 `^custom-[a-z0-9][a-z0-9_-]{0,62}$`)/`EnforceAuditInvariant`(audit 의 create/edit/delete 를 false 강제, 누락 resource 는 all-false 로 채움)/`Allows(matrix,r,a)`/`DefaultPermissionMatrix(roleID)`(4 system role 별 기본 매트릭스)/`SystemRoles`. seed migration 과 byte-for-byte 정합이 설계 의도.

## 발견 사항 (불일치/stale/부채)

1. **`AuditSourceKratos` deprecated enum 잔존** — `domain.go:197` 의 `AuditSourceKratos = "kratos"` 는 ADR-0001(superseded by ADR-0019)의 legacy 값. 코드 주석이 "historical audit_logs rows decode cleanly" 목적 보존을 명시하지만, 신규 코드가 실수로 참조할 수 있는 dead enum. 마찬가지로 `IdPSubject` 필드 주석(`domain.go:355`)이 "formerly Kratos identity.id" 로 Kratos 잔재를 설명만 함.

2. **평문 secret 필드가 도메인 struct 에 노출** — `IntegrationProvider.APIToken`/`AuthSecret`/`CredentialsRef` (`application.go:235`, `:241`, `:247`)가 평문 string 필드. `ResolveOutboundAuth` 가 이 raw 값을 `OutboundAuth.Token`/`Secret` 으로 그대로 복사. 도메인 레이어엔 마스킹/암호화 개념 없음 — write-only 보호는 handler 의 응답 매핑에만 의존(store.md 발견 1 과 동일 부채, #6 carve).

3. **`ProjectStatus` alias 의 enum 누설 위험** — `ProjectStatus = ApplicationStatus` (`application.go:23`)이므로 Project 에 `ApplicationStatus*` 상수를 그대로 쓴다. 향후 Project 전용 상태 추가 시 alias 를 끊어야 하나, 현 시점엔 두 도메인이 같은 CHECK 어휘를 공유해 한쪽 변경이 다른 쪽에 의도치 않게 전파될 구조적 위험(주석으로만 경고).

4. **상태머신이 코드↔DB 이중 정의** — Command/DevRequest 전이표는 도메인 코드에 있으나, Application/Project status·Repository draft·AppUser onboarding 전이는 도메인 코드에 전이 맵이 **없고** DB CHECK + handler 로직으로만 강제된다. enum 값 추가 시 도메인 상수·DB CHECK·handler 3곳을 수동 동기화해야 하는 부채(예: ApplicationStatus 5종이 `applications_status_check` 와 별도 관리).

5. **`AllResources()` 순서 의존성** — `EnforceAuditInvariant`(`rbac.go:126`)가 `AllResources()` 를 iterate 해 매트릭스를 재구성하므로, 신규 Resource 추가 시 `AllResources()` 등록을 누락하면 그 resource 권한이 조용히 누락된다(매트릭스에서 사라짐). seed migration(`migrations.md` 의 000018/000024/000026)과의 정합도 수동.

6. **`SyncErrorCode` 도메인 dictionary ↔ DB CHECK 동기화 부채** — 8종 SyncErrorCode(`application.go:58`)가 migration 000014 의 `application_repositories_sync_error_code_check` 와 동일 목록을 손으로 맞춘다. 한쪽만 늘리면 `IsRetryableSyncError` 의 default(false) 또는 DB CHECK 위반으로 갈림.
