# Integrated Work Backlog (main, post M1 RBAC)

- 문서 목적: main 브랜치 기준 상위 백로그 인덱스. 세부 sprint backlog 는 브랜치별 메모리 디렉터리 참조.
- 범위: 마일스톤 상태, 최근 머지, 잔여/후속 작업
- 대상 독자: 프로젝트 리드, 후속 에이전트, 트랙 담당자
- 상태: M1·M2·M3 done (1차). Application Domain backend 1차 (2026-05-14). DREQ 도메인 종합 closing. External Integration 도메인 1차 종합 closing. **2026-05-18 단일 세션 누적 23 PR** (EOD #1 #141~#152 12건 + post-EOD #1 #153~#158 6건 + post-EOD #2 #159~#164 6건 + 본 sprint x). **ADR carve out 종결 누적 8건**: ADR-0017 §6 atomicity (-o) + ADR-0015 §6 (1)+(2) (-p) + ADR-0016 §6 (1)+(2) (-s) + ADR-0017 §6 (a)+(c)+(d) (-t). **design 검토 2건**: single port reverse proxy (ADR-0018 후보, sprint -u) + Keycloak SSO federation (ADR-0019 후보, sprint -v). main HEAD `4dd02ad`. **다음 세션 directive**: (1) ADR-0015 §6 (3)+(4), (2) ADR-0016 §6 (3)+(4)+(5), (3) ADR-0017 §6 (b), (4) ADR-0018/0019 승격 + Phase 2 staging, (5) M4 RM-M4-XX — session_handoff.md "다음 세션 directive" 절 참조.
- 최종 수정일: 2026-05-18
- 관련 문서: [통합 로드맵](../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json), [M1 PR 리뷰 actions](./M1-PR-review-actions.md)

## 1. 마일스톤 진행 상황

| 마일스톤 | 상태 | 종료 일자 | 메모 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 | ✅ done | 2026-05-08 | PR #14·15·16·17·18·19. SEC-1~4 resolved. T-M0-10 운영 검증 PASS. |
| **M1** — 핵심 기능 contract 정합성 | ✅ done | 2026-05-11 | PR-B+C (#56) + PR-D (#57) 머지로 envelope / types / WS / CommandStatus / audit actor enrichment / request_id 완결. PR-D 후속 (#80/#82) 으로 commands enrichment + DEVHUB_TRUSTED_PROXIES 보강. |
| **M2** — 사용자 경험 정합 | ✅ done (1차 완성) | 2026-05-12 | login_action + work_26_05_11 + work_26_05_11-d 완료 후 PR #85 (`claude/login_usermanagement_finish`) 가 1차 완성 sprint 로 닫음: 로드맵 정합 + UX hygiene (PR-UX1+2+3) + Kratos audit (PR-M2-AUDIT) + 30 TC e2e 게이트. |
| **CI** — GitHub Actions 도입 | ✅ done (1차) | 2026-05-13 | PR #86 (`gemini/prepare-github-action`) 가 backend-unit + frontend-unit + e2e (Playwright 40 TC) 3잡 도입. 리뷰어 모드 2-pass 에서 5 blocker + follow-on 5 발견 → 보강 commit 7개로 그린 도달. PR-T5 의 핵심 잡 묶음은 본 PR 으로 1차 완료. 후속: FU-CI-1 (no-docker policy 정합), FU-CI-2/3/4 는 `claude/work_260513-a` 처리 중. |
| **M3** — Realtime 확장 + 외부 연동 1차 | ✅ done (1차) | 2026-05-13 | RM-M3-01..03 (Sign Up + 인사 DB + 조직 polish) 완료. ADR-0008/0009/0010 신규. |
| **Application Domain (backend 1차)** | ✅ done | 2026-05-14 | API-01~58 전체 activated (본 세션 #104~#110). 마이그레이션 000012~000018 (7), ADR-0011 accepted, RBAC 4 신규 resource (system_admin 일임), CI 5 job (Backend Integration 신설). 23 integration test (P1/P2 회귀 guard 포함) 가 CI 에서 실 실행. |
| **DREQ Domain (closing 1차 + TC-DREQ-* 13건)** | ✅ closing | 2026-05-18 | API-59..68 + API-79 (PATCH allowed_ips) activated + ADR-0012/0013/0014/**0017**. P2 carve out 6/6 모두 해소 + codex hotfix #5 (store revoked guard + 회귀 가드 3 unit test). **TC-DREQ-* 13건 정식 발급** (sprint d) + `dev-requests.spec.ts` 6 step + `test_cases_m5_dreq.md` 신규. 잔여 carve out: ADR-0017 §6 atomicity 실제 구현 (CTE refactor) + 자동 cron revoke + last_used staleness alert. |
| **External Integration Domain (1차 종합 closing)** | ✅ closing | 2026-05-18 | concept staged (PR #135) → backend 1차 (PR #139) → **ADR-0015/0016/0017** (sprint c) → provider frontend (sprint g/h) + API-80 DELETE (sprint j) → **bindings 관리 UI** (sprint m, PR #154) + **topology v2 시각화** (sprint n, PR #155) + **ADR-0017 §6 atomicity** 실 구현 (sprint o, PR #156) + **ADR-0015 §6 (1)+(2)** (sprint p, PR #157). API-69..80 모두 activated. TC-INT-FRONTEND-* 10건 + TC-INT-FRONTEND-BIND-* 3건 + TC-INT-HOMELAB-03 + TC-INT-FRONTEND-TOPOLOGY-V2-* 2건. ADR carve out 추가 종결 (atomicity + size limit + token rotation SOP). 잔여 carve: ADR-0016 §6 (Alertmanager YAML / Grafana JSON / p95 alert / push 알림 / 임계), ADR-0017 §6 잔여 (cron revoke / PATCH expires_at / 만료 alert), ADR-0015 §6 (3)+(4) (dedicated worker / push-pull dedup), React Flow group sub-node + WebSocket 실시간. |
| **M4** — 운영 / SSO / MFA / 후속 ADR | planned | — | 통합 로드맵 §3.5. RM-M4-01..09 (9 항목). |

## 2. 최근 머지 (M1 RBAC track, 2026-05-08)

| PR | 제목 | 머지 |
| --- | --- | --- |
| #20 | feat(http): SEC-5 mask 5xx errors + M1 sprint plan (PR-A) | `ae8aca1` |
| #21 | docs(adr): ADR-0002 RBAC policy edit API (PR-F) | `950a11f` |
| #22 | docs(api): RBAC §12 rewrite + route mapping (PR-G1) | `1a090a3` |
| #23 | feat(rbac): domain + rbac_policies migration (PR-G2) | `5239a87` |
| #29 | feat(rbac): postgres store + users.role FK (PR-G3 + FIX-A) | `24815b8` |
| #30 | feat(rbac): RBAC handlers (PR-G4 + FIX-B + FIX-C) | `27b6817` |
| #31 | feat(rbac): permission cache + enforcement (PR-G5) | `02eef35` |
| #27 | feat(frontend): RBAC PermissionEditor ↔ backend (PR-G6 + FIX-D) | `e02ba67` |
| #28 | docs(memory): M1 PR review actions tracker | `9bc30c9` |

원본 PR #24, #25, #26 은 stack base 자동 삭제로 close 후 main 위에서 #29/#30/#31 로 재등록.

## 3. M1 잔여 + DEFER 후속 작업

### 3.1 M1 잔여 (P1, 진입 시 분해)

- **T-M1-02 (PR-B)** — API envelope/role wire/필드 정합 (`backend_api_contract.md` ↔ 코드 1:1 강제, 통합 테스트 매트릭스).
- **T-M1-03 (PR-C)** — command lifecycle 6 상태 일관 적용 + dry-run/live 경계 테스트.
- **T-M1-04 (PR-D)** — Audit actor 보강 (`source_ip`, `request_id`, `source_type` + 마이그레이션 + 미들웨어 + 응답 헤더 `X-Request-ID`).
- **T-M1-05** — `auth_test.go` prod 가드 + role 가드 통합 테스트 매트릭스. PR-C 또는 PR-D 에 흡수 가능.
- **T-M1-07 (frontend)** — `frontend/lib/services/types.ts` UI vs wire 분리, 표시 포맷 프론트 이전. PR-B 에 묶거나 단독.
- **T-M1-08** — WebSocket envelope `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합. PR-B 에 묶거나 단독.

### 3.2 DEFER (M1 PR 리뷰 후속, [상세](./M1-PR-review-actions.md#3-다음-개발로-넘김--defer))

- **M1-DEFER-A** — `rbac_policies` 의 `is_system` ↔ `role_id` 일관성 CHECK 제약 (P2 방어선)
- **M1-DEFER-B** — `requireMinRole`/`roleMeetsMin`/`roleRank` deadcode 정리 + 단위 테스트 제거
- **M1-DEFER-C** — `writeRBACServerError` 임시 helper → `writeServerError` 통합 (PR-G4 의 TODO)
- **M1-DEFER-D** — `DeleteRBACRole` row-lock (다중 인스턴스 race 강화)
- **M1-DEFER-E** — `PermissionCache` 다중 인스턴스 일관성 (pub/sub 또는 polling)
- **M1-DEFER-F** — API contract §12.4 / §12.5 응답 예시 추가
- **M1-DEFER-G** — `MemberTable` role display 회귀 사용자 환경 검증

### 3.2 P1~P2 should-have

- **frontend `/auth/callback`**: Hydra authorization_code → `/oauth2/token` 교환 후 세션 저장 흐름. PR-D 의 자연 후속.
- **frontend `account.service.ts`**: Kratos public flow 호출 helper.
- **frontend `types.ts` 분리**: UI 표시명 vs API wire 타입 분리, 표시 포맷팅을 프론트로 이전.
- **WebSocket envelope 표준화**: `{schema_version, type, event_id, occurred_at, data}` 코드/문서 정합.
- **RBAC policy 편집 API**: 12.x — write/audit 경계 + persistence 또는 *static-default 유지* 결정.

### 3.3 P3 nice-to-have

- ADR-0002 (Gitea SSO 통합) 작성.
- ADR (commits 정규화 테이블 도입 여부).
- ADR (OS service wrapper 결정).

## 4. 운영 / 환경 메모

- Hydra/Kratos PoC: e2e 검증 완료. binary 위치 `$env:USERPROFILE\go\bin\` (Windows). 다음 세션에서 재가동은 `hydra serve all --config infra/idp/hydra.yaml --dev` + `kratos serve --config infra/idp/kratos.yaml`.
- 검증용 임시 OIDC client (client_credentials, id `43aa4b74-...`) 는 Hydra 안에 잔존. 보안 위험은 없으나 cleanup 가능.
- backend-core `go test ./...` 는 사내 GoProxy mirror 환경에서 PASS 검증됨.

## 5. 변경 이력 (요약)
자세한 변경 이력은 [Traceability Report](./report.md)의 §6 변경 이력을 참조하십시오.
