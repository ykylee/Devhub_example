# TASK-001 표준 AI 워크플로우 초기 도입 계획

- 문서 목적: 표준 AI 워크플로우 self-dogfood 온보딩 작업의 목표, 결정, 진행 상태, 다음 단계를 이어받기 쉽게 기록한다.
- 범위: workflow state docs 정렬, 배포 산출물 상태 검토, skill 보강 후보, 장기 작업 계획 문서 규칙 반영
- 대상 독자: 개발자, 운영자, AI agent, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../session_handoff.md`, `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/global_workflow_standard.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-001`
- 작업명:
- 표준 AI 워크플로우 초기 도입
- 현재 상태:
- `in_progress`
- 작업 카테고리:
- `workflow`, `documentation`, `development`
- 카테고리 설명:
- workflow self-dogfood, 배포 산출물 정리, 세션 복원/계획 문서 규칙 개발을 함께 다루는 장기 개선 작업
- 담당:
- Codex

## 2. 배경과 목표

- 배경:
- `ai-workflow/` 를 실제 배포 산출물로 보고, 이 저장소에서 self-dogfood 하며 workflow state docs 와 원본 workflow kit 의 경계를 정리한다.
- 목표:
- 세션 복원, backlog, handoff, 계획 문서, 검증 루프가 실제 작업에서 낭비를 줄이는지 확인하고 개선 사항을 원본 문서/템플릿에 반영한다.
- 기대 산출물:
- 갱신된 workflow state docs, 개선된 core 문서와 템플릿, 검증 결과, 다음 작업 후보

## 3. 범위

- 포함 범위:
- `ai-workflow/memory/*` 상태 문서 정렬
- `core/` 표준 문서 개선
- `templates/` 기본 양식 개선
- bootstrap 산출물 기본값 개선
- skill 미완성/보강 후보 정리
- 제외 범위:
- 전체 기능 구현 완료 선언
- MCP server 정식 승격
- 미관련 기존 worktree 변경 되돌리기
- 영향 파일/문서:
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/work_backlog.md`
- `ai-workflow/memory/backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`
- `ai-workflow/memory/state.json`
- `core/global_workflow_standard.md`
- `core/workflow_state_vs_project_docs.md`
- `templates/*`
- `scripts/bootstrap_workflow_kit.py`

## 4. 카테고리별 확장 기록

### 4.1 문서/워크플로우

- 갱신할 문서:
- `core/global_workflow_standard.md`
- `core/workflow_state_vs_project_docs.md`
- `templates/*`
- `ai-workflow/memory/*`
- 적용할 workflow 규칙:
- workflow state docs 와 project docs 를 분리한다.
- 장기 작업은 backlog 항목마다 계획 문서를 연결한다.
- 계획 문서는 카테고리 확장형으로 유지한다.
- 날짜별 backlog 는 호스트명/IP 폴더 아래로 분리한다.
- 배포 산출물 영향:
- `ai-workflow/` 를 최신 설치 결과물 테스트 대상으로 유지한다.
- 사용자 안내 필요 여부:
- 큰 작업을 이어받을 때는 backlog 항목보다 연결 계획 문서를 먼저 읽도록 안내한다.

### 4.2 개발 계획

- 사용자/운영 시나리오:
- 다음 세션이 `state.json` 과 `session-start` 결과에서 계획 문서를 발견하고 바로 이어서 작업한다.
- 기능 요구사항:
- backlog/handoff 링크 문서가 `next_documents` 로 노출된다.
- 비기능 요구사항:
- 계획 문서가 늘어나도 session-start 출력이 과도하게 장황해지지 않아야 한다.
- API/UI/데이터 계약:
- `state.json.next_documents` 에 계획 문서 경로를 포함한다.
- 구현 단위:
- `workflow_kit/common/project_docs.py`
- `workflow_kit/common/workflow_state.py`
- `skills/session-start/scripts/run_session_start.py`
- 호환성 고려:
- 기존 backlog 문서에 계획 문서가 없어도 session-start 는 경고 없이 동작해야 한다.

## 5. 현재까지 확인한 사실

- 코드베이스 분석 결과:
- skill 6종은 모두 smoke test 를 통과했다.
- `ai-workflow/` 삭제 항목 대부분은 루트 원본이 남아 있는 배포/설치 사본이다.
- `ai-workflow/README.md` 는 `Export Sample`/`new` 모드 산출물로 덮인 흔적이 있어, 배포 산출물 갱신 정책을 명확히 해야 한다.
- 관련 제약:
- 현재 worktree 에 기존 대량 수정/삭제가 있으므로 책임 파일 범위를 좁혀 작업한다.
- 전체 smoke 는 아직 실행하지 않았다.
- 의존성:
- Python 3.11 기준 검증
- `tests/check_*.py` smoke 기준선

## 6. 결정 기록

- 확정된 결정:
- `ai-workflow/` 는 배포 산출물로 보고 최신 설치 결과물 테스트 대상으로 유지한다.
- 큰 작업은 backlog 항목마다 계획 문서 또는 로드맵 문서를 연결한다.
- `state.json` 은 source-of-truth 문서 수정 후 재생성한다.
- 보류 중인 결정:
- `ai-workflow/` minimal 산출물을 그대로 둘지, core/support 사본을 다시 포함할지
- 전체 smoke 실행 시점
- skill 보강의 첫 구현 항목
- 되돌리면 안 되는 전제:
- 기존 미관련 worktree 변경은 사용자 변경으로 보고 임의로 되돌리지 않는다.

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | done | workflow state docs 를 현재 저장소 기준으로 정렬 | `check_docs`, `session-start`, `diff --check` |
| 2 | done | 실행 중 발견한 파서 친화 규칙과 검증 루프를 core 문서에 반영 | `core/workflow_state_vs_project_docs.md` |
| 3 | done | 삭제된 `ai-workflow/` 항목의 성격과 영향도 확인 | 루트 원본 존재 비교 |
| 4 | done | skill 6종 smoke 상태 확인 | `check_session_start`, `check_backlog_update`, `check_doc_sync`, `check_merge_doc_reconcile`, `check_validation_plan`, `check_code_index_update` |
| 5 | in_progress | 큰 작업별 계획 문서 연결 규칙 반영 | `core/global_workflow_standard.md`, `templates/work_item_plan_template.md` |
| 6 | done | 호스트명/IP 기준 backlog 경로 분리 | `latest-backlog`, `check_bootstrap`, `check_docs` |
| 7 | done | 개발 현황을 큰 작업으로 분류하고 계획 문서 등록 | `TASK-002` ~ `TASK-007` 계획 문서 |
| 8 | planned | 첫 skill 보강 항목 선택 | 후보: `session-start`, `validation-plan`, `code-index-update` |

## 8. 검증 계획과 결과

- 검증 계획:
- 문서 변경 후 `python3 tests/check_docs.py`
- bootstrap 변경 후 `python3 tests/check_bootstrap.py`
- session-start next document 후보 확인
- latest-backlog 의 중첩 backlog 탐색 확인
- 실행한 검증:
- `python3 tests/check_docs.py`
- `session-start` 경고 없음 확인
- skill 6종 smoke 확인
- `python3 mcp/latest-backlog/scripts/run_latest_backlog.py --backlog-dir-path ai-workflow/memory/backlog`
- 개발 현황 분류:
- `TASK-002` 기존 프로젝트 온보딩 자동 루틴 강화
- `TASK-003` 실제 적용 검증 및 파일럿 기록
- `TASK-004` 출력 계약 및 실패 스키마 안정화
- `TASK-005` 공통 라이브러리 추출 및 스크립트 정리
- `TASK-006` read-only MCP SDK transport 승격
- `TASK-007` 릴리즈 패키징 및 운영 절차 정리
- 실패한 검증:
- 없음
- 미실행 검증과 사유:
- 전체 smoke 는 변경 범위가 문서/bootstrap 중심이라 아직 실행하지 않았다.

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `ai-workflow/memory/state.json`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`
- 이 계획 문서
- 바로 할 일:
- `python3 tests/check_bootstrap.py` 와 `python3 tests/check_docs.py` 결과를 확인한다.
- 계획 문서 링크가 `state.json` 및 `session-start` next documents 에 노출되는지 확인한다.
- 계획 문서 템플릿의 카테고리 확장 구조를 bootstrap/export 산출물에도 반영할지 확인한다.
- backlog-update/doc-sync 출력 계약에 호스트별 backlog 경로와 계획 문서 필드를 더 깊게 반영할지 검토한다.
- 주의할 점:
- `ai-workflow/` 삭제 항목은 배포 산출물 정책 결정 전까지 손실로 단정하지 않는다.

## 10. 남은 리스크

- 계획 문서 경로가 너무 많아지면 session-start 출력이 장황해질 수 있다.
- backlog-update/doc-sync 출력 계약에는 아직 계획 문서 필드가 깊게 반영되지 않았다.
- bootstrap minimal 산출물과 copy-core-docs 산출물의 기대 구조가 더 명확해야 한다.
- 카테고리 섹션이 너무 많아지면 문서 작성 비용이 커질 수 있으므로 필요한 섹션만 남기는 원칙이 필요하다.
- 여러 호스트의 backlog 를 한 작업 흐름으로 합쳐 볼 요약 도구는 아직 없다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
- `2026-04-24`: 카테고리 확장형 구조 반영
- `2026-04-24`: 호스트명/IP 기준 backlog 경로 분리 반영
- `2026-04-24`: 개발 현황을 6개 장기 작업 계획으로 분류
