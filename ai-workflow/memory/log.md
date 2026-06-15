# ai-workflow/memory/log.md — v0.7.17+ in-repo wiki event log

본 file 은 standard_ai_workflow v0.7.17 의 wiki in-repo redirect 의 *event log target* — `tools/emit_wiki_l2_body.py --apply` 가 본 dir 에 wiki event (commit / push / PR / merge / release) 를 append.

본 저장소 (Devhub_example_minimax) 의 wiki 운영 = in-repo only:
- L1 raw mirror: `ai-workflow/memory/active/` + `ai-workflow/wiki/{concepts,decisions,entities,patterns,topics}/`
- L2 dense sources: `ai-workflow/wiki/sources/`
- Event log: 본 file (`ai-workflow/memory/log.md`)

외부 vault (`~/wiki/`) 연결 없음 (2026-06-15 v0.7.17 적용 결정).

위키 = 본 dir (in-repo) 단일 source. `vendor/standard_ai_workflow/tools/{refresh_wiki_memory,emit_wiki_l2_body,score_wiki_maintainability}.py` 가 본 dir 를 사용.
[2026-06-15T13:03:30Z] pr-update | PR #600 | f813493 | state=MERGED | files=0 | idem=pr-600-f813493a3dad8dbf4162684fe6f6d154397e1217
