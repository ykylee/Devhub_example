DELETE FROM public.unit_appointments
WHERE user_id IN ('u1', 'u2', 'u3');

DELETE FROM public.users
WHERE user_id IN ('u1', 'u2', 'u3');

DELETE FROM public.scm_providers
WHERE provider_key IN ('bitbucket', 'gitea', 'forgejo', 'github');

DELETE FROM public.org_units
WHERE unit_id IN (
    'org-root',
    'dept-eng',
    'dept-prod',
    'team-infra',
    'team-frontend',
    'team-ux',
    'part-security'
);
