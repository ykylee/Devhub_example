# DevHub 통합 개발 로드맵

- 문서 목적: DevHub 프로젝트의 전체 개발 방향을 단일 진입점에서 정리한다. 백엔드·프론트엔드·인증/IdP·운영 트랙이 동일 마일스톤 체계 위에서 진행되도록 하는 1차 참조 문서.
- 범위: 머지된 PR #12 이후 시점부터 다음 단계 작업의 마일스톤·우선순위·의존 관계. 트랙별 *세부* 작업은 각 트랙의 세부 로드맵에서 관리.
- 대상 독자: 프로젝트 리드, 백엔드/프론트엔드 개발자, 운영 담당자, 후속 작업자
- 상태: draft (2026-05-08 신규 작성)
- 최종 수정일: 2026-05-08
- 관련 문서:
  - 백엔드 세부 로드맵: [`ai-workflow/memory/backend_development_roadmap.md`](../ai-workflow/memory/backend_development_roadmap.md)
  - 프론트엔드 세부 로드맵: [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md)
  - 요구사항: [`./requirements.md`](./requirements.md)
  - 시스템 설계: [`./architecture.md`](./architecture.md)
  - API 계약: [`./backend_api_contract.md`](./backend_api_contract.md)
  - 인증 ADR: [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md)
  - 보안 리뷰: [`../ai-workflow/memory/codebase-security-review-2026-05-08.md`](../ai-workflow/memory/codebase-security-review-2026-05-08.md)

---

## 0. 사용 가이드

본 문서는 **모든 개발자가 작업 전 가장 먼저 읽는** 단일 진입점이다.

1. 본인 트랙(백엔드/프론트엔드)의 다음 마일스톤이 무엇인지 §3 마일스톤 표에서 확인한다.
2. 마일스톤 안의 작업 항목은 §4 트랙별 세부 표에서 출처·의존을 확인한다.
3. *그 작업의 구현 디테일* 은 백엔드/프론트엔드 세부 로드맵에서 관리한다 — 본 문서는 일정·우선순위·교차 의존만 다룬다.
4. 본 문서가 다른 문서(요구사항, 설계, ADR)와 충돌하면 §6 충돌 해소 표를 source-of-truth 로 본다.
5. 신규 결정이 생기면 §7 변경 이력에 한 줄 추가한다.

세부 로드맵과 본 통합 로드맵의 역할 분담:

| 본 문서 (통합) | 세부 로드맵 (트랙별) |
| --- | --- |
| 마일스톤 정의·일정·의존 | 작업 단위 세부, 코드 위치, 검증 절차 |
| 트랙 간 contract / 충돌 해소 | 트랙 내부 phase 진행 |
| 우선순위(P0~P3)·완료 정의 | DoD 의 구현 디테일 |

---

## 1. 트랙 정의

| 트랙 | 책임 영역 | 세부 로드맵 |
| --- | --- | --- |
| **B / Backend** | Go Core API, store, normalize, command worker, realtime hub | [`ai-workflow/memory/backend_development_roadmap.md`](../ai-workflow/memory/backend_development_roadmap.md) |
| **F / Frontend** | Next.js (대시보드, 조직, 인증 UI, 실시간 통합, RBAC UI) | [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md) |
| **A / Auth & IdP** | Ory Hydra + Kratos, 토큰 검증, 권한 가드. ADR-0001 결정 사항 | [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) |
| **X / Cross / Contract** | API 계약, 메시지 envelope, role wire format, 데이터 모델 | [`./backend_api_contract.md`](./backend_api_contract.md), [`./architecture.md`](./architecture.md) |
| **O / Operations** | 환경 셋업, 배포, 운영 모니터링, 데이터 보존 worker | [`./setup/environment-setup.md`](./setup/environment-setup.md) |
| **AI** | Python AI service (gRPC), Gardener suggestion, weekly report worker | [`./backend/requirements_review.md`](./backend/requirements_review.md) §3-P3 |

