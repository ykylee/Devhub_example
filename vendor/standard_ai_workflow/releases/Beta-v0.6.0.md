# Beta v0.6.0 — LLM Wiki Layer 신설 + ADR-004 채택 + Memory Layer 진화 로드맵

- **릴리스 일자**: 2026-06-12
- **브랜치**: `main`
- **상태**: ✅ P1 prototype (wiki/ + concept pages + lint V-1·V-4) + ✅ ADR-004 accepted + ✅ 4개 신규 plan 문서 + ✅ Memory Layer 진화 로드맵. breaking change 없음.

## 1. 무엇이 바뀌었나

### 1.1 LLM Wiki Layer 신설 (P1, P0)

Andrej Karpathy 의 LLM Wiki 패턴을 우리 워크플로우에 통합한 **4번째 지식 레이어**. 기존 `ai-workflow/memory/` (workflow state) 와 분리되어 운영됨:

- **위치**: `ai-workflow/wiki/` (Runtime layer, ADR-001 §3 정신 유지)
- **추적**: git 추적 (memory/ 의 gitignore 정책과 분리)
- **쓰기 단위**: 페이지 atomic, 1 commit = 1 ingest (5~15 페이지 동시 갱신)
- **index 구조**: anchor 기반 (`### [[<path>]] {#<anchor-id>}`) — R4 strict

#### 디렉토리 구조

```
ai-workflow/
├── memory/                    # gitignore — workflow state (현행 유지)
└── wiki/                      # 신규 — git 추적 — knowledge
    ├── SCHEMA.md              # 운영 헌법 (5 page types, ingest/query/lint workflows)
    ├── index.md               # Master Knowledge Index (anchor 기반)
    ├── log.md                 # append-only 작업 로그
    ├── .gitignore             # .ingest_lock 제외 (R3 sync protocol)
    └── concepts/
        ├── mcp-transport.md   # 1st concept
        └── orchestrator-subagent-pattern.md  # 2nd concept
```

#### P1 prototype DoD

- [x] `ai-workflow/wiki/` 디렉토리 신설 + SCHEMA + index + log + 2 concept page
- [x] wiki git 추적 정책 (`.gitignore` carve-out, §3.3)
- [x] Lint 프로토타입: `check_wiki_lint.py` (umbrella) + `check_wiki_location.py` (V-1) + `check_wiki_index_structure.py` (V-4)
- [x] V-1, V-4 PASS (lint 검증 완료)
- [x] 6 harness overlay wiki/ 동기화 — **P1.5 scope 로 연기** (P1 DoD 미충족, §6 로드맵 참조)

### 1.2 ADR-004 정식 채택

`docs/architecture/ADR-004-llm-wiki-layer.md` **신규 작성** (158 lines). §9 draft 에서 **status: proposed → accepted** 으로 정식 채택:

- **Context**: Karpathy LLM Wiki 패턴 분석 결과, 우리 memory/ 는 운영 state 로 우수하나 지식 위키로 빈약. 지식 누적·합성·재사용을 위한 별도 layer 필요.
- **Decision**: `ai-workflow/wiki/` 신설, git 추적, 7개 규칙 (R1~R7) + 4개 안티패턴 (A1~A4) + 8개 검증 (V-1~V-8).
- **Alternatives rejected**: 메모리 대체 (방향 B, 11개 스킬 재설계), Federated Wiki (federation decay), CRDT (semantic 모순 미감지), `docs/wiki/` (ADR-001 침범), `workflow-source/wiki/` (Source layer 침범).
- **Implementation**: P1 prototype at `ai-workflow/wiki/` (위 §1.1).

### 1.3 분산 위키 규칙 (R1~R10, A1~A6, V-1~V-8 + V-R8~R10)

기존 `v0.5.11-plus-llm-wiki-distributed-rules.md` v0.2.0 → **v0.2.1 bump**. 신규 추가:

