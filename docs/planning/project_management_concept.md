# 프로젝트 관리 도메인 컨셉

- 문서 목적: DevHub 의 신규 1차 도메인 **Application/Repository/Project 운영 모델** 의 컨셉 — 무엇을 묶는 entity 인지, 어떤 usecase 가 핵심인지, MVP 경계 — 를 정의해 후속 요구사항/usecase/설계 단계의 공통 기준으로 삼는다.
- 범위: 컨셉 단계. 도메인 정의 / 행위자 × 핵심 usecase 분리 / 운영 계층 모델 / Jira 연계 정책 / 로드맵·마일스톤 롤업 구조 / MVP scope / 데이터 모델 초안 / 템플릿·예시 운영안 진입점 / out-of-scope / 후속 단계 진입 hook. 본 문서는 **REQ-FR / ARCH / API ID 를 아직 발급하지 않는다** — 후속 sprint 에서 발급.
- 대상 독자: 프로젝트 리드, Backend / Frontend / Auth 트랙 담당자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-13
- 관련 문서: [요구사항 정의서 §5.3, §5.4](../requirements.md), [통합 개발 로드맵](../development_roadmap.md), [아키텍처](../architecture.md), [백엔드 API 계약](../backend_api_contract.md), [추적성 컨벤션](../traceability/conventions.md), [planning 진입점](./README.md), [운영 템플릿](./project_operating_model_template.md), [예시 운영안](./project_operating_model_example_2026.md).

## 1. 컨셉 정리 배경

- 기존 [`requirements.md §5.7 프로젝트 및 저장소 관리`](../requirements.md) 는 *Gitea 저장소 자동 생성 + 브랜치 보호 + 멤버 초대* 같은 **Gitea 연동 자동화 시나리오** 만 다루고, DevHub 자체의 **Project entity 자체의 lifecycle / CRUD / 조회 usecase** 는 정의가 없다.
- 다른 모듈 (조직 / 사용자 / RBAC / 인증) 은 §M2~§M3 으로 1차 안정화되었으나, 그 위에 얹히는 **상위 컨테이너 — "어떤 프로젝트에 무엇이 매여 있는가" — 가 비어 있다.**
- 또한 팀의 운영 관성은 **기간성 운영 과제를 명시적으로 발의**하는 방식이고, 실행 레벨은 repo 단위로 분화되어 있다.
- 따라서 본 컨셉은:
  - Application/Repository/Project 가 **어떤 역할을 갖는 계층 entity 인가** 를 정의하고,
  - **두 핵심 usecase** (일반 사용자 조회 / 시스템 관리자 등록·관리) 를 분리하고,
  - **Application > Repository > Project > GitHub 실행 단위**의 운영 계층을 공식화하고,
  - 기존 §5.7 (Gitea 자동화), M4 (실시간 / AI), 향후 RBAC / 알림 등을 **out-of-scope 후속** 으로 분리한다.
- 본 컨셉이 머지되면 후속 sprint 에서 단계적으로 요구사항(REQ-FR) → usecase → 설계(ARCH/API) 로 발전시킨다.

## 2. 운영 계층 모델 (합의안)

### 2.1 기본 구조

- **최상단 Application** 을 명시적으로 정의한다.
- Application 하위에 **N개의 Repository(실행 단위)** 를 둔다.
- 각 Repository 하위에 **기간성 Project(운영 단위)** 를 둔다.
- 저장소부터는 **GitHub 방식**을 따른다.
  - 이슈: Issue
  - 계획/진행판: Project (Board)
  - 일정 기준점: Milestone
  - 기술/운영 문서: Wiki/Docs

요약하면, **총괄 계층은 Application 중심, 실행 계층은 Repository 중심, 운영 계층은 Project 중심**이다.

### 2.2 대형 과제 구조

- 하나의 제품/서비스는 Application으로 선언한다.
- 하위 repo 각각을 실행 단위로 운영하고, repo 아래 Project를 기간 단위로 갱신한다.
- 필요 시 repo를 도메인별(backend/frontend/data/ops)로 나누고 각 repo에 독립 cadence 를 둔다.

### 2.3 역할 배치

- **Chief PL (Application 1명)**: 목표/KPI/예산/리스크 총괄, 우선순위 충돌 조정.
- **Delivery PL (repo별 1명)**: 스프린트 목표/백로그/납기 책임.
- 스프린트 실행 책임은 repo 레벨에 두고, Application 레벨은 롤업·조정에 집중한다.

