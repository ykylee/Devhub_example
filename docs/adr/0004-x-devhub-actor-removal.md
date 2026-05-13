# ADR-0004: `X-Devhub-Actor` 헤더 폐기 완료 선언

- 문서 목적: M0 SEC-4 시점에 prod 코드에서 이미 제거된 `X-Devhub-Actor` fallback 헤더에 대해 ADR-0001 §8 #4 가 보류한 "완전 제거 trigger" 가 충족됐음을 ex-post-facto 명문화한다.
- 범위: 인증 actor 도출 경로의 legacy fallback 헤더 폐기 결정. Bearer token verifier 도입과 audit/RBAC 흐름은 ADR-0001 / ADR-0002 가 다룬다.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent, 외부 API consumer, 후속 인증 phase 의사결정자.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-h`.
- 관련 문서: [ADR-0001 IdP 선정](./0001-idp-selection.md) §8 #4, [architecture.md §6.2](../architecture.md#62-사용자user--계정account-도메인-분리), [backend_api_contract.md §11.3](../backend_api_contract.md#113-go-core-bearer-token-경계-api-19), [추적성 매트릭스 §5.3](../traceability/report.md#53-문서-↔-코드-불일치).

## 1. 컨텍스트

`X-Devhub-Actor` 는 M0 이전 phase 의 fallback 헤더로, Bearer token verifier 가 도입되기 전 actor 도출 경로를 제공하기 위한 임시 장치였다. 도입 당시 [ADR-0001 §8 #4](./0001-idp-selection.md) 가 폐기 시점을 다음과 같이 결정했다.

> **`X-Devhub-Actor` 헤더 폐기 시점**: Phase 13 P1 7단계(Bearer token 검증 미들웨어) 도입 시점에 `X-Devhub-Actor` 폴백은 유지하되 deprecation warning 로그를 남긴다. 모든 backend 핸들러가 token 경로로 actor 도출하도록 전환됐는지 검증된 뒤 별도 phase 에서 완전 제거.

본 ADR 채택 시점 (2026-05-13) 의 실제 상태는 다음과 같다.

1. **prod 코드**: `backend-core/internal/httpapi/auth.go::authenticateActor` 가 `X-Devhub-Actor` 헤더를 어떤 분기에서도 읽거나 actor 로 채택하지 않는다 — SEC-4 (M0, 2026-05-07~08) 의 resolution 으로 처리 코드 자체가 제거됐다.
2. **회귀 방지 테스트**: `me_test.go`, `audit_test.go`, `auth_test.go`, `commands_test.go` 가 "`X-Devhub-Actor` must be ignored" 를 명시 검증 (응답 401 / actor=`system` / `X-Devhub-Actor-Deprecated` 헤더 부재).
3. **M1 PR-D**: `BearerTokenVerifier` interface + `AuthenticatedActor` context propagation + `authenticateActor` middleware 로 actor 도출 경로가 표준화됐다 (`backend_api_contract.md` §11.3 = API-19, IMPL-auth-01/02 — 매트릭스 §2.2 / §2.4 의 Auth API / IMPL-auth-XX 서브 표).
4. **문서 잔재**: `architecture.md` §6.2 본문은 여전히 "폐기 예정 폴백으로 유지" 라고 적혀 있고, `me.go::getMe` 의 doc comment 에도 `X-Devhub-Actor` 언급이 남아 있다.

즉 ADR-0001 §8 #4 의 trigger ("모든 backend handler 가 token 경로로 actor 도출되도록 전환") 가 SEC-4 시점에 이미 충족됐고, prod 코드도 즉시 제거 path 를 선택해 별도 마이그레이션 phase 를 도입하지 않았다. 본 ADR 은 그 사실의 ex-post-facto 기록이다.

## 2. 결정 동인

- **사실 정합**: 코드와 문서의 drift 가 reader 혼동을 유발 (`architecture.md` 가 "유지" 라고 적었는데 코드는 처리 안 함).
- **추적성**: 추적성 매트릭스 §5.3 의 "X-Devhub-Actor 완전 제거 trigger" open 항목이 코드 현실과 무관하게 open 으로 남아 있음.
- **회귀 방지 보존**: 신규 코드가 fallback 헤더를 다시 도입하지 않도록 명시적 결정 기록이 필요.
- **마이그레이션 phase 단순화**: 별도 phase 도입 없이 SEC-4 가 즉시 제거 path 를 선택한 사실 자체를 인정 — `X-Devhub-Actor-Deprecated` warning 헤더 단계를 건너뛴 채 완전 제거.

## 3. 검토한 옵션

| 옵션 | 설명 | 평가 |
| --- | --- | --- |
| **A. (선택)** 본 ADR 으로 폐기 완료 선언 | SEC-4 시점에 trigger 가 충족됐음을 명문화하고, 문서/주석 잔재만 정리. 동작 변경 0. | 사실 정합. 추적성 가시화. 추가 비용 0. |
| B. ADR-0001 §8 #4 만 인라인 갱신 | 새 ADR 없이 §8 의 결정만 갱신. | §8 정책상 §8 항목은 인라인 갱신 허용. 그러나 ex-post-facto 결정의 가시성이 매트릭스 §4 ADR 인덱스에 잡히지 않음. |
| C. deprecation warning header 재도입 | M0 이전 정책으로 회귀해 `X-Devhub-Actor-Deprecated: true` 헤더를 outbound 로 발행. | 사용 안 되는 헤더 처리 코드의 부활. 회귀 방지 테스트와도 충돌. 채택 거부. |
| D. 헤더 자체 차단 (`Vary` / 400) | inbound `X-Devhub-Actor` 을 명시적으로 거부. | 현재 무시 동작과 동일 결과. 신규 거부 분기 추가는 본 ADR 의 의도와 불일치 (단순 폐기). 후속 ADR 후보로 미해결 §6. |

## 4. 결정

**옵션 A 채택** — `X-Devhub-Actor` 폐기 완료 선언.

- ADR-0001 §8 #4 의 "별도 phase 에서 완전 제거" trigger 는 SEC-4 (M0, 2026-05-07~08) 시점에 prod 코드 제거로 이미 충족됐다. 별도 phase 를 거치지 않고 즉시 제거 path 를 채택한 것을 본 ADR 이 사후 명문화한다.
- 회귀 방지 테스트 (4 파일) 는 그대로 유지한다 — security regression 차단.
- 문서/주석 잔재 (architecture.md line 174, ADR-0001 §8 #4 인라인 갱신, me.go line 16 주석) 는 본 sprint 가 정리한다.

### 4.1 채택 후 인보크 패턴

- **새 코드**: `X-Devhub-Actor` 헤더에 대한 어떤 처리 (인보크, 로깅, 응답 헤더 발행) 도 추가하지 않는다.
- **테스트**: 회귀 방지 negative 테스트만 허용. 새 테스트가 이 헤더를 의미 있게 set 하면 안 된다.
- **문서**: `X-Devhub-Actor` 는 history 노트 (ADR-0004 link) 외에 어떤 spec 도 갖지 않는다.

## 5. 결과 (Consequences)

### 긍정적

- 추적성 매트릭스 §5.3 의 "X-Devhub-Actor 완전 제거 trigger" open 항목을 closed 처리할 수 있다.
- `architecture.md` 의 "유지 예정" 표기가 사실과 일치한다.
- 새 contributor 가 ADR-0001 §8 #4 를 읽었을 때 fallback 헤더의 현재 상태 (= 제거 완료) 를 즉시 알 수 있다.

### 부정적 / 트레이드오프

- 별도 deprecation phase 없이 SEC-4 가 즉시 제거 path 를 선택한 이력이 시계열 측면에서 "ADR-0001 §8 #4 → SEC-4 → ADR-0004" 의 3 단계로 분산된다. 본 ADR 의 §1 이 그 시계열을 한 곳에 모은다.

### 비변경 사항

- 코드 동작 0 변경 (auth.go / 회귀 방지 테스트 그대로).
- API contract (`backend_api_contract.md` §11) 변경 없음 — §11 본문은 이미 X-Devhub-Actor 를 spec 항목으로 갖지 않는다.

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| inbound `X-Devhub-Actor` 헤더의 명시적 거부 (400 / `Bad Request`) 채택 여부 | 후속 ADR 후보. 현재 무시 (no-op) 동작과 결과는 동일하지만, 명시적 거부는 client 쪽 잔재 헤더 사용을 빠르게 발견할 수 있다. 본 ADR 채택 시점에는 미결정. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-h`). X-Devhub-Actor 폐기 완료 선언 + ADR-0001 §8 #4 trigger 충족 사실 명문화. |
