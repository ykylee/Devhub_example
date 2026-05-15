# DevHub 요구사항 정의서 (Draft)

- 문서 목적: 팀 통합 개발 허브 (DevHub) 의 역할별 상세 기능 (Developer / Manager / System Admin) 과 데이터 구조, 비기능 요구사항을 정의한다.
- 범위: §2 역할별 기능 / §3 역할 확장성 / §4 데이터 보존 정책 / §5 비기능 + 운영 / §6 기술 스택 / §7 UX. 백엔드 측 상세는 `docs/backend/requirements.md`, 프론트 연동은 `docs/backend/frontend_integration_requirements.md`.
- 대상 독자: 프로젝트 리드, Backend / 프론트엔드 개발자, AI agent, QA, UX 검토자.
- 상태: accepted
- 작성일: 2026-04-28
- 최종 수정일: 2026-05-13 (메타 헤더 표준화, sprint `claude/work_260513-d`)
- 관련 문서: [통합 개발 로드맵](./development_roadmap.md), [아키텍처](./architecture.md), [기술 스택](./tech_stack.md), [백엔드 API 계약](./backend_api_contract.md), [추적성 매트릭스](./traceability/report.md), [거버넌스 — 문서 표준](./governance/document-standards.md).

## 1. 개요
본 문서는 팀 통합 개발 허브(DevHub)의 역할별 상세 기능과 데이터 구조를 정의하기 위해 작성되었습니다.

### 1.1 요구사항 범위 구분

본 문서의 기능 항목은 다음 상태로 구분합니다.

- **확정:** 현재 제품 방향과 초기 구현 기준에 포함되는 요구사항.
- **후보:** 사용자 가치가 있으나 세부 정책, 우선순위, 구현 범위 추가 검토가 필요한 항목.
- **MVP 이후:** 초기 버전 이후 단계적으로 도입할 확장 요구사항.

역할별 요구사항의 체크박스는 기능 후보를 추적하기 위한 목록이며, `핵심 기획 아젠다`의 결정 사항은 제품 방향과 정책 기준으로 우선 적용합니다. 단, 구현 범위는 각 항목의 `확정`, `후보`, `MVP 이후` 상태에 따라 별도로 관리합니다.

## 2. 사용자 역할별 요구사항 (진입 우선순위 기반)

### 2.1 개발자 (Developer)
- **핵심 니즈:** 정보 탐색 최소화, 개발 몰입도 향상.
- **기본 진입 우선순위:** 개발 대시보드 (Developer Dashboard)
- **주요 기능 (후보):**
    - [ ] 기술 스택별 가이드 및 Wiki 통합 검색.
    - [ ] 프로젝트별 환경 설정(Environment Setup) 원클릭 확인.
    - [ ] 팀 내 공통 라이브러리/컴포넌트 카탈로그.
    - [ ] (기타 제안) CI/CD 빌드 결과 및 실시간 에러 로그 요약.

### 2.2 관리자 (Manager)
- **핵심 니즈:** 프로젝트 가시성 확보, 리스크 선제 대응.
- **기본 진입 우선순위:** 관리 대시보드 (Management Dashboard)
- **주요 기능 (후보):**
    - [ ] 마일스톤별 진행률 시각화 대시보드.
    - [ ] 팀원별 작업량(Load) 및 할당 현황.
    - [ ] 일정 지연 및 차단(Blocked) 과제 알림.
    - [ ] (기타 제안) 투입 공수(Man-month) 및 예산 추정치 비교.

### 2.3 확장 역할: 테스트 담당자 (QA)
- **핵심 니즈:** 품질 지표 관리, 결함 추적 효율화.
- **주요 기능 (후보):**
    - [ ] 버전별 테스트 커버리지 및 패스율 리포트.
    - [ ] 결함(Bug) 수정 현황 및 재테스트 대기 목록.

### 2.4 시스템 관리자 (System Administrator)
- **핵심 니즈:** Gitea 연동 인프라와 DevHub 운영 설정을 안전하게 관리.
- **기본 진입 우선순위:** 시스템 대시보드 + 시스템 설정 메뉴
- **노출 정책 (확정):** 시스템 대시보드/시스템 설정은 `system_admin` 권한 보유자에게만 노출한다.
- **주요 기능 (확정):**
    - [x] Gitea 서버 및 Runner 상태 모니터링.
    - [x] Runner 재시작/설정 등 제한된 인프라 제어.
    - [x] Gitea 계정, 조직, 권한 연동 관리.
    - [x] DevHub 자체 사용자 계정(Account) 발급/회수 및 비밀번호 정책 관리 — `2.5 사용자 계정 관리` 참조.
    - [x] 백업 상태와 시스템 알림 임계치 확인.
    - [x] 시스템 관리 작업에 대한 Audit Log 조회.

### 2.5 사용자 계정 관리 (User Account Management)

DevHub 사용자(person)와 인증 자격(credential)을 분리해 관리한다. Gitea SSO 도입 전까지는 DevHub 자체 로컬 계정(Account)을 1차 인증 수단으로 사용한다.

> **구현 방식 (2026-05-07, [ADR-0001](./adr/0001-idp-selection.md))**: 본 절의 정책 invariant(1:1 매핑, 비밀번호 평문 미보관, 자동 lock, audit log 대상) 는 그대로 유지하되, 구현은 **자체 `accounts` 테이블이 아닌 Ory Hydra + Kratos** 가 책임진다. 신규 요구 — DevHub 의 계정 서비스를 다른 앱에도 OIDC IdP 로 제공 — 를 충족하기 위한 결정이다. 정책 변경 없음.

