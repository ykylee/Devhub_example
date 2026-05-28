# router.replace() Silent Failure After Idle — Root Cause Analysis

- 일자: 2026-05-28
- 상태: draft
- 관련 브랜치: `analysis/router-replace-idle-failure`
- 관련 이슈: Admin Catalog 탭 클릭 시 600초 idle 후 URL 변경 없음, 에러 없음
- 분석 방법: Playwright E2E 재현(15s/120s/600s) + Next.js 16.2.6 router 내부 소스 추적(5개 핵심 파일) + nginx/KC config 확인

---

## 1. 증상 요약

| 조건 | 결과 |
|------|------|
| 15초 idle 후 tab click | 정상 동작 |
| 120초 idle 후 tab click | 정상 동작 |
| **600초(10분) idle 후 tab click** | **URL 변경 없음, 콘솔 에러 없음, 사용자 피드백 없음** |
| `<Link>` navigation (다른 페이지) | 항상 정상 동작 |
| 브라우저 새로고침 후 즉시 tab click | 정상 동작 |

---

## 2. 전체 아키텍처 컨텍스트

### 2.1 배포 구조

```
nginx:13000
 ├── /devhub/          → frontend:3000 (Next.js standalone, basePath="/devhub")
 ├── /devhub/api/      → backend-core:8080 (rewrite strip prefix)
 └── /auth/            → keycloak:8080
```

### 2.2 인증 흐름

```
[브라우저] -- session cookie --> nginx --(proxy)--> Next.js standalone
  |
  ├── apiClient (direct fetch) -- Authorization: Bearer <access_token>
  │     └── 401 → refreshAccessToken()
  |
  ├── refresh-scheduler (proactive, 60s buffer)
  │     └── refreshAccessToken() → schedule next refresh
  |
  └── Next.js RSC fetch (fetchServerResponse) -- credentials: 'same-origin'
        └── session cookie only, NO Authorization header
```

### 2.3 Keycloak Token Lifespan

- accessTokenLifespan: 300s (기본값, realm config에 명시적 설정 없음)
- refreshTokenLifespan: 1800s (기본값)

---

## 3. 상세 원인 분석

### 3.1 타이밍 체인

```
T=0s    Access token 발급 (만료 T+300s)
        refresh-scheduler: T+(300-60)=T+240s 에 refresh 예약

T=240s  refresh-scheduler → refreshAccessToken() 성공
        새 token: 만료 T+240+300=T+540s
        refresh-scheduler: T+540-60=T+480s 에 refresh 예약

T=480s  refresh-scheduler → refreshAccessToken() transient_failed
        → line 90-92: console.warn 만 호출, reschedule() 호출 안 함
        → timer=null, scheduler 영구 중단 ⚠️

T=540s  Access token 만료. refresh-scheduler 중단 상태 → proactive refresh 없음

T=600s  사용자 tab 클릭 → router.replace("/admin/catalog?tab=repositories")
        → fetchServerResponse() → 서버 RSC 요청 (세션 쿠키만, 만료된 token)
        → 서버 응답 실패 (401/500)
        → navigation.js:102 .catch(() => state) → silent drop 🚫
```

### 3.2 refresh-scheduler: transient_failed 시 reschedule 누락 (`refresh-scheduler.ts`)

```typescript
// line 68-71: setTimeout 내부에서 timer=null
timer = setTimeout(() => {
    timer = null;       // ← 타이머 정리
    void runRefresh();  // ← refresh 실행
}, delay);

// line 87-92: runRefresh 결과 처리
if (outcome.kind === "auth_failed") {
    triggerSessionExpired();  // 세션 사망 → 로그인 페이지
} else if (outcome.kind === "transient_failed") {
    console.warn("... — will retry on next 401 or expiry");  // ← reschedule() 없음!
}
// ok: tokenStore.save → subscribeExpiryChange → reschedule 호출
```

**문제**: `transient_failed` 분기에서 `reschedule()`을 호출하지 않음. 주석에는 "will retry on next 401 or expiry"라고 되어 있지만:
- 401(reactive)이 발생하지 않으면 retry 기회 없음
- expiry 자체를 감지할 메커니즘도 refresh-scheduler 중단 이후에는 없음
- `outcome.kind === "ok"`인 경우만 `tokenStore.save` → `subscribeExpiryChange` → `reschedule()`이 체인으로 동작

