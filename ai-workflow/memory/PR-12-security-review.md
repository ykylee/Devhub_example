# PR #12 보안 리뷰 리포트

- 문서 목적: Claude Code `/security-review` 자동 보안 리뷰 결과를 PR #12 (`test/backend-integration` → `main`) 기준으로 정리한다.
- 범위: 본 PR 이 *새로 도입* 한 보안 위험만 다룬다. 기존 코드의 위험은 다루지 않는다.
- 대상 독자: PR 작성자, 리뷰어, 보안 담당자, 다음 세션 담당자
- 상태: stable
- 최종 수정일: 2026-05-08
- 관련 문서: [PR-12 액션 플랜](./PR-12-review-actions.md), [claude harness backlog](./claude/test/backend-integration/backlog/2026-05-08.md)

## 1. 실행 정보

| 항목 | 값 |
| --- | --- |
| 도구 | Claude Code `/security-review` |
| 브랜치 / HEAD | `test/backend-integration` / `fe593e7` |
| 비교 base | `origin/main` |
| 변경 규모 | 196 files (분할 후) — 이 중 보안 관련 영역(backend-core auth/audit/organization/commands/router/main, frontend AuthGuard/login, 일부 service) 만 검토 |
| 검토 단계 | 1차 식별(서브에이전트) → 후보별 FP 필터링(병렬 서브에이전트) → confidence ≥ 8 만 채택 |

## 2. 결과 요약

| 후보 | 분류 | 위치 | 처리 |
| --- | --- | --- | --- |
| 후보 A | AuthN/AuthZ — 무인증 통과 + 권한 미적용 + Actor 스푸핑 | `auth.go`, `main.go`, `router.go`, `commands.go` | **KEEP, severity High, confidence 10** |
| 후보 B | AuthN/AuthZ — Frontend mock 로그인 | `frontend/app/login/page.tsx`, `AuthGuard.tsx` | REJECT, confidence 9 (precedent: 클라이언트 측 권한 미흡 자체는 취약점이 아니다) |
| 후보 C | AuthN/AuthZ — `X-Devhub-Actor` 헤더 스푸핑 | `commands.go:271-287` | REJECT, confidence 9 (verifier 주입 후 fallback 도달 불가, 위조 값은 audit/display 컬럼만 채우고 권한 결정에 미사용 → log/audit 무결성은 정책상 single-issue 취약점 아님) |

→ **최종 채택 1건 (High)**.

## 3. 최종 채택 — Vuln 1: AuthN/AuthZ 통합 결함

**위치**

- `backend-core/internal/httpapi/auth.go:24-44`
- `backend-core/main.go:47-63`
- `backend-core/internal/httpapi/router.go:65-99`
- `backend-core/internal/httpapi/commands.go:271-287`

**심각도 / 신뢰도**: High / 1.0

### 3.1 결함 묘사

1. **무인증 통과 (auth.go:26-29)** — `authenticateActor` 미들웨어가 `Authorization` 헤더가 *비어 있으면* 그대로 `c.Next()` 호출. 즉 헤더만 안 붙이면 검사 자체를 건너뜀.
2. **Bearer verifier 미주입 (main.go:47-63)** — `RouterConfig` 생성 시 `BearerTokenVerifier` 필드를 설정하지 않음. 따라서 **Bearer 토큰이 붙어 있어도** `auth.go:41-44` 의 dev fallback 분기로 떨어져 무검증 통과. 본 PR 의 `main.go:46` SEC-2 주석이 이 사실을 인정함.
3. **권한(role) 미적용** — `auth.go:73` 에서 `devhub_actor_role` 을 gin context 에 set 하지만, **어느 핸들러도 이 값을 읽지 않음**. `defaultRBACPolicy()` 의 role × resource matrix 가 정의되어 있지만 실제 라우트 결정에 연결되지 않음. 즉 verifier 가 향후 주입돼도, role 차등이 동작하지 않음.
4. **Actor 스푸핑 (commands.go:271-287)** — 인증 컨텍스트가 비어 있을 때 `X-Devhub-Actor` 헤더 값을 그대로 `actor_login` 으로 사용. 무인증 상태에서 audit/command 의 행위자 신원을 임의로 위조 가능.

미들웨어는 `/api/v1` 전체 그룹(`router.go:65-66`) 에 마운트 — 결함 영향이 PR #12 가 추가한 모든 mutating 라우트에 적용된다.

### 3.2 영향 범위

다음 라우트가 무인증 + 무권한 호출 가능:

- `PATCH /api/v1/users/:user_id`, `DELETE /api/v1/users/:user_id`
- 조직 단위 CRUD: `POST/PATCH/DELETE /api/v1/organization/units...`, `PUT /api/v1/organization/units/:unit_id/members`
- `POST /api/v1/admin/service-actions`
- `POST /api/v1/risks/:risk_id/mitigations`
- `GET /api/v1/audit-logs` (감사 이력 노출)
- `GET /api/v1/realtime/ws` (실시간 채널)

