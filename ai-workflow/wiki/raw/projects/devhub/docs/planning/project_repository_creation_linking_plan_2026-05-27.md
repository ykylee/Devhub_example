# Project/Repository 생성·연결 개선 개발 계획 (2026-05-27)

- 문서 목적: Project 독립 생성 + 선택적 Application 연계 + 선택적 Repository 동반 생성 + Project-Repository N:M 연결 UX 개선을 구현하기 위한 실행 계획을 정의한다.
- 범위: backend-core API/도메인/저장소 계층, frontend 생성/상세 UI, 테스트(E2E/통합/단위), 문서 동기화
- 대상 독자: 구현 담당자, 리뷰어, QA, 제품 의사결정자
- 상태: draft
- 최종 수정일: 2026-05-27
- 관련 문서: `docs/requirements.md` (Platform/Project 요구사항), `docs/architecture.md`, `docs/traceability/sync-checklist.md`, `frontend/tests/e2e/*`, `backend-core/internal/httpapi/projects.go`

## 1. 배경 및 문제 정의

현재 Project 생성/연결 흐름은 다음 한계가 있다.

1. Project 생성 시 Application 종속이 강해 독립 프로젝트 생성 시나리오가 제한적이다.
2. Project 생성 중 Repository를 함께 만들고 즉시 연결하는 흐름이 없다.
3. Project-Repository N:M 연결 UX가 추가 확장(다중 연결) 관점에서 직관성이 낮다.

목표는 "필수 최소 입력으로 Project를 먼저 만들고, 필요 시 Application/Repository를 유연하게 연결"하는 운영 UX로 전환하는 것이다.

## 2. 목표 상태 (To-Be)

1. Project는 Application 없이 독립적으로 생성 가능해야 한다.
2. Project 생성 시 Application 선택은 optional이어야 한다.
3. Repository 없이 Project 생성이 가능해야 한다.
4. 사용자가 원하면 Project 생성 과정에서 Repository를 동반 생성하고 즉시 연결할 수 있어야 한다.
5. Project 상세에서 연결된 Repository를 수평 리스트로 노출하고, `+` 버튼으로 다중 연결을 빠르게 추가할 수 있어야 한다.

## 3. 기능 요구사항 (구현 대상)

### 3.1 Project 독립 생성

- `platform_id`는 nullable/optional 허용.
- 프로젝트 기본 필드(`key`, `name`, `owner_user_id`, `status`, `visibility`)만으로 생성 가능.
- Application 선택 시 기존과 동일하게 association 유지.

### 3.2 Repository 동반 생성(옵션)

- Project 생성 모달에 `Repository 함께 생성` 토글 제공.
- 토글 활성 시 추가 입력:
  - `repository_key`
  - `repository_slug`
  - `scm_provider`
- 성공 시:
  1. Repository 생성
  2. 생성된 Repository를 Project에 자동 연결

### 3.3 Project-Repository N:M 연결 UX

- Project 상세에서 연결된 Repository를 수평 카드/칩 형태로 표시.
- `+` 버튼 클릭 시 연결 모달 열기.
- 한 번에 여러 Repository 선택 후 연결 가능.
- 이미 연결된 Repository는 선택 불가 또는 중복 방지 메시지 노출.

## 4. API/도메인 설계 방향

### 4.1 기존 API 확장 우선

- `POST /api/v1/projects` 또는 현재 사용 중인 프로젝트 생성 endpoint에서 `platform_id` optional 처리.
- `POST /api/v1/projects/:id/repositories`는 다중 호출/다중 선택 UX를 지원하도록 유지.

### 4.2 동반 생성 방식

옵션 A (권장): 단일 endpoint 확장
- Project 생성 payload에 `repository_create_payload` optional 포함.
- backend transaction으로 project create + repository create + project_repository link 일괄 처리.

옵션 B: 프런트 orchestration
- 프로젝트 생성 후 repository 생성 API 호출, 이후 link API 호출.
- 실패 시 rollback UX 필요(권장도 낮음).

본 계획에서는 일관성과 실패 복구 단순화를 위해 **옵션 A**를 1순위로 채택한다.

## 5. 데이터/검증 규칙

1. `project.key` 유일성 기존 정책 유지.
2. `repository.slug`/`repository.key` 유효성 및 provider별 제약 검증.
3. Project-Repository link는 composite unique 제약으로 중복 연결 방지.
4. 권한 검증:
- 생성/연결/해제는 `system_admin` 또는 정책상 허용 role만 수행.

## 6. UI/UX 설계

### 6.1 Project 생성 모달

- 섹션 1: Project 기본 정보(필수)
- 섹션 2: Application 선택(optional, searchable)
- 섹션 3: `Repository 함께 생성` 토글
  - OFF: Project만 생성
  - ON: Repository key/slug/provider 입력 UI 노출

### 6.2 Project 상세 페이지

- `Connected Repositories` 수평 리스트
- 우측 `+` 버튼으로 연결 모달 진입
- 모달에서 다중 선택 후 `Connect` 실행

## 7. 구현 단계 (Sprint Plan)

### Phase 1 — 설계/계약 확정

- API payload 변경안 확정
- 에러 코드/응답 규약 합의
- traceability 영향 범위 식별

### Phase 2 — Backend 구현

- projects handler/store 도메인에서 `platform_id optional` 반영
- `repository_create_payload` 처리 + transaction 도입
- project-repository link 중복/권한 에러 정리
- 단위/통합 테스트 추가

### Phase 3 — Frontend 구현

- Project 생성 모달 개편(optional application + repo toggle)
- Project 상세 연결 UI를 수평 + `+` 추가 플로우로 전환
- 다중 선택 연결 모달 구현

### Phase 4 — 테스트/회귀/문서

- E2E 시나리오 추가:
  - Project 독립 생성
  - Application 선택 생성
  - Repository 동반 생성
  - 다중 Repository 연결
- 문서/추적성 동기화

## 8. 테스트 전략

### 8.1 Backend

- handler 테스트:
  - `platform_id` 없는 생성 성공
  - 동반 생성 payload 성공/실패 분기
  - 중복 link conflict
- integration 테스트:
  - transaction 원자성 검증(중간 실패 시 rollback)

### 8.2 Frontend

- 컴포넌트/서비스 테스트:
  - 토글 ON/OFF payload 분기
  - 다중 연결 payload 검증
- Playwright E2E:
  - 주요 happy path + conflict path

## 9. 리스크 및 완화

1. 트랜잭션 복잡도 증가
- 완화: 생성/연결 책임을 store 계층 단일 함수로 캡슐화

2. 기존 Project 생성 플로우 회귀
- 완화: 기존 payload backward compatibility 유지 + 회귀 E2E 포함

3. 권한 drift
- 완화: `enforceRowOwnership` 및 route permission 재검토, 테스트 케이스 고정

## 10. 완료 기준 (Definition of Done)

1. Project를 Application/Repository 없이 생성 가능
2. Project 생성 시 선택적으로 Application 연결 가능
3. Project 생성 시 선택적으로 Repository 동반 생성 + 자동 연결 가능
4. Project 상세에서 Repository 다중 연결 UX(`+` 기반) 동작
5. backend 단위/통합 + frontend lint + E2E 시나리오 통과
6. 추적성 문서(sync-checklist/report) 갱신

## 11. 다음 작업 시작점

1. API 계약 초안 작성 (`projects.go`/store payload)
2. frontend modal 필드 설계안 반영
3. E2E 테스트 골격 먼저 작성 후 구현 병행
