-- 000004: rollback project repository contribution weight
ALTER TABLE public.project_repositories DROP CONSTRAINT IF EXISTS project_repositories_weight_check;
ALTER TABLE public.project_repositories DROP COLUMN IF EXISTS contribution_weight;
