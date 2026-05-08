# Phase 5 운영 지능화 가이드

- 문서 목적: Phase 5에서 도입된 지능형 도구(`rotate_workflow_logs`, `assess_milestone_progress`, `summarize_git_history`)의 실전 활용 방법을 설명한다.
- 범위: 문서 비대화 방지, 마일스톤 관리, 작업 이력 요약
- 대상 독자: AI 에이전트, 워크플로우 운영자, 개발자
- 상태: final
- 최종 수정일: 2026-04-27
- 관련 문서: `ai-workflow/memory/comprehensive_workflow_report.md`, `core/workflow_kit_roadmap.md`

## 1. 작업 이력 자동 요약 (`summarize_git_history`)

세션 인계(Handoff) 문서 작성 시, 실제 Git 커밋 이력을 바탕으로 요약을 생성한다.

**실행 명령:**
```bash
python3 workflow_kit/server/read_only_entrypoint.py --tool summarize_git_history \
  --payload-json '{"repo_path": ".", "commit_range": "HEAD~5..HEAD"}'
```

## 2. 문서 비대화 방지 및 로테이션 (`rotate_workflow_logs`)

`session_handoff.md`의 '최근 완료 작업'이 너무 많아지면(기본 10개), 오래된 작업을 기준선(Baseline)으로 이동시키고 목록을 정리한다.

**실행 명령:**
```bash
python3 workflow_kit/server/read_only_entrypoint.py --tool rotate_workflow_logs \
  --payload-json '{"handoff_path": "ai-workflow/memory/session_handoff.md", "max_done_items": 10}'
```

## 3. 마일스톤 진척도 관리 (`assess_milestone_progress`)

`maturity_matrix.json`에 정의된 현재 마일스톤과 `backlog`의 태스크 상태를 대조하여 진척도를 계산하고, 다음 단계로의 전환을 제안한다.

**실행 명령:**
```bash
python3 workflow_kit/server/read_only_entrypoint.py --tool assess_milestone_progress \
  --payload-json '{"matrix_path": "core/maturity_matrix.json", "backlog_path": "ai-workflow/memory/backlog/2026-04-27.md"}'
```

## 4. 권장 자동화 루틴

- **세션 시작 시**: `assess_milestone_progress`를 실행하여 현재 마일스톤 상태 확인.
- **세션 종료 시**:
  1. `summarize_git_history`로 인계 문서 초안 작성.
  2. `rotate_workflow_logs`를 실행하여 문서 크기 최적화.
