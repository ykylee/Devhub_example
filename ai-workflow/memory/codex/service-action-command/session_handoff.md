# Session Handoff

- 문서 목적: codex/service-action-command 브랜치 세션 간 작업 상태 인계 및 다음 단계 제안
- 범위: 현재 기준선, 브랜치별 memory 전환, 작업 상태, 다음 행동, 리스크
- 대상 독자: 후속 에이전트, 프로젝트 리드, 개발자
- 상태: active
- 최종 수정일: 2026-05-08
- 관련 문서: [Work Backlog](./work_backlog.md), [Project Profile](../../PROJECT_PROFILE.md), [Memory Governance](../../../MEMORY_GOVERNANCE.md)
- Branch: `codex/service-action-command`
- Updated: 2026-05-08

## 현재 기준선

- `origin/main` 최신 커밋을 현재 브랜치에 병합해 PR #33/#35/#37 반영 후 PR #36 변경을 재정렬했다.
- main에는 Phase 12 조직/사용자 CRUD API 및 Organization UI, Phase 13 Ory Hydra/Kratos IdP PoC 문서와 설정이 포함되어 있다.
- 현재 브랜치의 TASK-020~023 command/service action/realtime worker 변경은 main 반영 뒤 다시 적용되어 유지된다.
- 브랜치별 memory 표준을 적용해 이 브랜치의 source of truth를 `ai-workflow/memory/codex/service-action-command/`로 전환했다.
- 백엔드 개발 로드맵을 main 반영 사항 기준으로 재검토해 Phase 12/13, command/realtime 상태, P0~P3 우선순위를 반영했다.
- P0 첫 구현으로 `GET /api/v1/audit-logs`, organization CRUD audit, `X-Devhub-Actor` deprecation 응답 헤더를 추가했다.
- P0 후속으로 `docs/backend_api_contract.md` §11을 Hydra/Kratos 기준으로 재작성하고 Go Core Bearer token verifier 경계를 추가했다.
- RBAC 후속으로 `GET /api/v1/rbac/policy`와 프론트 `rbacService`를 추가해 Organization Permissions 화면이 backend policy를 조회할 수 있게 준비했다.
- RBAC persistence 후속으로 `rbac_policy_versions`/`rbac_policy_rules` migration과 `PUT /api/v1/rbac/policy`를 추가해 전체 matrix 교체 및 audit 기록 경계를 만들었다.
- RBAC enforcement 후속으로 `PUT /api/v1/rbac/policy`에 `system_config: admin` 권한 체크를 적용했다.
- `/api/v1/me` 후속으로 인증 actor를 DevHub `users`와 매핑하고 effective permissions를 반환하는 경계를 추가했다.
- service action, risk mitigation, audit 조회, 조직/사용자 쓰기 API에 RBAC enforcement를 확장했다.
- WebSocket `types` query 기반 subscription filtering과 event type별 RBAC read permission check 1차 구현을 추가했다.
- 백엔드 리뷰 결과를 반영해 인증 actor가 DevHub user에 매핑되지 않거나 비활성 상태면 `X-Devhub-Role`/`role` fallback으로 우회하지 못하도록 강화했다. DB/조직 store가 붙은 환경에서는 `DEVHUB_AUTH_DEV_FALLBACK=true`일 때만 role fallback을 허용한다.
- WebSocket publish 경로를 client snapshot 방식으로 바꿔 hub lock 밖에서 network write를 수행하고 실패 client를 제거하도록 개선했다.
- service action approval boundary 1차로 `POST /api/v1/commands/{command_id}/approve|reject`를 추가하고 audit log를 남기도록 구현했다.
- approved live service action을 원자적으로 `running` claim하는 Postgres query와 `ServiceActionExecutor` adapter interface/worker 경계를 추가했다. 실제 executor는 아직 main에 주입하지 않는다.
- `SERVICE_ACTION_EXECUTOR_MODE=simulation`에서만 켜지는 simulation executor를 추가했다. `SERVICE_ACTION_ALLOWED_SERVICES`와 `SERVICE_ACTION_ALLOWED_ACTIONS` allowlist를 모두 통과해야 하며 외부 side effect는 만들지 않는다.
- 2026-05-08 기준 PR 리뷰 반영 변경을 정리했고 백엔드 테스트, 프론트 lint, diff whitespace 검증은 통과했다. 문서 smoke check는 main에 병합된 workflow 문서 메타데이터/링크 정합성 문제로 실패하며 후속 작업으로 남긴다.

## 현재 주 작업 축