## 3. Jira/Confluence 연결 정책

### 3.1 Jira 연결 옵션 비교 결론

- 단일 Project Jira only: 리포팅은 단순하지만 실무 티켓 혼재 위험이 큼.
- repo별 Jira only: 실행 정합성은 높으나 상위 집계 규칙이 필요.
- 연간+repo 동시 연결: 설계를 잘하면 가장 유연하지만 중복 입력 리스크 존재.

**본 컨셉의 기본 정책은 하이브리드다.**

1. 실행 이슈의 Source of Truth 는 **repo Jira**.
2. Project 이슈 관리는 기본적으로 Repository 도구(Jira/GitHub Projects)에 연결한다.
3. 상위 Jira 에 작업성 Story/Task 를 직접 생성하지 않는다.
4. 하위 Jira 이슈를 상위 Epic 에 링크하여 자동 집계한다.

### 3.2 Confluence 정책

- **Application Confluence**: 방향성 문서, 의사결정, 분기 계획, 경영/관리 보고.
- **repo Confluence/Docs**: 설계서, runbook, RFC, 장애/회고 등 실행 문서.
- 원칙: 문서도 실행 단위는 repo 우선, 상위는 요약·정책 중심.

## 4. 로드맵/마일스톤 계층 운영

### 4.1 계층 원칙

- 상위(Application): **큰 로드맵/마일스톤** (분기·반기 게이트).
- 하위(repo): **세부 로드맵/마일스톤** (스프린트·월 단위).

### 4.2 필수 규칙

1. 하위 마일스톤은 반드시 상위 마일스톤에 매핑한다 (`child -> parent`).
2. 상위 계층은 작업 티켓을 직접 소유하지 않는다.
3. 상위 완료 기준(DoD)은 수치로 명시한다.
4. 매핑 누락 항목은 리포팅에서 제외가 아닌 "경고"로 표시한다.

### 4.3 완료 기준 예시 (상위)

- 연결된 하위 마일스톤 완료율 90% 이상.
- P1 미해결 0건.
- 통합 테스트/릴리즈 게이트 통과.

## 5. 행위자 × 핵심 usecase 분리

### 5.1 일반 사용자 (Developer / Manager / QA) — 조회 중심

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 내가 멤버인 과제/저장소 묶음 보기 | 자신이 owner/member 인 active/on_hold 과제 및 연결 repo 요약 | ✅ |
| 공개(public) 과제 보기 | visibility=`public` 과제 목록 | ✅ |
| 과제 상세 보기 | 메타 + 멤버 + 상태 + 연결 repo + 상/하위 마일스톤 요약 | ✅ |
| 진행 현황/위험 보기 | 진행률, 최근 활동, blocker | 후속 (M4 연계) |
| 검색/필터 | status, owner, 부서, repo | ✅ (기본 필터) |

### 5.2 시스템 관리자 (System Admin) — 등록·관리 전용

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 전체 과제 목록 보기 | 보관(archived) 포함 옵션 | ✅ |
| Application 신규 등록 | name / description / code / owner / visibility / 기간 / KPI | ✅ |
| 과제 메타 수정 | 위 항목 + status 전환 | ✅ |
| 하위 repo 연결/해제 | 1:N repo 매핑 | ✅ |
| 멤버/책임자 관리 | owner 교체, 멤버 role 변경 | ✅ |
| 과제 보관 (archive) | soft-delete 성격 | ✅ |
| 영구 삭제 | 보관 후 정책 기반 삭제 | 2차 |
| Jira/Confluence 연결 정책 관리 | 상위/하위 연결 규칙 검증 | ✅ |

### 5.3 책임자(Owner) 권한 위양 범위

- MVP 1차: 시스템 관리자가 모든 쓰기 작업의 단일 주체.
- 2차: owner 에게 제한적 메타/멤버 수정 위양.
- 3차: RBAC row-scoping (ADR-0011 후보) 채택 시 부서 단위 위양 확장.

## 6. 스프린트/애자일 운영 모델

1. 스프린트는 **repo 단위 독립 운영**이 기본.
2. Application 레벨은 **주간 Program Sync + 월간 KPI 리뷰** cadence 로 동기화.
3. 공통 게이트(Alpha/Beta/GA 등)는 상위 마일스톤으로만 강제.
4. Jira 티켓 생성은 repo 레벨만 허용(상위는 롤업 전용).

## 7. 데이터 모델 초안 (운영 계층 반영)

