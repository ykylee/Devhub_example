# §11.1 Incident Runbook Index

> umbrella doc §11.1 — 6 incident runbook + 운영 runbook 진입점

## Runbook 목록

| # | Runbook | Trigger | Severity | MTTR target | 사외/사내 |
| --- | --- | --- | --- | --- | --- |
| §11.1.1 | [Source plugin sync 실패](./11.1.1-source-plugin-sync-failure.md) | sync HTTP 4xx/5xx / timeout | warning → critical | 15~30분 | 사외 |
| §11.1.2 | [Credential 만료](./11.1.2-credential-expired.md) | health check 401/403 | warning → critical | 30분~1시간 | 사내 (credential) |
| §11.1.3 | [Pi ingest pipeline timeout](./11.1.3-pi-ingest-timeout.md) | Pi subprocess hang | warning → critical | 30분~2시간 | 사내 (Pi vendor) |
| §11.1.4 | [Retention cron 실패](./11.1.4-retention-cron-failure.md) | cron fail / quota 90%+ | warning → critical | 1시간 | 사외 |
| §11.1.5 | [Integrity violation](./11.1.5-integrity-violation.md) | sha256 mismatch | high → critical | 30분 | 사외 |
| §11.1.6 | [Archive trigger 실패](./11.1.6-archive-trigger-failure.md) | superseded archive fail | warning → critical | 1시간 | 사외 |

## 공통 패턴

모든 runbook 5 단계:
1. **Trigger**: incident 발생 조건
2. **Detection**: 자동/수동 detection 방법 + audit log event
3. **Triage**: 원인 식별 절차
4. **Mitigation**: 즉시 적용 가능한 fix
5. **Recovery**: 정상 복구 + alert 해제 + audit log

## 3-tier alert routing (umbrella doc §11.3)

| Severity | Channel | Response time | Escalation |
| --- | --- | --- | --- |
| info | Slack `#backend-knowledge-info` (M-v0.2.1+) | 1 business day | (없음) |
| warning | Slack `#backend-knowledge-alerts` + on-call | 1시간 | on-call responder (§11.4) |
| critical | Slack `#backend-knowledge-critical` + page | 15분 | on-call → 30분 team lead → 1시간 director |

## Related

- umbrella doc §11.3 (Monitoring + alert routing)
- umbrella doc §11.4 (On-call 운영 + role 정의)
- umbrella doc §3.6.6.1 (7 audit event type)
- umbrella doc §3.7.6 (data normalization pipeline 자동 검증)
