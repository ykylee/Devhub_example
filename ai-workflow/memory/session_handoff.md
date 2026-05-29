# Session Handoff — main (2026-05-29 EOD, post-SDLC-restructure)

- 문서 목적: 2026-05-29 세션의 후속 carve out + SDLC 문서 재정비 sprint 7 PR 머지 후 main 상태와 인계 사항.
- 범위: PR #407 (cleanup-recovery) 머지 후 동일 세션에서 8 신규 PR (#408~#415) 머지로 carve out P1 처리 + SDLC 문서 도메인-모듈 재정비 완료.
- 상태: 모든 sprint PR 머지, main HEAD `273d9d4`.
- 최종 수정일: 2026-05-29 EOD

## 1. 본 세션 7 신규 PR (이전 #408 housekeeping 포함 = 8건)

| PR | Commit | 내용 |
|---|---|---|
| #408 | `68c2d15` | docs(memory) — main flat housekeeping (cleanup-recovery 결산) |
| #409 | `6eefda9` | refactor(backend/shared-integrationcaps) — providerHasCapability 3 카피 통합 + 11 unit test |
| #410 | `33594ed` | docs(governance-docs) — SDLC Phase 1: 도메인 디렉터리 골격 + README 진입점 (13 file) |
| #411 | `7d3d20d` | docs(governance-docs) — SDLC Phase 2: planning/+tests/ 도메인별 이관 + cross-reference 일괄 갱신 (25 rename + 91 file 갱신) |
| #412 | `0b5907a` | test(frontend/multi-domain-view) — view 컴포넌트 단위테스트 +210 (25 file, 584 tests PASS) |
| #413 | `7d390f7` | docs(governance-docs) — SDLC Phase 3: REQ/ARCH/API split (34 신규 + 3 master index 전환) |
| #414 | `c00b104` | docs(governance-docs) — SDLC Phase 4: traceability/report.md §3 매트릭스 10 도메인 SoT 재구성 (21 → 19 row) |
| #415 | `273d9d4` | docs(governance-docs) — SDLC Phase 5: document-standards.md §4 위치 가이드 갱신 |

## 2. SDLC 문서 도메인-모듈 재정비 결과

코드베이스의 3대 레이어 + 4대 계층 + 10 도메인 구조 (PR #406/#407) 와 SDLC 문서가 1:1 mirror 정합.

- `docs/domain/<도메인>/{requirements,architecture,api,test_cases}.md` × 10 도메인 = 40 sub-document.
- `docs/api/conventions.md` (cross-cutting envelope/enum 신규).
- master file 3건 (requirements / architecture / backend_api_contract) → index 전환.
- `docs/shared/README.md` + `docs/infrastructure/README.md` 신규 진입점.
- `docs/traceability/report.md` §3 매트릭스 19 row (10 core + 1 Shared + 7 Infra + 1 Cross-cutting).
- `docs/governance/document-standards.md` §4 위치 가이드 새 구조 명시.

## 3. 후속 carve out (별도 sprint)

| 우선순위 | 항목 | 사유 |
|---|---|---|
| P1 | CI e2e + backend-integration 복원 | refactor 정리 stabilize 됐을 때 `&& false` 제거 |
| P1 | view 컴포넌트 큰 modal coverage 70% | ApplicationCreationModal (57%) + ProjectCreationModal (39%) edit-mode + member CRUD 시퀀스 |
| P2 | ApplicationRepository cross-domain decouple | `*IntegrationRepository` embed 제거 (review agent P1) |
| P2 | ApplicationStore interface slim | 13+ integration 메서드 → integration domain |
| P2 | §2 인덱스 도메인 분류 정합 | traceability/report.md §2 (line 22-349) 의 cross-cutting row → 새 도메인 row 정합 (Phase 4 scope 외) |
| P3 | rbac/audit/org 신규 임시 ID 정규화 | Phase 3 임시 발급 ID (REQ-RBAC/AUDIT/ORG/ARCH-RBAC/AUDIT/ORG) 가 본 sprint 매트릭스 흡수됐으나 ID prefix 자체의 추적성 ID 컨벤션 (REQ-FR-<DOMAIN>-XXX) 와 정합 정리 |
| P3 | application-lifecycle/api.md §9.1 대시보드 JSON sample 위치 결정 | master SoT vs sub-document SoT |

## 4. 본 sprint 학습

1. **branch 침범 회피** — view carve agent 가 Phase 3 branch (`-f`) 에 commit 한 case. agent 위임 시 명확한 branch 명시 필수. cherry-pick 분리 패턴으로 회복.
2. **stash --include-untracked 활용** — 다른 branch 작업 사이 working tree 보존 (Phase 3 agent 결과 stash → 분리 branch → unstash).
3. **거대 sprint 의 Phase 분할** — 단일 거대 PR 보다 Phase 별 PR 분할이 review 부담 작음. 본 sprint 5 Phase + 추가 2 PR 으로 분할.
4. **sub-agent 위임 ROI** — Phase 3 (REQ/ARCH/API split XL) + view carve (24 file) 같은 대량 작업은 sub-agent 위임이 직접 작업보다 효율 큼.
5. **doc/code mirror 패턴** — code-taxonomy SoT 의 10 도메인 + 4 계층이 코드 + 문서 양쪽 1:1 mirror 일 때 navigation + ownership 명확.

## 5. 다음 세션 directive

후속 carve out 우선순위:
1. CI 복원 (사내 검증 후) → `e2e` + `backend-integration` 의 `&& false` 제거
2. view 큰 modal coverage 70% (별도 sprint)
3. ApplicationRepository decouple + ApplicationStore slim (P2 묶음 가능)
4. §2 인덱스 도메인 분류 정합 (별도 sprint)