**결과**: 한 번 transient 실패하면 스케줄러가 영구히 정지한다.

### 3.3 fetchServerResponse: token refresh 로직 미포함 (Next.js 내부)

`node_modules/next/dist/client/components/router-reducer/fetch-server-response.js`:

```javascript
// ≈line 316: 독립적인 fetch
const fetchUrl = /* RSC endpoint URL */;
const res = await fetch(fetchUrl, {
    // credentials: 'same-origin' — 세션 쿠키만 전송
    // Authorization header 없음
    // apiClient 경유하지 않음
});
```

- **apiClient의 401→refresh 로직 적용 안 됨**
- 서버가 세션 만료를 감지하면 non-200 응답
- Next.js 서버단에서도 token 갱신 없이 세션 만료 응답 반환

### 3.4 navigateToUnknownRoute: silent catch (`navigation.js`)

`node_modules/next/dist/client/components/segment-cache/navigation.js`:

```javascript
// ≈line 96-104
function navigateToUnknownRoute(url, state, ...) {
    return fetchServerResponse(url)
        .then(response => processResponse(response, state))
        .catch(err => {
            // ← console.error 없음, 모두 silent
            return state;  // ← 현재 state 그대로 반환
        });
}
```

- `.catch(() => state)`가 모든 에러를 삼킴
- 반환된 `state`는 기존 state와 동일한 참조
- React가 변화를 감지하지 못해 화면/URL 업데이트 생략

### 3.5 useActionQueue: state 동일성 검사로 업데이트 스킵

`node_modules/next/dist/client/components/use-action-queue.js`:

```javascript
// ≈line 104-107
// reducer가 동일한 state를 반환 → React setState가 skip
// → re-render 없음, URL 변화 없음
```

---

## 4. `<Link>`가 정상 동작하는 이유

```
<Link href="/applications/42">
  ├── prefetch: 뷰포트 진입 시점에 RSC prefetch 요청
  │     → prefetch cache에 route tree 저장
  │
  └── click: navigateUsingPrefetchedRouteTree
        → 서버 요청 없이 cache로 클라이언트 사이드 navigation
        → token 만료와 무관하게 동작 ✅
```

**핵심 차이**: `<Link>`는 prefetch cache 기반(client-side only), `router.replace()`는 항상 `navigateToUnknownRoute` → 서버 RSC fetch 필수.

---

## 5. 재현 환경

- Next.js 16.2.6 (standalone output, basePath="/devhub")
- React 19.2.4
- Keycloak 22+ (access token lifespan 기본 300s)
- nginx reverse proxy (localhost:13000)
- 배포 모드: 6 containers (nginx, frontend, backend-core, backend-ai, keycloak, db)
- 브라우저: Chromium (Playwright headless, E2E 테스트)

---

## 6. Fix 방안

### 6.1 [P0] refresh-scheduler: transient_failed 시 재시도 + reschedule

**대상 파일**: `frontend/lib/auth/refresh-scheduler.ts`

**변경 내용**:
```typescript
} else if (outcome.kind === "transient_failed") {
    console.warn("[refresh-scheduler] proactive refresh transient_failed:",
      outcome.reason, "— retrying");
    // reschedule()을 즉시 호출 — 현재 token의 남은 만료 시간 기준 재계산
    reschedule();
}
```

**효과**:
- transient 실패 시 현재 token의 남은 유효시간 기준으로 재스케줄
- 다음 타이머에서 재시도
- 연속 실패해도 token 만료 전 마지막 기회까지 retry

**리스크**:
- transient가 연속 실패하면 만료 직전까지 반복 시도 (OK — 의도된 동작)
- 서버가 과부하 상태면 retry가 부하加重 → 지수 백오프 추가 고려
- auth_failed 분기에는 영향 없음 (세션 사망은 그대로 처리)

### 6.2 [P1] Navigation: tab 전환을 `<Link>` prefetch 패턴으로 변경

**대상 파일**: `frontend/app/(dashboard)/admin/catalog/page.tsx`

**변경 내용**:
- `router.replace()` 대신 `<Link>` + prefetch로 tab 변경
- tab 버튼을 `<Link>`로 감싸고 prefetch 활성화
- 또는 navigation 발생 전 token 유효성 체크 로직 추가

**효과**: 서버 RSC 요청 의존성 제거, prefetch cache로 client-side navigation

