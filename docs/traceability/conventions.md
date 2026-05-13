# 추적성 ID 컨벤션

- 문서 목적: DevHub 의 SDLC 단계 별 항목에 부여하는 추적 ID 의 명명 규칙·영속성·발급 절차를 정의한다.
- 범위: 요구사항 → 설계 → 로드맵 → 구현 → 단위테스트 → E2E 6 단계.
- 대상 독자: 모든 contributor (사람 + AI agent).
- 상태: accepted
- 최종 수정일: 2026-05-13
- 관련 문서: [`README.md`](./README.md) (체계 개요), [`sync-checklist.md`](./sync-checklist.md) (동기화 절차), [`report.md`](./report.md) (1차 종합 매트릭스).

## 1. ID 접두사 표

| 단계 | 접두사 | 형식 | 예시 |
| --- | --- | --- | --- |
| 요구사항 (functional) | `REQ-FR-` | `REQ-FR-{nn}` | `REQ-FR-01` |
| 요구사항 (non-functional) | `REQ-NFR-` | `REQ-NFR-{nn}` | `REQ-NFR-03` |
| 설계 (architecture) | `ARCH-` | `ARCH-{nn}` | `ARCH-02` |
| 설계 (API contract) | `API-` | `API-{nn}` | `API-07` |
| 로드맵 | `RM-` | `RM-M{0..3}-{nn}` | `RM-M2-04` |
| 구현 | `IMPL-` | `IMPL-{module}-{nn}` | `IMPL-auth-01`, `IMPL-rbac-03` |
| 단위테스트 | `UT-` | `UT-{pkg}-{nn}` | `UT-httpapi-05`, `UT-domain-02` |
| E2E | `TC-` | `TC-{feature}-{nn}` | `TC-AUTH-01`, `TC-NAV-02` |

- `{nn}` 은 2자리 이상 0-padded 십진수. 단계 내에서 단조 증가.
- `{module}` (구현) 은 backend-core 의 `internal/<pkg>` 하위 경로 또는 frontend 의 기능 영역명 (`auth`, `rbac`, `audit`, `account`, `org`, `command`, `realtime` 등) 을 kebab-case 로.
- `{pkg}` (단위테스트) 는 Go 의 경우 패키지 디렉터리명, frontend 의 경우 영역명.
- `{feature}` (E2E) 는 spec 파일과 일치하는 영역 토큰 (`AUTH`, `USR-CRUD`, `NAV`, `ORG-CRUD`, `AUD` 등).

## 2. 영속성 (Immutability)

- **부여된 ID 는 재사용·재번호하지 않는다.** 항목이 삭제되어도 ID 는 "deprecated" 로 마킹만 하고 ID 자체는 회수하지 않는다. 이는 추적 chain 의 깨짐을 방지한다.
- 새 항목은 단계 내 마지막 ID + 1 로 할당.
- ID 의 의미(가리키는 항목) 가 변경되면 새 ID 발급 후 기존 ID 를 deprecate.

## 3. 발급 절차

### 3.1 신규 항목 — 작성자가 발급

PR 작성자가 영향 받는 단계 별로 새 ID 를 정의 + `docs/traceability/report.md` 의 해당 인덱스 표 (각 단계의 인덱스 섹션) 와 추적성 매트릭스에 row 추가한다. `sync-checklist.md` 의 절차 참조.

### 3.2 기존 항목 변경 — ID 유지

기존 항목의 범위 / 구현 / 테스트 가 변경되어도 **동일 ID 유지**. 매트릭스의 cross-references (file:line, commit, PR 번호) 만 갱신.

### 3.3 충돌 회피

여러 PR 이 동시에 같은 ID 를 사용하지 않도록, 매트릭스 갱신은 **머지 직전** PR 의 마지막 commit 으로 한다. 충돌 시 나중 PR 작성자가 한 번 더 ID 갱신.

## 4. 매핑 정책 (Required Coverage)

매트릭스의 행 = 추적 chain. 각 chain 은 다음 매핑을 가능한 만족:

- `REQ-FR-*` 는 ≥ 1 `ARCH-*` 또는 `API-*` 와 매핑.
- 구현된 `REQ-FR-*` 는 ≥ 1 `IMPL-*` 와 매핑.
- 구현된 `REQ-FR-*` 는 ≥ 1 `UT-*` 또는 `TC-*` 와 매핑 (둘 중 하나로 검증).
- 모든 `TC-*` 는 ≥ 1 `REQ-FR-*` 또는 `REQ-NFR-*` 로 reverse-trace.

이 정책은 hard requirement 가 아닌 권장. 위반 시 매트릭스의 "gap" 섹션에 기록.

## 5. 명명 충돌 / Special cases

- 단계 내에서 동일 의미를 가리키는 항목이 둘 이상 발견되면 하나로 통합. 통합되는 ID 중 하나가 "primary", 나머지는 "alias" 로 매트릭스에 명기.
- 외부 의존성 (Ory Hydra, Kratos, PostgreSQL 등) 자체는 추적 항목이 아니다. 외부 의존성을 사용하는 DevHub 측 wiring (예: `BearerTokenVerifier`) 만 `IMPL-` 부여.
- ADR 은 별개 식별 체계 (`ADR-{nnnn}`) 를 가짐. 매트릭스 행에는 포함하지 않고 `report.md` 의 "ADR 인덱스" 서브 섹션에서 별도 link.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). |
