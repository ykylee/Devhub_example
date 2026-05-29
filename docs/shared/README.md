# Shared Layer SDLC 문서

- 문서 목적: `docs/governance/code-taxonomy.md` §2.2 Shared 레이어 (공통 기능 영역) 의 SDLC 문서 진입점.
- 범위: 비즈니스 도메인에 결합되지 않는 시스템 전반 유틸리티/설정/공통 UI 레이어.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.2](../governance/code-taxonomy.md)

## 1. Shared 모듈 index

| 모듈 | 코드 위치 (Backend / Frontend) | 주 책임 |
|---|---|---|
| config | `backend-core/internal/shared/config/` / `frontend/.env` 매핑, `frontend/shared/config/{endpoints,mock-ui}.ts` | 전역 환경 설정 로더 |
| logger | (system 표준 logger 어댑터) | 시스템 표준 로그 수집 |
| utils | `backend-core/internal/shared/httphelp/` / `frontend/shared/utils.ts`, `frontend/shared/utils/{last-build,lifecycle-status}.ts` | 공통 유틸리티 헬퍼 |
| ui-foundation | — / `frontend/shared/ui-foundation/{components,layout}/` | 프론트엔드 공통 UI + 레이아웃 (Modal, Badge, Toast, PageState, FilterBar, ComboBox, DestructiveConfirmModal, Header, Sidebar, AuthGuard) |
| integrationcaps | `backend-core/internal/shared/integrationcaps/` (PR #409) | provider capability gate OR semantics 공용 helper |

## 2. SDLC 문서

| 단계 | 위치 | 상태 |
|---|---|---|
| README 진입점 | 본 파일 | active |
| Tech stack / Project profile | [`tech_stack.md`](./tech_stack.md), [`PROJECT_PROFILE.md`](./PROJECT_PROFILE.md), [`LOCAL_MACHINE_ENV.md`](./LOCAL_MACHINE_ENV.md) | active |
| ui-foundation design system | (별도 sprint — frontend UI cleanup, PR #248 등) | partial |

## 3. 호출 규칙 ([architecture.md §2.2](../architecture.md))

> `Shared` 레이어의 컴포넌트는 비즈니스 도메인의 특정 상태나 의미론에 의존하지 않고, 항상 중립적이고 재사용 가능한 유틸리티 성격을 유지해야 합니다.

Domain 또는 Infrastructure 가 Shared 를 활용하며, **Shared 가 Domain/Infrastructure 를 import 하는 것은 금지** (역방향 호출 금지).

## 4. cross-cutting 참조

- [code-taxonomy.md §2.2](../governance/code-taxonomy.md)
- [Domain index](../domain/README.md)
- [Infrastructure index](../infrastructure/README.md)
