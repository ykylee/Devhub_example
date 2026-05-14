# Application/Project 운영 모델 템플릿

- 문서 목적: Application > Repository > Project 기반 운영안을 표준 템플릿으로 기록한다.
- 범위: 계층 구조, 역할, Jira/Confluence 연결, 로드맵/마일스톤, cadence, KPI, 리스크/의존성.
- 대상 독자: Chief PL, Delivery PL, 시스템 관리자, 트랙 리드.
- 상태: template
- 최종 수정일: 2026-05-13
- 관련 문서: [project_management_concept.md](./project_management_concept.md)

## 1. 기본 정보

- Application 코드:
- Application 이름:
- 기간:
- Chief PL:
- 참여 조직/팀:
- 목표 KPI:

## 2. 계층 구조

- 상위 Application:
- 하위 Repository 목록:
  - repo-1:
  - repo-2:
  - repo-3:
- Repository 하위 Project:
  - repo-1:
  - repo-2:
  - repo-3:

## 3. 역할 배치

- Chief PL:
- Delivery PL (repo별):
  - repo-1:
  - repo-2:
- Tech Lead (repo별):
- Scrum Master/Agile Coach:

## 4. Jira 연결 정책

- 상위 Application Jira 키:
- 용도: `summary_only` (권장)
- 하위 repo Jira 키:
  - repo-1:
  - repo-2:
- 규칙:
  - 실행 이슈 생성은 repo Jira만 허용
  - 상위 Jira는 Epic/Project 롤업만 허용
  - 상위/하위 링크 필수 필드:

## 5. Confluence/Docs 정책

- 상위 Application 문서 공간:
- 하위 repo 문서 공간:
- 규칙:
  - 정책/의사결정 문서: 상위
  - 설계/운영/runbook: 하위 repo

## 6. 로드맵/마일스톤

### 6.1 상위(Application)

| 마일스톤 | 일정 | 완료 기준(DoD) | 상태 |
| --- | --- | --- | --- |
| M1 |  |  |  |
| M2 |  |  |  |
| M3 |  |  |  |

### 6.2 하위(repo)

| repo | 하위 마일스톤 | 상위 매핑 | 일정 | 상태 |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |
|  |  |  |  |  |

## 7. 스프린트/운영 cadence

- Sprint 길이:
- Repo별 Sprint Planning 요일:
- Repo별 Review/Retro 요일:
- 주간 Program Sync (상위):
- 월간 KPI/리스크 리뷰:

## 8. KPI/리포팅

- 상위 KPI:
  - 목표 달성률:
  - 일정 건전성:
  - P1 미해결 건수:
- 하위 KPI (repo별):
  - Sprint 목표 달성률:
  - Lead Time:
  - 결함 유출률:

## 9. 리스크/의존성 관리

- 주요 리스크:
- Cross-repo 의존성:
- 에스컬레이션 경로:

## 10. 운영 정책 체크리스트

- [ ] 상위 Jira에 실행 티켓 생성 금지 정책 합의
- [ ] 하위 Jira -> 상위 Epic 링크 자동화
- [ ] 상/하위 마일스톤 매핑 표 유지
- [ ] 상위 DoD 수치화
- [ ] 월간 운영 회고 일정 확정
