# Repository Assessment

- 문서 목적: 기존 프로젝트에 표준 AI 워크플로우를 도입하기 전에 현재 코드베이스와 문서 구조를 빠르게 진단한다.
- 범위: 저장소 구조, 추정 기술 스택, 문서 위치, 테스트 흔적, 초기 워크플로우 도입 포인트
- 대상 독자: 개발자, 운영자, AI agent, 프로젝트 온보딩 담당자
- 상태: done
- 최종 수정일: 2026-05-01
- 관련 문서: `./PROJECT_PROFILE.md`, `../core/workflow_adoption_entrypoints.md`, `./gemini/phase10/session_handoff.md`

## 1. 요약

- 분석 대상 프로젝트:
- `Standard AI Workflow`
- 분석 모드:
- `existing`
- 추정 기본 스택:
- `unknown`
- 감지된 스택 라벨:
- `없음`

## 2. 저장소 구조 관찰

- 상위 디렉터리 항목:
- `.DS_Store, .ai-workflow-backups, .codex, .git, .github, .gitignore, .venv, AGENTS.md, ANTIGRAVITY.md, QUICKSTART.md, README.md, dist, docs, requirements-dev.txt, scratch, split_checklist.md, workflow-source`
- 소스 디렉터리 후보:
- `없음`
- 문서 디렉터리 후보:
- `docs`
- 테스트 디렉터리 후보:
- `없음`

## 3. 추정 명령

- 설치:
- `python3 -m pip install -r requirements-dev.txt`
- 로컬 실행:
- `python3 workflow-source/scripts/bootstrap_workflow_kit.py --help`
- 빠른 테스트:
- `python3 workflow-source/tests/check_docs.py`
- 격리 테스트:
- `python3 workflow-source/tests/check_demo_workflow.py`
- 실행 확인:
- `python3 workflow-source/scripts/run_demo_workflow.py --example-project acme_delivery_platform`

## 4. package script 및 경로 샘플

- package script 목록:
- `없음`
- 분석 중 확인한 경로 샘플:
- `.DS_Store`
- `.gitignore`
- `AGENTS.md`
- `ANTIGRAVITY.md`
- `QUICKSTART.md`
- `README.md`
- `requirements-dev.txt`
- `split_checklist.md`
- `.ai-workflow-backups/20260424T125353Z/.opencode/.gitignore`
- `.ai-workflow-backups/20260424T125353Z/.opencode/package-lock.json`
- `.ai-workflow-backups/20260424T125353Z/.opencode/package.json`
- `.ai-workflow-backups/20260424T144730Z/.opencode/.gitignore`
- `.ai-workflow-backups/20260424T144730Z/.opencode/package-lock.json`
- `.ai-workflow-backups/20260424T144730Z/.opencode/package.json`
- `.codex/config.toml.example`
- `.github/workflows/smoke.yml`
- `docs/CODE_INDEX.md`
- `docs/DOCUMENT_INDEX.md`
- `docs/PROJECT_PROFILE.md`
- `docs/README.md`

## 5. 워크플로우 도입 초안

- 추천 문서 위키 홈:
- `docs/README.md`
- 추천 운영 문서 위치:
- `ai-workflow/memory/`
- 추천 backlog 위치:
- `ai-workflow/memory/backlog/`
- 추천 session handoff 위치:
- `ai-workflow/memory/session_handoff.md`

## 6. 자동 분석 기반 다음 작업

- 현재 추정 명령과 실제 운영 명령이 일치하는지 확인한다.
- 기존 문서 체계가 있으면 운영 문서 위치를 그대로 따를지, 별도 워크플로우 디렉터리로 분리할지 결정한다.
- 빠른 테스트와 실행 확인 기준이 약하면 우선 profile 문서에서 검증 규칙을 먼저 보강한다.

## 다음에 읽을 문서

- 프로젝트 프로파일: [./PROJECT_PROFILE.md](./PROJECT_PROFILE.md)
- 세션 인계 문서: [./gemini/phase10/session_handoff.md](./gemini/phase10/session_handoff.md)
- 도입 분기 가이드: [../core/workflow_adoption_entrypoints.md](../core/workflow_adoption_entrypoints.md)
