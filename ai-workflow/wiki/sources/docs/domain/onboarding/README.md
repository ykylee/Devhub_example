---
title: README
type: source
tags: [domain, README.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/onboarding/README.md]
git_commit: fb3894f7
git_branch: chore/260622-wiki-drift-cleanup-3
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:03:34Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# onboarding 도메인

- 문서 목적: `onboarding` 도메인의 SDLC 진입점.
- 범위: 인사 등록 후 최초 로그인하는 신규 사용자의 승인 절차, 온보딩 게이트 가드 및 초기 가이드.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.1.5](../../governance/code-taxonomy.md)

## 1. 도메인 정의

> 인사 등록 후 최초 로그인하는 신규 사용자의 승인 절차, 온보딩 게이트 가드 및 초기 가이드를 통제한다. ([code-taxonomy.md §2.1.5](../../governance/code-taxonomy.md))

`onboardingGate` 미들웨어 + `/onboarding` 페이지 + 관리자 review 워크플로우로 신규 사용자 진입을 통제한다.

## 2. 4 계층 모듈 매핑

| 계층 | Backend | Frontend |
|---|---|---|
| view | `backend-core/internal/domain/onboarding/view/` (`me_onboarding.go`, `onboarding_gate.go`, `onboarding_metrics.go`, `onboarding_pending_gauge.go`, `onboarding_roles.go`, `handler.go`) | `frontend/app/onboarding/`, `frontend/components/onboarding/OnboardingForm.tsx`, `OrganizationPicker.tsx` |
| service | 온보딩 게이트 통과 규칙 | `frontend/domain/onboarding/service/onboarding.service.ts` |
| repository | — (organization-management user 테이블 공유) | — |
| schema | 온보딩 폼 입력 필드 + 가벨 플래그 모델 | (frontend 내장) |

의존 도메인: [auth-session](../auth-session/), [organization-management](../organization-management/)

## 3. SDLC 문서 link

| 단계 | 위치 | 상태 |
|---|---|---|
| REQ | [`./requirements.md`](./requirements.md) | active (Phase 3 split, 2026-05-29) |
| ARCH | [`./architecture.md`](./architecture.md) | active (Phase 3) |
| API | [`./api.md`](./api.md) | active (Phase 3) |
| TC | `./test_cases.md` | planned (Phase 2 — `docs/domain/onboarding/test_cases.md`) |
| Concept | `./concept.md` | planned (Phase 2 — `docs/domain/onboarding/concept.md`) |
| Plan | `./impl_plan.md` | planned (Phase 2 — `docs/domain/onboarding/impl_plan.md`) |

## 4. 관련 ADR

- ADR-0021 (onboarding gate)

## 5. cross-cutting 참조

- `docs/backend_api_contract.md` §16 (onboarding API)
- `docs/architecture.md` §9 (onboarding)

## 6. E2E spec

- `frontend/tests/e2e/onboarding-first-login.spec.ts`
