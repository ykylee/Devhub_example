---
title: architecture
type: source
tags: [domain, architecture.md, project-devhub]
sources: [raw/projects/devhub/docs/domain/organization-management/architecture.md]
git_commit: 01f1969c
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-18T07:11:15Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# organization-management 도메인 아키텍처

- 문서 목적: `users` + `org_units` + `unit_appointments` 마스터 데이터 모델 및 HRDB lookup 어댑터 아키텍처를 정의한다.
- 범위: master `docs/architecture.md` 에 별도 § 가 없는 organization 도메인의 핵심 아키텍처 (ADR-0008 + ADR-0009 + ADR-0010 통합) 를 본 sub-document 에서 정리한다.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 아키텍처 검토자.
- 상태: draft (Phase 3 split)
- 최종 수정일: 2026-05-29
- 관련 문서: [도메인 README](./README.md), [requirements.md](./requirements.md), [api.md](./api.md), [기존 organizational_hierarchy_spec.md](./organizational_hierarchy_spec.md), [기존 org_chart_ux_spec.md](./org_chart_ux_spec.md), [master architecture](../../architecture.md), [ADR-0008](../../adr/0008-organization-model.md), [ADR-0009](../../adr/0009-organization-sync.md), [ADR-0010](../../adr/0010-hrdb-integration.md)

## 1. 컴포넌트 (ARCH-ORG-01)

```
┌────────────────────────────────────────────────┐
│  Go Core organization-management 도메인        │
│  ├── view: organization.go, organizations_     │
│  │         search.go, hr_lookup.go, handler.go │
│  ├── service: 조직 노드 유효성 (cycle 검사),   │
│  │            임원 할당 규칙                    │
│  └── repository: users_units.go                │
└────────────────┬───────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────┐
│  PostgreSQL                                    │
│  ├── users (000004 + onboarding 컬럼 §9.5)    │
│  ├── org_units (000004, parent FK,             │
│  │   leader_user_id FK)                        │
│  ├── unit_appointments (000019, 겸임)          │
│  ├── org_units_total_count_mv (000011, MV)     │
│  └── audit_logs (org_unit.*, user.* 액션)      │
└────────────────────────────────────────────────┘
                 │
                 ▼
┌────────────────────────────────────────────────┐
│  External HRDB (선택, ADR-0010)                │
│  └── ETL adapter — 직책/연락처 lookup          │
└────────────────────────────────────────────────┘
```

## 2. 데이터 모델 (ARCH-ORG-02)

상세 스키마와 제약은 [`./organizational_hierarchy_spec.md`](./organizational_hierarchy_spec.md) 가 source-of-truth. 본 §은 핵심만 정리.

### 2.1 `users`

- 기존 컬럼은 `docs/architecture.md` §6.2.1 에서 정의됨 (auth-session 도메인에서 인용).
- onboarding 신규 컬럼 (`onboarding_completed_at`, `review_status`) 은 onboarding 도메인 (ARCH-ONBOARD-05) 에서 정의됨.

### 2.2 `org_units`

```text
org_units
  unit_id           text  PK
  parent_unit_id    text  NULLABLE  FK org_units(unit_id)
  unit_type         text             -- team | department | org
  label             text
  leader_user_id    text  NULLABLE  FK users(user_id)
  position_x        int   NULLABLE  -- chart UI 좌표
  position_y        int   NULLABLE
  created_at, updated_at timestamptz
```

cycle 방지 검증은 PATCH 시점에 carve out (기존 `organizational_hierarchy_spec.md` §3).

### 2.3 `unit_appointments` (000019)

겸임/임명 매핑.

### 2.4 `org_units_total_count_mv` (000011)

materialized view — 조직별 누적 사용자 수 집계.

## 3. 사용자 검색/lookup (ARCH-ORG-03)

- 일반 사용자 조회: `GET /api/v1/users` (api.md API-33).
- onboarding 화면 검색: `GET /api/v1/organizations/search` (onboarding 도메인 API-84) — 권한 가드 없이 모든 사용자에게 모든 organization 후보 노출.
- HR lookup (직책/연락처): HRDB ETL adapter 경유 — ADR-0010, 운영 phase 진입 후 활성화.

## 4. Keycloak event 기반 user sync (ARCH-ORG-04)

ADR-0020 sub-carve C — Keycloak event listener 가 group/role/status 변경을 polling 후 DevHub `users.role`/`primary_unit_id`/`status` 갱신. 자세한 push/poll 흐름은 [audit-ops architecture](../audit-ops/architecture.md) 참조.

## 5. organization chart UI (ARCH-ORG-05)

좌표 영속화 + drag UX 는 [`./org_chart_ux_spec.md`](./org_chart_ux_spec.md) 의 spec 을 따른다.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/architecture.md` 에 별도 §이 없는 organization-management 도메인의 핵심 아키텍처 (기존 specs + ADR-0008/0009/0010 통합) 를 본 sub-document 로 정리. 기존 `organizational_hierarchy_spec.md` / `org_chart_ux_spec.md` / `backend_requirements_org_hierarchy.md` 는 detailed body 로 유지. ID는 ARCH-ORG-01..05 도메인 임시 발급. |
