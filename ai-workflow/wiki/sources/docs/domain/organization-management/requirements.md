---
title: requirements
type: source
tags: [domain, requirements.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/organization-management/requirements.md]
git_commit: 046e0c81
git_branch: chore/260622-wiki-drift-cleanup-4
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T06:22:35Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# organization-management 도메인 요구사항

- 문서 목적: 조직 계층(`org_units`)·사용자 마스터(`users`)·직무 Appointments·HR lookup 도메인의 요구사항을 정의한다.
- 범위: master `docs/requirements.md` 의 §2.5 사용자 master data 부분 (계정/credential 부분은 `docs/domain/auth-session/requirements.md` 로 분리) + 기존 `backend_requirements_org_hierarchy.md` 의 추적성 통합 진입점. onboarding self-service unit selection 은 `docs/domain/onboarding/requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: draft (Phase 3 split, 기존 `backend_requirements_org_hierarchy.md` 통합 진입점)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [기존 backend_requirements_org_hierarchy.md](./backend_requirements_org_hierarchy.md), [기존 org_chart_ux_spec.md](./org_chart_ux_spec.md), [기존 organizational_hierarchy_spec.md](./organizational_hierarchy_spec.md), [architecture.md](./architecture.md), [api.md](./api.md), [ADR-0008](../../adr/0008-organization-model.md), [ADR-0009](../../adr/0009-organization-sync.md), [ADR-0010](../../adr/0010-hrdb-integration.md)

## 1. 개요

본 도메인은 다음 마스터 데이터를 관리한다.

- **사용자(User) 마스터**: `users` row (`user_id`, `email`, `display_name`, `role`, `status`, `idp_subject`, `primary_unit_id`, `current_unit_id`, `is_seconded`, `joined_at` 등). credential lifecycle 은 Keycloak 책임 (auth-session 도메인).
- **조직 계층(`org_units`)**: parent-child 그래프, `unit_type`, `leader_user_id`, organization chart UI 좌표.
- **직무 Appointments**: 사용자-조직 간 겸임/임명 매핑 (`unit_appointments`, 000019).
- **HR lookup**: HRDB ETL 경유 사용자 직책/연락처 조회 어댑터.

## 2. 기능 요구사항 (high-level)

### 2.1 사용자 master CRUD

- **REQ-ORG-001 (MVP, 확정):** `system_admin` 은 사용자 master 를 생성/수정/조회/soft-delete 할 수 있어야 한다. credential 자체는 외부 IdP 가 관리한다 (auth-session 도메인 정책 정합).
- **REQ-ORG-002 (MVP, 확정):** `users.user_id` 는 시스템 전역 unique, `idp_subject` 는 OIDC `sub` 매핑. role 은 Keycloak event listener 가 자동 sync (rbac-permissions 정합).

### 2.2 조직 계층

- **REQ-ORG-003 (MVP, 확정):** 조직 계층은 parent-child 그래프로 관리하며 cycle 을 허용하지 않는다 (PATCH 시 cycle 검증 carve out — 기존 `organizational_hierarchy_spec.md` §3).
- **REQ-ORG-004 (MVP, 확정):** unit 의 삭제는 자식 unit 또는 members 가 존재하면 422 (`has_children` / `has_members`) — cascade 미지원 1차 정책.
- **REQ-ORG-005 (MVP, 확정):** unit 의 좌표(`position_x`, `position_y`) 는 organization chart UI 의 drag 위치를 영속화. 자세한 UX 는 [`./org_chart_ux_spec.md`](./org_chart_ux_spec.md) 참조.

### 2.3 직무 Appointments + 겸임 정책

- **REQ-ORG-006 (후속):** 사용자의 primary_unit_id 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수 등) 은 별도 결정 — 기존 [`./backend_requirements_org_hierarchy.md`](./backend_requirements_org_hierarchy.md) §1·2 의 미해결 항목.

### 2.4 HR lookup

- **REQ-ORG-007 (후속, ADR-0010):** HRDB ETL adapter 를 통한 사용자 직책/부서/연락처 조회 — 외부 HRDB scheme 정착 후 활성화.

## 3. 사용자 master data 운영 규칙 (auth-session 도메인 정합)

본 도메인이 `users` row 의 source-of-truth 이다. 다만 비밀번호/credential/세션 관리는 Keycloak 이 책임지며 (auth-session 도메인), DevHub 는 다음만 관리한다.

- `users` (프로필 + 권한 메타 + onboarding 상태)
- `org_units` + `unit_appointments` (조직 마스터 + 겸임)
- `org_units_total_count_mv` (000011, materialized view)

## 4. 추적성 진입점

- 기존 매트릭스 보존: [`./backend_requirements_org_hierarchy.md`](./backend_requirements_org_hierarchy.md), [`./organizational_hierarchy_spec.md`](./organizational_hierarchy_spec.md), [`./org_chart_ux_spec.md`](./org_chart_ux_spec.md) 가 본 sub-document 의 source-of-truth 로 유지된다.
- Phase 3 split 의 의도는 도메인 README 의 §3 SDLC link 표가 본 `requirements.md` 를 가리키고, 그 아래 기존 3 file 이 detailed body 로 cross-link 되는 구조이다.

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` 의 user master 관련 요구사항(§2.5 사용자 master 부분 — credential 은 auth-session) 을 도메인 sub-document 로 재집합. 기존 `backend_requirements_org_hierarchy.md` / `organizational_hierarchy_spec.md` / `org_chart_ux_spec.md` 는 detailed body 로 유지. ID는 REQ-ORG-001..007 도메인 임시 발급(기존 추적성 매트릭스와 별도 — Phase 4 재구성 시 정합). |
