# 조직 계층 구조 및 데이터 모델 명세

DevHub 플랫폼의 조직 관리 기능을 위한 계층 구조와 포함 관계에 대한 명세입니다.

## 1. 조직 계층 (Hierarchy Levels)

플랫폼은 다음과 같은 표준 계층 순서를 따릅니다:

1.  **사업부 (Division)**: 최상위 조직 단위
2.  **팀 (Team)**: 사업부 하위 조직
3.  **그룹 (Group)**: 팀 하위 조직
4.  **파트 (Part)**: 최하위 조직 단위

## 2. 포함 관계 규칙 (Inclusion Rules)

상위 조직은 하위 조직을 포함할 수 있으며, 다음과 같은 유연성을 가집니다:

- **표준 흐름**: 사업부 ⊃ 팀 ⊃ 그룹 ⊃ 파트
- **유연한 계층 (Skip-level)**: 상위 조직은 중간 단계를 건너뛰고 하위 계층 조직을 직접 포함할 수 있습니다.
    - 예: **팀(Team)** 아래에 **그룹(Group)** 없이 바로 **파트(Part)** 가 직할 조직으로 배치될 수 있음.
- **다대다 관계 방지**: 하나의 하위 조직은 원칙적으로 하나의 상위 조직에만 소속됩니다 (Tree Structure 유지).

## 3. 리더 및 멤버 소속 규칙 (Leadership & Membership)

### 3.1 리더십 (Leadership)
- 모든 조직 단위(사업부, 팀, 그룹, 파트)에는 반드시 1명의 **리더(Leader)**가 존재합니다.

### 3.2 겸임 (Dual Appointments)
- 특수한 경우, 한 명의 구성원이 여러 조직의 리더를 **겸임**할 수 있습니다.
- **주 소속 부서 판단 기준**:
    1. 계층상 가장 높은 단계의 부서를 주 소속으로 처리합니다 (예: 사업부 리더 > 팀 리더).
    2. 계층 단계가 같거나 독립적인 조직들을 겸임하는 경우, **하위 조직(Child Nodes)의 수가 더 많은 조직**을 대표 부서로 표시합니다.

### 3.3 파견 (Secondment)
- 구성원은 특정 부서로 **파견** 처리될 수 있습니다.
- **표시 기준**: 현재 활동 중인 '파견 부서'를 현재 소속으로 표시하되, 시스템은 **'원부서(Original Department)'** 정보를 유지하고 관리해야 합니다.

## 4. 데이터 모델 설계 (Expanded)

```typescript
interface OrgMember {
  id: string;
  name: string;
  primary_dept_id: string; // 원부서 ID
  current_dept_id: string; // 현재 소속 (파견 시 파견 부서 ID)
  is_seconded: boolean;    // 파견 여부
  appointments: {
    dept_id: string;
    role: 'leader' | 'member';
  }[];
}

interface OrgNode {
  id: string;
  name: string;
  type: 'division' | 'team' | 'group' | 'part';
  parent_id: string | null;
  leader_id: string; // 해당 조직의 리더 ID
}
```

## 5. UI/UX 반영 계획 (Updated)

- **Org Tree & Member List**:
    - 리더 아이콘 및 겸임 표시.
    - 파견 인원 별도 표기 (예: '파견 중' 배지 및 원부서 툴팁).
    - 겸임 리더의 경우 대표 부서 노드에 강조 표시.
- **조직 이동 관리**:
    - 파견 복귀 및 겸임 해제 프로세스 UI 제공.
