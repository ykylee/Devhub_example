# 문서 작성·관리 표준

- 문서 목적: DevHub 의 `docs/` + `ai-workflow/memory/` 하위 문서들의 작성 양식, lifecycle, 변경 기록, 추적성 ID 노출 방식을 정의한다.
- 범위: 본 저장소의 모든 long-lived 문서 (코드 주석, PR description, commit message 제외).
- 대상 독자: 모든 contributor (사람 + AI agent), 후속 리뷰어.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-c`.
- 관련 문서: [`../traceability/README.md`](../traceability/README.md) (추적성 체계), [`../traceability/conventions.md`](../traceability/conventions.md) (ID 표준).

## 1. 적용 대상

| 위치 | 유형 |
| --- | --- |
| `docs/requirements.md`, `docs/backend/requirements.md`, `docs/backend_requirements_org_hierarchy.md`, `docs/frontend_integration_requirements.md` | Requirements |
| `docs/architecture.md`, `docs/backend_api_contract.md`, `docs/org_chart_ux_spec.md`, `docs/organizational_hierarchy_spec.md` | Design (spec) |
| `docs/adr/000X-*.md` | Architecture Decision Record |
| `docs/development_roadmap.md`, `docs/frontend_development_roadmap.md`, `ai-workflow/memory/backend_development_roadmap.md` | Roadmap |
| `docs/tests/e2e_testing_strategy.md`, `docs/tests/test_cases_*.md` | Test strategy / TC |
| `docs/tests/reports/*.md` | Test report (timestamped) |
| `docs/setup/*.md` | Setup / operations guide |
| `docs/traceability/*.md`, `docs/governance/*.md` | Governance |
| `ai-workflow/memory/state.json`, `*/session_handoff.md`, `*/work_backlog.md` | Workflow memory (브랜치별 별도 lifecycle, 본 표준의 메타 헤더만 적용) |

## 2. 메타 헤더 (모든 문서)

문서의 첫 H1 직후, 다음 헤더를 부착한다. 누락 항목은 `-` 로 명시.

```markdown
# <문서 제목>

- 문서 목적: <한 줄, 왜 이 문서가 존재하는가>
- 범위: <한 줄, 무엇을 다루고 무엇은 다루지 않는가>
- 대상 독자: <한 줄, 누가 읽는가>
- 상태: draft | proposed | accepted | deprecated
- 결정일: <YYYY-MM-DD>  ※ ADR 만
- 최종 수정일: <YYYY-MM-DD>
- 관련 문서: [<제목>](<경로>), [<제목>](<경로>) ...
```

ADR 은 `결정일` 필수, 다른 문서는 `최종 수정일` 만.

`최종 수정일` 의 갱신 정책: 본문 의미 변경 시 갱신. 오타·서식 수정은 갱신하지 않는다 (`git log` 가 source-of-truth).

## 3. Lifecycle

| 상태 | 의미 | 진입 조건 |
| --- | --- | --- |
| `draft` | 초안. 검토 진행 중. | 작성자 commit 시 default. |
| `proposed` | 결정 대기. (ADR / 정책 문서에서 주로 사용) | 작성자가 PR 으로 제안. |
| `accepted` | 확정. 본문이 source-of-truth. | 리뷰 + 결정권자 승인 후 PR 머지. |
| `deprecated` | 폐기. 후속 문서로 대체됨. | 새 문서 또는 변경 ADR 머지 후 본 문서 상단에 deprecation 노트. |

`deprecated` 문서는 삭제하지 않는다 (cross-reference 깨짐 방지). 본문 첫 줄에 다음 노트 부착:

```markdown
> ⚠ **Deprecated** (YYYY-MM-DD). 후속 문서: [<제목>](<경로>). 본 문서는 참조용으로만 유지된다.
```

## 4. 단계별 문서 유형

### 4.1 Requirements

- 본문 구조: 도메인/모듈 단위 섹션 (예: 2.1 개발자, 2.2 관리자, ...).
- 각 요구사항은 1 문장 또는 짧은 문단. 본문에 backtick 으로 추적 ID 명기: `### 2.5.7 SSO 통합 (REQ-FR-23)`.
- 정량 기준 (성능, 가용성, 보안 threshold) 은 별도 NFR 섹션 또는 `REQ-NFR-*` 로 분리.
- 변경 이력 표는 선택. git log 위임 가능.

### 4.2 Design / Spec

- 컴포넌트 다이어그램, 데이터 흐름, API endpoint 표.
- 큰 결정 (트레이드오프 있는 선택) 은 inline 으로 적지 말고 별도 ADR 로 분기.
- API 계약은 endpoint 별 요청/응답 schema + envelope 규칙 + 상태 코드.
- 본문에 `(ARCH-XX)` / `(API-XX)` 명기.

### 4.3 ADR

