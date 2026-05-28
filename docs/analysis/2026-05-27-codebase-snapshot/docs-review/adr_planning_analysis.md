# docs/adr (0001..0024) + docs/planning (33) ↔ 코드/상호 정합 분석 (2026-05-27 스냅샷)

- 문서 목적: 24개 ADR 의 상태/supersession 체인/cross-ref 정합과, 33개 planning 문서의 상태/코드 반영/stale 여부를 main `cf19c94` 기준으로 점검하고, 갱신 권고(향후 작업용)를 근거와 함께 제시한다.
- 범위: `docs/adr/0001..0024` 메타 헤더 + supersession, `docs/planning/*.md` 33건의 lifecycle/코드 반영. 코드/대규모 본문 수정 없음 — critical cross-ref 보강만 별도 표기.
- 대상 독자: ADR governance 검토자, planning 문서 갱신자, 추적성 동기화 작업자, 리뷰어.
- 상태: snapshot (2026-05-27, main `cf19c94` / 마지막 기능 커밋 `99d6edc`)
- 근거 자료: `../01_codebase_state_analysis.md`, `../02_sdlc_chain_status.md`, `docs/traceability/report.md` §3/§4, `docs/adr/0001..0024` 헤더, `docs/planning/*.md` 헤더.

---

## 1. ADR 24개 인덱스 + supersession 체인

