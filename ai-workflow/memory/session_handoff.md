# Session Handoff — main (2026-05-21 Onboarding 도메인 종합 + 학습회 자료)

- 문서 목적: main 브랜치 기준 세션 상태와 다음 작업 진입점을 인계한다.
- 범위: 2026-05-21 단일 일자 12 머지 PR — Onboarding 도메인 phase 1~4 closing (#265~#271 + #276) + Carve A backend (#278) + codex deploy refactor (#277) + 학습회 자료 (#280 + #281). **v1.0 release gate (D-25) 잔여 1건** (#214 P1-3 Keycloak group staging-prod [사내 운영자 1회 작업]).
- 대상 독자: 후속 에이전트, 프로젝트 리드, 다음 세션 진입자.
- 상태: M1/M2/M3/M5/M6 done + **M7 Onboarding closing**. ADR-0019 + ADR-0020 + **ADR-0021 신규 발급** (Onboarding self-service + ADR-0020 partial supersession 5 위치). **Onboarding Carve 진척 1/4 done**: Carve A backend (PR #278) ✅. 잔여 carve: B frontend (#273) + C admin UI (#274) + D tests (#275) M-v1.1. Feature flag `DEVHUB_ONBOARDING_GATE_ENABLED` default OFF — main 동작 변경 없음, Carve D acceptance + 1주 staging monitoring 후 별도 hotfix 로 default ON flip. **외부 Keycloak 시나리오** (2026-05-20) + **단일 포트 deploy 정합** (PR #277, `DEVHUB_PUBLIC_BASE_URL`).
- 최종 수정일: 2026-05-21 (sprint claude/learning-session-slideshow-2026-05-21 머지)
- 관련 문서: [v1.0 릴리즈 로드맵](../../docs/planning/release_v1_roadmap.md), [Onboarding IMPL plan](../../docs/planning/onboarding_impl_plan.md), [ADR-0021](../../docs/adr/0021-onboarding-self-service-unit-selection.md), [학습회 자료 HTML](../../docs/learning-session/2026-05-21/index.html), [traceability/report](../../docs/traceability/report.md).
- 브랜치: `main` (HEAD `73f3b30` — PR #281 학습회 slideshow 머지 후).

## 2026-05-21 단일 일자 머지 PR 12건

| PR | sprint / 도메인 | sha | 핵심 |
| --- | --- | --- | --- |
| #265 | onboarding-concept-2026-05-21 | `e9b7543` | concept §5.9 skip-and-resume + §8 #7 결정 |
| #266 | onboarding-requirements-2026-05-21 | `4d882d5` | REQ §5.7 (REQ-FR/NFR-ONBOARD-* 20개) |
| #267 | onboarding-arch-2026-05-21 | `105b835` | UC-ONBOARD + ARCH §9 + API §16 (API-83..86) |
| #269 | onboarding-adr-2026-05-21 | `a2e751a` | ADR-0021 + ADR-0020 partial supersession 5 위치 |
| #270 | onboarding-codex-hotfix-2026-05-21 | `175bf9a` | codex P1 (§16.3 INSERT/UPDATE) + P2 (§6.1 scope) |
| #271 | onboarding-impl-carve-plan-2026-05-21 | `759f101` | IMPL carve plan + RM-ONBOARD-01..04 + reservation fix 14 위치 |
| #276 | onboarding-codex-hotfix2-2026-05-21 | `703b0f3` | codex P2 ADR-0021 본문 ↔ ADR-0020 cross-doc sync |
| **#278** | issue-272-onboarding-backend | `4a77d08` | **Carve A backend** ⚡ — migration 000033 + 5 handler + UT 13 + feature flag |
| #277 | codex/work_260521-a-next-work + claude follow-up | `d730fc6` | Codex deploy refactor (DEVHUB_PUBLIC_BASE_URL + db-migrate + preflight) + claude P2 보완 |
| #280 | learning-session-materials-2026-05-21 | `9fc12cd` | 학습회 자료 5 step + HTML + Chart.js 4.4.0 로컬 |
| **#281** | learning-session-slideshow-2026-05-21 | `73f3b30` | 학습회 자료 slideshow + SVG 다이어그램 (사용자 피드백) |

## 다음 세션 directive (우선순위)

1. **#214 P1-3 Keycloak group staging-prod 적용** (사내 운영자 1회 작업) — v1.0 release gate (D-25, 2026-06-15) 의 마지막 차단 carve.
2. **Onboarding Carve B/C/D 진입** (Gemini 영역, M-v1.1):
   - Carve B (#273) — `/onboarding` page + OrganizationPicker (typeahead + tree) + sessionStorage skip flag + dismissible banner + `(dashboard)/layout` 3-branch gating + `/account` self-service unit edit.
   - Carve C (#274) — `/admin/settings/users` 의 Confirm Review 액션 + pending_review filter.
   - Carve D (#275) — UT mega lifecycle + TC-ONBOARD-* 11 E2E + 6 test seed.
3. **Feature flag default ON flip** — Carve D acceptance + 1주 staging monitoring 후 별도 hotfix PR.
4. **lazy_auto_create.go deletion** — Carve D + default ON flip 후 별도 hotfix PR (M-v1.1 후반). **GitHub issue #284 등록** (P2-12, priority/p2 + worker/claude + domain/auth + type/refactor, M-v1.1) — prerequisite 5 단계 (Carve B/C/D 머지 + flag default ON flip + 1주 monitoring) + deletion scope (lazy_auto_create.go + onboarding_feature_flag.go 2 파일 삭제 + authenticateActor flag 분기 제거 + audit event emit 중단 + UT 정리) + acceptance 9건 + dependencies 다이어그램 명세.

## 학습 5건 (memory 저장)

1. `feedback_stacked_pr_base_merge_autoclose.md` — stacked PR + base merge auto-close 패턴 (PR #268 → #269 recreate).
2. `feedback_adr_supersession_cross_doc_sync.md` — partial supersession 시 양쪽 ADR + traceability §3/§4 동기 검증 (ADR-0021 3-round case).
3. Codex review cycle 2-round (PR #270 hotfix → PR #276 새 P2 유발).
4. Feature flag default-OFF 안전망 (lazy_auto_create flag conditional).
5. npm pack 으로 vendor 추출 (CDN 차단 환경 대응).

## 2026-05-20 sprint -r (본 PR) — housekeeping #9

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `claude/work_260520-r-housekeeping9` | (본 PR) | TBD | **housekeeping #9** — state.json (head_commit `3fdcf33` + status 재작성 + merged_prs_2026_05_20 에 PR #256/#257 2 entry 추가) + session_handoff (header + 본 sprint -r 표 + #215 cancel decision + v1.0 release gate 잔여 2건) + work_backlog (header + 다음 directive 2 우선순위) + sprint -r 본 state. |

### 흡수 2 PR

| PR | sha | sprint / issue | 핵심 |
| --- | --- | --- | --- |
| #256 | `e388868` | sprint -p housekeeping #8 | 9 PR (#243~#255) 흡수 + 개발현황 슬라이드쇼 11 슬라이드 PPT-style |
| #257 | `3fdcf33` | sprint -q HRDB ETL 폐기 정합 **(issue #215 cancelled)** | scripts/hrdb_etl_sync.sh deprecation + keycloak_offboarding_immediacy §3.1 deprecation + ADR-0019 §5.3 (7) decision shift + release_v1_roadmap P1-4 strikethrough |

### 사용자 결정 정착 (2026-05-20)

**외부 Keycloak 시나리오 채택** — DevHub 가 사내 IdP 팀이 별도 운영하는 Keycloak 사용 가정. HR ↔ Keycloak sync 책임이 외부 IdP 팀 (Federation 또는 사내 ETL → Keycloak Admin REST) 으로 이관. ADR-0020 결정 A 의 자연 확장.

영향:
- HRDB ETL cron (sub-carve, PR #184 sprint -p) 폐기 (script header DEPRECATED + design doc decision shift banner)
- DevHub off-boarding sync = sub-carve C event listener (PR #241) 가 정공법 — 외부 Keycloak user disable → admin event polling → user_sync.go::SyncUserProfile → users.status='deactivated' 자동
- service account `devhub-backend` 가 `manage-users` 없이 `view-users + view-events` 만으로 동작 가능 (sub-carve E PR #244)
- frontend `account.service.ts` 폐기 + admin actions UI 제거 (sub-carve B-frontend PR #246)
- `/api/v1/accounts/*` 4 endpoint 폐기 + lazy auto-create (sub-carve B-backend PR #239)

### v1.0 Release Gate (D-26, 2026-06-15) — 잔여 1건

| issue | 영역 | 상태 |
| --- | --- | --- |
| **#210 P0-2 UI polish** | Frontend (Gemini) | ✅ **done** — PR #248 이 성공적으로 E2E 검증 통과 및 main에 머지 완료되었습니다. semantic theme, responsive sidebar, a11y 가 전체 적용되었습니다. |
| **#214 P1-3 Keycloak group staging-prod** | Infra (사내 운영자) | Backlog — Keycloak admin console 1회 작업 (group 4 생성 + composite role assign) |
| ~~#215 P1-4 off-boarding cron~~ | — | **cancelled (2026-05-20)** — 외부 Keycloak 시나리오 |

### 오늘 머지된 PR 총 13건

| PR | sprint/scope | merge |
| --- | --- | --- |
| #243 | sprint -m housekeeping #6 | `2a1c627` |
| #244 | sprint -n sub-carve E (issue #217) | `6810384` |
| #245 | codex/issue-238 + Claude 인계 (issue #238) | `6656c2a` |
| #246 | gemini sub-carve B-frontend (issue #209) | `b1e34bd` |
| #249 | gemini P1-5 e2e (issue #216) | `7e8388d` |
| #251 | gemini P2-4 Bindings (issue #220) | `de69dd4` |
| #252 | gemini P2-5 Topology (issue #221) | `c44c33d` |
| #253 | gemini #247 P1-bug | `4769fc5` |
| #254 | gemini housekeeping #7 | `d92a01e` |
| #255 | codex docs format v2 | `ec487cf` |
| #256 | sprint -p housekeeping #8 | `e388868` |
| #257 | sprint -q HRDB ETL 폐기 정합 (issue #215) | `3fdcf33` |
| ~~#250~~ | ~~codex docs (closed, codex v2 로 대체)~~ | — |
| **#248** | gemini P0-2 UI polish | `05af1c7` |

## 2026-05-20 sprint -p (본 PR) — housekeeping #8 (9 PR 흡수)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `claude/work_260520-p-housekeeping8` | (본 PR) | TBD | **housekeeping #8** — state.json (head_commit `de69dd4` + merged_prs_2026_05_20 에 PR #243~#255 9 entry 추가 + status 재작성) + session_handoff (header + 본 sprint -p 표 + 9 PR 본문 + 다음 directive 갱신) + work_backlog (header + 변경 이력) + sprint -n/-o/-p branch state finalize. main flat memory drift 해소. |

### 흡수 9 PR

| PR | sha | sprint / issue | 핵심 |
| --- | --- | --- | --- |
| #243 | `2a1c627` | sprint -m housekeeping #6 | sprint -k (PR #241) + sprint -l (PR #242) 흡수 |
| #244 | `6810384` | sprint -n-214-service-account-min-role **(issue #217 P2-1 done)** | ADR-0020 sub-carve E 옵션 A 정공법 — 5 commit: organization.go password 분기 제거 + main.go seedLocalAdmin + KeycloakAdminClient write methods 4건 + IdentityAdmin interface 정리 + docs/planning/keycloak_service_account_min_role.md 신규 |
| #245 | `6656c2a` | codex/issue-238-single-port-nginx-v2 → claude/work_260520-o-238-augment **(issue #238 P0-4 done)** | 단일 포트 reverse proxy (ADR-0018) — codex 1차 3 commit + Claude 인계 보강 6 commit (healthcheck + TLS SOP + dev local mode + 404→302 + Dockerfile build args + Keycloak wildcard SOP) |
| #246 | `b1e34bd` | gemini/work_260520-a-209-accounts-cleanup **(issue #209 P0-1 frontend done)** | ADR-0020 sub-carve B-frontend — account.service.ts 폐기 + MemberTable admin actions 제거 + AdminSettingsUsersPage Keycloak Console link + Claude hotfix (broken merge marker + null 가드 + 빈 td) |
| #249 | `7e8388d` | gemini/work_260520-c-216-kratos-keycloak-e2e **(issue #216 P1-5 done)** | P1-5 e2e Kratos→Keycloak 실 코드 전환 — legacy SQL/scripts 5 file 삭제 + setup-keycloak.sh 신규 + dev-up.sh Keycloak 전환 + global-setup.ts 동적 idp_subject sync + Claude hotfix (38MB binary + manage-users → view-events + SQL injection 방어) |
| #251 | `de69dd4` | gemini/work_260520-d-220-bindings-ui-enhancement **(issue #220 P2-4 done)** | Bindings UI 강화 — Backend PATCH/DELETE + Frontend ComboBox + EditModal + pagination + Claude hotfix (interface method + routePermissionTable + e2e spec 정합 + codex P2×2: pagination clamp + ComboBox button type) |
| #252 | `c44c33d` | gemini/work_260520-e-topology-v2-websocket-grouping **(issue #221 P2-5 done)** | Topology v2 강화 — React Flow Environment 그룹화 + WebSocket 실시간 + Claude hotfix (realtime default unsupported event + isGrouped refetch 분리) |
| #253 | `4769fc5` | gemini/work_260520-f-247-user-creation-password-cleanup **(issue #247 P1-bug done)** | UserCreationModal password 필드 제거 + Claude rebase 회귀 정정 (codex #238 auth 라우팅 회귀 → 2 file 만 retain) |
| #254 | `d92a01e` | gemini/housekeeping-260520-status-update | housekeeping #7 — main flat memory 동기화 + Claude hotfix (state.json trailing garbage 제거) |
| #255 | `ec487cf` | codex/workflow_refactoring_v2 | docs format/lifecycle 표준화 (Markdown source-of-truth + HTML derived) — 50+ docs file 메타 헤더 + Claude hotfix (docs/setup 4 file 회귀 복원) |

### ADR-0020 sub-carve 진척 (5/8 → 6/8)

| sub-carve | 상태 | sprint / PR |
| --- | --- | --- |
| A | ✅ done | sprint -d PR #205 |
| B-backend | ✅ done | sprint -i PR #239 |
| B-frontend | ✅ done | **gemini PR #246** |
| C | ✅ done | sprint -k PR #241 |
| D | ✅ done | sprint -l PR #242 |
| E | ✅ done | **sprint -n PR #244** |
| F /login 정리 | carve | sprint -? 후속 (P3) |
| SPI provider JAR | carve | 사내 인프라 동반 (P2) |

### v1.0 Release Gate (D-26, 2026-06-15)

| issue | 영역 | 상태 |
| --- | --- | --- |
| **#210 P0-2 UI polish** | Frontend (Gemini) | **In review — PR #248 e2e shard 2 3번째 fail**. Header trigger 영역 fundamental incompatibility 의심. Playwright trace artifact 분석 필요 |
| **#214 P1-3 Keycloak group staging-prod** | Infra (사내 운영자) | Backlog — Keycloak admin console 1회 작업 |
| **#215 P1-4 off-boarding cron** | Infra (사내 운영팀 + Codex) | Backlog — scripts/hrdb_etl_sync.sh 실 deploy |

### 금일 진척 PR (8건)

| PR | sha | sprint / issue | 핵심 |
| --- | --- | --- | --- |
| #245 | `7e8388d` | issue-238-single-port | **ADR-0018 single-port reverse proxy 고도화** — nginx upstream 안정화 및 health check 도입. (Merged) |
| #246 | `e6a2b3c` | sub-carve B-frontend | **ADR-0020 sub-carve B (frontend) cleanup** — `account.service.ts` 삭제 및 admin UI actions 제거. (Merged) |
| #249 | `f9b8a7d` | p1-5-e2e-transition | **e2e Kratos legacy 제거** — dynamic idp_subject sync 도입으로 e2e 테스트 안정성 확보. (Merged) |
| #248 | issued | p0-2-ui-polish | **P0-2 UI polish 1차** — semantic theme 적용, responsive sidebar, a11y 개선. |
| #251 | issued | p2-4-bindings-ui | **P2-4 Bindings UI** — scope_id lookup 지원 및 CRUD 연동. |
| #252 | issued | p2-5-topology-v2 | **P2-5 Topology V2** — WebSocket 실시간 업데이트 및 노드 그룹화. |
| #253 | issued | issue-247-p3-1-cleanup | **Issue #247 + P3-1** — 사용자 생성 password 필드 제거 및 /login 페이지 정리. |

### ADR-0020 sub-carve 진척 (5/8 → 7/8)

| sub-carve | 결정 | 상태 | sprint / PR |
| --- | --- | --- | --- |
| A | `rbac_subject_roles` 완전 제거 (결정 D) | ✅ done | sprint -d / PR #205 `f2a389a` |
| B-backend | `/api/v1/accounts/*` 폐기 + lazy auto-create | ✅ done | sprint -i / PR #239 `d21e801` |
| B-frontend | `account.service.ts` 폐기 + admin/settings/users + e2e | ✅ done | PR #246 `e6a2b3c` |
| C | event listener 확장 + users sync + metric 3종 | ✅ done | sprint -k / PR #241 `9ea7e1c` |
| D | JWKS stale-while-error expiry case 확장 | ✅ done | sprint -l / PR #242 `cb6646d` |
| E | service account 권한 축소 + keycloak_operations §8.5c | carve | sprint -n 후속 |
| F | `/login` page 정리 (결정 B entry minimal 유지) | ✅ done | PR #253 |
| (신규) | Keycloak SPI provider JAR (PR #203 codex P2 carve out) | carve | 사내 인프라 동반 |

## 다음 세션 권장 진입

| 우선순위 | sprint 후보 | 작업 | 워커 |
| --- | --- | --- | --- |
| 1 | issued PR 머지 후 안정화 | PR #248, #251, #252, #253 머지 확인 및 e2e 회귀 검증 | Gemini/Claude |
| 2 | issue #214 P1-3 sub-carve E | service account 권한 축소 + governance 협약 SOP 갱신 | Claude |
| 3 | M4 RM-M4-XX 본격 진입 | WebSocket replay, AI Gardener gRPC 등 로드맵 과제 진입 | Claude/Codex |
| 4 | Keycloak SPI provider JAR | PR #203 codex P2 후속 — SPI 빌드 및 운영 SOP 작성 | Codex |

## 2026-05-20 sprint -m (PR #242, `cb6646d`) — housekeeping #6 (이전 finalize)

## 2026-05-20 sprint -j (PR #240, `80120c8`) — housekeeping #5 (이전 finalize)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `claude/work_260520-j-housekeeping` | #240 | `80120c8` | **housekeeping #5** — state.json (head_commit `d21e801` + merged_prs_2026_05_20 5 entry 추가 + `github_project_setup_2026_05_20` 객체 신규 + sub_carve_split 정정 sub-carve B backend done) + session_handoff + work_backlog + sprint -f/-g/-h/-i state finalize + 본 sprint -j state. main flat memory drift 해소. |

### 흡수 5 PR

| PR | sha | sprint / issue | 핵심 |
| --- | --- | --- | --- |
| #207 | `a08ffd3` | sprint -f-roadmap | release_v1_roadmap.md 신규 (306 lines, v1.0 scope + 30+ carve P0~P3 + 마일스톤 + 워커 분업 + GitHub plan) + worker_division.md 신규 (180 lines) + AGENTS/GEMINI 진입점 갱신 |
| #208 | `bb4a0f4` | sprint -g-codex-hotfix | codex P2 hotfix — worker_division.md §4.2 ADR reversal dead link → 5 step 표준 절차 본문 명시 + ADR-0019/0001 supersession canonical 사례 |
| #236 | `64d37f5` | sprint -i-cancel-signup | P3-12 Sign Up 영구 취소 (사용자 결정 — DevHub Keycloak admin 권한 없음). release_v1_roadmap.md §1.2/§3.4/§4.2/§5.2 strikethrough + issue #235 closed |
| #237 | `c202fa0` | sprint -h-screenshot **(issue #211, P0-3 done)** | Playwright screenshot mode — screenshots.spec.ts 신규 (19 페이지 fullPage, 4 test) + Upload UI Screenshots CI step + Stage 3 보강 2회 (timeout 180s + codex P1 networkidle 5s) |
| #239 | `d21e801` | sprint -i-209-accounts-deprecation **(issue #209, P0-1 backend done)** | ADR-0020 sub-carve B (backend) — /api/v1/accounts/* 4 endpoint 폐기 + lazy_auto_create.go 신규 + AuthenticatedActor Email/DisplayName + 신규 audit action 2종 + 회귀 test 5건 + 기존 test 3건 admin pre-seed fix + docs 4 file 정합 |

### GitHub 자산 등재 (사용자 + 본인 협업)

- **Project** `Devhub Example 1차 개발` (`PVT_kwHOCKCvOs4BYOna`, number 2) — 사용자 생성
- **Milestones** 3개 — v1.0 Release (2026-06-15) / v1.1 Stability (2026-07-31) / v2 Extension
- **Issues** 27개 (P0-1 ~ P0-3 + P1-1 ~ P1-5 + P2-1 ~ P2-7 + P3-1 ~ P3-11 + P0-4 신규 + P3-12 cancelled)
- **Labels** 19개 — priority/{p0..p3} + worker/{claude,codex,gemini,user} + domain/{auth,app-repo-project,dreq,integration,ui-polish,infra} + type/{feature,refactor,test,docs,ci}
- **Status 분류**: Ready 3 (#209/#210/#211→Done/#238) / In review 0 / Backlog 23 (P1~P3)
- **권한**: `gh auth refresh -s read:project,project` 사용자가 별도 터미널 실행 후 본인이 26 issue 일괄 추가 + Status 분류

### branch naming 신규 규칙 (worker_division.md §2.5)

`<worker>/work_<YYMMDD>-<sprint-seq>-<issue-num>-<short-key>` 패턴:
- 본 sprint -i 가 첫 적용 사례 (`claude/work_260520-i-209-accounts-deprecation`)
- 예외: housekeeping/hotfix (issue 없음, key 만): `claude/work_260520-j-housekeeping`, `claude/work_260520-g-codex-hotfix`
- 외부 contribution 자유 branch 허용

## 다음 sprint 권장 진입

| 우선순위 | sprint 후보 | 작업 | 워커 |
| --- | --- | --- | --- |
| 1 | sub-carve B frontend cleanup | `account.service.ts` 폐기 + `/admin/settings/users` admin actions 제거 + e2e TC-ACC-* 갱신 | Gemini |
| 2 | issue #210 P0-2 UI polish | semantic theme + responsive + a11y. screenshot artifact (PR #237 결과) review source | Gemini |
| 3 | issue #214 P1-3 sub-carve E | service account 권한 축소 (`manage-users` 제거) + governance 협약 SOP `keycloak_operations.md §8.5c` 신규. docs+SOP only, 낮은 위험 | Claude (sprint -n) |
| 4 | issue #215 P1-4 sub-carve F | `/login` page 정리 (결정 B entry minimal). 가장 낮은 우선순위 | Claude (sprint -o) |
| 5 | issue #238 P0-4 docker single-port review | Codex 작업 branch 발견 시 review 모드 | Claude (review only) |
| 6 | Keycloak SPI provider JAR (P2-6 carve) | PR #203 codex P2 후속 — devhub-event-listener SPI 빌드 + compose mount + 운영 SOP | 사내 인프라 동반 | sprint -d (PR #205 `f2a389a`) + 외부 인수 PR #203 (`a294baf`) 2 PR 흡수. state.json (head_commit `a294baf` + merged_prs_2026_05_20 에 PR #205/#203 추가 + sub_carve_split sprint label shift + external_carve_keycloak_spi_provider_jar 신규) + session_handoff (header + sprint -e 표 + sprint -d finalize + PR #203 인수 row + sub-carve B 진입 directive) + work_backlog (header + 변경 이력 2 row) + sprint -d branch state finalize (in_progress → done, merge_commit f2a389a) + sprint -e branch state. design doc §6.1 sprint label 정정 (sub-carve B → sprint -f 로 shift). |

### PR #205 sprint -d (`f2a389a`) finalize — 계정/사용자 관리 리팩토링 Phase 3 sub-carve A

| 단계 | 내용 |
| --- | --- |
| ADR-0020 발급 | accepted, Phase 2 명시 결정 6건 명문화 (A 전면 폐기 / B `/login` minimal / C event listener 확장 / D `rbac_subject_roles` 완전 제거 / E read-only self-reverse / F JWKS expiry 확장) |
| design doc §6 신규 | sub-carve A~F 분담 + Strangler Fig 패턴 |
| `rbac_subject_roles` 폐기 | 8 파일 변경 (handler 2개 + interface method 2개 + audit const + wire struct + 2 route + permissions entry + store impl 2개 + test cleanup) |
| Stage 3 보강 #1 | frontend rbac.service.ts dead method + backend_api_contract §12 정리 + design doc §6.4 IMPL 정정 |
| Stage 3 보강 #2 (codex P1) | `validAppRoles` 에 `pmo_manager` 추가 (sprint -f 의 event listener sync 가 정공법, sprint -d 는 backward compat) + 회귀 test 2건 (CreateUser + UpdateUser) + ADR-0020 §5.5 hotfix 섹션 |

### PR #203 (gemini → 본인 인수, `a294baf`) — Keycloak SPI webhook + semantic theme

| 단계 | 내용 |
| --- | --- |
| 원본 (gemini) | (1) `seedLocalAdmin` `temporary:false` + audience mapper → Keycloak 로그인 endless loop 해소, (2) 다수 modal hardcoded color → semantic theme variable (bg-warning/primary/destructive), (3) Keycloak SPI webhook (`/api/v1/internal/keycloak-events`) + handler + test 신규 |
| 본인 인수 fix `6ece5be` | (a) `TestRoutePermissionTable_CoversAllProtectedV1Routes` 회귀 해소 — `permissions.go` Bypass 섹션에 `POST /api/v1/internal/keycloak-events` entry 추가 + 주석. (b) codex P1 응답 — secret fail-closed (미설정 시 503, KratosWebhookToken 패턴 정합). (c) 회귀 test `TestReceiveKeycloakEventWebhook_SecretNotConfigured` 신규. main rebase (sprint -d 위, conflict 없음). |
| codex P2 carve out | `infra/idp/keycloak-realm.json` 의 `devhub-event-listener` SPI 가 provider JAR 미동반 → 사내 인프라 동반 carve: (a) Keycloak SPI plugin 빌드 + (b) compose volume mount + (c) 운영 SOP. ADR-0020 §6 미해결 carve 등재 예정. |

## 다음 sprint -f 진입 directive — sub-carve B (`/api/v1/accounts/*` 폐기 + lazy auto-create)

ADR-0020 §4.1 sub-carve B 본격 진입. 권장 작업 단위:

1. **backend handler 제거** — `accounts_admin.go` 의 4 endpoint (POST `/api/v1/accounts` createAccount, PUT `/api/v1/accounts/:user_id/password` resetAccountPassword, PATCH `/api/v1/accounts/:user_id` updateAccountStatus, DELETE `/api/v1/accounts/:user_id` deleteAccount) 모두 제거
2. **router.go** v1.POST/PUT/PATCH/DELETE accounts 4 route 제거
3. **permissions.go** 4 routePermissionTable entry 제거
4. **`KeycloakAdminClient` write 메서드 호출처 제거** — `CreateIdentity` / `UpdateIdentityPassword` / `SetIdentityState` / `DeleteIdentity` 메서드 자체는 보존 (lazy auto-create 가 향후 사용 가능)
5. **`authenticateActor` lazy auto-create 실 구현** — ADR-0020 §5.2 따름. `users` row 자체가 없을 때 `CreateUser` + audit `account.lazy_provisioned`. role 매핑 = `extractKeycloakRole(token.realm_access.roles)` (sprint -j PR #185 multi-role priority filter 공유) + fallback default `developer` + audit `user.role_default_assigned` (ADR-0020 §5.2.2 P1-3 결정)
6. **frontend `account.service.ts` 파일 제거** — 5 메서드 모두 폐기
7. **frontend `/admin/settings/users` page** — "Issue Account" 버튼 + modal 'Issue / Reset / Disable' 액션 제거. user list view + role/unit assignment 만 남김
8. **e2e spec** — TC-ACC-* 의 admin issue/disable 시나리오 제거 + Keycloak admin console flow 안내 link 검증으로 대체
9. **traceability 영향** — API-25 (POST /accounts) + API-25-PWD/STATUS/DELETE 폐기 strikethrough + IMPL-account-* 정리 + TC-ACC-* 갱신
10. **Strangler Fig 옵션 검토** — design doc §6.5 의 deprecation banner 단계 적용 여부 (외부 caller 없으면 즉시 제거 가능, frontend 만 caller 이므로 같은 PR 에서 frontend cleanup 동반 권장)

위험: e2e 회귀 (TC-ACC-*). 검증 = backend `go test ./...` + frontend `npm run build` + e2e Playwright shard 1/2 모두 PASS.

### sub-carve B~F (sprint label shift — sprint -e housekeeping #4 으로 소진)

| sub-carve | sprint | 영역 | 위험 |
| --- | --- | --- | --- |
| **B** | `-f` | `/api/v1/accounts/*` 4 endpoint 제거 + `KeycloakAdminClient` write 메서드 호출처 제거 + `authenticateActor` lazy auto-create 실 구현 + frontend `account.service.ts` 폐기 + admin/settings/users page + e2e spec | 중간 (e2e 회귀) |
| **C** | `-g` | event listener 확장 (USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑) + `users` write + metric 3종 (`devhub_keycloak_user_sync_*`) | 중간 (event listener 회귀) |
| **D** | `-h` | JWKS stale-while-error expiry case 확장 + cache flush SOP | 낮음 |
| **E** | `-i` | service account 권한 축소 (`manage-users` 제거) + governance 협약 SOP `keycloak_operations.md §8.5c` 신규 | 낮음 (docs only + Keycloak admin SOP) |
| **F** | `-j` | `/login` page 정리 (결정 B, 우선순위 가장 낮음) | 낮음 |
| (신규 carve) | TBD | **Keycloak SPI provider JAR** — PR #203 codex P2 후속 (devhub-event-listener 빌드 + compose mount + 운영 SOP). 사내 인프라 동반 carve | 중간 (사내 인프라) |

## 2026-05-20 housekeeping #3 (sprint claude/work_260520-c, PR #204 `12bb557`) — sprint -ad finalize + 2026-05-20 sprint -a/-b + 외부 PR #201/#202

본 housekeeping sprint 가 흡수한 5 PR (#198~#202) 종합:

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `claude/work_260519-ad` | #198 | `bcca86a` | **Kratos 잔재 residual cleanup (ADR-0019 정공법)** — 직전 handoff 시점 PR TBD 였음. 본 housekeeping 으로 finalize. 상세는 아래 sprint -ad 표 row 참조. |
| `claude/work_260520-a` | #199 | `b0bcced` | **계정/사용자 관리 리팩토링 Phase 1 — 현황 파악 매트릭스** (`docs/planning/account_user_management_redesign.md` 신규, 235 lines). §1 책임 분리 매트릭스 13 row + §1.1 source-of-truth 이중화 issue 3건 + §2 backend 17 endpoint 매트릭스 + §2.1 IdentityAdmin 5 메서드 → Keycloak Admin REST 매핑 + §3 frontend 4 page + §4 DB schema + §4.5 Phase 2 입력 옵션 A~D. 핵심 발견 5건: role 이중 source-of-truth / status 이중 source / POST `/accounts` vs POST `/users` 중복 / dead frontend code (unlockAccount, deleteAccount) / `PUT /rbac/subjects/:id/roles` UI 미구현. |
| `claude/work_260520-b` | #200 | `95d6909` | **계정/사용자 관리 리팩토링 Phase 2 — 책임 분리 design + 명시 결정 6건 확정** (`docs/planning/account_user_management_redesign.md` §5 신규 10 sub-section, 270+/16- LoC). 결정 6건: A 전면 폐기 / B `/login` minimal entry 유지 / C event listener 확장 + lazy hot path / D `rbac_subject_roles` 완전 제거 / E read-only mode self-reverse / F JWKS stale-while-error expiry case 확장. §5.1 결정 표 + §5.2 lazy auto-create + §5.3 event listener 매핑 표 (USER:CREATE/UPDATE/DELETE/GROUP_MEMBERSHIP/RESET_PASSWORD/DISABLE_CREDENTIALS) + §5.4 frontend cleanup + §5.5 service account 권한 축소 SOP + §5.6 JWKS expiry case + §5.7 `/login` 정리 + §5.8 `rbac_subject_roles` 폐기 + §5.9 Phase 1 매트릭스 오류 정정 (테이블 자체 없음) + §5.10 ADR-0020 후보 outline. |
| `codex/keycloak-only-refactor-plan` (post-#200) | #201 | `cff97d4` | **e2e 로그인/로그아웃/signup/audit 시나리오 안정화** — 외부 codex 작성. main PR #200 conflict 해소 (`bae9990`) + e2e/auth Keycloak login + runtime OIDC defaults (`08fab34`) + Phase 2 design duplicate (`09fefb3`). 변경 영역: frontend e2e (audit/auth/signout/signup) spec + `lib/services/auth.service.ts` + `global-setup.ts` + `fixtures.ts` + `api/runtime-config/route.ts`. frontend e2e 전체 47 passed. PR body: 작업 중 로컬 DB 에 000030 rename migration 적용해 users.idp_subject 정합. |
| `ci-fix` | #202 | `abc8cc9` | **Keycloak E2E CI pipeline 정합 + sync guard 신규 (큰 PR)** — 외부 본인 7 commit: `a66f67d` (e2e pipeline Keycloak 정합 + `scripts/ci-e2e-sync-check.sh` 51 lines 신규) + `9b288e5` (runner 호환) + `348d149` (Keycloak service-account admin role 부여) + `0c69f9c` + `257ae70` (audience mapper for e2e tokens) + `af743bc` (signout user-switch flow timeout relax) + `e285326` (signout 후 session clear 대기). 변경 영역: `.github/workflows/ci.yml` 210 lines 갱신 + `scripts/ci-e2e-sync-check.sh` 신규 + frontend e2e fixtures/global-setup. **잔존 검증**: 본 PR 이 잔여 carve "e2e Kratos → Keycloak 실 코드 전환" 의 CI 단 부분 해소 — sprint -m design (`e2e_keycloak_migration.md`) 정합 cross-check 권장. |
| `claude/work_260520-c` | #204 | `12bb557` | **main flat memory housekeeping #3** — 위 5 PR (#198~#202) 일괄 흡수. state.json (head_commit abc8cc9 + merged_prs_2026_05_19 에 PR #197/#198 finalize + merged_prs_2026_05_20 신규 4 PR + account_user_management_redesign_2026_05_20 객체 신규 + external_pr_2026_05_20_ci_e2e_keycloak 객체 신규) + session_handoff (본 표 + 다음 directive 갱신) + work_backlog (header + 변경 이력 5 row) + auto-memory project_2026_05_20_post_198_housekeeping.md. sprint -ad / -a / -b branch state.json finalize 3개. |

## 계정/사용자 관리 리팩토링 — Phase 3 (실 구현) carve out 8건

Phase 1 (PR #199) 현황 파악 + Phase 2 (PR #200) design 완료 — 실 구현 Phase 3 은 별도 sprint:

1. **backend code 제거** — `accounts_admin` handler 4 + `KeycloakAdminClient` manage-users 메서드 4 호출처 모두 제거 (결정 A)
2. **`authenticateActor` lazy auto-create 실 구현** — 첫 로그인 시 DevHub `users` row 자동 INSERT (Keycloak claim 기반)
3. **Keycloak event listener 확장** — sprint -u~-y 의 audit event puller 에 `USER:UPDATE` / `GROUP_MEMBERSHIP` / `USER:DELETE` 매핑 추가 + DevHub `users` write (결정 C)
4. **JWKS stale-while-error expiry case 확장** — sprint -r 의 kid mismatch fallback 패턴 정합 (결정 F)
5. **frontend cleanup** — `account.service.ts` 폐기 + `MemberTable.tsx` 정리 + `/admin/settings/users` UI 정리
6. **service account 권한 축소** — Keycloak admin SOP 갱신 (manage-users 제거, governance 협약 보강)
7. **governance 협약 SOP** — `keycloak_operations.md §8.5c` 신규 (9 운영 동작 책임 분리 표)
8. **ADR-0020 draft 작성** — 계정/사용자 관리 책임 경계 (옵션 A 채택 명문화) + 사내 검토

## 2026-05-20 sprint -ad — Kratos 잔재 residual cleanup (PR #198, `bcca86a`)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-ad` | #198 | `bcca86a` | **Kratos 잔재 residual cleanup (ADR-0019 정공법)** — backend 11 파일 삭제 (`account_password.go` + `kratos_login_client/settings_client/session_cache/admin_client` + `password_auth_types` + 각 `_test`) + `router.go`/`permissions.go` 의 `KratosLogin`/`IDPSessionCache` field + `/account/password` route 일괄 제거 + `identity_resolver.go` 신규 + `identity_admin_mock_test.go` 신규 (MockIdentityAdmin test-only 분리). `main.go` 변수명 `kratosAdmin` → `idpAdmin`, `main_test.go` `kratosAdminFake` → `idpAdminFake` rename. `domain.AuditSourceKratos` enum + audit action 이름 (`account.issue.kratos_failed` 등) + audit details key (`kratos_id`) 는 DB historical row 정합 보존, comment 만 generic 화. frontend `/account/page.tsx` 의 password form 제거 + Keycloak Account Console (`${OIDC_ISSUER_URL}/account/`) 외부 link 카드 대체. `account.service.ts` 의 `updateMyPassword`/`SettingsFlowError` 류 제거 (admin 메서드 5개 보존). `account.service.test.ts` 삭제. `account.spec.ts` password 시나리오 삭제 + `TC-ACC-KEYCLOAK-CONSOLE-01` 신규. `endpoints.ts` 의 dead export `KRATOS_ADMIN_URL_SERVER` 제거. `docs/traceability/report.md` §6 + `docs/setup/keycloak_operations.md` §8.5b 신규 + §11 변경 이력 row 추가. backend `go test ./...` + frontend `npm run build` + vitest 8 file/34 test 모두 그린. |

**변경 통계**: 31 파일 (11 D / 13 M / 2 신규 / 1 R), 239+ / 2939- LoC.

**핵심 발견**: `/api/v1/account/password` endpoint 는 wire 누락 dead path 였음 — `RouterConfig.KratosLogin` field 가 `main.go` 에서 set 되지 않아 (PR #167 KC-PR-D 가 wire 제거) `h.cfg.KratosLogin == nil` 으로 503 영구 응답. 사용자 confirm 으로 옵션 A (Keycloak Account Console redirect) 채택. self-service 비밀번호 변경 자체 proxy 코드는 **절대 부활시키지 않음**.

**보존 항목 (의식적)**:
- `domain.AuditSourceKratos` enum + audit action 이름 (`account.issue.kratos_failed`/`kratos_promoted` 등) + audit details key (`kratos_id`) — DB historical row 정합. rename 은 dual-write/dual-read carve.
- e2e `global-setup.ts` 의 Kratos admin API 직접 호출 — sprint -m design (e2e_keycloak_migration.md) 따른 별도 carve 로 보존 (사내 staging Keycloak 환경 동반).
- `infra/idp/` 의 historical Kratos/Hydra yaml — PR #169 deprecation banner 부착 후 보존.

## 2026-05-19 Phase 4 — ADR-0019 §5.3 (9) audit event listener Phase 2 풀스택 (sprint -u ~ -y, 5 PR)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-u` | #189 | `d6e0f99` | **PR-B skeleton** — migration 000031 `event_cursors` + `EventCursorStore` + `KeycloakAdminClient.List(User\|Admin)Events` (codex review #9 정정 정합 — `?dateFrom=` + `/admin-events`) + `RunKeycloakEventPuller` cron + 매핑 표 22 row + SHA256 dedup + 14건 test |
| `-v` | #190 | `2ddece2` | **PR-C wire + metric + integration test** — `domain.AuditSourceKeycloakEvent` enum + `audit/metrics.go` Prometheus 3종 (`devhub_keycloak_events_processed_total` / `_cursor_lag_seconds` / `_pull_errors_total`) + `HTTPAPIEventLister` adapter + main.go cron wire (`DEVHUB_KEYCLOAK_EVENT_LISTENER_*` 3 env) + `pullOnce` firstErr 패턴 + 5건 test |
| `-w` | #191 | `49bfb92` | **PR-D store-level dedup** — migration 000032 `audit_logs.source_event_id` + partial UNIQUE INDEX (source_type, source_event_id) WHERE NOT NULL AND source_type IS NOT NULL + `domain.AuditLog.SourceEventID` + `ON CONFLICT DO NOTHING` + `getAuditLogBySourceEventID` lookup + 3건 store integration test (DEVHUB_TEST_DB_URL skip pattern) |
| `-x` | #192 | `a72bde4` | **PR-E 운영 SOP** — `keycloak_operations.md §8.6` 신규 9 sub-section (활성화 사전 조건 / Keycloak admin Events 설정 / backend env / Prometheus dashboard 4 panel + PromQL / 알람 3종 / dedup 확인 SQL / 트러블슈팅 5 케이스 / disable·rollback / sub-carve) |
| `-y` | #193 | `f3f640b` | **codex review hotfix #10** — P1×3 (skip-only cursor advance + initial cursor 즉시 영구 seed + §8.6.1 seed 명시) + P2×2 (hash 7-tuple 확장 + Expiration wording 정정) + 4건 신규 test + PR #189 P2 same-ms false positive 응답 |
| `-z` | #194 | `61dcbb4` | main flat memory housekeeping #1 (sprint -u~-y 5 PR 흡수 + Phase 4 milestone block 신규). |
| `-aa` | #195 | `d9bcd82` | **codex hotfix #11** — PR #193 same-ms boundary 에서 emit-able event 의 hash 우선 (`latestEmittable bool` flag 도입). skip event hash 가 latestHash 로 set 되어 다음 tick 에서 emit-able event 가 re-emit + metric inflation 되던 side effect 해소. 신규 test 2건. |
| `-ab` | #196 | `86e5b0d` | **codex hotfix #12** — PR #194 codex P2 정정. sprint -n~-t 7 branch state.json finalize 누락 정정 (`status: in_progress` → `done` + ended_at + merged_pr + merge_commit 추가). sprint -z + -aa 본인도 finalize. work_backlog row wording 정정. AGENTS.md 세션 복원 위험 회피. |
| `-ac` | (본) | TBD | main flat memory housekeeping #2 (sprint -z + -aa + -ab 3 PR 흡수). |

## 2026-05-19 Phase 3 critical fixes (sprint -j ~ -t, 11 PR)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-j` | #178 | `22a1c15` | **codex review hotfix #9** — sprint -b~-h 7 PR P2 inline 9건 일괄 흡수. backend 확장 carve 4건 추가. |
| `-k` | #179 | `b3f8153` | **infra Keycloak 정합** — docker-compose + nginx |
| `-l` | #180 | `38e09aa` | **CRITICAL migration 000021 prefix 충돌 정정** — git mv → 000030 + 운영 DB 정정 SOP |
| `-m` | #181 | `12cdf81` | **frontend e2e Kratos → Keycloak design** — ADR-0019 §5.3 10/10 milestone |
| `-n` | #182 | `871d6f0` | main flat memory housekeeping (sprint -j~-m 흡수 + Phase 3 critical fixes 명문화) |
| `-o` | #183 | `1de2253` | CI `migration-prefix-lint` job 신규 (sprint -l §6 carve) |
| `-p` | #184 | `4cc6f2a` | off-boarding Phase 1 운영 cron — `scripts/hrdb_etl_sync.sh` 신규 (sprint -g design §3.1 옵션 C) |
| `-q` | #185 | `db5addf` | backend 확장 #1 — `extractKeycloakRole` multi-role priority filter (codex review #9 #6) |
| `-r` | #186 | `8470302` | backend 확장 #2 — JWKS kid mismatch stale-while-error fallback (codex review #9 #3) |
| `-s` | #187 | `5c36653` | backend 확장 #3 — basePath 포함 logout URI (codex review #9 #4) |
| `-t` | #188 | `cebf255` | backend 확장 #4 마지막 — `authenticateActor` 자동 idp_subject sync (codex review #9 #2). **backend 확장 carve 4건 모두 resolved**. |


## 2026-05-19 ADR-0019 §5.3 design 완결 milestone (sprint -a~-i 누적 9 PR)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-a` | #169 | `3018927` | **ADR-0019 재산정** — Keycloak 단일화 사후 명문화 ADR 신규 + ADR-0001 supersession 처리 (8개 inline banner) + 14 추가 정합 docs |
| `-b` | #170 | `d442e17` | main flat memory housekeeping — sprint -a 흡수 + 외부 누락 PR #165~#168 사후 등록 + auto-memory entry |
| `-c` | #171 | `23602f5` | **ADR-0019 §5.3 (1)+(2)+(3) SOP — Keycloak operations SOP 단일 통합 문서** (`docs/setup/keycloak_operations.md`, 11 section): realm/client/role + JWKS rotation + employee_id custom claim |
| `-d` | #172 | `d11917b` | **ADR-0019 §5.3 (4) SSO logout chain (RP-initiated) SOP** (keycloak_operations.md §8.5): frontend `auth.service.ts:100-126` 인용 + admin console SOP + chain order + 보안 4 위협 |
| `-e` | #173 | `6245bda` | **ADR-0019 §5.3 (9) audit event listener design** (`docs/planning/keycloak_event_audit_integration.md`): 옵션 3종 + 권장 B admin event polling + audit_logs action 매핑 25 row + 구현 PR-A..E + ADR-0020 후보 |
| `-f` | #174 | `947bd2f` | **ADR-0019 §5.3 (8) groups → RBAC role 자동 매핑 design** (`docs/planning/keycloak_groups_rbac_mapping.md`): 옵션 4종 + 권장 B group composite + **backend 변경 없음** + keycloak_operations §4.3/§8.1 갱신 |
| `-g` | #175 | `aa0c029` | **ADR-0019 §5.3 (7) off-boarding 즉시성 design + ADR-0008 §6 통합** (`docs/planning/keycloak_offboarding_immediacy.md`): 옵션 6종 + Phase 1 HR ETL push + Phase 2 LDAP federation + keycloak_operations §8.2 보강 |
| `-h` | #176 | `455556b` | **ADR-0019 §5.3 (6) Keycloak failover design — §5.3 design 완결 milestone** (`docs/planning/keycloak_failover.md`): 옵션 6종 + Phase 1 graceful degradation + Phase 2 HA active-active + 옵션 E backup IdP 명시 제외 (ADR-0019 충돌) |
| `-i` | (본) | TBD | main flat memory housekeeping — sprint -c~-h 6 PR 누적 흡수 + ADR-0019 §5.3 design 완결 milestone 명문화 + auto-memory entry + sprint -c~-h state finalize 6개 |

## ADR-0019 §5.3 design 완결 milestone (MFA 제외, 7/7 = 100%)

| # | 항목 | 상태 | 출처 |
| --- | --- | --- | --- |
| 1 | realm/client/role SOP | ✅ SOP | -c keycloak_operations §2~§4 + §7 + §8 |
| 2 | JWKS rotation SOP | ✅ SOP | -c §6 (rotation 주기 + cache + 비상 §6.5) |
| 3 | Keycloak ↔ HRDB sync (employee_id) | ✅ SOP | -c §5.2 (custom claim 매핑) |
| 4 | SSO logout chain (RP-initiated) | ✅ SOP | -d §8.5 (frontend 인용 + admin console + chain) |
| 5 | MFA 도입 | ❌ excluded | 사내 정책 |
| 6 | Keycloak failover (HA) | ✅ design | -h keycloak_failover.md (Phase 1 graceful + Phase 2 HA) |
| 7 | off-boarding 즉시성 | ✅ design | -g keycloak_offboarding_immediacy.md + ADR-0008 §6 통합 |
| 8 | groups → RBAC role 자동 매핑 | ✅ design | -f keycloak_groups_rbac_mapping.md (backend 무변경) |
| 9 | audit event listener | ✅ design | -e keycloak_event_audit_integration.md (ADR-0020 후보) |

**잔여 design+carve (실 구현)**
- ✅ resolved (sprint -u~-y, PR #189~#193) — **audit event listener 실 구현 Phase 2 종결** (PR-B skeleton + PR-C wire+metric + PR-D store dedup + PR-E 운영 SOP + codex hotfix #10)
- ✅ resolved (sprint -p, PR #184) — **off-boarding Phase 1 운영 cron — `scripts/hrdb_etl_sync.sh` 신규**. 실 deploy 는 사내 운영팀 동반 carve.
- group staging-prod 적용 — Keycloak admin 1회 작업 (group 4 생성 + composite role assign)
- HA Phase 2 — 사내 인프라 결정 동반 (Infinispan + shared PG + LB, ADR-0021 후보)
- e2e Kratos → Keycloak admin migration 실 코드 전환 — 사내 staging 동반 (sprint -m design)
- audit event listener SPI push 전환 (polling latency 30s → < 1s, sprint -x §8.6.9 sub-carve)



## 2026-05-18 외부 누락 PR 사후 등록 (post-EOD #2 이후)

직전 handoff (sprint -x EOD) 시점에는 PR #164 까지만 명시. 이후 외부 PR 4건이 누락된 상태로 main 머지됨:

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `claude/work_260518-x` | #165 | `3e25628` | **post-EOD #2 종합 housekeeping + 로드맵 갱신** (외부 ADR-0019 후보 메모 + sprint -r..w sync) — 본인 sprint 머지 |
| `gemini/reverse-proxy` | #166 | `694e694` | **단일 외부 포트 역프록시(Nginx) 구성 및 basePath 구현** — ADR-0018 실 구현. nginx skeleton + Next.js basePath /devhub + endpoints.ts dynamic + Same-Origin CORS 무력화 |
| `codex/keycloak-only-refactor-plan` | #167 | `dff487d` | **Keycloak-only refactor KC-PR-A..F 6단계 단일 묶음** — keycloak_verifier + admin client + migration 000021 (kratos_identity_id → idp_subject) + JWKS cache + resource_access fallback + frontend OIDC discovery + Hydra/Kratos 코드 대거 삭제. design 문서 PR #163 의 옵션 B 권장을 reversal, 옵션 A 채택. ADR governance 정정은 sprint -a (PR #169) 가 사후 처리. |
| `codex/keycloak-only-refactor-plan` (후속) | #168 | `09f4ca1` | **PR #167 머지 후 workflow memory sync** — main flat state/handoff/work_backlog 갱신 (외부 codex 작성) |



## 2026-05-18 post-EOD #2 후속 세션 9 PR (sprint r..w + reverse-proxy + keycloak-only + housekeeping)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-r` | #159 | `79c0ec3` | **Jira 보고서 self-review hotfix** — P1 3건 (ARCH 23→29 / TC 12 / M1 RBAC PR 범위) + P2 4건 (Mermaid Jira 호환 caveat / sprint code 안내 / PR #158 명시) |
| `-s` | #160 | `31cdad6` | **ADR-0016 §6 (1)+(2) resolved** — `docs/setup/prometheus_alertmanager_setup.md` (외부 git layout + stage/prod raw YAML + 라우팅 + 운영 SOP) + `docs/setup/grafana/homelab_dashboard.json` (5 panel + environment template) |
| `-t` | #161 | `8802a5b` | **ADR-0017 §6 (a)+(c)+(d) resolved** — `internal/devrequest/{metrics,intake_token_cron}.go` 신규 + `HardRevokeExpiredIntakeTokens` + `Count{ExpiringSoon,Stale}IntakeTokens` + audit `dev_request_intake_token.auto_revoked` + counter `devhub_intake_token_auto_revoked_total` + 5 cron unit test + 3 store integration sub-test |
| `-u` | #162 | `338c430` | **단일 포트 reverse proxy design 검토** — `docs/planning/single_port_reverse_proxy.md` (16 section, nginx 권장, /devhub prefix, ADR-0018 후보) |
| `-v` | #163 | `cdd73b0` | **Keycloak SSO federation design 검토** — `docs/planning/keycloak_sso_federation.md` (16 section, Kratos federation 권장, HRDB employee_id strict link, ADR-0019 후보, RM-M4-09 구체화) |
| `-w` | #164 | `4dd02ad` | **codex review hotfix #8** — P1 3건 (HTTP puller ErrUnexpectedEOF retry / PromQL env scoping / Hydra admin URL) + P2 6건 (scopeFilter / react-flow assertion / file puller trailing JSON / ADR-0006+0012 link / NoRecentSuccess aggregation) |
| `reverse-proxy` | #166 | `694e694` | **단일 외부 포트 역프록시(Nginx) 구성 및 basePath 구현** — Same-Origin CORS 무력화 + basePath /devhub 런타임 dynamic endpoints resolution + Codex P1 피드백 반영 완료 |
| `keycloak-only` | #167 | `dff487d` | **Keycloak OIDC 단일화 및 코드/문서/테스트 가이드 최종 정합화** — Kratos/Hydra 전면 제거 + OIDC Discovery 도입 + In-Memory JWKS 캐싱 및 client roles(resource_access) 매핑 완결 |
| `housekeeping` | #168 | `09f4ca1` | **PR #167 병합 후속 전역 메모리 갱신** — Keycloak 단일화 병합에 따른 global state/backlog 및 handoff 동기화 완결 |

## ADR carve out 진행 현황 (누적)

| ADR | §6 carve | 상태 | 출처 sprint |
| --- | --- | --- | --- |
| ADR-0015 | (1) size limit + streaming decode | ✅ resolved | -p PR #157 |
| ADR-0015 | (2) agent token rotation SOP | ✅ resolved | -p PR #157 |
| ADR-0015 | (3) dedicated worker binary | ⏳ carve (M4 진입 시) | — |
| ADR-0015 | (4) push/pull dedup | ⏳ carve (별도 ADR) | — |
| ADR-0016 | (1) Alertmanager 운영 가이드 | ✅ resolved | -s PR #160 |
| ADR-0016 | (2) Grafana dashboard JSON | ✅ resolved | -s PR #160 |
| ADR-0016 | (3) pull latency p95 alert | ⏳ carve (1주 baseline) | — |
| ADR-0016 | (4) push 경로 webhook 알림 | ⏳ carve (metric 도입 후) | — |
| ADR-0016 | (5) stage→prod 임계 확정 | ⏳ carve (1주 관찰) | — |
| ADR-0017 | §6 atomicity 실 구현 | ✅ resolved | -o PR #156 |
| ADR-0017 | (a) 자동 cron revoke | ✅ resolved | -t PR #161 |
| ADR-0017 | (b) PATCH expires_at | ⏳ carve (UI 영향) | — |
| ADR-0017 | (c) 만료 알림 metric | ✅ resolved | -t PR #161 |
| ADR-0017 | (d) staleness alert metric | ✅ resolved | -t PR #161 |
| **ADR-0018** | single port reverse proxy | ✅ Accepted + 실 구현 | docs/adr/0018-single-port-reverse-proxy-policy.md + PR #166 (gemini/reverse-proxy) |
| **ADR-0019** | Keycloak 단일화 (Hydra+Kratos 폐기) — ADR-0001 supersession | ✅ Accepted + 실 구현 + ADR governance 정정 | PR #167 (codex/keycloak-only-refactor-plan, KC-PR-A..F) + PR #169 (sprint -a, ADR-0019 신규 발행 + ADR-0001 정정) |
| ADR-0019 §5.3 | Keycloak realm/client/role 운영 SOP | ⏳ carve (별도 SOP 문서) | — |
| ADR-0019 §5.3 | JWKS rotation 운영 SOP | ⏳ carve (별도 SOP 문서) | — |
| ADR-0019 §5.3 | Keycloak ↔ HRDB sync (employee_id strict link) | ⏳ carve (Keycloak admin user attribute 매핑 SOP) | — |
| ADR-0019 §5.3 | Keycloak SSO logout chain (RP-initiated) | ⏳ carve | — |
| ADR-0019 §5.3 | MFA / failover / off-boarding 즉시성 | ⏳ carve (M4 진입 시 / 사내 정책) | — |
| ADR-0019 §5.3 | Keycloak `groups` claim → DevHub RBAC role 자동 매핑 | ⏳ carve (별도 ADR 후보) | — |
| ADR-0019 §5.4 | RM-M4-09 의미 재정의 — Keycloak identity broker / Gitea/AD 등 외부 SSO | ⏳ carve (후속 sprint) | — |

## 본 후속 세션 #2 핵심 학습 (다음 세션이 참조)

### 1. io.ErrUnexpectedEOF 의 다중 의미 (sprint -w codex hotfix #8 P1 #1)
`io.ErrUnexpectedEOF` 는 oversized payload (LimitReader cap 도달) + transient transport 실패 (upstream mid-response close) 둘 다 포함. byte counter (`*io.LimitedReader.N == 0`) 기반 명시 oversized 감지로 분리. 그 외는 retryable=true 유지.

### 2. PromQL single Prometheus + multi-env (sprint -w codex hotfix #8 P1 #2)
단일 Prometheus 가 stage + prod 둘 다 scrape 시 alert rule expr 에 `{environment="prod|stage"}` matcher 가 없으면 양쪽 metric 합쳐 평가 후 label 만 rewrite → cross-env contamination. 매처 명시 필수.

### 3. spec ts 의 always-mounted shell assertion (sprint -w codex hotfix #8 P2 #2)
`.react-flow` 같은 always-mounted container 만으로 검증 시 API fetch 실패도 false positive PASS. `page.waitForResponse` + body schema 검증 + 결과 기반 분기 패턴 정착 (sprint -m 의 page.request OIDC 회피 + 본 학습 합쳐 spec ts 작성 체크리스트).

### 4. planning 문서의 env 변수 안내 검증 (sprint -w codex hotfix #8 P1 #3)
design 문서가 안내한 env 변수 (`DEVHUB_HYDRA_ADMIN_URL` 등) 는 backend 코드 (`hydra_introspection.go`, `hydra_admin_client.go` 등) 의 실제 endpoint 사용처 (`/admin/oauth2/introspect`) 와 정합 확인 필수. design 문서 작성 시 grep + 실제 호출처 확인 SOP.

### 5. ADR §7 변경 이력 row 의 PR 번호 명시 (sprint -w + -x)
ADR §7 의 변경 이력 row 에 sprint code 뒤 PR 번호 추가 (`sprint X (PR #N)`) — 후속 codex hotfix 가 정확한 PR 참조 필요. sprint -w 가 ADR-0015/0016 §7 row 의 PR 번호 추가 정정.

## 다음 세션 directive (2026-05-20 housekeeping #3 종료 시점 재산정)

### 1순위 — 계정/사용자 관리 리팩토링 Phase 3 (실 구현) 진입

PR #199/#200 의 design 자산 (`docs/planning/account_user_management_redesign.md`) 따라 8건 carve 실 구현. 권장 순서:

1. **backend code 제거 + lazy auto-create** — accounts_admin handler 4 + KeycloakAdminClient manage-users 호출처 4 모두 제거 + `authenticateActor` lazy backfill (결정 A + 결정 D 의 `rbac_subject_roles` 폐기 자연 통합). 결정 C 의 event listener 확장 전에 우선 진입 가능.
2. **frontend cleanup** — `account.service.ts` 폐기 + `MemberTable.tsx` 정리 + `/admin/settings/users` UI 정리 (backend 제거 후 dead UI 정리).
3. **Keycloak event listener 확장** — sprint -u~-y 의 audit puller 에 USER:UPDATE / GROUP_MEMBERSHIP / USER:DELETE 매핑 추가 + DevHub `users` write (결정 C, hot path 정합).
4. **JWKS expiry case 확장** — sprint -r 의 kid mismatch fallback 패턴 정합 (결정 F).
5. **service account 권한 축소 + governance SOP** — Keycloak admin SOP 갱신 (manage-users 제거) + `keycloak_operations.md §8.5c` 신규 (9 운영 동작 책임 분리 표) + ADR-0020 draft 작성.

### 2순위 — ADR-0019 §5.3 design+carve 실 구현 잔여 (5 후보, 모두 사내 동반 carve)

- ✅ **resolved (sprint -u~-y, PR #189~#193)** — audit event listener 실 구현 Phase 2 종결
- ✅ **resolved (sprint -ad, PR #198 `bcca86a`)** — Kratos 잔재 residual cleanup

1. **group staging-prod 적용** — 가장 가벼움: Keycloak admin 1회 작업 (group 4 생성 + composite role assign).
2. **off-boarding Phase 1 운영 cron 실 deploy** — `scripts/hrdb_etl_sync.sh` (sprint -p PR #184 신규) 의 사내 운영 cron 배포 SOP + 1회 staging 검증.
3. **HA Phase 2** — 사내 SRE / 인프라팀 결정 동반. ADR-0021 후보. Keycloak HA active-active.
4. **e2e Kratos → Keycloak admin 실 코드 전환** — 사내 Keycloak e2e 환경 staging 진입 동반. sprint -m design 따름. PR #201 의 `scripts/ci-e2e-sync-check.sh` + Keycloak audience mapper 가 CI 단 일부 해소 가능성 검증 필요.
5. **audit event listener SPI push 전환** — polling latency 30s → < 1s. sprint -x §8.6.9 sub-carve.

### 3순위 — 기존 ADR carve out 잔여

6. ADR-0015 §6 (3)+(4) — dedicated worker binary / push-pull dedup
7. ADR-0016 §6 (3)+(4)+(5) — baseline 관찰 / push webhook metric / stage→prod 임계
8. ADR-0017 §6 (b) — PATCH expires_at + admin UI 편집 modal
9. ADR-0018 Phase 2 staging — 단일 포트 reverse proxy
10. ADR-0019 §5.4 RM-M4-09 후속 — Keycloak identity broker / Gitea/AD federation

### 4순위 — 다른 도메인

11. M4 RM-M4-XX 본격 진입 — WebSocket / AI Gardener gRPC / System Admin / Gitea Hourly Pull
12. External Integration 후속 강화 — React Flow group sub-node + WebSocket 실시간 + v2 node click action
13. Bindings UI 강화 — scope_id lookup combobox / Edit/Delete / pagination
14. historical infra/idp/ 정리 — Hydra/Kratos 가동 가이드 (deprecation banner 부착) 본문 정리 또는 archive 이전

---

(이하: 직전 post-EOD #1 (sprint q 종료) 시점의 누적 작업 기록은 historical reference 로 보존)

## 2026-05-18 post-EOD 후속 세션 6 PR (sprint k..p + q)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-k` | #152 | `4056ab4` | EOD 종합 housekeeping (단일 일자 11 PR sync) |
| `-l` | #153 | `f11a8c7` | **codex hotfix #7** — PR #149 merge_commit SHA d24a41e → bb83520 정정 |
| `-m` | #154 | `b5bc54b` | **External Integration bindings 관리 UI** — /admin/settings/integration-bindings + BindingsTable + CreateBindingModal + service/types (API-74 GET + API-75 POST) + admin-integration-bindings.spec.ts. **5 hotfix iteration** 으로 CI 그린화 (CSS locator → DOM attribute → SYNC gate → page.evaluate fetch → sessionStorage Bearer) |
| `-n` | #155 | `dc0ed1c` | **Infra topology v2 시각화** — /admin/topology-v2 + React Flow + nodes/services/edges + degraded banner + snapshot_at 메타 + node click modal. 1차 CI 그린 (sprint -m 학습 직접 효과) |
| `-o` | #156 | `eaa91a7` | **ADR-0017 §6 atomicity 실 구현** — UpdateDevRequestIntakeTokenIPs 단일 CTE + FOR UPDATE + (VALUES (1)) AS root(_) anchor + Happy/NotFound/Revoked/ConcurrentUpdateAndRevoke 4 sub-test |
| `-p` | #157 | `6648105` | **ADR-0015 §6 carve outs (1)+(2)** — size limit + streaming decode (file os.Stat + HTTP Content-Length + io.LimitReader) + DEVHUB_HOMELAB_PULL_MAX_BYTES env + 5 회귀 unit test + agent token rotation SOP |
| `-q` | (본) | TBD | post-EOD 종합 housekeeping + Jira 보고 status 문서 |

## 신규 ID 종합 (본 세션 후속)

| 종류 | ID | 출처 sprint |
| --- | --- | --- |
| TC | TC-INT-FRONTEND-BIND-{LIST,CREATE,RBAC}-01 (3건) | `-m` PR #154 |
| TC | TC-INT-HOMELAB-03 (carve out → active) + TC-INT-FRONTEND-TOPOLOGY-V2-{NAV,RBAC}-01 (2건) | `-n` PR #155 |
| env | `DEVHUB_HOMELAB_PULL_MAX_BYTES` (int64, default 0 unlimited, 운영 권장 5 MB) | `-p` PR #157 |
| 문서 | `docs/setup/homelab_agent_token_rotation.md` (운영 SOP) | `-p` PR #157 |
| 문서 | `docs/reports/jira_status_2026_05_18.md` (Jira 보고용 10 항목 status) | `-q` (본) |

## External Integration 도메인 1차 종합 closing

| 영역 | 상태 | 출처 |
| --- | --- | --- |
| Concept staged (REQ-FR-INT + ARCH-INT + API-69..78 + IMPL-int + TC-INT) | ✅ | 외부 PR #135 |
| Backend 1차 (HomeLab pull + Prometheus + integration_registry + infra_snapshots) | ✅ | 외부 PR #139 |
| ADR (pull strategy / alerts policy / intake token hardening) | ✅ ADR-0015/0016/0017 | sprint -c PR #143 |
| Provider frontend 진입점 (/admin/settings/integrations) | ✅ | sprint -g PR #148 |
| Provider E2E spec (5 TC) | ✅ | sprint -h PR #149 |
| API-80 DELETE + FK guard + Delete UI | ✅ | sprint -j PR #151 |
| **Bindings 관리 UI** (/admin/settings/integration-bindings) | ✅ | sprint -m PR #154 |
| **Topology v2 시각화** (/admin/topology-v2 + React Flow + degraded banner) | ✅ | sprint -n PR #155 |
| ADR-0015 §6 (1) size limit + streaming decode | ✅ resolved | sprint -p PR #157 |
| ADR-0015 §6 (2) agent token rotation SOP | ✅ resolved | sprint -p PR #157 |
| ADR-0015 §6 (3) dedicated worker binary | ⏳ carve (M4 진입 시) | — |
| ADR-0015 §6 (4) push/pull dedup | ⏳ carve (별도 ADR) | — |
| React Flow group sub-node (services as node children) | ⏳ carve | — |
| WebSocket 실시간 갱신 (infra.node.updated / infra.service.updated) | ⏳ carve | — |
| ADR-0017 §6 atomicity 실 구현 | ✅ resolved | sprint -o PR #156 |
| ADR-0017 §6 잔여 (cron revoke / PATCH expires_at / 만료/staleness alert) | ⏳ carve | — |
| ADR-0016 §6 carve outs (Alertmanager YAML / Grafana JSON / p95 alert / push 알림 / 임계 확정) | ⏳ carve | 다음 세션 directive 1 |

## 본 후속 세션 핵심 학습 (다음 세션이 참조)

### 1. page.request OIDC session propagation 의 CI flaky (sprint -m, 5 hotfix iteration)

직전 PR #145 (외부 codex docker hardening + Hydra prod mode + runtime-config + forwarded-header 신뢰 제거) 머지 후 `page.request.*` (GET/PATCH/DELETE) 가 CI 에서 일관 fail. 회피 패턴 정착:

- **DOM attribute 추출** — `tr[data-provider-id]` / `tr[data-token-id]` 등 row 에 attribute 추가, spec ts 가 `getAttribute()` 로 결정적 매핑. `page.request.get` 의존 제거.
- **mutation (PATCH/DELETE)** — `page.evaluate(async (id) => { const accessToken = sessionStorage.getItem("devhub_access_token"); return fetch(url, { method: "PATCH", headers: { ..., Authorization: \`Bearer ${accessToken}\` }, credentials: "include" }); }, id)` 패턴. apiClient 와 동일 인증 구조.
- **modal submit 우선** — POST 는 가능하면 modal form submit 으로 (apiClient 가 자동 Bearer 추가 → 안정적).

### 2. Playwright CSS list 안에 engine selector 못 둠 (sprint -m)

`page.locator("table, text=/regex/i")` 는 invalid CSS — `locator.or()` chain 으로 분리:
```ts
const el = page.locator("table").or(page.getByText(/regex/i)).first();
```

### 3. backend integration test 의 created_by FK (sprint -o)

`dev_request_intake_tokens.created_by` 는 `users(user_id)` FK. CI 환경 시드 user 는 migration 000004 의 u1/u2/u3 (Kratos 시드 alice/bob/charlie 와 별개). integration test 는 `CreatedBy: "u1"` 사용 패턴 (applications_integration_test.go 와 동일).

### 4. backend env 변수 default vs 권장 운영값 분리 (sprint -p)

ADR §4.3 환경 변수 표는 두 컬럼:
- **코드 default** — 보수적 (legacy 호환), 예: `MAX_BYTES=0` (unlimited)
- **권장 운영값** — ADR-0016 알림 임계가 가정하는 값, 예: `MAX_BYTES=5242880` (5 MB)

이전 codex hotfix #5 P2 (sprint -e) 가 발견한 default drift 패턴 정합.

### 5. CTE + FOR UPDATE + VALUES anchor 패턴 (sprint -o)

2 query (UPDATE + SELECT) 의 between-query race → 단일 CTE + FOR UPDATE row lock + `(VALUES (1)) AS root(_) LEFT JOIN locked LEFT JOIN upd` anchor. locked empty 여도 단일 row 반환 보장 (codex hotfix #6 P2 의 zero-rows 문제 정합).

## 다음 세션 directive (우선순위 순 — post-EOD 기준 재산정)

1. **ADR-0016 §6 carve outs** — (a) Alertmanager raw YAML 외부 git 이관, (b) Grafana JSON 모델, (c) pull latency p95 alert (baseline 1주 관찰 후), (d) push 경로 (API-73) 알림, (e) stage→prod 임계 확정.
2. **ADR-0017 §6 잔여 carve outs** — (a) 자동 cron revoke (expires_at 임박 정리), (b) PATCH expires_at 갱신 허용, (c) 토큰 만료 알림 metric (devhub_intake_token_expiring_soon), (d) last_used_at staleness alert.
3. **ADR-0015 §6 (3)+(4)** — dedicated worker binary (M4 진입 시) / push/pull dedup 정책 (별도 ADR).
4. **External Integration 후속 강화** — React Flow group sub-node + WebSocket 실시간 갱신 + 토폴로지 v2 의 node click action (v1 ServiceActionCommand 패턴 확장).
5. **M4 RM-M4-XX 본격 진입** — WebSocket 확장 (RM-M4-01..03), AI Gardener gRPC (RM-M4-04..05), Gitea Hourly Pull worker (RM-M4-06), System Admin 대시보드 (RM-M4-07).
6. **bindings UI 강화** — scope_id 의 application/project lookup combobox, Edit/Delete binding 액션, pagination.

---

(이하: 직전 EOD (sprint k 종료) 시점의 누적 작업 기록은 historical reference 로 보존)

## 2026-05-18 단일 세션 11 PR 누적 (sprint a..k + 외부 #145)

| Sprint | PR | sha | 핵심 |
| --- | --- | --- | --- |
| `-a` | #141 | `98fa4a7` | 메모리 갭 동기화 — 외부 8 PR (#133~#140) flat memory 흡수 + **DREQ P2 6/6 모두 해소 확인** |
| `-b` | #142 | `d1461bd` | traceability §2.2 (API-79 추가) + §3 + §5.1 + §6 — 외부 PR ID 매핑 사후 정합 |
| `-c` | #143 | `d2c0031` | **ADR-0015** (HomeLab pull strategy) + **ADR-0016** (Prometheus alerts) + **ADR-0017** (intake token hardening) + ADR-0014 §6 갱신 |
| `-d` | #144 | `d24a41e` | DREQ-E2E TC-DREQ-* 13건 정식 발급 + `dev-requests.spec.ts` 6 step + 신규 2 test (INTAKE-AUTH-NEG-01 + ADMIN-TOKEN-PATCH-01) + `test_cases_m5_dreq.md` 신규 |
| (외부) | #145 | `f2e18a2` | docker packaging hardening (IMAGE_REPO_PREFIX + local-db db-init + DEV_FALLBACK=0 + Hydra prod mode + system secret env + runtime-config forwarded-header 신뢰 제거) + codex P1/P2 해소 |
| `-e` | #146 | `52981b9` | **codex hotfix #5** — PR #142 (§2.2 API-78→79) + PR #143 (ADR-0015 default vs 권장값, ADR-0017 PATCH revoked guard + 3 회귀 unit test) P2 3건 |
| `-f` | #147 | `951c0c8` | PR #146 self-review carve out — ADR-0017 §6 atomicity (CTE + FOR UPDATE reference SQL) + test 회귀 시나리오 주석 |
| `-g` | #148 | `7b4ad18` | **External Integration frontend 진입점** — `/admin/settings/integrations` + ProviderTable + ProviderModal + integration.{service,types}.ts + layout subTabs Plug 아이콘 |
| `-h` | #149 | `bb83520` | External Integration E2E — `admin-integrations.spec.ts` (mega lifecycle 4 step + RBAC negative). TC-INT-FRONTEND-{LIST,CREATE,EDIT,SYNC,RBAC}-01 active |
| `-i` | #150 | `99c0833` | **codex hotfix #6** — PR #147+#148+#149 의 P1×2 (useEffect [toast] dep + syncProvider 응답 형식) + P2×2 (ADR-0017 CTE LEFT JOIN empty → not-found 분기 불가 + SYNC assertion false positive) |
| `-j` | #151 | `51c79af` | **API-80 DELETE integration provider + FK guard** (binding count > 0 → 409) + handler 4 unit test + DestructiveConfirmModal + spec ts DELETE step 5 |
| `-k` | (본) | TBD | 종합 housekeeping (state.json / session_handoff / work_backlog / auto-memory) |

## 신규 ID 종합 (본 세션)

| 종류 | ID | 출처 sprint |
| --- | --- | --- |
| ADR | ADR-0015 HomeLab adapter pull strategy | `-c` PR #143 |
| ADR | ADR-0016 Prometheus alerts policy | `-c` PR #143 |
| ADR | ADR-0017 DREQ intake token operational hardening | `-c` PR #143 |
| API | API-79 PATCH `/api/v1/dev-request-tokens/:token_id` (allowed_ips mutation, 이전 PR #137 활성화의 ID 정합) | `-b` PR #142 |
| API | API-80 DELETE `/api/v1/integration/providers/:provider_id` (FK guard) | `-j` PR #151 |
| TC | TC-DREQ-* 13건 (ADMIN-TOKEN-01/REVOKE-01/PATCH-01 + INTAKE-AUTH-01/NEG-01/03/EXPIRED-01/NEG-02 + WIDGET-FLOW-01 + PROMOTE-TX-01 + ADMIN-TOKEN-PATCH-NEG-01 + RBAC-NEG-01/02 + RBAC-ROW-01) | `-d` PR #144 |
| TC | TC-INT-FRONTEND-* 7건 (LIST/CREATE/EDIT/SYNC/RBAC/DELETE/DELETE-NEG-01) | `-g`+`-h`+`-j` |

## DREQ 도메인 종합 (closing 1차 + carve out 모두 해소)

| 영역 | 상태 | 출처 |
| --- | --- | --- |
| Concept / AuthADR | ✅ ADR-0012 옵션 A | sprint f/g (2026-05-15) |
| Backend 1차 (API-59..65) | ✅ activated | sprint i (2026-05-15) |
| Frontend 1차 (위젯 + 페이지) | ✅ activated | sprint j (2026-05-15) |
| Promote-Tx 단일 트랜잭션 (API-62) | ✅ activated + ADR-0013 | sprint m (2026-05-15) |
| Admin-UI backend (API-66..68) + ADR-0014 | ✅ activated | sprint o (2026-05-15) |
| Admin-UI frontend (페이지 + plain-1회 modal) | ✅ activated | sprint p (2026-05-15) |
| token expires_at + middleware 만료 체크 + API-79 PATCH allowed_ips | ✅ activated + ADR-0017 | 외부 PR #137 (gemini, 2026-05-16) + 본 세션 `-c` 사후 명문화 |
| Promote-Tx race 가드 (P2 #1) | ✅ `WHERE status IN ('pending','in_review')` | 외부 시기 미상, 본 세션 `-a` 에서 확인 |
| failPromote dead field 제거 (P2 #2) | ✅ grep 0건 | 외부 시기 미상 |
| DestructiveConfirmModal (P2 #3) | ✅ wire 완료 | 외부 PR #140 |
| plain_token Show/Hide toggle (P2 #4) | ✅ Eye/EyeOff | 외부 PR (시기 미상) |
| TC-DREQ-* 정식 발급 + spec ts active | ✅ 13건 | 본 세션 `-d` PR #144 |
| PATCH revoked guard (codex hotfix) | ✅ store guard + 회귀 가드 3 unit test | 본 세션 `-e` PR #146 |
| PATCH atomicity (race window 강화) | 🟡 ADR-0017 §6 carve out (CTE + FOR UPDATE reference SQL 만 명문화, 실 구현 미진입) | `-f`/`-i` ADR 본문 |

## External Integration 도메인 (frontend 진입점 + DELETE 까지 1차 완성)

| 영역 | 상태 | 출처 |
| --- | --- | --- |
| Concept staged (REQ-FR-INT + ARCH-INT + API-69..78 + IMPL-int + TC-INT) | ✅ | 외부 PR #135 (codex, 2026-05-15) |
| Backend 1차 (HomeLab pull adapter + Prometheus + integration_registry + infra_snapshots + API-73..78) | ✅ | 외부 PR #139 (codex, 2026-05-17) |
| ADR (pull strategy + alerts policy) | ✅ ADR-0015/0016 | 본 세션 `-c` PR #143 |
| Frontend 진입점 `/admin/settings/integrations` | ✅ activated | 본 세션 `-g` PR #148 |
| E2E spec ts (5 TC) | ✅ active | 본 세션 `-h` PR #149 |
| API-80 DELETE + FK guard + Delete UI | ✅ activated | 본 세션 `-j` PR #151 |
| Bindings 관리 UI (API-74, 75) | ⏳ 미진입 | 다음 세션 directive 1 |
| Infra topology v2 시각화 (API-76, 78, React Flow) | ⏳ 미진입 | 다음 세션 directive 2 |

## codex review cycle (본 세션 2회)

- **hotfix #5** (sprint `-e` PR #146) — PR #142 P2 (§2.2 API range) + PR #143 P2×2 (ADR-0015 default drift + ADR-0017 PATCH revoked guard 누락 → store 가드 + handler ErrConflict + 회귀 가드 3 unit test)
- **hotfix #6** (sprint `-i` PR #150) — PR #147 P2 (ADR-0017 CTE LEFT JOIN empty → not-found 분기 불가) + PR #148 P1×2 (useEffect [toast] dep 무한 재실행 + syncProvider 응답 형식 mismatch) + PR #149 P2 (SYNC assertion false positive)
- **clean** — PR #150 (hotfix #5 자체) + PR #151 (API-80) 모두 codex review 0건

## 다음 세션 directive (우선순위 순)

1. **Bindings 관리 UI** (API-74, 75) — provider ↔ application/project scope 매핑. 별도 페이지 `/admin/settings/integration-bindings` 또는 `/admin/settings/integrations` 의 sub-tab. backend 는 이미 activated, frontend 진입점만 필요. system_admin only RBAC.
2. **Infra topology v2 시각화** (API-76, 78) — React Flow 그래프. TC-INT-HOMELAB-03 활성화 후보. backend `/api/v1/infra/services` + `/infra/topology/v2` 가 hydrate 된 상태이므로 frontend 패턴은 기존 infra-topology.spec.ts + frontend/app/(dashboard)/infra/ 의 React Flow 패턴 따름.
3. **ADR-0017 §6 atomicity 실제 구현** (CTE refactor) — `store.UpdateDevRequestIntakeTokenIPs` 의 2 query 패턴을 단일 CTE + `FOR UPDATE` row lock 으로 atomic 보장. ADR §6 의 reference SQL 그대로 적용 + 회귀 race test (concurrent goroutine).
4. **ADR-0015 §6 carve outs** — (a) size limit + streaming decode (대용량 payload 보호), (b) agent token rotation SOP, (c) dedicated worker binary 평가 (M4 진입 시), (d) push/pull 동시 운영 시 `snapshot_at` + `source` tag dedup 정책.
5. **ADR-0016 §6 carve outs** — (a) Alertmanager raw YAML 외부 git 이관, (b) Grafana JSON 모델, (c) pull latency p95 alert (baseline 1주 관찰 후), (d) push 경로 (API-73) 알림, (e) stage → prod 임계 확정 1주 관찰 후.
6. **ADR-0017 §6 잔여 carve outs** — (a) 자동 cron revoke (expires_at 임박 정리), (b) PATCH 의 expires_at 갱신 허용, (c) 토큰 만료 알림 metric, (d) `last_used_at` staleness alert.

## 본 세션 학습 (다음 세션이 참조)

- **useEffect 의 hook 결과 dep 위험**: `useToast()` 가 매 render 새 callback → `[toast]` dep 이 effect 무한 재실행. 페이지 작성 시 hook 결과 dep 항목은 자동 의심 + ESLint disable 명시 패턴 (PR #150 hotfix #6 P1 #1).
- **service ↔ backend 응답 형식 verify**: typescript cast 만으로는 wire 검증 안 됨. service 작성 시 backend handler 의 `c.JSON(...)` 본문 grep 으로 실 응답 schema 확인 패턴 (PR #150 hotfix #6 P1 #2 — syncProvider 의 `{status, job_id}` vs provider envelope).
- **spec ts assertion 강도**: cell visibility 등 click 전후 변하지 않는 elementen 만 검증하면 false positive. `page.waitForResponse` + response body 검증 + optimistic update 결과 검증 3 layer 패턴 (PR #150 hotfix #6 P2 #4).
- **FK guard hard delete**: schema CASCADE 가 있더라도 운영 안전을 위해 handler 단에서 명시 차단. `integration_provider_has_bindings` 패턴은 다른 도메인의 DELETE 도입 시 reference (`-j` PR #151).
- **ADR reference SQL 의 row anchor 필요성**: CTE LEFT JOIN 만으로는 empty CTE 시 zero rows. `(VALUES (1)) AS root(_) LEFT JOIN ... LEFT JOIN ...` 패턴이 모든 분기 단일 row 보장 (PR #150 hotfix #6 P2 #3).

---

(이하: 직전 sprint 의 누적 작업 기록은 historical reference 로 보존)

## 2026-05-18 메모리 갭 동기화 세션 — 핵심 사실

직전 메모리 상태는 PR #132 (2026-05-15T07:22Z, sha `253063e`) 까지였고, 그 이후 외부 에이전트가 main 에 머지한 **8 PR (#133~#140)** 이 미반영. 본 세션이 일괄 흡수.

### A. P2 carve out 누적 6건 → 0건 모두 해소 (코드 검증 완료)

| # | 항목 | 해소 경로 | 검증 |
| --- | --- | --- | --- |
| 1 | promote-tx race 가드 | `WHERE id = $1 AND status IN ('pending', 'in_review')` | `dev_requests.go:266` + `dev_requests_promote.go:29` grep PASS |
| 2 | `memoryDevRequestStore.failPromote` dead field | 제거 | `grep -r failPromote backend-core` → 0 hit |
| 3 | window.confirm → DestructiveConfirmModal | 컴포넌트 신규 + dev-request-tokens 페이지 wire (PR #140) | `frontend/components/ui/DestructiveConfirmModal.tsx` 존재 + `grep window.confirm frontend` → 0 hit |
| 4 | plain_token Show/Hide toggle | Eye/EyeOff icon + aria-label | `IssueIntakeTokenModal.tsx:29` `showToken` state, line 253/257/259/261 |
| 5 | token rotation expires_at | migration 000027 + `auth_intake_token_expired` middleware (PR #137) | 마이그레이션 + 미들웨어 wire |
| 6 | allowed_ips mutation endpoint | PATCH /api/v1/dev-request-tokens/:token_id (PR #137) | endpoint 활성화 |

### B. DREQ carve out 3/4 (E2E) 부분 흡수

| PR | 기여 |
| --- | --- |
| #136 (codex/e2e-fix-20260515) | `frontend/tests/e2e/dev-requests.spec.ts` 신규 — 위젯→list→detail flow + Promote 조건부 + `signout.spec.ts` login form visibility 기반 검증 |
| #138 (gemini/main_review_260516) | DREQ E2E 전반 안정화 + 감사로그 selector 정정 + TC-USR-04 legacy advanced filters 호환 |
| #140 (gemini/ui_standardization_260517) | dev-requests page FilterBar 적용 (검색·상태 필터 표준화) + E2E selector 호환 |

**잔여**: intake auth Playwright spec (admin token issue → bearer header → POST 외부 의뢰 → assignee dashboard widget 노출 → promote → revoke 전체 흐름), TC-DREQ-* 정식 발급, traceability §3 row 갱신 (외부 에이전트 PR 들의 ID 매핑 미반영 가능성 확인).

### C. External Integration 도메인 신규 1차 완성

| 단계 | PR | 결과 |
| --- | --- | --- |
| Concept staged | #135 (codex/memory-next-step-20260515) | `docs/planning/external_system_integration_concept.md` + `external_integration_capability_matrix.md` + REQ-FR-INT + REQ-NFR-INT + UC-INT + ARCH-INT + API-69..78 + traceability §3 row + IMPL-int planned 분해 + TC-INT |
| Backend 1차 | #139 (codex/next-step-20260516) | `internal/integrations/adapters/{contract,homelab,homelab_file_puller,homelab_http_puller,homelab_pull_loop,metrics}.go` + `internal/store/integration_registry.go` + `infra_snapshots.go` + `httpapi/integration_registry.go` + `infra_integrations.go` + API-73..78 활성화 + Prometheus `/metrics` + migration 000028 (integration_registry) + 000029 (infra_service_snapshots) |
| Supporting docs | (PR #139 부수) | `docs/planning/homelab_adapter_pull_strategy.md` + `prometheus_homelab_alerts.md` + `docs/tests/test_cases_m4_integration.md` + `reports/report_20260516_m4_integration.md` |

### D. Frontend 대시보드 리브랜딩 + 현황 페이지 신설 (PR #138)

- 개발자/관리자 대시보드 → 업무 현황/품질 현황 리브랜딩
- Applications / Repositories / Projects 전용 현황 페이지 (`page.tsx` + `[id]/page.tsx` 6 신규)
- 공통 FilterBar 컴포넌트 (`frontend/components/ui/FilterBar.tsx`)
- 감사 로그 페이지 Glassmorphism 리디자인

### E. Docker deploy 패키지 안정화 (PR #133)

- `docker-compose.deploy.yml` + `local-db` profile (번들 DB vs 외부 DB 분리)
- `infra/idp/{hydra.deploy.yaml, kratos.deploy.yaml}` + `infra/nginx/devhub.deploy.conf`
- `/api/runtime-config` 도입 (배포 환경별 OIDC callback URL 정합)
- `.github/workflows/docker-image-publish.yml` 신규
- 메모: CLAUDE.md 의 docker = env-specific 정책상 본 자산이 git 추적되는 것은 `docs/setup/docker-packaging-deployment-guide.md` 의 가이드 + deploy compose 만이고, 실제 prod compose 는 사용자 로컬에 둠 (cf. [feedback_no_docker])

## 본 후속 세션 (2026-05-15 post-EOD) 누적 머지 — 13 PR (5 claude + 8 외부)

| PR | sha | sprint | 작업 |
| --- | --- | --- | --- |
| #128 | 1f9ec50 | claude/work_260515-m | DREQ-Promote-Tx 단일 트랜잭션 + ADR-0013 RBAC row-scoping |
| #129 | 5546a41 | claude/work_260515-n | codex review hotfix #4 — PR #128 의 P1 (CHECK 매핑) + P2 (SCM gate) + self-review P2 #1 (rejected_reason NULL) |
| #130 | 0bdf299 | claude/work_260515-o | DREQ-Admin-UI backend — intake token admin (API-66..68) + ADR-0014 |
| #131 | 2147d6d | claude/work_260515-p | DREQ-Admin-UI frontend — /admin/settings/dev-request-tokens 페이지 + plain-1회 modal |
| #132 | 253063e | claude/work_260515-q | post-EOD housekeeping (main flat memory sync + sprint finalize) |
| #133 | 4892a78 | codex/docker-packaging-guide | Docker deploy 패키지 안정화 + runtime-config + admin-permissions E2E 2건 안정화 |
| #134 | 09528bf | gemini/dreq_e2e_260515 | 대시보드 UI 안정화 (Header 통합, Recharts 실차트, 위젯 인터랙션) + LogoutOverlay |
| #135 | 62a2088 | codex/memory-next-step-20260515 | [Docs] External Integration 컨셉~설계 패키지 |
| #136 | c61274b | codex/e2e-fix-20260515 | dev-requests + signout E2E flow 안정화 (strict locator + unique title + UX 정합) |
| #137 | 72bf265 | gemini/dreq_e2e_260515 | DREQ token expires_at + PATCH allowed_ips — **P2 #5 + #6 해소** |
| #138 | c0134a1 | gemini/main_review_260516 | 대시보드 리브랜딩 + Applications/Repositories/Projects 현황 페이지 + FilterBar 공통화 + DREQ E2E 안정화 |
| #139 | e2a76fb | codex/next-step-20260516 | External Integration backend 1차 (HomeLab pull adapter + Prometheus + integration_registry + infra_snapshots + API-73..78 + migration 000028/000029) |
| #140 | caa80c7 | gemini/ui_standardization_260517 | FilterBar 표준화 종합 + dev-requests page 적용 + DestructiveConfirmModal — **P2 #3 해소** |

## 본 후속 세션 도입 핵심 (재참조 가능)

### 0. 대시보드 UI 안정화 및 성능 최적화 (sprint gemini, PR #134)

- **UI 표준화**: `DashboardHeader` 컴포넌트 통합으로 Developer/Manager 대시보드 헤더의 타이포그래피, 그라데이션, 애니메이션 정합성 확보.
- **데이터 시각화**: `recharts` 라이브러리 도입 및 Manager 대시보드의 "Resource Utilization" 플레이스홀더를 실제 `AreaChart` (Mock 시계열 데이터 연동)로 교체.
- **인터랙션 강화**: "Talent Load Balancing" 위젯에 호버 효과 및 이동 인디케이터 추가로 관리 유용성 증대.
- **성능 최적화**: OIDC 로그아웃 리다이렉트 과정의 시각적 플리커 해결을 위해 `LogoutOverlay` 및 `isLoggingOut` 상태 관리 도입.
- **Codex 리뷰 반영**:
  - `isLoggingOut` 상태를 Zustand `persist` 대상에서 제외(partialize)하여 비정상 종료 시 화면 갇힘 방지.
  - E2E 테스트(`dev-requests.spec.ts`)에서 마스킹된 토큰을 읽기 전 `show token` 버튼을 클릭하도록 로직 수정.

### 1. DREQ carve out 1/4 — RBAC-ADR + Promote-Tx (sprint m, PR #128)

- **ADR-0013** — `dev_requests` resource 의 RBAC row-scoping 정책 사후 명문화. ADR-0011 §4.2 helper `enforceRowOwnership(c, dr.AssigneeUserID, "pmo_manager")` 의 dev_requests 적용 사례. handler wire-up 은 PR #124 (sprint i) 에서 이미 도입.
- **Promote-Tx**: `store.RegisterDevRequestWithNewApplication` + `RegisterDevRequestWithNewProject` 신규 — `pool.BeginTx` → INSERT target (+ optional `application_repositories`) → UPDATE dev_requests → Commit. **REQ-FR-DREQ-005 정합 완성**.
- handler 분기: `target_id` (legacy) / `application_payload` / `project_payload` mutual exclusion.
- SQL drift 방지: `applications.go` 의 INSERT SQL 들을 const 로 추출.

### 2. codex hotfix #4 (sprint n, PR #129)

- **P1**: primary_repo 분기의 `application_repositories_role_check` CHECK 위반이 500 으로 surface → handler `validApplicationRepoRoles` gate + store `isCheckViolation` defense-in-depth.
- **P2**: promote primary_repo path 의 SCM provider enablement gate 우회 → handler `ListSCMProviders` + Enabled 검증.
- **self-review P2 #1**: `MarkDevRequestRegistered` + `dreqMarkRegisteredUpdateQuery` 에 `rejected_reason = NULL` clear 추가.
- 5 회귀 가드 test.

### 3. DREQ carve out 2/4 Admin-UI backend (sprint o, PR #130)

- **ADR-0014** — dev_request_intake_tokens resource RBAC + plain-1회-노출 + idempotent revoke 정책. accounts_admin temp_password 패턴과 정합.
- **신규 RBAC resource** `dev_request_intake_tokens` (system_admin 일임). migration 000026.
- **신규 endpoint 3**: API-66 POST (발급) / API-67 GET (목록) / API-68 DELETE (revoke). server 32-byte base64url plain 생성 → SHA-256(hex) 저장 + audit 에 plain/hashed 둘 다 미포함 + revoke `COALESCE(revoked_at, NOW())`.
- 8 신규 unit test.

### 4. DREQ carve out 2/4 Admin-UI frontend (sprint p, PR #131)

- **/admin/settings/dev-request-tokens** 페이지 (system_admin 보호 via AuthGuard + `/admin/*` prefix + layout subTabs 에 Intake Tokens 추가).
- **IssueIntakeTokenModal**: 2 phase (form → reveal). reveal phase 의 outside-click + ESC **차단** — 실수로 plain token 분실 방지. clipboard API copy.
- **IntakeTokenTable**: client/source + allowed_ips chips + Active/Revoked badge + revoke action.
- `dev_request_token.{service,types}.ts` — thin wrapper.
- npm run build PASS (26 static pages) + vitest 41 tests PASS.

## 다음 세션 directive (2026-05-18 sync 기준 재산정)

P2 carve out 6/6 모두 해소 + DREQ E2E 부분 흡수 + External Integration 1차 backend activated 상태에서 진입 후보:

| 우선순위 | 작업 | scope |
| --- | --- | --- |
| 1 | **traceability 동기화** | 외부 에이전트 PR #133~#140 의 ID 매핑(REQ-FR-INT / ARCH-INT / API-69..78 / IMPL-int / TC-INT / TC-DREQ-*) 이 `docs/traceability/report.md` §3 매트릭스 행에 반영됐는지 검수. PR #135 본문에 staged 가 있지만 후속 PR (#139) backend 활성화 후 status 갱신 필요. PR template 의 추적성 영향 섹션 누락 가능성. |
| 2 | **DREQ-E2E 잔여** | intake auth Playwright spec — admin token issue → bearer header → POST 외부 의뢰 → assignee dashboard widget 노출 → promote (신규 application/project 단일 tx) → revoke. TC-DREQ-* 정식 발급. `dev-requests.spec.ts` (PR #136) 가 widget→list→detail flow 만 커버, intake 경로 미커버 추정. |
| 3 | **External Integration design/impl carve out** | (a) Frontend 진입점 (system_admin 페이지 또는 admin/settings 통합), (b) API-69..72 (verifier strategy 외 endpoint) wire 검증 grep, (c) ADR 발급 (HomeLab adapter pull strategy + Prometheus alerts 정책 — `docs/planning/homelab_adapter_pull_strategy.md` + `prometheus_homelab_alerts.md` 의 결정 사항 ADR 화 후보). |
| 4 | **ADR-0014 §6 본문 검수** | PR #137 이 expires_at + allowed_ips PATCH 를 활성화했지만 ADR-0014 §6 의 carve out 항목이 "구현됨" 으로 갱신됐는지 외부 에이전트가 처리했는지 미확인. |
| 5 | **기존 m3/m4 carve out + Application 도메인 frontend** | 이전 session_handoff §3 후보 (M4 RM-M4-XX, Application 도메인 frontend UI, critical_warning_count 임계치 외부화 등) — 우선순위 낮음, 신규 작업 1~3 처리 후 재평가 권장. |

**메모리 메타 정합**: 외부 에이전트 sprint 디렉터리는 `ai-workflow/memory/{codex,gemini}/<sprint>/` 에 각각 존재 — 본 세션이 main flat memory 만 갱신했고, 외부 sprint 디렉터리는 그대로 유지 (해당 에이전트가 next session 에 자체 정리).

---

(이하: 직전 sprint q EOD 시점의 누적 작업 기록은 historical reference 로 보존)

## 다음 세션 directive (사용자 지시)

> "다음 작업 사항 모두 묶어서 진행할 거야" — DREQ carve out 4건을 한 sprint plan 으로 진입.

| Carve | 의존 | scope |
| --- | --- | --- |
| **DREQ-RBAC-ADR** | (독립) | pmo_manager 위양 정책 ADR — ADR-0011 §4.2 패턴 따라 dev_requests resource 의 row-level 위양 명문화 |
| **DREQ-Promote-Tx** | backend (Promote-Tx ↔ Admin-UI 둘은 backend 가 의존하지 않음) | store.RegisterDevRequest 가 신규 application/project 생성 + dev_request 상태 갱신 단일 트랜잭션 — REQ-FR-DREQ-005 정합 완성. 현재 handler 는 기존 target_id 매핑만 |
| **DREQ-Admin-UI** | backend admin endpoint 신설 → frontend UI | intake token 발급/revoke endpoint (`POST /api/v1/dev-request-tokens` 등) + `/admin/settings/dev-request-tokens` 페이지. accounts_admin 의 password issuance 패턴 (plain 1회 노출) 따름 |
| **DREQ-E2E** | 다른 3건 완료 후 | Playwright spec (intake → dashboard widget → register → close 흐름) + Vitest unit. TC-DREQ-* 발급. |

권장 진입 순서: **RBAC-ADR + Promote-Tx 병행 → Admin-UI → E2E**. 한 sprint 묶음 또는 4개 PR 로 나누는 선택은 진입 시점에 결정.

## 본 세션 (2026-05-15) 도입 핵심 (재참조 가능)

### 1. 도메인 / 인프라 4 패턴
1. `frontend/lib/config/endpoints.ts` — 모든 서비스 URL default 단일 진실 소스 (native default + env override)
2. `app/layout.tsx` inline script — theme FOUC 방지
3. `next.config.ts output: standalone` — NEXT_OUTPUT env gate
4. ADR-0011 §4.2 `enforceRowOwnership` + audit `auth.row_denied` + pmo_manager seed migration 000021

### 2. DREQ 도메인 4 결정 (sprint f~k)
- **컨셉** (sprint f, PR #121) — `docs/planning/development_request_concept.md` + REQ-FR-DREQ-001..011 + UC-DREQ-01..10 + ARCH-DREQ-01..06 + API-59..65 spec
- **AuthADR** (sprint g, PR #122) — ADR-0012 옵션 A (API 토큰 + IP allowlist) + `dev_request_intake_tokens` 테이블 스펙
- **Backend** (sprint i, PR #124) — 7 endpoint activated + `requireIntakeToken` middleware + 19 신규 unit test + migration 000022/023/024
- **Frontend** (sprint j, PR #125) — `/admin/settings/dev-requests` (system_admin 전체관리) + `/dev-requests` (일반 사용자, codex #125 hotfix) + DevRequestTable / DevRequestDetailModal / MyPendingDevRequestsWidget + developer/manager dashboard 통합

### 3. codex review cycle 3회
- hotfix #1 (sprint d, PR #119) — PR #114/#118 의 P1×3 + P2×1
- hotfix #2 (sprint h, PR #123) — PR #119/#120/#121 의 P1×2 + P2×2 (migration 000021 down FK + API-65 close 권한 + REQ source_system + sprint memory finalize)
- hotfix #3 (sprint k, PR #126) — PR #122/#124/#125 의 P1×2 + P2×2 (assignee FK rejected row + limit/offset + 일반 사용자 페이지 + session_handoff header)

## 본 세션 (2026-05-15) 누적 머지 — 15 PR

| PR | sha | sprint | 작업 |
| --- | --- | --- | --- |
| #112 | 3f387cd | codex/frontend_color_review | (흡수) Admin UI + ActionMenu + iPad |
| #115 | b669bc7 | gemini/frontend_redesign_260514 | Light theme + dropdown + endpoints 통일 |
| #114 | 25f97ba | codex/260514-a | Application leader/dev_unit + search 확장 |
| #116 | cbc36b0 | claude/work_260515-a | sprint a housekeeping |
| #117 | 68f031e | claude/work_260515-b | 모달 token 정책 sweep |
| #118 | 519a508 | claude/work_260515-c | enforceRowOwnership helper + ADR-0011 §4.2 |
| #119 | bca612e | claude/work_260515-d | codex hotfix #1 |
| #120 | feac299 | claude/work_260515-e | sprint e housekeeping |
| #121 | 52f6ad8 | claude/work_260515-f | DREQ 도메인 컨셉~설계 staged |
| #122 | 4d0277f | claude/work_260515-g | ADR-0012 DREQ AuthADR |
| #123 | 1d24acf | claude/work_260515-h | codex hotfix #2 |
| #124 | 333edc9 | claude/work_260515-i | DREQ Backend 1차 |
| #125 | 58033d2 | claude/work_260515-j | DREQ Frontend 1차 |
| #126 | bb164c4 | claude/work_260515-k | codex hotfix #3 |

## 0. 2026-05-15 머지 흐름 (7 PR)

```
bca612e PR #119 — fix(frontend,backend,migrations): codex review hotfix — PR #114 + PR #118 (claude/work_260515-d)
519a508 PR #118 — feat(httpapi,adr): enforceRowOwnership helper — ADR-0011 §4.2 + REQ-FR-PROJ-009 활성화 (claude/work_260515-c, 본인 4단계 리뷰)
68f031e PR #117 — refactor(frontend): PR #114 신규 컴포넌트 token 정책 일관 sweep (claude/work_260515-b, 본인 4단계 리뷰)
cbc36b0 PR #116 — docs(memory): 2026-05-15 sprint claude/work_260515-a housekeeping (claude/work_260515-a)
25f97ba PR #114 — feat(application): leader/dev_unit 모델 + search 확장 + auth_login canonical (codex/260514-a, 본인 4단계 리뷰)
b669bc7 PR #115 — feat(frontend): Light theme + dropdown refactor + endpoints 통일 (gemini/frontend_redesign_260514, 본인 4단계 리뷰)
3f387cd PR #112 — feat(frontend/org): Admin UI 개선 + iPad 터치 + 백엔드 트랜잭션 (codex/frontend_color_review, 2026-05-14 머지 누락 흡수)
```

## 1. 본 세션 도입 핵심 4 패턴 (재사용 가능)

### 1.1 endpoints 통일 모듈 (`frontend/lib/config/endpoints.ts`)
모든 서비스 URL default 단일 진실 소스. 정책: 코드 default = native(localhost), docker 는 env override (CLAUDE.md "native default, docker optional" + 메모리 [docker env-specific]). 사용처 8개 service 일괄 갱신 (`next.config`, `infra`, `rbac`, `realtime`, `websocket`, `auth`, `kratos-logout`). `frontend/.env.example` + `.gitignore` 의 .env.example 예외 nested 까지 확장.

### 1.2 theme FOUC 방지 (`app/layout.tsx` inline script)
paint 전 `<script>` 가 localStorage 의 `devhub-theme` 을 읽어 `theme-dark` class 적용. `Header.tsx` 의 useState 는 `document.documentElement.classList.contains` 로 lazy initialize. dead code `ThemeToggle.tsx` 제거 (Header dropdown 단일 진입점).

### 1.3 standalone gate (`next.config.ts`)
`output: "standalone"` 이 `next start` 와 호환되지 않아 CI/native dev 의 e2e webServer 가 깨졌던 문제. `NEXT_OUTPUT === "standalone"` env 일 때만 활성화 — docker 빌드 단계에서만 켠다.

### 1.4 ADR-0011 row-level 위양 (`enforceRowOwnership` + audit `auth.row_denied`)
- 시그니처: `func (h Handler) enforceRowOwnership(c *gin.Context, ownerUserID string, allowedRoles ...string) bool`
- allow 규칙: (1) `actor.role == system_admin`, (2) `actor.role ∈ allowedRoles`, (3) `actor.login == ownerUserID` (owner-self; ownerUserID == "" 자동 비활성화)
- deny: audit `auth.row_denied` + payload `{actor_role, owner_user_id, resource, action, denied_reason="owner_mismatch"}` + 403 `code=auth_row_denied`
- 운영 wire-up: `updateApplication / archiveApplication / updateProject / archiveProject` 4개 handler (PR #119)
- pmo_manager RBAC seed migration `000021_rbac_pmo_manager` — REQ-FR-PROJ-010 정책 매핑 (applications view+edit only, projects 전체, application_repositories view only)
- `devFallbackEnabled` 환경(test) 에서는 bypass — `enforceRoutePermission` 과 일관

## 2. 본 세션 리뷰어 모드 / codex review cycle 학습

### 2.1 본인 PR 4단계 리뷰 4회 (#114/#115/#117/#118)
diff 재검토 → gh pr comment (P0/P1/P2 분류) → 보강 commit (필요 시) → squash merge. **PR #115 의 E2E 가 2회 실패** 한 사례에서 **서비스 부팅 워닝 (Layer 1) + 실패 테스트 이름 (Layer 2)** 두 layer 분리 분석 패턴 정착:
- artifact `frontend.log` 의 "next start does not work with output standalone" → standalone gate fix
- raw job log 의 `header-switch-view.spec.ts:8/22/46/55` → Switch View 의도적 제거에 따른 e2e 정리

### 2.2 codex 외부 리뷰 hotfix 1회 (#119)
머지 후 codex 가 4 PR 에 inline 리뷰 (P1×3 + P2×1). PR #119 에서 일괄 처리:
- PR #114 P1: edit payload key 제외 (immutable 검증 회귀)
- PR #114 P2: date-only `new Date()` → `parseISO` (timezone shift 회귀)
- PR #115 P1: 이미 해소 (보강 commit `f621189`) — no action
- PR #118 P1: helper dead code → pmo_manager seed + handler wiring

본 sprint 가 정착한 패턴: **scope 와 effective production 효과의 불일치는 정당한 codex 지적**. ADR 의 carve out 이라도 PR body 가 "활성화" 라고 표기하면 effective 효과까지 가야 일관.

## 3. 다음 세션 진입 후보 (우선순위 순)

1. **owner-self route gate 활성화** — 현재 route-level RBAC gate 가 `applications:edit / projects:edit` 을 `system_admin/pmo_manager` 만 통과시키므로 owner-self 가 route gate 에서 막힘. owner-self effective 활성화는 route gate 정책 변경 (예: bypass owner-self for own row) + ADR 갱신 동반. ADR-0011 §4.2 의 자연 후속.
2. **critical_warning_count 임계치 외부화** — concept §13.2.1 운영 정책 테이블 신설 + handler lookup 동적화.
3. **CountApplicationCriticalWarnings lightweight SQL** — 현재는 전체 rollup compute 재호출, 성능 분리.
4. **Repository commit activity ingest** — pr_activities 외 commit 단위 이벤트.
5. **Project active→closed 가드 정책** — Application 만 critical 가드, Project 는 단순 transition.
6. **pr_activities.payload sanitization** — system_admin 외 노출 방어.
7. **quality_snapshots idempotency** — UNIQUE 추가 vs history retention.
8. **M4 RM-M4-XX 본격 진입** — WebSocket 확장 / replay / System Admin 대시보드 등.
9. **traceability follow-up** — TC-NAV-01/02/SIM-01 row 정리 (PR #115), endpoints 모듈 도입 ARCH 정책 1줄 (PR #115), pmo_manager seed traceability row (PR #119).

## 0. 2026-05-14 머지 흐름 (PR #104~#110, 7건)

```
d29f2ac PR #110 — fix(test,ci,store): PR #109 codex review hotfix — B1 fixture + I1 setup + I3 backend-integration job + 2 SQL fix + GITHUB_STEP_SUMMARY (claude/work_260514-f)
1e38c4d PR #109 — test(store): postgres integration test 도입 + P1/P2 회귀 guard 23 test (claude/work_260514-e)
7822a91 PR #108 — fix(application,store,httpapi): PR #107 codex review hotfix — P1 custom weight normalize + P2 UpdateIntegration unique mapping (claude/work_260514-d)
f11bdbb PR #107 — feat(application,store,httpapi,docs): API-51..58 세트 활성화 + active→closed critical 가드 흡수 (claude/work_260514-c)
66ab5ff PR #106 — feat(application,store,httpapi,docs): Application Design 2차 — A1+A2+A3 (API-41~50 activated) (claude/work_260514-b)
642d976 PR #105 — feat(application,adr,docs): Application Design 1차 + ADR-0011 결정 (claude/work_260514-a)
63a7ea2 PR #104 — docs: Application/Repository/Project 설계 고도화 + SCM 어댑터 모델 + self-review 13건 보강 (codex/project_concept_design)
```

## 0.1 Application 도메인 1차 완성 인벤토리

| 항목 | 결과 |
| --- | --- |
| API | API-01~58 전체 activated. 본 세션이 API-41..58 (18 endpoint group) 모두 활성화. |
| 마이그레이션 | 000012~000018 (7건) — scm_providers + applications + application_repositories + projects + project_members + project_integrations + pr_activities + build_runs + quality_snapshots + RBAC seed |
| ADR | ADR-0011 RBAC row-scoping accepted (옵션 C 1차 + 옵션 B 단계적 확장 옵션) |
| RBAC | 4 신규 resource (`applications` / `application_repositories` / `projects` / `scm_providers`) — system_admin only |
| Frontend | `PermissionMatrix` 9 resource 확장 (PR #105 self-review B1) |
| Domain types | Application / ApplicationRepository / SCMProvider / Project / ProjectMember / ProjectIntegration / PRActivity / BuildRun / QualitySnapshot / RepositoryActivity + WeightPolicy / Rollup* |
| Store interface | 27 메서드 (ApplicationStore) |
| Handler | 14 endpoint with 상태 전이 머신 + 가드 + audit emit (과거형 시제) |
| Unit test | 43 application 관련 (handler 25 + project 8 + integration 6 + rollup 4) |
| Integration test | 25 (Applications 9 + Repository ops 8 + Projects+Integrations 8 — DEVHUB_TEST_DB_URL 환경) |
| CI job | 4 → 5 (Workflow Lint / Backend Unit / Frontend Unit / E2E / **Backend Integration 신설**) |
| 본인 리뷰 | 4단계 × 5회 (#104, #105, #106, #107, #109/#110) — 모두 diff 재검토 → 코멘트 → 보강 → 머지 |
| codex 외부 리뷰 흡수 | 2회 (#107 → #108 hotfix / #109 → #110 hotfix) |

## 0.2 codex review cycle 학습 (다음 세션 적용)

머지 후 codex 외부 리뷰가 inline P1/P2 로 도착하는 시나리오가 본 세션에 2회 발생:
- PR #107 → P1 custom weight normalize 미실행 + P2 UpdateIntegration unique 매핑 누락 → PR #108 hotfix
- PR #109 → P1 fixture cleanup SQL multi-statement + bind args → PR #110 hotfix

본 세션이 정착한 패턴:
1. 머지 후 codex 외부 리뷰 확인
2. inline P1/P2 발견 시 hotfix PR 진입 (별도 브랜치)
3. 정정 + 회귀 guard test 추가 (integration test 가 결정적)
4. self-review 4단계 일관 유지

## 1. 다음 세션 진입 후보 (우선순위 순)

1. **frontend `/admin/settings/applications` UI** — IA 설계 + 화면 흐름. PR #105 self-review B1 의 PermissionMatrix 확장만 했고 페이지 자체는 미생성.
2. **ADR-0011 §4.2 enforceRowOwnership helper** — Owner 위양 2차 단계 (pmo_manager 활성화 sprint). 시그니처는 ADR-0011 §6 에 명시.
3. **critical_warning_count 임계치 외부화** — concept §13.2.1 운영 정책 테이블 신설 + handler lookup.
4. **CountApplicationCriticalWarnings lightweight SQL** — 성능 (현재는 전체 rollup compute 재호출).
5. **Repository commit activity ingest** — pr_activities 외 commit 단위 이벤트.
6. **Project active→closed 가드 정책 결정** — Application 만 critical 가드, Project 는 단순 transition.
7. **pr_activities.payload sanitization** — system_admin 외 노출 방어.
8. **quality_snapshots idempotency 결정** — UNIQUE 추가 vs history retention.
9. **M4 RM-M4-XX 본격 진입** — WebSocket 확장 / replay / System Admin 대시보드 등.
10. **세션 진입 시점에 사용자 환경 ops verification** — `scripts/setup-test-db.sh` 1회 실행 권장 (integration test 가 CI 외에 로컬에서도 정합).

## 0. 2026-05-13 머지 흐름

```
244f6b1 PR #102 — docs(planning): Project 도메인 컨셉 1차 — CRUD + 등록 + 조회 MVP scope (claude/work_260513-p)
118899b PR #101 — docs(memory): 2026-05-13 세션 종료 housekeeping sync (claude/work_260513-o)
4134b37 PR #100 — feat(domain,migrations),docs(adr,traceability),scripts: M4 전 잔여 일괄 (claude/work_260513-n)
c7c2f35 PR #99 — feat(hrdb,org,signup),docs(adr,traceability),test: M3 후속 1-4 일괄 (claude/work_260513-m)
b1268ce PR #98 — feat(auth),docs(adr,api,traceability),test: M3 진입 1차 — RM-M3-01..03 (claude/work_260513-l)
3d7d5a2 PR #97 — docs(roadmap,traceability): M3/M4 drift 정합 (claude/work_260513-k)
f551e6a PR #96 — feat(auth),docs(adr,api,traceability),test: M3 진입 전 잔여 후속 일괄 (claude/work_260513-j)
cb9e6d5 PR #95 — docs(traceability,adr,ci),test(frontend): 대형 묶음 B1~D5 (claude/work_260513-i)
ceb0f6f PR #94 — docs(adr): ADR-0004 X-Devhub-Actor 폐기 완료 선언 (B4) (claude/work_260513-h)
594be74 PR #93 — docs(traceability,api): B1 auth 도메인 2차 — §11 본문 ID 노출 + IMPL-auth-01..07 책임 정의 (claude/work_260513-g)
a73dba1 PR #92 — docs(traceability,api): B 묶음 — RBAC API §12 IMPL 정밀 매핑 + 본문 ID 노출 (claude/work_260513-f)
ae8b459 PR #91 — feat(backend): A 묶음 — M1 PR-D 정합 마무리 (writeRBACServerError + writeAuthLoginServerError 통합 + X-Request-ID validation + ctx 표준 request_id 전파) (claude/work_260513-e)
ea8df91 PR #90 — docs(governance,traceability): 갭 분석 정리 + 메타 헤더 표준화 + main flat sync (claude/work_260513-d)
7fac5bf PR #89 — docs: governance + traceability 체계 도입 + 1차 종합 매트릭스 (claude/work_260513-c)
9268227 PR #88 — docs(adr): ADR-0003 no-docker policy CI scope — drop services: postgres + native PG 15 (claude/work_260513-b)
e86f38f PR #87 — ci: FU-CI-2/3/4 (playwright scope, GHA cache, frontend readiness) + main flat memory sync (claude/work_260513-a)
450cc24 PR #86 — ci: E2E 테스트 안정화 및 GitHub Actions CI 파이프라인 구축 (gemini/prepare-github-action)
```

세부:
- **PR #86** — GitHub Actions CI 1차 도입. backend-unit + frontend-unit + e2e 3잡. 리뷰어 모드 2-pass 로 5 blocker + 5 follow-on fix.
- **PR #87** — FU-CI-2/3/4 처리. playwright install scope (chromium 단일), `actions/cache@v4`, frontend readiness 120s, install split (cache-hit 시 browser install skip).
- **PR #88** — ADR-0003 (no-docker 정책 CI 범위 명문화). `services: postgres:15` sidecar 제거 + pgdg native PG 15 native 설치 step. PG-14 dropcluster + `--port=5432` 강제.
- **PR #89** — `docs/governance/` (README + document-standards.md) + `docs/traceability/` (README + conventions + sync-checklist + report) + PR template + AGENTS/GEMINI 진입점. 1차 종합 매트릭스: REQ-FR 105 + REQ-NFR 26 + ARCH 17 + API 40 + RM 28 + IMPL 79 + UT 47 + TC 37 = 412 항목.
- **PR #90** — 매트릭스 §5 갭 표 형식 통일 + §5.2 auth.spec.ts TC 미흡수 + §5.3 ADR-0001 vs §3.8 모두 closed. document-standards §2 메타 헤더 9 문서 (누락 4 + 변형 4 + 부분 1) 일괄 적용. PR #87/#88/#89 누적 main flat memory sync.
- **PR #91** — A 묶음 (M1 PR-D 정합 마무리). `writeRBACServerError` (11 호출) + `writeAuthLoginServerError` (5 호출, Pass 1 review 발견) → `writeServerError` 일괄 통합. `requireRequestID` 미들웨어에 `validateCallerRequestID` (정규식 + 길이 + control char) 추가. request_id 를 `requestIDCtxKey{}` typed ctx key 에도 stash + `requestIDFromContext`/`logRequestCtx` ctx-aware helper. kratos_login_client.go 2건 + kratos_identity_resolver.go 1건 ctx-aware 치환. 단위테스트 11건 추가.
- **PR #92** — B 묶음 RBAC 1차. `backend_api_contract.md` §12.2~§12.10 의 9 헤더에 `(API-26..31, 38..40)` 본문 ID 노출. 매트릭스 §2.2 RBAC API + §2.4 IMPL-rbac-01..04 책임 정의 (handler / store / enforcement / cache) 서브 표 도입. §5.2 RBAC IMPL 정밀 매핑 항목 closed. Pass 1 review 보강으로 §3 RBAC 행을 ID 범위 + §2 서브 표 참조 패턴으로 정리 ("표 가독성 정책" 명문화).
- **PR #93** — B1 auth 도메인 2차. `backend_api_contract.md` §11.3 `(API-19)` + §11.5 표에 API ID 컬럼 (`API-20..24, 35`) + §11.5.1 `(API-35)` 본문 ID 노출. 매트릭스 §2.2 Auth API + §2.4 IMPL-auth-01..07 책임 정의 (verifier / actor / 5 endpoint handler) 서브 표 도입. §3 인증/회원가입/계정 관리 행 정리 (cross-cut API-23 / API-35 명시).
- **PR #101** — 2026-05-13 세션 종료 housekeeping sync 1차. main flat memory 의 PR #100 흡수 + 다음 세션 진입 후보 명문화 (RM-M4-01..09 + M3 carve out 6 항목).
- **PR #102** — Project 도메인 컨셉 1차. `docs/planning/project_management_concept.md` 신규 (10 절: 도메인 정의 / 행위자 × usecase / MVP scope / 데이터 모델 초안 / 다른 도메인 연계 / UI 컨셉 / 미해결 항목 / 후속 4 sprint hook). 일반 사용자 = 조회 중심, 시스템 관리자 = 등록·관리 전용 분리. RBAC row-scoping (ADR-0011 후보) 는 Design sprint 보류. `docs/planning/README.md §5.1` 도메인 컨셉 인덱스 신설 + `docs/development_roadmap.md §5` 백로그 1행 추가. 추적성 ID 미발급 (컨셉 단계).

## 1. 세션 종료 — 다음 세션 진입 후보

본 세션 (2026-05-13) 종료. 다음 세션 진입자는 본 §1 + §2 를 기준으로 작업 선택.

### 1.1 RM-M4 진입 후보 (매트릭스 §2.3.2 참조)

| RM ID | 항목 | 후보 sprint plan 단위 |
| --- | --- | --- |
| `RM-M4-01` | WebSocket 확장 — infra/ci/risk event publish | M4-WS sprint |
| `RM-M4-02` | WebSocket replay + 리소스 필터 | M4-WS (RM-M4-01 과 묶음 가능) |
| `RM-M4-03` | frontend command status WebSocket UI | M4-WS-UI |
| `RM-M4-04` | AI Gardener gRPC (Python AnalysisService + Go Core client) | M4-AI |
| `RM-M4-05` | AI Suggestion Feed 실데이터 바인딩 | M4-AI (RM-M4-04 직속 후속) |
| `RM-M4-06` | Gitea Hourly Pull worker (Phase 10) | M4-task |
| `RM-M4-07` | System Admin 대시보드 | M4-admin |
| `RM-M4-08` | RBAC PermissionCache LISTEN/NOTIFY ([ADR-0007](../../docs/adr/0007-rbac-cache-multi-instance.md)) | M4-infra (격리된 backend 작업) |
| `RM-M4-09` | 외부 SSO 통합 (Gitea 등) | M4-SSO |

### 1.1.b Project 도메인 후속 진입 후보 (PR #102 컨셉 1차 머지 후, 본 세션 신규)

본 세션에서 [`docs/planning/project_management_concept.md`](../../docs/planning/project_management_concept.md) 컨셉 1차 머지 완료. 후속은 컨셉 §9 의 4 sprint hook 을 따른다.

| 후속 sprint | 산출물 | 진입 조건 |
| --- | --- | --- |
| **Project-Req** | `docs/requirements.md §5.7` 확장 또는 `§5.8 Project 도메인` 신설, REQ-FR-* 일괄 발급 (개별 usecase 단위), NFR (응답시간/페이지네이션) 정의, 매트릭스 row 추가 | 컨셉 머지 직후 (즉시 가능) |
| **Project-Usecase** | 행위자 × usecase 매트릭스 + 핵심 시퀀스 (등록·조회·멤버 변경 3종) + RBAC 매트릭스 확장 후보 → ADR-0011 초안 | Project-Req 머지 직후 |
| **Project-Design (backend)** | `architecture.md` Project 컴포넌트 추가, `backend_api_contract.md` 신규 § `/api/v1/projects/*`, 마이그레이션 (`000012_projects.sql`) | Project-Usecase 머지 직후 |
| **Project-Design (frontend)** | `frontend_development_roadmap.md` 새 phase, 진입 경로 / 컴포넌트 / store 모델 초안 | Backend design 와 병행 |
| **Project-Impl** | IMPL-project-* / UT-project-* / TC-PROJ-* 발급 + 실 구현 | 모든 design 머지 후 |

### 1.2 M3 carve out (M4 와 병행 가능)

- `getHierarchy` MV join 코드 변경 ([ADR-0009](../../docs/adr/0009-org-secondary-memberships-and-total-count-mv.md) §4.2)
- ETL daily cron 운영 entry ([ADR-0008](../../docs/adr/0008-hrdb-production-adapter.md) §6)
- `primary_dept` backfill worker (signup 직후 + admin trigger, [ADR-0010](../../docs/adr/0010-primary-dept-resolution.md) §4.3)
- 파견 종료 (`is_seconded` 자동 갱신) trigger
- 본격 Vitest (AuthGuard / Header / Sidebar mock-heavy)
- TC-CMD/INFRA 인터랙션 spec ts (`test_cases_m3_command_infra.md` §4)

### 1.3 본 세션 머지 영역 요약 (PR #86~#102, 17건)

- **거버넌스/추적성** 체계 도입 + 1차 종합 매트릭스 + 갭 정리 + 메타 헤더 표준화 + M3/M4 drift 정합.
- **Project 도메인 컨셉 staged** (PR #102): 신규 1차 도메인의 컨셉 단계 문서 신규 작성, 후속 4 sprint (Req → Usecase → Design backend/frontend → Impl) 진입 hook 명문화. 추적성 ID 미발급 (컨셉 단계).
- **M1 PR-D 정합** (A 묶음): writeRBACServerError + writeAuthLoginServerError → writeServerError 통합, X-Request-ID validation, ctx 표준 request_id 전파.
- **X-Devhub-Actor 폐기**: ADR-0004 + ADR-0006 (inbound 400 거부) + 회귀 4 테스트 의도 갱신.
- **5 도메인 본문 ID 노출**: RBAC §12 / auth §11 / account §10.2 / users CRUD §10.3 / organization §10.4 / accounts admin / signup §11.5.2.
- **CI / lint**: GitHub Actions 4 잡 (workflow-lint + backend + frontend + e2e) + ADR-0003 no-docker + ADR-0005 actionlint.
- **M3 1차 closing**: Sign Up audit emit + 단위테스트 + §11.5.2 본격 spec / hrdb.MockClient → PostgresClient (ADR-0008) + migration 000010 / parent_id cycle 검증 + ResolvePrimaryUnit (ADR-0010) + 5 단위테스트 / total_count MV migration 000011 (ADR-0009) / scripts/hrdb_etl_seed.sql / frontend signup form alias.

### 1.4 ADR 인덱스 (본 세션 신규 5건)

| ADR | 제목 |
| --- | --- |
| ADR-0006 | inbound `X-Devhub-Actor` 헤더 명시 거부 (400) |
| ADR-0007 | RBAC PermissionCache 다중 인스턴스 일관성 (PG LISTEN/NOTIFY, 구현 carve) |
| ADR-0008 | HR DB production 어댑터 (PostgreSQL `hrdb` schema) |
| ADR-0009 | 파견/겸임 모델 + `total_count` Materialized View |
| ADR-0010 | `users.primary_unit_id` 자동 판정 알고리즘 |

### 1.5 다음 세션 진입 안내

다음 세션 첫 액션 권장:

1. 본 `session_handoff.md` + `state.json` + `work_backlog.md` 읽고 main HEAD 가 본 housekeeping 직후 commit 인지 확인.
2. 본 §1.1 RM-M4 / §1.1.b Project 도메인 후속 / §1.2 M3 carve out 중 우선순위 결정.
3. 새 sprint branch (`claude/work_260514-a` 또는 `claude/project-req-1` 등) + sprint memory 초기화 → 진행.

## 2. 다음 진입점 — 우선순위 후보

`state.json#next_actions` 참조. 본 sprint 머지 후 후속:

| 후보 | 주요 작업 | 규모 |
| --- | --- | --- |
| document-standards §8 우선순위 3 | 본문 ID 명기 (요구사항/설계 문서에 REQ-FR/ARCH/API 옆 backtick ID) | M |
| document-standards §8 우선순위 4 | deprecated 문서 식별 + 마킹 | S |
| E2E 신규 TC 작성 | TC-CMD-*, TC-INFRA-* — 매트릭스 §5.1 의 후보 구현 | M |
| frontend 컴포넌트 Vitest | Header, Sidebar, AuthGuard 등 | S |
| RBAC API §12 IMPL 정밀 매핑 | endpoint 별 IMPL ID 명시 | S |
| X-Devhub-Actor 폐기 ADR | architecture.md §6.2.3 의 완전 제거 trigger | S |
| RBAC cache 다중 인스턴스 일관성 | M1-DEFER-E | M-L |
| actionlint / workflow lint | ADR-0003 §6 의 후속 ADR | S |
| M3/M4 진입 | command status WebSocket UI, WebSocket 확장, AI Gardener gRPC, Gitea Hourly Pull worker | L |

## 3. 환경 / 운영 메모

- **CI 환경**: GitHub Actions ubuntu-24.04. native PostgreSQL 15 (pgdg) + native Ory Hydra/Kratos v26.2.0. `services:` 컨테이너 사용 안 함 (ADR-0003).
- **5 프로세스 native 기동** (prod / dev-server): PostgreSQL(시스템 서비스) + Hydra + Kratos + backend-core + frontend.
- **e2e 자동 시드**: `cd frontend && npm run e2e` 한 줄.

## 4. 잔여 결정 대기

- 본 sprint 후속 항목 (위 §2 표) 의 우선순위 결정.
- 운영 진입 전 hygiene: PoC `test/test` 시드 제거 (test-server-deployment.md §10).

## 5. 거버넌스 / 추적성 진입점

- `docs/governance/README.md` — 두 축 (문서 관리 + 추적성) 인덱스.
- `docs/governance/document-standards.md` — 문서 표준 (메타 헤더, lifecycle, 단계별 유형).
- `docs/traceability/conventions.md` — ID 컨벤션.
- `docs/traceability/sync-checklist.md` — 매 PR 동기화 절차.
- `docs/traceability/report.md` — 1차 종합 매트릭스.
- `.github/pull_request_template.md` — PR body 의 "추적성 영향" 섹션 자동 안내.
