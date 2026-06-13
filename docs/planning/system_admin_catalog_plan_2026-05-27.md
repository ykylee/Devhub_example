# System Admin 통합 관리 UI 계획 (Applications / Repositories / Projects)

- 문서 목적: 시스템 관리자가 등록된 Platform/Repository/Project 전체를 조회·관리할 수 있는 관리자 전용 UI 방향과 구현 단계를 정의한다.
- 범위: 메뉴 IA 결정, 권한/라우팅 가드, 목록/상세 UX, 단계별 구현/검증, 리스크 및 롤아웃
- 대상 독자: PMO, 시스템 관리자, 프론트엔드/백엔드 개발자, QA
- 상태: draft
- 최종 수정일: 2026-05-27
- 관련 문서:
  - `docs/planning/project_repository_creation_linking_plan_2026-05-27.md`
  - `docs/planning/ui_app_project_repo_upgrade_plan.md`
  - `docs/planning/view_menu_screen_api_matrix.md`
  - `docs/governance/worker_division.md`

---

## 1) 배경 및 문제 정의

현재 사용자 관점의 Platform/Repository/Project 화면은 존재하지만, 시스템 관리자가 전체 자산을 단일 운영 동선에서 관리하기에는 다음 제약이 있다.

1. 메뉴가 사용자 기능 중심으로 분산되어 운영자 동선이 길다.
2. 전체 검색/필터/관계 탐색(App↔Repo↔Project) 관점이 약하다.
3. 시스템 관리 작업(일괄 점검, 상태 정리, 연결 정합 확인)을 한곳에서 수행하기 어렵다.

목표는 `system_admin`이 운영 목적의 단일 진입점에서 3 도메인 자산을 조회/관리할 수 있도록 UI를 제공하는 것이다.

---

## 2) 메뉴 구조 옵션 검토

### 옵션 A. 기존 `Admin > Settings` 하위에 기능 확장

- 장점:
  - 초기 구현이 빠름
  - 기존 라우트/레이아웃 재사용 가능
- 단점:
  - Settings(설정)와 Operations(운영) 성격 혼재
  - 도메인 확장 시 정보 밀도 과다, 유지보수성 저하
  - 관리자 핵심 동선이 설정 메뉴 하위에 매몰됨

### 옵션 B. `Admin` 내 신규 운영 메뉴 신설 (권장)

- 제안 라벨: `Admin Catalog` (또는 `Resource Catalog`)
- 구성: `Applications / Repositories / Projects` 탭 기반
- 장점:
  - 운영 동선 분리(설정 vs 자산 운영)
  - 권한·감사 관점 명확
  - 확장성 우수(향후 Integrations/Users relation 탭 추가 용이)
- 단점:
  - 초기 라우팅/사이드바 정보구조 수정 필요

### 결론

`옵션 B (Admin Catalog 신설)` 채택.

---

## 3) 정보구조(IA) 및 라우팅 제안

1. 사이드바
- `Admin` 그룹에 `Catalog` 신규 항목 추가

2. 라우트
- `/admin/catalog` (진입, 기본 `applications` 탭)
- `/admin/catalog?tab=applications`
- `/admin/catalog?tab=repositories`
- `/admin/catalog?tab=projects`

3. 권한
- `system_admin`만 접근
- 비-admin 접근 시 기존 정책대로 `defaultLandingFor(role)`로 redirect

---

## 4) 화면 범위 (MVP)

### 4.1 공통

1. 상단 검색바(키/이름/owner 통합 검색)
2. 상태 필터(활성/보관/기타)
3. 정렬(최신 수정일/이름/키)
4. 공통 empty/loading/error/retry 상태

### 4.2 Applications 탭

1. 컬럼: key, name, status, visibility, owner, linked repo count, updated_at
2. 액션: 상세 보기, 수정 진입, 보관/복원(정책 허용 시)

### 4.3 Repositories 탭

1. 컬럼: full_name, provider(가능 시), private, linked project count, updated_at
2. 액션: 상세 보기, 연결된 project/app 탐색

### 4.4 Projects 탭

1. 컬럼: key, name, application, status, owner, linked repo count, updated_at
2. 액션: 상세 보기, 상태 변경, repository 연결 관리로 이동

---

## 5) API / 데이터 연계 전략

MVP는 기존 API 재사용을 우선한다.

1. Applications: `GET /api/v0-1/platforms`
2. Repositories: `GET /api/v0-1/repositories`
3. Projects:
- Application 맥락 목록: `GET /api/v0-1/platforms/:platform_id/projects`
- 필요 시 후속으로 Admin 전용 집계 API 검토

주의:
- 현재 project 목록 수집은 API 모델 혼합(hybrid) 영향이 있으므로 프론트 aggregation 정책을 명시
- 성능 이슈 발생 시 Admin 전용 검색 API(서버 집계)로 2차 확장

---

## 6) 구현 단계

### Phase 1. 라우트/권한/메뉴 골격

1. `/admin/catalog` 페이지 생성
2. `system_admin` 접근 가드 연결
3. 사이드바 `Catalog` 메뉴 추가

완료 기준:
- admin만 진입 가능
- 비-admin redirect 정상

### Phase 2. 3탭 목록 MVP

1. Applications / Repositories / Projects 탭 렌더링
2. 각 탭 데이터 로딩 + 공통 상태 UI
3. 검색/필터 기본 동작

완료 기준:
- 시스템 관리자가 3 도메인 자산을 단일 화면에서 조회 가능

### Phase 3. 관리 액션 연결

1. 상세 이동/편집 진입/상태 변경 액션 연결
2. App-Repo-Project 관계 drill-down 링크 연결

완료 기준:
- 조회뿐 아니라 주요 운영 액션 수행 가능

### Phase 4. 검증/E2E

1. RBAC E2E: non-admin 차단
2. 기본 조회 E2E: 3탭 로딩/검색/필터
3. 회귀: 기존 admin settings/app/project/repo 흐름 영향 점검

---

## 7) 리스크 및 대응

1. 프로젝트 목록 API 분산/중복 호출로 인한 성능 저하
- 대응: 초기 pagination + debounce, 후속 Admin 집계 API 검토

2. 기존 Admin Settings와 기능 중복
- 대응: Catalog는 운영 조회 중심, Settings는 정책/구성 변경 중심으로 역할 분리

3. 권한 누락으로 인한 노출 리스크
- 대응: 라우트 가드 + 서버 API RBAC 재확인 + E2E RBAC 케이스 추가

---

## 8) 테스트 전략

1. 단위
- 탭 전환, 필터 상태 계산, 검색 파서

2. 통합(UI)
- 로딩/에러/빈 상태
- 각 탭 리스트 렌더링

3. E2E
- `system_admin`: `/admin/catalog` 접근/조회/탭 전환 성공
- non-admin: `/admin/catalog` 접근 시 default landing redirect

---

## 9) 산출물 목록

1. `frontend/app/(dashboard)/admin/catalog/page.tsx` (신규)
2. `frontend/components/admin/catalog/*` (신규)
3. `frontend/components/layout/Sidebar.tsx` (메뉴 추가)
4. `frontend/tests/e2e/admin-catalog.spec.ts` (신규)
5. 필요 시 서비스 레이어 보강 (`project.service.ts` / `repository.service.ts`)

---

## 10) 의사결정 요약

- 채택안: `Admin Catalog` 신규 메뉴 (옵션 B)
- 이유: 운영 동선 분리, 확장성, 권한 명확성
- 구현 원칙: 기존 API 최대 재사용 → 성능 병목 시 Admin 집계 API로 점진 확장

