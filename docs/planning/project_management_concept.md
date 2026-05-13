# 프로젝트 관리 도메인 컨셉

- 문서 목적: DevHub 의 신규 1차 도메인 **Project(프로젝트)** 의 컨셉 — 무엇을 묶는 entity 인지, 어떤 usecase 가 핵심인지, MVP 경계 — 를 정의해 후속 요구사항/usecase/설계 단계의 공통 기준으로 삼는다.
- 범위: 컨셉 단계. 도메인 정의 / 행위자 × 핵심 usecase 분리 / MVP scope / 데이터 모델 초안 / 다른 도메인과의 연계 / out-of-scope / 후속 단계 진입 hook. 본 문서는 **REQ-FR / ARCH / API ID 를 아직 발급하지 않는다** — 후속 sprint 에서 발급.
- 대상 독자: 프로젝트 리드, Backend / Frontend / Auth 트랙 담당자, AI agent, 리뷰어.
- 상태: draft
- 최종 수정일: 2026-05-13
- 관련 문서: [요구사항 정의서 §5.7](../requirements.md) (기존 Project Manager / Admin 시나리오), [통합 개발 로드맵](../development_roadmap.md), [아키텍처](../architecture.md), [백엔드 API 계약](../backend_api_contract.md), [추적성 컨벤션](../traceability/conventions.md), [planning 진입점](./README.md).

## 1. 컨셉 정리 배경

- 기존 [`requirements.md §5.7 프로젝트 및 저장소 관리`](../requirements.md) 는 *Gitea 저장소 자동 생성 + 브랜치 보호 + 멤버 초대* 같은 **Gitea 연동 자동화 시나리오** 만 다루고, DevHub 자체의 **Project entity 자체의 lifecycle / CRUD / 조회 usecase** 는 정의가 없다.
- 다른 모듈 (조직 / 사용자 / RBAC / 인증) 은 §M2~§M3 으로 1차 안정화되었으나, 그 위에 얹히는 **상위 컨테이너 — "어떤 프로젝트에 무엇이 매여 있는가" — 가 비어 있다.**
- 따라서 본 컨셉은:
  - Project 가 **무엇을 묶는 1차 entity 인가** 를 정의하고,
  - **두 핵심 usecase** (일반 사용자 조회 / 시스템 관리자 등록·관리) 를 분리하고,
  - **CRUD + 등록 + 조회의 MVP** 경계를 명시한 뒤,
  - 기존 §5.7 (Gitea 자동화), M4 (실시간 / AI), 향후 RBAC / 알림 등을 **out-of-scope 후속** 으로 분리한다.
- 본 컨셉이 머지되면 후속 sprint 에서 단계적으로 요구사항(REQ-FR) → usecase → 설계(ARCH/API) 로 발전시킨다.

## 2. 도메인 정의

### 2.1 Project 는 무엇인가

DevHub 의 **Project** 는 다음을 묶는 단일 식별 가능한 작업 컨테이너다:

- **목적**: 하나의 비즈니스 목표 / 제품 / 산출물을 추적하기 위한 1차 단위.
- **소유**: 1명의 책임자 (project owner) + 0..N 명의 멤버.
- **기간**: 시작일 / 종료(예정)일 / 보관(archive) 일자.
- **상태**: 라이프사이클 phase (계획 / 진행 / 보류 / 종료 / 보관).
- **외부 자산 매핑** (후속): 0..N 개의 Gitea 저장소, 0..N 개의 마일스톤, CI 파이프라인.

> "프로젝트" 는 조직(org) / 사용자(user) / 권한(RBAC) 의 *상위 컨테이너* 가 아니다 — 그것들 위에 **얹히는** 별도 entity 다. 조직 단위 (`organizational_unit`) 가 사람의 소속을 표현하듯, 프로젝트는 사람의 **활동의 묶음** 을 표현한다.

### 2.2 무엇을 묶는가 — 1차 / 2차 / 후속

