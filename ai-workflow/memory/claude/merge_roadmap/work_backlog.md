# Work Backlog — claude/merge_roadmap

- 문서 목적: `claude/merge_roadmap` 브랜치의 큰 단위 작업 인덱스. 일자별 backlog 파일에 대한 인덱스 역할.
- 상태: in_progress
- 최종 수정일: 2026-05-08
- 관련 문서: [통합 로드맵](../../../../docs/development_roadmap.md), [세션 인계](./session_handoff.md), [상태 스냅샷](./state.json)

## 1. 활성 마일스톤

| 마일스톤 | 상태 | 진입 일자 | 종료 조건 |
| --- | --- | --- | --- |
| **M0** — 보안 게이트 통과 (인증·권한 정상화) | in_progress | 2026-05-08 | 통합 로드맵 §3.1 의 DoD 9 항목 PASS + SECURITY 마커 0건 |

## 2. 일자별 backlog 인덱스

| 일자 | 파일 | 핵심 |
| --- | --- | --- |
| 2026-05-08 | [`backlog/2026-05-08.md`](./backlog/2026-05-08.md) | M0 sprint 진입. T-M0-01~11 task 분해, 의존 그래프, PR 분할 전략 (PR-A~E) |

## 3. 주요 작업 항목 (task 단위)

본 브랜치에서 다루는 작업 일람. 세부 DoD 와 의존은 §2 의 일자별 backlog 참조.

| Task | 우선순위 | 상태 | 마일스톤 | 일자별 backlog |
| --- | --- | --- | --- | --- |
| T-M0-02 X-Devhub-Actor env gate | P0 | planned | M0 | 2026-05-08 |
| T-M0-01 empty Authorization 화이트리스트 | P0 | planned | M0 | 2026-05-08 |
| T-M0-03 Role 가드 미들웨어 구현 | P0 | planned | M0 | 2026-05-08 |
| T-M0-04 Role 가드 라우트 적용 | P0 | planned | M0 | 2026-05-08 |
| T-M0-06 Hydra introspection 구현체 | P0 | planned | M0 | 2026-05-08 |
| T-M0-05 config fail-fast 가드 | P0 | planned | M0 | 2026-05-08 |
| T-M0-07 main.go verifier 주입 | P0 | planned | M0 | 2026-05-08 |
| T-M0-09 통합 테스트 매트릭스 | P0 | planned | M0 | 2026-05-08 |
| T-M0-08 Frontend AuthGuard / login 실 토큰 | P0 | planned | M0 | 2026-05-08 |
| T-M0-10 ADR-0001 Phase 1 운영 검증 | P0 | planned | M0 | 2026-05-08 |
| T-M0-11 SECURITY 마커 일괄 제거 | P0 | planned | M0 | 2026-05-08 |

## 4. 차단 / 결정 대기

- 없음 (sprint 진입 직후).

## 5. 변경 이력

| 일자 | 변경 |
| --- | --- |
| 2026-05-08 | 본 인덱스 신설. M0 sprint 진입. |