- **핵심 니즈:** 식별 가능한 사람 단위 권한 관리, 분실/유출 시 빠른 회수, 감사 가능한 비밀번호 변경 기록.
- **용어 분리:**
    - **사용자(User):** 조직에 소속된 사람. 표시명, 이메일, 직책, 소속 조직 단위(Org Unit), 역할(Role)로 구성.
    - **계정(Account):** 사용자가 DevHub에 로그인하기 위한 인증 자격. 로그인 ID와 비밀번호 해시로 구성.
- **주요 기능 (확정):**
    - [x] **1:1 매핑 강제:** 사용자 1명은 정확히 0개 또는 1개의 계정을 가질 수 있다. 계정은 반드시 1명의 사용자에 귀속된다. 동일 사용자에 복수 계정을 만들 수 없으며, 1개의 계정이 복수 사용자를 대표할 수 없다.
    - [x] **계정 생성·회수 권한:** 시스템 관리자만 사용자에게 계정을 발급하거나 회수할 수 있다. 회수된 계정으로는 즉시 로그인할 수 없으며, 활성 세션은 무효화된다.
    - [x] **로그인 ID 정책:** 로그인 ID는 시스템 전역에서 unique 하다. 형식 정책(허용 문자, 길이)은 별도 정책 표로 관리한다. 사용자 식별자(`user_id`)와는 별도이며, 로그인 ID 변경이 사용자 식별자를 바꾸지 않는다.
    - [x] **비밀번호 정책:** 비밀번호는 평문으로 저장하지 않는다. 단방향 해시(예: bcrypt cost ≥ 12 또는 argon2id) 만 저장한다. 본인은 현재 비밀번호 확인 후 비밀번호를 변경할 수 있고, 시스템 관리자는 임시 비밀번호로 강제 재설정할 수 있다. 강제 재설정된 계정은 다음 로그인 시 비밀번호 변경을 요구한다.
    - [x] **계정 상태(Account Status):** `active`, `disabled`, `locked`, `password_reset_required` 의 4종 상태를 가진다. `locked`는 정책상 자동 잠금(예: 연속 실패) 또는 수동 잠금이며, `disabled`는 시스템 관리자 회수에 해당한다.
    - [x] **감사 로그 대상:** 계정 생성, 계정 비활성화, 비밀번호 변경(본인/관리자), 잠금 해제, 로그인 성공/실패는 Audit Log 기록 대상이다. 비밀번호 자체나 해시는 Audit Log에 기록하지 않는다.
- **주요 기능 (후보):**
    - [ ] 비밀번호 만료 정책(주기 강제 변경) — 운영 단계에서 정책 결정.
    - [ ] 다단계 인증(MFA/2FA) — 초기 단계는 미도입 (단계 5.3과 동일 기준).
    - [ ] Gitea SSO 연동으로 자체 계정 발급 없이 통합 인증 — `architecture.md` 6.2 RBAC 단계화에서 후속 phase로 관리.
- **데이터 주권 메모:** 사용자 자신은 자신의 계정 정보(로그인 ID, 비밀번호)를 변경할 수 있으나, 계정 발급/회수와 강제 재설정은 시스템 관리자 권한이다.

## 3. 공통 기능 및 시스템 요구사항
- **역할 확장성:** 새로운 역할이 추가되더라도 역할별 기본 진입 우선순위를 설정해 UX를 간접 제공하고, 메뉴/기능은 권한 기반으로 확장할 수 있어야 함.
- **데이터 일관성:** 개발자가 업데이트한 진행 상황이 관리자 대시보드에 즉시 반영되어야 함.

## 4. 공통 운영 원칙 (Common Operating Principles)

두 역할군(개발자, 관리자) 간의 뷰 공존과 데이터 신뢰를 위해 다음 원칙을 준수합니다.

1. **데이터 주권과 정합성의 조화:** 
    - 공식 업무 로그(PR, 이슈, 빌드 등)는 데이터 무결성을 유지하며 관리자 KPI의 기반이 됨.
    - 개인적 회고, 업무 설명, 개인 로드맵 영역에서는 개발자에게 완전한 수정/삭제 권한을 부여함. 
    - 보고용 데이터는 특정 시점의 **스냅샷(Snapshot)**을 기반으로 운영하여 데이터 변경으로 인한 혼선을 방지함.
2. **기술 태깅 기반의 전문가 맵:**
    - '스텔스 전문가'라는 용어 대신 **'기술 태깅(Tech Tagging)'** 개념을 사용함.
    - 시스템이 인식한 전문성은 개발자에게 긍정적인 피드백(Kudos)과 함께 투명하게 공유하여 자발적 정보 제공 동기를 부여함.
3. **AI 가드너 기반의 알림 중재 (v2 예정):**
    - 모든 강제 알림은 **AI 가드너**가 사용자의 현재 업무 몰입 상태를 판단하여 전달 시점을 중재함.
    - 관리자의 긴급 요청이라도 개발자의 집중 시간을 보호하는 것을 원칙으로 함.
4. **팀 성취 중심의 서비스 톤 정합:**
    - 게이미피케이션(Fun) 요소와 공식 거버넌스(Professional) 로그를 **'팀 마일스톤/업적'**으로 통합함.
    - 관리자의 의사결정이 개발자 뷰에서는 '팀 퀘스트 완료'나 '팀 성취'로 시각화되도록 설계함.

### 4.1 데이터 및 권한 운영 기준