| 묶음 대상 | 1차 (MVP) | 2차 | 후속 |
| --- | --- | --- | --- |
| 멤버 (User + role) | ✅ | — | — |
| 책임자 (Owner) | ✅ | — | — |
| 메타 정보 (name, description, code, status, dates) | ✅ | — | — |
| 가시성 (visibility — public / internal / restricted) | ✅ | — | — |
| 부서 / 조직 단위 매핑 (owner_org_unit) | 선택 | ✅ | — |
| Gitea 저장소 (0..N) | — | ✅ | — |
| 마일스톤 / 일정 (0..N) | — | ✅ | — |
| CI 파이프라인 / Runner 상태 | — | — | ✅ (M4 와 연계) |
| 진행률 / Burn-down | — | — | ✅ |
| AI 분석 / Suggestion | — | — | ✅ (M4 와 연계) |
| Audit log (CRUD 의 actor 추적) | ✅ | — | — |

### 2.3 비교 — 인접 도메인과의 경계

| 항목 | 어디서 책임지는가 |
| --- | --- |
| 사용자 식별 (login, identity) | Auth (Hydra + Kratos) — `users` 테이블 |
| 사용자의 조직 소속 | Org (`org_units`, `org_memberships`) |
| 권한 매트릭스 (resource × action) | RBAC (`rbac_policies`) |
| **Project 와 사용자의 관계 (소속 / 역할)** | **본 도메인 — `project_members`** |
| 작업 단위 (issue, PR, commit) | 외부 (Gitea) — DevHub 는 매핑만 보유 |
| 빌드 / CI 상태 | 외부 (Gitea Actions) — DevHub 는 캐시·시각화만 |

## 3. 행위자 × 핵심 usecase 분리

### 3.1 일반 사용자 (Developer / Manager / QA) — **조회 중심**

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 내가 멤버인 프로젝트 목록 보기 | 자신이 owner 또는 member 인 active/on_hold 프로젝트. **보관(archived) 프로젝트는 본인 멤버였던 경우에 한해 별도 toggle 로 표시** (1차에서는 default off) | ✅ |
| 공개(public) 프로젝트 목록 보기 | visibility=`public` 프로젝트 (멤버 아님 포함) | ✅ |
| 프로젝트 상세 보기 | 메타 + 멤버 + 상태 + (후속) 저장소·마일스톤 요약 | ✅ |
| 자신이 멤버인 프로젝트의 진행 현황 / 위험 보기 | 진행률, 최근 활동, 알림 | 후속 (M4 연계) |
| 프로젝트 검색 / 필터 (status, owner, 부서) | 목록 위에 검색 | ✅ (간단한 substring + status filter) |

> Manager 역할의 "관리 대시보드" (`requirements.md §2.2`) 에서 **자신이 owner 인 프로젝트 군집 시각화** 는 컨셉상 같은 조회 usecase 의 변형으로 본다. MVP 에서는 별도 화면 없이 *목록 + 필터* 로 흡수.

### 3.2 시스템 관리자 (System Admin) — **등록·관리 전용**

| usecase 후보 | 설명 | MVP |
| --- | --- | --- |
| 전체 프로젝트 목록 보기 (visibility 무관) | 보관(archived) 포함 옵션 | ✅ |
| 신규 프로젝트 등록 | name / description / code / owner / 초기 멤버 / 가시성 | ✅ |
| 프로젝트 메타 수정 | 위 항목 + status 전환 | ✅ |
| 멤버 추가 / 제거 / 역할 변경 | project-member relation | ✅ |
| 책임자 (owner) 교체 | 위양 | ✅ |
| 프로젝트 보관 (archive) | soft-delete 성격, 본문은 유지 | ✅ |
| 프로젝트 영구 삭제 | 멤버십·매핑·이력 전부 정리 (audit 보존) | 2차 (보관 후 N일 보존 정책 + 다단계 확인 — §8 결정 대기, Design sprint) |
| 일괄 변경 (CSV 가져오기, 부서 매핑 일괄 변경) | 운영 효율화 | 후속 |
| Gitea 저장소 자동 생성 / 브랜치 보호 적용 / 멤버 자동 초대 | `requirements.md §5.7` 의 자동화 시나리오 | 후속 (별도 sprint, §5.7 가 source-of-truth) |

