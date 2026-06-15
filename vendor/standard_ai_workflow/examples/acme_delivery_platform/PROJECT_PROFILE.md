# Project Workflow Profile

- 문서 목적: 공통 표준 워크플로우를 `Acme Delivery Platform` 저장소에 적용할 때 필요한 프로젝트 특화 규칙을 정리한다.
- 범위: 저장소 목적, 문서 구조, 기본 명령, 환경 기록 위치, 프로젝트 특화 검증 포인트, 예외 규칙
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-18
- 관련 문서: `../../core/global_workflow_standard.md`, `./session_handoff.md`, `./work_backlog.md`

## 1. 프로젝트 개요

- 프로젝트명:
- `Acme Delivery Platform`
- 프로젝트 목적:
- 주문 접수, 배차, 배송 상태 추적, 운영자 승인 흐름을 통합 관리하는 서비스 플랫폼을 운영한다.
- 주요 이해관계자:
- 플랫폼팀, 배송 운영팀, QA팀, 고객지원팀

## 2. 문서 구조

- 문서 위키 홈:
- `docs/README.md`
- 운영 문서 위치:
- `docs/operations/`
- 백로그 위치:
- `docs/operations/backlog/`
- 세션 인계 문서 위치:
- `docs/operations/session_handoff.md`
- 환경 기록 위치:
- `docs/operations/environments/`

## 3. 기본 명령

- 설치:
- `make install-dev`
- 로컬 실행:
- `make run-api`
- 빠른 테스트:
- `pytest -q`, `make lint`
- 격리 테스트:
- `docker compose --profile test run --rm api-test`
- UI/API 실행 확인:
- `curl http://localhost:8080/health`, `npm run e2e:smoke`

## 4. 프로젝트 특화 검증 포인트

- 코드 변경 시:
- API 변경은 단위 테스트와 lint를 필수로 수행하고, 스키마 변경 시 마이그레이션 검토를 포함한다.
- 문서 변경 시:
- 운영 허브 링크, 메타데이터, 관련 runbook 링크 무결성을 함께 확인한다.
- UI 변경 시:
- 운영 콘솔 화면 변경은 스크린샷 또는 e2e smoke 결과를 남긴다.
- 배포/운영 변경 시:
- 헬스체크 확인, 롤백 절차 검토, 운영팀 공지 여부를 함께 점검한다.

## 5. 프로젝트 특화 예외 규칙

- 병합 규칙:
- `session_handoff.md` 와 최신 날짜 백로그가 충돌하면 병합 후 handoff를 우선 재작성한다.
- 승인 규칙:
- 운영 절차 문서는 플랫폼팀 또는 운영팀 리뷰 중 하나를 반드시 거친다.
- 환경 제약:
- staging 검증은 사내 VPN 연결 상태에서만 수행 가능하다.
- 기타:
- 장애 대응 중 생성된 임시 작업은 다음 영업일 첫 세션에서 정규 백로그 항목으로 재정리한다.

## 다음에 읽을 문서

- 공통 표준: [../../core/global_workflow_standard.md](../../core/global_workflow_standard.md)
- 세션 인계 문서: [./session_handoff.md](./session_handoff.md)
- 작업 백로그 인덱스: [./work_backlog.md](./work_backlog.md)
