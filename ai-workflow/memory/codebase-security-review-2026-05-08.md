# DevHub Example — 저장소 전체 보안 리뷰 리포트 (2026-05-08)

- 문서 목적: Claude Code `/security-review` 도구로 수행한 **저장소 전체 코드베이스** 보안 검토 결과를 보고용으로 정리한다.
- 범위: backend-core (Go), backend-ai (Python/FastAPI), frontend (Next.js/TypeScript), ai-workflow scripts/tests (Python), 인프라/스크립트(Makefile, gitignore, global-snippets) 까지 포함. 문서/예제 JSON/릴리즈 노트는 정책상 검토 제외(skill FP rule §16, §11).
- 대상 독자: 프로젝트 리드, 보안 담당자, 운영 담당자, 후속 인증/권한 트랙 PM
- 상태: stable
- 최종 수정일: 2026-05-08
- 관련 문서: [PR-12 보안 리뷰 (변경 한정)](./PR-12-security-review.md), [PR-12 액션 플랜](./PR-12-review-actions.md), [SEC backlog](./claude/test/backend-integration/backlog/2026-05-08.md)

---

## Executive Summary

| 항목 | 결과 |
| --- | --- |
| **검토 대상** | 저장소 전체 (PR diff 가 아닌 모든 소스코드) |
| **검토 브랜치 / HEAD** | `test/backend-integration` / `fe593e7` |
| **신규 high-confidence 취약점** | **0 건** |
| **이미 식별된 high 취약점** | **1 건** — Auth 통합 결함 (SEC-2 외 SEC-3/SEC-4 신설 권장) |
| **방어적 개선 권고 (보고 임계 미달)** | 1 건 — DB 에러 메시지 그대로 노출 |
| **종합 판정** | 인증/권한 영역 외에는 **양호**. SQL/명령/역직렬화/경로/암호 영역에서 일관된 안전 패턴 확인. |

본 리포트는 코드베이스 전체를 대상으로 한 결과로, PR diff 한정 리포트(`PR-12-security-review.md`) 와 **상호 보완 관계**다. PR diff 한정 리포트가 *변경* 의 위험을 다룬다면, 본 리포트는 *현재 상태* 의 위험을 다룬다.

---

## 1. 검토 정보

| 항목 | 값 |
| --- | --- |
| 도구 | Claude Code `/security-review` (`security-review` skill) |
| 모델 | Claude Opus 4.7 (1M context) |
| 호출 시점 | 2026-05-08 |
| 검토 대상 브랜치 | `test/backend-integration` |
| HEAD 커밋 | `fe593e7 chore(repo): drop source-docs/workflow-source after split to PR #13` |
| 비교 base | `origin/main` (해당 사항 없음 — 코드베이스 전체 검토) |
| 검토 단계 | (a) 1차 후보 식별 (sub-agent) → (b) 후보별 false-positive 필터링 (병렬 sub-agent) → (c) confidence ≥ 8 만 채택 |

### 1.1 적용 가이드라인

skill spec 의 다음 가이드라인을 그대로 적용했다.

- **하드 제외 17 항** (DOS, 시크릿 디스크 저장, 레이트 리미팅, 메모리/CPU 고갈, 비핵심 입력 검증, GitHub Action input, 하드닝 결여, 이론적 race, 구버전 라이브러리, 메모리 안전 언어 메모리 안전, 단위 테스트 단독, 로그 스푸핑, path-only SSRF, AI 시스템 프롬프트 사용자 입력, 정규식 인젝션·DOS, 문서 단독, 감사 로그 결여)
- **선례 12 항** (URL 로깅 안전, UUID 추측 불가, 환경변수/CLI 신뢰, 자원 누수 무관, React/Angular XSS 한정, GitHub Action 보수, 클라이언트 측 권한 미흡 무관, MEDIUM 은 명백할 때만, ipynb 보수, 비-PII 로깅 안전, 셸 스크립트 cmd 인젝션 보수)
- **신뢰도 임계** ≥ 8 만 채택

### 1.2 검토 카테고리

