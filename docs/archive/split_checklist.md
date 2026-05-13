# Workflow Standard Kit 분리 체크리스트

> ⚠ **Deprecated** (2026-05-13, sprint `claude/work_260513-i`). 본 문서는 archive 위치로 이동된 초안. `workflow_standard_kit/` 분리는 본 저장소 외부의 별도 프로젝트 결정으로 이관됐다. 본 문서는 참조용으로만 유지된다.

- 문서 목적: `workflow_standard_kit/` 를 별도 프로젝트로 옮길 때 필요한 정리 항목과 완료 기준을 체크리스트로 제공한다.
- 범위: 디렉터리 구성, 링크 점검, 문서 self-contained 여부, 구현/미구현 경계, 배포 전 검증 항목
- 대상 독자: 개발자, 운영자, 프로젝트 분리 담당자
- 상태: deprecated
- 최종 수정일: 2026-05-13 (deprecated 마킹)
- 관련 문서: `README.md`, `core/global_workflow_standard.md`, `core/workflow_skill_catalog.md`, `core/workflow_mcp_candidate_catalog.md`, `core/workflow_agent_topology.md`

## 1. 분리 전 체크

- 외부 저장소 경로를 가리키는 링크가 없는가
- `core/` 와 `templates/` 만으로 최소 도입 흐름을 설명할 수 있는가
- 문서만으로 가능한 수동 운영 절차가 남아 있는가
- 미구현 항목이 구현 완료처럼 보이지 않는가

## 2. 별도 프로젝트에 포함할 파일

- `README.md`
- `split_checklist.md`
- `core/`
- `templates/`

## 3. 별도 프로젝트에서 추가 권장할 디렉터리

- `skills/`
- `mcp/`
- `examples/`
- `tests/`

## 4. 분리 후 바로 확인할 항목

- 문서 메타데이터 누락이 없는가
- 내부 링크가 모두 유효한가
- 템플릿만으로 새 프로젝트 문서를 생성할 수 있는가
- skill/MCP가 미구현이면 그 사실이 README에 드러나는가

## 5. 완료 기준

- 이 체크리스트를 따라가면 원 저장소 문서를 보지 않고도 구조를 이해할 수 있다.
- 키트 내부 링크와 메타데이터 검사가 모두 통과한다.
- 별도 프로젝트로 복사한 뒤 최소 세트만으로 첫 프로젝트 프로파일과 세션 문서를 만들 수 있다.
