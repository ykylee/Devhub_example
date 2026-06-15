# Unit of Work (UOW) Template

- 문서 목적: project 의 system-level 분해 — "어떤 unit 들로 구성되어 있고, 각 unit 이 무엇을 책임지며, 어떻게 연결되는가?" 를 SSOT 로 관리. v0.7.0 의 Unit of Work 3-layer 도입.
- 범위: unit definitions, dependency matrix, story mapping, code organization strategy
- 대상 독자: 시스템 설계자, AI agent (특히 orchestrator / code-worker / doc-worker), 운영자
- 상태: stable (v0.7.0 도입)
- 최종 수정일: YYYY-MM-DD
- 관련 문서: [`./work_backlog_template.md`](./work_backlog_template.md) (task-level 분해), [`./session_handoff_template.md`](./session_handoff_template.md), [`./project_workflow_profile_template.md`](./project_workflow_profile_template.md), [`../core/workflow_task_modes.md`](../core/workflow_task_modes.md), [`../core/global_workflow_standard.md`](../core/global_workflow_standard.md)

## 1. 왜 Unit of Work (UOW) 가 필요한가

기존 우리 워크플로우의 분해 단위:

- **Mode 6종** (Analysis / Requirements / Design / Planning / Implementation / Refactoring): *horizontal* 분류 (작업 성격)
- **Task 항목** (work_backlog.md 의 TASK-NNN): *operational* 단위 (날짜별 실행)

**빠진 차원**: *system-level* 분해. "이 project 가 어떤 module/unit 들로 구성되어 있고, 각 unit 이 무엇을 책임지며, 어떻게 연결되는가?" 가 SSOT 로 명시되지 않음.

AIDLC 의 `inception/units-generation.md` 가 해결: 시스템 분해 → unit 정의 → 의존성 matrix → user-story mapping. 본 템플릿은 우리 운영 컨텍스트에 적응.

## 2. 운영 원칙

1. **UOW 는 system-level**, work_backlog 는 task-level. 두 layer 분리.
2. **UOW 정의는 architectural decision** — 큰 변경 시만 갱신. 일상 task backlog 와 다르게 *낮은 churn*.
3. **UOW 간 dependency matrix** 명시 — orchestrator 가 sub-agent 위임 시 이 matrix 참조.
4. **UOW 별 story mapping** — user story 가 어느 UOW 에 속하는지 추적.
5. **Greenfield 만 mandatory**, Brownfield 는 optional (기존 코드 구조가 이미 UOW 의 역할).

## 3. UOW 정의 (필수)

### 3.1 UOW 식별자

`UOW-NNN` 형식. 3자리 zero-pad. 예: `UOW-001`, `UOW-002`.

### 3.2 UOW 메타필드

| 필드 | 필수 | 설명 |
|---|---|---|
| `id` | ✅ | `UOW-NNN` |
| `name` | ✅ | unit 의 canonical name (kebab-case). 예: `auth-service`, `user-profile` |
| `type` | ✅ | `service` (independently deployable) / `module` (logical grouping) / `library` (shared) |
| `responsibility` | ✅ | 1-2 line. 이 unit 이 무엇을 책임지는가. |
| `interfaces` | ✅ | 노출 API / RPC / event list |
| `owner` | (선택) | primary owner (human or team) |
| `status` | ✅ | `planned` / `in_progress` / `stable` / `deprecated` |
| `created` | ✅ | YYYY-MM-DD |
| `updated` | ✅ | YYYY-MM-DD |

### 3.3 Format

```markdown
### UOW-001: auth-service

- **Type**: service
- **Responsibility**: 사용자 인증 (login, logout, token refresh)
- **Interfaces**:
  - POST /auth/login
  - POST /auth/logout
  - POST /auth/refresh
- **Status**: in_progress
- **Owner**: backend-team
- **Created**: 2026-06-12
- **Updated**: 2026-06-12
```

## 4. Dependency Matrix (필수)

### 4.1 형식

UOW 간 의존 관계를 matrix 로 명시. `depends_on: [UOW-002, UOW-003]`.