> AI 트랙은 P3 이전에는 capacity 외 — §3 의 M3·M4 에서 다룬다.

---

## 2. 우선순위 분류 정책

본 로드맵은 4단계 우선순위를 사용한다. 백엔드 세부 로드맵의 P0~P3 / 보안 리뷰의 P0~P3 / `requirements_review.md` 의 P1~P3 / `frontend_integration_requirements.md` 의 P1~P3 모두를 동일 키로 매핑한다.

| Priority | 의미 | 처리 시점 |
| --- | --- | --- |
| **P0** | 머지된 코드의 보안/통합 결함, 다음 모든 작업의 전제 | M0 ASAP |
| **P1** | 핵심 기능의 contract·실행 경계 정합성. 사용자에게 보이는 결과물 | M1 ~ M2 |
| **P2** | 실시간 확장, RBAC UI 고도화, 운영 강화 | M2 ~ M3 |
| **P3** | 외부 연동(Gitea REST, AI gRPC, SSO), 후속 phase | M3 ~ M4 |

`requirements.md` §1.1 의 *확정 / 후보 / MVP 이후* 라벨도 본 문서에서는 다음으로 매핑:

| requirements 라벨 | 통합 로드맵 분류 |
| --- | --- |
| 확정 | M0~M2 의 must |
| 후보 | M2~M3 의 should — 마일스톤 진입 시 should/could 재분류 |
| MVP 이후 | M3~M4 의 could — 일정 미정 |

---

## 3. 마일스톤 (M0 ~ M4)

5개 마일스톤으로 분해. 각 마일스톤은 *완료 정의(DoD)* 가 통과해야 다음 마일스톤이 시작 가능하다 (직렬 강제는 아니지만, 병렬 실행 시 의존 위반은 §3.6 표로 관리).

### 3.1 M0 — 보안 게이트 통과 (인증/권한 정상화)

- **목표**: PR #12 머지 후 코드베이스가 *실제로* 인증·권한이 동작하는 상태로 만든다. 다른 모든 마일스톤의 전제.
- **우선순위**: P0
- **마감**: ASAP (다음 sprint)
- **트랙**: A 주도, B / F / X 협력

#### DoD

