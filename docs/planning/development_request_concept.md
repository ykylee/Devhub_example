# 개발 의뢰 (Dev Request, DREQ) 도메인 컨셉

- 문서 목적: DevHub 의 신규 1차 도메인 **개발 의뢰 (Dev Request)** 의 컨셉을 정의해 후속 요구사항/usecase/설계 단계의 공통 기준으로 삼는다. 외부 시스템에서 의뢰가 API 로 들어오고, 담당자가 검토 후 application/repository/project 중 하나로 등록(promote)하는 운영 흐름이다.
- 범위: 컨셉 단계 + 본 sprint 가 함께 발급할 후속 ID(REQ-FR-DREQ / UC-DREQ / ARCH-DREQ / API-DREQ) 의 사전 정의. lifecycle / 데이터 모델 초안 / 행위자 × 핵심 usecase / 외부 수신 인증 정책 후보 / 다른 도메인과의 연결 / MVP scope / out-of-scope / 미해결 항목 / 후속 sprint hook.
- 대상 독자: 프로젝트 리드, Backend / Frontend / Auth 트랙 담당자, 외부 시스템 통합 담당자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-15
- 결정 근거 sprint: `claude/work_260515-f`
- 관련 문서: [`docs/requirements.md §5.X DREQ`](../requirements.md), [`docs/planning/system_usecases.md` UC-DREQ-*](./system_usecases.md), [`docs/architecture.md` ARCH-DREQ §](../architecture.md), [`docs/backend_api_contract.md` API-DREQ §](../backend_api_contract.md), [`docs/planning/project_management_concept.md`](./project_management_concept.md), [`docs/traceability/report.md` §3 DREQ 도메인 row](../traceability/report.md).

## 1. 컨셉 정리 배경

- DevHub 가 Application / Repository / Project 3-tier 운영 모델을 정착(2026-05-14 백엔드 1차 완성)시킨 직후, **그 모델로 들어오는 작업이 어디서 시작되는가** 라는 upstream 흐름이 비어 있다.
- 현재는 system_admin 이 직접 application/project 를 생성한다 — 그러나 실제 운영 시나리오는 **외부 시스템 (운영 포털, ITSM, Jira, 사내 워크플로우 도구)** 에서 의뢰가 발생 → DevHub 가 그 의뢰를 받아 담당자에게 라우팅 → 담당자가 적절한 entity 로 등록한다.
- 따라서 본 컨셉은:
  - **DREQ entity** 의 lifecycle 과 데이터 모델을 정의하고,
  - **외부 수신 → 대기 → 검토 → 등록(promote)** 의 4-step 운영 흐름을 공식화하고,
  - DREQ 와 Application / Repository / Project 의 연결 관계를 정의하고,
  - **MVP scope** 와 **후속 (인증 정책, 자동 매핑, AI 분류 등)** 를 분리한다.
- 본 컨셉이 머지되면 후속 sprint 에서 단계적으로 backend store / API / frontend UI / e2e 로 발전시킨다.

## 2. 도메인 정의

### 2.1 entity: DevRequest

외부 시스템에서 들어온 1건의 개발 작업 의뢰. 본질은 "아직 어디에 매핑될지 결정되지 않은 작업 단위". 등록(promote) 시점에 application/project 와 매핑되며, 매핑 후에도 audit 보존 목적으로 row 는 유지된다.

### 2.2 필드 (1차)

| 필드 | 타입 | 필수 | 메모 |
| --- | --- | --- | --- |
| `id` | UUID | yes | 시스템 발급. |
| `title` | string(≤200) | yes | 의뢰 제목. |
| `details` | text | no | 상세 내용. markdown 허용 (XSS 정책은 §7). |
| `requester` | string(≤120) | yes | 외부 시스템상의 의뢰자 식별자. user 매핑은 1차 비강제 (외부 외 부서/시스템도 의뢰 가능). |
| `assignee_user_id` | string FK `users.user_id` | yes | 담당자. DREQ 가 dashboard 에 표시되는 기준. |
| `source_system` | string(≤60) | yes | 외부 시스템 이름 (예: `ops_portal`, `jira_pmo`). API 호출 시 인증된 클라이언트가 자동 채움. |
| `external_ref` | string(≤120) | no | 외부 시스템의 ticket id 등 역참조용. (source_system, external_ref) UNIQUE — 중복 수신 방지. |
| `status` | enum | yes | `received / pending / in_review / registered / rejected / closed` (§4 참조). |
| `registered_target_type` | enum | no | `application / project` — 등록(promote) 시 채워짐. repository 는 application/project 와 연동되므로 별도 type 으로 두지 않고 application 의 repository link 에 흡수. |
| `registered_target_id` | string | no | application_id 또는 project_id. |
| `rejected_reason` | text | no | rejected 상태로 전이 시 필수. |
| `received_at` | timestamptz | yes | 외부 시스템에서 들어온 시각. |
| `created_at` / `updated_at` | timestamptz | yes | 행 자체 생명주기. |

