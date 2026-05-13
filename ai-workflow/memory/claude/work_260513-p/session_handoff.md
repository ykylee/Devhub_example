# Session Handoff — claude/work_260513-p

- 문서 목적: 본 sprint 의 진입 상태 + 산출물 + 다음 sprint 진입점을 인계한다.
- 범위: 프로젝트 관리(Project Management) 신규 도메인 **컨셉 단계 1차**.
- 대상 독자: 후속 sprint 진입자, 리뷰어.
- 상태: in_progress (컨셉 문서 머지 직전)
- 최종 수정일: 2026-05-13
- 관련 문서: [컨셉 문서](../../../../docs/planning/project_management_concept.md), [통합 로드맵](../../../../docs/development_roadmap.md), [`work_backlog.md`](./work_backlog.md).
- 브랜치: `claude/work_260513-p` (base `main` @ `118899b`).

## 1. 본 sprint 작업 목표

기존 `docs/requirements.md §5.7` 에 *Gitea 자동화 시나리오* 수준으로만 존재하던 "프로젝트 및 저장소 관리" 항목을 **Project 단일 entity** 중심 도메인으로 격상하기 위한 **컨셉 1차** 정리.

구체적 1차 목표:

1. Project 의 도메인 정의 — 어떤 entity 들을 묶는 1차 컨테이너인가.
2. 두 핵심 usecase 의 분리 — *일반 사용자 조회* vs *시스템 관리자 등록·관리*.
3. MVP scope (CRUD + 등록 + 조회) 와 명시적 out-of-scope (실시간/AI/Gitea 자동화 등) 의 경계.
4. 데이터 모델 초안 + 향후 단계 (요구사항 / usecase / 설계) 의 진입 hook.

## 2. 산출물

- [`docs/planning/project_management_concept.md`](../../../../docs/planning/project_management_concept.md) — 컨셉 1차.
- [`docs/planning/README.md`](../../../../docs/planning/README.md) — 인덱스 1줄 추가.
- [`docs/development_roadmap.md`](../../../../docs/development_roadmap.md) §5 백로그 — "Project 도메인 — concept staged" 1행.

## 3. 다음 sprint 진입 안내

본 컨셉 머지 후 후속 sprint 에서 단계적으로 발전:

| 후속 sprint | 작업 | 산출물 |
| --- | --- | --- |
| **Req sprint** | 컨셉의 §3·§5 기반으로 REQ-FR-* 발급, `docs/requirements.md` §5.7 확장 또는 신규 절 신설 | requirements.md 갱신 + 매트릭스 row |
| **Usecase sprint** | 행위자 × usecase 매트릭스, 시퀀스, 권한 게이트, RBAC 매트릭스 확장 후보 | `docs/planning/project_management_usecases.md` (또는 requirements 본문에 흡수) |
| **Design sprint** | ARCH 추가, `backend_api_contract.md` 신규 § Project API, 데이터 스키마 마이그레이션 초안, frontend 진입점 | architecture.md / api_contract.md / migrations 디렉터리 |

본 sprint 에서는 **요구사항/설계 본문 수정 금지** — 컨셉 단계 격리.

## 4. 결정 / 보류

본 sprint 가 결정한 항목과 후속 sprint 로 보류한 항목은 컨셉 문서 §10 변경 이력에 명기.
