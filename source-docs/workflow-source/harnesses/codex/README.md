# Codex Harness Package

- 문서 목적: 표준 AI 워크플로우를 Codex 하네스에 맞춰 배포할 때 생성되는 파일과 검토 포인트를 정리한다.
- 범위: `AGENTS.md`, Codex 설정 예시, 공통 문서 연결 방식
- 대상 독자: Codex 사용자, 저장소 관리자, AI workflow 설계자
- 상태: draft
- 최종 수정일: 2026-04-23
- 관련 문서: `../../core/workflow_harness_distribution.md`, `../../scripts/bootstrap_workflow_kit.py`

## 생성 대상

- 프로젝트 루트 `AGENTS.md`
- 프로젝트 루트 `.codex/config.toml.example`

## 구성 원칙

- `AGENTS.md` 는 Codex 의 프로젝트 지침 진입점으로 사용한다.
- 상세 정책은 `ai-workflow/memory/` 문서를 먼저 읽도록 연결한다.
- `.codex/config.toml.example` 는 전역 Codex 설정에 병합 가능한 샘플 스니펫으로 유지한다.
- Codex 에서는 OpenCode 처럼 project-local agent 권한을 세밀하게 분리하기 어렵기 때문에, `AGENTS.md` 에 메인/worker 운영 패턴을 명시하는 방식을 우선 사용한다.
- `main`/`small` 모델 구조를 쓴다면, Codex 에서도 메인 에이전트는 `main`, bounded scope 서브 에이전트는 `small` 로 두는 운영 원칙을 문서로 먼저 합의하는 편이 좋다.
- 기존 프로젝트 도입 첫 세션에서는 `run_existing_project_onboarding.py` 결과의 `onboarding_summary`, `warnings`, `orchestration_plan` 을 먼저 읽고, 이를 Codex 내부 운영 힌트로 사용하는 구성이 적합하다.
- Codex 에서 기존 프로젝트 첫 세션 결과를 읽을 때 권장 순서는 `status -> onboarding_summary.recommended_next_steps -> warnings -> orchestration_plan -> validation_plan -> code_index_update -> session_start -> repository_assessment.summary` 다.
- `repository_assessment.summary` 와 `onboarding_summary.inferred_commands` 가 함께 채워진 예시는 [../../examples/output_samples/existing_project_onboarding.with_assessment.sample.json](../../examples/output_samples/existing_project_onboarding.with_assessment.sample.json) 을 참고하면 된다.
- 실제 기존 저장소에 Codex 하네스를 얹은 첫 파일럿 기록은 [../../examples/pilot_adoption_open_git_client_example.md](../../examples/pilot_adoption_open_git_client_example.md) 을 참고하면 된다.
- `status == "error"` 인 경우에는 `error`, `error_code`, `source_context.failed_step`, 누락된 입력 경로를 먼저 요약하고 복구 작업부터 배치하는 편이 낫다.
- 하네스 export bundle 을 적용할 때 read-only MCP descriptor 는 `bundle/source-docs/schemas/read_only_transport_descriptors.json` 에 들어 있다.
- descriptor 를 Codex 설정 검토용 draft 로 변환한 예시는 `bundle/source-docs/schemas/read_only_harness_mcp_examples.json` 의 `harness_examples.codex` 에 들어 있다.
- Codex 전역 MCP 설정으로 바로 복사하기 전에 [../../mcp_servers/read_only_bundle.md](../../mcp_servers/read_only_bundle.md) 의 draft 범위와 `transport_ready=false` 상태를 먼저 확인한다.

## bootstrap 예시

```bash
python3 scripts/bootstrap_workflow_kit.py \
  --target-root /tmp/example-repo \
  --project-slug example_repo \
  --project-name "Example Repo" \
  --harness codex \
  --copy-core-docs
```

## 도입 후 확인 포인트

- `AGENTS.md` 에 적힌 기본 명령이 실제 저장소와 맞는지 확인
- `ai-workflow/memory/` 경로가 현재 저장소 구조와 맞는지 확인
- `.codex/config.toml.example` 를 실제 사용자 전역 설정에 반영할지 결정
- 생성된 `AGENTS.md` 에 한국어 보고/컨텍스트 절약 원칙이 포함되는지 확인
- 생성된 `AGENTS.md` 에 메인 조정자와 worker 성격 서브 에이전트 분리 원칙이 포함되는지 확인
- 기존 프로젝트 도입 시 onboarding runner 결과를 어디에 붙여 읽을지 팀 운영 방식이 정해져 있는지 확인
- export bundle 의 read-only MCP descriptor 를 실제 Codex MCP 설정에 연결할지, 참고 산출물로만 둘지 결정
- `read_only_harness_mcp_examples.json` 의 Codex 예시가 `manual_review_only` 임을 확인

## 다음에 읽을 문서

- 적용 가이드: [./apply_guide.md](./apply_guide.md)
- 하네스 허브: [../README.md](../README.md)
- 배포 전략: [../../core/workflow_harness_distribution.md](../../core/workflow_harness_distribution.md)
- bootstrap 생성물 샘플: [../../examples/bootstrap_output_samples.md](../../examples/bootstrap_output_samples.md)
