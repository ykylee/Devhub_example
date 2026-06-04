# Project 상세 대시보드 컨셉 및 콘텐츠 구성 방안 (Concept Design)

- 문서 목적: DevHub의 프로젝트 상세 대시보드(PROJDASH)에서 제공해야 할 3대 페르소나별 핵심 정보 및 UX 구성을 정의하고 고도화한다.
- 범위: 개발자(Developer), 프로젝트 리더(PL), 조직 관리자(Manager)의 3단 뷰 세부 위젯 및 기능 설계.
- 대상 독자: 제품 리드, 프론트엔드 및 백엔드 개발자, QA
- 상태: draft
- 최종 수정일: 2026-06-04
- 관련 문서: [requirements.md](./requirements.md), [project_concept.md](./project_concept.md), [dashboard_concept.md](./dashboard_concept.md)

---

## 1. 개요 및 요구 사유

프로젝트 상세 대시보드(PROJDASH)는 단순 일정 관리를 넘어 프로젝트에 참여하는 구성원들이 실시간 작업 현황을 파악하고 기동성 있게 대응할 수 있는 홈 베이스여야 합니다. 그러나 단일한 관점으로만 대시보드를 구성할 경우, 정보 과부하가 발생하거나 혹은 필요한 핵심 지표가 누락되는 한계가 존재합니다.

이를 해결하기 위해 프로젝트 대시보드를 다음 **3대 페르소나**의 관점으로 명확히 분리하고, 상호 유기적으로 스위칭할 수 있도록 구성합니다.
1. **Developer (개발 실무자):** "오늘 내가 당장 집중해서 개발해야 하는 작업(My Work)과 빌드 상태 확인"
2. **Project Leader (PL, 프로젝트 리더):** "릴리즈 마일스톤 도달을 위한 코드 통합(PR) 병목 해소 및 기술적 장애물 중재"
3. **Org Manager (조직 관리자):** "인적 자원의 업무 부하 균형 상태 점검 및 마감일(SLA) 지연 리스크 관리"

---

## 2. 3대 페르소나별 상세 콘텐츠 및 데이터 매핑

### 2.1 개발자 뷰 (Developer View) — "일일 실무 기동력 극대화"
개발자가 출근해서 퇴근할 때까지 자신의 일일 업무와 코드 상태를 모니터링하기 위한 뷰입니다.
* **My Work Feed (개인화된 작업 요약):**
  * 로그인한 사용자 계정을 기준으로 **나에게 할당된 이슈(Active Tasks)** 목록을 우선순위별로 최상단에 리스트업합니다.
  * **"Review Requests (리뷰 요청)"**: 본인이 리뷰어로 지정된 PR 목록을 보여주어 협업 딜레이를 방어합니다.
* **Review Guard (내 PR 상태 확인):**
  * 본인이 작성한 PR 중 **코드 충돌(Merge Conflict)이 발생한 건**이나 **CI 빌드 실패 건**을 붉은색 경고 표시와 함께 전면 노출합니다.
* **My Code Health & Build Status:**
  * 본인이 담당하는 연결 저장소 브랜치들의 최종 빌드 헬스 정보를 제공합니다.

### 2.2 프로젝트 리더 뷰 (Project Leader View) — "딜리버리 및 통합 관제"
PL이 스프린트 및 마일스톤 달성을 위해 코드 병목과 기술적 차단 요인을 탐지하고 해소하기 위한 뷰입니다.
* **PR Integration Hub (머지 및 통합 병목 제어):**
  * 프로젝트에 연결된 모든 저장소들의 PR 목록 중 **빌드가 실패했거나, 충돌이 났거나, 48시간 이상 승인/리뷰 피드백 없이 방치된(Stale) PR**을 집중적으로 노출합니다.
  * PL이 해당 개발자들에게 리마인드 알림을 보내거나 담당자를 재지정할 수 있는 퀵 액션을 연결합니다.
* **Feature Progress Radar (기능 단위 릴리즈 진행률):**
  * 개별 이슈 레벨을 넘어, 특정 릴리즈 마일스톤이나 에픽(Epic)으로 묶인 **기능 묶음 단위의 완수 지표**를 추적합니다.
