# Work Backlog — feat/work_260611-a-n13-inbound-source-impl

- 문서 목적: N-13 backend foundation sprint 의 백로그. PR A (foundation 만).
- 범위: migration + domain + repository + view handler + 4 UT. **코드 변경 +340 line**. routing/auto_route.go + voc_handler 통합 + openapi.yaml = PR A-2 (별도).
- 상태: in_progress (PR 발행 pending, go test PASS)
- 최종 수정일: 2026-06-11

## 1. 태스크 (PR A)

- [x] WB-01: branch 생성 (`feat/work_260611-a-n13-inbound-source-impl`)
- [x] WB-02: domain.Platform 2 field 추가 + IsValidPlatformInboundSourceType helper
- [x] WB-03: migration 000007 up + down (CHECK + 2 인덱스 + consistency constraint)
- [x] WB-04: repository UpdatePlatformInboundSource + ListEnabledInboundSourcePlatforms + ErrInvalidInboundSourceType/Config sentinel
- [x] WB-05: repository ScanPlatform + PlatformsInsertQuery + UpdatePlatform 정합 (2 col 추가)
- [x] WB-06: PlatformStore interface 2 method 추가
- [x] WB-07: view/applications.go UpdatePlatform inbound_source 별도 처리 + platformResponse echo
- [x] WB-08: fake store (view + httpapi) 2 method + json.Valid check
- [x] WB-09: view/applications_handler_test.go 4 UT 추가
- [x] WB-10: go test ./... PASS
- [x] WB-11: branch memory directory (state.json + session_handoff.md + work_backlog.md + backlog/2026-06-11.md)
- [ ] WB-12: PR 발행 (push + gh pr create)
- [ ] WB-13: PR 머지 (사용자 confirm 후)
- [ ] WB-14: main flat memory 3 file finalize (post-merge sync)

## 2. 잔여 (PR A-2, 별도 sprint)

- routing/auto_route.go 신규 (pattern matcher 3 case + auto route 1 case)
- voc_handler.createOrGetVoc 자동 routing 호출 + 응답 (auto_routed + dev_request_id)
- backend IT (routing 1 case + auto route 1 case)
- openapi.yaml 정합 (PATCH /platforms inbound_source + POST voc auto_routed 응답)
- ADR-0028 §6 amendment (구현 sprint 에서)

## 3. 관련 PR

- **PR #516** (sprint plan, 2026-06-12 머지, `maintenance/work_260612-c-inbound-source-plan`)
- **PR #547** (N-13 housekeeping, 2026-06-11 머지, `feat/work_260611-a-n13-inbound-source-housekeeping`) — ID slot 9 row 발급
- **PR (pending, 본 sprint)**: N-13 backend foundation (PR A)

## 4. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | sprint 시작 + 9 file 변경 + 4 UT + go test PASS + branch memory + PR 발행 pending |
