# 추적성 보완 계획 — 로그인 세션 + 사용자/조직

- 문서 목적: 로그인 세션 모듈과 사용자/조직 관리 모듈의 추적성 미흡 항목을 계획적으로 보완한다.
- 범위: REQ→UC→ARCH/API→IMPL→UT/TC 체인 보강, API 계약 본문 보강, E2E 카탈로그-실구현 정합성 회복.
- 대상 독자: Backend/Frontend 담당, QA, 추적성 리뷰어.
- 상태: active
- 최종 수정일: 2026-05-13
- 관련 문서: [추적성 매트릭스](./report.md), [동기화 체크리스트](./sync-checklist.md), [API 계약](../backend_api_contract.md), [M2 인증 TC](../tests/test_cases_m2_auth.md), [M3 조직 TC](../tests/test_cases_m3_organization.md)

## 1. 미흡 항목 요약

| ID | 모듈 | 미흡 내용 | 심각도 |
| --- | --- | --- | --- |
| GAP-AUTH-01 | 로그인 세션 | 인증 E2E가 실패/리다이렉트 중심이고 positive/refresh/expiry 체인 추적 약함 | high |
| GAP-AUTH-02 | 로그인 세션 | `API-32 /api/v1/me` 본문 계약 미완 | high |
| GAP-ORG-01 | 사용자/조직 | 조직 TC 카탈로그와 실제 e2e spec 구현 범위 불일치 | high |
| GAP-AUTH-03 | 로그인 세션 | `REAUTH_REQUIRED`, `revoke_active_sessions` 후속 과제가 추적성상 open | medium |

## 2. 보완 원칙

1. 각 GAP은 최소 1개 `UC-*`, 1개 `API/ARCH`, 1개 `TC-*`로 닫는다.
2. 카탈로그에 있는 TC는 반드시 대응 spec 파일 존재 여부를 검증한다.
3. 본문 계약 부재 API는 추적성 `closed` 처리 금지.
4. 완료 기준은 "문서 수정"이 아니라 "매트릭스 열 전체 연결"로 판단한다.

## 3. 실행 계획

### 3.1 Phase R1 — 로그인 세션 추적성 보강

- 목표: GAP-AUTH-01, GAP-AUTH-02 해소.
- 작업:
  1. `backend_api_contract.md`에 `API-32 /api/v1/me` 본문 spec 절 신설.
  2. 세션 positive 경로 E2E TC 추가:
     - `TC-AUTH-POS-01` 로그인 성공 후 보호 라우트 접근
     - `TC-AUTH-SESSION-01` 재로딩/탭 이동 시 세션 유지
     - `TC-AUTH-EXP-01` 만료/무효 세션 처리
  3. `system_usecases.md`의 `UC-AUTH-*`와 신규 TC 직접 매핑.
- 산출물:
  - `docs/backend_api_contract.md` 갱신
  - `docs/tests/test_cases_m2_auth.md` 갱신
  - `frontend/tests/e2e/auth.spec.ts` 또는 신규 spec
- 완료 조건:
  - `report.md` 인증 행 USECASE/API/TC 열이 신규 TC를 포함
  - GAP-AUTH-01/02 상태 `closed`

### 3.2 Phase R2 — 사용자/조직 카탈로그-구현 정합

- 목표: GAP-ORG-01 해소.
- 작업:
  1. `TC-ORG-UNIT-02/03`, `TC-ORG-MEM-02`, `TC-ORG-CHART-02` 대응 e2e 구현.
  2. 미구현 유지 시 카탈로그를 planned로 격하하고 사유 명시.
  3. spec 파일과 TC 카탈로그 자동 대조 스크립트(간단 grep) 도입 검토.
- 산출물:
  - `docs/tests/test_cases_m3_organization.md` 정합화
  - `frontend/tests/e2e/admin-org-crud.spec.ts` 또는 분리 spec 보강
- 완료 조건:
  - 카탈로그 TC 100%가 spec에서 검색 가능하거나 planned 명시
  - `report.md` 조직 행 TC 열과 실제 spec 일치

### 3.3 Phase R3 — 세션 보안 후속 과제 추적 닫기

- 목표: GAP-AUTH-03 해소.
- 작업:
  1. `revoke_active_sessions` 정책 적용 여부 결정 및 문서화.
  2. `REAUTH_REQUIRED` 시나리오를 요구사항/TC로 승격.
  3. 보안 설정 변경 시 ADR 또는 운영 가이드 링크 추가.
- 산출물:
  - `docs/tests/test_cases_m2_auth.md` 후속 항목 상태 갱신
  - 필요 시 ADR/설정 문서 업데이트
- 완료 조건:
  - 후속 항목이 `open`에서 `planned(with owner/date)` 또는 `closed`로 전환

## 4. 추적성 반영 항목 (필수)

- `docs/traceability/report.md`
  - GAP 표 상태 갱신
  - 인증/조직 도메인 행 TC 열 업데이트
- `docs/traceability/sync-checklist.md`
  - TC 카탈로그 ↔ spec 일치 검증 단계 추가(선택→권장)

## 5. 소유자 및 일정

| Phase | owner(제안) | 목표 완료일 |
| --- | --- | --- |
| R1 | Auth + Frontend | 2026-05-20 |
| R2 | Org + Frontend QA | 2026-05-22 |
| R3 | Auth + Security | 2026-05-27 |

## 6. 상태 추적

| 점검일 | GAP-AUTH-01 | GAP-AUTH-02 | GAP-ORG-01 | GAP-AUTH-03 | 메모 |
| --- | --- | --- | --- | --- | --- |
| 2026-05-13 | open | open | open | open | 초기 계획 수립 |

