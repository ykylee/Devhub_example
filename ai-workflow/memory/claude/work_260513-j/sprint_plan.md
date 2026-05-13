# Sprint Plan — claude/work_260513-j (M3 진입 전 잔여 후속 일괄 처리)

- 문서 목적: M3 마일스톤 진입 전 잔여 후속 (open gap / carve out) 일괄 정리.
- 진입 base: main HEAD `cb9e6d5` (PR #95 직후).
- 최종 수정일: 2026-05-13
- 상태: in_progress

## 1. 작업 항목

| # | 항목 | 위치 | 규모 |
|---|------|------|------|
| D6 | inbound X-Devhub-Actor 명시 거부 (400) | `auth.go::authenticateActor` + 회귀 테스트 갱신 | S |
| ADR-0006 | D6 의 결정 ADR | `docs/adr/0006-x-devhub-actor-reject-inbound.md` | S |
| B2-2 | deprecated 추가 마킹 | `backend/requirements_review.md` 등 | S |
| 매트릭스 nit | §2.1 문구 + API-03 결손 명시 | `report.md` §2.1 | S |
| spec 절 신설 | account / users / organization | `backend_api_contract.md` §10.2 / §10.3 / §10.4 | M |
| AuthGuard smoke | C1 보완 (loading 상태 1 test) | `frontend/components/layout/AuthGuard.test.tsx` | S |
| TC-INFRA-RENDER-01 spec ts | C2 보완 (정적 렌더 1 spec) | `frontend/tests/e2e/infra-topology.spec.ts` | S+ |
| ADR-0007 | RBAC cache 일관성 결정 (구현 carve) | `docs/adr/0007-rbac-cache-multi-instance.md` | S |

## 2. carve out (M3 진입 후)

- C1 본격: AuthGuard / Header / Sidebar 의 mock-heavy 본격 Vitest
- C2 본격: TC-CMD-CREATE-01 / TC-CMD-STATUS-01 / TC-INFRA-NODE-CLICK-01 / TC-INFRA-GROUP-TOGGLE-01 spec ts
- RBAC cache 일관성 실제 구현 (ADR-0007 결정 따라)
- M3 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull)

## 3. 검증

- backend `go test ./...` PASS
- frontend `npm test` PASS + AuthGuard.test.tsx 신규
- e2e `npm run e2e` PASS (TC-INFRA-RENDER-01 추가)
- CI 4 잡 SUCCESS
