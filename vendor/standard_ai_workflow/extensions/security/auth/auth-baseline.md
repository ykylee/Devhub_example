# Auth Baseline Extension (v0.7.2, sub-cat)

- 문서 목적: standard_ai_workflow v0.7.2 의 auth-baseline extension. **sub-cat** of `security-baseline` — authentication 측면 (OAuth / API key / session token) 의 cross-cutting rule.
- 범위: 6 SEC-AUTH rule (SEC-WF-01~06 과 별도) + opt-in pattern + 8 smoke test
- 상태: stable (v0.7.2 도입)
- 최종 수정일: 2026-06-13
- 관련 문서: [`../security-baseline.md`](../security-baseline.md) (parent baseline), [`../SCHEMA.md`](../SCHEMA.md) (extension system SSOT), [`../security-baseline.opt-in.md`](../security-baseline.opt-in.md)
- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` 의 *authentication* 측면 (AIDLC 의 SECURITY-04, SECURITY-09, SECURITY-10, SECURITY-11, SECURITY-12, SECURITY-13 적응)

## §1 왜 Auth Baseline 이 필요한가

v0.7.0 step 8 의 `security-baseline` (6 SEC-WF rule) 은 *general security* — audit log, stage gate, question format, error handling, dependency integrity, R-9 skip marker. 그러나 **authentication** 측면 (API key, session token, OAuth scope) 의 *sub-cat* SSOT 가 없었음.

본 1차 출시: **auth-baseline 1종 (sub-cat)**, 우리 운영 컨텍스트에 applicable 한 6 rule.

## §2 6 Rule 정의

### 2.1 Rule SEC-AUTH-01: API Key Storage

**Rule**: harness API key (Claude / OpenAI / Google / Cohere 등) 는 *macOS keyring* 또는 *env var* 로만 저장. **plaintext file** (e.g. `~/.aws/credentials` 의 `[default]` 평문) 또는 *git commit message* 에 노출 금지.

**Verification**:
- `~/.myharness/providers.toml`, `~/.aws/credentials` 등이 `chmod 600` (소유자만 read)
- `printenv | grep -iE 'api[_-]?key|secret|token'` 실행 시 *empty* (env var 도 plaintext 위험)
- `git log -p | grep -iE 'sk-[a-zA-Z0-9]{20,}'` 시 0 매치 (git history 에 secret 없음)
- smoke test: `tests/check_auth_baseline.py` 의 `test_api_key_storage` 가 위 3가지 자동 검증

### 2.2 Rule SEC-AUTH-02: Session Token Rotation

**Rule**: session token (OAuth access_token / refresh_token / JWT) 은 *30일 미만* 주기로 rotation. rotation 시 *previous token 즉시 revoke*.

**Verification**:
- refresh_token 의 `exp` field 가 30일 (= 2592000 sec) 이내
- rotation log 가 `ai-workflow/memory/active/audit.md` 에 `event_type=token_rotation` 으로 append
- rotation event 가 `actor=user` (자동 rotation ❌)
- smoke test: `test_session_token_rotation` (3가지 자동 검증)

### 2.3 Rule SEC-AUTH-03: OAuth Scope 최소 권한

**Rule**: OAuth / API 호출 시 *최소 scope* 만 요청. `read:all` 또는 `admin:full` 같은 광범위 scope 사용 금지.

**Verification**:
- 각 harness 의 OAuth scope 가 5개 이하
- scope 가 `read` / `write` 의 fine-grained 분할
- 불필요한 scope (`admin`, `delete`, `*:all`) 미포함
- smoke test: `test_oauth_scope_minimal` (scope set 자동 검증)

### 2.4 Rule SEC-AUTH-04: 2FA / MFA 강제

**Rule**: production-grade workflow 의 *admin action* (release, state doc 변경, hard constraint disable) 시 2FA / MFA 인증 필수.

**Verification**:
- admin action 의 audit log entry 에 `mfa_verified: true` field
- 2FA 미통과 시 action 거부 (`require_explicit_approval` 와 통합)
- smoke test: `test_mfa_admin_action` (audit log 의 mfa_verified field 검증)

### 2.5 Rule SEC-AUTH-05: Password / Token Strength

**Rule**: token 또는 secret key 의 entropy ≥ 128 bit (= 32 hex char). `password123` 같은 약한 secret 사용 금지.

**Verification**:
- `secrets.token_hex(16)` (= 128 bit) 이상 생성
- entropy 측정: `secrets` 모듈의 `compare_digest` 또는 `hashlib.sha256`
- smoke test: `test_secret_strength` (각 secret 의 entropy 자동 측정)

### 2.6 Rule SEC-AUTH-06: Authentication Audit Log

**Rule**: 모든 authentication event (login / logout / token_refresh / mfa_challenge) 가 `audit.md` 에 append-only 기록. raw input 그대로 보존.

**Verification**:
- audit.md 의 `event_type` field 가 `auth_login` / `auth_logout` / `auth_token_refresh` / `auth_mfa_challenge` 4종 중 1
- actor field 가 user identifier
- `raw_input` field 가 PII redact (v0.7.1+ 의 privacy filter 통합)
- smoke test: `test_auth_audit_log` (4종 event 검증)

## §3 Compliance Summary

| Rule ID | Title | Status | Notes |
|---|---|---|---|
| SEC-AUTH-01 | API Key Storage | ✅ | macOS keyring + chmod 600 |
| SEC-AUTH-02 | Session Token Rotation | ✅ | 30일 rotation + audit log |
| SEC-AUTH-03 | OAuth Scope 최소 권한 | ✅ | 5개 이하 fine-grained |
| SEC-AUTH-04 | 2FA / MFA 강제 | ✅ | admin action + mfa_verified |
| SEC-AUTH-05 | Password / Token Strength | ✅ | entropy ≥ 128 bit |
| SEC-AUTH-06 | Authentication Audit Log | ✅ | 4종 event_type |

## §4 우리 사용 패턴 적응

| AIDLC 패턴 | 우리 적응 |
|---|---|
| AWS IAM | harness OAuth (Claude / OpenAI / Google) |
| AWS Cognito | (옵션) v0.7.3+ — 우리 SSO |
| AWS KMS | macOS keyring (Keychain) |
| API Gateway | MCP server 의 token 검증 |

**N/A 처리**: 우리 local runtime 은 cloud workload (Lambda / ECS / IAM role) 한정 rule (cross-account role, federated identity) N/A.

## §5 한계 / 예외

- **PoC / Prototype**: opt-in B) No 시 모든 rule advisory
- **CI/CD**: 자동 token rotation 미허용 (명시적 human approval)
- **External API**: OAuth scope 가 *provider 한정* (e.g. OpenAI `completions.write`)

## §6 Follow-up (v0.7.3+)

- SEC-AUTH-07: Hardware key (YubiKey) 통합
- SEC-AUTH-08: SSO / SAML (corporate 환경)
- v0.7.2: workflow_kit.common.auth helper (SEC-AUTH-01~06 evaluator)

## §7 References

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/extensions/security/baseline/security-baseline.md` (307 line, commit `b19c819`)
- 우리 SSOT: `extensions/SCHEMA.md` (v0.7.0 step 7, 200 line)
- 우리 parent: `extensions/security-baseline.md` (6 SEC-WF rule)
- 우리 wiki: `ai-workflow/wiki/concepts/extension-system.md` (210 line)
- 우리 검증: `tests/check_extension_system.py` (23 test PASS, v0.7.0)
- 우리 검증 (본 1차 출시): `tests/check_auth_baseline.py` (8 test PASS, v0.7.2)
- v0.7.1+ roadmap: `extensions/v0.7.1-roadmap.md` (sub-cat 구조)
