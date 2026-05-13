# Application/Project 운영 모델 예시 (2026)

- 문서 목적: Application > Repository > Project 기반 하이브리드 운영 모델의 샘플을 제공한다.
- 범위: Application 1건 + Repo 3건 + Repository 하위 Project 기준의 역할/연결/로드맵/cadence/KPI 예시.
- 대상 독자: PMO, 시스템 관리자, Delivery PL, 개발 리드.
- 상태: example
- 최종 수정일: 2026-05-13
- 관련 문서: [project_management_concept.md](./project_management_concept.md), [project_operating_model_template.md](./project_operating_model_template.md)

## 1. 기본 정보

- Application 코드: `APP-PLATFORM`
- Application 이름: DevHub Platform
- 운영 Project 코드: `INIT-2026-PLATFORM`
- 운영 Project 이름: 플랫폼 통합 고도화 2026
- 기간: 2026-01-01 ~ 2026-12-31
- Chief PL: 플랫폼 총괄 PM
- 목표 KPI:
  - 연말까지 핵심 시스템 가용성 99.9%
  - 배포 리드타임 30% 단축
  - 운영 장애 MTTR 20% 단축

## 2. 계층 구조

- 상위 Application: `APP-PLATFORM`
- 하위 Repository:
  - `org/devhub-backend-core`
  - `org/devhub-frontend`
  - `org/devhub-platform-ops`
- Repository 하위 Project:
  - `org/devhub-backend-core` -> `PLAT-BE-2026`
  - `org/devhub-frontend` -> `PLAT-FE-2026`
  - `org/devhub-platform-ops` -> `PLAT-OPS-2026`

## 3. 역할 배치

- Chief PL: 1명 (연간 KPI, 예산, 리스크 총괄)
- Delivery PL:
  - backend-core: 백엔드 리드 PM
  - frontend: 프론트엔드 리드 PM
  - platform-ops: 인프라 운영 PM
- Tech Lead:
  - backend-core: 백엔드 TL
  - frontend: 프론트엔드 TL
  - platform-ops: SRE TL

## 4. Jira 연결

- 상위 Application Jira: `PLAT-INIT` (summary_only)
- 하위 repo Jira:
  - backend-core: `PLAT-BE`
  - frontend: `PLAT-FE`
  - platform-ops: `PLAT-OPS`
- 운영 규칙:
  - 작업 티켓은 `PLAT-BE/FE/OPS`에서만 생성
  - `PLAT-INIT`에는 Epic/Project만 유지
  - 하위 Epic은 상위 `PLAT-INIT` Epic에 링크

## 5. Confluence/Docs

- 상위 공간: `PLATFORM-INIT` (전략, 의사결정, 분기계획)
- 하위 공간:
  - backend-core: `PLATFORM-BE`
  - frontend: `PLATFORM-FE`
  - platform-ops: `PLATFORM-OPS`

## 6. 로드맵/마일스톤

### 6.1 상위 마일스톤

| 마일스톤 | 일정 | DoD |
| --- | --- | --- |
| M1 아키텍처 확정 | 2026-03-31 | 하위 설계 milestone 완료율 90% 이상, P1 미해결 0 |
| M2 베타 운영 | 2026-07-31 | 통합 e2e 통과, 주요 장애 시나리오 훈련 완료 |
| M3 운영 전환 | 2026-11-30 | 가용성/KPI 기준 충족, 운영 인수인계 완료 |

### 6.2 하위 마일스톤 매핑

| repo | 하위 마일스톤 | 상위 매핑 |
| --- | --- | --- |
| backend-core | 인증/권한 API 안정화 | M1 |
| frontend | 관리자 콘솔 정합성 개선 | M1 |
| platform-ops | 배포 파이프라인 표준화 | M1 |
| backend-core | 실시간 이벤트 안정화 | M2 |
| frontend | 운영 대시보드 베타 | M2 |
| platform-ops | 관측성/알림 튜닝 | M2 |

## 7. cadence

- Sprint: repo별 2주
- Repo Planning: 격주 월요일
- Repo Review/Retro: 격주 금요일
- Program Sync: 매주 수요일 30분
- KPI/리스크 리뷰: 매월 마지막 주 목요일 60분

## 8. KPI

- 상위 KPI:
  - 분기 목표 달성률
  - 상위 milestone on-time 비율
  - P1/P0 리스크 체류일
- 하위 KPI:
  - Sprint 목표 달성률
  - PR lead time
  - 결함 유출률

## 9. 리스크/의존성 관리 예시

- 리스크: backend 인증 정책 변경이 frontend 일정에 영향
- 대응: Program Sync에서 cross-repo blocker를 주간 추적
- 에스컬레이션: 48시간 이상 blocker 지속 시 Chief PL 결정 회의 상정

## 10. 운영 핵심 원칙 (요약)

1. 상위는 전략/롤업, 하위는 실행/티켓 소유.
2. 상위 Jira/Confluence는 요약·결정 중심, 하위는 실행 문서 중심.
3. 모든 하위 마일스톤은 상위 마일스톤에 연결한다.
