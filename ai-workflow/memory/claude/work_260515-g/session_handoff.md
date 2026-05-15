# Session Handoff — claude/work_260515-g (DREQ-AuthADR)

- 브랜치: `claude/work_260515-g`
- Base: `main` @ `52f6ad8` (PR #121 머지 직후)
- 날짜: 2026-05-15
- 상태: in_progress
- 목적: DREQ 외부 수신 endpoint 인증 정책 ADR 결정 + 관련 문서 정합화.

## 결정 사항

- **채택**: 옵션 A — API 토큰 + IP allowlist
- **거부**: 옵션 B (HMAC), 옵션 C (OAuth client_credentials)
- **근거**: 1차 도입 비용 최소 + 외부 시스템 on-boarding 부담 최소. 운영 정책 (token rotation / revoke / IP allowlist 관리 / 감사 로깅) 으로 보안 보강.

## 작업 순서

1. (done) 메모리 초기화
2. (in progress) ADR-0012 작성
3. (planned) 관련 문서 6 갱신 (컨셉 §7/§9, REQ-NFR-DREQ-001, ARCH-DREQ-03, API §14.1, traceability ADR 인덱스 + DREQ row, roadmap M5)
4. (planned) PR + CI + self-review + merge

## 다음 sprint (본 PR 머지 후)

DREQ-Backend 진입 가능. ADR-0012 의 결정에 따라:
- `dev_request_intake_tokens` 테이블 도입 (token_id / hashed_token / client_label / allowed_ips JSON / created_at / revoked_at / last_used_at)
- intake auth middleware: header `X-DREQ-Token` + caller IP 검증 + last_used_at 갱신
- migration 000022 (dev_requests) + 000023 (intake tokens)
- API-59 의 handler 가 token 인증 통과 후 routePermissionTable Bypass 처리
