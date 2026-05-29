# API 공통 규약 (envelope + enum)

- 문서 목적: 모든 도메인 API 가 공유하는 응답 envelope, role wire format, 공통 상태값 enum 을 단일 source-of-truth 로 기록한다.
- 범위: REST envelope, WebSocket envelope 와 공통 enum. endpoint별 본문은 각 도메인 sub-document (`docs/domain/<도메인>/api.md`) 참조.
- 대상 독자: backend / 프론트엔드 / DevOps, AI agent, 외부 API consumer, QA.
- 상태: accepted
- 최종 수정일: 2026-05-29 (Phase 3 split, master `docs/backend_api_contract.md` §1 + §2 본문 이관)
- 관련 문서: [master backend_api_contract.md](../backend_api_contract.md), [governance/document-standards.md](../governance/document-standards.md)

## 1. 공통 응답 원칙

- 성공 응답은 `status`, `data`, `meta`를 기본 envelope로 사용한다.
- 단일 command성 endpoint는 `status`와 생성/처리 결과 key를 함께 반환할 수 있다.
- 실패 응답은 `status`, `error`를 반환한다.
- 시간 값은 ISO 8601/RFC3339 형식의 UTC timestamp를 사용한다.
- API role wire format은 `developer`, `manager`, `system_admin`을 사용하고 UI 표시명과 분리한다.

## 2. 공통 enum 및 상태 값

### 2.1 Role wire format

```text
developer
manager
system_admin
```

### 2.2 공통 상태 값

```text
ServiceStatus = stable | warning | degraded | down
RiskImpact = low | medium | high | critical
RiskStatus = detected | investigation | action_required | mitigated | dismissed
CommandStatus = pending | running | succeeded | failed | rejected | cancelled
WebhookEventStatus = received | validated | processed | failed | ignored
AccountStatus = active | disabled | locked | password_reset_required
```

Webhook event는 signature 검증과 raw 저장이 끝나면 `validated`가 되며, 정규화가 성공하면 `processed`, 재처리 가능한 오류는 `failed`, 지원하지 않거나 처리 대상이 아닌 이벤트는 `ignored`로 전환한다.

## 3. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-29 | Phase 3 split — master `docs/backend_api_contract.md` §1 (공통 응답 원칙) + §2 (공통 enum) 본문 그대로 이관. master file 은 master index 로 전환되면서 본 문서를 cross-cutting 진입점으로 사용한다. |