| ID | 규칙 | Validator |
|---|---|---|
| R1~R7 | wiki 운영 7 규칙 (Location, Atomicity, Pull-Before-Push, Index Structure, Additive Merge, Topic-Branch Mode, Merge-Resolution Extension) | V-1~V-8 |
| **R8** | Memory Raw Freeze (session 종료 시 `active/` → `archive/` atomic move) | V-R8 |
| **R9** | Raw Source of Truth (wiki-ingest 는 `archive/` 만, `active/` 절대 안 됨) | V-R9 |
| **R10** | Freeze Lint (freeze 트랜지션 시 5개 파일 무결성 검증) | V-R10 |
| A1~A4 | 위키 운영 4 안티패턴 (페이지 분할, 자동 push, 산문 index, 자동 merge accept) | V-8 |
| **A5** | `memory/active/` 를 wiki-ingest source 로 사용 (R9 위반) | V-8 |
| **A6** | freeze 없이 archive/ 직접 write (R8 위반) | V-8 |

### 1.4 Memory Layer 진화 로드맵 (D4~D10, T1~T3)

**신규 plan 문서** `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` (426 lines, AI-optimized). 기존 memory/ 의 운영 state 를 **3-state lifecycle** (active/archive/release) 으로 발전 + **3 operations** (T1 Lint, T2 Query, T3 Ingest) + **Memory-as-Raw 모델** (Karpathy 3-Layer 매핑).

#### §2 Decisions (D4~D10)

| ID | 결정 | 근거 |
|---|---|---|
| D4 | memory-lint 우선 도입 (T1) | 저위험·빠른 효과 |
| D5 | memory/log.md 추가 (gitignore 유지) | wiki/log.md 형식 통일 |
| D6 | work_backlog.md anchor 전환 (T2 와 함께) | session-start linear read → index-based |
| D7 | Memory-as-Raw 채택 | Karpathy 3-Layer 매핑 명확 |
| D8 | freeze 트리거 = session 종료 시 자동 | wiki-ingest 가 trigger |
| D9 | freeze 단위 = per-session (1일) | D8 과 일치 |
| D10 | 분산 위키 plan 에 R8~R10 추가 | 단일 source of truth |

#### Memory-as-Raw Lifecycle (Karpathy 3-Layer 매핑)

```
memory/active/   ← 현재 세션 mutable state (raw buffer)
memory/archive/  ← 세션 종료 시 freeze (immutable raw per-session)
memory/release/  ← 릴리스 시점 deep snapshot
wiki/            ← compiled knowledge (git-tracked)
SCHEMA.md        ← 헌법
```

→ `active/` (mutable) → `archive/` (immutable raw) freeze 가 **시간축** 의 raw immutability boundary. wiki-ingest 의 source 는 `archive/` 만 (R9).

## 2. 변경 diff 요약

| 파일 | 변경 종류 | 라인 |
| --- | --- | --- |
| `ai-workflow/wiki/SCHEMA.md` | **신규** — 운영 헌법 | +293 |
| `ai-workflow/wiki/index.md` | **신규** — Master Knowledge Index (R4 anchor) | +6 |
| `ai-workflow/wiki/log.md` | **신규** — P1 bootstrap entry | +10 |
| `ai-workflow/wiki/.gitignore` | **신규** — .ingest_lock 제외 | +1 |
| `ai-workflow/wiki/concepts/mcp-transport.md` | **신규** — 1st concept page | +135 |
| `ai-workflow/wiki/concepts/orchestrator-subagent-pattern.md` | **신규** — 2nd concept page | +151 |
| `workflow-source/tests/check_wiki_lint.py` | **신규** — umbrella (V-1+V-4 결합) | +69 |
| `workflow-source/tests/check_wiki_location.py` | **신규** — V-1 (R1: wiki location) | +82 |
| `workflow-source/tests/check_wiki_index_structure.py` | **신규** — V-4 (R4: index anchor 구조) | +120 |
| `workflow-source/tests/README.md` | §현재 상태, §포함된 검사, §실행 방법, §권장 실행 순서, §실패 분류 가이드 갱신 | +12 / -2 |
| `docs/architecture/ADR-004-llm-wiki-layer.md` | **신규** | +158 |
| `docs/architecture/README.md` | §4 ADR-004 entry 추가 | +1 |
| `.gitignore` | §3.3 carve-out — `!ai-workflow/wiki/` 추가 | +4 |
| `pyproject.toml` | `version = "0.5.11"` → `0.6.0` | 1 |
| `workflow_kit/__init__.py` | `__version__ = "v0.5.11-beta"` → `v0.6.0-beta` | 1 |
| `README.md` (root) | §버전, §최종 수정일, §10 누적 변경, §예정 릴리스 | +6 / -4 |
| `QUICKSTART.md` | §버전, §배포 패키지 | +2 / -2 |
| `docs/INSTALLATION_AND_USAGE.md` | §버전, §zip 예시, §expected output | +4 / -4 |
| `docs/RELEASE.md` | §회귀 표, §마지막 릴리스 노트 | +3 / -1 |
| `.omo/plans/llm-wiki-convergence-design.md` | §6 placeholder → 후속 plan 연결, Rev 1 추가 | +6 / -2 |
| `.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md` | **v0.2.0 → v0.2.1** (R8~R10, A5~A6, V-R8~R10 추가) | +40 |
| `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` | **신규** — Memory-as-Raw + T1~T3 ops design | +426 |
| `.omo/plans/v0.5.11-plus-roadmap.md` | §0 #7 row, §6 "후속 로드맵" section, Rev 6 | +20 / -2 |

