# 추적성 동기화 체크리스트

- 문서 목적: PR 작성자가 본 PR 의 추적성 영향을 매트릭스 + 단계 문서에 반영하는 절차를 정의한다.
- 범위: 매 PR (코드/문서/테스트 변경 모두 포함).
- 대상 독자: 모든 contributor (사람 + AI agent).
- 상태: accepted
- 최종 수정일: 2026-05-13
- 관련 문서: [`README.md`](./README.md), [`conventions.md`](./conventions.md), [`report.md`](./report.md), [`/.github/pull_request_template.md`](../../.github/pull_request_template.md).

## 1. 언제 적용되는가

본 체크리스트는 **모든 PR** 에 적용된다. 다만 변경 성격에 따라 적용 항목이 다르다:

| 변경 성격 | 필요 단계 |
| --- | --- |
| 새 요구사항 추가 | §3.1 (REQ ID 발급) + §3.6 (매트릭스 row 추가) + §3.7 (PR body 명시) |
| 기존 요구사항 구현 / 보강 | §3.4 (IMPL), §3.5 (UT/TC), §3.6 (매트릭스 cross-ref 갱신), §3.7 |
| 설계 변경 (architecture / API) | §3.2 (ARCH/API), §3.6, §3.7 + 영향받는 REQ 의 cross-ref 갱신 |
| 로드맵 갱신 | §3.3 (RM), §3.6, §3.7 |
| 단위/E2E 테스트 추가 | §3.5, §3.6, §3.7 + reverse trace 검증 |
| 문서 / 메모리 / 빌드 시스템 / CI 변경 | §3.7 만 (PR body 의 "추적성 영향: N/A" 명시) |
| 리팩토 (의미 변경 없음) | §3.7 만 ("추적성 영향: 없음, 의미 동등") |

## 2. 사전 조건

- `docs/traceability/report.md` 의 현재 상태 확인 (각 단계 마지막 ID 번호 + 매트릭스 영향 row 들).
- 본 PR 의 base branch 가 main 인지 확인 (다른 브랜치 위에서 fork 시 ID 충돌 위험).

## 3. 단계별 절차

### 3.1 새 요구사항 추가

1. `docs/requirements.md` 또는 `docs/backend/requirements.md` 에 요구사항 본문 추가.
2. 마지막 `REQ-FR-` / `REQ-NFR-` 번호 + 1 로 ID 할당, 본문의 해당 섹션/표 헤더 옆에 backtick 으로 ID 명기 (예: `### 2.5.7 SSO 통합 (REQ-FR-23)`).
3. `report.md` §`REQ-*` 인덱스 표에 1 row 추가.

### 3.2 설계 변경

1. `docs/architecture.md` / `docs/backend_api_contract.md` / 관련 spec 의 변경 섹션에 `ARCH-` 또는 `API-` ID 명기.
2. 새 ID 면 `report.md` 의 인덱스 표 + 매트릭스 row 갱신.
3. 영향받는 `REQ-*` 의 매트릭스 row 에서 cross-ref (ARCH/API 열) 갱신.
4. 큰 결정은 별도 ADR (`docs/adr/000X-*.md`) 으로 분리, `report.md` 의 "ADR 인덱스" 섹션 링크.

### 3.3 로드맵 갱신

1. `docs/development_roadmap.md` / `docs/frontend_development_roadmap.md` / `ai-workflow/memory/backend_development_roadmap.md` 의 마일스톤 표에 `RM-MX-YY` 명기.
2. 매트릭스의 영향 row 의 ROADMAP 열 갱신.

### 3.4 구현

1. 새 모듈 / 파일이면 `IMPL-<module>-XX` ID 부여.
2. 모듈 README 또는 첫 파일 상단 주석 ("// IMPL-rbac-03 — RBAC permission cache, see docs/traceability/report.md") 으로 명시.
3. 매트릭스의 IMPL 열 갱신 (file:line 또는 모듈 경로).

### 3.5 테스트 (UT + TC)

- **UT**: 새 `*_test.go` / `*.test.ts` 추가 시 `UT-<pkg>-XX` ID 를 테스트 파일 상단 주석 또는 `t.Run` 의 sub-test 이름에 명기.
- **TC** (E2E): 기존 패턴 (`TC-<feature>-XX`) 그대로. `docs/tests/test_cases_*.md` 의 표 + spec 파일 안의 `test()` 제목 양쪽에 ID 노출.
- 매트릭스의 UT/TC 열에 ID 추가 + 검증 대상 REQ 와 매핑 확인.

### 3.6 매트릭스 갱신

`docs/traceability/report.md` 의 종합 매트릭스 표:

- **새 row** 추가 시: REQ + ARCH/API + ROADMAP + IMPL + UT + TC 6 열 모두 채움 (`-` 로 미구현 표기 허용).
- **기존 row 갱신** 시: 변경된 열만 갱신, 다른 열은 그대로. commit SHA / PR 번호 추적은 git history 에 위임 — 매트릭스 본문에 절대 적지 않음.
- **deprecate** 시: row 의 status 열 (있다면) 또는 ID 옆에 ` (deprecated)` 마킹. row 자체는 제거하지 않는다 (`conventions.md` §2 영속성).

### 3.7 PR body / 커밋 메시지

`.github/pull_request_template.md` 의 "추적성 영향" 섹션 채움. 형식:

```
## 추적성 영향

- 추가: REQ-FR-23, IMPL-sso-01, UT-sso-01
- 갱신: TC-AUTH-04 (selector 변경, DoD 유지)
- Deprecate: (없음)
- 매트릭스: docs/traceability/report.md 에 row 1건 추가, row 3건 갱신
```

영향 없음일 때:

```
## 추적성 영향

- N/A — 본 PR 은 의미 변경 없는 리팩토 / 문서 정리.
```

## 4. 검증 (선택)

CI 게이트는 본 sprint 에 도입하지 않는다 (`README.md` §3 참조). 다만 PR 리뷰어는 다음을 manual 로 확인:

1. PR body 의 "추적성 영향" 섹션이 채워졌는가.
2. 추가/갱신된 ID 가 매트릭스에서 발견되는가.
3. 새 `REQ-FR-` 가 IMPL/UT/TC 매핑 없이 표류하지 않는가 (위반 시 "gap" 섹션에 추가).

## 5. 도구 부재 시 ad-hoc

본 체계는 정적 lint 가 없으므로 수작업 누락이 발생할 수 있다. **다음 PR 또는 sprint 마무리 시 한 번에 sync 하는 cleanup PR** 패턴 허용. 단 cleanup PR 의 body 에 "이전 PR #NN 의 추적성 누락 보강" 명시.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-13 | 1차 작성 (sprint `claude/work_260513-c`). |
