# Work Backlog — claude/work_260513-h (B4: X-Devhub-Actor 폐기 ADR)

- 문서 목적: 본 sprint 의 작업 추적.
- 최종 수정일: 2026-05-13

## 진행 상태

- [x] branch + sprint memory 초기화
- [ ] ADR-0004 작성 (`docs/adr/0004-x-devhub-actor-removal.md`)
- [ ] `docs/architecture.md` line 174 갱신
- [ ] `docs/adr/0001-idp-selection.md` §8 #4 인라인 갱신
- [ ] `backend-core/internal/httpapi/me.go` line 16 주석 정리
- [ ] `docs/traceability/report.md` §4 ADR 인덱스 + §5.3 closed + §6 변경 이력
- [ ] main flat memory sync (PR #93 흡수)
- [ ] PR open + 2-pass + squash merge

## 미진입 (다음 sprint 후보)

- B1 추가 도메인 (account / org / command / audit / infra) 본문 ID 노출 — 점진 진행
- B2 deprecated 문서 식별 + 마킹 (document-standards §8 우선순위 4)
- C1 frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard)
- C2 E2E 신규 TC (TC-CMD-*, TC-INFRA-*)
- D5 actionlint / workflow lint (ADR-0003 §6)
- M1-DEFER-E RBAC cache 다중 인스턴스 일관성
- M3 / M4 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker)
