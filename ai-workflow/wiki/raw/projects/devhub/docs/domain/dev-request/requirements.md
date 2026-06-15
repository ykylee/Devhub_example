# dev-request 도메인 요구사항 (DREQ)

- 문서 목적: 외부 시스템 개발 의뢰(DREQ) 수신 → 담당자 검토 → application/project 등록(promote) 흐름의 기능·비기능 요구사항을 정의한다.
- 범위: REQ-FR-DREQ-001..013 / REQ-NFR-DREQ-001..006. 외부 시스템 일반 연동(Integration) 은 `docs/domain/integration-registry/requirements.md`, application/project 모델은 `docs/domain/platform-lifecycle/requirements.md` 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/requirements.md` §5.5 에서 이관 — 본문 보존)
- 관련 문서: [도메인 README](./README.md), [컨셉](./concept.md), [architecture.md](./architecture.md), [api.md](./api.md), [master requirements](../../requirements.md), [ADR-0012](../../adr/0012-dreq-external-intake-auth.md), [ADR-0013](../../adr/0013-dreq-rbac-row-scoping.md), [ADR-0014](../../adr/0014-application-project-lifecycle.md)

## 1. 개요

본 도메인은 컨셉 문서([`./concept.md`](./concept.md), sprint `claude/work_260515-f`)에 정의된 외부 시스템 개발 의뢰 수신 → 담당자 검토 → application/project 등록(promote) 흐름의 요구사항을 정의한다.

## 2. 기능 요구사항 (REQ-FR-DREQ)

- **REQ-FR-DREQ-001 (MVP):** 외부 시스템은 인증된 API 호출로 개발 의뢰를 등록할 수 있어야 한다.
    - **Client 가 보내는 필수 필드**: `title` (≤200자), `requester` (외부 시스템상의 의뢰자 식별자), `assignee_user_id` (DevHub `users.user_id` FK).
    - **선택 필드**: `details` (markdown), `external_ref` (외부 시스템 ticket id 등).
    - **Server 가 자동 채우는 필드** (body 의 self-claim 무시 — ADR-0012 §4.1.2 spoofing 방지): `source_system` — 인증된 intake token 의 매핑 값에서 자동 추출.
    - `(source_system, external_ref)` 조합은 UNIQUE — 동일 외부 ticket 의 재수신은 409 또는 idempotent OK 응답.
- **REQ-FR-DREQ-002 (MVP):** DevHub 는 수신 직후 의뢰의 검증(필수 필드 / assignee_user_id 존재)을 수행하고, 성공 시 `pending` 상태로 저장한다. 검증 실패 시 `rejected` 상태 + `rejected_reason="invalid_intake"` 로 저장한다 (audit 보존 목적, 절대 drop 하지 않는다).
- **REQ-FR-DREQ-003 (MVP):** 의뢰의 상태 머신은 `received → pending → in_review → registered | rejected | closed` 로 한정되며, 모든 전이는 `dev_request.*` audit action 으로 기록되어야 한다. (받음/등록됨/거절됨/재오픈됨/닫힘)
- **REQ-FR-DREQ-004 (MVP):** 담당자는 자신의 `assignee_user_id` 와 일치하는 의뢰 목록을 dashboard 에서 조회할 수 있어야 한다. system_admin 은 모든 의뢰를 조회할 수 있어야 한다.
- **REQ-FR-DREQ-005 (MVP):** 담당자는 의뢰를 application 또는 project 로 등록(promote)할 수 있어야 한다. 등록 시 단일 트랜잭션으로 (a) 신규 application/project 생성, (b) DREQ.status → `registered`, (c) DREQ.registered_target_type / registered_target_id 갱신, (d) audit `dev_request.registered` 가 모두 이루어져야 한다. 부분 실패 시 모두 롤백한다.
- **REQ-FR-DREQ-006 (MVP):** 담당자 또는 system_admin 은 의뢰를 reject 할 수 있어야 하며, `rejected_reason` (텍스트) 은 필수다.
- **REQ-FR-DREQ-007 (MVP):** system_admin 은 의뢰의 `assignee_user_id` 를 변경(reassign)할 수 있어야 한다. 변경 이력은 `dev_request.reassigned` audit 으로 기록한다.
- **REQ-FR-DREQ-008 (MVP):** `registered` 또는 `rejected` 상태의 의뢰는 system_admin 이 `closed` 로 전이할 수 있어야 한다. `pending`/`in_review` 상태의 의뢰는 직접 `closed` 로 갈 수 없다 (먼저 reject 후 close).
- **REQ-FR-DREQ-009 (후속):** Platform/Project 의 `origin_dreq_id` 역참조 컬럼 도입 여부는 별도 ADR 에서 결정. 도입 시 nullable FK 로 추가하여 의뢰 없이 직접 생성된 entity 와 공존한다.
- **REQ-FR-DREQ-010 (후속):** 외부 시스템에 의뢰 상태 변경을 callback (webhook) 으로 알리는 기능은 MVP 안정화 후 결정한다.
- **REQ-FR-DREQ-011 (후속):** 의뢰 첨부파일, 댓글, 멘션, 알림, SLA/escalation, AI 자동 분류는 본 도메인의 1차 범위 밖이다.
- **REQ-FR-DREQ-012 (MVP, 확정):** 담당자는 네비게이션 헤더의 실시간 알림 배지와 카운트를 통해 인입된 대기 상태(pending/in_review)의 의뢰 정보를 파악할 수 있어야 한다.
- **REQ-FR-DREQ-013 (MVP, 확정):** 사용자가 헤더 알림에서 개별 의뢰를 클릭하면 의뢰 상세 모달이 열려야 하며, Promote 시 프로젝트 생성 모달이 팝업되고 의뢰의 정보(Key, Name, Description)가 자동으로 프리필되어 연계 생성될 수 있어야 한다.

## 3. 비기능 / 운영 요구사항 (REQ-NFR-DREQ)

- **REQ-NFR-DREQ-001 (MVP, 확정 — [ADR-0012](../../adr/0012-dreq-external-intake-auth.md), sprint `claude/work_260515-g`):** 외부 수신 endpoint 의 인증은 **옵션 A — API 토큰 + IP allowlist** 를 사용한다. token 은 `Authorization: Bearer <token>` 헤더로 전달하며, 외부 시스템별로 별도 발급 + DB 에 hashed 저장 + IP allowlist (CIDR) 로 2차 방어. 운영 정책: 12개월 회전 / 즉시 revoke 가능 / `last_used_at` 모니터링. 옵션 B (HMAC) 와 C (OAuth client_credentials) 는 후속 단계 마이그레이션 경로 (ADR-0012 §3.2).
- **REQ-NFR-DREQ-002 (MVP):** 외부 수신 요청은 idempotency 를 갖는다 — `(source_system, external_ref)` 동일한 재호출은 동일 의뢰 id 로 동일 응답을 반환한다 (혹은 명시적 409 + 기존 id 반환).
- **REQ-NFR-DREQ-003 (MVP):** 모든 상태 전이는 audit_logs 에 기록된다. payload 는 의뢰 id / 전이 from-to / actor / 변경 사유를 포함한다.
- **REQ-NFR-DREQ-004 (MVP):** `details` 필드는 markdown 렌더링 시 XSS 방어를 위해 sanitize 되어야 한다 (frontend 책임). backend 는 raw 저장.
- **REQ-NFR-DREQ-005 (MVP):** DREQ 목록 응답은 페이지네이션(limit/offset 또는 cursor)을 지원하고 기본 limit 50, 최대 100 으로 제한한다.
- **REQ-NFR-DREQ-006 (후속):** 외부 수신 endpoint 의 RPS 한계 / rate limiting 정책은 운영 진입 직전 결정한다.

## 4. 범위 경계 (Out of Scope)

- AI 자동 분류 / 자동 application 매핑 추천.
- 외부 시스템 callback (webhook 송신).
- 의뢰자(requester) 의 DevHub 직접 로그인 + 자기 의뢰 추적 UI.
- 의뢰 첨부 / 댓글 / 멘션 / 알림 / SLA / 자동 escalation.
- repository 단독 등록 (application 또는 project 만 선택).

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/requirements.md` §5.5 에서 본문 그대로 이관. ID(REQ-FR-DREQ-001..013, REQ-NFR-DREQ-001..006) 보존, 신규 발급/삭제 없음. |
