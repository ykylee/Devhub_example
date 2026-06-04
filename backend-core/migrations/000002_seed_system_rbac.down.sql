DELETE FROM public.rbac_policies
WHERE role_id IN ('developer', 'team_manager', 'system_admin');
