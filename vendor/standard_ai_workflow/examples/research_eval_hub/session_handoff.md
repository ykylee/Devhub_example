# 세션 인계 문서

- 문서 목적: 새 세션이나 새 환경에서 이전 작업 상태를 빠르게 복원할 수 있도록 현재 기준 상태를 요약한다.
- 범위: 진행 중 작업, 차단 작업, 최근 완료 작업, 잔여 작업, 환경별 검증 현황
- 대상 독자: 연구 엔지니어, 개발자, 리뷰어, 문서 작성자
- 상태: sample
- 최종 수정일: 2026-04-19
- 관련 문서: `./PROJECT_PROFILE.md`, `./work_backlog.md`, `./backlog/2026-04-19.md`

## 1. 현재 작업 요약

- 현재 기준선:
- 평가 리포트 템플릿 v2 초안이 정리되었고, dataset manifest 와 실험 결과 문서 간 링크는 최신 기준으로 맞춰졌다.
- 현재 주 작업 축:
- 외부 공유용 평가 리포트 패키지와 내부 실험 로그 문서의 동기화 기준을 표준화하는 작업
- 최근 핵심 기준 문서:
- `docs/evals/session_handoff.md`, `docs/evals/reports/release-report-v2.md`

## 2. 진행 중 작업

- 현재 `in_progress` 작업:
- `TASK-044 평가 리포트 패키지와 실험 메타데이터 정합성 점검`

## 3. 차단 작업

- 현재 `blocked` 작업:
- `TASK-041 secure runner 전체 데이터셋 검증 대기`

## 4. 최근 완료 작업

- 최근 완료 작업 목록:
- `TASK-039 리포트 템플릿 front matter 정리`
- `TASK-043 dataset manifest 링크 표준화`

## 5. 잔여 작업 우선순위

### 우선순위 1

- release-report-v2 와 실험 결과 요약 문서의 버전 태그 일치 여부 재확인
- secure runner 사용 가능 시 전체 데이터셋 dry-run 검증

### 우선순위 2

- 프롬프트 실험 요약 섹션을 주간 연구 리뷰 문서로 이관
- report export 체크리스트를 자동화 후보로 정리

## 6. 환경별 검증 현황

- 검증 완료 호스트:
- `eval-mac-01 / 10.20.8.15`
- 주요 제약:
- 내부 데이터셋 전문 검증은 secure runner 접근 권한이 있어야만 수행 가능

## 7. 현재 세션 운영 메모

- 메인 오케스트레이터는 release report, manifest, handoff 결과를 통합해 최종 기준선을 정한다.
- `doc-worker` 에게는 report, manifest, 실험 요약 문서 간 링크와 버전 태그 비교를 맡긴다.
- `code-worker` 에게는 report builder 나 export 설정처럼 범위가 명확한 코드/설정 수정만 넘긴다.
- `validation-worker` 에게는 secure runner dry-run 가능 여부, 로그 확인, 증빙 메모 수집을 맡긴다.
- 기본 모델 분배는 `main orchestrator + small workers` 로 두고, 실험 해석이 복잡해지면 doc-worker 또는 validation-worker 만 일시적으로 `main` 으로 승격한다.

## 다음에 읽을 문서

- 작업 백로그 인덱스: [./work_backlog.md](./work_backlog.md)
- 프로젝트 프로파일: [./PROJECT_PROFILE.md](./PROJECT_PROFILE.md)