- 입력 검증: SQL / Command / XXE / Template / NoSQL injection, Path traversal
- 인증/권한: Auth bypass, privilege escalation, session, JWT, authz bypass
- 암호/시크릿: 하드코딩, 약한 알고리즘, RNG, 인증서 검증 우회
- 코드 실행: 역직렬화 RCE, pickle, YAML, eval, **React unsafe 메서드 한정 XSS**
- 데이터 노출: 시크릿/PII 로깅, 디버그 정보 누출

---

## 2. 알려진 결함의 상세 (재진술)

본 리뷰에서도 가장 큰 위험으로 *재확인* 되었으며, 이미 PR-12 검토에서 채택된 항목이다. **신규 발견은 아니지만 보고용 완결성을 위해 상세 전체를 본 리포트에도 그대로 수록한다.**

### Vuln A — Auth 통합 결함 (Severity High, Confidence 1.0)

**위치 (4 개 파일에 분산)**

- `backend-core/internal/httpapi/auth.go:24-44` — `authenticateActor` 미들웨어
- `backend-core/main.go:47-63` — `RouterConfig` 생성
- `backend-core/internal/httpapi/router.go:65-99` — `/api/v1` 그룹에 미들웨어 마운트
- `backend-core/internal/httpapi/commands.go:271-287` — `requestActor`, `X-Devhub-Actor` 헤더 fallback

**구성 요소** (네 가지 결함의 통합)

1. **무인증 통과 (auth.go:26-29)** — `Authorization` 헤더가 비어 있으면 그대로 `c.Next()` 호출. 헤더만 안 붙이면 인증 검사 자체를 건너뜀.
2. **Bearer verifier 미주입 (main.go:47-63)** — `RouterConfig` 생성 시 `BearerTokenVerifier` 필드를 설정하지 않음. 따라서 `Bearer <임의값>` 토큰이 붙어 있어도 `auth.go:41-44` 의 dev fallback 분기로 떨어져 무검증 통과. 본 PR 의 `main.go:46` 에 SEC-2 주석으로 사실 자체는 명시됨.
3. **권한(role) 미적용** — `auth.go:73` 에서 `devhub_actor_role` 을 gin context 에 set 하지만, 어느 핸들러도 이 값을 읽지 않는다. `defaultRBACPolicy()` 의 role × resource 매트릭스 정의는 있으나 실제 라우팅 결정에 연결되지 않음.
4. **Actor 스푸핑 (commands.go:271-287)** — 인증 컨텍스트가 비어 있을 때 `X-Devhub-Actor` 헤더 값을 그대로 `actor_login` 으로 사용. 무인증 상태에서 audit/command 의 행위자 신원을 임의로 위조 가능.

**영향 범위**

미들웨어가 `/api/v1` 그룹 전체에 마운트(`router.go:65-66`) 되어 있어 PR #12 가 추가한 모든 mutating 라우트가 무인증·무권한 접근 가능:

- `PATCH /api/v1/users/:user_id`, `DELETE /api/v1/users/:user_id`
- 조직 단위 CRUD: `POST/PATCH/DELETE /api/v1/organization/units...`, `PUT /api/v1/organization/units/:unit_id/members`
- `POST /api/v1/admin/service-actions`
- `POST /api/v1/risks/:risk_id/mitigations`
- `GET /api/v1/audit-logs` (감사 이력 노출)
- `GET /api/v1/realtime/ws` (실시간 채널)

저장소 정책상 native 배포에 별도 게이트웨이가 없다 (BLK-1 결정 + `feedback_no_docker.md` 정책).

**익스플로잇 시나리오**

1. 공격자가 backend 에 네트워크 도달.
2. `PATCH /api/v1/users/<victim>` 본문 `{"role":"system_admin"}` 전송, Authorization 헤더 없음. `updateUser` 가 `validAppRoles` 검사를 통과시키고 그대로 저장 → 임의 계정을 System Admin 으로 승격.
3. 또는 `POST /api/v1/admin/service-actions` 본문 `{"service_id":"prod-db","action_type":"shutdown","reason":"x","force":true,"dry_run":false}` 전송 → command 즉시 큐잉.
4. `GET /api/v1/audit-logs` 로 감사 이력 전체 열람.
5. `X-Devhub-Actor: <임의 사용자>` 헤더로 위조된 actor 로 행위 기록 작성.