데이터 주권, 조회 권한, 보존 정책, 알림 정책은 다음 기준을 우선 적용합니다.

| 데이터 분류 | 원천 | 수정 권한 | 조회 권한 | 보존 기준 | 알림 기준 |
| --- | --- | --- | --- | --- | --- |
| 공식 업무 로그 | Gitea | Gitea 원천 기준, DevHub 직접 수정 불가 | 역할/프로젝트 권한 기준 | 운영 로그 1개월 | 상태 변경/지연/실패 시 역할별 알림 |
| 개인 업무 연혁 | DevHub | 본인 수정/삭제 가능 | 본인 기본, 공유 선택 시 승인된 대상 | 계정 활성 중 유지, 삭제 후 1개월 | 사용자가 선택한 경우만 공유 제안 |
| 보고 스냅샷 | DevHub | 생성 후 직접 수정 불가, 새 스냅샷으로 대체 | 관리자/승인된 리더 | 정책 기간 별도 정의 필요 | 보고 생성/변경 시 관리자 알림 |
| 기술 태깅/Kudos | DevHub | 시스템 추천 + 사용자 확인/관리자 승인 | PL/GL 이상 등 정책 기준 | 계정 활성 중 유지, 삭제 후 1개월 | 긍정 피드백 중심, 강제 알림 최소화 |
| 시스템 관리 로그 | DevHub/Gitea | 시스템 자동 기록, 직접 수정 불가 | 시스템 관리자 | 운영 로그 1개월 이상 검토 필요 | 보안/운영 이벤트는 즉시 알림 |
| 사용자 계정/자격 | DevHub | 본인은 자기 비밀번호 변경, 계정 발급/회수/강제 재설정은 시스템 관리자 | 본인(자기 자신), 시스템 관리자(전체) | 계정 활성 중 유지, 회수 후 90일 보존 후 삭제 | 비밀번호 변경/잠금/회수 시 본인 + 시스템 관리자 알림 |

## 5. 핵심 기획 아젠다 (Discussion Topics)

각 항목별로 심도 있는 논의를 거쳐 확정된 내용을 기록합니다.

### 5.1 [Agenda 1] 개발자용 '도움이 되는 정보'의 우선순위 및 세부 항목
- **논의 배경:** 개발자가 정보 탐색에 소비하는 시간을 줄이고 개발에만 몰입할 수 있는 환경 구축.
- **결정 사항:**
    1. **[최우선] 실시간 개발/운영 현황 (Dynamic):** 배포 버전, CI/CD 상태, 실시간 에러 로그 요약 등 상시 확인 정보 우선 배치.
    2. **개인화된 로드맵 및 현황 (Personalized):** '나의 마일스톤', '나의 잔여 작업량' 시각화.
    3. **기술 태깅 기반 전문가 맵 및 Kudos:** 데이터 가시성은 **파트장급(PL) 이상**으로 제한. 동료 간 피드백(Kudos) 가중치는 **그룹리더(GL) 이상**만 조회. 전문가 데이터를 기반으로 리뷰어 등을 **소극적으로 추천**하는 기능 포함.
    4. **내부 기반 상태 및 부재 관리:** DevHub 내 수동 상태 관리. 협업 타이밍 판단을 위한 최소 정보 공유 (업무 연혁과 분리 운영).
    5. **위키 중심 AI 가드너 (Passive AI, v2 예정):** 
        - 데이터 수집 범위를 **위키 문서**로 제한하여 정보의 정제성 확보.
        - UI 구석에 체크박스와 코멘트 형태로 **소극적인 공유 제안**을 수행하여 개발자 피로도 최소화.
        - 가드너(사람)는 AI 분석 결과를 최종 승인/분류하는 역할 수행.
    6. **민감 자산의 선택적 공유:** 트러블슈팅, 의사결정 히스토리 등은 **개인/과제별로 공유 여부를 선택(Opt-in)** 가능하도록 설계.
    7. **자율적 개인 업무 연혁:** 
        - 업무 이력을 연대기 형식으로 정리하되, **개발자에게 완전한 수정/삭제 권한** 부여.
        - 데이터 무결성보다 자기 어필 및 커리어 기록으로서의 사용자 경험에 집중.
    8. **원클릭 권한 신청 시스템:** 저장소 및 프로젝트별 조회 권한 등을 DevHub 내에서 간편하게 신청하고 승인받는 셀프 서비스 체계 구축.
    9. **팀 빌딩 이스터에그 (Gamification):** 팀 컨벤션 퀴즈나 숨겨진 업적 시스템을 이스터에그 형식으로 도입하여 재미와 팀 결속력 강화.
- **핵심 원칙:** "효율적인 업무 지원(권한, 정보)과 위트 있는 상호작용(이스터에그, Kudos)을 결합하여, 개발자가 자발적으로 머무르고 싶어 하는 팀의 광장을 구축함."

