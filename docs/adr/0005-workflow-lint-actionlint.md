# ADR-0005: GitHub Actions workflow lint (actionlint) CI 잡 도입

- 문서 목적: `.github/workflows/*.yml` 의 syntax / 보안 / 일관성 회귀를 머지 직전에 차단하기 위해 `actionlint` 잡을 CI 에 도입한다.
- 범위: CI workflow lint. backend / frontend / e2e 코드 lint 는 본 ADR 범위 밖.
- 대상 독자: CI/CD 담당자, Backend / 프론트엔드 개발자, AI agent.
- 상태: accepted
- 결정일: 2026-05-13
- 결정 근거 sprint: `claude/work_260513-i`.
- 관련 문서: [ADR-0003](./0003-no-docker-policy-ci-scope.md) §6 (후속 ADR 후보), [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml).

## 1. 컨텍스트

PR #86 ~ #94 동안 `.github/workflows/ci.yml` 에 5+ blocker (DSN 누락, build path 오류, Hydra cipher 길이, services: postgres 제거 후 native PG dropcluster, FU-CI-2/3/4 cache 효율 등) 가 발견됐고, 모두 리뷰어 모드 2-pass 에서 잡혔다. workflow YAML 의 syntax / 보안 / version-pin 회귀를 작성 시점에 차단하지 못해 매 PR 마다 1-2 회 CI 실패 → 보강 commit 사이클을 거쳤다.

`actionlint` 는 GitHub-recommended 표준 lint 도구로 다음을 검출한다:
- workflow YAML schema 오류 (잘못된 key, indentation, type mismatch).
- shell 명령의 shellcheck 통합 (run 블록 내 `$VAR` 누락 따옴표, unused 변수 등).
- `uses:` 의 미사용/typo 액션 이름, `secrets.*` 의 누락 검출.
- expression syntax (`${{ ... }}`) 의 부정확한 함수 호출.

ADR-0003 §6 의 미해결 항목 ("workflow YAML lint job 도입 여부") 의 후속 결정이다.

## 2. 결정 동인

- **회귀 차단**: workflow YAML 회귀의 cost (CI 실패 → 보강 commit → 재실행) 가 PR 당 5-15 분. lint 잡이 syntax 단계에서 차단하면 사이클 1회 단축.
- **보안**: pinned action version 미준수, 사용 안 되는 secrets 노출 같은 잠재적 보안 회귀 검출.
- **runtime 비용 낮음**: actionlint 잡은 ubuntu-latest + `raven-actions/actionlint@v2` 로 30초 이내 종료.
- **격리성**: 다른 잡 (backend / frontend / e2e) 과 병렬 실행 — 빨라야 할 backend-unit job 의 wallclock 영향 0.

## 3. 검토한 옵션

| 옵션 | 설명 | 평가 |
| --- | --- | --- |
| **A. (선택)** `raven-actions/actionlint@v2` 별도 잡 | 새 잡 `workflow-lint` 추가, 다른 잡과 병렬. | 격리성 + 작은 실행 비용. 채택. |
| B. backend-unit 잡 내부 step | 별도 잡 분리 없이 backend-unit step 으로 흡수. | 다른 잡 실패가 lint 실패를 가릴 수 있음 + backend 잡의 cache 키와 무관. 거부. |
| C. pre-commit hook | local 단계에서 차단. | CI 게이트가 아니라 local 누락 가능. 별도 후속 hygiene 후보. |
| D. dependabot 의 action 버전 관리 + lint 미도입 | version pin 만으로 충분. | dependabot 은 update PR 을 만들지만 작성 시점 lint 는 아님. 보완재로 별도 후속 후보. |

## 4. 결정

옵션 A 채택 — `raven-actions/actionlint@v2` 별도 잡 `workflow-lint` 신설.

### 4.1 잡 정의

```yaml
workflow-lint:
  name: Workflow Lint (actionlint)
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run actionlint
      uses: raven-actions/actionlint@v2
```

### 4.2 CI 잡 토폴로지

본 ADR 채택 후 CI 4 잡:
- `workflow-lint` (신규, ~30s)
- `backend-unit` (~15-20s)
- `frontend-unit` (~25-30s)
- `e2e` (~4-7m)

모두 ubuntu-latest 단일 runner. 4 잡 병렬 — wallclock 영향 0 (e2e 가 dominant).

## 5. 결과 (Consequences)

### 긍정적

- workflow YAML 회귀가 머지 직전에 차단됨.
- shellcheck 통합으로 run 블록 내 shell 회귀도 함께 검출.
- pinned action version 권장 메시지 노출.

### 부정적 / 트레이드오프

- `raven-actions/actionlint@v2` 자체가 외부 action 이라 maintainer 신뢰가 필요. 본 ADR 채택 시점에 raven-actions org 는 90+ repos / 500+ stars 의 OSS — 신뢰 acceptable.
- false positive 발생 시 lint 잡이 실패하므로 머지 차단. 본 ADR 채택 시점의 ci.yml 기준 false positive 미관측 (1차 도입 후 갱신).

### 비변경 사항

- 코드 lint (Go vet, ESLint, Vitest) 는 본 ADR 범위 밖 — backend-unit / frontend-unit 잡이 각각 책임.

## 6. 미해결 항목 (Open questions)

| 항목 | 후속 결정 |
| --- | --- |
| pre-commit hook 추가 (local 단계 차단) | 별도 hygiene 후보. local 미설치 사용자도 CI 가 차단하므로 priority 낮음. |
| actionlint config 파일 (`.github/actionlint.yaml`) — 회사 표준 self-hosted runner label 등 | 본 ADR 채택 시점에 사용 안 함. 필요 시 후속 갱신. |
| dependabot 으로 action 버전 자동 갱신 | 별도 후속 후보. 본 ADR 의 lint 와 보완재 관계. |

## 7. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-i`, D5). actionlint 잡 추가 + ADR-0003 §6 의 후속 결정. |
