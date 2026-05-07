# Project Workflow Profile

- 문서 목적: 공통 표준 워크플로우를 `Research Eval Hub` 저장소에 적용할 때 필요한 프로젝트 특화 규칙을 정리한다.
- 범위: 저장소 목적, 문서 구조, 기본 명령, 환경 기록 위치, 프로젝트 특화 검증 포인트, 예외 규칙
- 대상 독자: 연구 엔지니어, 개발자, 리뷰어, AI agent, 프로젝트 온보딩 담당자
- 상태: sample
- 최종 수정일: 2026-04-19
- 관련 문서: `../../core/global_workflow_standard.md`, `./session_handoff.md`, `./work_backlog.md`

## 1. 프로젝트 개요

- 프로젝트명:
- `Research Eval Hub`
- 프로젝트 목적:
- LLM 평가 데이터셋, 프롬프트 실험 결과, 배포용 리포트를 한 저장소에서 관리하고 재현 가능한 평가 흐름을 유지한다.
- 주요 이해관계자:
- applied research 팀, eval ops 팀, product quality 팀

## 2. 문서 구조

- 문서 위키 홈:
- `docs/README.md`
- 운영 문서 위치:
- `docs/evals/`
- 백로그 위치:
- `docs/evals/backlog/`
- 세션 인계 문서 위치:
- `docs/evals/session_handoff.md`
- 환경 기록 위치:
- `docs/evals/environments/`

## 3. 기본 명령

- 설치:
- `uv sync`
- 로컬 실행:
- `uv run python -m app.main`
- 빠른 테스트:
- `uv run pytest -q`, `uv run ruff check .`
- 격리 테스트:
- `uv run pytest tests/evals/test_pipeline.py -q`
- UI/API 실행 확인:
- `uv run python scripts/validate_dataset_manifest.py`, `uv run python scripts/build_eval_report.py --dry-run`

## 4. 프로젝트 특화 검증 포인트

- 코드 변경 시:
- 평가 파이프라인 로직 변경은 최소 한 개의 golden fixture 테스트와 lint 결과를 함께 남긴다.
- 문서 변경 시:
- dataset manifest, 실험 결과 요약, 배포용 report 링크가 서로 같은 버전을 가리키는지 확인한다.
- 프롬프트 변경 시:
- 프롬프트 템플릿 수정은 기준 실험 ID와 비교 요약을 backlog 또는 handoff에 남긴다.
- 배포/운영 변경 시:
- report 생성 dry-run, 산출물 경로 확인, 민감 데이터 제외 여부를 함께 검토한다.

## 5. 프로젝트 특화 예외 규칙

- 병합 규칙:
- `session_handoff.md` 와 최신 날짜 백로그가 충돌하면 최신 실험 결과 기준으로 handoff 요약을 다시 작성한다.
- 승인 규칙:
- 외부 공유용 평가 리포트는 applied research 팀과 product quality 팀 중 최소 한 곳의 검토를 거친다.
- 환경 제약:
- 내부 평가 데이터셋은 secure runner 에서만 전체 검증이 가능하며, 로컬에서는 축소 샘플만 사용할 수 있다.
- 기타:
- 실험 실패 로그는 삭제하지 않고 다음 세션에서 재현 여부와 함께 backlog 항목에 연결한다.

## 다음에 읽을 문서

- 공통 표준: [../../core/global_workflow_standard.md](../../core/global_workflow_standard.md)
- 세션 인계 문서: [./session_handoff.md](./session_handoff.md)
- 작업 백로그 인덱스: [./work_backlog.md](./work_backlog.md)
