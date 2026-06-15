# Prototype-v1 Pre-release

- 문서 목적: `prototype-v1` pre-release 의 주요 변경점, 배포 산출물, 적용 대상, 다음 릴리즈 이관 항목을 기록한다.
- 범위: workflow/skill 온보딩 패키지 변경 요약, MCP 준비 상태, 배포 패키지 위치, 검증 결과
- 대상 독자: 저장소 관리자, 배포 담당자, 하네스 통합 담당자, 파일럿 적용자
- 상태: draft
- 최종 수정일: 2026-04-23
- 관련 문서: `../README.md`, `../core/workflow_kit_roadmap.md`, `../core/workflow_release_spec.md`, `../scripts/export_harness_package.py`

## 1. 릴리즈 성격

- 버전: `prototype-v1`
- 채널: `pre-release`
- 릴리즈 초점: workflow/skill 온보딩 묶음과 하네스별 minimal runtime 패키지 배포
- 제외 범위: official MCP SDK server 기본 활성화, 하네스 MCP 자동 연결, 실제 multi-project 파일럿 결과 일반화

## 2. 핵심 변경

### workflow/skill 축

- 기존 프로젝트 온보딩 진입점과 권장 skill 묶음을 문서 기준선으로 재정렬했다.
- 파일럿 후보 체크리스트와 파일럿 적용 기록 템플릿을 이번 릴리즈 방향에 맞춰 보강했다.
- bootstrap, onboarding runner, skill 카탈로그를 workflow/skill 중심 소비 경로로 다시 정렬했다.

### read-only MCP 준비 축

- read-only JSON-RPC bridge 의 malformed input/error envelope, initialize capability validation, notification lifecycle, stdio session gating 을 보강했다.
- official MCP Python SDK candidate 와 stdio round-trip smoke 를 추가해 차기 릴리즈 승격 준비선을 만들었다.
- Python 3.11 + `mcp[cli]` 개발 의존성과 CI 재현성 경로를 저장소에 반영했다.

### 배포 패키징 축

- 하네스 export 를 `agent_runtime_minimal` 프로필로 전환해 AI agent 런타임 파일만 기본 포함하도록 바꿨다.
- 하네스별 버전 디렉터리와 버전 포함 zip 이름을 도입했다.
- 각 패키지 루트에 `PACKAGE_CONTENTS.md`, `APPLY_GUIDE.md`, `manifest.json` 을 생성해 다른 환경에서 바로 적용할 수 있게 했다.

## 3. 배포 산출물

### Codex

- package root: `dist/harnesses/codex/prototype-v1/`
- zip: `dist/harnesses/codex/prototype-v1/standard-ai-workflow-codex-prototype-v1.zip`
- entrypoints:
- `bundle/AGENTS.md`
- `bundle/ai-workflow/memory/session_handoff.md`
- `bundle/ai-workflow/memory/work_backlog.md`
- `bundle/ai-workflow/memory/PROJECT_PROFILE.md`

### OpenCode

- package root: `dist/harnesses/opencode/prototype-v1/`
- zip: `dist/harnesses/opencode/prototype-v1/standard-ai-workflow-opencode-prototype-v1.zip`
- entrypoints:
- `bundle/AGENTS.md`
- `bundle/opencode.json`
- `bundle/.opencode/skills/standard-ai-workflow/SKILL.md`
- `bundle/.opencode/agents/workflow-orchestrator.md`

## 4. 검증

- `python3 tests/check_export_harness_package.py`
- `python3 tests/check_docs.py`
- `python3 tests/check_read_only_jsonrpc_bridge.py`
- `python3 tests/check_read_only_jsonrpc_fixtures.py`
- `python3 tests/check_read_only_transport_promotion_spec.py`
- `./.venv/bin/python tests/check_read_only_mcp_sdk_candidate.py`
- `./.venv/bin/python tests/check_read_only_mcp_sdk_stdio.py`
- `git diff --check`

## 5. 이번 릴리즈에서 의도적으로 미룬 항목

- official MCP SDK server 를 기본 하네스 경로로 승격하는 일
- MCP descriptor/example 을 하네스 runtime 패키지에 기본 포함하는 일
- 실제 파일럿 저장소 2건 이상 적용 후 범용성 확정
- 추가 하네스 확장과 다중 배포 채널 운영

## 6. 다음 릴리즈 제안

1. official MCP SDK server 를 실제 하네스 경로에 연결할지 결정한다.
2. 파일럿 저장소 1~2건에 minimal runtime package 를 적용하고 적용 기록을 남긴다.
3. `workflow_kit/common` 추출 범위를 더 넓혀 orchestration helper 와 packaging helper 를 정리한다.
4. release note 와 package manifest 사이에 changelog 자동화를 붙일지 검토한다.
