# E2E 테스트 전략 및 지침 (E2E Testing Strategy & Guidelines)

- 문서 목적: DevHub 프로젝트의 품질 유지를 위한 지속 가능한 E2E 테스트 방법론과 자동화 지침을 정의한다.
- 범위: 테스트 분류, 실행 환경, CI/CD 연동, 데이터 관리 정책.
- 대상 독자: 모든 개발자, QA, AI 에이전트.
- 최종 수정일: 2026-05-12
- 상태: active

---

## 1. 테스트 원칙 (Core Principles)

1.  **현실성 (Realism)**: Mock IdP가 아닌 실제 Ory Hydra/Kratos 환경에서 테스트를 수행하여 OIDC 흐름의 완전성을 검증한다.
2.  **멱등성 (Idempotency)**: 테스트 전 `globalSetup`을 통해 필요한 시드 데이터를 자동 복구하며, 테스트 실행이 환경을 오염시키지 않도록 관리한다.
3.  **단일 워커 (Single Worker)**: Kratos 세션 충돌 방지를 위해 E2E 테스트는 원칙적으로 `workers: 1` 설정을 유지한다.
4.  **자동화 우선 (Automation First)**: 모든 신규 기능 개발 시 대응하는 E2E 테스트 케이스를 작성하고 CI 파이프라인에 통합한다.

---

## 2. 테스트 분류 및 실행 시나리오

### 2.1 Smoke Test (매 PR/커밋 시)
- **목적**: 핵심 경로(로그인, 대시보드 진입)의 치명적 결함 조기 발견.
- **대상**: `auth.spec.ts`, `signout.spec.ts`.
- **실행**: CI 파이프라인에서 자동 실행.

### 2.2 Functional Test (기능 개발 완료 시)
- **목적**: 개별 마일스톤(M2, M3 등)의 요구사항 충족 여부 검증.
- **대상**: `admin-users-crud.spec.ts`, `admin-org-crud.spec.ts` 등.
- **실행**: 개발 완료 후 로컬 및 staging 환경에서 수행.

### 2.3 Regression Test (배포 전)
- **목적**: 기존 기능의 영향도 파악 및 사이드 이펙트 방지.
- **대상**: 전체 E2E 테스트 스위트.
- **실행**: `main` 브랜치 머지 전 또는 배포 직전 단계에서 수행.

---

## 3. 지속적 테스트(Continuous Testing) 지침

### 3.1 로컬 개발 단계
- 개발자는 기능을 수정한 후 반드시 관련 E2E 테스트를 로컬에서 수행해야 한다.
- 명령어: `cd frontend && npm run e2e:ui` (특정 시나리오 선택 실행 권장)

### 3.2 CI/CD 파이프라인 (GitHub Actions)
- 모든 Pull Request는 `ci.yml` 워크플로우를 통과해야 한다.
- CI 환경에서는 `scripts/ci-setup.sh`를 통해 독립된 IdP 및 DB 환경을 구축한 후 테스트를 수행한다.
- 테스트 실패 시 Playwright Trace 및 Screenshot Artifact를 확인하여 원인을 파악한다.

### 3.3 시드 데이터 관리
- 테스트에 필요한 사용자는 `frontend/tests/e2e/fixtures.ts`의 `SEEDED` 객체에 정의된 표준 계정(`alice`, `bob`, `charlie`)을 사용한다.
- 테스트 중 데이터가 변경된 경우(비밀번호 변경 등), `finally` 블록에서 원복하거나 `globalSetup`의 자동 복구 기능에 의존한다.

## 4. 테스트 자산 관리 표준 (Management Standards)

### 4.1 테스트 케이스(TC) 문서 표준
- **파일명**: `docs/tests/test_cases_{milestone}_{feature}.md`
- **필수 섹션**: 메타데이터, 기능 맵, TC 상세(ID, 단계, 기대결과), 환경 전제.
- **ID 체계**: 
  - 기능: `F-{DOMAIN}-{SUB}` (예: `F-ORG-LIST`)
  - 케이스: `TC-{DOMAIN}-{ACTION}-{SEQ}` (예: `TC-ORG-UNIT-01`)

### 4.3 테스트 결과 보고서(Test Report) 표준
- **파일명**: `docs/tests/reports/report_YYYYMMDD_{milestone}.md`
- **필수 구성**:
  1. **요약**: 총 테스트 수, 성공/실패 수, 수행 시간.
  2. **환경 정보**: 실행 일시, OS/브라우저, 대상 URL.
  3. **상세 결과**: `TC ID | 결과 | 비고` 형식의 테이블.
  4. **결함 및 관찰 사항**: 실패한 TC의 원인 분석 및 스크린샷(있는 경우).
  5. **종합 의견**: 배포 가능 여부(Pass/Fail) 및 후속 조치.

---

## 5. AI 에이전트 작업 지침 (For Agents)

1.  **TC 우선주의**: 신규 기능 구현 전 반드시 위 표준에 맞는 TC 문서를 먼저 작성/수정한다.
2.  **Spec 정렬**: 작성된 TC ID를 Spec 파일의 `test` 타이틀이나 주석에 매핑하여 추적 가능성을 확보한다.
3.  **자동화 검증**: `npm run e2e`를 통해 구현 내용이 TC를 만족하는지 최종 확인한다.

---

## 5. 관련 링크
- [조직 관리 테스트 케이스](./test_cases_m3_org.md)
*   [인증/사용자 관리 테스트 케이스](./test_cases_m2_auth_user.md)
*   [E2E 실행 상세 가이드](../setup/e2e-test-guide.md)
*   [외부 시스템 연동 테스트 케이스 초안 (M4)](./test_cases_m4_integration.md)