1. ✅ **A**: `BearerTokenVerifier` 실 구현체 (Hydra introspection) — `internal/auth/hydra_introspection.go` (PR #17)
2. ✅ **A·B**: prod 모드 verifier nil → startup 거부 — `Config.Validate(hasVerifier)` (PR #17)
3. ✅ **B**: empty Authorization 분기 제거, `/api/v1/integrations/gitea/webhooks` 만 화이트리스트 (PR #15)
4. ✅ **B**: `requireMinRole` 미들웨어 + 10 보호 라우트 매핑 — `authz.go` (PR #16)
5. ✅ **B**: `requestActor` 의 `X-Devhub-Actor` fallback 코드 path 자체 제거 (PR #19, SEC-4 close)
6. ✅ **F**: `AuthGuard.tsx` mock 제거 → `/api/v1/me` 호출, `/login` Hydra OIDC redirect (PR #18)
7. ✅ **B·F**: 통합 테스트 매트릭스 — `auth_test.go`, `authz_test.go`, `me_test.go`, `config_test.go`, `auth/hydra_introspection_test.go` 누적 (PR #15·16·17·18·19)
8. ⏳ **A**: ADR-0001 §9 Phase 1 운영 검증 — 사용자 환경 의존 (T-M0-10, infra/idp/ scaffold 동작 검증)
9. ✅ 코드 내 `// SECURITY (SEC-1|SEC-2|SEC-3|SEC-4)` 마커 모두 제거 — backend-core/frontend 코드 영역 0건 (memory/문서 영역만 잔존, 의도)

**M0 sprint 종료: 2026-05-08.** DoD #8 (운영 검증) 만 사용자 환경에 의존하며 코드 측 모든 항목 resolved.

#### 입력 자료

- 보안 리뷰: [`PR-12-security-review.md`](../ai-workflow/memory/PR-12-security-review.md), [`codebase-security-review-2026-05-08.md`](../ai-workflow/memory/codebase-security-review-2026-05-08.md)
- SEC backlog: [`../ai-workflow/memory/claude/test/backend-integration/backlog/2026-05-08.md`](../ai-workflow/memory/claude/test/backend-integration/backlog/2026-05-08.md)
- ADR: [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md)

### 3.2 M1 — 핵심 기능 contract 정합성

- **목표**: command/audit/RBAC 의 *실행 경계* 와 wire contract 가 코드와 일치. snapshot/도메인/조직 API 의 응답 스키마가 안정.
- **우선순위**: P0~P1
- **마감**: M0 종료 후 ~2 sprint
- **트랙**: B 주도, X / F 협력

#### DoD

1. **B·X**: 모든 API 가 `backend_api_contract.md` 의 envelope/필드/role wire(`developer|manager|system_admin`)와 일치. 테스트로 강제.
2. **B**: command lifecycle 상태값(`pending|running|succeeded|failed|rejected|cancelled`) 일관 적용. dry-run/live 경계 테스트 고정.
3. **B**: Audit log auth actor 보강 — `source_ip`, `request_id`, 인증 출처(`source_type`).
4. **B**: `auth_test.go` 에 prod 가드 + role 가드 통합 테스트 추가.
5. **B**: DB 5xx 응답 마스킹(`writeServerError` 헬퍼) — SEC-5. 별도 PR 가능.
6. **B**: RBAC policy 편집 API (12.x) — write/audit 경계 + persistence (또는 *static-default 유지* 결정 명시).
7. **F**: `frontend/lib/services/types.ts` UI 타입 vs API wire 타입 분리. 표시 문자열 포맷팅을 프론트로 이전.
8. **X**: WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합.

#### 입력 자료

- [`./backend_api_contract.md`](./backend_api_contract.md), [`./backend/requirements_review.md`](./backend/requirements_review.md), [`./backend/frontend_integration_requirements.md`](./backend/frontend_integration_requirements.md)

### 3.3 M2 — 사용자 경험 정합 (Phase 4·5 잔여 + Phase 6/6.1)

- **목표**: 사용자에게 보이는 admin/RBAC/계정/실시간 흐름이 일관 동작.
- **우선순위**: P1~P2
- **마감**: M1 와 일부 병렬 가능, M0 종료 의존
- **트랙**: F 주도, B / A 협력

#### DoD

1. **F**: Phase 4 잔여 — command status transition WebSocket toast/status UI, AI Gardener suggestion API/UI 연결 범위 확정.
2. **F·A**: Phase 5 잔여 — `account.service.ts` 신설, `/login` 인증 가드 미들웨어, `/account` 비밀번호 변경 폼. 자체 accounts API 호출 *없이* Hydra OIDC + Kratos public flow + `/api/v1/admin/identities/*` 호출로 구현.
3. **F**: Phase 6 — PermissionEditor 복구.
4. **F·B**: Phase 6.1 — 정책 CRUD 연동, RBAC Guard 실체화.
5. **B·A**: `/api/v1/admin/identities/*` Kratos admin wrapper(GET/POST/PATCH/recovery-link/DELETE) 5종 구현.
6. **B·A**: Kratos identity webhook 수신 + `users.status` 동기화 + audit (`identity.created`, `identity.disabled`, `identity.recovery_link_created`, `auth.login.succeeded/failed`).
7. **X**: `UserRole` UI 표시명 vs API wire 분리 코드/타입 정리.

#### 입력 자료

- [`./frontend_development_roadmap.md`](./frontend_development_roadmap.md) Phase 4·5·6
- [`./backend_api_contract.md`](./backend_api_contract.md) §11 (Hydra/Kratos)
- [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) §9 Phase 2

### 3.4 M3 — Realtime 확장 + 외부 연동 1차

- **목표**: 실시간 채널이 인프라/CI/리스크/알림 전반을 커버. Gitea Hourly Pull 동기화. AI Gardener 1차.
- **우선순위**: P2~P3
- **마감**: M1·M2 종료 후
- **트랙**: B 주도, F / AI 협력

#### DoD

1. **B**: WebSocket publish 추가 — `infra.node.updated`, `infra.edge.updated`, `ci.run.updated`, `ci.log.appended`, `risk.critical.created`, `risk.updated`, `notification.created`.
2. **B**: Reconnect/replay 전략, role 기반 subscription filtering.
3. **X**: `ci.log.appended` 페이로드 (`ansi_text`, `line_number`, `stream_type`) 확정.
4. **B**: Gitea REST client + Hourly Pull reconciliation worker (Phase 10). dry-run + idempotency.
5. **B**: commits 정규화 테이블 도입 여부 결정 (ADR 형태로 기록).
6. **AI**: Python AI gRPC 서버 + Go Core client 연결 (Phase 9). 통합 테스트.
7. **AI·B**: AI Gardener suggestion → command/audit/risk 연결 1차.

#### 입력 자료

- [`../ai-workflow/memory/backend_development_roadmap.md`](../ai-workflow/memory/backend_development_roadmap.md) Phase 9·10
- [`./backend_api_contract.md`](./backend_api_contract.md) §8 (WebSocket)

### 3.5 M4 — 운영 / SSO / MFA / 후속 ADR

- **목표**: 프로덕션 운영 수준의 SSO·MFA·관측·데이터 보존.
- **우선순위**: P3
- **마감**: M3 이후, 일정 미정 (capacity 확보 후)
- **트랙**: A / O 주도

#### DoD

1. **A**: Gitea SSO 통합 (RBAC Phase 4) — 별도 ADR-0002 작성 후 진행.
2. **A**: MFA 정책 결정 + Kratos 설정 (현재는 1차 미도입).
3. **O**: OS 별 service wrapper 결정·문서화 (NSSM / sc / systemd).
4. **O**: 데이터 보존 자동 worker (운영 로그 1개월, Kudos 등 활성+1개월 후 영구 삭제).
5. **B**: System Admin 기능 고도화 (Phase 11) — Runner adapter, allowlist.
6. **AI**: Weekly report 생성 worker, AI Gardener 고도화, notification/focus mode 영속화.
7. **F**: Phase 7 통합 검증 + 전역 audit 연동.

#### 입력 자료

- [`./adr/0001-idp-selection.md`](./adr/0001-idp-selection.md) §8 미해결 항목 결정
- [`./requirements.md`](./requirements.md) §4·§5 (보존 정책)
- [`./architecture.md`](./architecture.md) §4.3

### 3.6 의존 관계 매트릭스

| 의존 → | M0 | M1 | M2 | M3 | M4 |
| --- | --- | --- | --- | --- | --- |
| **M0** | — | (M1 의 모든 인증·권한 의존) | (M2 의 모든 토큰 발급/가드 의존) | (M3 의 publish 권한 필터링 의존) | (M4 의 SSO/MFA 의존) |
| **M1** | | — | (M2 의 contract/role wire 의존) | (M3 의 envelope·command lifecycle 의존) | (M4 의 audit 보강 의존) |
| **M2** | | | — | (M3 의 RBAC 가드 UI 흐름 의존 — 일부) | (M4 의 사용자 흐름 의존) |
| **M3** | | | | — | (M4 의 realtime publish 분류 의존 — 일부) |

병렬 가능 영역: M1·M2 일부 (B 가 M1, F 가 M2 잔여 동시 진행), M3 의 AI 트랙은 M2 와 병렬 가능 (AI 단독).

---

## 4. 트랙별 세부 (마일스톤 매핑)

### 4.1 Backend (B)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Auth bypass 결함 해소 (4 결함 통합 — SEC-1~4) | M0 | P0 | codebase-security-review §3 |
| Role 가드 미들웨어 + 5 라우트 매핑 (SEC-3) | M0 | P0 | 동상 |
| `X-Devhub-Actor` fallback prod 제거 (SEC-4) | M0 | P0 | 동상 |
| `auth_test.go` prod 가드 + role 가드 통합 테스트 | M0~M1 | P0 | 동상 §2 권고 5 |
| Audit log actor 보강 (`source_ip`, `request_id`) | M1 | P1 | backend_roadmap §5 P0 |
| Service action approval 모델 + executor adapter | M1 | P1 | backend_roadmap §5 P1 |
| Live command 기본 거절·approval 강제 | M1 | P1 | 동상 |
| `/api/v1/admin/identities/*` Kratos admin wrapper | M2 | P1 | api_contract §11.4 |
| Kratos identity webhook 수신 + audit | M2 | P1 | api_contract §11.6 |
| RBAC policy 편집 API (12.x) — 또는 static 유지 결정 | M1~M2 | P1 | api_contract §12 |
| WebSocket publish 확장 (infra/ci/risk/notification) | M3 | P2 | api_contract §8 |
| Reconnect/replay + role subscription filtering | M3 | P2 | backend_roadmap §5 P2 |
| Gitea REST client + Hourly Pull reconciliation | M3 | P3 | backend_roadmap §5 P3 |
| Python AI gRPC client | M3 | P3 | backend_roadmap §5 P3 |
| Runner adapter / Allowlist (Phase 11) | M4 | P3 | backend_roadmap §2 Phase 11 |
| DB 5xx 응답 마스킹 (`writeServerError`) — SEC-5 | M1 (별도 PR) | P2 | codebase-security-review §5.1 |

### 4.2 Frontend (F)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| `AuthGuard.tsx` mock auth → 실토큰 검증 (SEC-1) | M0 | P0 | PR-12-review-actions SEC-1 |
| `/login` 인증 가드 미들웨어 (Kratos public flow) | M0~M2 | P0~P1 | frontend_roadmap Phase 5 잔여 |
| `account.service.ts` 신설 (Kratos public 흐름 호출) | M2 | P1 | 동상 |
| `/account` 비밀번호 변경 폼 | M2 | P1 | 동상 |
| Phase 4 잔여 — command status WebSocket toast/UI | M2 | P1 | frontend_roadmap Phase 4 |
| Phase 6 — PermissionEditor 복구 | M2 | P2 | frontend_roadmap Phase 6 |
| Phase 6.1 — 정책 CRUD 연동, RBAC Guard 실체화 | M2 | P2 | 동상 |
| `types.ts` UI vs API wire 분리 + 표시 포맷 프론트 이전 | M1 | P1 | frontend_integration §4 |
| AI Gardener suggestion 연결 범위 확정 | M3 | P3 | frontend_roadmap Phase 4 잔여 |
| Phase 7 — 통합 검증, 전역 audit | M4 | P3 | frontend_roadmap Phase 7 |
| Notification/focus mode 영속화 | M4 | P3 | frontend_integration §3.1·P3 |

### 4.3 Auth & IdP (A)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Hydra/Kratos PoC binary + schema 분리(`hydra`,`kratos`) | M0 | P0 | ADR-0001 §8-1·9 |
| First-party silent consent client 등록 | M0 | P0 | ADR-0001 §8-2 |
| `BearerTokenVerifier` 실 구현체 + `main.go` 주입 | M0 | P0 | ADR-0001 §9-7, codebase-security-review §3 |
| Kratos identity admin wrapper (5 endpoints) | M2 | P1 | api_contract §11.4, ADR-0001 §9-4 |
| Identity webhook 수신 + audit 매핑 | M2 | P1 | api_contract §11.6, ADR-0001 §9-3·5 |
| `X-Devhub-Actor` 단계적 폐기 마이그레이션 | M0~M1 | P0~P1 | ADR-0001 §8-4 |
| Gitea SSO 통합 (RBAC Phase 4) — 별도 ADR-0002 | M4 | P3 | ADR-0001 §8-5 |
| MFA 정책 결정 + 설정 | M4 | P3 | ADR-0001 §8-3 |
| OS service wrapper 결정 (NSSM/sc/systemd) | M4 | P3 | ADR-0001 §8-7 |

### 4.4 Cross / Contract (X)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| WebSocket envelope 표준화 (`schema_version`, `type`, `event_id`, `occurred_at`, `data`) | M1 | P1 | api_contract §8 |
| Role wire format (`developer|manager|system_admin`) 일괄 강제 | M1 | P1 | api_contract §2, requirements_review §3.3 |
| Command lifecycle 상태값(6) 표준 | M1 | P1 | api_contract §2 |
| `ci.log.appended` 페이로드 확정 | M3 | P2 | frontend_integration §6.1 |
| AccountStatus invariant 보존 (구현은 Kratos) | M0~M2 | P0 | requirements §2.5, ADR-0001 §5 |
| `requirements.md` / `architecture.md` 의 자체 accounts 흔적 → 정책 invariant 라벨링 | M0 (문서 작업) | P0 | §6 충돌 해소 |

### 4.5 Operations (O)

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Docker / native 운영 가이드 (완료) | — | — | environment-setup.md |
| `tech_stack.md` / `PROJECT_PROFILE.md` 의 docker default 표현 → 분기 안내 정리 | M0 (문서) | P1 | §6 충돌 해소 |
| 사내 Go module / npm / font mirror 가이드 보강 | M1 | P1 | environment-setup §5, backend_roadmap §7 |
| 데이터 보존 자동 worker (1개월 보관, 영구 삭제) | M4 | P3 | requirements §4.1·§5.3-5 |
| OS 서비스 wrapper 운영 진입 시점 결정 | M4 | P3 | ADR-0001 §8-7 |

### 4.6 AI

| 작업 | 마일스톤 | 우선순위 | 출처 |
| --- | --- | --- | --- |
| Python AI gRPC 서버 1차 | M3 | P3 | backend_roadmap §2 Phase 9 |
| AI Gardener suggestion 모델 + Go Core 연동 | M3 | P3 | backend_roadmap §5 P3 |
| Weekly report 생성 worker | M4 | P3 | frontend_integration §3.4 |
| AI 알림 중재 (집중 시간 보호) 모델 | M4 | P3 | requirements §4-3·§5.3-2 |

---

## 5. 백로그 (마일스톤 미배정 / 결정 필요)

다음 항목은 명세가 부족하거나 책임자 미정이라 마일스톤에 배정되지 않았다. 진입 시점에 ADR 또는 spec 작성 후 마일스톤으로 흡수.

| 항목 | 미정 부분 | 출처 |
| --- | --- | --- |
| `GET /api/v1/team/load`, `GET /api/v1/dashboard/velocity` | 데이터 source / 산출 기준 / 오너 | frontend_integration §6.3 |
| `GET /api/v1/me`, focus mode/notification settings 영속화 | 모델 / 저장 위치 | frontend_integration §3.1·P3 |
| Weekly report 생성 worker 실행 매체 | cron vs scheduled command | frontend_integration §3.4, api_contract §10 |
| 조직 도메인 — `parent_id` 검증, primary_dept 자동 판정 (겸임 우선순위, 동급 시 자식 노드 수), 파견/겸임 1:N 테이블, `total_count` Materialized View | spec / 마이그레이션 | backend_requirements_org_hierarchy §1·2, organizational_hierarchy_spec §3 |
| 알림 등급화 (Info / Action Required) 모델 | 모델 / 라우팅 정책 | requirements §5.2-7 |
| 기술 태깅 Kudos 가시성 | RBAC matrix와의 매핑 | requirements §5.1-3 |
| 외부 부서 의존성 수동 등록 | UI / 모델 | requirements §5.2-6 |
| `architecture/README.md`, `planning/README.md` TBD 스텁 | 본 통합 로드맵 채택 후 산출물로 채움 | 양자 |

---

## 6. 충돌 해소 (source-of-truth 정리)

`requirements.md`, `architecture.md`, `backend/requirements.md`, `backend_api_contract.md` 사이에 시점 차이로 인한 표현이 남아 있다. 본 표가 통합 로드맵의 source-of-truth 다.

| 주제 | 폐기된 표현 | 채택된 표현 | 결정 출처 |
| --- | --- | --- | --- |
| 인증/계정 구현 | 자체 `accounts` 테이블, 자체 7 endpoint (`requirements §2.5`, `architecture §6.2`, `backend/requirements §5`, `api_contract §11` historical) | 정책 invariant 만 보존, 구현은 **Hydra + Kratos** (Kratos가 credential master, `users` 테이블은 organizational metadata 만) | ADR-0001 (2026-05-07) |
| 브라우저↔서버 실시간 | gRPC stream (`backend/requirements §1`) | **REST snapshot + WebSocket** | requirements_review §3.1, frontend_integration §2.1 |
| 역할 wire 형식 | `DEVELOPER\|MANAGER\|ADMIN` (`backend/requirements §4`) | **`developer\|manager\|system_admin`** | api_contract §2, requirements_review §3.3 |
| 명령성 액션 응답 | boolean `ActionResponse` (`backend/requirements §2`) | **`command_id` + `command_status` lifecycle** | api_contract §9 |
| 환경 default | docker-compose default (`tech_stack §2`, 일부 `architecture`, `PROJECT_PROFILE`) | **native default**, docker 는 환경별 자산 (git 추적 외부) | PR #12 BLK-1 (2026-05-08), environment-setup §0 |
| Phase 8 상태 | "프론트 done" 만으로 완료 표기 (frontend_roadmap) | **백엔드 in_progress 가 source** — 인증/필터/replay 미완 | backend_roadmap §2 Phase 8 |
| RBAC 모델 | 1차원 `none\|read\|write\|admin` rank (`rbac.go defaultRBACPolicy`, `api_contract §6` legacy) | **per-resource 4-boolean** (`{view, create, edit, delete}`) — 5 resources (`infrastructure`, `pipelines`, `organization`, `security`, `audit`) | ADR-0002 + api_contract §12 (2026-05-08) |
| RBAC enforcement | `requireMinRole` 라우트별 정적 rank 비교 | **`requirePermission(resource, action)`** — 라우트-매핑 표 + DB-backed matrix + deny-by-default | ADR-0002 §4.3, api_contract §12.8·§12.9 (PR-G5 머지 시 발효) |

위 폐기 표현이 본문에 그대로 남은 위치(`requirements.md`, `architecture.md`, `backend/requirements.md`, `tech_stack.md`, `PROJECT_PROFILE.md`)는 M0~M1 의 문서 정리 작업으로 *재설계 박스* 또는 *링크 참조* 형태로 정리한다.

---

## 7. 변경 이력

| 일자 | 변경 | 메모 |
| --- | --- | --- |
| 2026-05-08 | 초판 작성. M0~M4 정의, 트랙 매핑, 충돌 해소 표 정리. | PR #12, #13 머지 직후. claude/merge_roadmap 브랜치. |
| 2026-05-08 | §6 충돌 해소 표에 RBAC 모델/enforcement 결정 2행 추가. | M1 PR-G1, ADR-0002 채택 반영. claude/m1-pr-g1-rbac-contract 브랜치. |

---

## 8. 다음 단계

1. 본 문서를 PR 로 머지하면 `docs/README.md`, `docs/DOCUMENT_INDEX.md`, 트랙별 세부 로드맵 상단에 진입점 링크가 추가된다.
2. M0 sprint 진입 직전, 본 문서 §3.1 의 DoD 항목별로 backlog 항목을 분해한다 (`ai-workflow/memory/claude/<branch>/backlog/<date>.md`).
3. 트랙 별 세부 로드맵은 본 문서의 마일스톤·우선순위 분류를 따르도록 갱신한다 — phase 표가 본 문서의 M0~M4 와 어떻게 매핑되는지 표 1개를 상단에 둔다.
4. `architecture/README.md`, `planning/README.md` 는 본 통합 로드맵 채택 후 산출물로 채운다.
