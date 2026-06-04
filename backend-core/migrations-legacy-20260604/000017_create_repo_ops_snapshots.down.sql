DROP INDEX IF EXISTS quality_snapshots_repository_tool_idx;
DROP INDEX IF EXISTS quality_snapshots_repository_measured_at_idx;
DROP TABLE IF EXISTS quality_snapshots;

DROP INDEX IF EXISTS build_runs_repository_status_idx;
DROP INDEX IF EXISTS build_runs_repository_started_at_idx;
DROP TABLE IF EXISTS build_runs;

DROP INDEX IF EXISTS pr_activities_external_pr_idx;
DROP INDEX IF EXISTS pr_activities_repository_occurred_at_idx;
DROP TABLE IF EXISTS pr_activities;
