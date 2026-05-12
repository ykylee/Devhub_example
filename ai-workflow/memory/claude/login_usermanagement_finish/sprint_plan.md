# Sprint Plan — login / 사용자 관리 1차 완성

- 문서 목적: `claude/login_usermanagement_finish` 브랜치의 sprint 목표·1차 완성 정의·진입 게이트·PR 단위·검증을 단일 진입점에서 기술한다.
- 범위: 본 sprint 의 in-scope / out-of-scope, PR 단위, DoD, 검증 명령, 위험 평가, 후속 sprint 인계
- 대상 독자: 본 브랜치 작업자, 후속 에이전트, 리뷰어
- 브랜치: `claude/login_usermanagement_finish` (분기 `main` @ `29a90bd`)
- 최종 수정일: 2026-05-12
- 상태: draft (사용자 확인 대기)
- 관련 문서: [세션 인계](./session_handoff.md), [백로그](./work_backlog.md), [상태](./state.json), [통합 로드맵](../../../../docs/development_roadmap.md), [프론트 로드맵](../../../../docs/frontend_development_roadmap.md), [백엔드 로드맵](../../backend_development_roadmap.md)

---

## 1. 목표와 1차 완성 정의

### 1.1 목표

DevHub 의 **로그인 / 로그아웃 / 사용자 관리** 트랙을 운영 환경 진입 직전에 통과할 수준으로 1차 완성한다. 핵심 흐름의 골격(M2)은 이미 완료된 상태이므로, 본 sprint 는 잔여 *결함·결손·정합 누락* 을 닫는다.

### 1.2 1차 완성 정의 (Definition of Done — sprint level)

다음을 모두 만족할 때 1차 완성으로 간주한다.

1. **사용자 가시 UX placeholder/결함 zero** — 화면 위에 노출되지만 동작하지 않는 컨트롤(검색·필터 등), 사용자 혼란을 일으키는 라벨/안내 부재가 남아있지 않는다.
2. **인증 행위의 audit 정합** — Kratos self-service 흐름(특히 비밀번호 변경)이 `audit_logs` 에 누락 없이 기록된다.
3. **모든 변경의 검증 통과** — frontend `npm run build`, backend `go test ./...`, `pytest ai-workflow/tests/check_docs.py` 가 PASS.
4. **로드맵 정합** — 본 sprint 의 결과가 통합 로드맵·프론트 로드맵·백엔드 로드맵에 반영되어 다음 sprint 의 진입 기준이 모호하지 않다.

### 1.3 1차 완성 진입 게이트 체크리스트

- [x] PR-UX1 머지 — `/admin/settings/users` 검색 필터 동작 (commit `0e27407`)
- [x] PR-UX2 머지 — `/account` Kratos privileged session 안내 노출 (commit `0e27407`)
- [x] PR-UX3 머지 — Header Switch View 의 한계 안내 노출 (commit `0e27407`)
- [x] PR-M2-AUDIT 머지 — Kratos webhook handler + 7 unit test PASS (commit `ead9f73`)
- [x] 로드맵 3종(통합/프론트/백엔드) 갱신 PR 머지 (commit `a9e2e3a`)
- [x] e2e 30 TC 작성 (commit `712b23e`) — 35 tests discovery PASS, tsc clean
- [ ] **실 e2e 실행 PASS** — `cd frontend && npm run e2e` (사용자 환경 의존: 5-process stack + DEVHUB_KRATOS_WEBHOOK_TOKEN)
- [ ] 리뷰 코멘트 보강 (필요 시)
- [ ] sprint close commit (state/handoff/backlog 종료 표기 + 후속 인계 6건)

---

## 2. In-Scope

결정 사항 (2026-05-12 확정): PR-UX 는 **묶음 1 PR**, webhook 인증은 **shared secret**, hook 범위는 **최소 `settings/password/after`**.

