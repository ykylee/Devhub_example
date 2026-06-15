# 추적성 관리 (Traceability)

- 문서 목적: DevHub 의 SDLC 자산 (요구사항 → Usecase → 설계 → 로드맵 → 구현 → 단위테스트 → E2E) 사이 추적 관계를 단일 매트릭스로 관리하고, 모든 PR 이 이 관계를 갱신하도록 유지한다. 본 체계는 [거버넌스의 추적성 축](../governance/README.md) 이며, 문서 관리 축 ([`docs/governance/document-standards.md`](../governance/document-standards.md)) 과 짝을 이룬다.
- 범위: 본 디렉터리의 4 파일 + PR 템플릿 + 에이전트 가이드 진입점.
- 대상 독자: 모든 contributor (사람 + AI agent), 후속 리뷰어, 외부 감사.
- 상태: accepted
- 최종 수정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-c`.

## 1. 왜 추적성?

`docs/requirements.md` → `docs/planning/*usecase*.md` → `docs/architecture.md` → `docs/development_roadmap.md` → `backend-core/internal/...` + `frontend/...` → `*_test.go` + `*.test.ts` → `frontend/tests/e2e/*.spec.ts` 가 각각 독립으로 진화하면 다음 위험이 발생:

- **회귀 누락**: 새 요구사항이 구현은 됐는데 단위/E2E 테스트가 누락.
- **고아 코드**: 어느 요구사항에서 출발했는지 모르는 코드가 누적.
- **deprecate 추적 손실**: 변경 ADR (예: ADR-0003) 의 영향 받는 코드/문서가 흩어져 있어 일괄 갱신 누락.
- **AI agent 가 문서를 참조해 작업할 때**: 어떤 문서가 source-of-truth 인지 모호하면 잘못된 가정으로 코드 작성.

본 체계는 단일 매트릭스 (`report.md`) + ID 컨벤션 (`conventions.md`) + 동기화 절차 (`sync-checklist.md`) 로 위 위험을 줄인다.

## 2. 디렉터리 구조

| 파일 | 역할 |
| --- | --- |
| [`README.md`](./README.md) | 본 문서 — 체계 개요와 인덱스. |
| [`conventions.md`](./conventions.md) | ID 접두사·형식·영속성·매핑 정책. |
| [`sync-checklist.md`](./sync-checklist.md) | PR 단위 추적성 갱신 절차. |
| [`report.md`](./report.md) | 단계별 인덱스 + 종합 추적성 매트릭스 + ADR 링크 + gap 요약. |
| [`traceability_remediation_plan_auth_org.md`](./traceability_remediation_plan_auth_org.md) | 로그인 세션 + 사용자/조직 추적성 미흡 항목 보완 계획. |

## 3. 무게감 (B. Medium)

본 체계는 다음 결정 (sprint `claude/work_260513-c`, 2026-05-13) 으로 운영:

- **포함**: 표준 + 동기화 절차 + PR 템플릿 의 추적성 영향 섹션 + 에이전트 가이드 진입점 안내.
- **제외**: 정적 매트릭스 무결성 검사 CI lint (후속 ADR 후보).

이 무게감의 의미: 사용자/AI agent 가 책임지고 매트릭스를 동기화하되, GH Actions 가 무결성을 강제하지는 않는다. 위반 시 다음 PR 또는 sprint 의 cleanup 으로 보완 가능 (`sync-checklist.md` §5).

## 4. 흐름 한 페이지

```
[새 작업 시작]
   |
   v
[PR 작성자는 본 PR 의 변경 범위를 식별]
   |
   v
[sync-checklist.md §1 표로 영향 단계 결정]
   |
   v
[영향 단계 별로 sync-checklist.md §3.1~§3.5 절차 적용]
   |  - 새 ID 발급 (conventions.md §3.1)
   |  - 기존 ID 유지 + cross-ref 갱신 (§3.2)
   v
[docs/traceability/report.md 의 인덱스 + 매트릭스 갱신 (§3.6)]
   |
   v
[PR body 의 "추적성 영향" 섹션 채움 (§3.7)]
   |
   v
[리뷰어는 §4 manual 검증]
   |
   v
[merge]
```

## 5. AI agent 진입점

- `CLAUDE.md` / `AGENTS.md` / `GEMINI.md` 의 "추적성" 단락이 본 디렉터리로 안내한다. 새 작업 진입 시 본 디렉터리의 4 파일을 우선 참조.
- 작업 종료 시 `report.md` 매트릭스 갱신을 누락하지 않을 것 — `sync-checklist.md` §3.6 참조.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). 4 파일 + PR 템플릿 + 에이전트 가이드 진입점 동시 도입. |
| 2026-05-13 | 로그인 세션 + 사용자/조직 추적성 보완 계획 문서 추가 (`traceability_remediation_plan_auth_org.md`). |
