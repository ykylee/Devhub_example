# Workflow Update Migration Spec

- 문서 목적: workflow 패키지 업데이트 시 사용자 데이터를 유지하기 위한 마이그레이션 규칙을 정의한다.
- 범위: 마이그레이션 대상, 보존 대상, 변환規則, 백업 전략
- 대상 독자: 개발자, 배포 담당자, 프로젝트 온보딩 담당자
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../scripts/apply_harness_update.py`, `../core/workflow_skill_catalog.md`

## 1. 마이그레이션 원칙

### 1.1 데이터 보존 순위

| 순위 |数据类型| 마이그레이션 방식 |
|------|--------|-----------------|
| 1 | profile/handoff/state.json | 항상 보존, 새 패키지로 덮어쓰지 않음 |
| 2 | backlog entries | 항상 보존 |
| 3 | skills 설정 | 기존 값 유지 |
| 4 | 문서 템플릿 | 새 패키지로 업데이트 |

### 1.2 변환 필요 경우

- 스키마 버전 변화 시: 기존 데이터를 새 스키마로 변환
- 양식 변화 시: 마이그레이션 스크립트 실행
- 경로 변화 시: 새 경로로 이동

## 2. 마이그레이션 대상 파일

### 2.1 항상 보존 (절대 덮어쓰지 않음)

- `ai-workflow/memory/PROJECT_PROFILE.md`
- `ai-workflow/memory/session_handoff.md`
- `ai-workflow/memory/state.json`
- `ai-workflow/memory/backlog/*.md`

### 2.2 새 패키지로 업데이트 (덮어 writ 가능)

- `ai-workflow/README.md`
- `ai-workflow/core/*.md` (공통 표준)
- `AGENTS.md`
- `opencode.json` (설정이 없는 경우에만)
- `.opencode/skills/*.md`

### 2.3 선택적 마이그레이션

- MCPS설정: 기존 설정 유지 또는 병합

## 3. 마이그레이션 흐름

```
업데이트 요청
     ↓
백업 생성 (.ai-workflow-backups/)
     ↓
변경 파일 분류 (보존/업데이트/변환)
     ↓
변환 필요 시 마이그레이션 스크립트 실행
     ↓
패키지 파일 적용
     ↓
결과 검증
```

## 4. 구현 상태

- 현재 `apply_harness_update.py`는 기본 백업을 생성하지만, 데이터 보존 로직은 아직 구현되지 않음
- profile/handoff/state.json 보존을 위한 --preserve-data 플래그 필요
- 스키마 변환을 위한 migration helper 필요

## 5. 다음 작업

- apply_harness_update.py에 --preserve-data 옵션 추가
- profile/handoff/state.json 보존 로직 구현
- 스키마 버전 감지 및 자동 마이그레이션