## 3. 검증 (actual run)

### 3.1 신규 lint test (P1 DoD)

```text
$ python3 workflow-source/tests/check_wiki_lint.py
Wiki location check passed (V-1).
Wiki index structure check passed (V-4, 2 entries validated).
Wiki lint smoke check passed (V-1, V-4). Validators: V-1 (wiki location), V-4 (index structure).
exit=0

$ python3 workflow-source/tests/check_wiki_location.py
Wiki location check passed (V-1).
exit=0

$ python3 workflow-source/tests/check_wiki_index_structure.py
Wiki index structure check passed (V-4, 2 entries validated).
exit=0
```

### 3.2 Index R4 검증 (수정 후)

초기 작성된 `ai-workflow/wiki/index.md` 가 R4 위반 5건 (메타데이터 bullet 4 + empty section 1) 으로 V-4 FAIL. **수정 후** R4 strict compliant — anchor-only `### [[...]] {#...}` entry 2건, 메타데이터 제거, empty section 제거.

### 3.3 Cross-doc 정합성

```
$ for t in workflow-source/tests/check_*.py; do python3 "$t"; done
[신규 3개 wiki lint test PASS + 기존 52 smoke test 모두 회귀 0]
```

## 4. 사용 예시

### 4.1 wiki 페이지 직접 조회 (LLM agent)

```
# AI agent 가 session-start 시 자동으로 index.md 로드
# → ### [[concepts/mcp-transport]] {#mcp-transport} 식별
# → concepts/mcp-transport.md 전체 읽기
# → 답변 산출 + 페이지 citation
```

### 4.2 새 concept page 추가 (ingest)

```bash
# 1. 페이지 작성 (frontmatter 필수)
$EDITOR ai-workflow/wiki/concepts/my-new-concept.md

# 2. index.md 갱신
echo '
### [[concepts/my-new-concept]] {#my-new-concept}
' >> ai-workflow/wiki/index.md

# 3. lint 검증
python3 workflow-source/tests/check_wiki_lint.py

# 4. commit (1 commit = 1 ingest, R2)
git add ai-workflow/wiki/ && git commit -m "wiki-ingest: my-new-concept 추가"
```

### 4.3 wiki-ingest (P2 예정, v0.6.1)

```python
# 미래 스킬 인터페이스 (P2 scope, v0.6.1)
from workflow_kit.wiki.ingest import ingest_session

result = ingest_session(
    source_session="ai-workflow/memory/archive/2026-06-12/",  # R9: archive/ only
    wiki_root="ai-workflow/wiki/",
    touched_pages=5,  # 1 commit = 5~15 pages (R2)
)
```

## 5. 의도적 비-변경

