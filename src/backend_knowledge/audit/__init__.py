"""Audit log core (umbrella doc §3.6.6 + §11.3 정합).

7 audit event type (§3.6.6.1):
1. audit.user.login       — X-DevHub-User-Context header 검증
2. audit.concept.access   — GET /concepts/{type}/{name}
3. audit.curation.edit    — PUT /concepts/{id} + POST /enrich
4. audit.query            — POST /query
5. audit.concept.archive  — POST /concepts/{id}/archive (§3.9.4)
6. audit.concept.publish  — POST /concepts/{id}/publish (§3.9.4)
7. audit.config.change    — operator 의 config / source_meta 변경

Storage: var/audit/audit-YYYY-MM-DD.jsonl (JSON Lines, daily rotation)
Retention: 7일 (configurable via AUDIT_LOG_RETENTION_DAYS)
"""
