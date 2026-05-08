# Planning Documentation

- 문서 목적: `docs/planning/` 디렉터리의 진입점. 마일스톤·로드맵·sprint plan·작업 트래커를 종류별로 안내한다.
- 범위: 통합 로드맵, 트랙별 세부 로드맵, 마일스톤별 backlog 위치, PR 별 트래커, 보안 리뷰 트래커, 향후 sprint plan
- 대상 독자: 프로젝트 리드, 백엔드/프론트엔드/Auth/AI/운영 트랙 담당자, 후속 작업자
- 상태: stable
- 최종 수정일: 2026-05-08 (TBD 스텁 → 인덱스 채움)
- 관련 문서: [../README.md](../README.md), [../development_roadmap.md](../development_roadmap.md), [../DOCUMENT_INDEX.md](../DOCUMENT_INDEX.md)

## 0. 진입점 — 무엇부터 읽는가

| 1차 (모든 트랙 공통) | 2차 (본인 트랙) | 3차 (sprint 진입 시) |
| --- | --- | --- |
| [통합 개발 로드맵](../development_roadmap.md) | 트랙별 세부 로드맵 (§2) | 마일스톤별 backlog (§3) |

본 README 는 그 진입점의 *지도* 다.

## 1. 로드맵 / 마일스톤 / 우선순위 체계

> [통합 개발 로드맵](../development_roadmap.md) 이 source-of-truth.

### 1.1 마일스톤

| 마일스톤 | 목표 | 우선순위 | 진입 조건 |
| --- | --- | --- | --- |
| **M0** | 보안 게이트 통과 (인증·권한 정상화) | P0 | PR #12 머지 직후 즉시 |
| **M1** | 핵심 기능 contract 정합성 | P0~P1 | M0 의 인증·권한 동작 확인 |
| **M2** | 사용자 경험 정합 (Phase 4·5 잔여 + Phase 6/6.1) | P1~P2 | M0 종료, M1 일부 병렬 가능 |
| **M3** | Realtime 확장 + 외부 연동 1차 (Gitea Hourly Pull, AI gRPC) | P2~P3 | M1·M2 종료 |
| **M4** | 운영 / SSO / MFA / 후속 ADR | P3 | M3 이후, capacity 확보 시 |

### 1.2 우선순위

| Priority | 의미 |
| --- | --- |
| P0 | 머지된 코드의 보안/통합 결함, 다음 모든 작업의 전제 |
| P1 | 핵심 기능의 contract·실행 경계 정합성 |
| P2 | 실시간 확장, RBAC UI 고도화, 운영 강화 |
| P3 | 외부 연동(Gitea REST, AI gRPC, SSO), 후속 phase |

## 2. 트랙별 세부 로드맵

| 트랙 | 세부 로드맵 | 책임 영역 |
| --- | --- | --- |
| **B — Backend** | [`../../ai-workflow/memory/backend_development_roadmap.md`](../../ai-workflow/memory/backend_development_roadmap.md) | Go Core API, store, normalize, command worker, realtime hub |
| **F — Frontend** | [`../frontend_development_roadmap.md`](../frontend_development_roadmap.md) | Next.js (대시보드, 조직, 인증 UI, 실시간 통합, RBAC UI) |
| **A — Auth & IdP** | [`../adr/0001-idp-selection.md`](../adr/0001-idp-selection.md) | Ory Hydra + Kratos, 토큰 검증, 권한 가드 |
| **X — Cross / Contract** | [`../backend_api_contract.md`](../backend_api_contract.md), [`../architecture.md`](../architecture.md) | API 계약, 메시지 envelope, role wire format, 데이터 모델 |
| **O — Operations** | [`../setup/environment-setup.md`](../setup/environment-setup.md) | 환경 셋업, 배포, 운영 모니터링 |
| **AI** | [`../backend/requirements_review.md`](../backend/requirements_review.md) §3-P3 | Python AI service (gRPC), Gardener, weekly report |

## 3. 마일스톤별 backlog 위치