**권고 조치**

1. **Fail-fast 가드** — `main.go` 또는 config layer 에서 `BearerTokenVerifier == nil` 일 때 prod 환경 startup 거부. dev 모드는 명시적 환경변수(예: `DEVHUB_ENV=dev` 또는 `DEVHUB_AUTH_DEV_FALLBACK=1`) 가 켜져 있을 때만 허용.
2. **무인증 통과 분기 제거** — `auth.go` 의 *empty Authorization → c.Next()* 분기 제거 또는 prod 에서 401. 인증 불필요 라우트(`/healthz`, `/api/v1/snapshot` 등)는 별도 화이트리스트.
3. **`X-Devhub-Actor` fallback 제거** — `requestActor` 의 헤더 fallback 을 dev 모드에서만 허용하고 prod 에서는 무시. `commands.go` 의 `Warning` 응답 헤더만으로는 감사·명령 컬럼 위조를 막지 못한다.
4. **Role 강제 미들웨어** — `devhub_actor_role` 을 읽어 `defaultRBACPolicy()` matrix 와 매칭하는 라우트별 가드 추가. 최소한 다음 라우트:
   - `/admin/*` → `system_admin`
   - 사용자 PATCH/DELETE → `system_admin`
   - `GET /audit-logs` → `manager` 이상
   - 조직 단위 CRUD → `system_admin`
   - command 생성 → `manager` 이상
5. **회귀 테스트 보강** — `auth_test.go` 에 prod 가드(env unset + verifier nil → 거부) 케이스 + role 가드 통합 테스트 케이스 추가.

**SEC backlog 매핑**

| 결함 요소 | 기존 SEC 항목 | 추가 처리 필요 |
| --- | --- | --- |
| 무인증 통과 + verifier 미주입 | SEC-2 | DoD 에 *empty Authorization 분기 제거* + *fail-fast* 항목 추가 |
| Role 미적용 | **신규 — SEC-3 신설 권장** | backlog 등록 + 라우트별 role 가드 미들웨어 |
| Actor 스푸핑 | **신규 — SEC-4 신설 권장** | backlog 등록 + `X-Devhub-Actor` fallback 제거 |
| Frontend mock 로그인 (단독으로는 비취약) | SEC-1 | 정책상 단독 취약점 아니지만 backlog 추적 유지 |

---

## 3. 신규 high-confidence 발견

**없음.** 저장소 전체에서 1차 식별 sub-agent 가 도출한 신규 후보는 0건이다. 따라서 false-positive 필터링 단계도 자명하게 통과(채택 0).

---

## 4. 영역별 클린 판정 근거 (Positive Coverage)

`/security-review` 가 *발견하지 못해서* 0건이 아니라, *각 영역에서 안전 패턴을 능동적으로 확인했기 때문에* 0건임을 보고용으로 정리한다.

### 4.1 SQL / Database (backend-core)

| 위치 | 패턴 | 검증 결과 |
| --- | --- | --- |
| `backend-core/internal/store/postgres.go` | `pgx` 매개변수 바인딩 (`$1, $2, …`). 모든 SELECT/INSERT/UPDATE/DELETE 가 prepared statement 형태 | clean |
| `backend-core/internal/store/audit_logs.go` | 동일 패턴, payload 는 `::jsonb` cast 후 매개변수로 주입 | clean |
| `fmt.Sprintf` UPDATE 빌더 (postgres.go 의 일부) | 동적 SET 절 구성 시 컬럼명을 *서버 제어 화이트리스트* 에서만 인라인. 외부 입력은 매개변수로만 들어감 | clean |
| `backend-core/cmd/idp-apply-schemas/main.go` | CLI flag (신뢰 입력) 로 SQL 파일 경로를 받아 `pool.Exec` 호출. 외부 영향 없음 (FP precedent §3 — env/CLI 신뢰) | clean |

### 4.2 명령 실행 / 외부 프로세스