- command/audit 기반 액션 API, RBAC enforcement, WebSocket backend delivery 경계 안정화.
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
- TASK-BACKEND-P0-RBAC-PERSISTENCE: done
- TASK-BACKEND-P0-RBAC-ENFORCEMENT: done
- TASK-BACKEND-P0-ME: done
- TASK-BACKEND-P0-RBAC-WRITE-ENFORCEMENT: done
- TASK-BACKEND-P0-REALTIME-AUTH-FILTER: done
- TASK-BACKEND-P0-ACTOR-FALLBACK-HARDENING: done
- TASK-BACKEND-P0-REALTIME-PUBLISH-LOCK: done
- TASK-BACKEND-P1-SERVICE-ACTION-APPROVAL: done
- TASK-BACKEND-P1-SERVICE-ACTION-EXECUTOR-BOUNDARY: done
- TASK-BACKEND-P1-SERVICE-ACTION-SIM-EXECUTOR: done
- TASK-007 AI Gardener Suggestions 및 Admin Service Actions 실체화: planned

## 통합 메모

- 신규 세션은 먼저 현재 git 브랜치를 확인하고 이 디렉터리의 `state.json`, `session_handoff.md`, `work_backlog.md`, 최신 backlog를 읽는다.
- flat `ai-workflow/memory/state.json`, `session_handoff.md`, `work_backlog.md`, `backlog/`는 legacy fallback 및 공용 색인으로만 사용한다.
- 현재 방향성 충돌은 발견되지 않았다. Phase 13 전환을 위해 command/audit actor는 Bearer token verifier context를 우선 사용하고 `X-Devhub-Actor`는 deprecation fallback으로 유지한다.

## Next Actions

- [x] 브랜치별 memory 전환 문서 검증
- [x] API 계약 §11 Hydra/Kratos 재작성
- [x] Bearer token 검증 middleware 설계 및 최소 구현
- [x] RBAC policy 조회 API 및 프론트 Permissions 연동 준비
- [x] RBAC policy persistence/edit API와 audit 경계
- [x] RBAC policy edit enforcement
- [x] `/api/v1/me` 및 DevHub user-role lookup
- [x] service action/risk/audit/organization RBAC enforcement
- [x] WebSocket `types` subscription filtering 및 RBAC read permission check
- [x] 인증 actor 미매핑/비활성 시 role fallback 우회 차단
- [x] WebSocket publish lock 개선 및 실패 client 제거
- [x] service action approve/reject API 및 audit boundary
- [x] approved live service action query 및 executor adapter boundary
- [x] simulation service action executor 및 명시적 main 주입 설정
- [x] audit log 조회 API와 organization CRUD audit 연결
- [x] `X-Devhub-Actor` deprecation warning 경로 추가
- [ ] WebSocket replay 및 resource/project scope filter 설계
- [ ] service action 운영 executor adapter 구현 범위 확정
- [ ] Gitea Runner adapter/Gitea REST client 연동 범위 확정
- [ ] AI Gardener suggestion API/UI 연결 범위 확정
- [ ] 실제 executor/approval boundary 설계
- [ ] RBAC 편집 UI 활성화 전 confirmation UX 설계
- [ ] Hydra JWKS/introspection verifier 실제 구현
- [ ] `ai-workflow/tests/check_docs.py` 실패 항목 정리: main workflow 문서 메타데이터 누락 및 archive/branch memory broken link 보정

## Risks & Blockers

- flat memory를 계속 갱신하는 에이전트가 있으면 브랜치별 source of truth와 상태가 다시 갈라질 수 있다.
- 전체 시스템 실행에는 Docker/Docker Compose 또는 Phase 13 native binary 운영 경로 중 선택된 로컬 환경이 필요하다.
- 프론트엔드 WebSocket 연결은 환경에 따라 `NEXT_PUBLIC_WS_URL` 설정이 필요할 수 있다.
- service action command API는 실제 executor를 호출하지 않으며, 현재는 승인 불필요 dry-run command만 자동 성공 전이한다.
- WebSocket publish lock은 1차 개선됐지만, 고부하 backpressure가 필요하면 client별 bounded send queue 설계를 추가한다.
- 문서 smoke check는 2026-05-08 병합 기준 `ai-workflow/memory/*` 메타데이터 누락과 archive/branch memory broken link로 실패한다. 코드 병합을 막는 회귀는 아니지만 다음 문서 정리 작업에서 우선 처리한다.

## 다음에 읽을 문서

- [작업 백로그](./work_backlog.md)
- [2026-05-08 백로그](./backlog/2026-05-08.md)
- [2026-05-07 백로그](./backlog/2026-05-07.md)
- [메모리 거버넌스](../../../MEMORY_GOVERNANCE.md)
- [프로젝트 프로파일](../../PROJECT_PROFILE.md)
- [API 계약](../../../../docs/backend_api_contract.md)
- [백엔드 개발 로드맵](../../backend_development_roadmap.md)
