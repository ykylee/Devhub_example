# Code Index

- 문서 목적: 코드베이스 구조 및 핵심 컴포넌트를 안내하여 개발자의 코드 이해를 돕는다.
- 범위: 소스 코드 구조, 기술 스택, 핵심 모듈 설명
- 대상 독자: 개발자, AI 에이전트
- 상태: stable
- 최종 수정일: 2026-05-01
- 관련 문서: [./DOCUMENT_INDEX.md](./DOCUMENT_INDEX.md), [../README.md](../README.md)

이 문서는 `Standard AI Workflow` 저장소의 코드 구조와 핵심 컴포넌트를 안내합니다.

## 1. 프로젝트 구조 개요

```text
.
├── ai-workflow/            # 워크플로우 source + runtime/state 레이어 (단일 진입점)
│   ├── workflow_kit/       # 워크플로우 핵심 엔진 및 공통 모듈 (.gitignore 로 추적 제외)
│   ├── skills/             # 독립적으로 실행 가능한 워크플로우 기능(Skill) 모음 (추적 제외)
│   ├── mcp_servers/        # Model Context Protocol 기반 에이전트 도구 서버 (추적 제외)
│   ├── scripts/            # 프로젝트 부트스트랩 및 관리용 스크립트 (추적 제외)
│   ├── tests/              # 스모크 테스트 및 유효성 검사 도구 (추적 제외)
│   ├── schemas/            # JSON 스키마 및 출력 계약 (추적 제외)
│   ├── templates/          # 워크플로우 템플릿 (추적 제외)
│   ├── examples/           # 워크플로우 적용 예제와 출력 샘플 (추적 제외)
│   ├── core/               # 워크플로우 표준/스킬 spec 등 핵심 가이드
│   ├── prompts/            # 워커 프롬프트 템플릿
│   ├── releases/           # 워크플로우 릴리즈 노트
│   └── memory/             # 브랜치별 상태/백로그 관리
├── docs/                   # 공유 지식 문서
├── AGENTS.md               # Codex 진입 규칙
└── README.md               # 저장소 홈
```

> 과거 색인이 `workflow-source/` 로 표기하던 항목은 모두 `ai-workflow/` 하위로 통합되어 있다. 루트 `workflow-source` 호환 symlink 는 PR #12 정리 단계에서 제거되었으므로 신규 참조는 `ai-workflow/` 경로만 사용한다.

## 2. 핵심 컴포넌트

### Workflow Kit (`ai-workflow/workflow_kit/`)
- `common/`: 경로 분석(`paths.py`), 문서 파서(`project_docs.py`), 상태 생성(`workflow_state.py`) 등 공통 로직.
- `server/`: 워크플로우 도구를 API로 노출하기 위한 서버 프레임워크.

### Skills (`ai-workflow/skills/`)
각 스킬은 특정 워크플로우 단계를 자동화하는 독립 패키지입니다.
- `backlog-update/`: 백로그 생성 및 업데이트 자동화.
- `session-start/`: 세션 복원 및 상태 요약 제공.
- `merge-doc-reconcile/`: 상태 문서 병합 충돌 해결 도구.
- `robust-patcher/`: 정밀 코드 편집 엔진.

### MCP Tools (`ai-workflow/mcp_servers/`)
- `smart-context-reader/`: 에이전트가 코드 문맥을 시맨틱하게 읽을 수 있도록 돕는 도구 서버.

## 3. 주요 진입점 (Entry Points)

- **부트스트랩**: `ai-workflow/scripts/bootstrap_workflow_kit.py`
- **상태 생성**: `ai-workflow/scripts/generate_workflow_state.py`
- **테스트 실행**: `ai-workflow/tests/check_docs.py`, `ai-workflow/tests/check_bootstrap.py`
- **배포 패키지 생성**: `ai-workflow/scripts/export_harness_package.py`

## 4. 에이전트 활용 팁
- 코드 수정 시 `ai-workflow/workflow_kit/common/`의 유틸리티를 먼저 확인하여 중복 구현을 방지하십시오.
- 새로운 스킬 추가 시 `ai-workflow/skills/` 하위의 기존 패키지 구조를 준수하십시오.