| ID | 제목 | 트랙 | 규모 | 비고 |
| --- | --- | --- | --- | --- |
| **PR-UX** (묶음) | `/admin/settings/users` 필터 + `/account` Kratos 안내 + Header Switch View 한계 안내 | F | S | 3 파일 ~50 LOC. UX-1/2/3 을 한 PR 에 묶음. |
| PR-M2-AUDIT | Kratos `settings/password/after` webhook → `audit_logs` | B | M | shared secret 인증, password 변경 audit 만 1차. |
| PR-DOCS | 로드맵 3종 + sprint_plan 정합 | X | S | 본 sprint 진입 직후 1차 갱신, 종료 시 close 갱신. |

## 3. Out-of-Scope (명시적 분리)

| 항목 | 분리 사유 | 인계 |
| --- | --- | --- |
| Hydra JWKS / introspection verifier 실구현 | 보안 영향 큰 변경, 별도 검증 필요. 현 introspection 안정. | 별도 sprint (`m2_followups.hydra_jwks`) |
| Sign Up (셀프 가입) | 인사 DB 스키마 결정 필요 (M3). | M3 |
| MFA / Two-Factor | 별도 ADR 필요 (M4). | M4 |
| caller-supplied `X-Request-ID` content validation | PR-D 후속, audit hygiene 트랙. | `pr_d_followups` |
| `ctx` 표준 request_id 전파 | PR-D 후속. | `pr_d_followups` |
| `writeRBACServerError → writeServerError` 통합 | 단순 리팩터, M1 cleanup. | `pr_d_followups` |
| PR-T5 CI 도입 | 사용자 보류 결정(2026-05-12). | 미정 |
| M4 (WebSocket UI, AI Gardener gRPC, Gitea Hourly Pull) | 본 sprint 범위 외 | M4 sprint |

---

## 4. PR 단위 세부

### 4.1 PR-UX1 — `/admin/settings/users` SearchInput 실 필터링

- **파일**: `frontend/app/(dashboard)/admin/settings/users/page.tsx`
- **변경 요지**:
  - `query` state 추가, `<input>` 에 `value`/`onChange` 바인딩
  - `MemberTable` 로 넘기는 `members` 를 `query` 로 필터 (login / email / role 부분일치, case-insensitive)
  - Filter 버튼: 본 sprint 에서는 비활성화(`disabled` + tooltip "filters coming soon") — 또는 제거. PR 작성 시 확정.
- **DoD**:
  - `cd frontend && npm run build` PASS
  - 검색어 입력 시 행 즉시 필터링 (수동 검증)
  - dead 버튼이 사용자 혼란을 일으키지 않는 상태
- **위험**: 낮음. 클라이언트 필터만 추가하므로 서버 contract 변경 없음.

### 4.2 PR-UX2 — `/account` Kratos privileged session 안내

- **파일**: `frontend/app/(dashboard)/account/page.tsx`
- **변경 요지**:
  - Current Password 라벨 아래 1줄 안내 (예: "최근 로그인 직후가 아니면 현재 비밀번호가 필요합니다. 세션이 만료되면 Sign In Again 으로 재인증해주세요.")
  - REAUTH_REQUIRED 경로에 표시되는 메시지/버튼 텍스트 검토 (사용자 행동을 1step 안내)
- **DoD**:
  - `cd frontend && npm run build` PASS
  - 라벨 옆/아래 안내 노출, REAUTH 발생 시 Sign In Again 버튼이 명확
- **위험**: 매우 낮음. 텍스트 + 보조 안내만.

### 4.3 PR-UX3 — Header Switch View 한계 안내

- **파일**: `frontend/components/layout/Header.tsx`
- **변경 요지**:
  - dropdown 의 "Switch View" 라벨 아래 1줄 보조 텍스트 (예: "메뉴 미리보기 — 실제 권한은 서버 `actor.role` 기준")
  - 또는 info 아이콘 + tooltip
- **DoD**:
  - `cd frontend && npm run build` PASS
  - 사용자가 메뉴 미리보기와 실권한 변경을 혼동하지 않음
- **위험**: 매우 낮음.

### 4.4 PR-M2-AUDIT — Kratos self-service webhook → `audit_logs`

