# ADR-0031: build tag 정책 재검토 (v0.1.1 sprint -a follow-up PR1 PR #540 carry-over C-j)

- **문서 목적**: [ADR-0030 §2.3](./0030-sso-integrations-and-auth-session-port.md) 의 **runtime injection (옵션 2) 채택 결정** 을 현시점 (2026-06-12, e2e-internal job 폐기 결정, sprint `fix/work_260612-6-e2e-internal-removal`) 에서 재평가하고, build tag (옵션 1) 로의 전환 trade-off 를 정량적으로 측정하여 본 결정의 지속 / 갱신 / supersession 을 명문화한다.
- **범위**: ADR-0030 §2.3 의 두 옵션 (옵션 1 build tag / 옵션 2 runtime injection) 의 cost 측정 + 현시점 결정 + 재검토 trigger 정의. **코드 변경 0**. 본 ADR 은 결정의 re-evaluation 만, 신규 구현 변경 없음.
- **대상 독자**: Backend / IdP / ops 트랙 담당자, 후속 sprint 작업자, owner, CI maintainer.
- **상태**: accepted (2026-06-10, **partial supersession — 2026-06-12 baseline 변경 정공법**)
- **최종 수정일**: 2026-06-12 (e2e-internal job 폐기에 따른 baseline 갱신, sprint `fix/work_260612-6-e2e-internal-removal`)
- **결정 근거 sprint**: `docs/work_260610-c-j-build-tag-review` (PR 의 sprint) + `fix/work_260612-6-e2e-internal-removal` (본 baseline 갱신 sprint)
- **Tier**: **공용** (ADR 만, 사내 한정 정보 미포함)
- **관련 문서**: [ADR-0030 §2.3 runtime injection 결정](./0030-sso-integrations-and-auth-session-port.md) (본 ADR 의 re-evaluation source), [release_v0-1_roadmap.md §3.5 N-13](../planning/release_v0-1_roadmap.md) (PR #540 의 carry-over C-j 의 source), [PR #539 `feat/work_260610-v0-1-1-sprint-a-followup`](https://github.com/ykylee/Devhub_example/pull/539) (saovae_stub + main wiring), [PR #540 `feat/work_260610-v0-1-1-sprint-a-real-adapter`](https://github.com/ykylee/Devhub_example/pull/540) (real adapter), [PR #542 `feat/work_260610-c-i-e2e-internal-job`](https://github.com/ykylee/Devhub_example/pull/542) (C-i E2E Internal job — **2026-06-12 폐기 결정, sprint `fix/work_260612-6-e2e-internal-removal`**).

## 1. 배경

### 1.1 ADR-0030 §2.3 의 결정

ADR-0030 §2.3 (2026-06-10 채택) 의 두 옵션:

| # | 옵션 | 장점 | 단점 | 결정 |
| --- | --- | --- | --- | --- |
| 1 | **Build tag** (`//go:build saovae \|\| dev`) | binary 2개, 명확한 분리 | CI / build matrix 복잡 | ❌ |
| 2 | **Runtime injection** (main.go 에서 `DEVHUB_BUILD_TIER` env var 로 분기) | 단일 binary, CI 단순, saovae default | stub 이 binary 에 항상 포함 (binary size) | ⭐ **채택** |

ADR-0030 의 근거 (2026-06-10): "옵션 2 의 단점 (binary size) 은 미미할 것, 옵션 1 의 단점 (CI / build matrix) 은 본질적 복잡도" 라고 추정. **그러나 정당 정량 측정 없음**. sprint -a follow-up PR #539 + #540 + #542 머지로 **현시점에서 정량 측정 가능** → 본 ADR 의 의의.

### 1.2 sprint -a follow-up 의 적용 결과 (2026-06-10)

| PR | 변경 | 영향 |
| --- | --- | --- |
| **#539** (`feat/work_260610-v0-1-1-sprint-a-followup`) | `backend-core/internal/sso-integrations/keycloak/saovae_stub.go` (NEW, 105 lines) + main.go `DEVHUB_BUILD_TIER` env var 분기 | saovae_stub 가 단일 binary 에 포함, default = stub. |
| **#540** (`feat/work_260610-v0-1-1-sprint-a-real-adapter`) | `verifier.go` (506 lines) + `admin_client.go` (456 lines) + `metrics.go` (70 lines) + 4 test file (~1,200 lines) — `sso-integrations/keycloak/` 단일 트리 | real adapter 가 stub 와 동거 (단일 트리). |
| **#542** (`feat/work_260610-c-i-e2e-internal-job`) | `.github/workflows/ci.yml` 에 `e2e-internal` job (23 step, port 8181, `DEVHUB_BUILD_TIER=internal`) 신규 → **2026-06-12 폐기** (sprint `fix/work_260612-6-e2e-internal-removal`). 사용자 결정: e2e-internal 은 사내 환경용 셋팅 (real Keycloak adapter 검증) 이므로 GitHub Actions 로 체크 불요. CI 의 1쌍 matrix 폐기. runtime injection 결정 자체는 유지 (사내 staging/prod-smoke 가 real adapter 검증). |

**즉, 현시점 (2026-06-12, e2e-internal job 폐기 후) 에서 "runtime injection 의 cost" 와 "build tag 의 cost" 를 e2e-internal 분 제외 후 정당 정량 비교 가능**. 본 ADR 의 baseline = 1쌍 (e2e shard 1/2/3 saovae_stub default 만). runtime injection 결정 = 유지.

## 2. 정량 측정 (2026-06-12, e2e-internal job 폐기 후)

### 2.1 Runtime injection (옵션 2, 현재) 의 cost

| 항목 | 측정값 | 비고 |
| --- | --- | --- |
| `sso-integrations/keycloak/saovae_stub.go` | 105 lines, 3,831 bytes | 사외 default (4 port stub + webhook handler) |
| `sso-integrations/keycloak/metrics.go` | 70 lines | JWKS stale-while-error metric (real adapter 만 사용) |
| `sso-integrations/keycloak/` 전체 | 8 file, 2,335 lines, ~70KB | real adapter + stub + metrics + 4 test file (test file 은 binary 미포함) |
| Stub 의 binary overhead | **< 5KB** (CGO_ENABLED=0 Go build, 3.8KB source) | 사외 build 시 stub code 가 binary 에 항상 link, 5KB 미만 overhead |
| Stub 의 production 위험 | **0** (stub 가 call path 에 들어가지 않음) | main.go 의 `DEVHUB_BUILD_TIER` 분기가 production (internal) 시 stub path 미사용, runtime zero overhead |
| Stub 의 운영 위험 | **0** (no external dep, no env read, no network) | saovae_stub.go 가 `os.Getenv` 등 호출 0, production binary 가 실수로 stub 경로 진입 가능성 0 |
| CI matrix | **1 axis (shard only)** — e2e shard 1/2/3 (saovae_stub default) | e2e-internal 폐기 후. PR #542 의 1쌍 matrix → 단일 matrix |
| CI runtime | e2e shard 1/2/3 (3 jobs × 4-5min) = baseline | e2e-internal 의 +15~20min 제거 |

**런타임 cost 정당 정량 합** (e2e-internal 폐기 후): stub binary overhead < 5KB + CI 0 min 추가 (e2e shard 1/2/3 만) + Keycloak container 1개 (e2e shard 공용 port 8180).

### 2.2 Build tag (옵션 1) 의 cost (이론적 측정, e2e-internal 폐기 후)

| 항목 | 측정값 | 비고 |
| --- | --- | --- |
| `sso-integrations/keycloak/saovae_stub.go` 가 production binary 에 미포함 | **-3.8KB** binary overhead 절감 | 사외 build tag `//go:build saovae \|\| dev` 가 production (real) build 시 exclude |
| `sso-integrations/keycloak/metrics.go` 가 production binary 에 미포함 | **-2.5KB** | 동일 |
| `verifier_test.go` (853 lines) + `admin_client_*_test.go` (345 lines) 가 real tag 시 production binary 에서 제외 (단 test file 은 어차피 binary 미포함) | **0** 절감 | test code 는 production binary 영향 0 |
| CI matrix | **2배** (e2e shard 1/2/3 × 2 tags = 6 jobs) | 각 shard 가 2 build + 2 deploy + 2 test invocation. 또는 e2e shard 만 2 tags (real / saovae) × 1 set 의 Keycloak container 1개 + 2 backend deploy |
| CI runtime | **+15~30min** (6 jobs vs 3 jobs) | build matrix 2배. Go build cache 가 module cache hit 시 +0min, miss 시 +5min per build |
| 코드 변경 | **5~10 file** (`//go:build` tag 5개 위치, `_test.go` tag 2개 위치, main.go 분기 단순화) | sprint -a follow-up 의 본질적 architecture 재구성. PR #539 의 runtime injection 코드를 build tag 로 변환. |
| 운영 위험 (new) | **2개 binary 배포/롤백** | staging + production 환경 변수 + Keycloak 컨테이너 + binary 2개 |

**Build tag 의 cost 정량 합** (e2e-internal 폐기 후): binary -6.3KB + CI +15~30min + 5~10 file 변경 + 2개 binary 운영.

### 2.3 비교 (e2e-internal 폐기 후)

| 측정 항목 | Runtime injection (현재) | Build tag (이론) | 차이 |
| --- | --- | --- | --- |
| Binary overhead | < 5KB | -6.3KB (절감) | -6.3KB (build tag 유리) |
| Stub production 위험 | 0 (call path 미진입) | 0 (build 시 exclude) | 0 |
| CI runtime | e2e shard 1/2/3 (baseline) | +15~30min (e2e × 2 tags = 6 jobs) | build tag 가 +15~30min 더 |
| CI matrix jobs | 3 (e2e 1/2/3) | 6 (e2e × 2 tags) | build tag 가 +3 jobs |
| 코드 변경 | 0 (현재 상태 유지) | 5~10 file (`//go:build` tag) | build tag 가 +5~10 file |
| 운영 복잡도 | 1 binary | 2 binary | build tag 가 +1 |

**결론 (정당 정량, e2e-internal 폐기 후)**: build tag 의 binary size 절감 (~6KB) 은 무시 가능 수준 (전체 backend-core binary < 50MB 대비 0.01% 미만). **runtime injection 의 cost 가 build tag 의 cost 보다 본질적으로 작음**. e2e-internal 폐기 후 비교는 폐기 전보다 build tag 의 cost 가 더 명확히 부각됨 (이전엔 e2e-internal 1 job 이 build tag 의 1 set 와 중복 → trade-off 모호; 폐기 후 build tag 가 pure additional cost 가 됨).

## 3. 후보 옵션

### 3.1 옵션 1: Build tag 로 전환

| 장점 | 단점 |
| --- | --- |
| Production binary size -6.3KB | CI matrix 2배 (+15~30min runtime, +3 jobs) |
| Stub code 가 production binary 에 미포함 (architectural cleanliness) | 5~10 file `//go:build` tag 추가 |
| 명시적 2 build 의 architectural 정합성 | 2개 binary 운영 (배포/롤백) |
| | Stub code 가 사외 build path 로 분리 (cross-tag test 어려움) |
| | **2026-06-12 e2e-internal 폐기 결정으로 비교 baseline 변경 — build tag 의 추가 가치 더욱 부각 (이전엔 e2e-internal 1 job 이 build tag 의 1 set 와 중복)** |

**Trade-off 평가**: binary size -6.3KB vs CI +15~30min + 코드 +5~10 file + 운영 +1 binary. **불리 (1:3+ 의 cost ratio)**.

### 3.2 옵션 2: Runtime injection 유지 (현재)

| 장점 | 단점 |
| --- | --- |
| 단일 binary, 배포 단순 | Stub code 가 production binary 에 link (< 5KB overhead) |
| CI matrix 단순 (e2e shard 1/2/3 saovae_stub default 만) | Stub code 가 production binary 에 존재 (architectural cleanup 미비) |
| main.go 분기 명확 (`DEVHUB_BUILD_TIER`) | |
| 코드 변경 0 (현재 상태 유지) | |
| 사외 build 시 saovae_stub 자동 사용 — production 동작과 다른 환경에서 e2e 가능 | |
| **사내 staging/prod-smoke 가 real adapter 검증 (e2e-internal 폐기 결정 정합)** | |

**Trade-off 평가**: < 5KB binary overhead vs 0 CI/runtime/코드/운영 cost. **유리**.

### 3.3 옵션 3: Hybrid (build tag + runtime injection)

| 장점 | 단점 |
| --- | --- |
| Production binary size -6.3KB (옵션 1) | CI matrix 2배 + main.go 분기 (옵션 1 + 2 의 모든 cost) |
| Stub code 가 production binary 에 미포함 | 코드 +5~10 file + 2개 binary 운영 |
| CI 단순 (옵션 2) | **두 메커니즘 동시 사용 = complexity × 2** |

**Trade-off 평가**: 양쪽 cost 의 합 + 복잡도 × 2. **가장 불리**.

## 4. 결정

**옵션 2 (Runtime injection 유지) 채택** (2026-06-12 e2e-internal 폐기 후 재확인).

근거:
1. **정당 정량 측정 (2026-06-12, e2e-internal 폐기 후) 결과**: build tag 의 binary size 절감 (~6KB) 은 backend-core binary < 50MB 대비 0.01% 미만. **무시 가능 수준**. e2e-internal 폐기 후 비교는 build tag 의 cost 가 더 명확히 부각됨 (이전엔 e2e-internal 1 job 이 build tag 의 1 set 와 중복 → trade-off 모호; 폐기 후 build tag 가 pure additional cost 가 됨).
2. **CI runtime 영향**: e2e shard 1/2/3 (3 jobs, baseline) vs build tag 의 e2e × 2 tags (6 jobs) = +15~30min 더.
3. **코드 / 운영 영향**: build tag 전환 시 +5~10 file 변경 + 2개 binary 운영. runtime injection 의 코드 변경 0 + 1 binary 운영과 대비.
4. **현시점 측정 가능 (e2e-internal 폐기 후)**: ADR-0030 §2.3 의 결정 시점 (2026-06-10 sprint -a PR) 에는 stub code size 와 CI runtime 의 정량 측정 불가. sprint -a follow-up PR #540 (real adapter) + PR #542 (e2e-internal job) 머지로 **정량 측정 가능해졌고**, 2026-06-12 e2e-internal 폐기 결정으로 **build tag 의 trade-off 가 더 명확히 부각됨**. 측정 결과 runtime injection 의 cost 가 더 작음.
5. **stub production 위험 0**: saovae_stub 가 production (internal) build 시 call path 미진입. main.go 의 `DEVHUB_BUILD_TIER` 분기가 zero-cost.
6. **architectural cleanliness 우선순위**: 현재 DevHub 는 **architectural cleanliness** (단일 binary, 단일 CI matrix) 를 **binary size** (6KB) 보다 우선시. e2e-internal 폐기 후 cost ratio 1:3+ 일관.
7. **e2e-internal 폐기 결정 정합 (2026-06-12)**: 사용자는 e2e-internal 을 "사내 환경용 셋팅" 이라 명시. CI 의 1쌍 matrix 폐기 = **runtime injection 결정은 유지하되 CI 의 1 axis (build tier) 가 사라짐**. 사내 staging/prod-smoke 가 real adapter 검증 책임. **이 결정은 본 ADR §5 재검토 trigger 3번 (CI matrix 5+ axes) 의 정반대 방향 — CI axis 축소**.

**결론**: ADR-0030 §2.3 의 옵션 2 (runtime injection) 채택 결정을 **현시점에서 confirmed (재확인)**. 본 ADR 은 ADR-0030 을 supersede 하지 않음. ADR-0030 §2.3 의 옵션 2 row 에 본 ADR reference 추가. e2e-internal 폐기는 runtime injection 결정과 독립 — runtime injection 의 사내 staging/prod-smoke 적용은 유지.

## 5. 재검토 trigger (future)

본 결정의 재검토가 필요한 시점 (= build tag 로의 전환이 trade-off 우위로 바뀌는 시점):

| Trigger | Threshold | 영향 |
| --- | --- | --- |
| **Stub code size 가 production binary 대비 > 0.5%** | 현재 ~6KB / 50MB = 0.012%. threshold = 250KB. | Stub code 가 40× 커져야 함. 즉, saovae_stub + 39개 추가 stub (다른 IdP / 외부 시스템) 등. |
| **Stub 가 production runtime path 에 진입할 위험** | main.go 의 `DEVHUB_BUILD_TIER` 분기 + 운영 SOP 가 모두 미흡. | 즉, stub 가 실수로 production 에서 호출 가능. **현시점 위험 0**. |
| **CI matrix 가 5+ axes 로 확장** | **2026-06-12 기준 1 axis (shard only)** — e2e-internal 폐기 후. threshold = 5 axes. | e2e shard 가 아닌 다른 변종 (예: multi-region, multi-IdP) 가 추가되어 build matrix 5+ 차원이 되면 build tag 가 유리. |
| **Phase 2 (v1.2) agentic RAG 가 ssoKeycloakPort 외 추가 port 필요** | 본 ADR 의 trade-off 가 다른 port (e.g. git port, CI port) 와 함께 결정되어야 함. | 새 ADR 필요. |
| **Stub 자체가 production 위험** (예: stub 가 본질적으로 production 환경에서 호출 가능) | **현시점 0**. main.go 의 `DEVHUB_BUILD_TIER` 분기가 zero-cost. | trigger 시 build tag 전환. |

**현시점 trigger 0건**. 결정 유지.

## 6. Cross-tier impact

### 6.1 tier 매핑

| Module | Tier | 비고 |
| --- | --- | --- |
| `sso-integrations/keycloak/saovae_stub.go` | **사외** (runtime default) | production (internal) build 시 call path 미진입, runtime zero overhead |
| `sso-integrations/keycloak/verifier.go` + `admin_client.go` | **사내** (runtime internal) | production 시 wire. stub 와 단일 트리 동거. |
| `main.go` (DEVHUB_BUILD_TIER 분기) | **공용** (env var read, no 사내 한정 패턴) | runtime injection 의 분기. |
| `scripts/ci-e2e-sync-check.sh` | **공용** | DEVHUB_BUILD_TIER token 미포함 (e2e shard 1/2/3 env block 미설정 유지, e2e-internal 폐기 결정 정합) |

### 6.2 .gitignore / CI / ADR 업데이트 필요

- **CI**: 추가 변경 없음. e2e shard 1/2/3 (saovae_stub default) 가 현시점 optimal. e2e-internal 폐기 결정 (2026-06-12, sprint `fix/work_260612-6-e2e-internal-removal`) 정합.
- **ADR**: 본 ADR (ADR-0031) 의 §5 재검토 trigger 정의로 future 갱신 trigger 명시. 본 sprint = **partial supersession (baseline 변경 정공법)**, ADR-0030 자체는 supersede X.
- **tier lint** (`scripts/check-tier-separation.sh`): 본 ADR 결정 = runtime injection 유지 = **변경 없음** (PR #541 의 lint 정합 유지).

## 7. Risks + Open questions

### 7.1 Risks

1. **Stub code size 가 예상보다 빠르게 증가**: 다른 IdP (Auth0, Okta 등) 또는 외부 시스템 stub 가 추가될 경우. **현시점 0** (단일 Keycloak stub). 본 ADR §5 trigger 1번 threshold 가 이 risk 의 trigger.
2. **운영 SOP 누락**: production 환경에서 `DEVHUB_BUILD_TIER` env var 미설정 시 stub 가 default 로 wire 될 위험. **현시점 0** (PR #539 의 main.go 가 default = stub 이지만 production SOP 가 `DEVHUB_BUILD_TIER=internal` 강제). 본 ADR §5 trigger 2번 threshold 가 이 risk 의 trigger.
3. **CI matrix 5+ axes 확장 시 본 결정의 한계**: **2026-06-12 기준 1 axis (shard only)** — e2e-internal 폐기 후. multi-region, multi-IdP 등 추가 시 build tag 가 유리. 본 ADR §5 trigger 3번.
4. **e2e-internal 폐기 결정의 운영 위험 (new)**: e2e shard 1/2/3 의 saovae_stub default 검증만으로 real adapter 의 production 회귀를 미검증. **사내 staging/prod-smoke 가 real adapter 검증 책임** (e2e-internal 폐기 결정의 의도). 본 risk 는 e2e-internal 폐기 정공법 정합 — 별도 risk 아님.

### 7.2 Open questions

1. **Phase 2 (v1.2) agentic RAG 의 ssoKeycloakPort 외 다른 port 도입 시 본 ADR 의 trade-off 재평가 필요**: 본 ADR 은 ssoKeycloakPort 만 다룸. 다른 port (git port, CI port, hrdb port) 가 도입되면 별도 ADR 필요.
2. **사내 운영 SOP 의 `DEVHUB_BUILD_TIER=internal` 강제 검증 자동화**: `docker-compose.deploy.yml` 의 env block 에 `DEVHUB_BUILD_TIER=internal` 가 강제되는지 lint 자동화. 현시점 manual. 자동화 시 별도 PR.

## 8. 결정 supersession

본 ADR 은 **ADR-0030 을 supersede 하지 않음**. ADR-0030 §2.3 의 옵션 2 (runtime injection) 결정을 confirmed (재확인). 본 ADR 의 의의 = 정당 정량 측정 + 재검토 trigger 명시.

**2026-06-12 partial supersession (baseline 변경 정공법)**: e2e-internal job 폐기 (sprint `fix/work_260612-6-e2e-internal-removal`) 가 본 ADR 의 **측정 baseline 을 변경** (CI matrix 1쌍 → 단일 matrix). 본 ADR 의 결론 (옵션 2 runtime injection 유지) 자체는 변동 없음 — runtime injection 결정과 e2e-internal job 폐기는 **독립 결정**. runtime injection 의 사내 staging/prod-smoke 적용은 유지.

후속 trigger 발효 시 새 ADR (e.g. ADR-0032 "build tag 전환") 필요.

## 9. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-06-10 | 1차 작성 — ADR-0030 §2.3 의 runtime injection (옵션 2) 결정을 sprint -a follow-up PR #540 (real adapter) + PR #542 (e2e-internal job) 머지 후 정량 측정. binary overhead < 5KB vs CI matrix 2배. 결론: runtime injection 유지. §5 재검토 trigger 5건 정의 (size / prod risk / axes / Phase 2 / stub safety). ADR-0030 을 supersede 하지 않음 (confirmed). | `docs/work_260610-c-j-build-tag-review` |
| 2026-06-12 | **partial supersession (baseline 변경 정공법)** — e2e-internal job 폐기 결정 (사용자: "사내 환경용 셋팅, GitHub Action 으로 체크 불요") 정합. ci.yml e2e-internal job (PR #542, +202 lines) + scripts/ci-e2e-sync-check.sh DEVHUB_BUILD_TIER 코멘트 정리. 본 ADR §1.2 / §2.1 / §2.2 / §2.3 / §3.1 / §3.2 / §4 / §5 / §6.1 / §7.1 / §8 모두 baseline 변경 정공법 (e2e-internal 분 제외 + 결론 자체는 유지). **결론 변동 0건** (runtime injection 결정 = confirmed, e2e-internal 폐기 = 독립 결정). `fix/work_260612-6-e2e-internal-removal` (본 sprint). | `fix/work_260612-6-e2e-internal-removal` |
