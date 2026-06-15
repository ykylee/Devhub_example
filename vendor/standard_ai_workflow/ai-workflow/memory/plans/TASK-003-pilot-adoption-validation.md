# TASK-003 실제 적용 검증 및 파일럿 기록 계획

- 문서 목적: workflow kit 을 실제 저장소 또는 별도 환경에 적용해 운영 마찰과 보정 사항을 기록하는 장기 작업 계획을 관리한다.
- 범위: 파일럿 후보 선정, pre-release package 적용, 적용 기록, 피드백 반영
- 대상 독자: 저장소 관리자, 파일럿 적용자, AI workflow 설계자, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/workflow_kit_roadmap.md`, `../../../templates/pilot_candidate_checklist.md`, `../../../templates/pilot_adoption_record_template.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-003`
- 작업명:
- 실제 적용 검증 및 파일럿 기록
- 현재 상태:
- `planned`
- 작업 카테고리:
- `research`, `operations`, `workflow`, `release`
- 카테고리 설명:
- 샘플이 아닌 실제 소비 환경에서 workflow kit 적용 가능성과 마찰 지점을 검증한다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- 현재 저장소는 실행형 프로토타입 단계지만 여러 실제 프로젝트 적용 기록은 부족하다.
- 목표:
- 최소 1개 이상 실제 적용 기록을 남기고 규칙이 특정 샘플에 과적합됐는지 확인한다.
- 기대 산출물:
- 파일럿 적용 기록, 적용 전후 차이, 피드백 기반 개선 항목

## 3. 범위

- 포함 범위:
- 파일럿 후보 선정
- 하네스 package 적용
- 첫 세션 복원 마찰 기록
- 제외 범위:
- 모든 하네스 자동 연결 완성
- MCP server 승격 구현
- 영향 파일/문서:
- `templates/pilot_candidate_checklist.md`
- `templates/pilot_adoption_record_template.md`
- `examples/pilot_*`
- `releases/prototype-v2-pre-release.md`

## 4. 카테고리별 확장 기록

### 4.1 운영/릴리스

- 배포 단위:
- `dist/harnesses/*/prototype-v2/*.zip`
- 운영 절차:
- package 를 별도 저장소에 적용하고 `state.json -> handoff -> latest backlog` 복원 흐름을 기록한다.
- 모니터링/알림:
- 적용 중 깨진 링크, 과한 규칙, 누락된 명령, 하네스별 instruction 충돌을 기록한다.
- 롤백 기준:
- 대상 저장소 변경이 workflow state docs 범위를 넘어 의도치 않은 코드 변경을 만들면 중단한다.

### 4.2 코드베이스 분석

- 분석 질문:
- 신규/기존 프로젝트 도입 경로가 실제 저장소에서 충분히 짧고 명확한가
- 불확실한 지점:
- 후보 저장소의 보안/권한/CI 제약

## 5. 현재까지 확인한 사실

- 현재 상태:
- 파일럿 후보 체크리스트와 기록 템플릿은 존재한다.
- 로드맵상 4단계 진입 기준에 실제 적용 기록이 포함된다.
- 관련 제약:
- 적용 대상 저장소 선정이 필요하다.

## 6. 결정 기록

- 확정된 결정:
- 첫 파일럿은 pre-release package 소비 검증에 집중한다.
- 보류 중인 결정:
- 대상 저장소와 하네스 조합

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | 파일럿 후보 선정 | checklist |
| 2 | planned | package 적용 | 적용 기록 |
| 3 | planned | 첫 세션 복원 실행 | session-start 결과 |
| 4 | planned | 개선 항목 backlog 반영 | adoption record |

## 8. 검증 계획과 결과

- 검증 계획:
- `python3 tests/check_export_harness_package.py`
- 대상 저장소에서 `session-start` 또는 equivalent 복원 확인
- 실행한 검증:
- 아직 없음
- 미실행 검증과 사유:
- 파일럿 대상 미선정

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `templates/pilot_candidate_checklist.md`
- `templates/pilot_adoption_record_template.md`
- 바로 할 일:
- 후보 저장소 1개를 정하고 적용 범위를 제한한다.
- 주의할 점:
- 파일럿 적용 중 대상 저장소의 기존 변경을 되돌리지 않는다.

## 10. 남은 리스크

- 샘플에서는 통과하지만 실제 저장소에서는 문서 구조가 더 다양할 수 있다.
- 하네스별 전역 설정과 프로젝트 설정 충돌이 발생할 수 있다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