### 5.2 [Agenda 2] 관리자용 '현황 파악'의 시각화 범위와 깊이
- **논의 배경:** 단순 상태 조회를 넘어 실질적인 리스크 관리와 의사결정에 도움이 되는 데이터 정의.
- **결정 사항:**
    1. **계층형 가시성 및 조직 경계 정책:**
        - **초기:** 조직 간 데이터 경계를 두는 **Strict Boundary** 적용.
        - **향후:** 협업 체계 안착 시 조직 간 비교 및 콜라보레이션이 가능하도록 설계.
    2. **보고 자동화 및 KPI 관리:**
        - **주간 보고서 자동 생성** 및 선택적 외부 발송.
        - 조직별 목표 KPI 설정 및 리더 관심도에 따른 **카테고리 우선순위 조절**.
    3. **AI 협업 리소스 시뮬레이션 (What-if):**
        - 리소스 조정에 따른 일정 변화를 AI 에이전트와 의견 교환하며 검토 (참고용 수치 제공).
    4. **리스크 관리 및 공식 전파:**
        - 관리자가 식별한 리스크를 **조직/과제별 공지사항**으로 즉시 전파하여 가시성 확보.
    5. **의사결정 및 협의 거버넌스:**
        - 주요 계획 변경 및 리더 간 리소스/우선순위 충돌 발생 시 **공식 협의 및 의사결정 로그**를 남겨 추적성 확보.
    6. **복합 의존성 관리:**
        - 내부 과제 간 의존성은 자동 시각화, **외부 부서(인프라, 보안 등)와의 접점은 수동 등록** 및 일정 영향도 관리.
    7. **알림 계층화 (Notification Grading):**
        - 중요도에 따라 **단순 공유(Info)**와 **즉각 조치 필요(Action Required)** 단계로 알림을 차등화하여 피로도 감소.
    8. **구조화된 개인화 UI:** 표준 레이아웃 기반 위에서 우선순위 및 즐겨찾기 조정 지원.
    9. **데이터 무결성 피드백:** 소스 간 불일치 시 시스템 워닝 및 멘션 기반 피드백 루프 가동.
- **제외/후순위:** 워룸(War-room) 모드는 현재 조직 성격상 불필요하므로 검토 대상에서 제외.
- **핵심 원칙:** "데이터 시각화를 넘어 리더 간의 협의를 돕고, 조직 내외의 복잡한 의존성을 관리하여 의사결정의 투명성과 속도를 높임."




### 5.3 [Agenda 3] 데이터 원천(Source) 및 연동 전략
- **논의 배경:** Gitea를 중심으로 한 데이터 통합 및 자동화 수준, 예외 처리 정책 구체화.
- **결정 사항:**
    1. **데이터 기반 리스크 판단 (관리자 뷰):**
        - **판단 기준:** 이슈 아이템의 생존 시간(Duration)이 **7일**을 초과하거나, 관련 Commit/PR 액티비티 추이가 급감하는 경우를 리스크로 자동 식별.
        - **시각화 (v2):** AI 가드너가 위 지표를 분석하여 관리자 대시보드에 '주의/위험' 신호로 표시.
    2. **집중 시간 보호 및 알림 제어 (개발자 뷰):**
        - **기능:** 알림 On/Off 스위치, 특정 시간대 자동 차단(Schedule) 기능 제공.
        - **예외:** 'Emergency' 태그가 붙은 긴급 이슈는 집중 시간 설정과 무관하게 전달 가능하도록 설계.
    3. **데이터 상충 및 동기화 전략 (하이브리드):**
        - **상충 해결:** Gitea 상태와 DevHub 기록이 불일치할 경우 우선 개발자에게 알림. 불일치가 일정 시간 지속되면 **PL(Part Leader) 뷰**에 해당 항목을 노출하되, PL은 **조회(Read-only)**만 가능하도록 하여 데이터 주권을 보호함.
        - **동기화 방식:** Gitea의 **모든 Webhook 이벤트**를 수집하여 실시간 반영하는 것을 기본으로 하며, **시간 단위(Hourly)**로 전체 상태를 체크하여 누락된 데이터를 Pull 하는 하이브리드 방식 채택.
    4. **CI/CD 및 인프라 연동:**
        - **Gitea Actions:** Gitea Actions를 이용한 CI/CD 파이프라인 상태를 실시간 연동함. 
        - **상세 지표:** 단순히 성공/실패 여부뿐만 아니라 **테스트 커버리지, 소스코드 품질 현황(Static Analysis), 빌드 로그 요약**을 추출하여 각 역할군 뷰에 제공함.
    5. **데이터 수집 및 보존 정책 (Retention):**
        - **이벤트 우선순위:** 모든 Webhook 이벤트를 수집하되, **이슈(Issue) 및 PR 액티비티**를 최우선 순위로 처리하여 실시간 가시성을 확보함.
        - **로그 데이터:** DevHub 내의 일반 운영 로그는 **1개월** 동안 보존함.
        - **개인화 데이터 (Kudos 등):** 해당 계정이 활성 상태인 동안 지속적으로 유지함. 계정 삭제 시 즉시 **아카이브** 처리하며, 아카이브 후 **1개월**이 지나면 영구 삭제함.
    6. **인프라 및 시스템 관리 (System Administrator):**
        - **권한 격리:** 인프라 모니터링, Runner 제어, 시스템 설정 등 핵심 관리 기능은 **시스템 관리자(System Admin) 역할에만 배타적으로 부여**함.
        - **인프라 제어:** 알림 임계치 설정, Runner 직접 제어(재시작/설정), 백업 상태 실시간 확인 기능을 제공함.
        - **계정 및 조직 관리:** Gitea와 연동된 계정, 사용자 권한, 조직 구성을 통합 관리하며, 향후 **Gitea와 SSO(Single Sign-On) 연동**을 통해 통합 인증 체계를 구축함 (초기 단계에서 2FA는 미도입).
    7. **Application/Repository/Project 관리 (System Admin):**
        - **Application/Repository 생애주기 자동화:** 신규 Application 등록 이후 Repository 실행 환경을 다음 순서로 자동화함.
            1. Gitea 내 신규 저장소 자동 생성.
            2. 기본 브랜치 보호 정책(Branch Protection) 자동 적용.
            3. 저장소 멤버 자동 초대 및 권한 설정.
        - **Project 운영 갱신:** Repository 하위 기간성 Project 생성/보관 정책과 연동하여 운영 주기(연간/반기/분기) 갱신을 지원함.
        - **연결 매핑 관리:** Application-Repository, Repository-Project 연결 상태를 관리하며, 추가 자동화 시나리오는 운영 단계에서 지속 고도화함.
    8. **내부 전용 데이터 관리 (Internal Only):**
        - 팀 문화, Kudos(동료 칭찬), 이스터에그 등 Gitea에 기록하기 부적절하거나 팀 내 폐쇄성이 필요한 정보는 **DevHub 자체 DB**에만 보관.

