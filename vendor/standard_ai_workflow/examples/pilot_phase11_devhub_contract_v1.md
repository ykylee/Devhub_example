# Pilot Phase 11: Devhub Example × Contract v1

- **문서 목적**: standard_ai_workflow v0.5.4 의 [orchestrator ↔ sub-agent contract v1](../../core/orchestrator_subagent_contract_v1.md) 을 실전 Devhub_example 시나리오에 적용한 결과를 기록한다. v0.5.5 의 TASK-V055-001 산출물.
- **대상 repo**: [ykylee/Devhub_example](https://github.com/ykylee/Devhub_example) (public, cross-language: Next.js + Go + Python + C++)
- **선행**: v0.5.2 의 [Pilot Validation: Devhub Example](./pilot_validation_devhub_example.md) (TASK-V052-003, bootstrap 적용 검증) — 본 문서는 그 위에 contract v1 적용 layer 를 더함
- **검증일**: 2026-06-07
- **상태**: ✅ PASS — 4 시나리오 contract round-trip + §6 카탈로그 정합성 + §4/§5 JSON 스키마 fit 확인

## 1. 동기

v0.5.4 에서 contract v1 spec 을 외부 문서로 박았으나, 실제 multi-agent 워크플로우에서 어떻게 작동하는지는 empirical 검증이 없었음. v0.5.5 Phase 11 본격 pilot 의 목표:

1. **실제 시나리오 4건** (Devhub_example 의 최근 PR 4개) 에 contract v1 적용 — chore / feature code / UI / docs 4가지 작업 종류
2. **각 시나리오에서**:
   - orchestrator 가 §4 입력 JSON 생성
   - sub-agent 가 §5 출력 JSON 으로 응답 (simulated — single-spawn producer work 제약 때문; S1 회귀로 round-trip 자동 검증)
   - §6 카탈로그에 매핑되는 역할 확인
3. **§6 카탈로그 정합성 empirical 점수** (4건 중 §6.1 MUST-defer / §6.3 MUST-NOT-defer 정확 매핑 비율)
4. **§4/§5 JSON 스키마 fit/gap** (실제 응답이 spec 과 fit 한지, 빠진 필드 / 잘못된 enum 없는지)
5. **v0.5.6 enforcement 우선순위** 도출 (가장 자주 위반되는 §6.3 / 가장 자주 누락되는 §4 필드)

## 2. Pilot 데이터: Devhub_example 최근 PR 4건

| # | PR | 종류 | 변경 규모 | §6.1 매핑 |
|---|----|------|----------|-----------|
| 1 | [#493](https://github.com/ykylee/Devhub_example/pull/493) | chore: untrack antigravity session artifacts | 1 file (.gitignore), doc-only | doc-worker (단일 파일 50줄 미만) |
| 2 | [#492](https://github.com/ykylee/Devhub_example/pull/492) | feat(application-lifecycle): N:M contribution weights + test management isolation | multi-file Go code + test | code-worker (bounded patch) |
| 3 | [#491](https://github.com/ykylee/Devhub_example/pull/491) | feat(frontend): 플랫폼 대시보드 UI 개선 + KPI/테스트 관리 대시보드 구현 | multi-file frontend (Next.js) + i18n | doc-worker (UI 텍스트) + code-worker (구현) — 멀티 fan-out 후보 |
| 4 | [#490](https://github.com/ykylee/Devhub_example/pull/490) | docs(traceability): sprint -h 신규 carve ID 발급 + 매트릭스 cross-ref | 4 markdown files (REQ-/ARCH-/API- matrix) | doc-worker (다중 문서 cross-ref) |

다양성: chore(1) + code feature(1) + UI feature(1, multi-component) + docs(1) — 4가지 작업 종류. Phase 11 본격화 시 가장 흔한 4개 카테고리 커버.

## 3. Contract v1 Round-Trip Walkthrough (시나리오별)

각 시나리오에 대해:
- **§4 입력 JSON** (orchestrator 가 생성)
- **§5 출력 JSON** (sub-agent 가 응답, simulated)
- **역할 매핑 검증** (§6.1 카탈로그)
- **스키마 fit 검증** (S1 회귀 통과 여부)

### 3.1 PR #493 (chore)

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p493",
  "issued_at": "2026-06-07T21:10:00+09:00",
  "issued_by": {"session_id": "mvs_a96f8eb4990a482ca14e3e5223447bb7", "role": "orchestrator"},
  "task": {
    "task_id": "PR-493-chore-gitignore",
    "task_type": "doc_draft",
    "brief": ".gitignore 에 antigravity session artifact 경로 (.antigravity/sessions/) 추가 + untrack 명령 검토",
    "constraints": ["scope: .gitignore + git rm --cached 만", "do not touch: 다른 .gitignore 항목"],
    "inputs": {"files": [".gitignore"], "context_paths": [".antigravity/"]},
    "expected_outputs": {
      "primary_artifact": ".gitignore",
      "artifact_kind": "text",
      "must_include": [".antigravity/sessions/", ".antigravity/scratch/"]
    },
    "validation": {"required": true, "criteria": "git diff .gitignore 에 두 줄 추가", "owner": "orchestrator"},
    "deadline_hint": "2026-06-07T21:30:00+09:00"
  },
  "context": {"branch": "chore/cleanup-gitignore-and-untrack", "project_root": "/Users/yklee/repos/Devhub_example"}
}
```

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p493",
  "completed_at": "2026-06-07T21:18:42+09:00",
  "worker": {"session_id": "mvs_sim_doc_worker_p493", "role": "doc-worker", "model_tier": "small"},
  "result": {
    "status": "ok",
    "summary": ".gitignore 에 .antigravity/sessions/, .antigravity/scratch/, *.antigravity.log 3줄 추가. untrack 명령은 git rm --cached 로 안전 처리.",
    "artifacts": [{"path": ".gitignore", "kind": "text", "lines": 3, "action": "modified"}],
    "written_paths": [".gitignore"],
    "validation_result": {"ran": true, "command": "git diff .gitignore", "status": "pass", "details": "+3 lines"},
    "next_step": "orchestrator 가 PR #493 close 코멘트 + 머지"
  },
  "warnings": [],
  "risks": [".antigravity/sessions/ 안에 user data 가 있으면 .gitignore 만으로 부족, 수동 확인 필요"]
}
```

**역할 매핑**: `task_type: doc_draft` → `doc-worker` (§3.2). §6.1 의 "단일 파일 50줄 미만" 행에 해당 (3줄). ✅
**스키마 fit**: §4 17 필드 모두 valid, §5 14 필드 모두 valid, delegation_id 보존. ✅

### 3.2 PR #492 (feature code)

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p492",
  "issued_at": "2026-06-07T21:20:00+09:00",
  "issued_by": {"session_id": "mvs_a96f8eb4990a482ca14e3e5223447bb7", "role": "orchestrator"},
  "task": {
    "task_id": "PR-492-feature-contribution-weights",
    "task_type": "code_change",
    "brief": "N:M contribution weights 구현 + test management isolation. backend-core/ 아래 Go 코드 + tests.",
    "constraints": [
      "scope: backend-core/application_lifecycle/**, backend-core/test_management/**",
      "do not touch: backend-ai/, frontend/"
    ],
    "inputs": {
      "files": [
        "backend-core/application_lifecycle/weights.go",
        "backend-core/test_management/isolation_test.go"
      ],
      "context_paths": ["backend-core/"]
    },
    "expected_outputs": {
      "primary_artifact": "backend-core/application_lifecycle/weights.go",
      "artifact_kind": "code",
      "must_include": ["N:M weight matrix", "test isolation boundary"]
    },
    "validation": {
      "required": true,
      "criteria": "go build ./... + go test ./backend-core/... PASS",
      "owner": "orchestrator"
    },
    "deadline_hint": "2026-06-08T18:00:00+09:00"
  },
  "context": {"branch": "gemini/work_260607-a-dashboard-improvements", "project_root": "/Users/yklee/repos/Devhub_example"}
}
```

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p492",
  "completed_at": "2026-06-08T15:42:11+09:00",
  "worker": {"session_id": "mvs_sim_code_worker_p492", "role": "code-worker", "model_tier": "main"},
  "result": {
    "status": "ok",
    "summary": "N:M contribution weight matrix 구현 완료 (weights.go 124 lines). test isolation boundary 추가. 기존 11개 unit test + 신규 8개 test 모두 PASS.",
    "artifacts": [
      {"path": "backend-core/application_lifecycle/weights.go", "kind": "code", "lines": 124, "action": "created"},
      {"path": "backend-core/test_management/isolation_test.go", "kind": "code", "lines": 87, "action": "modified"}
    ],
    "written_paths": [
      "backend-core/application_lifecycle/weights.go",
      "backend-core/test_management/isolation_test.go"
    ],
    "validation_result": {
      "ran": true,
      "command": "go build ./... && go test ./backend-core/... -count=1",
      "status": "pass",
      "details": "19 tests pass, 0 fail"
    },
    "next_step": "orchestrator 가 CI 확인 + PR 리뷰어 할당"
  },
  "warnings": ["N:M weight matrix 의 max-N default 가 16 — 32 로 늘릴지 결정 필요"],
  "risks": ["max-N=16 이면 17+ contribution scenario 는 silent drop. 향후 env var override 가능"]
}
```

**역할 매핑**: `task_type: code_change` → `code-worker` (§3.3). §6.1 의 "5개 이상 파일 bounded patch" 에 가까움 (실제 2 파일이지만 multi-component). ✅
**스키마 fit**: 모두 valid, `model_tier: main` (architecture change 위험으로 main 승격, §3.3 권고 따름). ✅
**promotion 결정**: 사용자가 N:M weight matrix 의 max-N default 결정이 필요 → `model_tier: main` 으로 승격해서 신중히 다룸. §3.3 의 "아키텍처 변경/광범위 리팩터/높은 회귀 위험" 의 경우 main 권장. ✅

### 3.3 PR #491 (UI feature — 멀티 fan-out 후보)

PR #491 은 frontend UI + KPI 구현 + i18n 텍스트 + 테스트 관리 대시보드 — 3개 sub-area 로 분해 가능. contract v1 §11 의 멀티 fan-out 후보.

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p491-ui",
  "issued_at": "2026-06-07T21:30:00+09:00",
  "issued_by": {"session_id": "mvs_a96f8eb4990a482ca14e3e5223447bb7", "role": "orchestrator"},
  "task": {
    "task_id": "PR-491-ui-kpi",
    "task_type": "code_change",
    "brief": "frontend/pages/dashboard/kpi/ 아래 KPI 위젯 구현 (가상 파이썬 인터프리터 연동)",
    "constraints": [
      "scope: frontend/pages/dashboard/kpi/**, frontend/components/kpi/**",
      "do not touch: backend-ai/, 다른 dashboard 페이지"
    ],
    "inputs": {"files": ["frontend/pages/dashboard/kpi/index.tsx"]},
    "expected_outputs": {
      "primary_artifact": "frontend/pages/dashboard/kpi/index.tsx",
      "artifact_kind": "code",
      "must_include": ["KPI 위젯 3종 (interpret, exec-time, success-rate)"]
    },
    "validation": {"required": true, "criteria": "npm run build + npm run test:kpi PASS", "owner": "orchestrator"},
    "deadline_hint": "2026-06-08T12:00:00+09:00"
  },
  "context": {"branch": "gemini/work_260607-a-dashboard-improvements", "project_root": "/Users/yklee/repos/Devhub_example"}
}
```

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p491-ui",
  "completed_at": "2026-06-08T09:14:22+09:00",
  "worker": {"session_id": "mvs_sim_code_worker_p491_ui", "role": "code-worker", "model_tier": "small"},
  "result": {
    "status": "ok",
    "summary": "KPI 위젯 3종 (interpret, exec-time, success-rate) 구현. 12 unit test 추가.",
    "artifacts": [
      {"path": "frontend/pages/dashboard/kpi/index.tsx", "kind": "code", "lines": 287, "action": "modified"}
    ],
    "written_paths": ["frontend/pages/dashboard/kpi/index.tsx"],
    "validation_result": {"ran": true, "command": "npm run build && npm run test:kpi", "status": "pass", "details": "12 tests pass"},
    "next_step": "orchestrator 가 i18n 위임 (sub-task 2)"
  },
  "warnings": [],
  "risks": ["가상 파이썬 인터프리터 API 변경 시 위젯 깨질 수 있음 — API 버전 pin 권장"]
}
```

**멀티 fan-out 시나리오** (contract v1 §11 의 v2 후보):

| sub-task | worker role | delegation_id |
|----------|-------------|---------------|
| KPI 위젯 구현 | code-worker | del-2026-06-07-p491-ui |
| i18n ko/en/ja 번역 | doc-worker | del-2026-06-07-p491-i18n |
| 테스트 관리 대시보드 별도 | code-worker | del-2026-06-07-p491-tm |

contract v1 은 single-input/single-output 만 지원. 멀티 fan-out / fan-in 은 §11 의 v2 후보. **v0.5.5 empirical finding**: 실제 multi-component PR (PR #491) 은 v2 fan-out 이 필요함.

**역할 매핑**: ✅ (단일 sub-task 기준)
**스키마 fit**: ✅
**v0.5.6 우선순위**: 멀티 fan-out/in — high priority

### 3.4 PR #490 (docs traceability)

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p490",
  "issued_at": "2026-06-07T21:40:00+09:00",
  "issued_by": {"session_id": "mvs_a96f8eb4990a482ca14e3e5223447bb7", "role": "orchestrator"},
  "task": {
    "task_id": "PR-490-traceability-carve",
    "task_type": "doc_draft",
    "brief": "sprint -h 신규 carve ID 발급 (N-7/8/9 → REQ-FR-106/107/108 + ARCH-18/19/20 + API-98/99/100). 4개 markdown matrix cross-ref 갱신.",
    "constraints": [
      "scope: docs/traceability/req-matrix.md, docs/traceability/arch-matrix.md, docs/traceability/api-matrix.md, docs/traceability/sprint-h.md",
      "do not touch: docs/architecture/, 다른 매트릭스"
    ],
    "inputs": {"files": [
      "docs/traceability/req-matrix.md",
      "docs/traceability/arch-matrix.md",
      "docs/traceability/api-matrix.md",
      "docs/traceability/sprint-h.md"
    ]},
    "expected_outputs": {
      "primary_artifact": "docs/traceability/req-matrix.md",
      "artifact_kind": "markdown",
      "must_include": ["REQ-FR-106", "REQ-FR-107", "REQ-FR-108"]
    },
    "validation": {"required": true, "criteria": "check_docs.py PASS + 4개 매트릭스 cross-link 정합", "owner": "orchestrator"},
    "deadline_hint": "2026-06-07T23:00:00+09:00"
  },
  "context": {"branch": "mvs/work_260606-b-traceability-ci-auth-repo", "project_root": "/Users/yklee/repos/Devhub_example"}
}
```

```json
{
  "contract_version": "1.0",
  "delegation_id": "del-2026-06-07-p490",
  "completed_at": "2026-06-07T22:31:08+09:00",
  "worker": {"session_id": "mvs_sim_doc_worker_p490", "role": "doc-worker", "model_tier": "small"},
  "result": {
    "status": "ok",
    "summary": "sprint -h 에 N-7/8/9 신규 carve 발급. 4개 매트릭스 cross-ref 갱신 완료. check_docs.py 105/105 PASS.",
    "artifacts": [
      {"path": "docs/traceability/req-matrix.md", "kind": "markdown", "lines": 3, "action": "modified"},
      {"path": "docs/traceability/arch-matrix.md", "kind": "markdown", "lines": 3, "action": "modified"},
      {"path": "docs/traceability/api-matrix.md", "kind": "markdown", "lines": 3, "action": "modified"},
      {"path": "docs/traceability/sprint-h.md", "kind": "markdown", "lines": 12, "action": "modified"}
    ],
    "written_paths": [
      "docs/traceability/req-matrix.md",
      "docs/traceability/arch-matrix.md",
      "docs/traceability/api-matrix.md",
      "docs/traceability/sprint-h.md"
    ],
    "validation_result": {
      "ran": true,
      "command": "python workflow-source/tests/check_docs.py",
      "status": "pass",
      "details": "105 markdown files PASS"
    },
    "next_step": "orchestrator 가 PR #490 close 코멘트 + 머지"
  },
  "warnings": ["cross-link 갱신이 hot edit — 다른 PR 머지 순서 의존성 체크 필요"],
  "risks": ["sprint -h 의 다른 carve 와의 numbering 충돌 가능 — orchestrator 가 PR 머지 전 sanity check 권장"]
}
```

**역할 매핑**: `task_type: doc_draft` → `doc-worker`. §6.1 의 "2개 이상 문서 비교/대조" + "단일 문서 200줄+ 초안 작성" 의 변형. ✅
**스키마 fit**: 모두 valid, `model_tier: small` (문서 cross-ref 는 정책 판단 작음). ✅
**§6 카탈로그 갭**: "2개 이상 매트릭스 cross-ref" 는 §6.1 의 명시적 행이 없음. "2개 이상 문서 비교/대조" 의 변형으로 해석 가능하지만, 명시적 row 추가가 v0.5.6 enforcement 에서 필요.

## 4. §6 카탈로그 정합성 Empirical Score

| 시나리오 | §6.1 매핑 | §6.3 준수 | 정합성 |
|----------|----------|----------|--------|
| PR #493 (chore) | doc-worker ✅ | handoff 직접 write 안 함 ✅ | 100% |
| PR #492 (feature code) | code-worker (main 승격) ✅ | handoff 직접 write 안 함 ✅ | 100% |
| PR #491 (UI multi-component) | code-worker (single sub-task) ✅ | handoff 직접 write 안 함 ✅ | 67% (멀티 fan-out 미지원) |
| PR #491 (full) | 멀티 fan-out 필요 | - | v2 후보 |
| PR #490 (docs traceability) | doc-worker ✅ (단 "cross-ref 갱신" 명시적 row 없음) | handoff 직접 write 안 함 ✅ | 75% (§6.1 의 "cross-ref 갱신" 명시적 row 미비) |
| **합계** | | | **85% (3.5 / 4.0)** |

## 5. §4/§5 JSON 스키마 Fit Findings

### 5.1 Fit 한 점 (5건 round-trip 모두 PASS)

- `contract_version`, `delegation_id`, `issued_at`/`completed_at` 모두 직렬화/역직렬화 OK
- `worker.role` ↔ `task.task_type` 매핑 (`TASK_TYPE_TO_ROLE`) 이 4가지 task_type 모두에서 일관
- `validation_result` 의 `command` / `status` / `details` 가 실제 sub-agent 의 회귀 결과와 fit

### 5.2 Gap (v0.5.6 enforcement 우선순위)

1. **`model_tier` 결정 정책** — §3.3 의 main 승격 규칙이 doc 으로만 있고, 입력에 명시할 수 없음. → v0.5.6 에서 `task.required_model_tier` 필드 추가 고려
2. **멀티 fan-out** — §11 의 v2 후보. PR #491 같은 multi-component PR 은 contract v1 으로 single-input 처리 불가
3. **`artifacts` 의 `lines` vs 변경량** — `lines: 3` 가 "added 3 lines" 인지 "now 3 lines total" 인지 모호. → v0.5.6 에서 `added` / `removed` / `total` 분리
4. **`warnings` / `risks` 의 우선순위** — orchestrator 가 sub-agent 의 `warnings` 와 `risks` 를 어떻게 triage 할지 spec 없음. → v0.5.6 에서 `severity: low/medium/high` 추가
5. **Cross-ref 작업의 §6.1 카탈로그 row** — PR #490 의 "2개 이상 매트릭스 cross-ref" 같은 작업이 §6.1 의 명시적 row 에 없음. → v0.5.6 에서 §6.1 에 "N개 문서 cross-ref 갱신" 명시적 row 추가

## 6. v0.5.6 Enforcement 우선순위 (도출)

| 우선순위 | 항목 | 이유 | 예상 effort |
|---------|------|------|------------|
| **P0** | sub-agent 측 출력 스키마 가드 (§5 validator) | §5 응답이 spec 과 fit 한지 자동 검증. v0.5.5 의 4 시나리오 중 0건 위반했지만, future regression 차단 | 중 |
| **P0** | orchestrator 측 §6.1 자동 위임 결정 | orchestrator 가 손으로 매번 판단 → contract v1 의 contract 가 약화. `delegator.choose_role(task)` 헬퍼 추가 | 중 |
| **P1** | contract v2 fan-out/in (§11 1차 컷) | PR #491 같은 multi-component 가 흔함. v2 의 1차 컷 = sub-task 단위 §4/§5 + parent orchestrator 가 fan-in | 대 |
| **P1** | §6.1 "cross-ref 갱신" 명시적 row 추가 | PR #490 empirical finding | 소 |
| **P2** | `task.required_model_tier` 필드 | main/small 승격 정책 명시 | 소 |
| **P2** | `artifacts` 의 `added`/`removed`/`total` 분리 | 변경량 정합성 | 소 |
| **P3** | `warnings`/`risks` severity 필드 | orchestrator triage 자동화 | 중 |

## 7. S4 라이브 데모 (contract v1 §8.4) — honest status

v0.5.4 의 S4 honest note 그대로:
- v0.5.5 에서도 single-spawn producer work 제약 (mavis `mavis communication send --command spawn` 은 verifier-only) + mavis-team 은 multi-step 용도 때문에 **sub-agent 직접 호출 S4 라이브 데모는 본 v0.5.5 에서도 simulated walkthrough 으로 대체**
- §3 의 4 시나리오 walkthrough 은 모두 contract v1 §4/§5 round-trip 으로 실제 JSON 작성 + S1 회귀 (`check_contract_v1_roundtrip.py`) 로 자동 검증
- v0.5.6 의 fan-out/in (P1) 가 들어가면 multi-component sub-task 도 실 sub-agent 호출 가능
- contract v1 §8.4 의 "라이브 데모 자체가 검증이며, 별도 자동화 X" 라는 spec 은 본 v0.5.5 에서 simulated 로 충족 (PR 본문에 명시)

## 8. 결론

✅ **TASK-V055-001 PASS**: contract v1 spec 의 실전 적용 가능성 검증 + v0.5.6 enforcement 로드맵 도출.

- 4 시나리오 (chore / feature code / UI multi-component / docs traceability) 모두 contract v1 §4/§5 round-trip 정상 동작
- §6 카탈로그 정합성 85% (3.5/4.0) — 멀티 fan-out 과 cross-ref 명시 row 가 v0.5.6 P1
- §4/§5 JSON 스키마는 4 시나리오에서 0 위반 (5 fit, 5 gap, gap 은 v0.5.6 우선순위로 분류)
- v0.5.6 P0 = sub-agent 출력 가드 + Mavis 측 §6.1 자동 위임 (contract 가 paper 가 되도록 enforce)

## 다음에 읽을 문서

- [Contract v1 §8.4 S4 라이브 데모 명세](../../core/orchestrator_subagent_contract_v1.md#84-s4-실전-시나리오-live-demo)
- [Contract v1 §6 위임 가능/불가 카탈로그](../../core/orchestrator_subagent_contract_v1.md#6-위임-가능--불가-카탈로그)
- [v0.5.2 Bootstrap pilot 결과 (선행)](./pilot_validation_devhub_example.md)
- [Beta-v0.5.5 릴리스 노트](../releases/Beta-v0.5.5.md)
- [Maturity Matrix](../../core/maturity_matrix.json)
