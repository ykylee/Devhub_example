# Session Handoff — docs/x6-keycloak-groups-public

- 문서 목적: X-6 (Keycloak group staging-prod 적용, issue #214) 의 사외 PR 정공법
- 범위: ADR-0036 + sprint plan + memory 4 file (docs only, Tier: 공용)
- 상태: 작업 완료, push + PR 발행 pending
- 최종 수정일: 2026-06-14

## 0. 본 세션 핵심 결과

### X-6 정공법 (2-PR 분할)

**문제**: issue #214 P1-3 잔여 = staging/prod Keycloak admin console 적용 + 운영 SOP 보강. 단 **사내 한정 정보 (Keycloak admin URL, client secret) 포함** → GitHub public 저장소 push 안 됨 (AGENTS.md §6).

**결정**: 2-PR 분할
- **사외 PR (본 PR)**: ADR-0036 + sprint plan + memory (Tier: 공용, GitHub)
- **사내 PR (사내 Gitea)**: `scripts/setup-keycloak-groups.sh` + runbook (Tier: 사내, 사내 SCM)

### ADR-0036 핵심 결정

- group 4종 + composite realm role 1:1 매핑 (keycloak_groups_mapping.md §3.1)
- Default Groups 미설정 (codex review #9)
- backend 변경 없음 (sub-carve C event listener 가 P1-1 완료)
- staging 1주 관찰 후 prod 적용 (운영 안전성)
- 사내 script = idempotent + dry-run + check-only + remove flag (3 flag)

### 발견: docs 이미 정합

`docs/setup/keycloak_operations.md §4.3/§4.4` + `scripts/verify-keycloak-groups.sh` 가 이미 merge. **X-6 의 docs only 부분은 status row update 만**.

## 1. 변경 (사외 PR, 4 file + commit)

1. `docs/adr/0036-keycloak-groups-staging-prod.md` (NEW, ~10KB, 9 section) — Accepted
2. `docs/planning/2026-06-14-x6-keycloak-groups-sprint-plan.md` (NEW, ~10KB)
3. 메모리 4 file (`ai-workflow/memory/docs/x6-keycloak-groups-public/`)
+ 메모리 commit

## 2. Tier 분류

- **공용** (이 PR)
- **사내** (사내 Gitea PR — `scripts/setup-keycloak-groups.sh` + runbook)

## 3. 신규 ID (4 row)

- `REQ-FR-KEYCLOAK-GROUPS-01`
- `ARCH-KEYCLOAK-GROUPS-01`
- `RM-KEYCLOAK-GROUPS-01`
- `IMPL-KEYCLOAK-GROUPS-01` (사내 PR reference)

## 4. 잔여 follow-up (사용자 결정 영역)

- **staging Keycloak admin console 적용** (사용자 수동, 1주 관찰 후 prod)
- **prod Keycloak admin console 적용** (사용자 수동)
- **staging test user 의 token role emit 검증** (사용자 follow-up)
- **prod smoke test** (staging 1주 후)
- **사내 PR (infra/x6-keycloak-groups-internal)**: `scripts/setup-keycloak-groups.sh` + runbook — 사내 Gitea push

## 5. 다음 세션 directive

- 본 사외 PR commit + push + PR 발행 + 머지
- 위키 mirror 갱신: `bash scripts/wiki-sync-devhub.sh` 1회 실행 (PR 머지 후)
- 사내 PR commit (사내 Gitea push) — 사용자 follow-up
- 다음 sprint: X-8 (Keycloak SPI realm events) 또는 X-4 Phase 2 (handler post-commit wire)
