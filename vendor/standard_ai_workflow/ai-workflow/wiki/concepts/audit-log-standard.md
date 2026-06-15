---
type: concept
status: active
last_ingested_from: workflow-source/core/audit_log_standard.md + workflow_kit/common/contracts/stage_gate.py
related_pages: [concepts/extension-system, concepts/stage-gate-pattern, concepts/reverse-engineering, decisions/adr-001-3-layer-separation, topics/aidlc-benchmark-analysis-2026-06-12]
created: 2026-06-13
updated: 2026-06-13
---

# Audit Log Standard (v0.7.0 step 10)

- 문서 목적: standard_ai_workflow v0.7.0 step 10 의 per-project audit log 표준. 분산 정의된 audit log 정책을 단일 표준 spec 으로 통합 + 2 latent bug fix.
- 범위: per-project `audit.md` 정책, 8 field, append-only 강제, lifecycle, 자동화 hook
- 최종 수정일: 2026-06-13

## §1 TL;DR  {#s1-tldr}

| # | 항목 | 값 |
|---|---|---|
| 1 | 외부 spec | `workflow-source/core/audit_log_standard.md` (200 line, v0.7.0 stable) |
| 2 | runtime helper | `workflow_kit/common/contracts/stage_gate.py` 의 `append_audit_log()` (updated) |
| 3 | smoke test | `workflow-source/tests/check_audit_log_compliance.py` (290 line, 13 test PASS) |
| 4 | Source 1차 출처 | AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/build-and-test.md` (audit log 정책) |
| 5 | 도입 버전 | v0.7.0 (commit `54e96a9`) |
| 6 | Trigger | v0.7.0 stage_completion required 격상에 따른 audit log mandatory 표준화 |
| 7 | 2 latent bug fix | leading newline + microsecond leak |
| 8 | 관련 토픽 | [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 H (Audit Log 표준화) |

## §2 왜 Audit Log 표준이 필요한가  {#s2-why}

v0.6.4-5 의 분산 정의된 audit log 정책:
- `wiki-prompt-log` 가 L1 raw mirror 만. in-project audit 채널 ❌
- stage_gate 의 `append_audit_log()` 의 2 latent bug (leading newline, microsecond leak)
- timestamp 형식 비일관 (datetime.now() vs ISO 8601)

v0.7.0 stage_completion required 격상에 따라 audit log 도 mandatory 표준화.

## §3 Audit Log 위치 + Format  {#s3-location-format}

### 3.1 Primary 위치

`ai-workflow/memory/active/audit.md` — per-project audit 채널. wiki-prompt-log 와 role 분리:
- `audit.md`: short-term (현재 project 의 in-memory layer)
- `vault/~/wiki/`: long-term archive (wiki-prompt-log 의 raw mirror)

### 3.2 Format (8 field)

```markdown
## [Stage: <name>] [<ISO 8601 timestamp>]

- **event_type**: opt_in | stage_completion | user_response | compliance_check | ...
- **stage_name**: <name>
- **actor**: <user|orchestrator|default>
- **raw_input**: <verbatim user input, no trimming>
- **decision**: <approved|rejected|request_changes>
- **reason**: <optional free text>
- **source_context**: <file path or session id>
- **notes**: <optional additional context>
```

### 3.3 Header (Stage prefix)

- `## [Stage: X] [<ISO 8601>]` 형식
- e.g. `## [Stage: design] [2026-06-13T22:00:00+09:00]`

## §4 Append-only 강제 (CRITICAL)  {#s4-append-only}

- 한 번 기록된 entry 는 *수정/삭제 금지*
- 같은 entry 중복 시 dedup (idempotency 보장) — `f(f(x)) = f(x)`
- read-modify-write 만 허용 (audit log 외 다른 file 수정)
- file lock (atomic_write + flock) 으로 동시 append 방지
- corruption 시 `audit.md.corrupt.<ts>` 로 격리 + 새로 시작

## §5 Latent Bug Fix  {#s5-bug-fix}

### 5.1 Leading newline bug

- 증상: `append_audit_log()` 호출 시 첫 호출 entry 앞에 `\n` 삽입
- 원인: file open mode 의 newline handling
- fix: explicit `open(path, "a", encoding="utf-8", newline="\n")` → leading newline 제거
- 영향: audit log reader 가 첫 entry 의 header 인식 실패 → reject

### 5.2 Microsecond leak bug

- 증상: `normalize_timestamp()` 가 `datetime.now()` 의 microsecond 포함 → ISO 8601 string 길이 가변
- 원인: `fromisoformat()` 가 microsecond 포함 string 만 parse
- fix: 명시적 `.isoformat(timespec="seconds")` 또는 `strftime("%Y-%m-%dT%H:%M:%S%z")`
- 영향: state.json 의 reverse_engineering.last_generated 비교 시 type mismatch

## §6 Lifecycle  {#s6-lifecycle}

- **Retention**: 90일 default, v0.7.1+ 설정 가능
- **Freeze**: 90일 후 archive (`audit.md.archive.<ts>`) + wiki raw mirror 로 propagate
- **Privacy**: raw input 그대로 보존 — PII 포함 가능, MD5 hash 로 dedup 만
- **Auto cleanup**: `stale-90day-lint` 가 R-5 정책으로 자동 detect + alert

## §7 자동화 Hook  {#s7-automation}

- **Runtime**: `stage_gate.append_audit_log()` 호출 시 자동 write
- **Orchestrator**: stage transition 시 자동 emit (session-start, design → impl, etc.)
- **Linter**: `stale-90day-lint` 가 90일 stale detect + rotate 권고
- **Wiki sync**: vault 의 `~/wiki/log.md` 가 wiki-event-sync 로 자동 propagate

## §8 한계 / 예외  {#s8-limitations}

- **CI/cron 자동 approval**: 1+ case 보호. ensure_stage_completion() 의 auto-approval 은 audit log 기록
- **PII**: raw input 그대로 보존. privacy filter 미구현 (v0.7.1+ follow-up)
- **Cross-project archive**: vault 의 raw mirror 가 L1 source-of-truth — L1 변경 시 vault 가 자동 propagate

## §9 Follow-up (v0.7.1+)  {#s9-followup}

- `stale-90day-lint` 의 자동 rotate (현재 *수동 lint*)
- v0.7.1: privacy filter (PII 자동 redact)
- v0.7.1: audit log dashboard (Grafana / Obsidian Canvas)
- v0.8.0: cross-project audit log aggregator (multi-project 통합 view)

## §10 References  {#s10-references}

- 1차 출처: AIDLC `awslabs/aidlc-workflows/aidlc-rules/aws-aidlc-rule-details/construction/build-and-test.md` (audit log 정책)
- 우리 SSOT: `workflow-source/core/audit_log_standard.md` (200 line)
- 우리 runtime: `workflow_kit/common/contracts/stage_gate.py` (`append_audit_log` + `normalize_timestamp`)
- 우리 검증: `workflow-source/tests/check_audit_log_compliance.py` (13 test PASS)
- 우리 wiki: [[topics/aidlc-benchmark-analysis-2026-06-12]] §4.2 H