| # | 제목(요약) | 상태 | supersedes | superseded-by | 영향 도메인 |
| --- | --- | --- | --- | --- | --- |
| 0001 | IdP 선택 (Ory Hydra + Kratos) | **superseded** | — | **ADR-0019** (2026-05-19) | 인증/계정 |
| 0002 | RBAC policy 편집 API (Option A DB-backed) | accepted | — | — | RBAC |
| 0003 | No-Docker 정책 CI 범위 (services:postgres 제거) | accepted | — | — | 인프라/CI |
| 0004 | `X-Devhub-Actor` 폐기 완료 선언 | accepted | (ADR-0001 §8#4 trigger) | — | 인증 |
| 0005 | workflow lint (actionlint) CI 잡 | accepted | (ADR-0003 §6 후속) | — | CI |
| 0006 | inbound `X-Devhub-Actor` 명시 거부 (400) | accepted | (ADR-0004 §6 후속) | — | 인증 |
| 0007 | RBAC PermissionCache 다중 인스턴스 일관성 (PG LISTEN/NOTIFY) | accepted | — | — | RBAC/운영 |
| 0008 | HRDB production 어댑터 (PG `hrdb` schema) | accepted | — | — | 계정/조직 |
| 0009 | 파견/겸임 모델 + total_count MV | accepted | — | — | 조직 |
| 0010 | `primary_dept` 자동 판정 알고리즘 | accepted | — | — | 조직 |
| 0011 | Application/Project Owner 위양 + RBAC row-scoping | accepted | — | — | App/Project/RBAC |
| 0012 | DREQ 외부 수신 endpoint 인증 (API token + IP allowlist) | accepted | — | — | DREQ/인증 |
| 0013 | DREQ 도메인 RBAC row-scoping | accepted | (ADR-0011 §4.2 적용) | — | DREQ/RBAC |
| 0014 | DREQ intake token admin 발급/조회/revoke | accepted | — | — | DREQ |
| 0015 | HomeLab adapter pull source 전략 (file+HTTP) | accepted | — | — | External Integration |
| 0016 | Prometheus 알림 규칙 정책 (HomeLab) | accepted | — | — | External Integration/운영 |
| 0017 | DREQ intake token 운영 hardening (expires_at + PATCH) | accepted | (ADR-0014 §6 carve) | — | DREQ |
| 0018 | 단일 외부 포트 reverse proxy + `/devhub` prefix | accepted | — | — | 인프라/운영 |
| 0019 | Keycloak 단일화 (Hydra+Kratos 폐기) | accepted | **ADR-0001** | (§3.2 등 일부 → ADR-0021 partial) | 인증/계정 |
| 0020 | 외부 Keycloak 하 계정/사용자 관리 책임 경계 | accepted (**partial superseded**) | — | **ADR-0021** (partial, 5 위치) | 계정/사용자/RBAC |
| 0021 | Onboarding self-service unit selection + lazy 폐기 | accepted | **ADR-0020 (partial)** | — | Onboarding/계정 |
| 0022 | Keycloak 버전 pin 25.0 | **superseded** | — | **ADR-0023** (2026-05-26) | 인프라/인증 |
| 0023 | Keycloak 버전 pin 26.0 (0022 retreat reversal) | accepted | **ADR-0022** | — | 인프라/인증/운영 |
| 0024 | WebSocket 인증 — `?access_token=` query → ticket-only cutover | accepted | — | — | 인증/운영/realtime |

### 1.1 supersession 체인 정합 — 결론: 양방향 링크 모두 정상

3개 체인이 존재하며, immutable-history 패턴(`feedback_adr_supersession_pattern`)을 모두 따른다:

1. **ADR-0001 → ADR-0019** (full): 0001 상태/§0 banner 가 0019 를 명시, 0019 헤더 `supersedes: ADR-0001` 명시. 본문 immutable 보존(+ PR #167 의 partial heading 수정을 sprint `-a` 가 복원). ✅ 양방향 정합.
2. **ADR-0020 → ADR-0021** (partial, 5 위치): 0020 상태/§0 banner 가 "5 위치 partial supersession" + `partial superseded by: ADR-0021` 명시, 0021 헤더 `partial supersedes: ADR-0020` 명시(§3.2 row + §4.1 sub-carve B + §4.2 + §6.1 + §6.2). 핵심 결정(옵션 A 책임 경계)은 유지 명시. ✅ 양방향 정합 + scope wording 일치(`feedback_adr_supersession_cross_doc_sync` 위반 없음).
3. **ADR-0022 → ADR-0023** (full): 0022 상태/§0 banner/§3 banner 가 0023 을 명시, 0023 헤더 `Supersedes: ADR-0022` 명시. 0022 본문 immutable + §3.1 placeholder("사내 정합 사유 finalize 필요")가 finalize 못 된 채 supersession 된 사실까지 명문화. ✅ 양방향 정합.

### 1.2 cross-ref 점검 — critical 누락 없음

- **ADR-0024 (WS auth)** 는 ADR-0019(Keycloak 단일 IdP)만 cross-ref. ADR-0020/0021 미참조 — 그러나 WS 인증은 계정/onboarding 책임 경계와 무관하므로 **논리적 누락 아님**. 보강 불필요.
- **ADR-0022/0023 (버전 pin)** 은 둘 다 ADR-0019 를 cross-ref(버전 미명시 부분의 보완). 상호 supersession 링크 완비.
- ADR-0019 헤더의 cross-ref 에 ADR-0020/0021 이 명시돼 있지 않으나, 역방향(0020/0021 → 0019)은 명시. ADR governance 상 superseder→superseded 링크가 필수이고 0019 가 후속 ADR 을 모를 수 있음(시점상)은 정상. **선택적 보강 후보**(§4 권고 R-ADR-1)이며 critical 아님.
- **결론**: critical cross-ref 누락 없음 → 본 분석에서 ADR 메타 헤더 Edit 적용 없음(절차 5 조건 미충족).

### 1.3 ADR ↔ 추적성 §4 인덱스 정합

- 직전까지 추적성 §4 ADR 인덱스에 **ADR-0022/0023 누락**이 있었으나, `docs/traceability/report.md`(최종수정 2026-05-27)가 이미 ADR-0022/0023 2 row 추가 + ADR-0024 제목 ticket-only cutover 반영을 완료(라인 398-400). → 현재 추적성 §4 = 24/24 ADR 등재. **이 갭은 진행 중 분석 PR 에서 이미 해소**.

---

## 2. planning 33 문서 인덱스 + 상태/코드 반영/stale

> 상태 컬럼: 헤더 명시 lifecycle. "코드 반영" = 해당 plan 의 대상 기능이 main 코드에 활성인지. "stale" = 헤더 상태 또는 본문 진행표기가 현재 코드보다 뒤처졌는지.

| # | 문서 | 목적(요약) | 헤더 상태 | 코드 반영 | stale |
| --- | --- | --- | --- | --- | --- |
| 1 | README.md | planning 진입점 지도 | accepted | — (메타) | 🟡 경미 (최종수정 2026-05-20, M0~M6 기준) |
| 2 | account_user_management_redesign.md | 계정/사용자 책임 경계 redesign (Phase 1/2/3) | **draft** | ✅ Phase 1/2/3 실 구현 완료(ADR-0020/0021, accounts 폐기·lazy 폐기) | 🟠 **stale** — 실 구현 종료됐는데 draft (→ accepted/superseded-by-ADR) |
| 3 | application_management_hotfix_2026-05-27.md | App 등록/수정 UI 오류·개선 | **draft** | ✅ key regex `{1,10}` 코드 반영됨(applications.go:27) | 🟠 **stale** — 관찰 이슈 #1(regex) 이미 수정, plan draft 잔존 |
| 4 | development_request_concept.md | DREQ 도메인 컨셉 | draft(concept) | ✅ DREQ 풀스택 closing(M5) | 🟡 컨셉 문서 특성상 draft 허용, 단 도메인은 done |
| 5 | e2e_keycloak_migration.md | e2e Kratos→Keycloak admin API migration design | draft | ✅ Kratos 잔재 정리 완료(sprint -ad), e2e Keycloak 실연동 CI | 🟠 **stale** — migration 실행 완료, draft 잔존 |
| 6 | external_integration_capability_matrix.md | provider capability 매트릭스 + MVP 우선순위 | draft | ✅ capability gate(pull/sync/push) + Gitea sync + auth_mode 구현 | 🟠 **stale** — MVP 표기가 실 구현(API-87..90)보다 뒤처짐 |
| 7 | external_system_integration_concept.md | External Integration 도메인 컨셉 | draft(concept) | ✅ M6 + 2026-05-27 깊이 확장 | 🟡 컨셉 draft 허용, 도메인 활성 |
| 8 | homelab_adapter_pull_strategy.md | HomeLab pull source 전략 초안 | accepted | ✅ ADR-0015 로 승격(SoT 이관) | 🟢 정합 (ADR-0015 가 SoT) |
| 9 | keycloak_event_audit_integration.md | Keycloak event→audit_logs design | draft | ✅ 실 구현 완료(`internal/audit` event polling, ADR-0020 sub-carve C) | 🟠 **stale** — design→구현 완료, draft 잔존 |
| 10 | keycloak_failover.md | Keycloak failover(HA/DR) design | draft | 🔴 미구현 (Phase 2 HA carve) | 🟢 정합 (의도된 미진입 design) |
| 11 | keycloak_groups_rbac_mapping.md | Keycloak group→RBAC role 자동 매핑 design | draft | 🔴 미구현 (Phase 2 carve, staging-prod #214) | 🟢 정합 (의도된 미진입 design) |
| 12 | keycloak_offboarding_immediacy.md | Off-boarding 즉시성 chain design | draft | ⚠️ **cancelled** (Phase 1 cron 폐기, issue #215 not-planned) | 🟡 §0 banner 로 cancellation 명시됨 — 헤더 상태는 draft 잔존 |
| 13 | keycloak_only_refactor_execution_plan.md | Keycloak 단일화 실행 계획 | accepted | ✅ PR #167 완료(§0 banner ✅실행완료) | 🟢 정합 (ADR-0019 가 결정 SoT) |
| 14 | keycloak_service_account_min_role.md | service account manage-users 제거 design (ADR-0020 sub-carve E) | draft | 🟡 부분 (사내 Keycloak admin 동반 carve) | 🟡 사내 동반 carve — draft 합리적 |
| 15 | keycloak_sso_federation.md | 외부 Keycloak SSO federation(옵션 B) design | **deprecated** | 🔴 rejected (옵션 A 채택, ADR-0019) | 🟢 정합 (deprecated + rejected banner) |
| 16 | keycloak_user_onboarding_concept.md | Onboarding UI 컨셉 | draft(concept) | ✅ Onboarding 도메인 풀스택 closing | 🟡 컨셉 draft 허용, 도메인 done |
| 17 | onboarding_impl_plan.md | Onboarding IMPL 4 carve plan + RM-ONBOARD-01..04 | **draft** | ✅ **Carve A/B/C/D 전부 머지** + flag flip + lazy 삭제 | 🔴 **stale (최우선)** — 4 carve 모두 done 인데 plan 은 진입 전 상태 |
| 18 | ops_ui_transition_plan.md | 운영 UI 전환(mock 정리) 계획 | (상태 헤더 없음) | ✅ mock 비노출 + PageState + 실데이터(#334/#340/#342/#369) | 🟠 **stale** — 전환 완료, 계획 문서 잔존 + 상태 헤더 부재 |
| 19 | project_creation_dreq_notification_concept.md | 하이브리드 프로젝트 생성 + DREQ 알림 design | active | ✅ DREQ↔notification 연계(#323) | 🟢 정합 (active) |
| 20 | project_management_concept.md | Application/Repo/Project 도메인 컨셉 | draft(concept) | ✅ 도메인 풀스택 + v2 | 🟡 컨셉 draft 허용 |
| 21 | project_operating_model_example_2026.md | 운영 모델 예시 | accepted | — (참고 자산) | 🟢 정합 |
| 22 | project_operating_model_template.md | 운영 모델 템플릿 | accepted | — (참고 자산) | 🟢 정합 |
| 23 | project_repository_creation_linking_plan_2026-05-27.md | Project 독립생성+Repo 동반+N:M 연결 plan | **draft** | ✅ 대부분 구현(#354 단일 tx, standalone, project_repositories) | 🟠 **stale** — 코드 반영 진행됐는데 plan draft, 완료 표기 부재 |
| 24 | prometheus_homelab_alerts.md | Prometheus alert/dashboard 초안 | accepted | ✅ ADR-0016 로 승격 | 🟢 정합 (ADR-0016 SoT) |
| 25 | release_v1_roadmap.md | v1.0 릴리즈 로드맵 + 워커 분업 | **draft** | 부분 — 다수 carve done 인데 ⏳ 표기 | 🔴 **stale (최우선)** — §2.3/§3.3/§4.2 의 RM-ONBOARD ⏳, P2-8..12 미해소 표기 |
| 26 | single_port_reverse_proxy.md | 단일 포트 reverse proxy design | accepted | ✅ ADR-0018 로 승격 + nginx 구현 | 🟢 정합 (ADR-0018 SoT) |
| 27 | system_admin_catalog_plan_2026-05-27.md | System Admin 통합 관리 UI plan | **draft** | ✅ `/admin/catalog` 구현(#357/#361) | 🟠 **stale** — admin catalog UI 활성, plan draft |
| 28 | system_erd.md | 시스템 ERD 카탈로그 | draft | 🟡 부분 — 최신 migration(000038~000045) 미반영 가능 | 🟡 ERD 카탈로그 후행(신규 컬럼/테이블 추가분) |
| 29 | system_usecases.md | Usecase 카탈로그 | draft | ✅ UC 전 도메인(§2.13 Onboarding 포함) | 🟡 draft 잔존, 내용은 최신(2026-05-21) |
| 30 | ui_app_project_repo_upgrade_plan.md | UI 고도화 실행 계획 | **in_progress** | ✅ 대부분 완료(상세 mock 제거 + 실데이터) | 🟠 **stale** — in_progress 잔존, 1차/대부분 done |
| 31 | ui_e2e_followup_after_merge.md | UI E2E 후속 정비 메모 | active | 🟡 부분 (E2E 후속 잔여) | 🟢 정합 (active, 잔여 추적용) |
| 32 | view_menu_screen_api_matrix.md | 뷰별 메뉴/화면/API 매트릭스 | draft | 🟡 부분 — 2026-05-13 기준, M4/v2/신규 메뉴 미반영 | 🟠 **stale** — 최종수정 2026-05-13, 신규 도메인 메뉴(catalog/integration/onboarding) 누락 |
| 33 | ws_subprotocol_vs_ticket_poc.md | WS subprotocol vs ticket PoC 비교 | accepted | ✅ ADR-0024 §6 carve 4 resolved 근거 | 🟢 정합 |

### 2.1 cancelled 항목 잔재 점검 (절차 2-b)

- **Sign Up 흐름** (`POST /api/v1/auth/signup`, ADR-0008 의존): ADR-0019 Keycloak 단일화로 auth proxy 전체 폐기 → Sign Up cancelled. P3-12 로 release_v1_roadmap.md §4.2 라인 256 에 "cancelled (외부 Keycloak + ADR-0021 self-service 로 무효화)" 명시됨. ✅ 명문화 정합. 단 **ADR-0008** 본문은 여전히 Sign Up 의존 production 어댑터로 서술(accepted) — supersession 미표기(권고 R-ADR-2).
- **off-boarding cron** (keycloak_offboarding_immediacy.md Phase 1 §3.1 옵션 C): issue #215 cancelled, 문서 §0 banner 로 폐기 + historical 보존 명시. ✅ 정합. release_v1_roadmap §4.2 P1-4 cancelled 표기.
- **HRDB ETL unit** (P2-7, issue #223): release_v1_roadmap §3.3/§4.2 에 cancelled 명시. ✅ 정합.
- → 3개 cancelled 항목 모두 문서에 명문화됨. **잔재 risk 낮음** (`feedback_backlog_hygiene_post_decision_reversal` 위반 없음). 단 ADR-0008 만 supersession 표기 부재.

---

## 3. stale planning 문서 Top (심각도순)

| 순위 | 문서 | stale 내용 | 근거 |
| --- | --- | --- | --- |
| **1** | release_v1_roadmap.md | §2.3/§3.3 P2-8..P2-12 + §4.2 M-v1.1 라인 256 이 RM-ONBOARD-01..04 를 ⏳/미해소로 표기. 실제는 Carve A(#278)+B/C(#288)+D(#289)+flag flip&lazy 삭제(#290)+codex hotfix(#291) 전부 머지. P2-8 만 ✅ 표기, P2-9~12 미정정 | ../02 §2.5 G1 / 라인 91-94,189-193,256 / MEMORY onboarding carve_bcd |
| **2** | onboarding_impl_plan.md | 헤더 draft + Carve A/B/C/D 4건 모두 "진입 전" 서술. 실제 4 carve 전부 done + DoD(§7) 충족 + lazy_auto_create.go 삭제. status/changelog 섹션 부재 | ../02 §2.5 / 라인 31-178 / 추적성 §3 row 369 |
| **3** | account_user_management_redesign.md | 헤더 draft. Phase 1(현황)+Phase 2(결정 6건)+Phase 3(accounts 폐기·lazy 폐기 실구현) 모두 종료(ADR-0020/0021). draft→accepted 또는 "ADR-0020/0021 로 승격" 표기 필요 | ADR-0020/0021 §0 banner / ../01 §2.1 |
| **4** | external_integration_capability_matrix.md | MVP/후속 구분이 실 구현(auth_mode full, base_url, test-connection API-87, Gitea sync, SCM 양방향) 이전 기준(2026-05-15). capability gate 실 동작 미반영 | ../02 §2.4 / ../01 §2.3 |
| **5** | view_menu_screen_api_matrix.md | 최종수정 2026-05-13. 신규 메뉴(`/admin/catalog`, integration/integration-bindings, onboarding, dev-request-tokens) + M4/v2 구분 미반영 | ../01 §3.2 페이지 목록 |
| 6 | project_repository_creation_linking_plan_2026-05-27.md | draft. To-Be 대부분 구현됨(#354 단일 tx + standalone + N:M) — 완료/잔여 분리 표기 부재 | ../01 §2.3 Project model |
| 7 | system_admin_catalog_plan_2026-05-27.md | draft. `/admin/catalog` 구현(#357/#361) 완료 — 진행상태 갱신 부재 | ../01 §3.2 |
| 8 | ui_app_project_repo_upgrade_plan.md | in_progress. 1차/상세 mock 제거 대부분 done — 단계별 done 마크 부재 | ../01 §3.5 |
| 9 | ops_ui_transition_plan.md | 상태 헤더 부재 + 전환 완료(#334/#340/#342/#369) — 메타 헤더 + 완료 표기 보강 | ../01 §3.5 |
| 10 | keycloak_event_audit_integration.md / e2e_keycloak_migration.md | design draft. 둘 다 실 구현 완료(audit event polling / e2e Keycloak 전환) — draft→accepted 또는 "구현 완료" banner | ../01 §2.1 / sprint -ad |

> **공통 근본 원인**: ../01 §6 + ../02 §4 의 "코드는 앞섰는데 문서(로드맵·plan)가 1주 뒤처진 drift". 추적성 매트릭스(report.md)는 2026-05-27 정합됐으나, planning 문서 본문의 진행상태 표기가 따라가지 못함.

---

## 4. 갱신 권고 (향후 작업용)

> 본 절은 **권고만** 기록한다(절차 4). 코드/대규모 본문 수정 없음. critical cross-ref 보강(절차 5)은 §1.2 결론대로 적용 대상 없음.

### 4.1 ADR 권고

- **R-ADR-1** (선택, low): ADR-0019 헤더 관련 문서 목록에 ADR-0020/0021(후속) 역참조 1줄 추가 검토. 단 superseded→superseder 방향은 이미 0020/0021 측에 있어 governance 필수 아님 → **보류 가능**.
- **R-ADR-2** (low): ADR-0008(HRDB production adapter)이 Sign Up 의존으로 서술됐으나 Sign Up 흐름은 ADR-0019 로 cancelled. ADR-0008 메타 헤더에 "Sign Up 경로 폐기(ADR-0019) — 본 어댑터의 Sign Up 의존 부분은 historical, HRDB lookup 자체는 onboarding/조직 매핑에서 유효" 1줄 banner 검토. (즉시성 design 의 #215 cancellation 과 정합)
- ADR 인덱스(추적성 §4)는 24/24 등재 완료 — 추가 조치 불필요.

### 4.2 planning 권고 (문서별 위치 + 정정 방향)

1. **release_v1_roadmap.md** — §3.3 표의 P2-9/P2-10/P2-11/P2-12 행 + §4.2 M-v1.1 라인 256 + §2.3 Onboarding 표의 RM-ONBOARD-02/03/04 (라인 92-94) 의 `⏳ M-v1.1` → `✅ done (#288/#289/#290/#291)`. P2-12 lazy_auto_create deletion(#284) → resolved. M-v1.1 backlog 에서 P2-8..12 제거.
2. **onboarding_impl_plan.md** — 헤더 상태 `draft` → `done` (또는 "실행 완료") + 문서 상단에 §0 완료 banner(Carve A #278 / B·C #288 / D #289 / flag flip+lazy 삭제 #290 / hotfix #291) 추가. §7 DoD 표에 done 체크.
3. **account_user_management_redesign.md** — 헤더 `draft` → `accepted` + "Phase 1/2/3 결정은 ADR-0020(책임 경계) + ADR-0021(onboarding/lazy 폐기)로 SoT 이관, 본 문서는 분석 자산으로 immutable 보존" banner 추가.
4. **external_integration_capability_matrix.md** — §2 매트릭스에 실 구현 capability gate(pull=import / sync=mirror / push=create) + auth_mode 5종 + test-connection(API-87) 반영. 최종수정일 갱신.
5. **view_menu_screen_api_matrix.md** — 신규 메뉴 행 추가(`/admin/catalog`, integrations / integration-bindings, onboarding, dev-request-tokens) + 최종수정일 2026-05-13 → 현재. (또는 deprecated 표기 후 `docs/analysis/.../code/frontend/pages.md` 로 위임)
6. **project_repository_creation_linking_plan_2026-05-27.md** / **system_admin_catalog_plan_2026-05-27.md** / **ui_app_project_repo_upgrade_plan.md** — 각 단계 표에 done 마크 + 헤더 상태 정정(draft/in_progress → done 또는 잔여만 active).
7. **ops_ui_transition_plan.md** — 메타 헤더(목적/범위/상태/최종수정일) 추가 + 전환 완료(#334/#340/#342/#369) banner.
8. **keycloak_event_audit_integration.md** / **e2e_keycloak_migration.md** — design draft 상단에 "실 구현 완료" banner(audit event polling / e2e Keycloak 전환) + 상태 정정.
9. **keycloak_offboarding_immediacy.md** — §0 cancellation banner 는 이미 있음. 헤더 상태 `draft` → `partially cancelled` 또는 `deprecated (Phase 1)` 로 명확화(선택).
10. **README.md (planning)** — 최종수정일 + M0~M6 → M0~M7/v1.0 진입점 반영(경미).

### 4.3 우선순위

- **즉시(문서 drift 해소)**: 권고 1, 2, 3 (Onboarding 도메인 done 정합 — ../02 G1 최우선).
- **단기**: 권고 4, 5, 6, 7 (코드 반영 완료 plan 의 상태 정정).
- **선택**: 권고 8, 9, 10 + R-ADR-2 (design banner / 헤더 정밀화).

---

## 5. 결론

1. **ADR 정합**: 24개 ADR 의 3 supersession 체인(0001→0019, 0020→0021 partial, 0022→0023)이 메타 헤더 + §0 banner + 본문 immutable 보존까지 양방향 정합. **critical cross-ref 누락 없음** → 본 분석에서 ADR Edit 미적용(절차 5 조건 불충족). 선택 보강 후보 2건(R-ADR-1/2)만 권고.
2. **stale planning Top**: release_v1_roadmap.md / onboarding_impl_plan.md / account_user_management_redesign.md 3건이 최우선(Onboarding 도메인 풀스택 done 이 ⏳/draft 로 잔존). 그 다음 external_integration_capability_matrix / view_menu_screen_api_matrix / 27·30·23·18 plan 들.
3. **근본 원인**: 코드·추적성은 2026-05-27 정합됐으나 planning 문서 본문 진행표기가 1주 뒤처진 drift. 위 §4 권고로 정합 가능 — 의미/결정 변경 없는 진행상태 정정이 대부분이라 추적성 ID 영향 없음(N/A).
