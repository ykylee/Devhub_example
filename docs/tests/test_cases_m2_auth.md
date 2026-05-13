# Test Cases — M2 1차 완성 검증 (기능별 분류)

- 문서 목적: `claude/login_usermanagement_finish` sprint 의 1차 완성 정의를 e2e/단위 테스트로 검증하기 위한 TC 리스트. **서비스 기능별로 분류**하여 spec 파일을 따로 둔다.
- 범위: PR-UX1/2/3 + PR-M2-AUDIT 변경에 대한 e2e 시나리오 + 기존 회귀 sanity 매핑.
- 대상 독자: e2e 작성자, 리뷰어, sprint close 검증 담당.
- 최종 수정일: 2026-05-12
- 상태: draft (사용자 확인 대기)
- 관련 문서: [sprint plan](./sprint_plan.md), [세션 인계](./session_handoff.md)

---

## 1. 1차 완성 정의 → TC 매핑

| 1차 완성 정의 | 검증 방법 | TC ID |
| --- | --- | --- |
| (1) 사용자 가시 UX placeholder/결함 zero | Playwright e2e | TC-USR-01..06, TC-USR-CRUD-01..03, TC-ACC-01..03, TC-NAV-01..03 |
| (2) Kratos self-service 흐름의 audit 정합 | Playwright e2e + Go integration | TC-AUD-01..02 + (BE) 7 케이스 (PR-M2-AUDIT 머지본) |
| (3) 모든 변경의 검증 통과 | CI/로컬 명령 | (자동) frontend `npm run build`, backend `go test ./...`, `pytest ai-workflow/tests/check_docs.py` |
| (4) 로드맵 3종 정합 갱신 | 머지 후 점검 | (수동) PR-DOCS 머지 후 main 에서 §3 M2 확인 |
| (5) 운영 진입 직전 게이트 — 인증/권한 핵심 흐름 회귀 | Playwright e2e | TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01, TC-AUTH-SIGNOUT-REDIR-01, TC-RBAC-SUB-01, TC-RBAC-MGR-01, TC-SIGNUP-01..04 |

---

## 2. 환경 전제

- 5-process native stack 가동 (PostgreSQL + Hydra + Kratos + backend-core + frontend), `docs/setup/e2e-test-guide.md` 절차.
- backend-core 와 Kratos 양쪽에 **`DEVHUB_KRATOS_WEBHOOK_TOKEN=<같은 값>`** export 후 Kratos 재기동 (PR-M2-AUDIT 변경 반영).
- Kratos 설정파일 `infra/idp/kratos.yaml` 의 `settings.after.password.hooks` 가 새 jsonnet (`infra/idp/kratos_webhooks/settings_password_after.jsonnet`) 으로 작동.
- e2e seed: globalSetup 이 alice/bob/charlie 3명 idempotent 시드 + 비밀번호 force-reset.
- 실행: `cd frontend && npm run e2e`.

---

## 3. 기능별 TC 분류

본 sprint 의 1차 완성 검증은 **4개 서비스 기능** 으로 분류한다. 각 기능에 대응되는 spec 파일을 하나씩 둔다.

| # | 기능 | spec 파일 | TC IDs | 기존 spec 와의 관계 |
| --- | --- | --- | --- | --- |
| F1 | **사용자 관리** (`/admin/settings/users`) | `admin-users-search.spec.ts` + `admin-users-crud.spec.ts` (신규 2) | TC-USR-01..06, TC-USR-CRUD-01..03 | 검색 + CRUD UI smoke. 기존 spec 없음. |
| F2 | **내 계정 / 비밀번호** (`/account`) | `account.spec.ts` (신규, `password-change.spec.ts` 흡수) | TC-ACC-01..03 | 기존 `password-change.spec.ts` 삭제 후 본 spec 으로 통합 — UX 안내 + 변경 round-trip + 클라이언트 mismatch 검증. |
| F3 | **헤더 / 네비게이션** | `header-switch-view.spec.ts` (신규) | TC-NAV-01..03 | 신규. Switch View 안내 + role 전환 회귀 + Account Settings 메뉴 회귀. |
| F4 | **감사 (audit) — Kratos webhook** | `kratos-audit-webhook.spec.ts` (신규) | TC-AUD-01..02 | 신규. password 변경 → audit detail 의 source_type=kratos 확인 + target_id 가 alice identity_id 와 일치. |
| F5 | **인증 가드 (AuthGuard / 로그인 실패)** | `auth.spec.ts` 확장 + `signout.spec.ts` 확장 | TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01, TC-AUTH-SIGNOUT-REDIR-01 | 기존 auth/signout spec 확장 — 로그인 실패 메시지 + 비로그인 보호 페이지 진입 + Sign Out 후 보호 페이지. |
| F6 | **권한 매트릭스 (sub-routes)** | `rbac-routes.spec.ts` (신규) | TC-RBAC-SUB-01, TC-RBAC-MGR-01 | 신규. developer → /admin/settings/{users,permissions,audit} sub-routes + manager → /admin 차단. 기존 auth.spec 의 단일 path gating 확장. |
| F7 | **회원가입 (Sign Up)** | `signup.spec.ts` (신규) | TC-SIGNUP-01..04 | 신규. /auth/signup 페이지 + HR mock round-trip + HR 실패 + password mismatch. M3 트랙이지만 코드 main 반영 — 운영 진입 직전 회귀. |
| F8 | **권한 편집** (`/admin/settings/permissions`) | `admin-permissions.spec.ts` (신규) | TC-PERMISSIONS-SMOKE-01 | 신규. PermissionEditor 진입 + matrix 노출. M1 RBAC track 의 검증 0건이던 핵심 도구. |

