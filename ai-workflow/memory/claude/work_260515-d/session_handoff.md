# Session Handoff — claude/work_260515-d (codex hotfix)

- 브랜치: `claude/work_260515-d`
- Base: `main` @ `519a508`
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: Codex 외부 리뷰의 P1/P2 inline 코멘트 4건 처리.

## codex 코멘트 정리 + 처리 방침

| PR | 등급 | 위치 | 처리 |
| --- | --- | --- | --- |
| #114 | P1 | ApplicationCreationModal edit payload `key` | 본 sprint 에서 fix |
| #114 | P2 | Tables/모달의 `new Date(YYYY-MM-DD)` timezone shift | 본 sprint 에서 fix |
| #115 | P1 | next.config localhost fallback | 이미 해소 (`f621189`) — no action |
| #118 | P1 | enforceRowOwnership production dead code | pmo_manager seed + handler 호출 일괄 도입 |

## carve out (다음 sprint)

- **owner-self 활성화** — 현재 route-level RBAC gate 가 applications:edit/delete 를 system_admin (또는 pmo_manager seed 후) 만 허용하므로 owner 본인이 route gate 에서 막힘. owner-self 가 통과하려면 route gate 정책 변경이 필요한데, 이는 ADR 갱신과 함께 정책 결정. 본 hotfix scope 외.
- **owner_user_id NULL backfill** — 새 row guard 가 NULL owner 인 경우 어떻게 행동하는지 정합.

## 작업 순서

1. (in progress) 브랜치 메모리 초기화
2. (planned) PR #114 P1 fix — ApplicationCreationModal edit 분기에서 key 제외
3. (planned) PR #114 P2 fix — date-only safe parse
4. (planned) PR #118 P1 fix — pmo_manager seed + handler 호출
5. (planned) 검증 + PR + self-review + merge
