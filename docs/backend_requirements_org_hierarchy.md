# 백엔드 개발 요구사항: 조직 계층 및 인사 관리 시스템

- 문서 목적: 복잡한 조직 계층 구조와 인사 이동 (겸임, 파견) 을 지원하기 위한 백엔드 기능 요구사항을 정의한다.
- 범위: 조직 노드 (Division/Team/Group/Part) 의 hierarchy, 리더십, 멤버십, 파견 / 겸임 / 주 소속 자동 판정, 관련 API 명세.
- 대상 독자: Backend 개발자, AI agent, 프론트엔드 Phase 5+ 작업자.
- 상태: accepted
- 최종 수정일: 2026-05-13 (메타 헤더 표준화, sprint `claude/work_260513-d`)
- 관련 문서: [조직 계층 구조 명세](./organizational_hierarchy_spec.md), [조직도 UX 명세](./org_chart_ux_spec.md), [백엔드 API 계약 §12.8](./backend_api_contract.md), [통합 요구사항 정의서 §5.2](./requirements.md).

프론트엔드 Phase 5 고도화에 따라, 복잡한 조직 계층 구조와 인사 이동(겸임, 파견)을 지원하기 위한 백엔드 기능 개발을 요청합니다.

## 1. 도메인 모델 및 스키마 요구사항

### 1.1 조직 노드 (OrgNode)
- **Hierarchy**: `parent_id` (Self-referencing) 구조를 통해 무제한 계층 및 계층 건너뛰기(Skip-level) 지원.
- **Types**: `division`, `team`, `group`, `part` 열거형 지원.
- **Leadership**: 각 노드는 1명의 `leader_id`를 가짐.

### 1.2 구성원 및 인사 상태 (OrgMember)
- **Secondment (파견)**: `primary_dept_id` (원부서)와 `current_dept_id` (현재 소속) 필드 필요.
- **Dual Appointments (겸임)**: 한 명의 구성원이 여러 부서의 리더 역할을 가질 수 있는 1:N 관계 테이블 필요.

## 2. 주요 로직 요구사항

### 2.1 주 소속 부서 자동 판정 (Primary Dept Logic)
- 구성원의 주 소속 부서를 다음 우선순위에 따라 계산하여 반환해야 함:
    1. 겸임 시: 최상위 계층 부서 (사업부 > 팀 > 그룹 > 파트).
    2. 동급 계층 겸임 시: 하위 조직(자식 노드) 수가 가장 많은 부서.
    3. 파견 시: 파견 부서를 `current_dept`로 반환하되 원부서 정보 포함.

### 2.2 계층 무결성 검증
- 하위 조직 추가/이동 시 상위 조직의 타입보다 낮은 계층인지 검증 (예: 파트 아래에 팀 배치 불가).

## 3. API 엔드포인트 (Proposed)

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/v1/org/hierarchy` | 전체 조직 트리 및 노드별 리더 정보 반환 |
| `GET` | `/api/v1/org/members` | 구성원 리스트 (파견/겸임 상태 포함) |
| `POST` | `/api/v1/org/members/:id/second` | 파견 처리 (원부서 유지, 현재 부서 변경) |
| `POST` | `/api/v1/org/members/:id/appoint` | 리더 임명/겸임 설정 |
| `PATCH` | `/api/v1/org/nodes/:id` | 조직 계층 이동 (부모 변경) |

## 4. 데이터 예시 (JSON)

```json
{
  "member": {
    "id": "u123",
    "name": "홍길동",
    "is_seconded": true,
    "primary_dept": "팀-A",
    "current_dept": "팀-B",
    "appointments": [
      {"dept_id": "팀-B", "role": "leader"},
      {"dept_id": "그룹-C", "role": "leader"}
    ]
  }
}
```
