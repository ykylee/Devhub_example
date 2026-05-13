# Work Backlog — claude/work_260513-d (갭 분석 + 문서 재정리)

- 문서 목적: 본 sprint 의 작업 추적.
- 최종 수정일: 2026-05-13

## 진행 상태

- [ ] main flat memory sync (PR #87/#88/#89 누적)
- [ ] 매트릭스 §4 ADR-0003 정상 link 활성화 + §3 ※ 정리
- [ ] 갭 §5.1 E2E 미커버 — 분석 후 매트릭스 §5 갱신 + 별도 후보 work_backlog 등재
- [ ] 갭 §5.2 ID 부재 — auth.spec.ts TC-AUTH-01..06 카탈로그 흡수 (직접 fix) + 나머지 후속 등재
- [ ] 갭 §5.3 문서↔코드 불일치 — frontend_integration_requirements §3.8 deprecated 노트 (직접 fix) + 나머지 후속 등재
- [ ] 메타 헤더 표준화 — 누락 4 + 변형 4 (+ frontend_integration_requirements 부분 보강)
- [ ] 검증 + PR + 2-pass + 머지

## 미진입 (다음 sprint 후보)

- document-standards.md §8 우선순위 3 (본문 ID 명기) — 점진 적용
- document-standards.md §8 우선순위 4 (deprecated 마킹) — 별도 hygiene
- E2E 신규 TC 작성: TC-CMD-* (command lifecycle), TC-INFRA-* (인프라 토폴로지)
- frontend 컴포넌트 Vitest 단위테스트 추가 (Header, Sidebar, AuthGuard)
- RBAC API §12 IMPL 정밀 매핑
- X-Devhub-Actor 완전 제거 일정 결정 + ADR
- RBAC cache 다중 인스턴스 일관성 (M1-DEFER-E)
- backend-ai 실 구현 (M3-04)
- actionlint / workflow lint (ADR-0003 §6)