| From \ To | UOW-001 | UOW-002 | UOW-003 |
|---|---|---|---|
| **UOW-001** | - | ❌ | ✅ |
| **UOW-002** | ✅ | - | ❌ |
| **UOW-003** | ❌ | ✅ | - |

✅ = depends on, ❌ = independent.

### 4.2 Format (markdown)

```markdown
## Dependency Matrix

| UOW | depends_on | depended_by |
|---|---|---|
| UOW-001 | [] | [UOW-002, UOW-003] |
| UOW-002 | [UOW-001] | [UOW-003] |
| UOW-003 | [UOW-001, UOW-002] | [] |

## Dependency Graph (Mermaid)

\`\`\`mermaid
graph LR
    UOW-001 --> UOW-002
    UOW-001 --> UOW-003
    UOW-002 --> UOW-003
\`\`\`
```

## 5. Story Mapping (선택, 권장)

User story 가 어느 UOW 에 속하는지 매핑:

```markdown
## Story Mapping

| User Story | UOW | Notes |
|---|---|---|
| "사용자가 로그인한다" | UOW-001 | auth-service |
| "사용자가 프로필을 수정한다" | UOW-002 | user-profile, UOW-001 의 auth 필요 |
| "결제 트랜잭션이 처리된다" | UOW-003 | 결제, UOW-001 + UOW-002 의 user 정보 필요 |
```

## 6. Code Organization Strategy (Greenfield mandatory)

Greenfield project 의 경우, UOW 별 code organization 전략 명시:

```markdown
## Code Organization

**Multi-unit (microservices)**:
\`\`\`
{uow-name}/
├── src/
│   ├── api/
│   ├── domain/
│   └── infra/
├── tests/
├── deploy/
└── README.md
\`\`\`

**Single-unit (monolith)**:
\`\`\`
src/
├── {uow-name}/
│   ├── api/
│   ├── domain/
│   └── infra/
├── tests/
│   └── {uow-name}/
└── README.md
\`\`\`
```

## 7. Template 전체 구조 (한 page 요약)

```markdown
# Unit of Work — {project_name}

- 문서 목적: ...
- 상태: stable
- 최종 수정일: YYYY-MM-DD
- 관련 문서: ...

## 1. UOW 정의

### UOW-001: {name}
... (위 3.3 형식)

### UOW-002: {name}
...

## 2. Dependency Matrix

| UOW | depends_on | depended_by |
|---|---|---|

## 3. Dependency Graph (Mermaid)

\`\`\`mermaid
graph LR
    ...
\`\`\`

## 4. Story Mapping (선택)

| User Story | UOW | Notes |
|---|---|---|

## 5. Code Organization (Greenfield)

\`\`\`
...
\`\`\`

## Changelog

- YYYY-MM-DD: initial (creator)
- YYYY-MM-DD: dependency matrix updated
```

## 8. 우리 운영 적용

- **신규 project**: bootstrap_lib 의 `--adoption-mode new` 가 `unit_of_work.md` 자동 emit
- **Brownfield project**: optional (이미 코드 구조가 UOW 역할). *필요시* UOW 정의 추가.
- **Orchestrator sub-agent 위임 시**: UOW matrix 참조하여 worker role 결정 (auth-service 작업 → code-worker + UOW-001 tag)
- **v0.7.0+ 통합**: workflow_kit.common.contracts.uow (선택) — UOW matrix parsing + sub-agent 위임 결정 helper

## 9. 다음에 읽을 문서

- [`./work_backlog_template.md`](./work_backlog_template.md) — task-level 분해 (UOW 의 operational 단위)
- [`./session_handoff_template.md`](./session_handoff_template.md) — session 시작/종료
- [`./project_workflow_profile_template.md`](./project_workflow_profile_template.md) — project 특화 규칙
- [`../core/workflow_task_modes.md`](../core/workflow_task_modes.md) — mode 6종 (horizontal)
- [`../core/global_workflow_standard.md`](../core/global_workflow_standard.md) §1.2
- AIDLC 원본: `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/inception/units-generation.md` (188 line)
