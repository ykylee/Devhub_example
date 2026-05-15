# Session Handoff — claude/work_260515-f (DREQ 도메인 컨셉~설계)

- 브랜치: `claude/work_260515-f`
- Base: `main` @ `feac299`
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: 개발 의뢰 (Dev Request, **DREQ**) 도메인의 컨셉 + 요구사항 + Usecase + 설계 + 로드맵 + 추적성 일괄 문서화. 구현/UT/TC 는 carve out.

## 도메인 정의 (한 줄)

외부 시스템 API 로 개발 의뢰를 접수하고, 담당자가 dashboard 에서 확인 후 application/repository/project 중 하나로 등록(promote)하는 운영 흐름.

## 입력 가정

- 의뢰 필드: 제목 / 상세내용 / 의뢰자 / 담당자
- 수신 시 "접수 대기 목록" 에 추가, 담당자 dashboard 표시
- 등록 시점에 application/repository/project 중 하나 선택

## 약어 결정

`DREQ` — REQ-FR-DREQ-* / UC-DREQ-* / TC-DREQ-* / IMPL-dreq-*. PROJ/APP/AUD 패턴 정합.

## 산출물 8개

1. `docs/planning/development_request_concept.md` (신규)
2. `docs/requirements.md` § 신규 (REQ-FR-DREQ-* + REQ-NFR-DREQ)
3. `docs/planning/system_usecases.md` (UC-DREQ-*)
4. `docs/architecture.md` § 신규 (ARCH-DREQ)
5. `docs/backend_api_contract.md` § 신규 (API-59..N)
6. `docs/development_roadmap.md` row
7. `docs/traceability/report.md` 도메인 row
8. `docs/planning/README.md` 인덱스 hook

## 작업 순서

1. (in progress) 메모리 초기화
2. (planned) 컨셉 문서 작성
3. (planned) 요구사항 §
4. (planned) Usecase
5. (planned) Architecture §
6. (planned) API contract §
7. (planned) 로드맵 + 추적성 매트릭스
8. (planned) planning README + PR open + merge

## 후속 sprint hook (본 sprint 머지 후)

- DREQ backend 구현 (store / handler / migration)
- DREQ frontend UI (담당자 dashboard 위젯 + 관리 페이지)
- 외부 수신 인증 ADR (HMAC / API key / OAuth client_credentials)
- DREQ → application/project promotion 정책 ADR (owner / leader 자동 매핑)
- UT-dreq + TC-DREQ 발급
