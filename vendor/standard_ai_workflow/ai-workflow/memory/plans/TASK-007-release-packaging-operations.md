# TASK-007 릴리즈 패키징 및 운영 절차 정리 계획

- 문서 목적: pre-release package, 릴리즈 노트, 하네스 zip 산출물, changelog 흐름을 반복 가능한 운영 절차로 정리하는 장기 작업 계획을 기록한다.
- 범위: harness export, release note, package manifest, changelog, GitHub pre-release asset
- 대상 독자: 배포 담당자, 저장소 관리자, 하네스 통합 담당자, 리뷰어
- 상태: draft
- 최종 수정일: 2026-04-24
- 관련 문서: `../work_backlog.md`, `../backlog/iyeong-gyun-ui-MacBookAir.local/192.168.0.139/2026-04-24.md`, `../../../core/workflow_release_spec.md`, `../../../releases/prototype-v2-pre-release.md`

## 1. 작업 개요

- 작업 ID:
- `TASK-007`
- 작업명:
- 릴리즈 패키징 및 운영 절차 정리
- 현재 상태:
- `planned`
- 작업 카테고리:
- `release`, `operations`, `documentation`, `workflow`
- 카테고리 설명:
- 하네스별 runtime package 와 릴리즈 문서를 반복 가능하게 만들고 배포 전후 검증 기준을 정리한다.
- 담당:
- 미정

## 2. 배경과 목표

- 배경:
- prototype-v2 pre-release 산출물과 export script 는 있지만 changelog/asset upload 흐름은 더 정리할 여지가 있다.
- 목표:
- 버전별 package, manifest, release note, 검증 결과가 같은 구조로 반복 생성되게 한다.
- 기대 산출물:
- 릴리즈 노트 템플릿, export 검증 체크, changelog 연결 기준

## 3. 범위

- 포함 범위:
- `scripts/export_harness_package.py`
- `releases/*.md`
- `dist/harnesses/*`
- package manifest/apply guide
- 제외 범위:
- 실제 GitHub release 생성 자동화 완성
- MCP server 기본 활성화
- 영향 파일/문서:
- `core/workflow_release_spec.md`
- `releases/prototype-v2-pre-release.md`
- `tests/check_export_harness_package.py`

## 4. 카테고리별 확장 기록

### 4.1 운영/릴리스

- 배포 단위:
- Codex/OpenCode harness zip
- 운영 절차:
- export -> smoke -> release note -> asset upload 순서
- 모니터링/알림:
- package contents, manifest, apply guide 누락 확인
- 롤백 기준:
- export package 가 state docs 또는 apply guide 를 누락하면 배포 중단

### 4.2 문서/워크플로우

- 갱신할 문서:
- `core/workflow_release_spec.md`
- `releases/prototype-vNext-pre-release.md`
- 사용자 안내 필요 여부:
- minimal runtime 과 source docs 포함 profile 의 사용 기준을 명확히 한다.

## 5. 현재까지 확인한 사실

- 현재 상태:
- prototype-v2 release note draft 와 export smoke 가 존재한다.
- 로드맵상 우선순위 5 작업이다.
- 관련 제약:
- GitHub release asset 업로드는 환경 권한과 네트워크 상태에 의존한다.

## 6. 결정 기록

- 확정된 결정:
- 기본 export 는 minimal runtime profile 을 우선한다.
- 보류 중인 결정:
- changelog 자동 생성 여부

## 7. 작업 단계

| 단계 | 상태 | 내용 | 검증/증빙 |
| --- | --- | --- | --- |
| 1 | planned | release spec 과 v2 note 차이 점검 | docs |
| 2 | planned | export manifest/check 강화 | export test |
| 3 | planned | 릴리즈 노트 템플릿화 | docs smoke |
| 4 | planned | asset upload 절차 문서화 | release checklist |

## 8. 검증 계획과 결과

- 검증 계획:
- `python3 tests/check_export_harness_package.py`
- `python3 tests/check_docs.py`
- 실행한 검증:
- 아직 없음
- 미실행 검증과 사유:
- 계획 등록 단계

## 9. 다음 세션 시작 포인트

- 먼저 읽을 파일:
- `core/workflow_release_spec.md`
- `releases/prototype-v2-pre-release.md`
- 바로 할 일:
- prototype-v2 release note 를 기준으로 반복 가능한 템플릿 항목을 추출한다.
- 주의할 점:
- dist 산출물 재생성은 기존 파일 변경 범위를 먼저 확인한다.

## 10. 남은 리스크

- 릴리즈 문서와 실제 export manifest 가 쉽게 어긋날 수 있다.
- source docs 포함 패키지와 minimal runtime 패키지의 기준이 혼동될 수 있다.

## 11. 변경 이력

- `2026-04-24`: 계획 문서 생성
