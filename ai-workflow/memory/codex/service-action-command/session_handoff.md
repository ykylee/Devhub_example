# Session Handoff

- 문서 목적: codex/service-action-command 브랜치 세션 간 작업 상태 인계 및 다음 단계 제안
- 범위: 현재 기준선, 브랜치별 memory 전환, 작업 상태, 다음 행동, 리스크
- 대상 독자: 후속 에이전트, 프로젝트 리드, 개발자
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [Work Backlog](./work_backlog.md), [Project Profile](../../PROJECT_PROFILE.md), [Memory Governance](../../../MEMORY_GOVERNANCE.md)
- Branch: `codex/service-action-command`
- Updated: 2026-05-07

## 현재 기준선

- `origin/main` 최신 커밋 `7316ddd`를 현재 브랜치에 fast-forward로 반영했다.
- main에는 Phase 12 조직/사용자 CRUD API 및 Organization UI, Phase 13 Ory Hydra/Kratos IdP PoC 문서와 설정이 포함되어 있다.
- 현재 브랜치의 TASK-020~023 command/service action/realtime worker 변경은 main 반영 뒤 다시 적용되어 유지된다.
- 브랜치별 memory 표준을 적용해 이 브랜치의 source of truth를 `ai-workflow/memory/codex/service-action-command/`로 전환했다.
- 백엔드 개발 로드맵을 main 반영 사항 기준으로 재검토해 Phase 12/13, command/realtime 상태, P0~P3 우선순위를 반영했다.
- P0 첫 구현으로 `GET /api/v1/audit-logs`, organization CRUD audit, `X-Devhub-Actor` deprecation 응답 헤더를 추가했다.
- P0 후속으로 `docs/backend_api_contract.md` §11을 Hydra/Kratos 기준으로 재작성하고 Go Core Bearer token verifier 경계를 추가했다.
- RBAC 후속으로 `GET /api/v1/rbac/policy`와 프론트 `rbacService`를 추가해 Organization Permissions 화면이 backend policy를 조회할 수 있게 준비했다.

## 현재 주 작업 축

- command/audit 기반 액션 API와 프론트 Admin Service Actions의 command status 표시 연결.
- Phase 12 조직 관리 API/UI와 command/realtime 작업을 같은 backend-core 라우터와 Postgres store 경계에서 공존시키기.
- 브랜치별 memory 구조를 표준화해 flat memory 충돌을 줄이기.

## Work Status

- TASK-020 service action command API: done
- TASK-021 Workflow kit v0.5.0-beta 적용 및 existing project onboarding 갱신: done
- TASK-022 Command lifecycle 안전화 및 Admin service action 실연동: done
- TASK-023 Command status transition worker 및 WebSocket publish 경계 구현: done
- TASK-CODEX-MEMORY-BRANCH-SPLIT: done
- TASK-CODEX-BACKEND-ROADMAP-REVIEW: done
- TASK-BACKEND-P0-AUDIT-ACTOR: done
- TASK-BACKEND-P0-AUTH-BOUNDARY: done
- TASK-BACKEND-P0-RBAC-POLICY: done
- TASK-007 AI Gardener Suggestions 및 Admin Service Actions 실체화: planned

## 통합 메모

- 신규 세션은 먼저 현재 git 브랜치를 확인하고 이 디렉터리의 `state.json`, `session_handoff.md`, `work_backlog.md`, 최신 backlog를 읽는다.
- flat `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/`는 legacy fallback 및 공용 색인으로만 사용한다.
- 현재 방향성 충돌은 발견되지 않았다. Phase 13 전환을 위해 command/audit actor는 Bearer token verifier context를 우선 사용하고 `X-Devhub-Actor`는 deprecation fallback으로 유지한다.

## Next Actions

- [x] 브랜치별 memory 전환 문서 검증
- [ ] 프론트 `RealtimeService` command status event UI 반영
- [x] API 계약 §11 Hydra/Kratos 재작성
- [x] Bearer token 검증 middleware 설계 및 최소 구현
- [x] RBAC policy 조회 API 및 프론트 Permissions 연동 준비
- [x] audit log 조회 API와 organization CRUD audit 연결
- [x] `X-Devhub-Actor` deprecation warning 경로 추가
- [ ] Gitea Runner adapter/Gitea REST client 연동 범위 확정
- [ ] AI Gardener suggestion API/UI 연결 범위 확정
- [ ] 실제 executor/approval boundary 설계
- [ ] RBAC policy persistence/edit API와 audit/approval 경계 설계
- [ ] Hydra JWKS/introspection verifier와 role/permission lookup 구현

## Risks & Blockers

- flat memory를 계속 갱신하는 에이전트가 있으면 브랜치별 source of truth와 상태가 다시 갈라질 수 있다.
- 전체 시스템 실행에는 Docker/Docker Compose 또는 Phase 13 native binary 운영 경로 중 선택된 로컬 환경이 필요하다.
- 프론트엔드 WebSocket 연결은 환경에 따라 `NEXT_PUBLIC_WS_URL` 설정이 필요할 수 있다.
- service action command API는 실제 executor를 호출하지 않으며, 현재는 승인 불필요 dry-run command만 자동 성공 전이한다.

## 다음에 읽을 문서

- [작업 백로그](./work_backlog.md)
- [2026-05-07 백로그](./backlog/2026-05-07.md)
- [메모리 거버넌스](../../../MEMORY_GOVERNANCE.md)
- [프로젝트 프로파일](../../PROJECT_PROFILE.md)
- [API 계약](../../../../docs/backend_api_contract.md)
- [백엔드 개발 로드맵](../../backend_development_roadmap.md)
