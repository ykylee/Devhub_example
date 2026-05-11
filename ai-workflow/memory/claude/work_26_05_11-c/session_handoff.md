# Session Handoff — claude/work_26_05_11-c (M1 PR-D)

- 문서 목적: `claude/work_26_05_11-c` sprint 의 세션 간 상태 인계
- 범위: M1 sprint 의 마지막 잔여 PR-D (T-M1-04 audit actor + request_id + audit_logs 마이그레이션)
- 대상 독자: 후속 에이전트, 프로젝트 리드
- 상태: planned (sprint 주제 결정 완료 2026-05-11. 결정 3건 대기)
- 브랜치: `claude/work_26_05_11-c` (HEAD `9b6b3ea`, main fast-forward)
- 최종 수정일: 2026-05-11

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

## 3. 진입 순서

1. 결정 3건 확정 → backlog §2 갱신
2. **PR-D**:
   a. 마이그레이션 작성 (`000008_audit_logs_actor_enrichment.{up,down}.sql`)
   b. domain `AuditLog` 필드 3개 추가 + store 갱신
   c. request_id 미들웨어 신설
   d. authenticateActor 가 source_type 분류 (Bearer→oidc, webhook bypass→webhook, dev fallback→system)
   e. recordAudit 가 request_id/source_ip/source_type 채움
   f. writeServerError 가 request_id 포함 로그
   g. 통합 테스트 — X-Request-ID 응답 헤더 == audit_logs.request_id
3. 검증 + sprint closure + PR

## 4. 위험 / 운영 의존

- **마이그레이션 운영 적용**: PoC 환경은 단순 `make migrate-up`. 운영은 down-time 정책 결정 필요 (M1 sprint plan §5 위험 표기).
- **logger 통합점**: writeServerError 의 log 형식 변경 — 기존 audit log 파싱 도구 영향 가능 (dev 환경엔 영향 없음).

## 5. 다음에 읽을 문서

- [본 sprint backlog](./backlog/2026-05-11.md)
- [M1 원본 sprint plan §T-M1-04](../../m1-sprint-plan/backlog/2026-05-08.md)
- [API 계약](../../../../docs/backend_api_contract.md)
- [backend 로드맵](../../../backend_development_roadmap.md)
- [직전 sprint closure (work_26_05_11-b)](../work_26_05_11-b/session_handoff.md)
