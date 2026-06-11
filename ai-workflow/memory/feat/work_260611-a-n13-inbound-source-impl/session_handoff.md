# Session Handoff — feat/work_260611-a-n13-inbound-source-impl

- 문서 목적: N-13 (project.inbound_source 자동 routing) 의 backend foundation sprint. PR #547 (N-13 housekeeping) 의 ID slot 9 row 의 코드 변경 본. 본 PR = PR A (foundation 만, routing/voc_handler/openapi 는 별도 PR A-2).
- 범위: migration 000007 + domain.Platform + repository (3 method) + view handler + view/applications.go + PlatformStore interface + 2 fake store + 4 UT. **코드 변경 +340 line, 신규 ID 매핑 = PR #547 의 9 row 정합**.
- 상태: branch `feat/work_260611-a-n13-inbound-source-impl` 작업 완료, push/PR 발행 pending.
- 최종 수정일: 2026-06-11

## 0. 본 세션 핵심 결과

### 변경 요약 (9 file, +382 -37, **build + 모든 UT PASS**)

| 파일 | 변경 | line |
|---|---|---|
| `backend-core/migrations/000007_platform_inbound_source.up.sql` | 신규 — platforms.inbound_source_type + inbound_source_config 컬럼 + CHECK 제약 + 2 인덱스 + consistency constraint | +41 |
| `backend-core/migrations/000007_platform_inbound_source.down.sql` | 신규 — 롤백 | +21 |
| `backend-core/internal/domain/application.go` | `domain.Platform` +2 field + `IsValidPlatformInboundSourceType` helper | +18 |
| `backend-core/internal/domain/application-lifecycle/repository/applications.go` | `platformsSelectColumns` +2 col + `ScanPlatform` +2 scan + `PlatformsInsertQuery` +2 col + `UpdatePlatform` +2 col + `ErrInvalidInboundSourceType/Config` sentinel + `UpdatePlatformInboundSource` method + `ListEnabledInboundSourcePlatforms` method | +98/-1 |
| `backend-core/internal/domain/application-lifecycle/view/handler.go` | `PlatformStore` interface +2 method | +3 |
| `backend-core/internal/domain/application-lifecycle/view/applications.go` | `updatePlatformRequest` +2 field + `UpdatePlatform` 핸들러 inbound_source 별도 처리 + `platformResponse` echo | +123 |
| `backend-core/internal/domain/application-lifecycle/view/applications_handler_test.go` | 4 UT 추가 (GiteaOK / InvalidType400 / InvalidConfig400 / DisableEmpty) | +82 |
| `backend-core/internal/domain/application-lifecycle/view/fake_store_test.go` | `UpdatePlatformInboundSource` + `ListEnabledInboundSourcePlatforms` fake impl + json.Valid check | +49 |
| `backend-core/internal/httpapi/applications_test.go` | `memoryPlatformStore` +2 method + json import | +46 |

### 정공법 핵심

1. **PR A scope = backend foundation 만**. routing/auto_route.go + voc_handler 통합 + openapi.yaml 정합은 PR A-2 (별도 sprint).
2. **inbound_source 격리**: `inboundTouched` 일 때 `UpdatePlatform` 호출 skip + `GetPlatform` 로 row 확인 + `UpdatePlatformInboundSource` 단독 호출. **inbound_source 부분 실패가 다른 필드 변경에 영향 없도록 격리**.
3. **migration 000007 의 CHECK 제약**: type='' 일 때 config NULL/'{}' 만 허용 (consistency). type whitelist = `''|gitea|jira|other` (4 종).
4. **routePermissionTable 변경 불요** — 기존 PATCH /platforms entry 가 `ResourcePlatforms + ActionEdit` 으로 inbound_source 도 cover.
5. **fake store parity**: fakeViewPlatformStore + memoryPlatformStore 모두 production UpdatePlatformInboundSource 동작 (json.Valid check) mirror.

### Pre-flight / Safety

- **Tier**: 사내 (backend 코드 변경, 사내 repo = GitHub push-only). main branch 보호 — CI 통과 + 사용자 confirm 후 머지.
- **go build ./...**: PASS
- **go test ./...**: PASS (모든 package ok)
- **go vet ./...**: PASS (사소한 self-assignment dev_requests.go:74 = 본 PR scope 외, 기존 코드)
- **routePermissionTable 변경 불요** (기존 PATCH /platforms entry 가 inbound_source cover)

## 1. 다음 세션 directive

1. **본 PR 발행 + 머지** (사용자 confirm 후).
2. **PR 머지 후 main flat memory 3 file finalize** (state.json / session_handoff.md / work_backlog.md 의 N-13 implementation close 마킹).
3. **PR A-2** (별도 sprint): `routing/auto_route.go` 신규 (pattern matcher 3 case + auto route 1 case) + `voc_handler.createOrGetVoc` 자동 routing 호출 + `openapi.yaml` 정합 + ADR-0028 §6 amendment. 본 PR A 가 foundation 의 ID slot SoT 역할.
4. 또는 다른 sprint (N-6 staging 운영 사용자 결정 / backend-integration DEVHUB_BUILD_TIER matrix / 다른 housekeeping).

## 2. 후속 (사용자 결정 영역)

- **PR 머지 시점**: 사용자 confirm 후.
- **PR A-2 진입 시점**: 사용자 결정.

## 3. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — N-13 backend foundation + 4 UT + go test PASS + branch memory + PR 발행 pending |