저장소 정책상 native 배포에 별도 게이트웨이가 없음 (BLK-1 결정 + `feedback_no_docker.md` 정책).

### 3.3 익스플로잇 시나리오

1. 공격자가 backend 에 네트워크 도달.
2. `PATCH /api/v1/users/<victim>` 본문 `{"role":"system_admin"}` 전송 — Authorization 헤더 없음. `updateUser` 가 `validAppRoles` 검사를 통과시키고 그대로 저장 → 임의 계정을 System Admin 으로 승격.
3. 또는 `POST /api/v1/admin/service-actions` 본문 `{"service_id":"prod-db","action_type":"shutdown","reason":"x","force":true,"dry_run":false}` 전송 → command 가 즉시 큐잉.
4. `GET /api/v1/audit-logs` 로 감사 이력 전체 열람. 동시에 `X-Devhub-Actor: <임의 사용자>` 를 붙여 위조된 actor 로 행위 기록을 만들 수 있음.

### 3.4 권고 조치

1. **fail-fast 가드** — `main.go` 또는 config layer 에서 `BearerTokenVerifier == nil` 이면 prod 환경에서 startup 거부. dev 모드는 명시적 환경변수(예: `DEVHUB_ENV=dev` 또는 `DEVHUB_AUTH_DEV_FALLBACK=1`) 가 켜져 있을 때만 허용.
2. **무인증 통과 분기 제거** — `auth.go` 의 *empty Authorization → c.Next()* 분기를 제거하거나 prod 에서는 401 반환. 인증이 필요 없는 라우트는 `/healthz`, `/api/v1/snapshot` 등으로 명시적 화이트리스트.
3. **`X-Devhub-Actor` fallback 제거** — `requestActor` 의 헤더 fallback 을 dev 모드에서만 허용하고 prod 에서는 무시. `commands.go` 의 `Warning` 응답 헤더만으로는 감사/명령 컬럼 위조를 막지 못함.
4. **Role 강제 미들웨어** — `devhub_actor_role` 을 읽어 `defaultRBACPolicy()` matrix 와 매칭하는 라우트별 가드 추가. 최소한 다음 라우트:
   - `/admin/*` → `system_admin`
   - 사용자 PATCH/DELETE → `system_admin`
   - `GET /audit-logs` → `manager` 이상
   - 조직 단위 CRUD → `system_admin`
   - command 생성 → `manager` 이상
5. **테스트 보강** — `auth_test.go` 에 prod 가드(env unset + verifier nil → 거부) 케이스 + role 가드 통합 테스트 케이스 추가.

## 4. PR-12 액션 플랜 매핑

| 본 리포트 항목 | PR-12-review-actions.md 매핑 | 추가 처리 필요 |
| --- | --- | --- |
| 무인증 통과 (3.1-1, 3.1-2) | SEC-2 (verifier 미주입) 와 동일 사안 | 본 PR 에 코드 마커는 부착됨. backlog DoD 에 *empty Authorization 분기 제거* + *fail-fast* 항목 추가 필요 |
| Role 미적용 (3.1-3) | **신규** — 기존 SEC 항목에 미반영 | SEC-3 신설 후 backlog 등록 권장 |
| Actor 스푸핑 (3.1-4) | **신규** — 기존 SEC 항목에 미반영 | SEC-4 신설 또는 SEC-2 의 sub-item. backlog 의 DoD 에 `X-Devhub-Actor` fallback 제거 명시 필요 |
| Frontend mock 로그인 (후보 B, REJECT) | SEC-1 그대로 유효 | 정책상 단독 취약점은 아니지만 backlog 추적은 유지 |

## 5. 머지 영향 평가

- 본 결함들은 **PR 머지 블로커가 아니라 다음 PR 의 블로커**. 즉 본 PR 은 인증 인프라 *틀* 만 도입했고 verifier 주입은 follow-up. 그러나 다음 PR 이 가시 가능한 인증 의존 기능(예: 권한 의존 UI, 외부 노출 endpoint) 을 추가할 때까지는 **본 결함이 모두 해소돼야 한다**.
- 본 PR 의 `// SECURITY (SEC-1|SEC-2)` 마커가 코드에 부착되어 있어 grep 추적 가능. SEC-3/SEC-4 추가 시 동일 패턴 권장.

## 6. 다음 단계 (제안)

1. SEC-3 (role 미적용), SEC-4 (Actor 스푸핑) 항목을 `claude/test/backend-integration/backlog/2026-05-08.md` 에 신설.
2. SEC-2 의 DoD 에 *empty Authorization 분기 제거* + *fail-fast* 추가.
3. 본 리포트를 PR #12 코멘트로 등록 (보안 검증 결과 첨부).
4. SEC-1 ~ SEC-4 를 묶어 다음 PR (인증/권한 트랙) 에서 일괄 해소.
