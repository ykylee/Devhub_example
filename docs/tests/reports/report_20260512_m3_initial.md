# Test Report - 2026-05-12 (Milestone M3 Initial)

## 1. 요약 (Summary)
- **총 테스트 수**: 40
- **성공(Pass)**: 29
- **실패(Fail)**: 11
- **성공률**: 72.5%
- **수행 시간**: 약 5.6분

## 2. 환경 정보 (Environment)
- **실행 일시**: 2026-05-12 23:20 (KST)
- **OS/브라우저**: macOS / Chromium (Playwright)
- **대상 URL**: `http://localhost:3000`
- **시드 데이터**: `globalSetup` (alice/bob/charlie) 기반

## 3. 상세 결과 (Detailed Results)

### M3 조직 관리 (Organization Management)
| TC ID | 결과 | 비고 |
| --- | --- | --- |
| TC-ORG-LIST-01 | **PASS** | 리스트 뷰 노출 확인 |
| TC-ORG-LIST-02 | **FAIL** | 검색 입력 필드 앵커 텍스트 불일치 (Timeout) |
| TC-ORG-UNIT-01 | **FAIL** | 부서 생성 모달 진입 실패 (Timeout) |
| TC-ORG-MEM-01 | **FAIL** | 멤버 관리 모달 진입 실패 (Timeout) |
| TC-ORG-CHART-01 | **FAIL** | 뷰 전환 확인 실패 (Timeout) |

### M2 인증 및 사용자 관리 (Auth & User Management)
| TC ID | 결과 | 비고 |
| --- | --- | --- |
| TC-USR-01~06 | **PASS** | 검색 필터 정상 동작 |
| TC-USR-CRUD-01 | **FAIL** | 모달 열림 확인 실패 (UI 지연) |
| TC-USR-CRUD-03 | **FAIL** | 액션 메뉴 닫기 확인 실패 (Timeout) |
| TC-ACC-01~03 | **PASS** | 비밀번호 변경 및 계정 정보 확인 정상 |
| TC-NAV-01~03 | **PASS** | 헤더 네비게이션 및 Switch View 정상 |
| TC-SIGNUP-01~04 | **FAIL** | 회원가입 페이지 로딩 지연 (Timeout) |

## 4. 결함 및 관찰 사항 (Failures & Observations)
- **조직 관리(ORG)**: 구현된 UI의 텍스트와 테스트 코드의 Selector(`getByText`)가 일치하지 않아 대부분의 테스트가 실패함. (앵커 텍스트 수정 필요)
- **회원가입(SIGNUP)**: `/auth/signup` 페이지 로딩 시 간헐적으로 Timeout이 발생함. 이는 Kratos 응답 속도 또는 렌더링 지연으로 보임.
- **UI 타이밍**: 모달이 열리거나 메뉴가 닫히는 동작에서 `expect`가 너무 일찍 실행되어 실패하는 케이스가 확인됨. `waitFor` 로직 보강 필요.

## 5. 종합 의견 (Conclusion)
- **상태**: **FAIL (재시험 필요)**
- **의견**: 핵심 인증 및 사용자 검색 기능은 안정적이나, 이번에 신규 추가된 **조직 관리 기능**의 E2E 테스트가 Selector 불일치로 인해 정상 수행되지 않았습니다. 테스트 코드의 Selector를 실제 UI에 맞춰 보정하고, UI 렌더링 대기 시간을 최적화한 후 재시험이 필요합니다.
