INSERT INTO public.org_units (unit_id, parent_unit_id, unit_type, label, leader_user_id, position_x, position_y)
VALUES
    ('org-root', NULL, 'company', 'DevHub Global', 'u1', 400, 0),
    ('dept-eng', 'org-root', 'division', 'Engineering', 'u1', 200, 150),
    ('dept-prod', 'org-root', 'division', 'Product', 'u2', 600, 150),
    ('team-infra', 'dept-eng', 'team', 'Infrastructure', 'u1', 50, 300),
    ('team-frontend', 'dept-eng', 'team', 'Frontend', 'u3', 350, 300),
    ('team-ux', 'dept-prod', 'team', 'UX Strategy', 'u2', 600, 300),
    ('part-security', 'team-infra', 'part', 'Security Part', NULL, 50, 450);

INSERT INTO public.scm_providers (provider_key, display_name, enabled, adapter_version)
VALUES
    ('bitbucket', 'Bitbucket', TRUE, '0.0.1'),
    ('gitea', 'Gitea', TRUE, '0.0.1'),
    ('forgejo', 'Forgejo', TRUE, '0.0.1'),
    ('github', 'GitHub', TRUE, '0.0.1');

INSERT INTO public.users (
    user_id,
    email,
    display_name,
    role,
    status,
    primary_unit_id,
    current_unit_id,
    is_seconded,
    joined_at,
    user_type
)
VALUES
    ('u1', 'yklee@example.com', 'YK Lee', 'system_admin', 'active', 'dept-eng', 'dept-eng', FALSE, '2026-01-15', 'human'),
    ('u2', 'alex@example.com', 'Alex Kim', 'team_manager', 'active', 'dept-prod', 'team-ux', TRUE, '2026-02-01', 'human'),
    ('u3', 'sam@example.com', 'Sam Jones', 'developer', 'active', 'team-infra', 'team-infra', FALSE, '2026-05-01', 'human');

INSERT INTO public.unit_appointments (user_id, unit_id, appointment_role)
VALUES
    ('u1', 'org-root', 'leader'),
    ('u1', 'dept-eng', 'leader'),
    ('u2', 'dept-prod', 'leader'),
    ('u3', 'team-infra', 'member');