> 시스템 관리자의 시스템 메뉴는 기존 `/admin/settings/{users,organization,permissions}` 옆에 **`/admin/settings/projects`** 가 추가되는 형태로 본다. RBAC enforcement 는 §5.3 참조.

### 3.3 결정 — 책임자(Owner) 의 권한 위양 범위

본 컨셉 단계에서는 **권한 위양을 최소화** 한다.

- **MVP 1차 (본 컨셉)**: 시스템 관리자가 *모든 쓰기 작업의 단일 주체*. 책임자(owner) 는 식별자 (`owner_user_id`) 로 표현되지만, MVP 1차에서는 쓰기 권한과 분리. UI 에서는 "이 프로젝트의 책임자" 로 정보 표시만.
- **2차 (1차 머지 후 별도 sprint)**: 책임자에게 자신이 owner 인 프로젝트의 **메타 수정 (description, dates) + 멤버 추가/제거** 위양. 단 *등록 / 가시성 변경 / 보관(archive) / 영구 삭제* 는 시스템 관리자 전용 유지.
- **3차 (후속)**: RBAC row-scoping ([§5.3](#53-rbac-매트릭스-영향-컨셉-단계-결정-대기) ADR-0011 후보) 가 채택되면 *부서 단위* 위양 (`owner_org_unit_id` 기반) 으로 단계 확장.

## 4. MVP scope — "CRUD 우선 + 등록 + 조회"

### 4.1 1차 (본 컨셉이 가르치는 MVP)

1. **C — 등록 (Create)**: 시스템 관리자의 신규 프로젝트 등록.
2. **R — 조회 (Read)**:
   - 일반 사용자: 자신이 멤버 + 공개 프로젝트 목록 + 상세.
   - 시스템 관리자: 전체 프로젝트 목록 (filter, 보관 포함) + 상세.
3. **U — 수정 (Update)**: 시스템 관리자의 메타 / 멤버 / owner / 가시성 / status 수정.
4. **D — 삭제 (Delete)**: 시스템 관리자의 보관(archive) + (후속 정책 후) 영구 삭제.

### 4.2 2차 (1차 머지 후 follow-on)

- 책임자(owner) 의 메타/멤버 수정 권한 위양.
- 부서(`org_unit`) 매핑 기반 *부서 관리자* 위양 가능성.
- 검색·필터 고도화 (full-text, tag).

### 4.3 명시적 out-of-scope (별도 sprint / 마일스톤)

| 항목 | 분류 | 후속 진입 |
| --- | --- | --- |
| Gitea 저장소 자동 생성 / 브랜치 보호 / 멤버 자동 초대 | §5.7 자동화 시나리오 | 별도 sprint, §5.7 source-of-truth 유지 |
| 마일스톤 / 진행률 / Burn-down | 진행 시각화 | M4 또는 후속 |
| 실시간 진행 상태 (WebSocket) | M4 — RM-M4-01..03 위에서 publish 한 event 의 project-scope 필터 | M4 RM-M4-02 (리소스 필터링) 와 같이 |
| AI Suggestion / 위험 자동 탐지 | AI 가드너 | M4 RM-M4-04·05 |
| 외부 칸반 / 일정 도구 연동 | 외부 SaaS | M4 이후 |
| MFA / 위험 작업 (영구 삭제) 의 다단계 확인 | 운영 보안 | 운영 진입 직전 별도 결정 |
| CSV 일괄 가져오기 / 내보내기 | 운영 효율화 | 운영 진입 후 |

## 5. 데이터 모델 초안

> 본 절은 컨셉 단계 초안이다. 실 마이그레이션 / 인덱스 / 제약은 후속 design sprint 에서 `backend_api_contract.md` + 마이그레이션 디렉터리에 정합화한다.

### 5.1 핵심 테이블

```text
projects
  id                 UUID PK
  code               TEXT UNIQUE     -- 사람 읽는 짧은 식별자 (예: "DEVHUB", "ML-PLATFORM"). 형식 제약은 §8 미해결 항목.
  name               TEXT NOT NULL
  description        TEXT
  status             TEXT NOT NULL   -- planning | active | on_hold | closed | archived
  visibility         TEXT NOT NULL   -- public | internal | restricted
  owner_user_id      UUID FK users.id   -- 책임자 (1명)
  owner_org_unit_id  UUID FK org_units.id NULL  -- 책임 부서 (선택, 후속 활용)
  start_date         DATE NULL
  due_date           DATE NULL
  created_at         TIMESTAMPTZ NOT NULL
  updated_at         TIMESTAMPTZ NOT NULL
  archived_at        TIMESTAMPTZ NULL

project_members
  project_id   UUID FK projects.id
  user_id      UUID FK users.id
  role         TEXT NOT NULL   -- lead | contributor | observer
  joined_at    TIMESTAMPTZ NOT NULL
  PRIMARY KEY (project_id, user_id)
```

### 5.2 후속 (2차 / 별도 sprint)

```text
project_repositories     -- §5.7 자동화 시나리오의 1:N 저장소 매핑 (후속)
  project_id, gitea_repo_id, kind (primary | docs | ...), linked_at

project_milestones       -- 진행 시각화의 후속 (M4 또는 별도)
  id, project_id, name, due_date, status

project_audit_view       -- audit_logs 의 project_id 필터 view (후속, audit 매트릭스 확장)
```

### 5.3 RBAC 매트릭스 영향 (컨셉 단계 결정 대기)

현재 RBAC 5 resources: `infrastructure`, `pipelines`, `organization`, `security`, `audit` ([ADR-0002](../adr/0002-rbac-policy-edit-api.md), `backend_api_contract.md §12`).

본 컨셉의 후보 — **신규 `projects` resource 추가** + 4-boolean (`view`, `create`, `edit`, `delete`).

- 기본 정책 후보:
  - `developer`: `view=true` (자신이 멤버 또는 public), 나머지 `false`.
  - `manager`: `view=true`, `edit=true` (자신이 owner 인 프로젝트 한정 — row-level), `create`/`delete` `false`.
  - `system_admin`: 전체 4 `true`.
- **row-level enforcement** (자신이 owner / member 인 프로젝트 한정 수정) 은 현 RBAC matrix 가 표현하지 못함 → 후속 ADR 후보 (RBAC row-scoping).
- 본 sprint 에서는 결정 보류, 후속 design sprint 에서 ADR-0011 후보로 정리.

## 6. 다른 도메인과의 연계

| 인접 도메인 | 연계 지점 | 1차 / 후속 |
| --- | --- | --- |
| **Users (M2)** | `owner_user_id`, `project_members.user_id` | 1차 |
| **Organization (M3)** | `owner_org_unit_id` (선택), 부서 단위 위양 후보 | 1차 (옵션 컬럼) / 후속 (위양 정책) |
| **HRDB (M3)** | 신규 멤버 lookup 시 hrdb 조회 재사용 | 후속 |
| **RBAC (M1)** | 신규 `projects` resource + row-level enforcement | 후속 (§5.3) |
| **Audit (M1·M2)** | 모든 CRUD 의 actor / request_id / IP 기록 | 1차 |
| **Gitea (M4)** | `project_repositories` 매핑 + §5.7 자동화 | 2차 / 후속 |
| **WebSocket (M4 RM-M4-01..03)** | `project.updated` event publish + 리소스 필터 | 후속 (M4 와 같이) |
| **AI Gardener (M4 RM-M4-04·05)** | 프로젝트별 Suggestion / 위험 탐지 | 후속 (M4 와 같이) |
| **System Admin Dashboard (M4 RM-M4-07)** | `/admin/settings/projects` 메뉴 추가 | 1차 (M4 와 별개로 진입 가능) |

## 7. UI / UX 컨셉 (초안)

> 상세 화면 설계는 후속 design sprint. 본 절은 진입 hook 만.

| 화면 | 진입 경로 | 1차 / 후속 |
| --- | --- | --- |
| 프로젝트 목록 (일반 사용자) | 상단 nav `Projects` 또는 사이드바 | 1차 |
| 프로젝트 상세 | `/projects/{code}` | 1차 |
| 프로젝트 생성 (system admin) | `/admin/settings/projects` → "신규 등록" | 1차 |
| 프로젝트 관리 (system admin, 전체 목록 + filter) | `/admin/settings/projects` | 1차 |
| 멤버 관리 / Owner 위양 | `/admin/settings/projects/{code}/members` | 1차 |
| Gitea 저장소 매핑 | `/admin/settings/projects/{code}/repositories` | 후속 |

역할별 기본 진입 우선순위 ([`requirements.md §2`](../requirements.md)) 와의 정합:

- Developer / Manager: 기존 dashboard 진입 후 *Projects* 진입.
- System Admin: 기본 진입은 system dashboard, 시스템 메뉴에서 *Projects* 관리.

## 8. 미해결 / 결정 대기

| 항목 | 결정 후보 | 결정 시점 |
| --- | --- | --- |
| Project `code` 의 형식 (대문자/숫자/하이픈 패턴, 길이) | 후속 design | Req sprint |
| 영구 삭제(`delete`) 의 정책 — 즉시 vs 보관 후 N일 | 30일 보관 후 admin 재확인 후 삭제 (보안 검토 입력) | Design sprint |
| Owner 위양 범위 (책임자가 메타·멤버 수정 가능 여부) | 2차 sprint 로 보류 | 후속 |
| RBAC row-level enforcement (자신이 owner / member 인 한정) | ADR-0011 후보 | Design sprint |
| Gitea 자동화 (§5.7) 와의 시점 통합 (등록 즉시 자동화 vs 별도 단계) | 별도 sprint | Gitea 자동화 sprint |
| 시스템 관리자 대시보드 (`RM-M4-07`) 와의 진입 통합 | `/admin/settings/projects` 가 RM-M4-07 진입 시 흡수되는 형태로 정합 | M4 진입 시 |

## 9. 후속 단계 진입 hook

| 단계 | 산출물 후보 | 진입 조건 |
| --- | --- | --- |
| **Req sprint** | `docs/requirements.md` 의 §5.7 확장 또는 신규 절 `§5.8 Project 도메인` 신설, REQ-FR-* 발급 (개별 usecase 단위), NFR (응답시간/조회 페이지네이션) 정의 | 본 컨셉 머지 직후 |
| **Usecase sprint** | 행위자 × usecase 매트릭스, 핵심 시퀀스 (등록·조회·멤버 변경 3종), RBAC 매트릭스 확장 후보 → ADR-0011 초안 | Req sprint 머지 직후 |
| **Design sprint (backend)** | `architecture.md` 의 Project 컴포넌트 추가, `backend_api_contract.md` 신규 § `/api/v1/projects/*` , 마이그레이션 (`000012_projects.sql` 등) | Usecase sprint 머지 직후 |
| **Design sprint (frontend)** | `frontend_development_roadmap.md` 의 새 phase 추가, 진입 경로 / 컴포넌트 / store 모델 초안 | Design sprint (backend) 와 병행 |
| **Implementation sprint** | IMPL-project-*, UT-project-*, TC-PROJ-* 발급 | 모든 design sprint 머지 후 |

> 각 단계 진입 시점에 본 컨셉 문서의 §10 변경 이력에 1줄 추가 + 본문 §5 / §6 / §7 의 후보 항목을 확정·삭제·이동한다. 본 컨셉이 *의사결정의 origin* 으로 살아 있도록 유지한다.

## 10. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-13 | 초안 — 컨셉 1차. 도메인 정의 + 핵심 usecase 분리 + MVP scope + 데이터 모델 초안 + RBAC 영향(보류) + 후속 단계 진입 hook. | sprint `claude/work_260513-p`. base: `118899b` (PR #101 머지 직후). 추적성 ID 미발급 (컨셉 단계). |
| 2026-05-13 | 리뷰 1차 보강 — (1) §3.1 archived 멤버 가시성 명시, (2) §3.2 영구 삭제를 *2차* 로 통일 (§4.1.D / §8 와 정합), (3) §3.3 1·2·3차 단계로 명확화 + ADR-0011 forward link, (4) §5.1 `code` 형식 제약을 §8 미해결로 forward link. | 동일 sprint, PR 직전 정합성 패스. |