* **Escalation & Blocker Feed (장애 요인 감지):**
  * 개발자들이 이슈 상태를 `Blocked`로 변경했거나, 코멘트 내에서 "도움 필요", "의존성 병목" 등의 키워드가 검출된 일감을 수집해 우선 대응할 수 있도록 돕습니다.

### 2.3 조직 관리자 뷰 (Org Manager View) — "자원 배분 및 SLA 리스크 관리"
라인 매니저 및 부서장이 팀원들의 업무 과부하 상태를 감시하고, 품질 거버넌스 준수 및 납기 지연 가능성을 거시적으로 파악하기 위한 뷰입니다.
* **Workload Meter & Resource Balancing (리소스 부하 감지):**
  * 팀 멤버별로 현재 할당된 **오픈 이슈 개수 및 할당된 PR 수**를 게이지 바 형태로 시각화합니다.
  * 특정 팀원에게 태스크가 과도하게 쏠릴 경우(예: 5개 이상) 경고 배지를 띄워 업무 과적을 방지합니다.
* **Delivery Health & Forecast (마감 지연 위험도 예측):**
  * 남은 잔여 이슈량과 최근 팀의 개발 속도(Velocity)를 대조 연산하여 마감일(Due Date)까지 납기 가능성을 예측하는 **지연 리스크 신호등(Green/Yellow/Red)**을 작동시킵니다.
* **Governance Shield & Tech Debt Rollup (품질 및 보안 거버넌스):**
  * 연동된 전 저장소의 SonarQube 정적 분석 데이터(Blocker 버그 수, 보안 취약점 수, 테스트 커버리지율)를 종합 롤업해 프로젝트의 전반적인 거버넌스 스코어를 제공합니다.

---

## 3. UI/UX 구성 및 기술 사양

### 3.1 페르소나 모드 스위처 (3-Way Mode Switcher)
사용자의 역할(Role)에 맞춰 가장 최적화된 초기 뷰를 렌더링하고, 필요 시 전환 가능한 UI 컴포넌트를 상단 헤더에 구현합니다.
* **자동 렌더링 규칙:**
  * Keycloak 토큰의 Resource Access Role 분석:
    * `contributor` -> **Developer View** 기본 활성화
    * `project_leader` / `lead` -> **PL View** 기본 활성화
    * `pmo_manager` / `team_manager` -> **Manager View** 기본 활성화
* **수동 세그먼트 컨트롤:**
  * 상단 우측에 `[ Developer ] | [ Project Leader ] | [ Org Manager ]` 형태의 슬라이딩 토글을 노출하여, 권한을 가진 사용자(예: PL 및 Manager)가 유연하게 화면을 전환할 수 있도록 합니다. 권한이 없는 뷰 선택 시 제한 안내 팝업을 노출합니다.

### 3.2 Premium Design & Interaction (Wow 요소)
* **Glassmorphism Grid Layout:** 반응형 레이아웃이 적용된 글래스모피즘 카드를 배치하여 정돈된 프리미엄 다크 모드 감성을 제공합니다.
* **Red Neon Pulsing & Glowing:** 빌드 실패나 Blocker PR 등 긴급 상황이 발생한 카드 테두리에 은은한 네온 펄스 애니메이션(`animate-pulse` 및 HSL 테마 기반 glow 효과)을 적용하여 PL과 개발자의 기동성 있는 대응을 자극합니다.

---

## 4. 데이터 맵핑 및 흐름도

```mermaid
graph TD
    subgraph SCM Data (Gitea / GitHub)
        A1[Pull Requests]
        A2[Issues / Tasks]
        A3[Commits]
        A4[SonarQube Rating]
    end

    subgraph DevHub Dashboard view
        D1[Developer View]
        D2[Project Leader View]
        D3[Org Manager View]
    end

    A1 -->|CI Build & Conflict status| D1
    A1 -->|Stale Review & Build failures| D2
    A2 -->|Assigned Issues| D1
    A2 -->|Blocked & Technical Help needed| D2
    A2 -->|Load per dev & Velocity| D3
    A4 -->|Rollup Quality Shield| D3

    style D1 fill:#f5f,stroke:#333,stroke-width:1px
    style D2 fill:#5ff,stroke:#333,stroke-width:1px
    style D3 fill:#ff5,stroke:#333,stroke-width:1px
```
