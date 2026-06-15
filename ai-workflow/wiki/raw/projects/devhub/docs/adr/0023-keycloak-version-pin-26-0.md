# ADR-0023: Keycloak 운영 image 버전 pin (26.0) — ADR-0022 25.0 retreat reversal

## 1. 상태
- **상태**: Accepted
- **작성일**: 2026-05-26
- **수정일**: 2026-05-26
- **결정 근거 sprint**: `claude/work_260526-adr-0023-keycloak-26-forward`
- **Supersedes**: [ADR-0022 (Keycloak 25.0 pin)](./0022-keycloak-version-pin-25-0.md)
- **Tier**: 사내
- **관련 문서**: [ADR-0019 Keycloak 단일화](./0019-keycloak-only-idp.md), [`docker-compose.deploy.yml`](../../docker-compose.deploy.yml), [`docs/setup/keycloak_operations.md`](../setup/keycloak_operations.md). ADR governance 패턴: `feedback_adr_supersession_pattern` (claude memory, immutable history 원칙).

## 2. 컨텍스트

### 2.1 ADR-0022 의 retreat 결정 (이전 상태)

[ADR-0022 (2026-05-22, Draft)](./0022-keycloak-version-pin-25-0.md) 가 Keycloak 운영 image 를 `26.0` → `25.0` 으로 retreat 한 결정을 명문화. 사유는 §3.1 placeholder ("사내 정합 사유, 사용자 finalize 필요") 로 보류 상태. active code 3 위치는 25.0 으로 변경 commit (`42b18b1`, sprint `codex/work_260521-c-db-docker-option`).

### 2.2 reversal 식별 시점

2026-05-26 사내 운영 환경 **재확인 결과 26.x 사용 가능** 확정. ADR-0022 §2.3 의 retreat 식별 시점 (사내 deploy smoke test 진행 중 25.0 retreat 결정) 의 판단이 후속 재테스트로 정정됨.

