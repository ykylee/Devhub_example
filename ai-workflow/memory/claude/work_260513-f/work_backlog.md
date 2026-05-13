# Work Backlog — claude/work_260513-f (B 묶음: RBAC IMPL 정밀 매핑 + API §12 본문 ID 노출)

- 문서 목적: 본 sprint 의 작업 추적.
- 최종 수정일: 2026-05-13

## 진행 상태

- [x] branch + sprint memory 초기화
- [ ] backend_api_contract.md §12.2 ~ §12.10 endpoint 헤더 (API-26..31, 38..40) 노출
- [ ] report.md §2.2 RBAC API 매핑 표 추가
- [ ] report.md §2.4 IMPL-rbac-XX 정의 명시 (handler / store / enforcement / cache)
- [ ] report.md §3 RBAC 행 IMPL 컬럼 정밀 매핑
- [ ] report.md §5.2 "RBAC API §12 IMPL 정밀 매핑" closed
- [ ] report.md §6 변경 이력 + main flat memory sync (PR #91 흡수)
- [ ] PR open + 2-pass + squash merge

## 미진입 (다음 sprint 후보)

- B1 의 다른 도메인 (auth / account / org / command / audit / infra) 본문 ID 노출 — 점진 진행
- B2 (document-standards §8 우선순위 4) — deprecated 문서 식별 + 마킹
- B4 — X-Devhub-Actor 폐기 ADR
- C1 — frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard)
- C2 — E2E 신규 TC (TC-CMD-*, TC-INFRA-*)
- D5 — actionlint / workflow lint (ADR-0003 §6 후속 ADR)
- M1-DEFER-E RBAC cache 다중 인스턴스 일관성
- M3 / M4 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker)