**리스크**: `<Link>`는 full page navigation(soft)이 아닌 hard navigation을 유발할 수 있어 UX 차이 발생

### 6.3 [P1] fetchServerResponse: Authorization header 주입 (Next.js patch)

**대상**: `node_modules/next/dist/client/components/router-reducer/fetch-server-response.js`
(패치 불가능한 라이브러리 — 직접 수정 시 유지보수 불가)

**결정**: **제외**. Next.js 내부 코드를 직접 패치하면 업그레이드 시마다 conflict 발생. 유지보수 불가.

### 6.4 [P2] Idle 감지 + token keepalive

**대상 파일**: `frontend/lib/auth/refresh-scheduler.ts` 또는 별도 idle watchdog

**변경 내용**:
- window event (click, keydown, scroll) 기반 idle 감지
- 일정 idle 시간(예: 240s) 경과 후 주기적 whoAmI() 또는 token 남은 만료 확인
- token 만료 임박 시 refresh-scheduler가 중단되었으면 수동 재시작

**효과**: 네트워크 실패 등 예외 상황에도 robust

**리스크**: 구현 복잡도 증가, 배터리/네트워크 비용

### 6.5 [P2] Keycloak access token lifespan 연장

**대상**: `infra/idp/keycloak-realm.dev.json`

**변경 내용**: `accessTokenLifespan: 900` (15분) 또는 그 이상

**효과**: 600초 idle 버그의 재현 확률을 낮춤 (근본 해결 아님). 300s → 900s면 transient 실패 후에도 token이 15분간 유효.

**리스크**: 보안 — access token 탈취 시 피해 기간 증가. 운영 환경에서는 더 짧은 수명 권장.

---

## 7. 권장 Fix 우선순위

```
1차 (P0): refresh-scheduler transient_failed → reschedule() 추가
  └─ 최소 변경, 최대 효과. 단 1줄 추가.
  └─ 적용 후 600초 idle 테스트 통과 예상

2차 (P1): tab navigation `<Link>` prefetch 패턴 검토
  └─ 근본적인 설계 개선. 서버 의존성 제거.
  └─ router.replace()가 필요한 다른 페이지에도 영향 분석 필요

3차 (P2): idle watchdog + token keepalive
  └─ 추가 안전장치. refresh-scheduler가 어떤 이유로 중단되어도 복구.

4차 (P2): Keycloak token lifespan 연장
  └─ 임시 방편. 근본 해결 아님.
```

---

## 8. 검증 방법

P0 fix 적용 후 동일 E2E 시나리오로 재현:

```
1. 로그인 (charlie@example.com / ChangeMe-12345!)
2. /admin/catalog 이동
3. 600초 대기 (idle)
4. "Repositories" 탭 클릭
5. URL /admin/catalog?tab=repositories 확인
6. 콘솔 에러 없음 확인
```

---

## Appendix A — 추적한 Next.js 내부 소스 파일

| 파일 | 역할 | 주요 발견 |
|------|------|----------|
| `app-router-instance.js` | `router.replace()` → `startTransition()` → `dispatchNavigateAction()` | Transition wrapper로 동작 |
| `use-action-queue.js` | `dispatchAction` → deferredPromise + `setState(deferredPromise)` → Router 컴포넌트 suspend | 동일 state 참조 시 skip |
| `fetch-server-response.js` | RSC 서버 요청 — `fetch(fetchUrl, { credentials: 'same-origin' })` | **apiClient 미경유**, Authorization 헤더 없음 |
| `segment-cache/navigation.js` | `navigateToUnknownRoute()` → `.catch(() => state)` | **silent drop 지점** |
| `app-router.js` | Router reducer — `completeSoftNavigation` vs `completeHardNavigation` (MPA redirect) | |

## Appendix B — console-2026-05-28T07-57-30-970Z.log 분석

| T(초) | 이벤트 |
|-------|--------|
| 10 | 401 on `/api/v1/me` → refresh 성공 (recovered) |
| 36 | RealtimeService reconnect |
| 46 | RealtimeService reconnect |
| 140 | RealtimeService reconnect |
| 278 | RealtimeService reconnect |
| 285 | RealtimeService reconnect |
| 600-660 | **에러 없음** — navigation silent failure 확인 |

600-660s 구간에 아무 에러도 없는 것이 이 버그의 핵심 특징: Next.js가 내부에서 에러를 완전히 삼킴.