### 5.4 [Agenda 4] Project 운영 모델 고도화 (Requirements Phase)
- **논의 배경:** 제품 수명 단위의 상위 관리와 연간 운영 단위의 행정 갱신을 함께 지원해야 한다. 이를 위해 `Application > Repository > Project > GitHub 실행 단위` 계층을 요구사항으로 확정한다.
- **용어 정책 (확정):**
    - **Application:** 제품 수명 주기와 함께 가는 최상위 총괄 단위.
    - **Repository:** Application 하위 실행 단위 (1 Application : N Repository).
    - **Project:** Repository 하위 기간성 운영 단위 (1 Repository : N Project, 연간/반기/분기 등 정책 기반 갱신 가능).
    - **Execution Artifact:** Repository 내부의 Issue / Milestone / Project Board / Wiki(또는 Docs).

#### 5.4.1 기능 요구사항 (REQ-FR)

- **REQ-FR-PROJ-000 (MVP, 확정):** `Application > Repository > Project` 관리 쓰기 권한은 기본적으로 `system_admin`에 한정해야 한다.
    - 대상 기능: Application 생성/수정/보관, Repository 연결/해제, Project 생성/수정/보관, Project 멤버/owner 관리, Integration 정책 변경, 마일스톤 매핑 관리.
    - 예외 역할: `pmo_manager`는 후보 role로 정의할 수 있으나 정책 확정 전까지 `disabled` 상태로 유지한다.
    - `pmo_manager` 활성화 전 요청은 `403 role_not_enabled`로 거절한다.
- **REQ-FR-APP-001 (MVP, 확정):** 시스템 관리자는 Application을 생성/수정/보관(archive)할 수 있어야 한다.
    - 필수 필드: `key`, `name`, `owner`, `start_date`, `due_date`, `visibility`, `status`.
    - `status` 최소 상태: `planning`, `active`, `on_hold`, `closed`, `archived`.
- **REQ-FR-APP-002 (MVP, 확정):** 하나의 Application은 0개 이상의 Repository를 연결할 수 있어야 한다.
    - 연결 단위 필드: `repo_provider`, `repo_full_name`, `role(primary|sub|shared)`.
    - 동일 Application 내 서로 다른 provider Repository를 동시에 연결할 수 있어야 한다 (예: bitbucket + gitea 병행).
- **REQ-FR-APP-003 (MVP, 확정):** `Application.key`는 시스템 전역 고유값(unique)이어야 하며 관리 식별자로 사용해야 한다.
    - 표시명(`name`) 변경과 무관하게 `key`는 안정 식별자로 유지한다.
    - **변경 불가(immutable):** 발급 후 `key`는 변경할 수 없다. rename 이 필요하면 신규 Application 생성 + 기존 Application archive 절차로만 수행한다 (PATCH 응답 422 `application_key_immutable`).
    - 현재 입력 정책: 영문숫자 조합 10자 (`^[A-Za-z0-9]{10}$`).
    - 데이터베이스 컬럼은 정책 변경 여지를 위해 더 긴 길이(예: VARCHAR(32) 이상)를 허용하고, 실제 길이 제한은 애플리케이션 검증 정책으로 강제한다.
- **REQ-FR-APP-004 (MVP, 확정):** Repository는 외부 형상관리 도구와 연결되는 구조여야 하며, DevHub는 운영/분석용 관리 데이터를 보유해야 한다.
    - 외부 SoT: 코드/PR/빌드 원본.
    - DevHub 보유: 연결 메타데이터, 동기화 상태, 운영 스냅샷.
    - 지원 정책: 특정 SCM 단일 종속이 아니라 provider 추상화(`repo_provider`)를 사용하며, `bitbucket`, `gitea`, `forgejo` 등 복수 provider를 동등하게 지원/확장할 수 있어야 한다.
- **REQ-FR-APP-005 (MVP, 확정):** Repository 작업현황을 수집/조회할 수 있어야 한다.
    - 최소 지표: commit 활동량, active contributor 수, 작업 추이.
- **REQ-FR-APP-006 (MVP, 확정):** PR/PR Activity 정보를 수집/조회할 수 있어야 한다.
    - 최소 정보: PR 상태(open/draft/merged/closed), 생성/리뷰/코멘트/머지 이벤트 타임라인.
- **REQ-FR-APP-007 (MVP, 확정):** 빌드 정보를 수집/조회할 수 있어야 한다.
    - 최소 정보: run status, duration, branch/commit, 시작/종료 시각.
- **REQ-FR-APP-008 (MVP, 확정):** 소스코드 품질 지표(정적분석/스코어링)를 수집/조회할 수 있어야 한다.
    - 최소 정보: tool, quality score, gate pass/fail, metric 상세(coverage, bug/vuln, duplication 등).
