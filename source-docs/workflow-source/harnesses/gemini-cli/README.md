# Gemini CLI Harness Package

- 문서 목적: 표준 AI 워크플로우를 Gemini CLI 하네스에 맞춰 배포할 때 생성할 파일과 검토 포인트를 정리한다.
- 범위: `GEMINI.md`, 공통 workflow 문서 연결 방식
- 대상 독자: Gemini CLI 사용자, 저장소 관리자, AI workflow 설계자
- 상태: draft
- 최종 수정일: 2026-04-25
- 관련 문서: `../../core/workflow_harness_distribution.md`, `../../scripts/bootstrap_workflow_kit.py`

## 생성 대상

- 프로젝트 루트 `GEMINI.md`

## 구성 원칙

- `GEMINI.md` 는 Gemini CLI 의 프로젝트 지침 진입점으로 사용한다.
- 상세 정책은 `ai-workflow/memory/` 문서를 먼저 읽도록 연결한다.
- Gemini CLI 에서는 `GEMINI.md` 가 시스템 프롬프트보다 우선하는 강력한 지침이므로, 핵심 운영 원칙을 여기에 명시한다.
- `invoke_agent` 를 사용해 메인 조정자와 worker 성격 서브 에이전트를 분리하는 패턴을 권장한다.
- 기존 프로젝트 도입 첫 세션에서는 `run_existing_project_onboarding.py` 결과의 `onboarding_summary`, `warnings`, `orchestration_plan` 을 먼저 읽고, 이를 Gemini CLI 내부 운영 힌트로 사용하는 구성이 적합하다.

## bootstrap 예시

```bash
python3 scripts/bootstrap_workflow_kit.py \
  --target-root /tmp/example-repo \
  --project-slug example_repo \
  --project-name "Example Repo" \
  --harness gemini-cli \
  --copy-core-docs
```

## 도입 후 확인 포인트

- `GEMINI.md` 가 생성됐는지 확인
- `ai-workflow/memory/` 경로가 현재 저장소 구조와 맞는지 확인
- 생성된 `GEMINI.md` 에 한국어 보고/컨텍스트 절약 원칙이 포함되는지 확인
- 생성된 `GEMINI.md` 에 메인 조정자와 worker 성격 서브 에이전트 분리 원칙이 포함되는지 확인

## 다음에 읽을 문서

- 적용 가이드: [./apply_guide.md](./apply_guide.md)
- 하네스 허브: [../README.md](../README.md)
- 배포 전략: [../../core/workflow_harness_distribution.md](../../core/workflow_harness_distribution.md)
