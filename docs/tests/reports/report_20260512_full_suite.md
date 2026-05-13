# E2E 테스트 결과 보고서

- **일시**: 2026-05-12 23:59 KST
- **브랜치**: `gemini/prepare-github-action`
- **커밋**: `00f89a5`
- **환경**: Local (macOS, Chromium, Workers=1)
- **총 테스트**: 40 | **통과**: 40 | **실패**: 0
- **실행 시간**: 1분 36초

## 테스트 결과 요약

| Spec File | TC Count | Status | Duration |
|---|---|---|---|
| `account.spec.ts` | 4 | ✅ All Pass | ~12s |
| `admin-org-crud.spec.ts` | 5 | ✅ All Pass | ~14s |
| `admin-permissions.spec.ts` | 1 | ✅ All Pass | ~4s |
| `admin-users-crud.spec.ts` | 3 | ✅ All Pass | ~13s |
| `admin-users-search.spec.ts` | 6 | ✅ All Pass | ~10s |
| `audit.spec.ts` | 1 | ✅ All Pass | ~3s |
| `auth.spec.ts` | 6 | ✅ All Pass | ~15s |
| `header-switch-view.spec.ts` | 4 | ✅ All Pass | ~8s |
| `kratos-audit-webhook.spec.ts` | 1 | ✅ All Pass | ~3s |
| `rbac-routes.spec.ts` | 2 | ✅ All Pass | ~4s |
| `signout.spec.ts` | 3 | ✅ All Pass | ~8s |
| `signup.spec.ts` | 4 | ✅ All Pass | ~55s |

## 이번 세션에서 수정된 결함

### 1. 모달 ARIA 역할 누락 (M3 Org, M2 Users)
- **증상**: `getByRole("dialog")` 로 모달을 찾지 못해 TC-ORG-UNIT-01, TC-ORG-MEM-01 실패
- **원인**: `UnitManagementModal`, `MemberManagementModal`, `UserCreationModal`에 `role="dialog"` 미설정
- **조치**: 3개 모달 모두 `role="dialog"`, `aria-modal="true"`, `aria-labelledby` 추가

### 2. Escape 키 모달/메뉴 닫기 미지원
- **증상**: Escape 키로 모달/액션메뉴가 닫히지 않아 TC-ORG-MEM-01, TC-USR-CRUD-03 실패
- **원인**: 각 컴포넌트에 keydown 이벤트 리스너 없음
- **조치**: 4개 컴포넌트(`UnitManagementModal`, `MemberManagementModal`, `UserCreationModal`, `MemberTable`)에 Escape 리스너 추가

### 3. Signup 폼 라벨 연결 미흡
- **증상**: `getByLabel(/full name/i)` 등의 쿼리가 요소를 찾지 못해 TC-SIGNUP-01~04 전부 실패
- **원인**: `InputField` 컴포넌트의 `<label>`에 `htmlFor` 속성 없음, `<input>`에 `id` 없음
- **조치**: 라벨 텍스트 기반 `id` 자동 생성 + `htmlFor`/`id` 연결

### 4. Signout 후 User Switch ERR_ABORTED
- **증상**: 로그아웃 후 다른 사용자로 로그인 시 `page.goto("/login")` 에서 `net::ERR_ABORTED`
- **원인**: OIDC redirect chain 중 Next.js dev mode에서 navigation abort 발생
- **조치**: `ERR_ABORTED` catch 패턴 + 올바른 로그인 폼 셀렉터(`System ID` 라벨) 사용

## 기술 상세

### 변경된 파일
| 파일 | 변경 유형 |
|---|---|
| `organization/page.tsx` | Loading/Error 상태, aria-label 추가 |
| `UnitManagementModal.tsx` | role="dialog", Escape 리스너, aria-label |
| `MemberManagementModal.tsx` | role="dialog", Escape 리스너, aria-label |
| `UserCreationModal.tsx` | role="dialog", Escape 리스너 |
| `MemberTable.tsx` | Escape 리스너 (Action 메뉴) |
| `auth/signup/page.tsx` | InputField htmlFor/id 연결 |
| `admin-org-crud.spec.ts` | 전면 재작성 (aria 기반 셀렉터) |
| `admin-users-crud.spec.ts` | 전면 재작성 (loading wait, toBeAttached) |
| `signout.spec.ts` | user-switch ERR_ABORTED 처리 |
| `fixtures.ts` | 타임아웃 30s 상향 |