- **REQ-FR-APP-009 (MVP, 확정):** 형상관리 도구 연동은 provider별 어댑터 구조를 따라야 한다.
    - 공통 도메인 계약(Repository/PR/Build/Quality 이벤트/스냅샷)과 provider 전용 구현을 분리한다.
    - 신규 provider 추가 시 기존 도메인 API/화면 계약을 깨지 않고 어댑터 추가만으로 확장 가능해야 한다.
    - provider별 인증/웹훅 검증/속도제한/에러 포맷 차이는 어댑터 내부에서 흡수한다.
- **REQ-FR-APP-010 (MVP, 확정):** Application 상태 전이는 정의된 상태 머신 규칙을 따라야 한다.
    - 상태 집합: `planning`, `active`, `on_hold`, `closed`, `archived`.
    - `archived`는 기본적으로 종료 상태이며 일반 상태 전이로 복구하지 않는다.
    - 상태 전이 권한: 기본적으로 `system_admin`만 허용한다 (`pmo_manager` 활성 전 `403 role_not_enabled`).
    - 전이 검증 가드:
      - `planning -> active`: 연결된 활성 Repository 1개 이상 필요.
      - `active -> closed`: `severity=critical` 롤업 경고 0건 + 연결 Repository 1개 이상 필요.
      - `on_hold -> active`: `due_date` 만료 시 재개 사유(`resume_reason`) 기록 필요.
      - `* -> archived`: soft-delete 처리, `archived_reason` 기록 필요.
- **REQ-FR-APP-011 (MVP, 확정):** Application-Repository 연결은 라이프사이클 상태를 가져야 한다.
    - 최소 상태: `requested`, `verifying`, `active`, `degraded`, `disconnected`.
    - 연결 검증 실패/일시 장애 시 `sync_error_code`를 기록해야 한다.
    - `sync_error_code`는 표준 코드 사전을 사용해야 하며(`provider_unreachable`, `auth_invalid`, `permission_denied`, `rate_limited`, `webhook_signature_invalid`, `payload_schema_mismatch`, `resource_not_found`, `internal_adapter_error`), 임의 문자열 사용을 금지한다.
    - `sync_error_code`에는 재시도 가능 여부(`retryable`)와 최근 발생 시각이 함께 관리되어야 한다.
- **REQ-FR-APP-012 (MVP, 확정):** Application 롤업은 누락/장애 데이터를 숨기지 않고 `data_gap` 또는 경고 상태로 표시해야 한다.
    - 최소 롤업 대상: PR 분포, 빌드 성공률/평균 시간, 품질 점수, gate 실패 건수.
    - 기본 `weight_policy`는 `equal`(동일 가중)이다.
    - 선택 `weight_policy`는 `repo_role`(primary/sub/shared 가중), `custom`(관리자 정의)를 지원할 수 있어야 한다.
    - `custom` 정책은 가중치 합이 1.0(±허용오차)이어야 하며, 음수 가중치는 허용하지 않는다.
- **REQ-FR-PROJ-001 (MVP, 확정):** 시스템 관리자는 Repository 하위 Project를 생성/수정/보관(archive)할 수 있어야 한다.
    - 필수 필드: `key`, `name`, `owner`, `start_date`, `due_date`, `visibility`, `status`.
    - `status` 최소 상태: `planning`, `active`, `on_hold`, `closed`, `archived`.
- **REQ-FR-PROJ-002 (MVP, 확정):** 일반 사용자는 자신이 멤버인 Project 및 공개 Project를 조회할 수 있어야 한다.
    - `archived` Project는 기본 숨김이며, 명시적 토글로 노출한다.
- **REQ-FR-PROJ-003 (MVP, 확정):** 시스템 관리자는 Project별 멤버/책임자(owner)를 관리할 수 있어야 한다.
- **REQ-FR-PROJ-004 (MVP, 확정):** 상위(Application) 로드맵/마일스톤과 하위(Repository) 로드맵/마일스톤을 연결(매핑)할 수 있어야 한다.
    - 모든 하위 마일스톤은 상위 마일스톤에 `child -> parent` 매핑 가능해야 한다.
- **REQ-FR-PROJ-005 (MVP, 확정):** Jira 연동은 하이브리드 정책을 지원해야 한다.
    - 실행 이슈 Source of Truth는 Repository Jira.
    - Project는 Repository 하위 기간성 운영 단위로 관리한다.
    - Project Jira에 작업성 Story/Task 직접 생성은 정책 위반으로 취급.
- **REQ-FR-PROJ-006 (MVP, 확정):** Confluence(또는 문서 체계)는 상/하위 분리 정책을 지원해야 한다.
    - Project 문서: 방향성/의사결정/분기 계획.
    - Repository 문서: 설계/RFC/runbook/회고.
- **REQ-FR-PROJ-007 (MVP, 확정):** 스프린트는 Repository 단위로 운영되어야 하며, Application 레벨은 주간/월간 cadence로 상태를 롤업해야 한다.
    - 권장 cadence: 주간 Program Sync, 월간 KPI/리스크 리뷰.