```text
applications
  id                 UUID PK
  code               TEXT UNIQUE
  name               TEXT NOT NULL
  description        TEXT
  status             TEXT NOT NULL   -- planning | active | on_hold | closed | archived
  visibility         TEXT NOT NULL   -- public | internal | restricted
  owner_user_id      UUID FK users.id
  start_date         DATE NULL
  due_date           DATE NULL
  created_at         TIMESTAMPTZ NOT NULL
  updated_at         TIMESTAMPTZ NOT NULL
  archived_at        TIMESTAMPTZ NULL

application_repositories
application_integrations
  application_id      UUID FK applications.id
  repo_provider      TEXT NOT NULL   -- github | gitea
  repo_full_name     TEXT NOT NULL   -- org/repo
  role               TEXT NOT NULL   -- primary | sub | shared
  linked_at          TIMESTAMPTZ NOT NULL
  PRIMARY KEY (application_id, repo_full_name)

application_integrations
  application_id      UUID FK applications.id
  scope              TEXT NOT NULL   -- project | repository
  integration_type   TEXT NOT NULL   -- jira | confluence
  external_key       TEXT NOT NULL
  url                TEXT NOT NULL
  policy             TEXT NOT NULL   -- summary_only | execution_system
```

## 8. 템플릿 / 예시 운영안

- 운영 템플릿: [`project_operating_model_template.md`](./project_operating_model_template.md)
- 예시 운영안: [`project_operating_model_example_2026.md`](./project_operating_model_example_2026.md)

## 9. 명시적 out-of-scope (별도 sprint / 마일스톤)

| 항목 | 분류 | 후속 진입 |
| --- | --- | --- |
| Gitea 저장소 자동 생성 / 브랜치 보호 / 멤버 자동 초대 | §5.7 자동화 시나리오 | 별도 sprint, §5.7 source-of-truth 유지 |
| 실시간 진행 상태 (WebSocket) | M4 이벤트 확장 | M4 |
| AI Suggestion / 위험 자동 탐지 | AI 가드너 | M4 |
| MFA / 위험 작업(영구 삭제) 다단계 확인 | 운영 보안 | 운영 진입 직전 |
| CSV 일괄 가져오기 / 내보내기 | 운영 효율화 | 운영 진입 후 |

## 10. 미해결 / 결정 대기

| 항목 | 결정 후보 | 결정 시점 |
| --- | --- | --- |
| Application `code` 형식 (대문자/숫자/하이픈 패턴, 길이) | 후속 design | Req sprint |
| 영구 삭제 정책 | 보관 후 N일 + admin 재확인 | Design sprint |
| Owner 위양 범위 | 2차 sprint | 후속 |
| RBAC row-level enforcement | ADR-0011 후보 | Design sprint |
| GitHub/Gitea 동시 운영 시 표준 provider 정책 | github 우선 + gitea 호환 어댑터 | Integration sprint |

## 11. 후속 단계 진입 hook

| 단계 | 산출물 후보 | 진입 조건 |
| --- | --- | --- |
| Req sprint | `docs/requirements.md` 확장 (`Project/하위 repo/Jira 정책`) + REQ-FR 발급 | 본 컨셉 머지 직후 |
| Usecase sprint | 행위자 × usecase 매트릭스, Jira/Confluence 운영 규칙 상세 | Req sprint 머지 직후 |
| Design sprint (backend) | `/api/v1/applications/*`, `/api/v1/repositories/:repo_id/projects/*` 계약 + 마이그레이션 초안 | Usecase sprint 머지 직후 |
| Design sprint (frontend) | `/admin/settings/applications` IA/화면 흐름 | backend design 과 병행 |
| Implementation sprint | IMPL-application-*, UT-application-*, TC-application-* | design 머지 후 |

## 12. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-13 | 초안 — 컨셉 1차. 도메인 정의 + 핵심 usecase 분리 + MVP scope + 데이터 모델 초안 + RBAC 영향(보류) + 후속 단계 진입 hook. | sprint `claude/work_260513-p`. |
| 2026-05-13 | 리뷰 1차 보강 — archived 가시성/영구삭제 단계/Owner 위양 단계/`code` 제약 미해결 정합화. | 동일 sprint. |
| 2026-05-13 | 컨셉 심화 — Application > Repository > Project 운영 계층, Jira/Confluence 하이브리드 정책, 상/하위 로드맵·마일스톤 롤업 규칙, PL·스프린트 운영 모델, 템플릿/예시 문서 링크 추가. | 사용자 워크숍 합의 반영. |
