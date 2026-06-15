# Session Handoff — claude/work_260515-h (codex hotfix #2)

- 브랜치: `claude/work_260515-h`
- Base: `main` @ `4d0277f` (PR #122 머지 직후)
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: Codex 외부 리뷰 hotfix #2 — PR #119/#120/#121 의 P1×2 + P2×2 일괄 처리.

## codex 4건 처리 방침

| PR | 등급 | 위치 | fix |
| --- | --- | --- | --- |
| #119 | P1 | `migrations/000021.down.sql` | DELETE 전에 `UPDATE users SET role='developer' WHERE role='pmo_manager'` 추가 |
| #121 | P1 | `api_contract §14.7` (API-65 close) | 인증을 `system_admin only` 로 (REQ-FR-DREQ-008 + ARCH-DREQ-04 정합) |
| #121 | P2 | `requirements.md REQ-FR-DREQ-001` | source_system 을 client required → server-derived 표기 |
| #120 | P2 | `work_260515-*/state.json` | 머지된 7 sprint 의 status → done finalize 패턴 정착 |

## 작업 순서

1. (done) 메모리 초기화
2. (planned) migration 000021 down FK reassign
3. (planned) API §14.7 close 권한 fix
4. (planned) REQ-FR-DREQ-001 source_system 표기
5. (planned) 머지된 7 sprint state.json status finalize
6. (planned) commit + push + PR + CI + self-review + merge