> `source_system` / `external_ref` / `requester` 는 외부 시스템 호출에서 채워진다. DREQ 가 외부 시스템과 1:1 매핑된 경우 `external_ref` 가 우리 측 dedup key 가 된다.

### 2.3 상태 전이 (lifecycle)

```
[external POST]
     │
     ▼
  received ─── (자동, 검증 통과 시) ──▶ pending
     │
     └── (자동, 검증 실패: assignee 미존재 / 필수 필드 누락) ──▶ rejected (reason: invalid_intake)

  pending ─── (담당자 dashboard 열람 시 자동) ──▶ in_review
  pending ─── (담당자 명시 reject) ──▶ rejected
  pending ─── (담당자가 application/project 로 등록) ──▶ registered

  in_review ─── (담당자 명시 reject) ──▶ rejected
  in_review ─── (담당자가 application/project 로 등록) ──▶ registered

  rejected ─── (담당자 명시 reopen) ──▶ pending  (이력 보존)
  registered ─── (담당자 명시 close, 30일 보존 후 archive 정책 carve out) ──▶ closed
```

- `pending → in_review` 는 dashboard 열람 자동 전이 (선택사항 — 담당자가 "맡았다" 는 신호로 사용). 명시적 acknowledge 도 허용한다.
- 모든 상태 전이는 audit log 에 기록 (`dev_request.{received,registered,rejected,reopened,closed}`).

## 3. 행위자 × 핵심 usecase

| 행위자 | 핵심 usecase |
| --- | --- |
| **외부 시스템** | 의뢰를 API 로 등록한다 (DREQ-Submit). |
| **담당자 (assignee)** | 내 대기 의뢰를 본다 (DREQ-MyQueue) / 의뢰 상세를 본다 (DREQ-View) / 의뢰를 application 또는 project 로 등록한다 (DREQ-Promote) / 의뢰를 reject 한다 (DREQ-Reject). |
| **시스템 관리자** | 전체 대기 목록을 본다 (DREQ-AdminList) / 임의 의뢰를 다른 담당자로 재할당한다 (DREQ-Reassign) / 만료/이상 의뢰를 close 한다 (DREQ-Close). |
| **의뢰자 (requester, 외부)** | 1차에서는 DevHub 내 직접 조회 권한 없음 (외부 시스템에서 자체 추적). 2차 carve out — webhook 또는 callback API 로 상태 알림. |

## 4. UI surface (예상 — frontend 구현 sprint 에서 확정)

- **담당자 dashboard**: 기존 dashboard 에 "내 대기 의뢰" 위젯 신설. 카운트 + 최신 5건. 클릭 시 `/admin/settings/dev-requests?assignee=me&status=pending` 로 이동.
- **관리 페이지** (`/admin/settings/dev-requests`): system_admin / 담당자 (본인 의뢰만) 가 접근. 목록 + 필터 (status / source_system / assignee). 행 클릭 시 상세 modal.
- **상세 modal**: title / details / requester / assignee / external_ref + 액션 버튼 ("Register as Application" / "Register as Project" / "Reject" / "Reassign").
  - "Register as Application/Project" 클릭 시 기존 `ApplicationCreationModal` / `ProjectCreationModal` 로 이동하되, 의뢰 정보가 초기값으로 prefill + 등록 성공 시 DREQ 가 `registered` 로 자동 전이.

## 5. 다른 도메인과의 연결

- **Application / Project** — DREQ.registered_target_type/id 로 1:1 (또는 1:N) 매핑. application/project 의 row 에 역방향 `origin_dreq_id` (nullable) 컬럼을 둘지는 carve out — 의뢰 없이 system_admin 이 직접 생성한 application 도 존재해야 하므로 nullable.
- **RBAC** — DREQ resource 신규. system_admin (전체 관리) / 담당자 (본인 의뢰만 view/edit) 가 1차. pmo_manager 의 위양 범위는 carve out (ADR-0011 §4.2 패턴 적용).
- **Audit** — 모든 상태 전이가 `dev_request.*` action 으로 기록. 외부 수신은 `dev_request.received` + source_system payload.
- **외부 시스템 인증** — §7 참조.

## 6. MVP scope (본 컨셉 + 후속 sprint)

### 6.1 In-scope (MVP)
- 외부 수신 endpoint `POST /api/v1/dev-requests` (API 토큰 + IP allowlist 1차)
- 대기 목록 조회 `GET /api/v1/dev-requests` (필터: status / assignee / source_system)
- 상세 조회 `GET /api/v1/dev-requests/:id`
- 등록(promote) `POST /api/v1/dev-requests/:id/register` (target_type/target_id 명시) — application/project 신규 생성과 단일 트랜잭션
- reject `POST /api/v1/dev-requests/:id/reject` (reason 필수)
- reassign `PATCH /api/v1/dev-requests/:id` (assignee_user_id 변경 — system_admin)
- close `DELETE /api/v1/dev-requests/:id` (registered/rejected 만)
- audit 기본 6 action

