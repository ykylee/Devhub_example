# Session Handoff

- 문서 목적: 세션 간 작업 상태 인계 및 다음 단계 제안
- 범위: 현재 기준선, 작업 상태, main 동기화 결과, 다음 행동, 리스크
- 대상 독자: 후속 에이전트, 프로젝트 리드, 개발자
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [Work Backlog](./work_backlog.md), [Project Profile](./PROJECT_PROFILE.md), [Backend Roadmap](./backend_development_roadmap.md), [ADR-0001](../../docs/adr/0001-idp-selection.md)
- Branch: `codex/service-action-command`
- Updated: 2026-05-07

> 이 flat handoff는 legacy fallback 및 공용 색인이다. 현재 브랜치의 source of truth는 [codex/service-action-command/session_handoff.md](./codex/service-action-command/session_handoff.md)다.

## 현재 기준선

- `origin/main` 최신 커밋 `7316ddd`를 현재 브랜치에 fast-forward로 반영했다.
- 신규 세션 상태는 `ai-workflow/memory/codex/service-action-command/` 아래에서 브랜치별로 관리한다.
- main에는 Phase 12 조직/사용자 CRUD API 및 Organization UI, Phase 13 Ory Hydra/Kratos IdP PoC 문서와 설정이 포함되어 있다.
- 현재 브랜치의 TASK-020~023 command/service action/realtime worker 변경은 main 반영 뒤 다시 적용되어 유지된다.
- workflow kit `v0.5.0-beta` 기준 운영 문서 위치는 `ai-workflow/memory/`다.

## 현재 주 작업 축

- command/audit 기반 액션 API와 프론트 Admin Service Actions의 command status 표시 연결.
- Phase 12 조직 관리 API/UI와 command/realtime 작업을 같은 backend-core 라우터와 Postgres store 경계에서 공존시키기.
- Phase 13 IdP 도입 방향을 고려해 production command actor verification을 `X-Devhub-Actor` header에서 JWT/session 기반으로 전환 준비.

## Work Status

- TASK-003 Frontend REST API 및 WebSocket 실시간 연동: done
- TASK-019 command/audit 최소 schema 및 risk mitigation command API: done
- TASK-020 service action command API: done
- TASK-021 Workflow kit v0.5.0-beta 적용 및 existing project onboarding 갱신: done
- TASK-022 Command lifecycle 안전화 및 Admin service action 실연동: done
- TASK-023 Command status transition worker 및 WebSocket publish 경계 구현: done
- Phase 12 조직/사용자 CRUD API 및 Organization UI: main 반영 완료
- Phase 13 Ory Hydra/Kratos IdP PoC scaffold: main 반영 완료
- TASK-007 AI Gardener Suggestions 및 Admin Service Actions 실체화: planned

## 통합 메모

- `backend-core/internal/httpapi/router.go`는 `OrganizationStore`와 `RealtimeHub`를 함께 받도록 통합해야 한다.
- `backend-core/main.go`는 Postgres store를 organization, command, worker store로 동시에 사용하고 realtime hub publisher를 worker에 연결해야 한다.
- 현재 방향성 충돌은 발견되지 않았다. 다만 Phase 13에서 actor 인증이 JWT/session으로 이동하면 command API의 `X-Devhub-Actor` 사용은 deprecation 경로를 따라 교체해야 한다.

## Next Actions

- [ ] 충돌 해소 후 `cd backend-core && go test ./...` 실행
- [ ] 충돌 해소 후 `cd frontend && npm run lint` 실행
- [ ] 프론트 `RealtimeService` command status event UI 반영
- [ ] Gitea Runner adapter/Gitea REST client 연동 범위 확정
- [ ] AI Gardener suggestion API/UI 연결 범위 확정
- [ ] 실제 executor/approval boundary 설계
- [ ] production command API actor verification을 JWT/session 기반으로 전환

## Risks & Blockers

- 전체 시스템 실행에는 Docker/Docker Compose 및 backend-core API `:8080` 실행 환경이 필요하다.
- 프론트엔드 WebSocket 연결은 환경에 따라 `NEXT_PUBLIC_WS_URL` 설정이 필요할 수 있다.
- service action command API는 실제 executor를 호출하지 않으며, 현재는 승인 불필요 dry-run command만 자동 성공 전이한다.
- Phase 13 IdP PoC는 Hydra/Kratos binary와 PostgreSQL schema 준비가 필요하다.
- 새 workflow script는 `pydantic`이 필요하므로 bundled Python 또는 pydantic 설치 환경에서 실행해야 한다.

## 다음에 읽을 문서

- [작업 백로그](./work_backlog.md)
- [2026-05-06 백로그](./backlog/2026-05-06.md)
- [Claude Phase 13 handoff](./claude/phase13/session_handoff.md)
- [백엔드 개발 로드맵](./backend_development_roadmap.md)
- [API 계약](../../docs/backend_api_contract.md)