- **파일 (예정)**:
  - `backend-core/internal/httpapi/kratos_webhook_handler.go` (신규) — payload 파싱, audit_logs INSERT
  - `backend-core/internal/httpapi/router.go` — `/api/v1/kratos/hook/{event}` 라우트 + auth middleware
  - `backend-core/internal/audit/*` — 필요 시 `source_type=kratos` enum 보강 (현 enum 검토 필요)
  - `infra/idp/kratos.yaml` — `selfservice.flows.*.after.hooks` 추가
  - `docs/setup/test-server-deployment.md` — webhook 운영 안내 §보강
- **결정 (2026-05-12 확정)**:
  - 등록할 hook: 최소 `settings/password/after` 만. 나머지(`login/after`, `registration/after`) 는 후속 sprint.
  - webhook 인증: shared secret `Authorization: Bearer $DEVHUB_KRATOS_WEBHOOK_TOKEN`. env 누락/불일치 시 401.
- **DoD**:
  - `cd backend-core && go test ./...` PASS (handler 단위 테스트 1+)
  - Kratos password 변경 → `audit_logs` 에 source_type=kratos 1행 INSERT (수동 검증)
  - secret 미설정 또는 잘못된 secret 시 401 응답 (handler 단위 테스트 포함)
- **위험**:
  - Kratos 의 hook payload 스키마는 self-service flow 마다 다름 → 본 sprint 는 password 변경 1종만 enforced, 나머지는 best-effort.
  - 운영 환경에서 webhook secret 분실 시 audit 누락. 운영 가이드에 secret 운영 절차 명시 필요.

### 4.5 PR-DOCS — 로드맵 3종 + sprint_plan 정합

- **파일**:
  - `docs/development_roadmap.md` (§3 M2 + §7 변경 이력)
  - `docs/frontend_development_roadmap.md` (§2 Phase 6/6.1 상태 + §7 다음 작업 큐)
  - `ai-workflow/memory/backend_development_roadmap.md` (§5 P0 M2 + §6 작업 큐)
  - 본 sprint_plan 자체
- **DoD**: `pytest ai-workflow/tests/check_docs.py` PASS

---

## 5. 검증 명령 (한 곳에 정리)

```
# Frontend
cd frontend && npm run build

# Backend
cd backend-core && go test ./...

# 문서
PYTHONPATH=ai-workflow python ai-workflow/tests/check_docs.py
```

E2E 는 본 sprint 범위에서 필수가 아니지만, PR-M2-AUDIT 머지 후 1회 수동 확인 권장:

```
cd frontend && npm run e2e
```

---

## 6. 위험 평가

| 위험 | 영향 | 완화 |
| --- | --- | --- |
| Kratos hook payload 스키마가 버전마다 다름 | PR-M2-AUDIT 의 미래 호환성 | password 변경 1종 우선, payload 검증 fail-safe (모르는 필드는 무시) |
| webhook secret 누설 | audit 위·변조 위험 | env 관리 + `test-server-deployment.md` §보강 + secret rotation 절차 1줄 |
| UX 텍스트 한국어/영어 혼용 | i18n 부재 | 본 sprint 는 영어 UI 유지(현 UI 와 정합), 한국어화는 별도 sprint |
| Filter 버튼 제거 시 UX 변화 | 사용자 시각 혼란 | 비활성화 + tooltip 으로 가시성 유지 권장 |

---

## 7. 일정 가이드 (제안)

| 일차 | 작업 |
| --- | --- |
| D0 (오늘) | sprint_plan + 로드맵 3종 머지 (PR-DOCS) |
| D1 | PR-UX 묶음 PR 작성 + 머지 |
| D2 | PR-M2-AUDIT 작성 + 머지, sprint close |

위는 단순 제안이며, 사용자 일정에 맞춰 조정.

---

## 8. 후속 sprint 인계 (sprint close 시점에 갱신)

(sprint 종료 후 채움)

- [ ] Hydra JWKS verifier 실구현 (별도 sprint)
- [ ] caller-supplied X-Request-ID validation (PR-D follow-up)
- [ ] M3 Sign Up 흐름 진입 결정
