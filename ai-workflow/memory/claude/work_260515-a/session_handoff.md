# Session Handoff — claude/work_260515-a

- 브랜치: `claude/work_260515-a`
- Base: `main` @ `3f387cd` (rebased onto `25f97ba` 후 housekeeping)
- 날짜: 2026-05-15
- 상태: in_progress (housekeeping)
- 목적: 본인 PR 2건(#114, #115) 4단계 리뷰어 모드 수행 후 main flat memory housekeeping.

## 흡수 대상

| PR | merge | scope |
| --- | --- | --- |
| #112 | 3f387cd | (2026-05-14 머지, flat memory 누락) Admin UI + ActionMenu + iPad 터치 + 백엔드 트랜잭션 |
| #115 | b669bc7 | (본 sprint 머지) Light theme + dropdown refactor + endpoints 통일 모듈 + standalone gate |
| #114 | 25f97ba | (본 sprint 머지) Application leader/dev_unit + search 확장 + auth_login canonical + search predicate refactor |

## 본 sprint 핵심 도입

- **frontend/lib/config/endpoints.ts** — 모든 서비스 URL default 단일 진실 소스. native default + env override 정책 명문화. 8개 service 일괄 갱신. `frontend/.env.example` + `.gitignore` 의 nested `.env.example` 예외 확장.
- **theme FOUC 방지 패턴** — `app/layout.tsx` inline `<script>` + `Header.tsx` lazy initializer + `ThemeToggle.tsx` dead code 제거.
- **`output: "standalone"` NEXT_OUTPUT gate** — `next start` 호환 + docker 빌드 단계 활성화.
- **Application 검색 helper** — `applicationsSearchPredicate` const (count/list 33줄 중복 해소).
- **migration 000020 backfill** — `leader_user_id ← owner_user_id` (기존 row 회귀 방지).

## 본 sprint 디버깅 학습

PR #115 E2E 2회 실패 — service-logs artifact + raw job log 두 layer 분석.
- **Layer 1**: frontend.log 워닝 "next start does not work with output standalone" → standalone gate fix.
- **Layer 2**: raw log 실패 테스트 명 → header-switch-view.spec.ts 의 4건. Switch View 가 dropdown refactor 로 의도적 제거됐는데 e2e 가 잔존 → TC-NAV-01/02/SIM-01 삭제 + TC-NAV-03 selector 갱신.

## 작업 순서

1. (done) PR #114 / #115 CI green 확인 + squash merge
2. (done) main pull + rebase
3. (done) flat state.json / session_handoff.md / work_backlog.md 갱신
4. (in progress) 본 sprint memory 디렉터리 작성
5. (pending) auto memory MEMORY.md 인덱스 갱신
6. (pending) commit + push + open PR + merge

## 다음 세션 우선 작업

state.json `next_session_candidates` + main flat `session_handoff.md §1` 참조.
