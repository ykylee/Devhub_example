# ADR-0011: Application/Project Owner 위양과 RBAC row-level scoping (proposed, placeholder)

- 문서 목적: Application/Repository/Project 도메인의 Owner / Member 가 자신이 소속된 리소스에 대해서만 쓰기 권한을 행사하도록 하는 **row-level RBAC scoping** 정책을 결정하기 위한 ADR 의 placeholder. 본 ADR 은 후속 design sprint 의 시작점이며, **현재 상태 = proposed (구현 carve out)**.
- 범위: `applications`, `application_repositories`, `projects`, `project_members` 의 read/write 권한이 (a) 전역 system_admin, (b) 후속 활성화될 `pmo_manager`, (c) 리소스의 owner / member 로 어떻게 분기되는지 결정한다. application/project 외 모듈 (auth/account/org/RBAC policy 자체) 은 본 ADR 의 범위 밖이다.
- 대상 독자: Backend 개발자, RBAC 정책 stakeholder, AI agent, 추적성 리뷰어.
- 상태: proposed (placeholder, decision pending — Design sprint 진입 전)
- 작성일: 2026-05-14
- 결정 근거 sprint: 후속 (`docs/planning/project_management_concept.md` §10 미해결 항목 + REQ-FR-PROJ-009 후속).
- 관련 문서: [`docs/planning/project_management_concept.md`](../planning/project_management_concept.md) §5.3, §10, [`docs/requirements.md`](../requirements.md) §5.4.1 (REQ-FR-PROJ-000, REQ-FR-PROJ-009, REQ-FR-PROJ-010), [ADR-0002](./0002-rbac-policy-edit-api.md), [ADR-0007](./0007-rbac-cache-multi-instance.md), [추적성 매트릭스 §2.2 RBAC API + §3 Application/Project 행](../traceability/report.md).

## 1. 컨텍스트

`docs/planning/project_management_concept.md` §5.3 는 Owner 권한 위양을 3단계로 정의한다:

- **1차 (MVP)**: 모든 쓰기 작업은 `system_admin` 단일 주체.
- **2차**: `owner` 에게 제한된 메타/멤버 수정 위양.
- **3차**: RBAC row-scoping — 부서/리소스 단위 위양 확장.

3차 단계의 enforcement 가 본 ADR 의 결정 대상이다. 현재 RBAC 모델 (ADR-0002, ADR-0007) 은 (role, resource, action) 의 4축 boolean 매트릭스이며, **row-level 조건 (= 이 row 의 owner_user_id 가 caller 인가?) 은 없다**. 즉 현재 매트릭스 그대로는 owner-self 만 PATCH 가능 같은 표현이 불가능하다.

REQ-FR-PROJ-009 와 REQ-FR-PROJ-010 (pmo_manager 활성 후 권한 범위) 가 본 ADR 의 결과를 의존하고 있어, 결정 전까지는 `system_admin` 일임 정책 (REQ-FR-PROJ-000) 이 강제된다.

## 2. 결정 동인 (decision drivers, draft)

- **deny-by-default 유지**: 모든 row-level scoping 결정은 explicit allow 만 추가하고 deny 는 그대로.
- **현행 RBAC 매트릭스와의 호환성**: 4축 boolean 의 확장이 아니라 별도 row-predicate 레이어로 분리할지, 매트릭스에 row 조건 컬럼을 추가할지.
- **enforcement 위치**: handler / service / store 중 어느 계층에서 row 조건을 검증할지. SQL `WHERE owner_user_id=?` 형태로 store 에 push down 할지, application code 에서 post-filter 할지.
- **감사 (audit)**: row-level 거부도 deny 이벤트로 audit log 에 남길지.

## 3. 검토할 옵션 (carve out — design sprint 에서 채울 자리)

| 옵션 | 요약 | 장점 | 단점 |
| --- | --- | --- | --- |
| A. Casbin 식 ABAC | 별도 정책 엔진 도입, row 조건을 정책 표현으로 기술 | 표현력 ↑ | 학습 비용, 또 하나의 SoT |
| B. RBAC 매트릭스 확장 | 매트릭스 row 에 `row_predicate` 컬럼 추가 (예: `owner=$caller`) | 단일 SoT 유지 | 표현력 제한, 매트릭스 복잡도 ↑ |
| C. handler/service 코드 검증 | row 조건은 코드로, RBAC 는 4축만 | 단순함 | 정책 분산, 누락 위험 |
| D. SQL row-level security (PG RLS) | PostgreSQL native RLS | 효율적, DB 가 강제 | 정책 외부 (DB) 로 빠짐, audit 통합 어려움 |

본 ADR 의 후속 sprint 에서 각 옵션의 비용/리스크/마이그레이션 영향을 평가하고 단일 옵션을 선택한다.

## 4. 결정

**(placeholder — decision pending)**.

본 ADR 은 forward reference 의 anchor 로 발급된 placeholder 이며, design sprint 진입 후 정식 결정 (`status: accepted`) 으로 갱신한다. 그때까지 본 문서를 참조하는 모든 코드/문서는 아래 잠정 규칙을 따른다:

- **잠정 규칙 (MVP 진입 시)**: REQ-FR-PROJ-000 의 `system_admin` 일임 + `pmo_manager` `disabled` 상태가 그대로 강제된다.
- Owner 위양 코드 경로는 만들지 않는다 (RBAC 4축 매트릭스로 모두 차단).

## 5. 결과

(decision 이후 갱신)

## 6. 후속 작업

- design sprint 에서 옵션 A~D 의 비용/리스크 평가 + 단일 옵션 선택.
- REQ-FR-PROJ-009, REQ-FR-PROJ-010 의 활성화 조건을 본 ADR 결정으로 묶기.
- 매트릭스 §2.2 RBAC 인덱스 + §3 Application/Project row 에 본 ADR 의 IMPL 영향 반영.

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-14 | placeholder 초안 (proposed) — concept.md 가 4회 forward reference 하던 ADR-0011 의 anchor. 결정 carve out. | PR #104 보강 commit |
