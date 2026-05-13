# Work Backlog — claude/work_260513-g (B1 auth 도메인 확장)

- 문서 목적: 본 sprint 의 작업 추적.
- 최종 수정일: 2026-05-13

## 진행 상태

- [x] branch + sprint memory 초기화
- [ ] backend_api_contract.md §11.3 (API-19) / §11.5 표 ID 컬럼 / §11.5.1 (API-35) 본문 ID 노출
- [ ] report.md §2.2 auth API 매핑 서브 표 (RBAC 동일 패턴)
- [ ] report.md §2.4 IMPL-auth-XX 책임 정의 서브 표 (7 IMPL)
- [ ] report.md §3 인증 행 — ID 범위 + §2 서브 표 참조
- [ ] report.md §3 회원가입 행 — API-23 ID 본문 매핑
- [ ] report.md §3 계정 관리 행 — API-35 cross-cut 노트
- [ ] report.md §6 변경 이력 + main flat memory sync (PR #92 흡수)
- [ ] PR open + 2-pass + squash merge

## 미진입 (다음 sprint 후보)

- B1 추가 도메인 (account / org / command / audit / infra) 본문 ID 노출 — 점진 진행
- B2 deprecated 문서 식별 + 마킹
- B4 X-Devhub-Actor 폐기 ADR (architecture §6.2.3)
- C1 frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard)
- C2 E2E 신규 TC (TC-CMD-*, TC-INFRA-*)
- D5 actionlint / workflow lint (ADR-0003 §6)
- M1-DEFER-E RBAC cache 다중 인스턴스 일관성
- M3 / M4 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker)
