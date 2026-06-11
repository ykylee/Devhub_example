# Session Handoff — feat/work_260611-a-n13-inbound-source-housekeeping

- 문서 목적: N-13 (project.inbound_source 자동 routing, ADR-0028 §6 (a)) 의 housekeeping 정공법 sprint. PR #516 (sprint plan) + PR #517 (N-13 row) + ADR-0028 §6 (2026-06-12 정합) 의 결과를 정합. 본 sprint = docs only, 구현은 v1.1 진입 시점 별도 sprint.
- 범위: docs/traceability/{conventions.md, report.md} + docs/adr/0028-dev-requests-voc-external-ref.md + docs/planning/release_v1_roadmap.md. 코드 변경 0줄. **신규 ID 발급 9 row** (REQ-FR-113 / UC-DEV-REQ-15 / ARCH-23 / API-103 / RM-DEV-REQ-15 / IMPL-inbound-source-01 / IMPL-platform-patch-02 / UT-inbound-source-01 / TC-INBOUND-SRC-01).
- 상태: branch `feat/work_260611-a-n13-inbound-source-housekeeping` 작업 완료, push/PR 발행 pending.
- 최종 수정일: 2026-06-11

## 0. 본 세션 핵심 결과

### 변경 요약 (4 file, +22 -17줄, docs only)

| 파일 | 변경 | line |
|---|---|---|
| `docs/traceability/conventions.md` | §1 표에 RM 표기 정책 확장 — `RM-{domain}-{nn}` (도메인 prefix) 관행 명문화 + §1 의 RM row 예시에 `RM-DEV-REQ-15` 추가 + `{nn}` 가이드 직후 RM 표기 정책 3-bullet 추가 | 2 row + 5 line |
| `docs/traceability/report.md` | 헤더 메타 (상태 / 최종 수정일 / 결정 근거 sprint) + §2.1 REQ-FR-113 + §2.1.5 UC-DEV-REQ-15 + §2.2 ARCH-23 / API-103 + §2.3 RM-DEV-REQ-15 + §2.4 N-13 IMPL callout + §2.5 UT-inbound-source-01 + §2.6 TC-INBOUND-SRC-01 + §3 dev-request row + §4 ADR 인덱스 ADR-0028 row | 11 row + 1 callout |
| `docs/adr/0028-dev-requests-voc-external-ref.md` | §6 (a) 의 ID slot 9 row 발급 사실 + 도메인 prefix 관행 정합 + §7 변경 이력 1 row | 2 row |
| `docs/planning/release_v1_roadmap.md` | 헤더 메타 (최종 수정일) + §3.5 N-13 row 의 status `⏳ planned` 보강 + ID slot 정공법 명시 + §9 변경 이력 1 row | 3 row |

### 정공법 핵심

1. **N-13 정공법은 이미 PR #516 (sprint plan, 2026-06-12) + ADR-0028 §6 amendment 에서 결정 완료** — 본 sprint 는 그 결과를 traceability 매트릭스 + conventions.md + roadmap 에 정합.
2. **구현은 v1.1 milestone 진입 시점 별도 sprint** (`feat/work_260611-a-n13-inbound-source-impl`) — 본 sprint 는 docs only housekeeping.
3. **RM 표기 관행 (도메인 prefix)** 정공법 — `RM-M{n}-{nn}` (legacy milestone) + `RM-{domain}-{nn}` (M-v1.1+ 도메인 특화) 두 관행 혼용 가능. RM-DEV-REQ-15 가 정공법 형식.

### 신규 ID 발급 9 row (planned, v1.1 진입 시점 코드 변경)

| ID | 단계 | 의미 |
|---|---|---|
| `REQ-FR-113` | Requirements | N-13 inbound_source 자동 routing 정책 + 매칭 전략 |
| `UC-DEV-REQ-15` | Usecase | 외부 시스템 의뢰 → applications 매칭 → dev-request 직접 등록 flow |
| `ARCH-23` | Design | applications.inbound_source schema + sync routing 결정 |
| `API-103` | Design | PATCH `/api/v1/platforms/:platform_id` inbound_source + POST `/api/v1/dev-requests/:external_ref` auto_routed 응답 |
| `RM-DEV-REQ-15` | Roadmap | N-13 sprint -d 구현 (planned) |
| `IMPL-inbound-source-01` | IMPL | repository UpdatePlatformInboundSource + routing/auto_route.go + voc_handler 통합 |
| `IMPL-platform-patch-02` | IMPL | UpdatePlatform 입력 validate + inbound_source_type CHECK 매핑 |
| `UT-inbound-source-01` | UT | pattern matcher 3 case + auto route 1 case + store 1 case |
| `TC-INBOUND-SRC-01` | E2E | PATCH inbound_source → POST voc → auto_routed 검증 |

### Pre-flight / Safety

- **Tier**: 공용 (docs only, 사내 한정 정보 미포함)
- **check-tier-separation.sh**: PASS 예상
- **check_docs.py**: PASS 예상 (report.md / conventions.md / ADR-0028 / roadmap 정합)
- **CI 4/4 PASS 예상** (path-detect → docs 만 변경 감지, backend/e2e/frontend skip)

## 1. 다음 세션 directive

1. **본 PR 발행 + 머지** (사용자 confirm 후).
2. **PR 머지 후 main flat memory 3 file finalize** (state.json / session_handoff.md / work_backlog.md 의 N-13 housekeeping close 마킹).
3. **구현 sprint 진입 시점** (v1.1 milestone): `feat/work_260611-a-n13-inbound-source-impl` 분기, ID slot 9 row 의 코드 변경 (migration 000007 + domain.Platform 2 field + repository.UpdatePlatformInboundSource + routing/auto_route.go + voc_handler 통합 + PATCH /platforms + UT/IT + E2E + ADR-0028 §6 amendment). 본 sprint 의 housekeeping 정합이 ID 의 SoT 역할.
4. 또는 다른 sprint (N-6 staging 운영 / N-13 housekeeping 후속 / backend-integration matrix / C-i 후속).

## 2. 후속 (사용자 결정 영역)

- **PR 머지 시점**: 사용자 confirm 후.
- **구현 sprint 진입 시점**: v1.1 milestone 결정 후.

## 3. 변경 이력

| 일자 | 변경 |
|---|---|
| 2026-06-11 | 본 sprint — N-13 housekeeping 정공법 + 9 ID row 발급 + PR 발행 |
