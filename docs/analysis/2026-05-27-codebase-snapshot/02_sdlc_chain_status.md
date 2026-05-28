# 02. SDLC 체인 구성 현황 (컨셉→요구사항→유스케이스→설계→구현→단위테스트→E2E)

- 문서 목적: DevHub 의 SDLC 자산 체인을 도메인별로 점검하고, 현재 코드 상태와의 정합/갭을 식별한다. 본 점검 결과가 [추적성 매트릭스](../../traceability/report.md) 갱신의 근거다.
- 기준: main `cf19c94` (PR #374) / 마지막 기능 커밋 `99d6edc`.
- 단계 정의: 컨셉(`docs/planning/*_concept.md`) → 요구사항(`docs/requirements.md` REQ-FR/NFR) → 유스케이스(`docs/planning/system_usecases.md` UC) → 설계(`docs/architecture.md` ARCH + `docs/backend_api_contract.md` API + ADR) → 구현(IMPL) → 단위테스트(UT) → E2E(TC).

---

## 1. 단계별 자산 현황 (전역)

| 단계 | 자산 위치 | 현황 |
| --- | --- | --- |
| 컨셉 | `docs/planning/*_concept.md` (project_management / development_request / external_system_integration / keycloak_user_onboarding) | 4 핵심 도메인 컨셉 + 보조 plan 다수. |
| 요구사항 | `docs/requirements.md` §2~§5 (+ backend/org hierarchy spec) | REQ-FR-01..105 + NFR-01..26 + APP/PROJ/DREQ/INT/ONBOARD 도메인 확장(§5.4~§5.7). |
| 유스케이스 | `docs/planning/system_usecases.md` §2.1~§2.13 | UC-AUTH/ACCOUNT/ORG/RBAC/GITEA/CMD/AUD/RT/APP/PROJ/DREQ/INT/ONBOARD. |
| 설계(ARCH) | `docs/architecture.md` §1~§9 | ARCH-01..17 + DREQ(§7)/INT(§8)/ONBOARD(§9) 서브. |
| 설계(API) | `docs/backend_api_contract.md` §3~§16 | API-01..90 (실제 코드는 90까지 활성). |
| 설계(ADR) | `docs/adr/0001..0024` | 24 ADR. |
| 구현(IMPL) | `backend-core/` + `frontend/` | §2.4 IMPL-* (추적성 §2.4). |
| 단위테스트(UT) | `*_test.go`(72) + `*.test.ts(x)`(10) | UT-* (추적성 §2.5). |
| E2E(TC) | `frontend/tests/e2e/*.spec.ts`(28) | TC-* (추적성 §2.6). |

---

## 2. 도메인별 체인 정합 점검

각 도메인이 컨셉→...→E2E 중 어디까지 도달했는지 + **추적성 매트릭스에 반영 필요한 갱신 항목**.

### 2.1 인증/계정/조직/RBAC/감사 (기반 도메인)

- 체인: ✅ 컨셉(역할 우선순위) → ✅ REQ → ✅ UC-AUTH/ACCOUNT/ORG/RBAC/AUD → ✅ ARCH/API(19,26..40,32..34) + ADR-0002/0011/0019/0020/0021 → ✅ IMPL → ✅ UT → 🟡 E2E.
- **현재 상태**: Keycloak 단일 IdP 완전 전환(Hydra/Kratos·`/api/v1/accounts/*`·`/api/v1/auth/*` 모두 폐기). lazy_auto_create 도 폐기(Onboarding gate 정공법).
- **갱신 필요**: 추적성 §4 ADR 인덱스에 **ADR-0022(Keycloak 25.0 pin)·ADR-0023(Keycloak 26.0 pin)** 누락 → 추가. 백엔드 세부 로드맵 §2 Phase 13 이 여전히 "Hydra/Kratos in_progress" 로 stale → 정정.

### 2.2 Application / Repository / Project

- 체인: ✅ 컨셉(project_management) → ✅ REQ-FR-APP/PROJ → ✅ UC-APP/PROJ → ✅ ARCH/API-41..58 + ADR-0011 → ✅ IMPL → ✅ UT(46) → 🟡 E2E(admin-projects/applications/repositories + project-model-modes/v2).
- **현재 상태(신규)**: project model v2(application-project-repo 관계 + standalone project), repository **draft→publish lifecycle**(#368), **SCM↔시스템 repository 소유권 분리 + 양방향 import/create**(#363/#366), provider_id 단일화(#373).
- **갱신 필요**: 추적성에 **repository draft/publish(#368) + SCM 양방향 연동(#363/#366/#373) IMPL/UT 행 부재** → 신규 행. API-88/89/90 미등재 → 추가. **#368 draft→publish handler 가 무테스트 머지** → UT/TC 보강 carve(향후 §6).

### 2.3 Dev Request (DREQ) — M5 closing

- 체인: ✅ 컨셉(development_request) → ✅ REQ-FR-DREQ-001..013 → ✅ UC-DREQ-01..12 → ✅ ARCH-DREQ(§7)/API-59..68,79 + ADR-0012/0013/0014/0017 → ✅ IMPL(domain/store/handler/intake_auth/promote_tx/cron) → ✅ UT → ✅ TC-DREQ-* 13건.
- **현재 상태**: 완성. PATCH expires_at + EditIntakeTokenModal(ADR-0017 §6(b)) resolved. DREQ↔notification 연계(#323).
- **갱신 필요**: 추적성 정합 양호. development_roadmap M5 를 "closing" 으로 확정 표기(현재 부분 ⏳ 잔존).

### 2.4 External Integration — M6 + 2026-05-26/27 깊이 확장

- 체인: ✅ 컨셉(external_system_integration) → ✅ REQ-FR-INT-001..012 → ✅ UC-INT-01..14 → ✅ ARCH-INT(§8)/API-69..90 + ADR-0015/0016 → ✅ IMPL → 🟡 UT → 🟡 TC.
- **현재 상태(대폭 확장)**: HomeLab pull + provider/binding CRUD + topology v2 에 더해 — **Gitea SCM 동기화 워커**(pull, `internal/gitea/`, integration_sync_jobs 큐, RM-M4-06 1차), **provider 등록 UX 고도화**(vendor 템플릿 7종 + 가이드 자격증명 + base_url + 연결 테스트 API-87), **auth_mode full 모델**(token/basic/app_password/oauth2/agent + write-only `auth_secret`, migration 000041), **api_token write-only 슬롯**(000040), **webhook 헤더 alias**(X-Gitea/X-Gogs fallback), **admin catalog UI**(#357/#361), **고정메뉴 Phase 2b**(#362).
- **갱신 필요**: 추적성에 API-87(test-connection)·88(scm-repositories list)·89(import-repositories)·90(create-repository) + migration 000038/000040/000041/000042/000043/000045 + auth_mode/api_token IMPL 행 보강. RM-M4-06 1차 구현 반영(이미 §3 일부 반영됨 — 깊이 추가).

### 2.5 Onboarding — M7, **IMPL/UT/E2E 까지 완성 (문서엔 미반영)**

- 체인: ✅ 컨셉(keycloak_user_onboarding §5.9) → ✅ REQ-FR-ONBOARD-001..012 → ✅ UC-ONBOARD-01..11 → ✅ ARCH-ONBOARD(§9)/API-83..86,32/33 + ADR-0021 → ✅ **IMPL 완료** → ✅ **UT 완료** → ✅ **TC 완료**.
- **현재 상태(중대 drift)**: 로드맵/추적성은 RM-ONBOARD-01..04 를 "⏳ M-v1.1 진입 예정" 으로 표기하나, **실제로는 Carve A(backend, #278) + B/C(frontend·admin, #288) + D(tests, #289) 모두 머지 완료 + feature flag default ON flip + lazy_auto_create.go 삭제(#290) + codex hotfix(#291)** 까지 끝났다. `frontend/app/onboarding/page.tsx` + `components/onboarding/*` + `admin/users/ConfirmReviewModal` + `onboarding_test.go` + `onboarding-first-login.spec.ts` 모두 active.
- **갱신 필요(최우선)**: development_roadmap M7 + release_v1_roadmap §2.3/§3.3(P2-8..12)/§4.2 의 ⏳ → ✅ done 정정. 추적성 Onboarding 행은 Carve A 만 반영 → B/C/D 완료 반영. lazy_auto_create 폐기(P2-12/#284) closed 반영.

### 2.6 Realtime / WebSocket — 부분 (RM-M4 잔여)

- 체인: ✅ REQ → ✅ UC-RT → ✅ ARCH-05/API-14,36 + **ADR-0024(WS ticket auth)** → 🟡 IMPL(ticket store PG/in-memory 완성, event publish 는 command 만) → 🟡 UT → 🔴 TC(WS replay E2E 없음).
- **현재 상태**: ADR-0024 §6 carve 1/3/4/5/6 모두 resolved(ticket-only cutover #348). 잔여 = RM-M4-01 infra/ci/risk event publish + RM-M4-02 replay + 리소스 필터링.
- **갱신 필요**: 추적성에 ADR-0024 §6 종결 + realtime-02(ticket store) 이미 반영됨. RM-M4-01/02 는 향후 carve 유지.

### 2.7 Gitea Webhook + 도메인 데이터 (push + pull)

- 체인: ✅ REQ-FR-49..55 → ✅ UC-GITEA → ✅ ARCH-06..08/API-02 → ✅ IMPL(push: signature.go / pull: client·syncer·worker) → ✅ UT(gitea-01..04) → 🔴 TC(외부 Gitea 의존).
- **현재 상태**: push(webhook ingest) + pull(sync worker, #341) 양쪽 구현. 추적성 §2.4 IMPL-gitea-XX + §3 'Gitea SCM 동기화 워커' 행 이미 반영(#344).

---

## 3. 단계 전이 갭 요약 (체인 누수 지점)

| 갭 | 도메인 | 심각도 | 조치 |
| --- | --- | --- | --- |
| **G1** Onboarding IMPL/UT/TC 완료가 로드맵·추적성에 미반영 | Onboarding | 🔴 높음 | 본 PR: 로드맵 ⏳→✅, 추적성 행 갱신 (Step 3/4) |
| **G2** ADR-0022/0023 추적성 §4 인덱스 누락 | 인증/인프라 | 🟠 중간 | 본 PR: ADR 인덱스 2행 추가 (Step 3) |
| **G3** API-87..90 + migration 000038~000045 추적성 미등재 | Integration/Repository | 🟠 중간 | 본 PR: API 매핑 + IMPL 행 보강 (Step 3) |
| **G4** repository draft→publish(#368) 무테스트 머지 | Repository | 🟠 중간 | 향후 carve: UT/TC 보강 (Step 6/8) |
| **G5** SCM 양방향 연동(#363/#366/#373) IMPL 행 부재 | Integration | 🟠 중간 | 본 PR: IMPL/UT 행 추가 (Step 3) |
| **G6** 평문 secret 저장(credentials_ref/api_token/auth_secret) ADR/REQ-NFR 부재 | Integration/보안 | 🟡 낮음 | 향후 carve: envelope 암호화 ADR (#6, Step 8) |
| **G7** FE 단위테스트 밀도 < backend (서비스 18 중 2 커버) | 프론트엔드 | 🟡 낮음 | 향후 carve: service/component vitest 보강 (Step 7/8) |
| **G8** 백엔드/프론트 세부 로드맵 stale(Hydra·accounts·Phase) | 문서 | 🟠 중간 | 본 PR: 세부 로드맵 정정 (Step 4) |

---

## 4. 결론 — 체인 성숙도

| 도메인 | 컨셉 | REQ | UC | ARCH/API | ADR | IMPL | UT | E2E |
| --- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
| 인증/계정/조직/RBAC/감사 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 |
| Application/Repository/Project | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 |
| Dev Request (DREQ) | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| External Integration | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 | 🟡 |
| Onboarding | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Realtime/WebSocket | ✅ | ✅ | ✅ | ✅ | ✅ | 🟡 | 🟡 | 🔴 |
| Gitea (push+pull) | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ | 🔴(외부의존) |
| backend-ai (AI Gardener) | 🟡 | 🟡 | — | — | — | 🔴 | 🔴 | 🔴 |

**핵심 메시지**: 5개 핵심 도메인(인증/앱·프로젝트/DREQ/Integration/Onboarding)은 컨셉→E2E 체인이 사실상 닫혀 v1.0 기능 완성에 도달했다. 미성숙은 (a) Realtime event publish/replay, (b) backend-ai, (c) FE 단위테스트 밀도, (d) 최신 backend 기능의 E2E 후행. 가장 큰 실질 부채는 **코드는 앞섰는데 문서(추적성·로드맵)가 1주 뒤처진 drift** — 본 PR 로 정합한다.
