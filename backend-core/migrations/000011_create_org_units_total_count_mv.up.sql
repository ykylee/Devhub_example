-- 000011: org_units_total_count Materialized View (ADR-0009, RM-M3-03).
--
-- Sub-tree member count caching for the organization hierarchy. The wire
-- `orgUnitResponse.total_count` is currently computed by a recursive CTE
-- every call — at 100+ unit sizes this adds 50-200ms to GET /api/v1/organization/hierarchy.
-- This MV captures `(unit_id, total_count)` and is refreshed concurrently
-- on a cron cadence (operations sprint owns the schedule; the documented
-- default is every 5 minutes).
--
-- See docs/adr/0009-org-secondary-memberships-and-total-count-mv.md §4.2
-- for the design rationale and the carve-out items (cron, getHierarchy
-- MV join, leader uniqueness store invariant).

CREATE MATERIALIZED VIEW org_units_total_count AS
WITH RECURSIVE descendants(root_unit_id, descendant_unit_id) AS (
    SELECT unit_id, unit_id FROM org_units
    UNION ALL
    SELECT d.root_unit_id, child.unit_id
    FROM descendants d
    JOIN org_units child ON child.parent_unit_id = d.descendant_unit_id
)
SELECT u.unit_id AS unit_id,
       COUNT(DISTINCT ua.user_id) AS total_count
FROM org_units u
LEFT JOIN descendants d ON d.root_unit_id = u.unit_id
LEFT JOIN unit_appointments ua ON ua.unit_id = d.descendant_unit_id
GROUP BY u.unit_id
WITH NO DATA;

-- REFRESH MATERIALIZED VIEW CONCURRENTLY 는 unique index 가 필요.
CREATE UNIQUE INDEX org_units_total_count_pk ON org_units_total_count (unit_id);

-- 1차 populate. 이후 갱신은 운영 cron 의 책임:
--   REFRESH MATERIALIZED VIEW CONCURRENTLY org_units_total_count;
REFRESH MATERIALIZED VIEW org_units_total_count;