- **REQ-FR-PROJ-008 (후속):** Project 영구 삭제는 `archive 후 N일 보존 + 관리자 재확인` 정책을 따라야 한다.
- **REQ-FR-PROJ-009 (활성화, 2026-05-15 sprint `claude/work_260515-c`):** Owner 위양(RBAC row-level)은 ADR-0011 §4.2 의 `enforceRowOwnership(c, ownerUserID, allowedRoles...)` helper 로 활성화한다. allow 규칙: (1) `system_admin`, (2) `allowedRoles` 화이트리스트, (3) `actor.login == ownerUserID`. deny 시 `auth.row_denied` audit + 403 + `code=auth_row_denied`. handler 단위 호출은 별도 sprint (pmo_manager seed 결정 후).
- **REQ-FR-PROJ-010 (후속):** `pmo_manager` 역할 활성화 시 권한 범위는 정책 확정 후 단계적으로 허용한다.
    - 기본 후보 범위: `project.manage`, `project.member.manage`, `milestone.mapping.manage`.
    - 제한 후보 범위: `application.manage`(수정만), `application.repo.link`(초기 비허용 권장).
    - 금지 범위: 시스템 설정, 계정/조직/RBAC 정책 변경.

#### 5.4.2 비기능/운영 요구사항 (REQ-NFR)

- **REQ-NFR-PROJ-001 (MVP):** Project/Repository 매핑 정보는 감사(audit) 가능해야 하며 생성/수정/해제 이력을 기록해야 한다.
- **REQ-NFR-PROJ-002 (MVP):** 상위 롤업 지표는 매핑 누락 항목을 조용히 제외하지 않고 경고 상태로 표시해야 한다.
- **REQ-NFR-PROJ-003 (후속):** Project 대시보드 응답시간 목표(예: p95 2초 이내)와 페이지네이션 한계는 설계 단계에서 별도 계약한다.
- **REQ-NFR-PROJ-004 (MVP):** 외부 형상관리/CI/품질 도구 연동 데이터는 idempotency key 기반 중복 방지 및 재동기화(reconciliation) 정책을 가져야 한다.
- **REQ-NFR-PROJ-005 (MVP):** 어댑터 장애는 provider 단위로 격리되어야 하며, 특정 provider 장애가 전체 수집 파이프라인 중단으로 전파되지 않아야 한다.
- **REQ-NFR-PROJ-006 (MVP):** Application 롤업 계산은 동일 요청 조건에서 재현 가능해야 하며, 집계 기준(기간/필터/가중치)을 메타데이터로 함께 제공해야 한다.
    - `weight_policy`와 실사용 가중치 맵(`applied_weights`)을 응답 메타에 포함해야 한다.
    - 가중치 누락 repository는 기본값 fallback(`equal`) 적용 여부를 메타에 명시해야 한다.

#### 5.4.3 Usecase 산출물 (확정)

- 본 아젠다의 설계 진입 직전 Usecase 산출물은 [`docs/planning/system_usecases.md`](./planning/system_usecases.md) 를 source-of-truth 로 사용한다.
- 해당 문서의 `UC-*` 는 REQ와 ARCH/API 사이의 중간 추적 단계다.

#### 5.4.4 ERD 산출물 (확정)

- 데이터 모델 기준 문서는 [`docs/planning/system_erd.md`](./planning/system_erd.md) 를 사용한다.
- 설계/구현 단계의 신규 엔티티·관계는 ERD 문서와 동기화해야 한다.

#### 5.4.5 범위 경계 (Out of Scope)

- 신규 Project 생성 시 Gitea 저장소 자동 생성/브랜치 보호/멤버 초대 자동화는 별도 sprint에서 진행한다 (`§5.3-7` 후속).
- WebSocket 기반 실시간 위험 탐지는 M4 범위에서 다루고, AI 제안 자동화는 v2 범위에서 다룬다.
- MFA 기반 위험 작업 다단계 확인은 운영 진입 직전 정책으로 별도 확정한다.

### 5.5 개발 의뢰 (Dev Request, DREQ) 도메인

본 절은 컨셉 문서([`docs/planning/development_request_concept.md`](./planning/development_request_concept.md), sprint `claude/work_260515-f`)에 정의된 외부 시스템 개발 의뢰 수신 → 담당자 검토 → application/project 등록(promote) 흐름의 요구사항을 발급한다.

#### 5.5.1 기능 요구사항 (REQ-FR-DREQ)

- **REQ-FR-DREQ-001 (MVP):** 외부 시스템은 인증된 API 호출로 개발 의뢰를 등록할 수 있어야 한다.
    - 필수 필드: `title` (≤200자), `requester` (외부 시스템상의 의뢰자 식별자), `assignee_user_id` (DevHub `users.user_id` FK), `source_system`.
    - 선택 필드: `details` (markdown), `external_ref` (외부 시스템 ticket id 등).
    - `(source_system, external_ref)` 조합은 UNIQUE — 동일 외부 ticket 의 재수신은 409 또는 idempotent OK 응답.
