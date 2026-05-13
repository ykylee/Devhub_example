# ADR-0003: No-Docker 정책의 CI 범위 명문화 — `services: postgres:` 제거

- 문서 목적: DevHub 의 "No-Docker" 정책 (2026-05-07 결정, [feedback_no_docker memory](../../ai-workflow/memory/PROJECT_PROFILE.md) §정책) 이 CI 환경까지 적용됨을 명시하고, GitHub Actions 의 `services: postgres:15` sidecar 컨테이너를 native PostgreSQL 설치로 교체하는 결정을 기록한다.
- 범위: `.github/workflows/ci.yml` 의 E2E 잡 의 PostgreSQL 기동 방식, no-docker 정책 문구의 범위 해석.
- 대상 독자: Backend 개발자, AI Agent, 시스템 관리자, CI 작업자.
- 상태: accepted
- 결정일: 2026-05-13
- 관련 문서: [ADR-0001 IdP 선정](./0001-idp-selection.md) §2.5 (no-docker 정책 1차 진술), [test-server-deployment.md §1.2](../setup/test-server-deployment.md) (prod PostgreSQL 15), [PR #86 review comment](https://github.com/ykylee/Devhub_example/pull/86#issuecomment-4435695605) (FU-CI-1 발견 시점).

## 1. 컨텍스트

### 1.1 현재 정책 진술

`feedback_no_docker` 메모리 (2026-05-07 origin) 와 ADR-0001 §2.5 는 다음과 같이 명시:

> DevHub 서비스 전체 (backend-core, backend-ai, frontend, IdP 등 **모든 외부 컴포넌트 포함**) 는 Docker / docker-compose / 컨테이너를 사용하지 않는 것을 전제로 구성하고 개발한다. 새 컴포넌트 도입 시에도 컨테이너 이미지/composer 경로를 가정하지 말고 native binary 또는 호스트 시스템 서비스 경로로 계획한다.

2026-05-08 PR #12 리뷰 시 사용자가 한 단계 구체화:

> docker 같은 개발 환경 셋업은 작업 환경마다 제약이 다르므로 작업 환경에 특화된 자산은 git 에 커밋하지 않고, git 에는 docker / no-docker 분기 가이드만 둔다.

→ **컨테이너 자산은 git 에 커밋하지 않는다** 가 정책의 운영 기준.

### 1.2 위반 지점

PR #86 (`gemini/prepare-github-action`, 2026-05-13 머지) 가 도입한 `.github/workflows/ci.yml` E2E 잡에 다음 블록이 포함:

```yaml
services:
  postgres:
    image: postgres:15
    env:
      POSTGRES_USER: runner
      POSTGRES_DB: devhub
      POSTGRES_HOST_AUTH_METHOD: trust
    ports:
      - 5432:5432
    options: >-
      --health-cmd pg_isready
      ...
```

`services: postgres:15` 는 GitHub Actions 가 ephemeral Docker sidecar 컨테이너를 띄우는 메커니즘. 이는 다음을 위반한다:

1. **"모든 외부 컴포넌트 포함"** — PostgreSQL 은 외부 컴포넌트이고, CI 도 운영 환경의 한 단면이다.
2. **"컨테이너 자산은 git 에 커밋하지 않는다"** — `services:` 선언이 git-tracked `ci.yml` 안에 존재.

PR #86 리뷰어 모드 2-pass 의 코멘트에서 본 항목을 "Non-blocker / 후속 논의용 FU-CI-1" 로 분류하여 별도 sprint 로 인계했다. 본 ADR 이 그 결정을 처리한다.

## 2. 결정 동인 (Decision drivers)

1. **정책 일관성** — 메모리 본문이 "전체 / 모든 외부 컴포넌트 포함" 으로 명시. 예외를 만드는 것은 정책 텍스트를 약화시킨다.
2. **버전 정합성** — `test-server-deployment.md §1.2` 의 prod 환경은 PostgreSQL 15. CI 도 동일 버전이 운영 등가성 보장.
3. **재현성** — git-tracked 셋업 스크립트만으로 CI 환경이 결정되어야 함. Docker 사용 시 GitHub 의 service container 런타임에 의존하게 됨.
4. **사내 SSL inspection 환경 대응** — prod 에서 사내 corp 환경 binary mirror 가능성을 메모리가 명시. Docker 이미지 mirror 와 별개 경로.
5. **CI 시간 영향 최소화** — apt install postgresql-15 의 추가 시간은 30–60 초 수준으로 e2e 잡의 총 4–6 분 중 미미.

## 3. 검토한 옵션

### 3.1 Option A — 정책을 CI 로 명시 확장, native PostgreSQL 15 (선정)

ubuntu-latest 러너에서 PostgreSQL 공식 apt repo (`apt.postgresql.org`) 를 추가하고 `postgresql-15` 를 설치, `systemctl start postgresql` 로 기동. `runner` 유저 + `devhub` DB 생성, `pg_hba.conf` 의 localhost TCP entries 를 `trust` 로 설정 (CI 전용 — 외부 노출 없음).

- 장점:
    - 정책 텍스트와 100% 일치.
    - prod 와 동일 PostgreSQL 15.
    - git-tracked `.github/workflows/ci.yml` 에 모든 셋업이 명시되어 재현 가능.
- 단점:
    - CI 잡 시간 ~30–60 초 추가 (apt install).
    - workflow yml 에 셸 라인 ~15줄 추가.

### 3.2 Option B — 정책 단서 추가, CI ephemeral sidecar 예외

`feedback_no_docker` 메모리에 "단, GitHub Actions service container 같은 CI ephemeral sidecar 는 예외" 를 명시. `ci.yml` 그대로 유지.

- 장점: 코드 변경 0, 추가 시간 0.
- 단점:
    - 정책 텍스트 약화 ("전체 / 모든" → "전체 / 모든, 단 ...").
    - "어디까지가 ephemeral 인지" 가 모호 — 향후 sidecar 가 늘어날 때마다 ADR 필요.
    - 정책 결정의 spirit (사내 SSL inspection 환경에서 Docker 의존 차단) 와 거리감.

### 3.3 Option C — third-party test dep 만 예외

DevHub 자체 서비스 (backend-core, backend-ai, frontend, IdP) 는 native, 그 외 "test 의존성" (postgres 같은 데이터스토어) 만 컨테이너 허용.

- 장점: 운영성과 정책 사이 절충.
- 단점:
    - "test 의존성" 의 정의가 주관적. PostgreSQL 은 prod 의존성이기도 함.
    - 두 카테고리 사이 경계가 향후 모호해질 위험 (예: Redis, Kafka 가 CI 에 추가되면?).
    - Option A 대비 정책 텍스트 분기.

## 4. 결정

**Option A 선정** — no-docker 정책이 CI 까지 적용됨을 본 ADR 로 명문화하고, `.github/workflows/ci.yml` 의 `services: postgres:15` 블록을 제거. PostgreSQL 15 를 pgdg apt repo 로 native 설치하는 step 으로 교체.

본 ADR 의 accepted 와 동시에 `ai-workflow/memory/feedback_no_docker` 메모리는 별도 변경 없이 그대로 유효 (CI 가 "DevHub 전체" 의 일부임을 본 ADR 이 해석으로 못 박음).

## 5. 결과 (Consequences)

### 5.1 즉시 변경

- `.github/workflows/ci.yml` 의 e2e 잡:
    - `services: postgres:15` 블록 제거.
    - 새 step `Setup PostgreSQL (native)`: pgdg repo 추가 → `apt-get install postgresql-15` → `systemctl start` → `runner` SUPERUSER + `devhub` DB 생성 → `pg_hba.conf` host entries 를 trust 로 변경 → restart → `pg_isready` 대기.
- 본 ADR 파일 (`docs/adr/0003-no-docker-policy-ci-scope.md`) 추가.
- `docs/setup/test-server-deployment.md` 에 본 ADR 링크 추가 (CI 등가성 메모).
- `ai-workflow/memory/state.json#next_actions.ci_followups` 에서 FU-CI-1 항목 제거.

### 5.2 후속 영향

- 향후 CI 에 새 데이터스토어 (Redis, ClickHouse 등) 가 필요해지면 본 ADR 의 패턴 (apt install + systemctl) 따라 추가. `services:` 사용 금지.
- 사내 corp 환경에서 CI 를 self-hosted runner 로 전환할 가능성 시, pgdg repo mirror 또는 사전 install 이미지 사용 검토.
- `services:` 사용을 막는 workflow 린트 (Hadolint, actionlint) 도입은 별도 ADR 후보.

### 5.3 검증

- 본 ADR 과 함께 푸시되는 PR 의 CI 그린 = 검증 완료. 3 잡 (backend-unit / frontend-unit / e2e) 모두 SUCCESS 확인.

## 6. 미해결 항목 (Open issues)

- **사내 corp 환경 self-hosted runner 도입 시점** — 본 ADR 은 GitHub-hosted ubuntu-latest 기준. self-hosted 전환은 별도 결정.
- **actionlint / workflow 린트 도입** — `services:` 정책 위반을 정적 검사로 잡는 게이트. 후속 ADR 후보.
