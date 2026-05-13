# Work Backlog — claude/work_260513-e (A 묶음 PR-D 정합 마무리)

- 문서 목적: 본 sprint 의 작업 추적.
- 최종 수정일: 2026-05-13

## 진행 상태

- [x] branch + sprint memory 초기화
- [ ] A3 — writeRBACServerError 통합 (rbac.go helper 삭제 + 11곳 치환)
- [ ] A1 — caller-supplied X-Request-ID validation (정규식/길이/control char)
- [ ] A1 단위테스트 (4-5건)
- [ ] A2 — ctx 표준 request_id 전파 (requireRequestID + ctx-aware helper)
- [ ] A2 — kratos_login_client.go 2건 + kratos_identity_resolver.go 1건 ctx-aware 치환
- [ ] A2 단위테스트 (ctx fallback + logRequestCtx)
- [ ] go test ./backend-core/... PASS
- [ ] 추적성 매트릭스 row 갱신 (§3, §6)
- [ ] PR open + 2-pass + squash merge

## 미진입 (다음 sprint 후보)

- E2E 신규 TC (TC-CMD-*, TC-INFRA-*)
- frontend 컴포넌트 Vitest (Header, Sidebar, AuthGuard)
- RBAC API §12 IMPL 정밀 매핑
- X-Devhub-Actor 완전 제거 일정 결정 + ADR
- document-standards §8 우선순위 3 (본문 ID 명기)
- document-standards §8 우선순위 4 (deprecated 마킹)
- RBAC cache 다중 인스턴스 일관성 (M1-DEFER-E)
- backend-ai 실 구현 (M3-04)
- actionlint / workflow lint (ADR-0003 §6)
- M3/M4 진입 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull worker)
