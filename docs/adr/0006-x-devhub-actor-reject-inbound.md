# ADR-0006: inbound `X-Devhub-Actor` 헤더 명시 거부 (400)

- 문서 목적: ADR-0004 §6 의 미해결 항목 ("inbound `X-Devhub-Actor` 명시 거부 채택 여부") 의 후속 결정. silent ignore → 명시 reject 로 전환한다.
- 범위: `authenticateActor` 미들웨어의 1줄 분기 + 회귀 방지 테스트 의도 갱신.
- 대상 독자: Backend / 프론트엔드 개발자, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-j`.
- 관련 문서: [ADR-0004](./0004-x-devhub-actor-removal.md) §6, [`backend-core/internal/httpapi/auth.go`](../../backend-core/internal/httpapi/auth.go).

## 1. 컨텍스트

[ADR-0004](./0004-x-devhub-actor-removal.md) (2026-05-13) 가 `X-Devhub-Actor` 폐기를 ex-post-facto 명문화하면서 §6 미해결 항목으로 남긴 것이 다음이다:

> inbound `X-Devhub-Actor` 헤더의 명시적 거부 (400 / `Bad Request`) 채택 여부 — 후속 ADR 후보. 현재 무시 (no-op) 동작과 결과는 동일하지만, 명시적 거부는 client 쪽 잔재 헤더 사용을 빠르게 발견할 수 있다.

`raven-actions/actionlint` 도입 (ADR-0005) + 4 회귀 방지 negative 테스트로 prod 코드의 헤더 처리 부재는 보장되지만, **client 쪽 잔재 사용** (다른 앱 / 사내 자동화 / 외부 통합) 은 silent ignore 로 surface 되지 않는다.

## 2. 결정 동인

- **빠른 surface**: client 가 잔재 헤더를 보내면 400 응답으로 즉시 인지 — silent 100/200 보다 debug 비용 낮음.
- **회귀 차단 강화**: silent ignore 는 새 contributor 가 "X-Devhub-Actor 가 동작한다" 고 오해할 가능성 존재. 명시 reject 는 오해 불가.
- **비용 낮음**: middleware 1 분기 + 회귀 테스트 의도 갱신. 인프라 추가 0.

## 3. 검토한 옵션

| 옵션 | 설명 | 평가 |
| --- | --- | --- |
| **A. (선택)** 400 + `code=x_devhub_actor_removed` 응답 | `authenticateActor` 미들웨어가 헤더 발견 시 즉시 abort. | client surface + 회귀 차단 강화. 채택. |
| B. silent ignore 유지 (ADR-0004 §5 상태 유지) | 현 상태. | client 잔재가 silent 로 동작 — debug 비용. |
| C. 401 + 헤더 명시 안내 | 인증 실패로 분류. | 의미적 부정확 — 헤더 syntax 거부지 인증 부재가 아니다. |
| D. `Warning` 헤더만 발행 + 200 통과 | 폐기 안내 헤더 추가. | ADR-0004 §3 옵션 C ("deprecation warning header 재도입") 거부 정책과 충돌. 거부. |

## 4. 결정

옵션 A 채택 — middleware 분기:

```go
if c.GetHeader("X-Devhub-Actor") != "" {
    c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
        "status": "rejected",
        "error":  "X-Devhub-Actor header is removed; use Authorization: Bearer ... (see ADR-0004 / ADR-0006)",
        "code":   "x_devhub_actor_removed",
    })
    return
}
```

`authenticateActor` 의 첫 분기로 배치 (publicAPIPaths 확인 이전). dev fallback / Bearer 검증 / role gate 모두 본 분기 이전에 차단된다.

### 4.1 회귀 테스트 의도 갱신

기존 4 negative 테스트 (`audit_test`, `auth_test`, `commands_test`, `me_test`) 의 검증은 "header 가 ignored, response 200/401" 패턴이었다. 본 ADR 채택으로 다음과 같이 갱신:

- `TestGetMeRejectsXDevhubActorHeader`: 401 → **400 + `code=x_devhub_actor_removed`**.
- `TestXDevhubActorRejectedWhenDevFallbackOff`: 401 → **400**.
- `TestXDevhubActorRejectedEvenWhenDevFallbackOn`: 202 → **400 + command 미작성**.
- `TestCreateUserWritesAuditLogWithSystemFallback`: 헤더 의존 split — `TestCreateUserRejectsXDevhubActorHeader` 분리.
- `TestApproveServiceActionCommandMarksApprovalAndAudits` / `TestRejectServiceActionCommandMarksRejectedAndAudits`: 본 PR 에서 X-Devhub-Actor / X-Devhub-Role 헤더 제거. dev fallback 으로 actor=system 처리 (그 자체는 ADR-0006 의 의도와 무관).

## 5. 결과 (Consequences)

### 긍정적

- client 쪽 잔재 헤더 사용이 응답 코드로 즉시 surface.
- 회귀 테스트 의도가 "ignore" 에서 "reject" 로 명시화되어 새 contributor 가 정확히 동작 이해.

### 부정적 / 트레이드오프

- 본 ADR 채택 시점에 client 가 잔재 헤더를 보내고 있으면 break. **DevHub frontend 는 본 ADR 채택 시점에 X-Devhub-Actor 를 보내지 않는다** (Bearer token 으로 통일됨, sprint `claude/work_260513-g` 의 §11.5 매핑 + `IMPL-auth-02` actor context). 외부 client / 사내 자동화의 잔재 사용은 본 ADR 의 surface 의도와 정합 — break 가 곧 surface.

### 비변경 사항

- 다른 헤더 (`X-Devhub-Role`, `X-Devhub-Auth`) 는 본 ADR 범위 밖.
- `X-Devhub-Actor-Deprecated` 응답 헤더는 발행되지 않음 (ADR-0004 §3 옵션 C 와 동일 정책).

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| `X-Devhub-Role` query / header 도 명시 거부? | 별도 ADR 후보. `X-Devhub-Role` 은 dev fallback realtime ws 흐름에서 여전히 사용 — 그 spec 자체가 deprecated 인지부터 결정 필요. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-j`). ADR-0004 §6 미해결 항목 closing. |