### 6.2 Out-of-scope (후속)
- AI 자동 분류 / 자동 application 매핑 추천
- 외부 시스템 callback (의뢰 상태 변경 시 webhook 송신)
- 의뢰자 (requester) 가 DevHub 에 직접 로그인 + 자기 의뢰 추적
- 의뢰 첨부파일 / 이미지 업로드
- 의뢰 댓글 / 멘션 / 알림
- 의뢰 SLA / 자동 escalation (assignee 가 N일 이내 미응답 시 system_admin 으로 전환)
- repository 단독 등록 (application 없이 repository 만 만들기) — 1차는 application 또는 project 만, repository 는 application 의 link 로 흡수

## 7. 외부 수신 인증 정책 — [ADR-0012](../adr/0012-dreq-external-intake-auth.md) 결정

**[ADR-0012](../adr/0012-dreq-external-intake-auth.md)** (sprint `claude/work_260515-g`, accepted 2026-05-15) 가 옵션 A — **API 토큰 + IP allowlist** 를 1차 채택. 옵션 B (HMAC) 와 C (OAuth client_credentials) 는 후속 단계 마이그레이션 경로.

요약:
- 외부 호출은 `Authorization: Bearer <plain-token>` 헤더로 도착.
- middleware `requireIntakeToken` 가 `SHA-256(token)` 으로 `dev_request_intake_tokens` lookup + IP allowlist (CIDR) 검증 + revoke 상태 확인.
- 성공 시 `source_system` 컨텍스트 주입 + audit `dev_request.intake_auth_succeeded` + `last_used_at` 갱신.
- 실패 시 401 + audit `dev_request.intake_auth_failed`.
- 토큰은 발급 직후 1회만 admin 에게 노출, 이후 DB 에는 hashed 만 저장.
- 운영 정책: 12개월 회전 / 즉시 revoke 가능 / 30일 미사용 시 운영자 알림.

자세한 결정 근거 / 단계적 마이그레이션 경로 (A → B → C) / 데이터 모델 / audit 카탈로그는 ADR-0012 참조.

## 8. 미해결 항목 (carve out)

- **외부 수신 인증 정책** (§7) — ADR 발급 sprint 분리.
- **DREQ → application/project promotion 시 owner/leader 자동 매핑** — DREQ.assignee 가 application.leader 가 되는가? `requester` 가 외부라 user 매핑이 부재한 경우의 fallback.
- **registered 후 origin_dreq_id 역참조 컬럼**: application/projects 테이블에 nullable FK 추가 여부. ADR 후보.
- **repository 단독 등록** 1차 비허용 결정 — 후속 사용자 피드백에 따라 확장 가능.
- **PMO Manager 의 DREQ 위양 범위** — ADR-0011 §4.2 패턴으로 후속.
- **외부 시스템 callback (의뢰 상태 알림)** — webhook 송신 모델.
- **의뢰 SLA / 자동 escalation** — 운영 정책 결정 후 진입.
- **AI Gardener 와의 연계** — 의뢰 내용 분류, 유사 application 추천, 자동 라우팅. M4 AI track 진입 후.

## 9. 후속 sprint hook

| 후속 sprint | 산출물 | 진입 조건 |
| --- | --- | --- |
| **DREQ-AuthADR** | 외부 수신 인증 ADR (A/B/C 중 선택) | 컨셉 머지 직후 |
| **DREQ-Backend** | `internal/domain/dev_request.go`, `internal/store/dev_requests.go`, `internal/httpapi/dev_requests.go`, migration `000022_dev_requests.up.sql` | DREQ-AuthADR 머지 후 |
| **DREQ-Frontend** | 담당자 dashboard 위젯 + `/admin/settings/dev-requests` 페이지 + 상세 modal + Promote-to-Application/Project 연계 | DREQ-Backend 머지 후 |
| **DREQ-RBAC-ADR** | PMO Manager / 담당자 위양 정책 (ADR-0011 §4.2 패턴) | DREQ-Backend 와 병행 |
| **DREQ-E2E** | UT-dreq + TC-DREQ-* 발급 + Playwright spec | DREQ-Frontend 머지 후 |
| **DREQ-Callback** (carve) | 외부 시스템 webhook 송신 + 상태 알림 | MVP 안정화 후 |

## 10. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-15 | 1차 작성 — DREQ 도메인 컨셉 staged. 후속 sprint hook 6건 명시. ID 미발급 (본 sprint 가 함께 발급할 REQ-FR-DREQ / UC-DREQ / ARCH-DREQ / API-DREQ 는 본 컨셉 머지 PR 의 다른 문서에서 발급). |
| 2026-05-15 | §7 외부 수신 인증 정책 결정 — [ADR-0012](../adr/0012-dreq-external-intake-auth.md) 가 옵션 A (API 토큰 + IP allowlist) 1차 채택. sprint `claude/work_260515-g`. |
