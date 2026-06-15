# Work Backlog — claude/merge_roadmap

- 문서 목적: `claude/merge_roadmap` 브랜치의 큰 단위 작업 인덱스. 일자별 backlog 파일에 대한 인덱스 역할.
- 범위: M0 sprint task/PR 상태, 다음 sprint 후보
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: M0 sprint 마무리 단계 (코드 측 종료, 운영 검증만 잔존)
- 최종 수정일: 2026-05-08
- 관련 문서: [통합 로드맵](../../../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json)

## 1. 활성 마일스톤

| 마일스톤 | 상태 | 진입 일자 | 종료 일자 | 종료 조건 |
| --- | --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 (인증·권한 정상화) | **done (코드)** | 2026-05-08 | 2026-05-08 | DoD #1~7,9 코드 측 resolved (PR #15·16·17·18·19); DoD #8 Phase 1 운영 검증만 사용자 환경 의존 (T-M0-10) |

## 2. 일자별 backlog 인덱스

| 일자 | 파일 | 핵심 |
| --- | --- | --- |
| 2026-05-08 | [`backlog/2026-05-08.md`](./backlog/2026-05-08.md) | M0 sprint 진입. T-M0-01~11 task 분해, 의존 그래프, PR 분할 전략 (PR-A~E) |

## 3. 주요 작업 항목 (task 단위)

| Task | 우선순위 | 상태 | 마일스톤 | PR |
| --- | --- | --- | --- | --- |
| T-M0-02 X-Devhub-Actor env gate (SEC-4 dev-only) | P0 | done | M0 | #15 |
| T-M0-01 empty Authorization 화이트리스트 (SEC-2) | P0 | done | M0 | #15 |
| T-M0-03 Role 가드 미들웨어 구현 (SEC-3) | P0 | done | M0 | #16 |
| T-M0-04 Role 가드 라우트 적용 (SEC-3) | P0 | done | M0 | #16 |
| T-M0-06 Hydra introspection 구현체 (SEC-2) | P0 | done | M0 | #17 |
| T-M0-05 config fail-fast 가드 (SEC-2) | P0 | done | M0 | #17 |
| T-M0-07 main.go verifier 주입 + URL Redacted (SEC-2) | P0 | done | M0 | #17 |
| T-M0-09 통합 테스트 매트릭스 | P0 | done | M0 | #15·16·17·18·19 |
| T-M0-08 Frontend AuthGuard / login 실 토큰 (SEC-1) | P0 | done | M0 | #18 |
| T-M0-11 SECURITY 마커 일괄 제거 + SEC-4 fallback path 제거 | P0 | done | M0 | #19 |
| T-M0-10 ADR-0001 Phase 1 운영 검증 | P0 | **pending** (사용자 환경) | M0 | — |

## 4. 다음 sprint 후보 (M1 진입 시)

통합 로드맵 §3.2 M1 의 DoD 와 PR-12 액션 트래커의 후속 항목들.

- M1: contract 정합성 — envelope/lifecycle/role wire 일관, audit actor 보강, RBAC policy edit 결정, types.ts UI/wire 분리
- SEC-5 (DB 에러 노출 마스킹) — `writeServerError` 헬퍼 도입, 별도 PR
- 통합 로드맵 §3.5 M4 의 SSO/MFA/OS service wrapper

## 5. 차단 / 결정 대기

- T-M0-10 운영 검증만 사용자 환경 (Hydra/Kratos PoC 가동) 의존. 검증 결과를 본 backlog 또는 PR 코멘트로 첨부.

## 6. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-08 | 본 인덱스 신설. M0 sprint 진입. |
| 2026-05-08 | M0 sprint 코드 측 종료. PR #15·16·17·18·19 머지. T-M0-10 만 운영 검증 잔존. |
