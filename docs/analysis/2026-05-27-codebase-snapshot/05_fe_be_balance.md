# 05. 프론트엔드 ↔ 백엔드 개발 균형 분석

- 문서 목적: FE/BE 의 개발 진척·완성도·테스트·최신 기능 노출 균형을 비교하고 불균형 지점과 시정 방향을 도출한다.
- 기준: main `cf19c94`.

---

## 1. 정량 비교

| 지표 | 백엔드 | 프론트엔드 | 해석 |
| --- | --- | --- | --- |
| 구현 파일 | 89 (Go) | 50 컴포넌트 + 18 서비스 + 33 페이지 | 규모 균형 양호 |
| HTTP 표면 | ~100 라우트 (API-01..90) | 18 서비스가 소비 | 서비스가 API 대부분 래핑 |
| 단위 테스트 | **72 파일** | **10 파일** | 🔴 큰 불균형 |
| E2E/통합 | go integration test 다수 | **28 Playwright spec** | E2E 는 FE 가 담당(균형 양호) |
| 도메인 커버 | 8 도메인 모두 store/handler | 8 도메인 모두 page/service | 도메인 폭 균형 |

## 2. 도메인별 FE/BE 정합 매트릭스

| 도메인 | Backend | Frontend | 균형 평가 |
| --- | :-: | :-: | --- |
| 인증/세션 | ✅ | ✅ | 균형 |
| 조직/사용자/RBAC | ✅ | ✅ | 균형 |
| 감사 | ✅ | ✅ | 균형 |
| Onboarding | ✅ | ✅ | 균형 (동시 완성) |
| Application/Project | ✅ | ✅ | 균형 |
| DREQ | ✅ | ✅ | 균형 (M5 closing) |
| External Integration (provider/binding) | ✅ | ✅ | 균형 |
| **Repository draft/publish + SCM 양방향** | ✅(앞섬) | 🟡 | **BE 선행** — UI 는 따라왔으나 happy-path E2E 후행 |
| **auth_mode / api_token / base_url 깊이** | ✅(앞섬) | ✅ | UI 동적 입력 따라옴(균형 회복) |
| **Gitea SCM pull sync worker** | ✅ | — (운영 view 없음) | **BE only** — Runner/sync 상태 UI 없음(RM-M4-07) |
| **Realtime event publish (infra/ci/risk)** | 🟡 | 🟡 | 양쪽 미완(command 만) |
| **AI Gardener** | 🔴 | 🔴(suggestion UI placeholder) | 양쪽 v2 |
| **평문 secret 암호화** | 🟡 | n/a | BE 보안 carve |

## 3. 불균형 진단

### 3.1 백엔드 선행 (FE 후행)

2026-05-26~27 의 작업(Gitea sync, SCM 양방향, auth_mode, draft/publish)은 **백엔드 주도로 빠르게 확장**됐고, UI 는 모달/테이블 수준으로 따라왔으나:

- **운영 가시성 부족**: Gitea sync 워커의 큐/상태, integration_sync_jobs 진행 현황을 보는 **System Admin 운영 대시보드(RM-M4-07)가 없다**. 백엔드는 데이터를 만들지만 운영자가 볼 화면이 없다.
- **happy-path E2E 후행**: repository draft→publish, SCM import/create 는 negative-path E2E 위주로만 커버.

### 3.2 테스트 밀도 불균형 (가장 큰 부채)

- 백엔드 72 test 파일 vs 프론트 10 vitest 파일. 프론트 서비스 18개 중 **단위 테스트가 있는 건 2개**(`project.service`, `integration-provider-presets`).
- E2E(Playwright 28)가 프론트 회귀를 상당 부분 방어하지만, **service/유틸 로직의 단위 회귀 가드는 얇다**. 신규 service 메서드(예: integration `createScmRepository`, repository draft) 추가 시 회귀 위험.

### 3.3 균형이 잘 잡힌 영역

- 5개 핵심 도메인(인증·조직·DREQ·Onboarding·Application)은 BE store/handler ↔ FE page/service 가 1:1 로 정합. Onboarding 은 Carve A(BE)→B/C(FE)→D(tests) 가 순차 완성되어 모범 사례.

## 4. 시정 방향 (우선순위)

| # | 시정 항목 | 영역 | 우선순위 |
| --- | --- | --- | --- |
| B1 | **프론트 service/component 단위테스트 보강** (integration·repository·dashboard service + 신규 모달) | FE | 높음 |
| B2 | **최신 backend 기능 happy-path E2E** (repository draft→publish, SCM import/create) | FE+BE | 높음 |
| B3 | **System Admin 운영 대시보드** (Gitea sync job 큐/상태 + provider health) — BE 데이터 ↔ FE view 균형 회복 | FE+BE | 중간 (RM-M4-07) |
| B4 | **#368 draft→publish backend UT 보강** (무테스트 머지분) | BE | 중간 |
| B5 | Realtime event publish(infra/ci/risk) + FE 구독 UI 동시 진행 | FE+BE | 중간 (RM-M4-01/02) |
| B6 | 평문 secret envelope 암호화 (BE 단독) | BE | 중간 (#6) |

## 5. 결론

- **폭(도메인 커버리지)** 은 FE/BE 균형이 좋다. 8개 도메인 모두 양쪽에 존재한다.
- **깊이(최신 기능)** 는 2026-05 후반 들어 **백엔드가 약간 앞선다** — SCM/Gitea/auth_mode 확장이 backend 주도였고, FE 는 입력 UI 까지는 따라왔으나 운영 가시성·E2E 가 뒤처졌다.
- **테스트** 는 명확히 불균형 — backend 72 vs frontend 10. **프론트 단위테스트 보강(B1)이 균형 회복의 1순위**.
- 권장: 다음 사이클은 **신규 기능보다 "따라잡기"(E2E happy-path + FE 단위테스트 + 운영 대시보드)** 에 무게를 두어 깊이/테스트 균형을 회복한다.