backlog 는 *브랜치별 메모리* 에서 관리한다 (CLAUDE.md 정책). 작업이 진행되는 브랜치에 따라 위치가 결정된다.

| 환경 | 위치 패턴 |
| --- | --- |
| Claude Code 사용 | `ai-workflow/memory/claude/<branch>/backlog/<date>.md` |
| Codex CLI 사용 | `ai-workflow/memory/codex/<branch>/backlog/<date>.md` |
| Gemini CLI 사용 | `ai-workflow/memory/gemini/<branch>/backlog/<date>.md` |
| Antigravity 사용 | `ai-workflow/memory/antigravity/<branch>/backlog/<date>.md` |
| Agent prefix 없는 브랜치 | `ai-workflow/memory/branches/<branch>/backlog/<date>.md` |
| flat (legacy fallback) | `ai-workflow/memory/backlog/<date>.md` (현재는 `_archive/legacy-flat-backlog/` 로 이전) |

마일스톤 진입 시점에 backlog 항목을 분해한다 — 각 항목은 (위치, 우선순위, DoD, 의존, 책임자) 를 포함.

## 4. PR · 보안 리뷰 트래커 (영구 보존)

PR 단위의 의사결정과 보안 리뷰 결과는 *통합 로드맵 산출물* 로 영구 보존한다.

| 트래커 | 위치 | 메모 |
| --- | --- | --- |
| PR #12 액션 트래커 | [`../../ai-workflow/memory/PR-12-review-actions.md`](../../ai-workflow/memory/PR-12-review-actions.md) | BLK/SEC/HYG/INF/FU 분류와 처리 결과. M0 에 SEC-1~5 가 흡수됨. |
| PR #12 보안 리뷰 (PR diff 한정) | [`../../ai-workflow/memory/PR-12-security-review.md`](../../ai-workflow/memory/PR-12-security-review.md) | Auth 통합 결함 1건 High. M0 의 입력 자료. |
| 코드베이스 전체 보안 리뷰 (2026-05-08) | [`../../ai-workflow/memory/codebase-security-review-2026-05-08.md`](../../ai-workflow/memory/codebase-security-review-2026-05-08.md) | 신규 0건, 기존 1건 재확인 + DB 에러 노출(SEC-5) 권고. M0~M1 입력 자료. |

향후 PR 별 트래커도 같은 디렉터리에 `PR-N-review-actions.md`, 보안 리뷰는 `PR-N-security-review.md` 또는 `codebase-security-review-<date>.md` 로 일관되게 둔다.

## 5. 향후 추가될 자료

본 디렉터리에 추가될 가능성이 있는 자료 (현재 미작성):

- **Sprint plan**: 마일스톤별 sprint 단위 분해 (예: `sprint-2026-Q3-W3.md`).
- **Quarter / 분기 OKR**: 분기별 핵심 결과 정의.
- **Release plan**: 릴리즈 단위 묶음 (M0+M1 → v0.4.0, M2 → v0.5.0 등).
- **Capacity / 일정 시각화**: 마일스톤별 인력·기간 추정.
- **회의록**: 트랙 간 sync 결정 (대안: 각 트랙 backlog 에 분산).

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-04-30 | 초기 TBD 스텁 생성 |
| 2026-05-08 | TBD 스텁 → 디렉터리 진입점 + 마일스톤·트랙·트래커 인덱스 (claude/merge_roadmap 브랜치) |

## 7. 신규 자료 작성 규칙

1. 파일명은 영문 소문자 + `_` 또는 `-`. 일자가 들어가면 `YYYY-MM-DD` 형식.
2. 본문 시작에 표준 frontmatter (목적/범위/대상/상태/최종 수정일/관련 문서) 를 포함한다.
3. 새 자료가 마일스톤·트랙·트래커 중 어디에 속하는지 본 README 에 1줄 추가.
4. *영구 결정* 은 [`../adr/`](../adr/) 에 ADR 로, *마일스톤 산출물* 은 본 디렉터리에, *작업 트래커* 는 브랜치별 메모리에 둔다 (§3 표).