이외 회귀 sanity (role landing / audit page smoke) 는 기존 spec (`auth.spec.ts` 의 role-based landing, `audit.spec.ts`) 가 커버하므로 본 sprint 신규 작성 없음.

---

## 4. F1 사용자 관리 — `admin-users-search.spec.ts`

### TC-USR-01 — 검색어 이름 부분일치 + 빈 검색어 복귀

- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. `/admin/settings/users` 진입.
  2. MemberTable 에 행이 ≥ 2 노출 확인 (seed alice/bob/charlie).
  3. SearchInput 에 `alice` 입력.
  4. alice 행만 남고 다른 행이 사라지는지 확인.
  5. SearchInput 비움.
  6. 전체 행 (≥ 2) 복귀 확인.
- **DoD**: 즉시 행 수 감소 + 빈 query 시 전체 복귀.

### TC-USR-02 — 검색어 email 부분일치

- **단계**:
  1. `/admin/settings/users` 진입.
  2. SearchInput 에 `bob@` 입력.
  3. bob 행만 남는지 확인.
- **DoD**: email 의 일부로 필터 동작.

### TC-USR-03 — 검색어 role 부분일치

- **단계**:
  1. `/admin/settings/users` 진입.
  2. SearchInput 에 `manager` 입력.
  3. role=manager 인 bob 만 남는지 확인 (alice/charlie 사라짐).
- **DoD**: role 문자열의 일부로 필터 동작.

### TC-USR-04 — Filter 버튼 disabled

- **단계**:
  1. `/admin/settings/users` 진입.
  2. Filter 버튼이 `disabled` 속성 + `title="Advanced filters coming soon"` 를 갖는지 확인.
- **DoD**: 미구현 컨트롤이 명시적으로 disabled.

### TC-USR-05 — Case-insensitive 매칭

- **단계**:
  1. `/admin/settings/users` 진입.
  2. SearchInput 에 `ALICE` (대문자) 입력.
  3. alice 행이 노출되는지 확인.
- **DoD**: `toLowerCase()` 정합성. 대문자 입력으로도 동일 결과.

### TC-USR-06 — 매칭 0건 (empty result)

- **단계**:
  1. `/admin/settings/users` 진입.
  2. SearchInput 에 `zzzz-no-match` 입력.
  3. MemberTable 의 사용자 행이 0개임을 확인.
  4. 페이지가 깨지지 않고 SearchInput 자체는 정상 동작 (다시 비우면 전체 복귀).
- **DoD**: 매칭 0건 시 행 수 0 + UI 깨짐 없음.
- **참고**: 현 MemberTable 은 empty state placeholder UI 없음 (그냥 빈 tbody). 빈 검색 결과용 명시적 UI 추가는 본 sprint 범위 밖 — 후속 sprint 로 인계 (`m2_followups.ux_member_table_empty_state`).

---

## 4-B. F1 사용자 관리 — `admin-users-crud.spec.ts` (CRUD UI smoke)

