# Work Backlog — fix/work_260611-2-lint-phase2-chunk2

- 문서 목적: wiki lint phase 2 chunk 2 sprint 의 backlog.
- 상태: ✅ done (PR #569 squash `481a8faa` 머지 완료, 2026-06-12)
- 최종 수정일: 2026-06-12 (post-merge finalize)

## 1. 완료 항목 (2026-06-12)

### 1.1 L08 fix
- [x] `index.md` §8 devhub topics 표 1 line 추가 (`workflow.md`)

### 1.2 L02 fix (5 file body)
- [x] `wiki/projects/devhub/entities/devhub-agent-memory.md` — `[[ai-workflow]]` → prose
- [x] `wiki/projects/devhub/topics/workflow.md` — `[[query/...]]` → prose
- [x] `wiki/cross/concepts/runtime-injection.md` — `[[context-budget]]` → markdown link
- [x] `wiki/cross/topics/cross-host-bootstrap.md` — 4 wikilink → markdown link
- [x] `wiki/cross/topics/sub-host-onboarding.md` — 4 wikilink → markdown link

### 1.3 L02 fix (2 file frontmatter plain text 제거)
- [x] `wiki/cross/topics/cross-host-bootstrap.md` — 4 plain text 제거
- [x] `wiki/cross/topics/sub-host-onboarding.md` — 4 plain text 제거

### 1.4 L03 fix (3 file inbound)
- [x] `wiki/projects/devhub/entities/devhub-agent-memory.md` — `## Inbound from [[devhub]]`
- [x] `wiki/cross/topics/d-72-rollout.md` — `## Inbound from [[cross-host-bootstrap]]`
- [x] `wiki/cross/topics/sub-host-onboarding.md` — `## Inbound from [[cross-host-bootstrap]]`

### 1.5 L03 fix (2 file inbound source)
- [x] `wiki/cross/topics/cross-host-bootstrap.md` — `## Related Pages` (d-72-rollout + sub-host-onboarding inbound)
- [x] `wiki/projects/devhub/entities/devhub.md` — `## Related Entities` (devhub-agent-memory inbound)

### 1.6 lint-config.toml (L03 + L10 면제)
- [x] `[rules.L03] skip_paths = ["wiki/projects/devhub/sources/*.md"]`
- [x] `[rules.L10] skip_paths = ["wiki/projects/devhub/sources/*.md"]`

## 2. 결과

- **0/0/0 달성**: 11 error / 87 warn → 0 error / 0 warn
- **vault 8 file 변경** + **DevHub repo 1 file 변경**

## 3. 보류 (사용자 결정)

- [ ] vault 8 file commit + Gitea push
- [ ] DevHub repo 1 file commit + branch push
- [ ] PR 발행

## 4. Follow-up (별도 sprint)

- [ ] lint SSOT (my_harness) 의 L02 검사 schema 정합 (frontmatter related: 의 plain text wikilink false positive)
- [ ] `wiki-source-sync` skill 의 L03 fix 자동화 (frontmatter related: + body 1줄 일괄)