| 위치 | 검증 항목 | 결과 |
| --- | --- | --- |
| `backend-core/internal/commandworker/worker.go` | `os/exec`, `syscall.Exec`, 외부 셸 호출 사용처 — **없음**. DB 폴링 + 상태 전이 + WebSocket publish 만 | clean |
| `ai-workflow/scripts/*.py` (`bootstrap_workflow_kit.py`, `export_harness_package.py`, `generate_workflow_state.py`, `run_demo_workflow.py`, `run_existing_project_onboarding.py`, `scaffold_harness.py`) | `subprocess.*(shell=True)`, `os.system`, `commands.getoutput` — **없음**. 셸 호출이 있는 곳도 인자 리스트 형태로 안전 호출 | clean |
| `tests/repro_validation.py` | skill FP rule §11 (테스트 단독 파일 제외). 그럼에도 검토 결과 외부 입력 미관여 | excluded |
| `backend-ai/main.py` | 현 시점 헬스 엔드포인트 + TODO 만 있는 스텁. 처리할 사용자 입력 없음 | clean |

### 4.3 역직렬화 / 코드 실행

| 검색어 / 패턴 | 결과 |
| --- | --- |
| Python `pickle.loads`, `pickle.load`, `cPickle` | 사용처 없음 |
| Python `yaml.load(...)` (unsafe loader) | 사용처 없음 (`safe_load` 사용 또는 미사용) |
| Python `eval(`, `exec(` (런타임 코드 실행) | 사용처 없음 |
| Go `gob.NewDecoder`, `xml.Unmarshal` | 사용처 없음 (JSON/protobuf 만) |
| Go `json.Unmarshal` 의 대상 구조체 | 모두 정의된 타입에 대한 unmarshal — 임의 타입 역직렬화 없음 |

### 4.4 경로 처리 / 파일 I/O

| 위치 | 검증 결과 |
| --- | --- |
| `backend-core/cmd/idp-apply-schemas/main.go` | CLI flag 로 받은 경로 → CLI 신뢰 입력 |
| `ai-workflow/scripts/export_harness_package.py` | 출력 경로는 신뢰된 인자 + `pathlib.Path` 정상화 |
| 기타 `os.ReadFile`, `os.Open` 사용처 | 모두 컴파일 시점 상수 또는 신뢰 인자. 사용자 입력 기반 path traversal 없음 |

### 4.5 인증/세션 패턴 (Auth 통합 결함 외)

| 위치 | 검증 결과 |
| --- | --- |
| `backend-core/internal/httpapi/gitea_webhook.go` | HMAC-SHA256 signature 검증, **`hmac.Equal` 상수시간 비교** 사용. 빈 시크릿/시그니처 거부. replay 방지는 `dedupe_key` 로 처리 | clean |
| `backend-core/internal/httpapi/realtime.go` (WebSocket) | `gorilla/websocket` upgrader 의 `CheckOrigin` 동일 origin 강제. 권한 자체는 SEC-2 결함이 모든 권한 차이를 무력화하므로 별도 CSWSH 위험 *추가* 없음 — Vuln A 와 통합 |
| Cookie / 세션 쿠키 처리 | 본 코드에는 없음 (Bearer 만) |

### 4.6 암호 / 시크릿

| 검사 항목 | 결과 |
| --- | --- |
| ID/토큰 생성기 | `crypto/rand` 사용 (postgres.go 의 `rand.Read` + hex). `math/rand` 사용처 없음 |
| TLS 인증서 검증 우회 | `tls.Config{InsecureSkipVerify: true}` 사용처 없음 |
| 약한 해시 (MD5, SHA1) 의 보안 용도 사용 | 없음 (있다면 dedupe key 등 비보안 용도) |
| 하드코딩 시크릿 | 코드 본체에 API 키/토큰/비밀번호 없음. `docker-compose.yml` 의 `user/pass` placeholder 는 BLK-1 결정으로 git 추적 외부 |
| `.env*` 파일 처리 | `.gitignore` 가 `.env`, `.env.*` 차단 (예외는 `.env.example`, `.env.template`) |

### 4.7 Frontend (Next.js / React)

