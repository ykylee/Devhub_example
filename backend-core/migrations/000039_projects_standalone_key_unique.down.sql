-- 000039 rollback: standalone project key partial unique index 제거.

DROP INDEX IF EXISTS projects_standalone_key_uniq;
