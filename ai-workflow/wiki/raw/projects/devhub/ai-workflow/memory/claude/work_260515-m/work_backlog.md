# Work Backlog — claude/work_260515-m (DREQ-RBAC-ADR + DREQ-Promote-Tx)

- 상태: in_progress (PR open 대기)
- 최종 수정일: 2026-05-15
- 위치: DREQ carve out 4건 묶음의 1/4

## items

- [done] sprint 메모리 초기화
- [done] ADR-0013 — DREQ RBAC row-scoping 정책 ADR 신규
- [done] store.RegisterDevRequestWithNewApplication + RegisterDevRequestWithNewProject 트랜잭션 (BeginTx + INSERT + UPDATE + Commit)
- [done] applications.go INSERT SQL 들을 const 로 추출 (drift 방지)
- [done] handler registerDevRequest 분기 (legacy target_id / application_payload / project_payload)
- [done] unit test 6건 신규 + memoryDevRequestStore 확장 (knownRepoIDs / knownDevUnits / createdApps 등)
- [done] backend_api_contract §14.4 API-62 spec 갱신
- [done] traceability report §3 DREQ row + §4 ADR-0013 row 갱신
- [done] development_roadmap §3 M5 DREQ entry 갱신
- [planned] main flat state.json + session_handoff.md + work_backlog.md 갱신
- [planned] commit + push + PR + 4단계 self-review + squash merge
