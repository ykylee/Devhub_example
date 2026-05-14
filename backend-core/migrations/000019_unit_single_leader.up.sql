-- 000019_unit_single_leader.up.sql
-- Enforce at most one leader appointment per unit at the SQL level. Pairs with
-- the application-level demote-then-promote logic in CreateOrgUnit/UpdateOrgUnit
-- so concurrent admin actions cannot leave a unit with two leaders.

CREATE UNIQUE INDEX IF NOT EXISTS unit_single_leader_idx
    ON unit_appointments (unit_id)
    WHERE appointment_role = 'leader';
