-- 000004: add project repository contribution weight
ALTER TABLE public.project_repositories
ADD COLUMN contribution_weight numeric(5,2) DEFAULT 100.00 NOT NULL;

ALTER TABLE public.project_repositories
ADD CONSTRAINT project_repositories_weight_check CHECK (contribution_weight >= 0.00 AND contribution_weight <= 100.00);