본 sprint 는 PR-UX1 의 검색 필터만 변경했고 CRUD 자체 코드는 손대지 않았다. 본 spec 는 회귀 sanity 차원의 **UI smoke** — backend mutation 은 발생시키지 않고 모달 열림 / dropdown 옵션 / 액션 메뉴 노출만 확인. 실제 round-trip 은 backend Go test (PR #54 의 10 케이스) 가 커버.

### TC-USR-CRUD-01 — Invite Member 모달 열림 + 닫기

- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. `/admin/settings/users` 진입.
  2. "Invite Member" 버튼 클릭.
  3. `UserCreationModal` 노출 확인 (모달 제목 또는 form 요소).
  4. 모달 닫기 (Close 버튼 / overlay 클릭 / ESC).
  5. 모달이 사라지는지 확인.
- **DoD**: 모달 정상 열림/닫힘. backend mutation 없음 (form submit 안 함).

### TC-USR-CRUD-02 — Role select dropdown 옵션 노출

- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. `/admin/settings/users` 진입.
  2. 임의의 row 의 Role `<select>` 요소 확인.
  3. option 갯수가 ≥ 3 이고 적어도 "Developer", "Manager", "System Admin" 3 option 이 포함되는지 확인.
- **DoD**: roles 가 dropdown 에 정상 노출. 선택값 변경은 안 함 (mutation 회피).

### TC-USR-CRUD-03 — Action 메뉴 (`...`) 에 system_admin 액션 4종 노출

- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. `/admin/settings/users` 진입.
  2. 임의의 row 의 `...` (MoreHorizontal) 버튼 클릭.
  3. dropdown menu 노출 + "Issue Account" / "Force Reset Password" / "Revoke Account" 3 액션 노출 확인.
  4. 메뉴 닫기 (overlay 클릭).
- **DoD**: system_admin 으로 진입 시 3 액션 노출. 실제 클릭은 안 함 (mutation 회피).
- **참고**: developer/manager 진입 시 미노출 검증은 redundant — `auth.spec.ts` 의 "developer cannot reach /admin/settings" 가 이미 페이지 진입 자체를 차단.

---

## 5. F2 내 계정 / 비밀번호 — `account.spec.ts`

본 spec 는 기존 `password-change.spec.ts` 를 흡수한다. 비밀번호 변경 round-trip + Kratos privileged session UX 안내 노출 + cleanup 을 한 spec 에 묶음.

### TC-ACC-01 — Current Password 라벨 안내 노출

- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. `/account` 진입.
  2. Current Password 입력란 아래에 "Required by our identity provider (Ory Kratos) for sensitive changes..." 텍스트 노출 확인.
  3. Current Password 입력란에 `aria-describedby="current-password-help"` 속성 확인.
  4. `#current-password-help` 요소 존재 확인.
- **DoD**: 안내 텍스트가 입력란 아래에 노출 + ARIA 연결.

### TC-ACC-02 — `/account` 비밀번호 변경 → Sign Out → 새 비밀번호로 재로그인 (기존 password-change.spec 흡수)

- **사전 조건**: alice (developer) 로 로그인.
- **단계**: (기존 `password-change.spec.ts` 내용 그대로 옮김)
  1. `/account` 진입.
  2. Current/New/Confirm 입력으로 password 변경.
  3. "Password updated successfully" 노출 확인.
  4. Sign Out.
  5. 새 password 로 로그인.
  6. (finally) password 원복 (best-effort; globalSetup 이 다음 run 에서 자동 복구).
- **DoD**: 양 password 모두로 로그인 round-trip 성공.

### TC-ACC-03 — New ≠ Confirm 클라이언트 측 검증

- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. `/account` 진입.
  2. Current Password 에 임의 값 (정답이 아니어도 무관 — 폼이 클라이언트 단에서 차단되므로 backend 호출 발생 안 함).
  3. New Password 에 `Foo-12345!`, Confirm New Password 에 `Bar-12345!` (서로 다른 값) 입력.
  4. Save Changes 클릭.
  5. "New passwords do not match." 에러 메시지 노출 확인.
  6. Kratos / DevHub 호출 발생 안 함 (network 활동 검증은 옵션, smoke 수준).
- **DoD**: 클라이언트 검증으로 mismatch 차단 + 명확한 에러 메시지. mutation 발생 안 함.

### TC-ACC-PROFILE-01 — `/account` 의 사용자 정보 정확성

- **목적**: alice 로 로그인하면 alice 의 login/email/role 정보가 정확히 노출 + 다른 사용자 정보 누출 없음 (회귀 + 보안).
- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. `/account` 진입.
  2. Profile Info 섹션의 `actor.login` 표시가 `alice` 확인.
  3. `actor.role` 표시가 `developer` (또는 UI 표시명 "Developer") 확인.
  4. 이메일 표시가 `alice@example.com` 확인.
- **DoD**: actor 데이터가 올바른 사용자의 정보. 다른 사용자 정보 누출 없음.

---

## 6. F3 헤더 / 네비게이션 — `header-switch-view.spec.ts`

### TC-NAV-01 — Switch View 한계 안내 노출

- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. 로그인 후 임의의 화면.
  2. Header 의 사용자 영역 (avatar + login) 클릭하여 dropdown 열기.
  3. dropdown 안에 "Switch View" 헤더 노출 확인.
  4. 그 아래에 "Menu preview only — actual permissions follow server actor.role." 안내 텍스트 노출 확인.
- **DoD**: 안내 텍스트가 Switch View 헤더 아래에 노출.

### TC-NAV-02 — Switch View role 전환 회귀

- **목적**: PR-UX3 의 dropdown 안내 추가가 기존 role 전환 동작을 깨지 않았는지 회귀 확인.
- **사전 조건**: alice (developer) 로 로그인 (시작 시 role=Developer).
- **단계**:
  1. Header dropdown 열기.
  2. "Manager" 버튼 클릭.
  3. URL 이 `/manager` 로 이동하는지 확인.
  4. Header 의 role 표시가 "Manager" 로 변경됐는지 확인.
- **DoD**: dropdown 클릭 → store role 변경 + path 전환. 기존 동작 유지.
- **구현 노트 (race-aware)** — PR #86 리뷰어 모드에서 두 가지 race 가 동시에 노출됨:
  1. `<header> getByText("Manager", exact)` 는 role badge `<span>` 과 dropdown 의 `<button>Manager</button>` 두 element 에 동시 매칭 — AnimatePresence 가 exit 중인 동안 strict-mode 위반.
  2. AuthGuard 의 `useEffect([pathname])` 가 `whoAmI()` → `setActor()` → store role 을 actor.role 로 되돌림. dropdown 이 닫히는 것을 기다리면 그 사이 role 이 "Developer" 로 reset.
  - 회피책: selector 를 `header span.uppercase.tracking-wider` 로 좁혀 badge 한 곳만 매칭 + `waitForURL` 직후 즉시 단언. 향후 Header CSS 클래스 변경 시 이 selector 도 같이 업데이트해야 함 (또는 `data-testid="header-role-badge"` 로 옮길 것 — 보강 후보).

### TC-NAV-03 — Account Settings 메뉴 → `/account` 이동

- **목적**: dropdown 의 Account Settings 진입 경로 회귀.
- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. Header dropdown 열기.
  2. "Account Settings" 버튼 클릭.
  3. URL 이 `/account` 로 이동하는지 확인.
- **DoD**: dropdown → /account 진입 정상.

### TC-NAV-SIM-01 — Switch View 시뮬레이션은 서버 RBAC 우회 불가

- **목적**: PR-UX3 안내 텍스트("Menu preview only — actual permissions follow server actor.role.") 의 효과를 직접 증거로 확인.
- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. Header dropdown 으로 Switch View → "Developer" 클릭.
  2. URL 이 `/developer` 로 이동, store role 가 Developer 로 시뮬레이션.
  3. Header 의 system menu 항목들이 시뮬레이션상 사라지는지 확인.
  4. URL 로 직접 `/admin/settings/users` 진입 시도.
  5. AuthGuard 의 `actor.role === system_admin` 기준으로 페이지 정상 노출 (시뮬레이션 우회 못함).
- **DoD**: 시뮬레이션은 UI 미리보기, 실제 권한은 서버 actor.role 기준 — 안내가 사실과 일치.

---

## 7. F4 감사 (audit) — `kratos-audit-webhook.spec.ts`

### TC-AUD-01 — `/account` 비밀번호 변경 → audit detail 에 `source_type=kratos` 노출

- **목적**: Kratos self-service password 변경이 DevHub `audit_logs` 에 **`source_type=kratos`** 행으로 자동 기록되는지 확인. 단순 행 노출이 아니라 entry detail 의 `source_type` 필드까지 검증해 webhook 이 실제로 발화됐다는 강한 증거를 확보 (false-positive 회피).
- **사전 조건**:
  - 5-process stack 가동
  - 양쪽에 `DEVHUB_KRATOS_WEBHOOK_TOKEN` 동일 값 export
  - Kratos 재기동으로 새 hook 활성
  - 본 spec 는 alice 의 password 를 임시 변경하므로 `account.spec.ts` 와 같은 run 에 함께 실행될 때 ordering 충돌 위험 → 본 spec 의 try/finally rollback + globalSetup 자동 복구로 mitigated.
- **단계**:
  1. alice (developer) 로 로그인.
  2. `/account` 진입.
  3. password 변경.
  4. "Password updated successfully" 확인.
  5. Sign Out.
  6. charlie (system_admin) 로 로그인.
  7. `/admin/settings/audit` 진입.
  8. `account.password_changed` 행이 audit 목록에 노출되는지 확인 (최신 entry prepend 가정).
  9. **해당 entry 클릭 → detail 영역에 `source_type` = `kratos` 확인 (필수)**.
  10. (옵션) `target_type` = `kratos_identity` 도 함께 확인.
  11. Sign Out (charlie).
  12. (finally) alice 로 로그인 + password 원복 (best-effort; globalSetup 이 다음 run 에서 자동 복구).
- **DoD**:
  - password 변경 성공.
  - audit 페이지에 `account.password_changed` entry 노출.
  - **entry detail 에 `source_type=kratos` 노출** (webhook 실작동 강한 증거).
- **위험 / 한계**:
  - Kratos `response.ignore: true` 라 webhook 실패해도 password 변경 자체는 success → false-positive 위험. 그러므로 단계 9 의 source_type 검증이 핵심.
  - 본 spec 와 F2 `account.spec.ts` 의 TC-ACC-02 가 모두 alice 의 password 를 임시 변경 → ordering 의존. Playwright `fullyParallel: false, workers: 1` + 파일명 알파벳 순으로 `account.spec.ts` (a) → `kratos-audit-webhook.spec.ts` (k) 순 실행. 양쪽 모두 finally rollback 하므로 안전. **운영 가정**: 양 spec 모두 알파벳 순 보장 (수동 ordering 미주입).

### TC-AUD-02 — audit detail 의 `target_id` 가 alice 의 Kratos identity_id 와 일치

- **목적**: webhook 가 올바른 사용자 (alice) 의 password 변경에 대해 발화됐다는 엄격한 증거 확보. `target_id` ↔ Kratos identity_id 매칭.
- **사전 조건**: TC-AUD-01 과 동일.
- **단계**:
  1. (fixture 확장) Kratos `/admin/identities` 를 호출해 `traits.email = alice@example.com` 인 identity 의 `id` 추출 → 변수 `aliceIdentityID`.
  2. TC-AUD-01 단계 1~9 수행 (audit detail 영역까지).
  3. detail 영역에서 `target_id` 필드 확인.
  4. `target_id === aliceIdentityID` 검증.
  5. cleanup 동일.
- **DoD**: detail 의 target_id 가 Kratos 가 보유한 alice identity_id 와 정확히 일치.
- **위험 / 한계**:
  - `account.spec.ts` 와 같이 ordering 영향 받음 (TC-AUD-01 과 동일).
  - fixture 확장 필요: `fixtures.ts` 에 `getKratosIdentityIdByEmail(email)` helper 추가. global-setup.ts 의 `listExistingIdentities` 패턴 그대로 활용 가능.
  - 본 TC 는 TC-AUD-01 과 같은 password 변경을 공유 → 한 spec 안에서 두 TC 가 같은 변경을 두 번 일으키지 않도록 helper 재사용 또는 단일 변경 + 두 검증 패턴.

---

## 7-A. F5 인증 가드 — `auth.spec.ts` + `signout.spec.ts` 확장

기존 spec 들에 회귀 TC 를 추가한다. 새 spec 파일 신설하지 않음 — 동일 기능 카테고리.

### TC-AUTH-NEG-01 — 잘못된 비밀번호 → 에러 메시지 (auth.spec.ts 확장)

- **사전 조건**: 어떤 사용자도 로그인 안 한 상태.
- **단계**:
  1. `/login` 진입 → OIDC dance 후 `/auth/login?login_challenge=...` 도달.
  2. System ID `alice`, Password `wrong-password` 입력.
  3. Sign In 클릭.
  4. Kratos 에러 메시지 노출 확인 (정확한 문구는 Kratos 버전 의존이지만 "credentials are invalid" 또는 "incorrect" 류).
  5. URL 이 여전히 `/auth/login` (login challenge 살아있음).
- **DoD**: 로그인 실패 시 명시적 에러 메시지, 임의 페이지로 진행 안 됨.

### TC-AUTH-NOAUTH-01 — 비로그인 상태 보호 페이지 진입 → /login 리다이렉트 (auth.spec.ts 확장)

- **목적**: AuthGuard 가 401/no-actor 시 `/login` 으로 리다이렉트하는지.
- **사전 조건**: 어떤 사용자도 로그인 안 한 상태.
- **단계**:
  1. 직접 `/developer` URL 진입 (로그인 없이).
  2. AuthGuard 의 whoAmI 401 → `router.replace("/login")` → OIDC dance 시작.
  3. `/auth/login?login_challenge=...` 에 도달하는지 확인.
- **DoD**: 비로그인 보호 페이지 접근이 항상 로그인 흐름으로 강제.

### TC-AUTH-SIGNOUT-REDIR-01 — Sign Out 후 보호 페이지 진입 → /login (signout.spec.ts 확장)

- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  1. Header dropdown → Sign Out.
  2. (signout.spec 의 기존 검증과 동일 시점) `/` 또는 `/login` 진입.
  3. 직접 `/developer` URL 진입 시도.
  4. AuthGuard 가 401 → `/login` → OIDC dance.
- **DoD**: Sign Out 후 보호 페이지 진입이 차단되고 재인증 요구.

### TC-USER-SWITCH-01 — 사용자 전환 (signout.spec.ts 확장)

- **목적**: alice 로그아웃 후 bob 으로 로그인 → bob 페이지 정상 노출, alice 정보 잔재 없음.
- **사전 조건**: 어떤 사용자도 로그인 안 한 상태.
- **단계**:
  1. alice (developer) 로 로그인 → `/developer` landing.
  2. `/account` 진입 → Profile Info 의 login 이 `alice` 인 것 확인.
  3. Sign Out.
  4. bob (manager) 로 로그인 → `/manager` landing.
  5. `/account` 진입 → Profile Info 의 login 이 `bob`, email 이 `bob@example.com`, role 이 manager 인 것 확인. alice 정보 잔재 0.
- **DoD**: 사용자 전환이 깨끗 — 이전 사용자 데이터가 store / UI 어디에도 남지 않음.

---

## 7-B. F6 권한 매트릭스 — `rbac-routes.spec.ts`

기존 `auth.spec.ts` 의 "system route gating" 은 단일 path (`/admin/settings`) 만 검증. 본 spec 는 sub-routes 매트릭스로 확장.

### TC-RBAC-SUB-01 — developer → /admin/settings/* sub-routes 차단

- **사전 조건**: alice (developer) 로 로그인.
- **단계**:
  - `/admin/settings/users` 진입 → `/developer` 로 redirect.
  - `/admin/settings/permissions` 진입 → `/developer` 로 redirect.
  - `/admin/settings/audit` 진입 → `/developer` 로 redirect.
  - `/admin/settings/organization` 진입 → `/developer` 로 redirect.
- **DoD**: 모든 admin sub-routes 가 일관되게 차단.

### TC-RBAC-MGR-01 — manager → /admin 차단

- **사전 조건**: bob (manager) 로 로그인.
- **단계**:
  1. `/admin` 진입.
  2. AuthGuard 의 `pathRequiresSystemAdmin(/admin) && !isSystemAdmin('manager')` 참 → `/manager` 로 redirect.
- **DoD**: manager 도 system admin 영역 차단됨.

---

## 7-C. F7 회원가입 — `signup.spec.ts`

`/auth/signup` 페이지 + `POST /api/v1/auth/signup` 흐름. HRDB Mock (3명: yklee/akim/sjones) 에 의존.

### TC-SIGNUP-01 — `/auth/signup` 페이지 진입 + 폼 + Sign In 링크

- **사전 조건**: 비로그인 상태.
- **단계**:
  1. `/auth/signup` 직접 진입.
  2. 폼 필드 (Full Name / System ID / Employee ID / Password / Confirm Password) 노출 확인.
  3. "Sign In" 링크 노출 확인 + 클릭 시 `/login` 로 이동.
- **DoD**: 페이지 진입 + 핵심 form 요소 노출.

### TC-SIGNUP-02 — HR mock 매칭 → 가입 round-trip

- **목적**: HR DB lookup → Kratos identity 생성 → DevHub user 생성 라운드트립.
- **사전 조건**: 비로그인. yklee (Kratos identity 미존재 가정 — 이전 run 의 cleanup 가정).
- **단계**:
  1. `/auth/signup` 진입.
  2. Full Name `YK Lee`, System ID `yklee`, Employee ID `1001`, Password / Confirm `Signup-12345!` 입력.
  3. Register 클릭.
  4. "Identity Verified! ... Redirecting to sign in..." success 메시지 노출.
  5. `/login` 으로 redirect 확인 (3초 setTimeout).
  6. (옵션) 새 yklee 계정으로 로그인 시도 → 성공.
  7. **(finally) cleanup**: Kratos `/admin/identities/{id}` DELETE 로 yklee identity 삭제 + DevHub `users` 행 삭제 (또는 status disable). fixture 의 `deleteKratosIdentityByEmail("yklee@example.com")` helper 사용.
- **DoD**: Sign Up round-trip 성공 + cleanup 으로 다음 run 에 영향 없음.
- **위험 / 한계**:
  - cleanup fail 시 다음 run 의 TC-SIGNUP-02 가 fail (already exists). 운영자 수동 cleanup 절차 필요 — `kratos delete identity --id <id>` 또는 admin API.
  - DevHub `users` 행은 backend `authSignUp` 가 OrganizationStore.CreateUser 호출. cleanup 은 별도 API 가 없어 SQL 직접 또는 admin endpoint 필요. 본 sprint 는 Kratos identity cleanup 만 — DevHub users 행은 globalSetup 의 002_seed_e2e_users.sql 이 다음 run 마다 idempotent 실행하므로 잔재해도 무관.
  - 새 cleanup helper 는 단일 spec 한정 사용. 다른 spec 영향 없음.

### TC-SIGNUP-03 — HR DB 매칭 실패 → 거절

- **사전 조건**: 비로그인.
- **단계**:
  1. `/auth/signup` 진입.
  2. Full Name `Nobody`, System ID `unknown-id`, Employee ID `9999`, Password / Confirm `Signup-12345!` 입력.
  3. Register 클릭.
  4. "identity verification failed" 또는 "hr_lookup_failed" 류 에러 노출.
- **DoD**: HR mock 미매칭 시 forbidden 메시지. Kratos identity 생성 안 됨.

### TC-SIGNUP-04 — Password mismatch 클라이언트 검증

- **단계**:
  1. `/auth/signup` 진입.
  2. 임의 HR 정보 + Password `A`, Confirm `B` 입력.
  3. Register 클릭.
  4. "Passwords do not match." 에러 노출.
  5. backend 호출 발생 안 함.
- **DoD**: 클라이언트 측 mismatch 차단.

---

## 7-D. F8 권한 편집 — `admin-permissions.spec.ts`

`/admin/settings/permissions` 의 PermissionEditor 가 M1 RBAC track (PR-G6) 으로 머지됐지만 e2e 검증 0건. 본 sprint 게이트로 smoke 1건 추가.

### TC-PERMISSIONS-SMOKE-01 — PermissionEditor 진입 + matrix 노출

- **사전 조건**: charlie (system_admin) 로 로그인.
- **단계**:
  1. `/admin/settings/permissions` 진입.
  2. PermissionEditor 영역 노출 확인.
  3. matrix 가시화 — 5 resource (infrastructure / pipelines / organization / security / audit) 가 각각 4 action (view / create / edit / delete) 와 함께 노출되는지 행/열 수 검증 (또는 명시적 cell 존재 확인).
  4. mutation 없음 — 편집/저장은 클릭 안 함.
- **DoD**: PermissionEditor 가 정상 렌더 + matrix 가 가시. backend mutation 없음.

---

## 8. 기존 Backend 단위 테스트 (참고)

PR-M2-AUDIT 머지에서 이미 작성 + PASS — 본 sprint 추가 작성 없음.

| TC | Go 테스트 |
| --- | --- |
| webhook 성공 path → audit_logs INSERT | `TestKratosPasswordWebhookRecordsAuditLog` |
| Bearer prefix 없는 bare token 도 허용 | `TestKratosPasswordWebhookAcceptsBareToken` |
| Authorization 헤더 누락 → 401 | `TestKratosPasswordWebhookRejectsMissingAuth` |
| secret 불일치 → 401 | `TestKratosPasswordWebhookRejectsWrongSecret` |
| 토큰 unset → 503 | `TestKratosPasswordWebhookUnavailableWhenSecretUnset` |
| identity_id 누락 → 400 | `TestKratosPasswordWebhookRejectsMissingIdentityID` |
| invalid JSON → 400 | `TestKratosPasswordWebhookRejectsInvalidJSON` |

---

## 9. 결정 사항 (확정)

### 9.1 초기 결정 (2026-05-12)

- **spec 배치**: 기능별 분할 (사용자 결정).
- **alice 충돌 처리**: `account.spec.ts` 와 `kratos-audit-webhook.spec.ts` 모두 alice 사용. 양쪽 try/finally rollback + globalSetup 자동 복구로 안전. 알파벳 순 ordering 보장.
- **`password-change.spec.ts` 처리**: `account.spec.ts` 에 흡수 후 삭제.
- **Kratos hook 실작동 검증 깊이**: smoke 수준 (UI 에서 audit 행 노출만 확인). DB 직접 쿼리는 본 sprint 범위 밖.

### 9.2 기능별 상세 검토 결정 (2026-05-12, 사용자와 1:1 검토)

| 기능 | 결정 |
| --- | --- |
| F1 | (1) "빈 검색어 복귀" 케이스를 TC-USR-01 에 step 추가. (2) case-insensitive TC-USR-05 추가. (3) 매칭 0건 TC-USR-06 추가 (empty state UI 부재는 후속 sprint 인계). (4) CRUD UI smoke 3 TC 추가 (admin-users-crud.spec.ts) — Invite Member 모달 / Role dropdown / Action 메뉴 admin 노출. backend mutation 회피. round-trip 은 backend Go test 가 커버. |
| F2 | (5) TC-ACC-03 클라이언트 mismatch 검증 추가. weak password / REAUTH 는 후속 sprint 인계. |
| F3 | (6) TC-NAV-02 Switch View role 전환 회귀 추가 (PR-UX3 의 dropdown 변경이 기존 동작 깨지 않았다는 직접 증거). (7) TC-NAV-03 Account Settings 메뉴 회귀 추가. |
| F4 | (8) TC-AUD-01 의 source_type detail 검증을 옵션 → 필수로 격상. (9) TC-AUD-02 추가 — target_id 가 alice Kratos identity_id 와 일치 (가장 엄격한 webhook 발화 검증). fixture 확장 필요 (`getKratosIdentityIdByEmail` helper). |

### 9.3 후속 sprint 인계 (`m2_followups.*`)

**TC / e2e 차원**:
- `m2_followups.ux_member_table_empty_state` — MemberTable 의 empty state UI (현재 없음).
- `m2_followups.weak_password_e2e` — Kratos 의 약한 password 거절 e2e.
- `m2_followups.reauth_e2e` — privileged session 만료 시 REAUTH_REQUIRED 흐름 e2e.
- `m2_followups.disabled_user_login_e2e` — status=disabled 사용자 로그인 차단 e2e.
- `m2_followups.role_landing_content_e2e` — 각 role 의 초기 페이지 콘텐츠 렌더링 (URL 만 보던 검증을 위젯 단까지).
- `m2_followups.signup_devhub_user_cleanup` — Sign Up e2e 의 DevHub `users` 행 cleanup 절차 표준화.
- `m2_followups.audit_filter_e2e` — `/admin/settings/audit` 의 action/target_type filter 동작 e2e.
- `m2_followups.organization_admin_e2e` — `/admin/settings/organization` 의 부서/멤버/리더 관리 e2e.

**코드 / 보안 차원 (TC 가 아닌 결함 후보)**:
- `m2_followups.kratos_revoke_active_sessions` — `infra/idp/kratos.yaml` 의 `settings.after.password.hooks` 에 `revoke_active_sessions` 미설정 → password 변경 후 기존 access token 살아있음. 운영 진입 직전 정책 결정 + Kratos 설정 1줄 추가 + 회귀 TC.
- `m2_followups.first_login_force_change` — `must_change_password` 흐름 미구현 (frontend roadmap §6.1 에 명시됐지만 코드 0건). 신규 가입자 / admin 발급 임시 비번 사용자의 첫 로그인 강제 변경. backend `users.must_change_password` 필드 + frontend redirect 로직 + e2e.

### 9.4 추가 결정 (2026-05-12, 2차 검토)

| 카테고리 | 결정 |
| --- | --- |
| 인증 가드 회귀 | TC-AUTH-NEG-01 + TC-AUTH-NOAUTH-01 → 기존 `auth.spec.ts` 확장. TC-AUTH-SIGNOUT-REDIR-01 → 기존 `signout.spec.ts` 확장. 새 파일 분리 안 함 (동일 기능 카테고리). |
| 권한 매트릭스 sub-routes | 새 spec `rbac-routes.spec.ts` 신설 — 기존 auth.spec 의 단일 path gating 와 분리해 매트릭스 단위로 관리. |
| 회원가입 | 새 spec `signup.spec.ts` 신설 (M3 트랙이지만 코드 main 반영, 회귀 가치 큼). |
| 디스에이블 / 콘텐츠 로딩 | 본 sprint 미포함, 후속 sprint 인계 (위 §9.3). |

### 9.5 추가 결정 (2026-05-12, 3차 검토 — 사용자 journey 관점)

| 카테고리 | 결정 |
| --- | --- |
| 사용자 전환 (alice → bob) | TC-USER-SWITCH-01 → `signout.spec.ts` 확장 (Sign Out + 다른 사용자 로그인 흐름은 signout 기능 연장). |
| /account 정보 정확성 | TC-ACC-PROFILE-01 → `account.spec.ts` 확장 (TC-ACC-01..03 와 같은 파일). |
| Switch View 시뮬레이션 우회 검증 | TC-NAV-SIM-01 → `header-switch-view.spec.ts` 확장 (TC-NAV-01..03 와 같은 파일). PR-UX3 안내의 사실성 직접 증거. |
| PermissionEditor smoke | 새 spec `admin-permissions.spec.ts` 신설 — M1 RBAC track 검증 0건이던 핵심 도구 1회 회귀. |
| Kratos revoke_active_sessions 미설정 | 코드 결함 후보. 본 sprint TC 작성 단계에서 발견된 갭으로 §9.3 의 `m2_followups.kratos_revoke_active_sessions` 에 인계. |
| must_change_password 미구현 | 코드 0건 — 본 sprint 미진입, §9.3 의 `m2_followups.first_login_force_change` 인계. |

---

## 10. spec 작성 순서 (제안)

작성 순서 — 외부 stack 의존이 적은 것 → 무거운 것 순:

1. `admin-users-search.spec.ts` — TC-USR-01..06 (가장 단순, 검색만)
2. `admin-users-crud.spec.ts` — TC-USR-CRUD-01..03 (UI smoke, mutation 없음)
3. `admin-permissions.spec.ts` 신규 — TC-PERMISSIONS-SMOKE-01 (smoke 1)
4. `header-switch-view.spec.ts` — TC-NAV-01..03, TC-NAV-SIM-01
5. `auth.spec.ts` 확장 — TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01
6. `signout.spec.ts` 확장 — TC-AUTH-SIGNOUT-REDIR-01, TC-USER-SWITCH-01
7. `rbac-routes.spec.ts` 신규 — TC-RBAC-SUB-01, TC-RBAC-MGR-01
8. `account.spec.ts` (`password-change.spec.ts` 흡수 후 삭제) — TC-ACC-01..03, TC-ACC-PROFILE-01
9. `signup.spec.ts` 신규 — TC-SIGNUP-01..04 (HRDB mock + Kratos identity cleanup 의존)
10. `kratos-audit-webhook.spec.ts` 신규 — TC-AUD-01..02 (Kratos webhook 실작동 + identity_id fixture 확장 의존)

작성 후 `npm run e2e` 로 사용자 환경에서 실행.

### 신규 / 변경 / 삭제 spec 요약

| 파일 | 작업 | TC |
| --- | --- | --- |
| `frontend/tests/e2e/admin-users-search.spec.ts` | 신규 | TC-USR-01..06 (6) |
| `frontend/tests/e2e/admin-users-crud.spec.ts` | 신규 | TC-USR-CRUD-01..03 (3) |
| `frontend/tests/e2e/header-switch-view.spec.ts` | 신규 | TC-NAV-01..03, TC-NAV-SIM-01 (4) |
| `frontend/tests/e2e/account.spec.ts` | 신규 (password-change.spec 흡수) | TC-ACC-01..03, TC-ACC-PROFILE-01 (4) |
| `frontend/tests/e2e/kratos-audit-webhook.spec.ts` | 신규 | TC-AUD-01..02 (2) |
| `frontend/tests/e2e/auth.spec.ts` | 확장 (기존 + 2 TC) | + TC-AUTH-NEG-01, TC-AUTH-NOAUTH-01 |
| `frontend/tests/e2e/signout.spec.ts` | 확장 (기존 + 2 TC) | + TC-AUTH-SIGNOUT-REDIR-01, TC-USER-SWITCH-01 |
| `frontend/tests/e2e/rbac-routes.spec.ts` | 신규 | TC-RBAC-SUB-01, TC-RBAC-MGR-01 (2) |
| `frontend/tests/e2e/signup.spec.ts` | 신규 | TC-SIGNUP-01..04 (4) |
| `frontend/tests/e2e/admin-permissions.spec.ts` | 신규 | TC-PERMISSIONS-SMOKE-01 (1) |
| `frontend/tests/e2e/password-change.spec.ts` | **삭제** (account.spec.ts 가 흡수) | — |
| `frontend/tests/e2e/fixtures.ts` | 확장 — `getKratosIdentityIdByEmail(email)`, `deleteKratosIdentityByEmail(email)` helper 추가 | — |
| 기존 `audit.spec.ts` | 변경 없음 | (회귀 sanity) |

**총 30 TC, 8 신규 spec, 2 기존 spec 확장, 1 spec 삭제, 1 fixture 확장 (2 helper).**

---

## 11. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-12 | sprint 검증 TC 리스트 초안 작성. |
| 2026-05-12 | 사용자 결정 반영 — 기능별 4 spec 분할, password-change.spec 흡수, smoke 깊이. |
| 2026-05-12 | 기능별 1:1 상세 검토 — F1 검색 6 + CRUD 3, F2 안내+round-trip+mismatch 3, F3 안내+role 전환+Account 메뉴 3, F4 source_type 필수 + target_id 매칭 2. 총 17 TC, 5 spec. |
| 2026-05-12 | 2차 갭 검토 — F5 인증 가드 (auth/signout 확장) 3 TC + F6 권한 매트릭스 sub-routes 2 TC + F7 회원가입 4 TC. 총 26 TC, 7 신규 spec, 2 기존 확장. fixture 2 helper. |
| 2026-05-12 | 3차 사용자 journey 검토 — 사용자 전환 (USER-SWITCH-01) + /account 정보 정확성 (ACC-PROFILE-01) + Switch View 우회 검증 (NAV-SIM-01) + PermissionEditor smoke (PERMISSIONS-SMOKE-01) 추가. Kratos `revoke_active_sessions` 미설정 + `must_change_password` 미구현 발견 — 후속 sprint 인계. 총 30 TC, 8 신규 spec. |
