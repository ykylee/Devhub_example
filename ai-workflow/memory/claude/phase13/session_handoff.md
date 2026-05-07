# 세션 인계 문서 (Session Handoff)

- 문서 목적: Phase 13 (IdP 도입) 구현 세션 간 작업 상태 인계 및 다음 단계 제안
- 범위: Phase 13 P1 큐 진행 상태와 환경 제약, 차기 권장 사항
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: active
- 최종 수정일: 2026-05-07
- 관련 문서: [작업 백로그](./work_backlog.md), [ADR-0001](../../../../docs/adr/0001-idp-selection.md), [backend roadmap](../../backend_development_roadmap.md)

- 작성자: Claude Code
- 현재 브랜치: `claude/phase13`
- 분기 기준: `claude/init` @ `f77fa8a` (2026-05-07)
- 부모 브랜치 PR: [PR #10 — claude/init → main](https://github.com/ykylee/Devhub_example/pull/10) (미머지)

## 현재 세션 요약

`claude/init` 브랜치에서 ADR-0001(Ory Hydra+Kratos IdP 도입) 채택과 §8 미해결 항목 7종 결정을 완료한 뒤 PR #10 을 main 으로 올렸다. 본 브랜치 `claude/phase13` 은 PR #10 머지를 기다리지 않고 Phase 13 P1 1단계(PoC) 작업을 병행 진입하기 위해 분기됐다. PR #10 이 머지되면 본 브랜치를 main 에 rebase 한다.

## 진행 중

- **TASK-PHASE13-POC** (in_progress) — Phase 13 P1 1단계 PoC 환경 구성. sub-step (a) 사용자 수동 binary 설치 + (b)~(f) AI 자동화 분리.

## ADR-0001 §8 결정 7종 (참고)

이번 세션 시작 전 모두 결정 완료, ADR 인라인 반영됨.

1. DB 분리: `devhub` 단일 DB + `hydra`, `kratos` schema 분리 (`?search_path=...`).
2. 외부 client: 단계적 — 1차 사내 first-party only (`skip_consent=true`), 외부 SaaS 후속 phase.
3. MFA: 1차 미도입, schema 확장 가능 상태로 유지.
4. `X-Devhub-Actor` 폐기: P1 7단계에서 deprecation warning, 완전 제거 별도 phase.
5. Gitea SSO: 본 ADR 범위 밖, ADR-0002 예정.
6. Binary 획득: `go install` 사용자 터미널 수동 (사내 GoProxy 미러 재사용).
7. 호스트 서비스: PoC 직접 실행, 운영 진입 시점 별도 결정.

## Phase 13 P1 큐 (9단계, ADR-0001 결정 반영)

1. PoC 환경 구성 — **현재 진입**
   - (a) **사용자 수동, 샌드박스 외 터미널**: `go install github.com/ory/hydra/v2@latest`, `go install github.com/ory/kratos@latest`. 두 프로젝트 모두 `main.go` 가 모듈 루트에 있으므로 install path 에 `/cmd/...` 를 붙이면 실패한다 (2026-05-07 사용자 1차 시도 실패로 확인).
   - (b) DB schema 생성 SQL: `devhub` 안에 `hydra`, `kratos` schema.
   - (c) 설정 파일: `infra/idp/hydra.yaml`, `infra/idp/kratos.yaml` (위치 시작 시 확정).
   - (d) Kratos identity schema JSON: `traits.email`, `traits.display_name`, `metadata_public.user_id`.
   - (e) DevHub OIDC client 등록 스크립트 (PowerShell 또는 hydra CLI 호출).
   - (f) Next.js `/login` → Hydra `/oauth2/auth` → Kratos public flow → callback → token endpoint round-trip 검증.
2. DevHub OIDC client 등록 (PKCE + refresh token rotation + `skip_consent=true`).
3. `users.user_id` ↔ Kratos identity 1:1 매핑 검증.
4. identity ↔ users 동기화 어댑터 (Kratos webhook).
5. `/api/v1/admin/identities/*` Kratos admin API wrapper.
6. Kratos 이벤트 → DevHub audit log 6종 매핑.
7. Bearer token 검증 미들웨어 + `X-Devhub-Actor` deprecation warning.
8. `backend_api_contract.md §11` 재작성 (Hydra 표준 + admin wrapper + Kratos public flow 3개 절).
9. 테스트.

## 환경 가동 메모

- 로컬 PostgreSQL `:5432` (postgres/postgres, db=devhub) — 마이그레이션 `000001~000004` 적용됨.
- backend-core, frontend 가동 상태는 매 세션 시작 시 `Get-Process node, go` 로 확인.
- Hydra/Kratos 는 PoC 1단계 (a) 완료 후 직접 실행 (별도 PowerShell 창 또는 백그라운드).

## 다음에 읽을 문서

- [ADR-0001 IdP 선택](../../../../docs/adr/0001-idp-selection.md)
- [backend_development_roadmap.md P1](../../backend_development_roadmap.md)
- [2026-05-07.md](./backlog/2026-05-07.md)
