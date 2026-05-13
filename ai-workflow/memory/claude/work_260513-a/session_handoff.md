# Session Handoff — claude/work_260513-a (2026-05-13 진행 중)

- 문서 목적: PR #86 머지 직후 잔여 작업 sprint 인계.
- 범위: main flat memory sync (#85/#86 반영) + FU-CI-2/3/4 (GHA 최적화 3건).
- 최종 수정일: 2026-05-13
- 상태: IN PROGRESS. 브랜치 분기 직후.

## 1. 진입 컨텍스트

- 직전 sprint = `gemini/prepare-github-action` (PR #86, `450cc24` 로 squash merge). 머지 직후 cleanup 으로 본 슬롯 진입.
- PR #86 의 `work_backlog.md` (merged) 에 FU-CI-1..4 follow-up 4건 인계.
- FU-CI-1 (no-docker policy 정합) 만 정책 결정 필요로 본 sprint 에서 제외. 나머지 3건 묶음 + main flat memory sync 한 PR.

## 2. 작업 목록

### 2-1. main flat memory sync (FU-A)

- `ai-workflow/memory/state.json` — `head_commit`, `merged_prs_2026_05_*` 에 #85/#86 추가. `next_actions.tdd_followups` 의 PR-T5 항목을 partial-done 으로 마킹하고 FU-CI-1 만 잔여 항목으로 남김.
- `ai-workflow/memory/session_handoff.md` — EOD 2026-05-13 으로 헤더 갱신. 0 절에 PR #85/#86 머지 사실 + 본 sprint 진행 사실 기록.
- `ai-workflow/memory/work_backlog.md` — FU-CI-1 등재 + FU-CI-2/3/4 는 본 sprint 결과 commit 으로 처리됨을 후속 기록.

### 2-2. FU-CI-2: Playwright install 범위 축소

- 변경 위치: `.github/workflows/ci.yml` 의 `Install Playwright Browsers` step.
- 변경: `npx playwright install --with-deps` → `npx playwright install --with-deps chromium`.
- 근거: `frontend/playwright.config.ts` 의 project 가 `chromium` 단일 + `channel: "chrome"`. 다른 브라우저 다운로드는 dead weight.

### 2-3. FU-CI-3: GHA cache (go modules + playwright browsers)

- 변경 위치: `.github/workflows/ci.yml` 의 backend-unit, e2e job (frontend-unit 은 setup-node 캐시 이미 잡힘).
- 변경: `actions/cache@v4` 로 `~/go/pkg/mod` + `~/.cache/go-build` + `~/.cache/ms-playwright` 캐시.
- Key: backend-unit/e2e 의 go 캐시는 `${{ runner.os }}-go-${{ hashFiles('backend-core/go.sum') }}`. playwright 캐시는 `${{ runner.os }}-pw-${{ hashFiles('frontend/package-lock.json') }}`.

### 2-4. FU-CI-4: Wait for App Readiness frontend timeout 분리 상향

- 변경 위치: `.github/workflows/ci.yml` 의 `Wait for App Readiness` step.
- 변경: backend `/health` 60 s 유지, frontend `/` 60 s → 120 s. Next.js 16 cold boot 여유 확보.

## 3. 검증 계획

- 로컬: `go test ./backend-core/...` + `cd frontend && npm run test` (둘 다 그대로 통과).
- CI: PR 생성 후 backend-unit / frontend-unit / e2e 3 잡 SUCCESS 재확인. 캐시 첫 회 적용은 보통 cache miss 라 1회차는 기존 시간과 비슷, 2회차부터 단축 검증.

## 4. 다음 행동

1. ci.yml 수정 (FU-CI-2/3/4).
2. main flat memory 3 파일 sync.
3. commit + push + PR 생성.
4. CI 그린 확인 → 리뷰어 모드 2-pass → 머지.
5. 머지 후 본 슬롯 close 커밋 (이전 sprint 패턴: `docs(memory): close work_260513-a slot`).