| 검사 항목 | 결과 |
| --- | --- |
| `dangerouslySetInnerHTML`, `bypassSecurityTrust*`, `innerHTML` 직접 할당 | 사용처 없음 — React 자동 탈출만 사용 |
| `target="_blank"` + `rel` 누락 (tabnabbing) | precedent §5 — 보고 임계 미달 (subtle 항목은 매우 high confidence 일 때만) |
| URL/origin 동적 사용 (`websocket.service.ts`, `*.service.ts`) | host 는 `NEXT_PUBLIC_*` 환경변수 (신뢰), path 는 사용자 입력 미관여 |
| 클라이언트 측 권한/인증 미흡 (mock 로그인) | precedent §8 — 클라이언트 측 권한 미흡은 단독 취약점 아님. SEC-1 backlog 추적 유지 |

### 4.8 로깅 / 데이터 노출

| 검사 항목 | 결과 |
| --- | --- |
| 시크릿/토큰/비밀번호 로깅 | 없음. 로그는 동작 메시지 또는 식별자(URL, ID) 뿐 |
| PII 로깅 | 사용자 이메일이 audit_logs payload 에 포함되지만, 이는 정상 *감사 데이터* 로 사용자가 의도한 보존 (기능 자체) — 로깅 누출이 아님 |
| 디버그 모드 / verbose 정보 누출 | 별도의 디버그 endpoint 없음. 단, **DB 에러 메시지가 그대로 응답 본문에 포함** — §5 의 minor info leak 참고 |

---

## 5. 보고 임계 미달 — 방어적 개선 권고 (Defense-in-depth)

본 리뷰의 채택 임계(confidence ≥ 8) 에는 미치지 못하지만 보고 완결성을 위해 기록한다.

### 5.1 DB 에러 메시지의 응답 본문 노출

**위치**

다수의 backend-core 핸들러:

- `organization.go` 의 `c.JSON(http.StatusInternalServerError, gin.H{"status":"failed","error": err.Error()})` 패턴
- 유사 패턴이 `commands.go`, `audit.go`, `domain` 핸들러에도 분포

**관찰**

`pgx` 에러는 자격증명을 포함하지 않지만, **스키마/제약/내부 SQL 구문 단편** 을 외부에 노출할 수 있다. 예: `duplicate key value violates unique constraint "users_email_key"`, `column "primary_unit_id" of relation "users" does not exist` 등. 이는 high-confidence "데이터 노출" 취약점 기준에는 미치지 않지만 (시크릿/PII 가 아님), 외부 공격자에게 스키마 정찰 단서를 제공할 수 있다.

**권고 (선택적, 비-블로커)**

- prod 빌드에서는 5xx 응답 본문에 `err.Error()` 대신 일반화된 메시지 (`internal error`) 만 노출하고, 상세는 서버 로그로만 기록.
- 또는 에러 분류기를 두어 *사용자에게 의미 있는 에러* (제약 위반 등) 만 응답에 노출하고 나머지는 마스킹.

---

## 6. 검토 흐름 (방법론 투명성)

skill spec 의 3 단계 흐름을 그대로 적용했다.

```
[1단계: 1차 후보 식별 — sub-agent]
  ↓ 코드베이스 전체 + 카테고리 + FP 룰 + 출력 형식 전달
  ↓ Read/Grep/Glob 으로 핵심 파일 컨텍스트 분석
  → 도출 후보: 0 건 (NO_FINDINGS)

[2단계: 후보별 false-positive 필터링 — 병렬 sub-agent]
  → 후보가 0 이므로 자명하게 통과

[3단계: confidence ≥ 8 채택]
  → 채택: 0 건 (신규)
  → 알려진 결함 (Vuln A) 은 PR-12 검토에서 이미 채택, 본 리뷰가 재확인
```

본 리뷰는 sub-agent 가 능동적으로 다음 패턴을 grep/read 로 확인했다.

