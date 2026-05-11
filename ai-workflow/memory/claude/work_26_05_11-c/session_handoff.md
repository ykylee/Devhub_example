# Session Handoff — claude/work_26_05_11-c (M1 PR-D)

- 문서 목적: `claude/work_26_05_11-c` sprint 의 세션 간 상태 인계
- 범위: M1 sprint 의 마지막 잔여 PR-D (T-M1-04 audit actor + request_id + audit_logs 마이그레이션)
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: **CLOSED 2026-05-11**. PR #57 (`4e831a3`) 머지 완료. M1 sprint 전 부채 청산.
- 브랜치: `claude/work_26_05_11-c` (sprint 산출물은 main 머지됨)
- 최종 수정일: 2026-05-11 (sprint closure)

## 0. 현재 기준선

- main HEAD `9b6b3ea` — 직전 sprint (work_26_05_11-b) closure 직후
- 본 sprint 의 첫 commit 이 메모리 정리 (work_26_05_11-b closure 표기 + 본 sprint baseline + backlog)
- 사용자가 다음 작업 묶음을 **M1 완결 (PR-D 단독)** 로 선택. M1 sprint 의 마지막 부채 청산.

## 1. 본 sprint 작업 축

[`./backlog/2026-05-11.md`](./backlog/2026-05-11.md) §3 PR-D 가 단일 source-of-truth.

T-M1-04 spec 요약:
- `audit_logs` 테이블에 컬럼 3개 추가: `source_ip TEXT`, `request_id TEXT`, `source_type TEXT` (`oidc | webhook | system`)
- `request_id` 미들웨어로 부착, `gin.Context` 에 저장, audit/log 양쪽 동일 값 사용
- 응답 헤더 `X-Request-ID` 출력 (운영 트레이싱)
- `writeServerError` 가 `request_id` 도 로그에 포함

## 2. PR 진입 전 결정 3건

| ID | 결정 | 영향 PR |
| --- | --- | --- |
| **DEC-1** | 기존 `audit_logs` 행의 `source_type` 처리 (NULL 허용 vs `'legacy'` backfill) | PR-D |
| **DEC-2** | `source_type` enum 범위 (`oidc|webhook|system` 만 vs `cli|api_token` 추가) | PR-D |
| **DEC-3** | `request_id` 형식 (UUIDv4 vs prefixed `req_<uuid>`) | PR-D |

세부는 [`./backlog/2026-05-11.md`](./backlog/2026-05-11.md) §2 참조.

## 3. 머지 결과 (2026-05-11 closure)

### 결정 확정
| ID | 채택 | 의미 |
| --- | --- | --- |
| **DEC-1** | A NULL 허용 | 기존 audit_logs 행은 backfill 안 함 |
| **DEC-2** | A `oidc | webhook | system` 만 | M1 spec 그대로 |
| **DEC-3** | B `req_<hex24>` prefix | realtime `evt_` 컨벤션 일관 |

### Commits
- `1bacea3` baseline (work_26_05_11-b closure + 본 sprint baseline)
- `51b7f05` PR-D 본체 (마이그레이션 000008 + AuditSourceType + requireRequestID + recordAudit/writeServerError 보강 + 3 신규 audit 테스트)
- `2bcd861` Codex P1 fix-up (`SetTrustedProxies(nil)` + IP 위조 방지 검증 테스트)
- `4e831a3` main 머지 (PR #57)

### 다음 sprint 후보 (PR-D follow-up hygiene)
- `store/postgres.go` 의 commands audit INSERT 3 곳 (586/773/1173) 도 actor context 채우기 — 본 sprint 범위 밖
- 모든 log 라인에 request_id 부착 (현재는 audit_logs + writeServerError 만)
- 운영 reverse proxy 환경용 `DEVHUB_TRUSTED_PROXIES` env 도입

## 4. 운영 의존 메모

- **마이그레이션**: 본 sprint PoC 가정. 운영은 zero-downtime 정책 결정 필요 (column ADD 만이라 사실상 영향 작음).
- **proxy 환경**: 현재 `SetTrustedProxies(nil)` 로 X-Forwarded-For 무시. 실 reverse proxy 환경 진입 시 env 기반 설정 필요.

## 5. 다음에 읽을 문서

- [본 sprint backlog](./backlog/2026-05-11.md)
- [M1 원본 sprint plan §T-M1-04](../../m1-sprint-plan/backlog/2026-05-08.md)
- [API 계약](../../../../docs/backend_api_contract.md)
- [backend 로드맵](../../../backend_development_roadmap.md)
- [직전 sprint closure (work_26_05_11-b)](../work_26_05_11-b/session_handoff.md)
