# Work Backlog — feat/work_260611-a-n13-inbound-source-housekeeping

- 문서 목적: N-13 housekeeping 정공법 sprint 의 백로그.
- 범위: docs/traceability/{conventions, report}.md + docs/adr/0028 + docs/planning/release_v1_roadmap.md. 코드 변경 0줄. 신규 ID 발급 9 row.
- 상태: in_progress (PR 발행 pending, CI 4/4 PASS 예상)
- 최종 수정일: 2026-06-11

## 1. 태스크 (sprint)

- [x] WB-01: branch 생성 (`feat/work_260611-a-n13-inbound-source-housekeeping`)
- [x] WB-02: conventions.md §1 RM 표기 정책 확장 (도메인 prefix 관행 명문화)
- [x] WB-03: traceability report.md 9 row 발급 (§2.1/§2.1.5/§2.2/§2.3/§2.4/§2.5/§2.6 + §3 dev-request row + §4 ADR 인덱스 ADR-0028 + 헤더 메타)
- [x] WB-04: ADR-0028 §6 (a) ID slot 정공법 + §7 변경 이력 1 row
- [x] WB-05: release_v1_roadmap §3.5 N-13 row + §9 변경 이력 + 헤더 메타
- [x] WB-06: branch memory directory 작성 (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-11.md)
- [ ] WB-07: PR 발행 (push + gh pr create, 사용자 confirm 후)
- [ ] WB-08: PR 머지 (사용자 confirm 후)
- [ ] WB-09: main flat memory 3 file finalize (post-merge sync)

## 2. 잔여 (별도 sprint, v1.1 진입 시점)

- `feat/work_260611-a-n13-inbound-source-impl` 분기:
  - migration 000007 (applications.inbound_source_type + inbound_source_config + CHECK)
  - domain.Platform 2 field 추가
  - platform-lifecycle repository UpdatePlatformInboundSource method + view API
  - routing/auto_route.go 신규 (pattern matcher + 3 매칭 전략)
  - voc_handler.createOrGetVoc 자동 routing 호출 + 응답 (auto_routed + dev_request_id)
  - backend UT + IT (pattern matcher 3 case + auto route 1 case + store 1 case)
  - openapi.yaml 정합
  - ADR-0028 §6 amendment (구현 sprint 에서)

## 3. 관련 PR

- **PR #516** (`2b3c7661`, 2026-06-12 머지, `maintenance/work_260612-c-inbound-source-plan`): sprint plan 본 검토
- **PR #517** (`2222cb09`, 2026-06-12 머지, `docs/planning/2026-06-12-inbound-source-routing-sprint-plan.md` 정합)
- **PR #546** (이전 sprint, `chore/n10-memory-drift-2026-06-11`): N-10 housekeeping (별도)
- **본 PR (pending)**: N-13 housekeeping 정합

## 4. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | sprint 시작 + 4 file 정공법 + branch memory + PR 발행 pending |
