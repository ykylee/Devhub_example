# Governance — DevHub 거버넌스 체계

- 문서 목적: DevHub 의 문서 관리 표준 + 추적성 관리 체계를 단일 진입점으로 안내한다.
- 범위: 본 디렉터리 + `docs/traceability/` 의 cross-reference.
- 대상 독자: 모든 contributor (사람 + AI agent), 외부 감사.
- 상태: accepted
- 최종 수정일: 2026-06-16

## 두 축

DevHub 의 거버넌스는 다음 두 축으로 운영된다:

| 축 | 위치 | 역할 |
| --- | --- | --- |
| **문서 관리** | [`docs/governance/`](./) | 문서의 작성 양식·lifecycle·변경 기록·ID 노출 표준. |
| **추적성 관리** | [`docs/traceability/`](../traceability/) | SDLC 단계 (요구사항/설계/로드맵/구현/UT/E2E) 사이 항목 식별과 매핑. |
| **코드 분류** | [`code-taxonomy.md`](./code-taxonomy.md) | 코드베이스 12 카테고리 + 4 횡단 카테고리 + 모듈 SoT + 작업 prefix 컨벤션 + 리팩토링 후보 P0~P3. |

두 축은 서로 참조한다:

- 문서 관리의 §5 (추적성 ID 본문 노출) 가 추적성의 conventions.md 를 따른다.
- 추적성의 sync-checklist.md (§3.7 PR body 의 추적성 영향) 는 문서 관리의 §7 (리뷰 워크플로우) 안에서 운영된다.

## 문서 포맷 기본 원칙

- 원본 문서는 Markdown(`.md`)으로 관리한다.
- HTML은 보고/취합/배포용 파생 산출물로만 사용하며 수동 수정하지 않는다.

## 디렉터리 구조

```
docs/
├── governance/
│   ├── README.md                              ← 본 문서
│   ├── document-standards.md                  ← 문서 작성·관리 표준
│   ├── worker_division.md                     ← 워커 분업 historical + 사외/사내 2-tier 정책 (§0, §6)
│   ├── keycloak_admin_responsibility.md       ← IdP 팀 ↔ DevHub 운영자 책임 분리 협약 (ADR-0020 §3.2 승격)
│   ├── code-taxonomy.md                       ← 코드 분류 12+4 카테고리
│   ├── l1-format.md                           ← in-repo wiki L1 page format SSOT
│   ├── wiki-cross-reference.md                ← wiki ↔ code cross-reference SSOT
│   └── branch-cleanup-sop.md                  ← 원격/로컬 branch 일괄 정리 SOP (3-tier 분류 + backup 절차)
└── traceability/
    ├── README.md             ← 추적성 체계 개요
    ├── conventions.md        ← ID 컨벤션 표준
    ├── sync-checklist.md     ← PR 동기화 절차
    └── report.md             ← 종합 추적성 매트릭스

## AI agent 진입점

- `CLAUDE.md`, `AGENTS.md`, `GEMINI.md` 모두 본 README 와 traceability/README 를 진입점으로 안내한다.
- 새 작업 시작 시: 본 README → document-standards.md → traceability/README.md → traceability/conventions.md 순으로 확인.
- 작업 종료 시: traceability/sync-checklist.md §3 절차 수행 + document-standards.md §6 변경 기록 정책 준수.

## 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). |
| 2026-05-22 | `keycloak_admin_responsibility.md` 신규 — ADR-0020 §3.2 책임 매트릭스 승격 + escalation path + 명시 금지 5건 (sprint `claude/work_260522-internal-coordinated-carve-docs`). |
| 2026-06-09 | **워커 분업 전면 취소 (사용자 결정)** — `worker_division.md` 가 §0 + §1~§4 의 historical 標記 + §2.5 branch prefix 자유화 + §5 Owner 권한 명시로 전면 改 編. Claude/Codex 의 자유 이용 불가로 분배 무효화. 유지 정책: §4.2 ADR supersession 정공법, §5 Owner 권한, 우선순위 P0~P3 (sprint `maintenance/work_260609-a-cancel-worker-division`). |
| 2026-06-16 | `branch-cleanup-sop.md` 신규 — 원격 78 + 로컬 41 일괄 cleanup 의 SOP 표준화 (3-tier 분류, backup dump, dry-run 절차, retention). 백업 위치: `.git/branch-backup/<UTC-timestamp>/`. 디렉터리 트리 갱신. |