직접적 reversal 동인:
- 사내 환경 (외부 100.90.113.29:13000 → 호스트 → VM → docker 단일 포트 ingress) 의 26.x 재테스트 PASS
- ADR-0022 §3.1 retreat 사유 placeholder 가 finalize 되지 못한 상태에서 사용 가능 재확인 → retreat 자체 무효화
- ADR-0019 §5.3 사내 동반 carve (Keycloak group staging-prod) 의 26.x 정합 확인 ([PR #306 verify-keycloak-groups.sh](https://github.com/ykylee/Devhub_example/pull/306) — 25.0 / 26.0 무관 동작)

### 2.3 reversal 절차의 ADR governance

`feedback_adr_supersession_pattern` (claude memory) 정공법:

- ADR-0022 본문 partial 수정 금지 (immutable history 원칙)
- 본 ADR-0023 신규 발행으로 결정 reversal 명문화
- ADR-0022 메타 헤더 + §0 inline banner 로 supersession 표기, 본문 immutable 보존
- 결정 시점의 snapshot 보존 — historical 26.0 retreat → 25.0 시점 검토 정보 유지

## 3. 결정

**Keycloak 운영 image 를 `quay.io/keycloak/keycloak:26.0` (마이너 pin) 으로 환원**한다. ADR-0022 의 25.0 retreat 결정은 superseded.

### 3.1 reversal 사유

사내 운영 환경 재확인 결과 Keycloak 26.x 사용 가능 확정. ADR-0022 의 retreat 식별 시점 판단 (smoke test 단계의 호환성 우려) 이 후속 재테스트로 정정됨.

본 결정은 ADR-0019 (Keycloak 단일 IdP, 2026-05-19) 의 운영 image 26.x LTS 후보 잠정 채택을 환원한다 — 즉 ADR-0019 의 본래 운영 자산 (sprint `claude/work_260519-k`, 2026-05-19 commit) 으로 forward 정합.

### 3.2 pin 등급

마이너 pin (`26.0`) 채택. 26.0.x 의 patch 자동 흡수 + 27.x 메이저 변경 차단의 balance. ADR-0022 의 25.0 마이너 pin 결정 형식은 보존 (등급 정책 유지).

### 3.3 영향 범위 (active code only)

본 ADR-0023 commit 시점에 active code 3 위치를 `25.0` → `26.0` 으로 정정한다:

| 위치 | 직전 (ADR-0022) | 본 ADR-0023 |
| --- | --- | --- |
| `docker-compose.deploy.yml:106` | `quay.io/keycloak/keycloak:25.0` | `quay.io/keycloak/keycloak:26.0` |
| `.github/workflows/ci.yml:348` | CI E2E job image `25.0` | `26.0` |
| `dev-up.sh:121` | dev 안내 echo `25.0` | `26.0` |

ADR-0022 의 §3.3 영향 범위 표는 본문 immutable 보존 (결정 시점 snapshot). historical 25.0 명시 = 2026-05-22 ~ 2026-05-26 사이 짧은 retreat window 의 실 사실 기록.

### 3.4 ADR-0022 §3.4 (port 13000 정합) 의 보존

ADR-0022 §3.4 의 "외부 ingress 포트 13000 정합" 결정은 **version pin 과 무관** 한 외부 ingress 매핑 정책. 본 reversal 은 §3.4 결정에 영향 없음:

- realm.dev.json 의 13000 6 위치 entry 보존
- `scripts/deploy-from-env.sh` 의 `PUBLIC_ACCESS_PORT:=13000` default 보존
- host:VM forward (13000 → 3000) 의 외부 인프라 의존성 보존
- ADR-0023 도 동일 port 13000 정합 적용

### 3.5 Keycloak admin bootstrap env 호환성

ADR-0022 의 `feedback_keycloak_25_26_admin_env` 패턴 — `KEYCLOAK_ADMIN` + `KEYCLOAK_ADMIN_PASSWORD` (legacy 25.x) + `KC_BOOTSTRAP_ADMIN_USERNAME` + `KC_BOOTSTRAP_ADMIN_PASSWORD` (26+ 표준) 동시 주입 — 은 **본 ADR-0023 에서도 유지**한다. 26.x 가 표준 `KC_BOOTSTRAP_*` 우선이나 legacy 도 동작 — 양쪽 주입은 보수적 안전망 + version 변경 silent fail 회피.

`docker-compose.deploy.yml:117-121` + `.github/workflows/ci.yml` 의 양쪽 env 주입 패턴 그대로 보존.

## 4. 결과

### 4.1 긍정

- ADR-0019 본래 운영 자산 (26.x LTS 후보) 으로 forward 정합 회복
- 26.x 의 신규 기능 (event listener SPI 변경 / Account Console UI 개선 / OIDC PKCE 보강) 사용 가능
- 25.0 EOL 우려 해소 (Keycloak 의 release cadence 상 ~6개월 후 25.x EOL 예상이었음)
- 사내 운영 표준이 26.x 로 진입할 경우 별도 ADR 발급 불필요

### 4.2 부정

- 26.x → 25.x → 26.x 의 짧은 retreat / re-advance window 가 historical record 에 남음 (ADR-0022 + ADR-0023 2 ADR 발행 부담)
- ADR-0022 의 retreat 결정에 따라 commit 된 active code 3 위치를 본 PR 에서 다시 26.0 으로 정정 — git history 의 churn

### 4.3 ADR governance

- ADR-0022 의 본문은 수정하지 않음 (immutable history 원칙 — `feedback_adr_supersession_pattern`)
- ADR-0022 메타 헤더 + §0 inline banner 로 supersession 표기
- ADR-0019 의 결정 (Keycloak 단일 IdP) 은 reverse 되지 않음 — version pin 만 정정
- 본 ADR-0023 가 ADR-0022 를 supersede, ADR-0019 와 자연 정합 (보조 명세 관계)

## 5. 후속 작업

| 항목 | 상태 | 비고 |
| --- | --- | --- |
| 본 ADR commit 시 active code 3 위치 정정 (25.0 → 26.0) | resolved (본 PR) | docker-compose / CI / dev-up.sh |
| ADR-0022 메타 헤더 + §0 inline banner 추가 | resolved (본 PR) | immutable 본문 보존 |
| Keycloak 26.0 image pull + redeploy smoke | pending | 사내 운영자 (staging) |
| 26.0 의 realm import 정합 검증 (`infra/idp/keycloak-realm.dev.json` + `prod.json`) | pending | 26.x 호환 — ADR-0022 §5 의 25 검증과 별개로 26 재검증 |
| `--proxy-headers=xforwarded` 동작 검증 | pending | 26.x 에서도 동일 syntax 유효 확인 |
| 26.0 patch 정기 bump carve out | open | 보안 패치 follow 절차 |
| PR #306 verify-keycloak-groups.sh 의 26.x 정합 재실행 | pending | 사내 admin console 작업 후 PASS 확인 |

## 6. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-26 | 1차 발행 (Accepted). ADR-0022 supersede. active code 3 위치 25.0 → 26.0 정정 + ADR-0022 메타 헤더 + §0 inline banner. | `claude/work_260526-adr-0023-keycloak-26-forward` |
