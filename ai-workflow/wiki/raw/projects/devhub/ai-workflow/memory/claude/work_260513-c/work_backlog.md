# Work Backlog — claude/work_260513-c (추적성 분석 sprint)

- 문서 목적: 본 sprint 의 단계별 진행 추적.
- 최종 수정일: 2026-05-13

## Phase 진행 상태

- [x] Phase 1 — 요구사항 분석 (REQ-FR-01..105, REQ-NFR-01..26)
- [x] Phase 2 — 설계 분석 (ARCH-01..17, API-01..40, ADR 인덱스)
- [x] Phase 3 — 로드맵 분석 (RM-M0..M3 = 28 항목, M4 planned)
- [x] Phase 4 — 구현 분석 (IMPL backend 47 + frontend 32)
- [x] Phase 5 — 단위테스트 분석 (UT backend 41 + frontend 6)
- [x] Phase 6 — E2E 분석 (TC M2 28 + M3 9)
- [x] 거버넌스 체계 추가 (사용자 확장 요청, `docs/governance/`)
- [x] Phase 8 — 추적성 보고서 작성 (`docs/traceability/report.md`)
- [ ] Phase 9 — 검증 + PR + 2-pass + 머지 (진행 중)
- [-] Phase 7 — 기존 문서 일괄 재정리 — 별도 PR 로 분기 (본 PR 스코프 밖)

## 발견 사항 누적 (Phase 진행하며 채움)

### Gap 후보 — 문서 ↔ 코드 불일치

- (Phase 진행하며 채움)

### Gap 후보 — ID 누락 / 표 결측

- (Phase 진행하며 채움)

## 미진입 (다음 sprint 후보)

- M4 (WebSocket 확장, AI Gardener gRPC, Gitea Hourly Pull worker) 의 단계별 추적성 추가
- 외부 도구용 JSON/CSV 산출물 추출
- actionlint / workflow 린트 도입 (ADR-0003 §6 의 미해결 항목)
- caller-supplied X-Request-ID validation
- ctx 표준 request_id 전파
- writeRBACServerError → writeServerError 통합
