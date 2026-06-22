---
title: README
type: source
tags: [infrastructure, core, project-devhub]
sources: [raw/projects/devhub/docs/infrastructure/README.md]
git_commit: baf1cf24
git_branch: main
version_system: v0.1.1-alpha
version_workflow: v0.5.11-beta
last_touched: 2026-06-22T03:37:45Z
mirror_dirty: |
related: [none]
status: draft
contradictions: [none]
---

# Infrastructure Layer SDLC 문서

- 문서 목적: `docs/governance/code-taxonomy.md` §2.3 Infrastructure 레이어 (외부 기술 및 연동 구현체) 의 SDLC 문서 진입점.
- 범위: 외부 시스템 연동 + 운영 인프라 구체 구현 레이어.
- 상태: draft (Phase 1 골격)
- 최종 수정일: 2026-05-29
- 관련 문서: [code-taxonomy.md §2.3](../governance/code-taxonomy.md)

## 1. Infrastructure 모듈 index

| 모듈 | 코드 위치 | 주 책임 | 관련 ADR |
|---|---|---|---|
| keycloak-idp | `backend-core/internal/auth/keycloak_verifier.go`, `infra/idp/keycloak-event-listener-spi/`, `infra/idp/sql/` | Keycloak IdP 연동 (JWKS + admin client + event listener SPI) | ADR-0019, ADR-0020, ADR-0022, ADR-0023 |
| gitea-scm | `backend-core/internal/infrastructure/gitea/`, `backend-core/internal/normalize/gitea/` | Gitea API 클라이언트 + 백그라운드 sync 워커 + webhook 서명 검증 + JSON 정규화 | ADR-0003 |
| hrdb | `backend-core/internal/infrastructure/hrdb/` (`postgres.go`, `mock.go`) | 인사망 데이터 어댑터 (실 PG / mock) | ADR-0008, ADR-0010 |
| commandworker | `backend-core/internal/infrastructure/commandworker/{worker,live_worker}.go`, `backend-core/internal/infrastructure/serviceaction/executor.go` | 실시간 인프라 명령어 폴링/실행 에이전트 + sandbox | — |
| database-migration | `backend-core/migrations/000001~000046_*.sql` | golang-migrate SQL 마이그레이션 전체 (~46 file) | — |
| deployment-automation | `scripts/`, `infra/nginx/`, `docker-compose.{deploy,}.yml` | 배포 전처리 + Nginx 역프록시 + compose 구성 | ADR-0018 |

## 2. SDLC 문서

| 단계 | 위치 | 상태 |
|---|---|---|
| README 진입점 | 본 파일 | active |
| Setup / Operations | `docs/setup/` (mixed — keycloak_operations.md, test-server-deployment.md, environment-setup.md 등) | active (Phase 2 일부 이관 후보) |
| Concept / Design | `docs/planning/` (keycloak_*, single_port_reverse_proxy.md, prometheus_homelab_alerts.md 등) | Phase 2 이관 예정 → `docs/infrastructure/<영역>/` |
| TC | `docs/infrastructure/commandworker/test_cases.md` | Phase 2 rename 후보 → `docs/infrastructure/commandworker/test_cases.md` |

## 3. 호출 규칙 ([architecture.md §2.2](../architecture.md))

> `Infrastructure` 레이어는 `Domain` 레이어의 구체 비즈니스 서비스나 엔티티를 직접 소유하거나 지배하지 않습니다. `Domain`은 외부 연동 대상에 대해 추상화된 어댑터 인터페이스만 노출하며, `Infrastructure`는 이 인터페이스의 기술 구현체로만 작동합니다.

상향 호출 금지 — Infrastructure 가 Domain 의 interface 를 구현하되, Domain 의 concrete 비즈니스 로직을 import 하지 않는다.

## 4. cross-cutting 참조

- [code-taxonomy.md §2.3](../governance/code-taxonomy.md)
- [Domain index](../domain/README.md)
- [Shared index](../shared/README.md)
- ADR-0003 (no-docker CI policy)
- ADR-0018 (single-port reverse proxy)
- ADR-0019 (Keycloak 단일화)