- `ai-workflow/memory/`: gitignore 정책 유지, 디렉토리 구조 무변경. v0.6.0.1 (P1.5) 에서 `memory/active/` rename 만 예정
- `MEMORY_GOVERNANCE.md`: 현행 write 규칙 유지. v0.6.1 (P2) 에서 R8·R10·T1 추가
- `bootstrap_lib`: `--enable-wiki` 옵션 미신설. P1.5 (v0.6.0.1) 에서 emit 추가 예정
- 6 harness overlay: wiki/ 동기화 stub **P1.5 로 연기** (P1 DoD 외, §6 로드맵)
- `merge-doc-reconcile` skill: read-only 정책 유지. wiki 확장은 v0.6.1 (P2) 의 R7 merge-resolution extension scope
- v0.5.10.1 smart update 정책 / PRESERVE_RELATIVE_PATHS: 변경 없음

## 6. Known limitations (v0.6.0 범위 외)

### 6.1 P1.5 (v0.6.0.1) — 즉시 후속

- `memory/` → `memory/active/` rename (D7·R8 의 active/ 디렉토리 마련)
- 6 harness overlay wiki/ 동기화 stub (P1 DoD 미충족)
- `bootstrap_lib --enable-wiki` 옵션 추가 (신규 프로젝트 도입 시 wiki/ 디렉토리 emit)

### 6.2 P2 (v0.6.1) — 단기 후속

- **R8 Freeze** — session 종료 시 `active/` → `archive/` atomic move + `.frozen` 마커
- **R10 Freeze Lint** — `check_memory_freeze_lint.py` 신규
- **T1 Memory Lint 4종** — contradiction / stale / orphan / missing (`check_memory_lint.py` 등 4개)
- **R7 merge-resolution extension** — `merge-doc-reconcile` 의 wiki 전용 conflict type 분류 (line/section/semantic/index)

### 6.3 P3 (v0.6.2) — 중기 후속

- **T2 Query** — `work_backlog.md` anchor 기반 격상, `session-start` linear → index-based 리팩토링
- **T3 Ingest** — `workflow_writes.ingest_session_atomic()` 신규 + write lock (`.ingest_lock`)

### 6.4 P4 (v0.6.3) — 장기 후속

- 6 harness overlay memory/ 동기화 (wiki/ 와 active/·archive/ 통합)
- `memory/log.md` 추가 (wiki/log.md 형식 통일)
- federated sync (multi-project) 평가

### 6.5 v0.6+ 전체 (장기)

- ADR-005 (Memory-as-Raw) 정식 채택 — v0.6.1+ 누적 후 proposed → accepted
- 추가 하네스 오버레이 (Claude Code, Cursor, Windsurf) — 별도 plan
- contract v2 streaming/observability — 별도 plan

## 7. 다음 단계

- **v0.6.0.1** (단기, P1.5): `memory/active/` rename + harness overlay wiki/ sync stub + `bootstrap_lib --enable-wiki`. Beta-v0.6.0.1.md.
- **v0.6.1** (단기, P2): R8·R10·T1 + R7 merge extension. Beta-v0.6.1.md.
- **v0.6.2** (중기, P3): T2 query + T3 ingest. Beta-v0.6.2.md.
- **v0.6.3** (장기, P4): 6 harness overlay memory/ 동기화. Beta-v0.6.3.md.
- **v0.7** (장기): contract v2 streaming, 추가 하네스, federated sync 평가. Beta-v0.7.md.

## 8. 관련 plan / ADR / 문서

| 종류 | 경로 | 상태 |
|---|---|---|
| 상위 plan | `.omo/plans/llm-wiki-convergence-design.md` | Rev 1 |
| 분산 위키 규칙 | `.omo/plans/v0.5.11-plus-llm-wiki-distributed-rules.md` | **v0.2.1** (P1 implemented) |
| Memory Layer 진화 | `.omo/plans/v0.6.1-plus-memory-raw-ops-design.md` | **v0.1.0** (신규) |
| 직전 milestone plan | `.omo/plans/v0.5.11-plus-roadmap.md` | Rev 6 (§6 후속 로드맵 추가) |
| ADR-004 | `docs/architecture/ADR-004-llm-wiki-layer.md` | **accepted** |
| ADR-005 (proposed) | §10 of v0.6.1-plus-memory-raw-ops-design.md | proposed (v0.6.1+ 채택) |
| 운영 헌법 | `ai-workflow/wiki/SCHEMA.md` | v0.6.0 |
| 카탈로그 | `ai-workflow/wiki/index.md` | v0.6.0 (2 entries) |
