# Governance — DevHub 거버넌스 체계

- 문서 목적: DevHub 의 문서 관리 표준 + 추적성 관리 체계를 단일 진입점으로 안내한다.
- 범위: 본 디렉터리 + `docs/traceability/` 의 cross-reference.
- 대상 독자: 모든 contributor (사람 + AI agent), 외부 감사.
- 상태: accepted
- 최종 수정일: 2026-05-13

## 두 축

DevHub 의 거버넌스는 다음 두 축으로 운영된다:

| 축 | 위치 | 역할 |
| --- | --- | --- |
| **문서 관리** | [`docs/governance/`](./) | 문서의 작성 양식·lifecycle·변경 기록·ID 노출 표준. |
| **추적성 관리** | [`docs/traceability/`](../traceability/) | SDLC 단계 (요구사항/설계/로드맵/구현/UT/E2E) 사이 항목 식별과 매핑. |

두 축은 서로 참조한다:

- 문서 관리의 §5 (추적성 ID 본문 노출) 가 추적성의 conventions.md 를 따른다.
- 추적성의 sync-checklist.md (§3.7 PR body 의 추적성 영향) 는 문서 관리의 §7 (리뷰 워크플로우) 안에서 운영된다.

## 디렉터리 구조

```
docs/
├── governance/
│   ├── README.md             ← 본 문서
│   └── document-standards.md ← 문서 작성·관리 표준
└── traceability/
    ├── README.md             ← 추적성 체계 개요
    ├── conventions.md        ← ID 컨벤션 표준
    ├── sync-checklist.md     ← PR 동기화 절차
    └── report.md             ← 종합 추적성 매트릭스 (1차)
```

## AI agent 진입점

- `CLAUDE.md`, `AGENTS.md`, `GEMINI.md` 모두 본 README 와 traceability/README 를 진입점으로 안내한다.
- 새 작업 시작 시: 본 README → document-standards.md → traceability/README.md → traceability/conventions.md 순으로 확인.
- 작업 종료 시: traceability/sync-checklist.md §3 절차 수행 + document-standards.md §6 변경 기록 정책 준수.

## 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). |
