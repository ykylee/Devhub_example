# 코드 항목 분석 — 통합 발견 사항 백로그

- 문서 목적: 항목별 코드 분석(`code/backend/*`, `code/frontend/*`, `code/infra/*`)에서 식별된 불일치·부채·stale 를 단일 우선순위 백로그로 통합한다.
- 기준: main `cf19c94` (PR #374).
- 매핑: 각 항목은 [향후 방향 §3](../06_future_direction.md) 의 N/X/E 백로그에 연결. 코드 근거는 `file:줄`.
- 처리 방침: 본 PR 은 **(a) 명백한 stale 주석/문자열/오타 = 즉시 정정**, **(b) 동작 변경·테스트 필요 항목 = 백로그 등록(향후 sprint)**. 동작 변경을 doc-review PR 에 끼워 넣지 않는다.

---

## 1. HIGH — 즉시 후속 sprint

| ID | 발견 | 근거 | 매핑 | 조치 |
| --- | --- | --- | --- | --- |
| H-1 | repository **draft→publish 핸들러 + store 무테스트** 머지(#368). SCM 생성 부분실패(502) 경로 검증 공백 | `httpapi/repository_ops.go` createRepositoryDraft/requestRepositoryPublish; `store` CreateRepositoryDraft/MarkRepositoryDraftPublishRequested | N-2 | UT + 통합테스트 보강 |
| H-2 | **credentials_ref 평문 raw 응답** (HMAC/SDK secret 포함 가능). api_token/auth_secret 은 write-only 인데 legacy 만 예외 | `httpapi/integration_registry.go:36` `integrationProviderResponse`; migration 000028/000040/000041 | X-3 (#6) | envelope 암호화 + 응답 redact |

## 2. MEDIUM — 보안/정확성 (백로그 등록, 테스트 동반)

| ID | 발견 | 근거 | 조치 |
| --- | --- | --- | --- |
| M-1 | **onboardingGate allowlist method-agnostic** — `c.FullPath()` path 키라 `PATCH /api/v1/me` 도 통과 → 미완료 사용자가 onboarding 우회 profile 변경 가능 (`patchMe` 완료여부 미검사) | `httpapi/onboarding_gate.go` + `me.go::patchMe` | gate 에 method 분기 또는 patchMe 완료검사 추가 + 회귀 test |
| M-2 | realtime WS `CheckOrigin` 무조건 true (CSWSH 표면) | `httpapi/realtime.go` upgrader | Origin allowlist (사내 origin) |
| M-3 | infra snapshot process-global 가변 전역 → multi-instance 비일관 | `httpapi/snapshot_provider.go` / runtime_snapshot_provider | PG 백킹 또는 단일 인스턴스 전제 명문화 |
| M-4 | mock fallback 이 backend 장애를 정상 데이터로 가림 (infra/gardener/risk service) + `identityService.mockHierarchy()` OrgTree fallback + account email 합성 + MFA "Disabled" 하드코딩 | `frontend/lib/services/{infra,gardener,risk,identity}.service.ts`, `app/(dashboard)/account/page.tsx` | 운영 UI 전환 2차 — mock fallback 제거 + 에러 표면화 |
| M-5 | `risk.service` raw fetch + 하드코딩 `X-Devhub-Actor: 'yklee'` → Bearer/401 refresh 누락 (호출 페이지 현재 archived) | `frontend/lib/services/risk.service.ts` | apiClient 경유로 전환 (또는 service 폐기) |

## 3. MEDIUM — 정리/일관성 (백로그)

| ID | 발견 | 근거 | 조치 |
| --- | --- | --- | --- |
| C-1 | **dead code**: `requireMinRole`(라우터 미연결) / `resolveIdPSubject`(핸들러 미참조) / `websocket.service.startMockEvents` + `mockUsers` / `dashboard.service` dead-path | `httpapi/authz.go`, `identity_resolver.go`; `frontend/lib/services/{websocket,dashboard}.service.ts` | dead code 제거 (별도 cleanup sprint) |
| C-2 | **WS 이중화 미정리** — ticket 기반 `realtime.service`(Header/topology) ↔ legacy `?access_token=` `websocket.service`(`AuthGuard.tsx:96`) 공존. ADR-0024 ticket-only cutover 가 frontend 일부 미완 | `frontend/lib/services/{realtime,websocket}.service.ts`, `components/layout/AuthGuard.tsx:96` | AuthGuard 를 realtime.service 로 통일 + websocket.service 폐기 |
| C-3 | **command status WS UI 미완** — `command.status.updated` 소비처 0 (Phase 4 잔여). Header `dev_request.created` 구독이 publish 타입에 없어 영구 미발화 | `frontend` Header/realtime 구독 | command toast/status UI 연결 (RM-M4-03) |
| C-4 | env 자격증명 모델 이원화 — integration provider 는 env fallback 금지(#359)인데 legacy SCM catalog `has_credentials` 는 GITEA env 참조 | `store/applications.go:78` | legacy SCM catalog 정합 또는 폐기 |
| C-5 | gitea webhook 헤더 alias 이원화 (전용 핸들러 X-Gitea/X-Gogs only, 범용 ingest X-Integration fallback 포함) | `httpapi/gitea_webhook.go` vs `integration_registry.go` ingest | 정규화 일원화 (X-2) |
| C-6 | priority role 로직 2곳 중복 | `audit/user_sync.go` + `httpapi/onboarding_roles.go` (또는 keycloak group 매핑) | DRY helper 추출 |
| C-7 | gitea 패키지만 Prometheus metric 부재 + 30s 하드코딩 | `internal/gitea/worker.go` | metric 추가 + interval env |
| C-8 | hrdb `Client` interface dead 선언 + Postgres adapter 미배선 (Mock 만 wire) | `internal/hrdb/*` | adapter wire 또는 dead 선언 정리 |
| C-9 | scm_providers(000012) ↔ integration_providers(000028) SCM catalog 이중화 | migration 000012/000028 | 통합 검토 (별도 ADR 후보) |
| C-10 | column scan 비대칭 — `GetUser`/`GetUserByIdPSubject`/`ListUnitMembers` 가 `user_type` 미조회 → 단건 `AppUser.Type` 목록 불일치 | `store/users_units.go` | scan 정합 |
| C-11 | total_count MV(000011) 가 GetHierarchy 실경로에서 미사용 (dead 가능) | migration 000011 + `store` GetHierarchy | MV join 활성화 또는 제거 |
| C-12 | 데이터 파괴형 down migration (000021 role 재할당 / 000025·000037·000039 NULL row / 000045 비대칭 backfill) | 해당 down.sql | down 위험 주석 명시 |
| C-13 | lint disable 회귀 패턴 8 파일 9건 (set-state-in-effect 위주, Header/layout 우회) | `frontend` 다수 | useSyncExternalStore 정공법 통일 |

## 4. LOW — stale 주석/오타 (본 PR 즉시 정정 대상)

| ID | 발견 | 근거 | 본 PR 처리 |
| --- | --- | --- | --- |
| L-1 | `dev_requests` 에러 텍스트 `{10}` (실제 정규식 `{1,10}`) | `httpapi/dev_requests.go` 에러 메시지 | ✅ 정정 |
| L-2 | migration `000030` 헤더 주석 "-- 000021" 오기 | `migrations/000030_*.up.sql:1` | ✅ 정정 |
| L-3 | `router.go` onboardingGate "default OFF" 주석 (실제 flag default true) | `httpapi/router.go` 주석 | ✅ 정정 |
| L-4 | applications "501 stubs" 주석 (이미 구현 완료) | `httpapi/applications.go` 또는 router 주석 | ✅ 정정 (존재 시) |

## 5. INFRA (백로그 — 사내/CI)

| ID | 발견 | 근거 | 매핑 |
| --- | --- | --- | --- |
| I-1 | migration prefix guard 가 **CI-bypass 머지를 못 잡음** (branch protection required check 부재) — #363↔#368 000042 충돌 사후 적발 | `.github/workflows/ci.yml:58-90` | N-5 |
| I-2 | **Keycloak SPI realm events 미배선** — realm JSON 은 `devhub-event-listener` 선언하나 운영 compose 가 stock `keycloak:26.0` 사용 (`Dockerfile.keycloak` SPI JAR 미사용) → 존재하지 않는 listener 참조. 현재 backend cron polling 만 동작 | `infra/idp/*` | X-8 (P3-5/P2-6) |
| I-3 | proto/gRPC 미구현 — `backend-ai/main.py` TODO 주석만, 생성물 부재. 실 연동은 HTTP | `backend-ai/main.py:10-11`, `proto/` | E-3 |
| I-4 | `hrdb_etl_sync.sh` deprecated 잔재 (full ETL loop 코드 잔존) | `scripts/hrdb_etl_sync.sh` | cleanup |
| I-5 | nginx WS token redact 미적용 (ADR-0024 §6 carve 2) — README 권장 patch 문서화만 | `infra/nginx/*` | 사내 |
| I-6 | `ci-setup.sh` stale 참조 (파일 부재, 주석만 잔존) | `ci.yml` 주석 | cleanup |

---

## 6. 본 PR 처리 요약

- **즉시 정정(L-1..L-4)**: stale 주석/오타 — 코드 동작 무변경, doc-review PR 안전 범위.
- **백로그 등록(H/M/C/I)**: 동작 변경·테스트 필요 항목은 [향후 방향 §3](../06_future_direction.md) N/X/E 에 연결하여 후속 sprint 로 이관. 본 PR 에서 코드 동작은 변경하지 않는다.