- **REQ-FR-DREQ-002 (MVP):** DevHub 는 수신 직후 의뢰의 검증(필수 필드 / assignee_user_id 존재)을 수행하고, 성공 시 `pending` 상태로 저장한다. 검증 실패 시 `rejected` 상태 + `rejected_reason="invalid_intake"` 로 저장한다 (audit 보존 목적, 절대 drop 하지 않는다).
- **REQ-FR-DREQ-003 (MVP):** 의뢰의 상태 머신은 `received → pending → in_review → registered | rejected | closed` 로 한정되며, 모든 전이는 `dev_request.*` audit action 으로 기록되어야 한다. (받음/등록됨/거절됨/재오픈됨/닫힘)
- **REQ-FR-DREQ-004 (MVP):** 담당자는 자신의 `assignee_user_id` 와 일치하는 의뢰 목록을 dashboard 에서 조회할 수 있어야 한다. system_admin 은 모든 의뢰를 조회할 수 있어야 한다.
- **REQ-FR-DREQ-005 (MVP):** 담당자는 의뢰를 application 또는 project 로 등록(promote)할 수 있어야 한다. 등록 시 단일 트랜잭션으로 (a) 신규 application/project 생성, (b) DREQ.status → `registered`, (c) DREQ.registered_target_type / registered_target_id 갱신, (d) audit `dev_request.registered` 가 모두 이루어져야 한다. 부분 실패 시 모두 롤백한다.
- **REQ-FR-DREQ-006 (MVP):** 담당자 또는 system_admin 은 의뢰를 reject 할 수 있어야 하며, `rejected_reason` (텍스트) 은 필수다.
- **REQ-FR-DREQ-007 (MVP):** system_admin 은 의뢰의 `assignee_user_id` 를 변경(reassign)할 수 있어야 한다. 변경 이력은 `dev_request.reassigned` audit 으로 기록한다.
- **REQ-FR-DREQ-008 (MVP):** `registered` 또는 `rejected` 상태의 의뢰는 system_admin 이 `closed` 로 전이할 수 있어야 한다. `pending`/`in_review` 상태의 의뢰는 직접 `closed` 로 갈 수 없다 (먼저 reject 후 close).
- **REQ-FR-DREQ-009 (후속):** Application/Project 의 `origin_dreq_id` 역참조 컬럼 도입 여부는 별도 ADR 에서 결정. 도입 시 nullable FK 로 추가하여 의뢰 없이 직접 생성된 entity 와 공존한다.
- **REQ-FR-DREQ-010 (후속):** 외부 시스템에 의뢰 상태 변경을 callback (webhook) 으로 알리는 기능은 MVP 안정화 후 결정한다.
- **REQ-FR-DREQ-011 (후속):** 의뢰 첨부파일, 댓글, 멘션, 알림, SLA/escalation, AI 자동 분류는 본 도메인의 1차 범위 밖이다.

#### 5.5.2 비기능 / 운영 요구사항 (REQ-NFR-DREQ)

- **REQ-NFR-DREQ-001 (MVP):** 외부 수신 endpoint 의 인증은 운영 진입 전에 [ADR (DREQ-AuthADR)] 으로 결정한다. 후보: (A) API 토큰 + IP allowlist, (B) HMAC 시그니처, (C) OAuth client_credentials.
- **REQ-NFR-DREQ-002 (MVP):** 외부 수신 요청은 idempotency 를 갖는다 — `(source_system, external_ref)` 동일한 재호출은 동일 의뢰 id 로 동일 응답을 반환한다 (혹은 명시적 409 + 기존 id 반환).
- **REQ-NFR-DREQ-003 (MVP):** 모든 상태 전이는 audit_logs 에 기록된다. payload 는 의뢰 id / 전이 from-to / actor / 변경 사유를 포함한다.
- **REQ-NFR-DREQ-004 (MVP):** `details` 필드는 markdown 렌더링 시 XSS 방어를 위해 sanitize 되어야 한다 (frontend 책임). backend 는 raw 저장.
- **REQ-NFR-DREQ-005 (MVP):** DREQ 목록 응답은 페이지네이션(limit/offset 또는 cursor)을 지원하고 기본 limit 50, 최대 100 으로 제한한다.
- **REQ-NFR-DREQ-006 (후속):** 외부 수신 endpoint 의 RPS 한계 / rate limiting 정책은 운영 진입 직전 결정한다.

#### 5.5.3 범위 경계 (Out of Scope)

- AI 자동 분류 / 자동 application 매핑 추천.
- 외부 시스템 callback (webhook 송신).
- 의뢰자(requester) 의 DevHub 직접 로그인 + 자기 의뢰 추적 UI.
- 의뢰 첨부 / 댓글 / 멘션 / 알림 / SLA / 자동 escalation.
- repository 단독 등록 (application 또는 project 만 선택).

## 6. 기술 스택 결정 사항 (Technology Stack Decisions)

기술 스택 상세 계약, 버전, 설치/검증 명령은 **[tech_stack.md](./tech_stack.md)**를 기준으로 관리합니다. 본 요구사항 문서에서는 제품 요구사항과 직접 연결되는 기술 결정 요약만 유지합니다.

- **하이브리드 백엔드:** Gitea 연동, Webhook 수집, 시스템 제어, 권한 관리는 Go Core가 담당합니다. AI 가드너와 분석성 작업(Python AI)은 v2에서 도입합니다.
- **내부 통신:** Go Core와 Python AI의 분석 요청/응답은 gRPC를 기본 계약으로 사용합니다.
- **프론트엔드:** 역할별 진입 우선순위와 실시간 상태 시각화는 Next.js 기반 UI에서 제공합니다.
- **데이터베이스:** Gitea 원본 이벤트, 프로젝트/저장소/사용자/권한 관계, 비정형 분석 결과 저장에는 PostgreSQL을 사용합니다.

## 7. 상세 시스템 아키텍처 설계 (Detailed System Architecture)

상세한 시스템 아키텍처 설계 내용은 별도 문서인 **[architecture.md](./architecture.md)**에서 관리하며, 구체적인 기술 스택 및 환경 설정 가이드는 **[tech_stack.md](./tech_stack.md)**를 참조합니다.

### 주요 아키텍처 결정 사항:
- **내부 통신:** Go Core ↔ Python AI 간 gRPC 도입.
- **실시간성:** WebSocket을 통한 프론트엔드 실시간 상태 전송. SSE는 초기 구현 범위에서 제외하고 운영 환경 제약 발생 시 fallback으로 재검토.
- **시각화:** React Flow를 이용한 인터랙티브 인프라 구성도.

---
