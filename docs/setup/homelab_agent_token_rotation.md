# HomeLab agent token rotation SOP

- 문서 목적: HomeLab agent 인증 token (`DEVHUB_HOMELAB_PULL_TOKEN`, `DEVHUB_INFRA_AGENT_TOKEN`) 의 발급, 배포, 주기적 rotation, 비상 revoke 절차 정의.
- 범위: backend-core 의 pull/push 양쪽 채널의 agent 인증 token 운영. backend ↔ HomeLab agent 신뢰 경계의 운영 책임.
- 대상 독자: 운영 담당자 (HomeLab agent + backend ops), External Integration 도메인 stakeholder.
- 상태: draft
- 최종 수정일: 2026-05-18
- 관련 문서: [ADR-0015 HomeLab adapter pull strategy](../adr/0015-homelab-adapter-pull-strategy.md) §4.3 + §6, [ADR-0016 Prometheus alerts policy](../adr/0016-prometheus-alerts-policy.md), `backend-core/internal/config/config.go` (`HomeLabPullToken`, `InfraAgentToken`).

## 1. 배경

HomeLab agent 와 DevHub backend 사이의 두 채널 — pull (`DEVHUB_HOMELAB_PULL_TOKEN`) + push (`DEVHUB_INFRA_AGENT_TOKEN`) — 은 공유 secret bearer token 으로 인증한다. token 이 leak / staleness 또는 운영 정책 변경 시 안전한 rotation 절차가 필요. ADR-0015 §6 (carve) 에서 "agent token rotation 정책" 으로 식별된 운영 항목.

## 2. Token 종류 (요약)

| Token | 환경변수 | 용도 | leak 시 영향 |
| --- | --- | --- | --- |
| **Pull token** | `DEVHUB_HOMELAB_PULL_TOKEN` | backend → HomeLab agent snapshot endpoint 요청 시 Bearer 헤더 | leak 자체는 HomeLab agent endpoint 의 enforcement 책임. snapshot 조회 권한 노출. |
| **Ingest agent token** | `DEVHUB_INFRA_AGENT_TOKEN` | HomeLab agent → backend `POST /api/v1/infra/services/snapshot` (API-77) 요청 시 Bearer 헤더 | backend infra_service_snapshots 의 위/변조 가능 — fake snapshot ingest → false alert / topology v2 오염 가능. **leak 시 즉시 revoke + rotation** |

## 3. 발급 절차

1. **strong random 생성** — 256-bit URL-safe 권장:
   ```bash
   openssl rand -base64 32
   # 또는
   python3 -c "import secrets; print(secrets.token_urlsafe(32))"
   ```
2. **저장 위치** — backend 운영 환경의 secret store (예: vault, systemd EnvironmentFile, .env 의 비공개 사본). git 추적 금지 ([CLAUDE.md] feedback_no_docker 정책 + .gitignore 의 DEV ENVIRONMENT 섹션 참조).
3. **배포** — backend-core 인스턴스 + HomeLab agent 양쪽에 동일 값 주입. 차이 발생 시 backend 가 인증 실패 (push 는 401 `infra_agent_unauthorized`, pull 은 HomeLab agent 가 401/403 반환 → degraded metric).
4. **검증** — token 적용 후:
   - **Pull**: backend log 의 `homelab pull request failed: status=401` 부재 + `devhub_homelab_pull_runs_total{result="success"}` counter 증가.
   - **Push**: HomeLab agent 가 `POST /api/v1/infra/services/snapshot` 호출 → backend 가 `202 Accepted` 응답 + audit log 의 ingest entry.

## 4. 주기적 rotation

### 4.1 권장 주기

- **운영 환경**: 90 일마다 rotation (기본 권장). 보안 정책에 따라 단축 가능 (예: 30 일).
- **개발 / stage**: 6 개월마다 또는 인스턴스 재구축 시.

### 4.2 zero-downtime rotation 절차

backend 와 HomeLab agent 사이에 약간의 시차가 있을 수 있어 단계별 진행:

1. **새 token 발급** (§3.1) — old / new 두 값을 동시에 보관.
2. **HomeLab agent 먼저 갱신** — agent 의 outgoing push token 을 new 로 변경. **이 시점에는 push 가 401** (backend 는 old 만 인증). 짧은 ingest 공백 (≤ 5 분) 허용 시 진행, 그렇지 않으면 다음 4.3 의 atomic rotation 활용.
3. **backend 갱신** — `DEVHUB_INFRA_AGENT_TOKEN` 을 new 로 변경 + backend-core 재시작 (또는 환경 reload). reload 직후 push 정상 복구.
4. **검증** (§3.4) + audit log 의 push 성공 entry 확인.
5. **old token 폐기** — vault / EnvironmentFile 에서 제거.

### 4.3 atomic rotation (downtime 0)

backend 가 두 token (old + new) 동시 수용을 지원하지 않는다 — 현재 단일 string compare. zero-downtime 이 강 요건이면:

- **옵션 A**: backend-core 다중 인스턴스 (load balancer 뒤) — 인스턴스별 sequential restart 로 token 교체 (전체 인스턴스 다운 0).
- **옵션 B**: `InfraAgentToken` 을 다중 값 (`,` separated) 으로 받는 enhancement — 본 SOP 시점 carve out (코드 변경 필요). 후속 ADR 후보.
- **옵션 C**: pull 채널만 사용 (push 인증 token 제거) — push endpoint 자체 비활성. ADR-0015 §1 의 두 채널 구조 변경 필요.

**현재 권장**: §4.2 의 단계별 + ingest 짧은 공백 허용 (HomeLab agent 의 retry/backoff 가 다음 tick 에 push 재시도하므로 5 분 이하 공백은 운영 영향 미미).

## 5. 비상 revoke

token leak (예: GitHub repo push 사고, log 노출) 발견 시:

1. **즉시 새 token 발급** (§3.1) 및 backend `DEVHUB_INFRA_AGENT_TOKEN` 갱신 + 재시작. **이 순서를 먼저** — old token 으로의 위/변조 push 차단.
2. **HomeLab agent 갱신** — new token 으로 변경 + agent 프로세스 재시작.
3. **audit log 검수** — `audit_logs` 의 `integration.snapshot.ingested` 또는 유사 action 에서 의심 ingest 패턴 (예: 비정상 snapshot_at, 비정상 IP) 점검. 의심 ingest 발견 시 `infra_service_snapshots` 의 해당 row 검수 / 정리.
4. **Pull token leak 케이스** — HomeLab agent endpoint 의 access log 점검 + 필요 시 agent 측 토큰도 rotation.
5. **사후 보고서** — leak 경로 + 영향 범위 + 후속 강화 (예: vault 도입, secret 검사 hook) 운영 일지에 기록.

## 6. 자동화 후보 (M4 carve out)

- secret scanner (예: `gitleaks` GHA workflow) 가 PR push 시 token 패턴 차단.
- vault / sealed-secrets / SOPS 등 secret management 도입 — 현재는 EnvironmentFile (운영 부담 ↑).
- backend 의 multi-token 수용 (옵션 B) — atomic rotation 단순화.

## 7. 변경 이력

| 일자 | 변경 | sprint |
| --- | --- | --- |
| 2026-05-18 | draft 1차 — ADR-0015 §6 carve (agent token rotation 정책) 의 운영 SOP 명문화 | `claude/work_260518-p` |
