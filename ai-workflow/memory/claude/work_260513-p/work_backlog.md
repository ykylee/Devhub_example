# Work Backlog — claude/work_260513-p (프로젝트 관리 컨셉)

- 문서 목적: 본 sprint 의 작업 항목 진행 상태와 결정 기록을 추적한다.
- 범위: 프로젝트 관리 도메인 컨셉 1차 문서 작성 + 진입점 인덱스 갱신.
- 대상 독자: 본 sprint 진입자, 후속 sprint 리뷰어.
- 상태: in_progress
- 최종 수정일: 2026-05-13
- 관련 문서: [session_handoff.md](./session_handoff.md), [state.json](./state.json), [컨셉 문서](../../../../docs/planning/project_management_concept.md).
- 스프린트 목표: 프로젝트 관리 도메인 **컨셉 1차** 정리. 요구사항/설계 본문 수정은 후속 sprint 로 보류.

## 진행 상태

- [x] sprint state.json + session_handoff.md + work_backlog.md 초기화
- [x] docs/planning/project_management_concept.md 신규 작성
- [x] docs/planning/README.md §5.1 도메인 컨셉 인덱스 신설 (Project 1행)
- [x] docs/development_roadmap.md §5 백로그 표에 Project 도메인 1행 추가 + 최종 수정일/§7 변경 이력 갱신
- [x] 컨셉 doc cold read 후 4건 정합성 보강 (§3.1 archived / §3.2 영구 삭제 / §3.3 단계 / §5.1 code forward link) + §10 이력 1행 추가
- [x] commit + PR #102 생성 (https://github.com/ykylee/Devhub_example/pull/102)
- [x] CI 4잡 PASS (Backend Unit / Frontend Unit / Workflow Lint / E2E Playwright) → squash merge (commit `244f6b1`)
- [x] 후속 housekeeping sprint `claude/work_260513-q` 에서 main flat memory + slot close 마킹

## 결정 기록

- (TBD) 본 sprint 진입 시점 (2026-05-13) — 사용자 요청: CRUD 우선 + 일반 사용자 조회 / 시스템 관리자 등록·관리 분리.

## 다음 sprint 진입 후보 (컨셉 머지 후)

1. **Req sprint** — REQ-FR-* 발급 + requirements.md §5.7 확장.
2. **Usecase sprint** — 행위자 × usecase 매트릭스 + 시퀀스 + RBAC 게이트.
3. **Design sprint** — ARCH/API contract 추가 + 데이터 스키마 마이그레이션 초안.