- SQL 빌더 패턴: parametrized 여부, `fmt.Sprintf` 호출 컨텍스트
- 외부 프로세스 호출: `os/exec`, `subprocess`, `os.system`, `commands.getoutput`
- 역직렬화 위험: `pickle`, `yaml.load`, `eval`, `exec`, `gob`, `xml.Unmarshal`
- 암호 자산: `crypto/rand` vs `math/rand`, `tls.Config`, `InsecureSkipVerify`, 하드코딩 시크릿
- 웹 위험: `dangerouslySetInnerHTML`, `bypassSecurityTrust*`, `innerHTML`
- Webhook 시그니처 검증: `hmac.Equal`, 빈 시크릿 거부 여부

---

## 7. 결론과 권고

### 7.1 결론

- 저장소 전체에서 **신규 high-confidence 보안 취약점은 0 건**.
- 가장 큰 위험은 **이미 식별된 Auth 통합 결함** (Vuln A, SEC-2 + 신규 SEC-3/SEC-4) — 본 리뷰가 이를 재확인.
- SQL/명령/역직렬화/경로/암호 영역은 일관된 안전 패턴 확인.
- 그 외 보고 임계 미달의 방어적 개선 항목 1건 (§5.1) — 선택적.

### 7.2 다음 단계

| 우선순위 | 액션 | 위치 / 책임 |
| --- | --- | --- |
| **P0** | Vuln A 의 권고 5 항을 다음 sprint (인증/권한 트랙 PR) 에서 일괄 해소 | backend-core 팀 |
| **P0** | SEC-3, SEC-4 backlog 신설 후 SEC-1/2 와 함께 동일 sprint 에 묶음 | `claude/test/backend-integration/backlog/2026-05-08.md` 갱신 |
| **P1** | 인증/권한 트랙 PR 머지 후 `/security-review` 재실행 (회귀 검증) | 다음 PR 머지 시점 |
| **P2** | §5.1 의 DB 에러 메시지 마스킹은 보안 강화의 일환으로 별도 backlog 또는 follow-up | backend-core |
| **P3** | 본 리뷰는 자동 도구 결과. 머지 전 보다 깊은 분석을 원하면 `/ultrareview` (다중 에이전트 클라우드 리뷰, 별도 과금) 실행 검토 | 사용자 결정 |

### 7.3 적용 보장 한계

- 본 리뷰는 **정적 분석 + 휴리스틱** 기반이며, 동적 분석/퍼징/exploit POC 는 포함하지 않는다.
- 외부 의존(`pgx`, `gin`, `gorilla/websocket`, `next`, `react`) 라이브러리 자체의 CVE 는 본 리뷰 범위 외. SCA(예: `govulncheck`, `npm audit`, `pip-audit`) 별도 권장.
- 사내 SSL inspection 환경에서 backend-core build 가 완료되지 않은 상태로 검토 — 컴파일 오류로 가려진 추가 면이 있을 가능성은 낮지만 0 은 아님.

---

## 부록 A. 검토 대상 파일 영역

- `backend-core/internal/httpapi/*.go` — auth, audit, commands, organization, rbac, realtime, router, gitea_webhook, snapshot, runtime_snapshot_provider, domain
- `backend-core/internal/store/*.go` — postgres, audit_logs
- `backend-core/internal/normalize/*.go`, `backend-core/internal/commandworker/*.go`, `backend-core/internal/config/*.go`
- `backend-core/cmd/idp-apply-schemas/main.go`, `backend-core/main.go`
- `backend-ai/main.py`
- `frontend/app/**`, `frontend/components/**`, `frontend/lib/services/*.ts`, `frontend/lib/store.ts`
- `ai-workflow/scripts/*.py`, `ai-workflow/tests/*.py`
- `tests/repro_validation.py`, `tests/mock_backlog.md` (문서)
- 인프라: `.gitignore`, `Makefile`, `global-snippets/codex/*`

## 부록 B. 적용한 하드 제외와 선례 일람

§1.1 와 동일. 본 리뷰가 *어떤 항목을 보고하지 않았는지* 의 근거는 모두 거기에 있다.

---

*본 리포트는 자동 도구 결과를 기반으로 한 보안 검토 결과이며, 보안 인증·감사를 대체하지 않는다. 실제 운영 배포 전에는 별도의 보안 검수 절차(외부 침투 테스트, SCA, DAST 등) 를 거치기를 권장한다.*
