# system_usecases.md ↔ 현재 코드 정합 분석

- 문서 목적: `docs/planning/system_usecases.md` 의 UC 인벤토리를 main `cf19c94`(PR #374) 코드 상태와 대조해, 누락/미반영 유스케이스를 근거와 함께 식별하고 갱신 계획을 정의한다.
- 기준 시점: 2026-05-27 (main `cf19c94`).
- 참고: `docs/analysis/2026-05-27-codebase-snapshot/code/backend/httpapi.md`(라우트 전수표), `docs/analysis/2026-05-27-codebase-snapshot/02_sdlc_chain_status.md`(§2.2/§2.4 갭 G3/G4/G5).
- 원칙: 기존 UC ID 삭제/재번호 금지. 추가만.

---

## 1. 현재 UC 인벤토리 (system_usecases.md §2 기준)

| 섹션 | 도메인 | UC ID 범위 | 비고 |
| --- | --- | --- | --- |
| §2.1 | Auth/OIDC | UC-AUTH-01..03 | Keycloak 단일화 |
| §2.2 | Account | UC-ACCOUNT-01..03 | |
| §2.3 | Organization | UC-ORG-01..04 | |
| §2.4 | RBAC | UC-RBAC-01..03 | |
| §2.5 | Gitea Ingest/Snapshot | UC-GITEA-01..03 | push(webhook) 중심 |
| §2.6 | Command | UC-CMD-01..03 | |
| §2.7 | Audit | UC-AUD-01..02 | |
| §2.8 | Realtime | UC-RT-01..02 | |
| §2.9 | Application | UC-APP-01..10 | |
| §2.10 | Project | UC-PROJ-01..10 | |
| §2.11 | Dev Request (DREQ) | UC-DREQ-01..12 | |
| §2.12 | External Integration | UC-INT-01..14 | |
| §2.13 | Onboarding | UC-ONBOARD-01..11 | |

→ §2.14 부재. Repository 라이프사이클(draft→publish) + SCM↔시스템 양방향 연동을 담는 섹션이 없다.

---

## 2. 누락 UC 식별 (코드 근거)

### 2.A Repository SCM 연동 + Lifecycle (섹션 자체 부재 → §2.14 신규)

현재 §2.5 Gitea 는 webhook 수집(push)만 다루고, §2.9 Application 의 UC-APP-02/06/07/08 은 "Application-Repository 링크 + 운영지표 조회"만 다룬다. 2026-05-27 도입된 다음 흐름은 **어느 UC 에도 매핑되지 않는다**:

| 코드 사실 | 근거(httpapi.md / 라우트표) | 매핑 UC |
| --- | --- | --- |
| 원격 SCM repo 목록 조회 (imported 플래그) | `GET /api/v1/integration/providers/:id/scm-repositories` `listSCMRepositories` (API-88), httpapi.md §5.3 | (없음) |
| 선택 import (SCM 재조회 값 upsert, source=scm) | `POST .../import-repositories` `importSCMRepositories` (API-89) | (없음) |
| 시스템→SCM 저장소 생성 (push capability, source=system, `gitea.CreateRepo`) | `POST .../create-repository` `createSCMRepository` (API-90) | (없음) |
| repository draft 생성 (provider_key→provider_id FK 해석, migration 000045) | `POST /api/v1/repositories` `createRepositoryDraft`, httpapi.md §5.2 / domain.go:141 | (없음) |
| draft → publish 요청 (draft only → SCM/push/gitea-compat 검사 → `gitea.CreateRepo` → `UpsertRepository`) | `POST /api/v1/repositories/:id/publish` `requestRepositoryPublish` | (없음) |
| 소유권 구분 (scm-owned mirror vs system-owned 보존, ON CONFLICT) | `UpsertRepository` source/provider_id (#363/#366/#373), 02_sdlc §2.2 | (없음) |
| capability 기능 gate (import=pull / sync=pull\|sync / create=push + gitea-compat) | `scmProviderForCapability`/`isGiteaCompatibleProvider`, httpapi.md §5.3 | (없음) |

근거 보강: 02_sdlc_chain_status.md G4(repository draft→publish 무테스트 머지), G5(SCM 양방향 연동 IMPL 행 부재) 가 동일 갭을 지목.

→ **신규 §2.14 "Repository SCM 연동 + Lifecycle"** 으로 UC-REPO-01..07 발급.

관련 REQ: 신규 REQ-FR 발급 없음. 기존 REQ-FR-APP-002(0..N repo 연결)/004(외부 형상관리 연동)/009(provider 어댑터) + REQ-FR-INT-004(SCM 정규화)/012(write-back 양방향, 시스템→SCM 생성이 이에 해당) 에 매핑.

### 2.B §2.12 INT — 신규 INT UC (UC-INT-15.. 이어서)

기존 UC-INT-01(provider 등록), 03(webhook ingest), 04(scheduled pull), 05(SCM 정규화) 가 있으나 2026-05-27 깊이 확장은 미반영:

| 코드 사실 | 근거 | 기존 UC 와의 차이 |
| --- | --- | --- |
| auth_mode 별 자격증명 등록 (token/basic/oauth2/app_password/agent + write-only `api_token`/`auth_secret`, migration 000040/000041, `ResolveOutboundAuth`) | httpapi.md §5.3 / API-69..71 | UC-INT-01 은 "인증 모드 관리" 추상 언급뿐. auth_mode별 write-only secret 슬롯·outbound auth 해석은 별도 UC 가치 |
| 연결 테스트 (GET 5s timeout, SSRF 사내 수용) | `POST /api/v1/integration/test-connection` `testIntegrationConnection` (API-87), httpapi.md §5.3 / F-3 | UC 없음 |
| Gitea pull sync (per-provider, env fallback 금지, 주기/큐 integration_sync_jobs) | `syncIntegrationProvider` (API-72) + `internal/gitea/` worker, 02_sdlc §2.4/§2.7 | UC-INT-04 는 일반 scheduled pull. Gitea 전용 per-provider sync(provider 우선 env fallback)+SKIP LOCKED 큐는 구체화 가치 |
| webhook 헤더 alias (X-Integration-*→X-Gitea-*→X-Gogs-* fallback) | `ingestIntegrationProviderWebhook` (API-73), httpapi.md §5.3 (#358) | UC-INT-03 은 ingest 일반. 헤더 alias 정규화(헤더 불일치 정정)는 구체화 가치 |

→ UC-INT-15..18 추가(기존 UC-INT-01..14 유지, 다음 번호 이어서). 관련 REQ-FR-INT-001/002/003/004 재사용.

---

## 3. 계획

1. **§2.14 신규** "Repository SCM 연동 + Lifecycle": UC-REPO-01..07.
   - UC-REPO-01 원격 SCM repo 목록 조회 / 02 선택 import / 03 시스템→SCM 저장소 생성 / 04 repository draft 생성 / 05 draft→publish 요청 / 06 소유권(scm vs system) 구분 / 07 capability 기능 gate.
2. **§2.12 INT 확장**: UC-INT-15(auth_mode 자격증명 등록) / 16(연결 테스트) / 17(Gitea pull sync) / 18(webhook 헤더 alias) 추가.
3. **§1 모듈 표** 에 SCM repo 연동 모듈 row 보강(`integration_scm_repositories.go`, `domain.go` draft/publish).
4. **§3 규칙** 에 "API-87..90 + repository draft/publish 는 위 UC 에 매핑, 무테스트 머지분(#368)은 UT/TC 보강 carve 대상" 메모.
5. 메타 헤더 `최종 수정일` → 2026-05-27, 상태 line 갱신.
6. 기존 UC ID 일절 삭제/재번호 없음.

### 발급 예정 UC ID 범위
- 신규: **UC-REPO-01 ~ UC-REPO-07** (§2.14), **UC-INT-15 ~ UC-INT-18** (§2.12 추가).
