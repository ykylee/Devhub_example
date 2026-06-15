# Work Backlog — gemini/prepare-github-action

- 문서 목적: PR #86 리뷰어 모드 2-pass 에서 의도적으로 본 PR 스코프 밖으로 둔 follow-up 작업을 기록한다.
- 최종 수정일: 2026-05-13
- 상태: open (PR #86 본체는 그린 상태로 머지 대기)

## 본 PR 에서 미반영 (별도 PR 권장)

머지 직전 리뷰 코멘트 (PR #86 https://github.com/ykylee/Devhub_example/pull/86#issuecomment-4435695605) 의 "Non-blocker / 후속 논의용" 4건을 그대로 인계한다. 본 PR 의 스코프 (E2E CI 그린 만들기) 를 넘기 때문에 별도 PR 로 분리한다.

### FU-CI-1 — No-Docker policy 와 `services: postgres:` 충돌 검토

- **현황**: `.github/workflows/ci.yml` 의 E2E 잡이 `services: postgres:15` 컨테이너를 띄움 — GitHub Actions ephemeral sidecar 이라 deploy 와는 별개 layer 지만, 메모리의 "DevHub 서비스 전체는 Docker/컨테이너 미사용 전제" 와 형식적으로 충돌.
- **선택지**:
  1. 현 상태 유지 (CI 한정 예외) — 정책 본문에 "CI ephemeral sidecar 는 제외" 명시.
  2. ubuntu-latest 의 preinstalled PostgreSQL 14 를 `systemctl start postgresql` 로 띄우고 user/db 만 생성 — 정책 완전 일관, 단 트레이드오프: PG 14 (현 service 는 15).
- **How to apply**: 정책 본문 우선 결정 후 PR 진행. 의사결정 ADR 1 페이지 권장.

### FU-CI-2 — Playwright install 최적화

- **현황**: `npx playwright install --with-deps` 가 Chromium/Firefox/WebKit + 시스템 deps 까지 모두 다운로드 (~100 MB+) 하지만 `frontend/playwright.config.ts` 의 project 는 `chromium` 단일 + `channel: "chrome"` (호스트 시스템 Chrome). bundled 브라우저는 사용 안 함.
- **변경 안**: `npx playwright install-deps chromium` (OS deps 만) 또는 `npx playwright install --with-deps chromium` (chromium + deps) 로 좁힘. ubuntu-latest 에는 google-chrome-stable preinstall.
- **이득**: install 단계 30–60 s 단축, CI 비용/시간 절감.

### FU-CI-3 — actions cache 미적용 (Go modules, Playwright browsers)

- **현황**: `actions/setup-node@v4` 가 npm 캐시는 자동으로 잡지만, Go modules (`~/go/pkg/mod`, `~/.cache/go-build`) 와 playwright browsers (`~/.cache/ms-playwright`) 는 캐시 없음. PR 마다 cold install.
- **변경 안**: `actions/cache@v4` 로 `go-build` + `go-mod` + `ms-playwright` 캐시 추가. Key 는 `${{ runner.os }}-go-${{ hashFiles('backend-core/go.sum') }}` 등.
- **이득**: E2E 잡 총 시간 8 분 → 4–5 분.

### FU-CI-4 — Wait for App Readiness 60s 여유 부족 가능성

- **현황**: `.github/workflows/ci.yml` 의 `Wait for App Readiness` 가 `timeout 60s` 로 backend `/health` + frontend `/` 응답을 기다림. Next.js 16 cold boot (next start) 가 CI 컨테이너에서 종종 30–50 s 라 빠듯.
- **변경 안**: backend 60 s 유지, frontend 120 s 로 분리 상향. 또는 backend `wait_for_http` helper 함수로 retry 로직 명시.
- **트리거**: 첫 flake 시 즉시 반영.

## 머지 후 PR-T5 정합 확인

- `ai-workflow/memory/state.json` (main 브랜치 flat copy) 의 `tdd_followups` 항목 중 "PR-T5 (CI 도입, DEC-4) — GitHub Actions: backend go test + frontend Vitest + e2e matrix 또는 nightly. (사용자가 본 시점 진입 보류 결정, 2026-05-12)" 가 본 PR 머지로 일부 충족. main backlog 갱신 시 PR-T5 상태를 partial-done 으로 마킹하고 본 backlog 의 FU-CI-1..4 를 PR-T5 후속으로 링크.
