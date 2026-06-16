# Work Backlog — opencode/work_260604-i-n4-frontend-tests

- Branch: `opencode/work_260604-i-n4-frontend-tests`
- Agent: opencode (Sisyphus, model `MiniMax-M3`)
- Updated: 2026-06-04

## 0. 본 sprint 의 목표

release_v1_roadmap §3.5 N-4. baseline 82% statements / 980 tests / 74 files. recent 변경 (PR #470/471/472) 컴포넌트의 vitest 단위테스트 추가.

## 1. 작업 단위 분해

| ID | 작업 | 상태 |
| --- | --- | --- |
| WB-01 | 브랜치 + memory set up | done |
| WB-02 | RepositoryEditModal.test.tsx | planned |
| WB-03 | RepositoryCreationModal.test.tsx | planned |
| WB-04 | projects/page.test.tsx | planned |
| WB-05 | admin/catalog/page.test.tsx | planned |
| WB-06 | vitest + tsc + lint 검증 | planned |
| WB-07 | 커밋 + push + PR | planned |

## 2. baseline (2026-06-04)

```
Statements   : 82.03% ( 2950/3596 )
Branches     : 76.63% ( 2033/2653 )
Functions    : 79.86% ( 833/1043 )
Lines        : 83.33% ( 2660/3192 )
Tests        : 980 passed (74 files, 10.55s)
```

## 3. 추가 테스트 범위

| 컴포넌트 | 시나리오 | 테스트 수 |
| --- | --- | --- |
| RepositoryEditModal | pre-fill, SCM dropdown, PATCH submit, change-detection, error, onClose | 8 |
| RepositoryCreationModal | POST submit, SCM dropdown, error, onClose | 5 |
| projects/page | list render, status filter, progress loading/empty/actual, delete/archive | 6 |
| admin/catalog | tabs render, repo filter, edit button (draft only), delete confirm, modal open | 8 |

목표: +27 tests (980 → 1007)

## 4. 검증 기준 (DoD)

- [ ] 4 test files 신규 (test/test.tsx)
- [ ] vitest 0 fail (목표 1007+ tests)
- [ ] tsc 0 errors
- [ ] eslint 0 new errors
- [ ] CI frontend-unit PASS
- [ ] coverage statements 82% → ≥84%

## 5. carry-over (sprint 종료 후)

- **#1 Gitea issue sync (backend)** — Claude
- **Codex P2 잔여** + **Application progress_percent** + **v1.0 N-6/P1-6/P2**