- 표준 양식: §1 컨텍스트 → §2 결정 동인 → §3 검토한 옵션 → §4 결정 → §5 결과 (Consequences) → §6 미해결 항목 → §7 변경 이력.
- ID 는 `ADR-{4자리}` (예: `ADR-0001`).
- 결정 후에는 본문 수정 금지. 보충은 새 ADR 으로.
- 변경 이력 표는 필수 (`| 일자 | 변경 |` 형식).

### 4.4 Roadmap

- 마일스톤 단위 표 (`| ID | 항목명 | 출처 PR | 상태 |`).
- 상태: `planned | in_progress | done | discarded`.
- 항목별로 `RM-M{0..N}-YY` 명기.
- 머지된 PR 은 PR 번호 + merge_commit hash 기록 (git 위임 가능).

### 4.5 Test strategy / TC

- 전략 (`e2e_testing_strategy.md`): 시드 정책, 환경 가정, retry 정책.
- TC 카탈로그 (`test_cases_*.md`): TC 별 (목적, 사전조건, 단계, DoD).
- TC ID 는 `TC-<feature>-XX` 표준.

### 4.6 Test report (`docs/tests/reports/`)

- 파일명: `report_<YYYYMMDD>_<scope>.md`.
- 본문: 실행 환경 + spec 별 결과 (PASS/FAIL) + 실패 분석 + 다음 행동.
- timestamped — 갱신 안 함 (스냅샷).

### 4.7 Setup / Guide

- 절차 중심 (단계 별 명령어).
- 환경별 분기 (Linux / Windows / corp SSL inspection).
- 최종 수정일 갱신 시점: 절차 변경 시.

### 4.8 Workflow memory

- 브랜치별 디렉터리 (`ai-workflow/memory/<agent>/<branch>/`).
- `state.json` (JSON, 자동화 친화) + `session_handoff.md` + `work_backlog.md`.
- 머지 후 close 커밋으로 상태 갱신.
- 메타 헤더의 `상태` 는 작업 진행 enum 사용: `planned | in_progress | blocked | done` (CLAUDE.md / AGENTS.md / GEMINI.md 의 작업 원칙과 정합). 본 문서 §3 의 lifecycle (`draft | proposed | accepted | deprecated`) 과는 다른 축임에 주의 — workflow memory 는 작업 lifecycle 이지 문서 lifecycle 이 아니다.

## 5. 추적성 ID 본문 노출

`docs/traceability/conventions.md` 의 ID 체계를 본문에 노출하는 형식:

| 단계 | 본문 노출 위치 | 형식 |
| --- | --- | --- |
| Requirements | 섹션 제목 끝 | `### 2.5.7 SSO 통합 (REQ-FR-23)` |
| Design | 섹션 제목 끝 또는 표 헤더 | `### 6.2.3 BearerTokenVerifier (ARCH-11)` |
| API contract | endpoint 표의 ID 컬럼 또는 endpoint 헤더 옆 | `### POST /api/v1/auth/login (API-20)` |
| Roadmap | 표의 ID 컬럼 | `| RM-M2-01 | OIDC 프록시 ... |` |
| Implementation | 모듈 README 또는 첫 파일 상단 주석 | `// IMPL-rbac-03 — RBAC permission cache. See docs/traceability/report.md` |
| Unit test | 테스트 파일 상단 주석 또는 `t.Run` 명 | `// UT-httpapi-05` |
| E2E | `test()` 제목 (기존 패턴 유지) | `test("TC-NAV-02 — ...", ...)` |

ID 노출은 권장. 누락 시 매트릭스 (`docs/traceability/report.md`) 만으로 추적 가능하지만 본문 노출 시 reader 가 매트릭스 없이도 ID 파악 가능.

## 6. 변경 이력 정책

- ADR: §변경 이력 표 필수.
- 기타 문서: 의미 변경 시 `최종 수정일` 갱신. 큰 sprint 단위 변경은 본문 끝 변경 이력 표 추가 (선택).
- workflow memory: 별도 close 커밋 패턴.

## 7. 리뷰 워크플로우

- 모든 문서 변경은 PR 기반.
- PR body 의 "추적성 영향" 섹션 (`docs/traceability/sync-checklist.md` §3.7) 채움.
- 새 ADR / Roadmap 항목 / 신규 단계 문서는 결정권자 (project lead 또는 sprint owner) 의 명시적 승인.
- 리팩토 / 오타 정리는 약식 PR 허용 (단 PR body 에 "의미 변경 없음" 명시).

## 8. 본 표준의 적용 단계

본 표준 도입 시 기존 문서들은 다음 우선순위로 재정리:

1. **메타 헤더 누락된 문서** → 추가 (sprint `claude/work_260513-c` 이후 별도 PR).
2. **상태 필드 부재** → 추가 (`accepted` default 로 마킹 후 리뷰).
3. **추적성 ID 본문 노출** → 점진 — 새 항목 추가 시점부터 적용, 기존 항목은 큰 sprint 단위로 일괄.
4. **deprecated 문서 마킹** → 식별 후 별도 hygiene sprint.

## 9. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). 본 표준 + 추적성 체계 (`docs/traceability/`) 동시 도입. |
