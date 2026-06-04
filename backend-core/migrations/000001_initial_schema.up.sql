
SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

CREATE SCHEMA hrdb;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

SET default_tablespace = '';

SET default_table_access_method = heap;

CREATE TABLE hrdb.persons (
    system_id text NOT NULL,
    employee_id text NOT NULL,
    name text NOT NULL,
    department_name text NOT NULL,
    email text,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.platform_repositories (
    platform_id uuid NOT NULL,
    repo_provider text NOT NULL,
    repo_full_name text NOT NULL,
    external_repo_id text,
    role text NOT NULL,
    sync_status text DEFAULT 'requested'::text NOT NULL,
    sync_error_code text,
    sync_error_retryable boolean,
    sync_error_at timestamp with time zone,
    last_sync_at timestamp with time zone,
    linked_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT platform_repositories_role_check CHECK ((role = ANY (ARRAY['primary'::text, 'sub'::text, 'shared'::text]))),
    CONSTRAINT platform_repositories_sync_error_code_check CHECK (((sync_error_code IS NULL) OR (sync_error_code = ANY (ARRAY['provider_unreachable'::text, 'auth_invalid'::text, 'permission_denied'::text, 'rate_limited'::text, 'webhook_signature_invalid'::text, 'payload_schema_mismatch'::text, 'resource_not_found'::text, 'internal_adapter_error'::text])))),
    CONSTRAINT platform_repositories_sync_error_consistency CHECK ((((sync_error_code IS NULL) AND (sync_error_retryable IS NULL) AND (sync_error_at IS NULL)) OR ((sync_error_code IS NOT NULL) AND (sync_error_retryable IS NOT NULL) AND (sync_error_at IS NOT NULL)))),
    CONSTRAINT platform_repositories_sync_status_check CHECK ((sync_status = ANY (ARRAY['requested'::text, 'verifying'::text, 'active'::text, 'degraded'::text, 'disconnected'::text])))
);

CREATE TABLE public.platforms (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    description text,
    status text NOT NULL,
    visibility text NOT NULL,
    owner_user_id text,
    start_date date,
    due_date date,
    archived_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    leader_user_id text,
    development_unit_id text,
    CONSTRAINT platforms_archived_consistency CHECK ((((status = 'archived'::text) AND (archived_at IS NOT NULL)) OR ((status <> 'archived'::text) AND (archived_at IS NULL)))),
    CONSTRAINT platforms_due_date_after_start CHECK (((start_date IS NULL) OR (due_date IS NULL) OR (due_date >= start_date))),
    CONSTRAINT platforms_key_format CHECK ((key ~ '^[A-Za-z0-9]{1,10}$'::text)),
    CONSTRAINT platforms_status_check CHECK ((status = ANY (ARRAY['planning'::text, 'active'::text, 'on_hold'::text, 'closed'::text, 'archived'::text]))),
    CONSTRAINT platforms_visibility_check CHECK ((visibility = ANY (ARRAY['public'::text, 'internal'::text, 'restricted'::text])))
);

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    audit_id text NOT NULL,
    actor_login text NOT NULL,
    action text NOT NULL,
    target_type text NOT NULL,
    target_id text NOT NULL,
    command_id text,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    source_ip text,
    request_id text,
    source_type text,
    source_event_id text
);

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;

CREATE TABLE public.build_runs (
    id bigint NOT NULL,
    repository_id bigint NOT NULL,
    run_external_id text NOT NULL,
    branch text NOT NULL,
    commit_sha text NOT NULL,
    status text NOT NULL,
    duration_seconds integer,
    started_at timestamp with time zone NOT NULL,
    finished_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT build_runs_finished_consistency CHECK (((finished_at IS NULL) OR (finished_at >= started_at))),
    CONSTRAINT build_runs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'success'::text, 'failed'::text, 'cancelled'::text, 'skipped'::text, 'unknown'::text])))
);

CREATE SEQUENCE public.build_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.build_runs_id_seq OWNED BY public.build_runs.id;

CREATE TABLE public.ci_runs (
    id bigint NOT NULL,
    external_id text NOT NULL,
    repository_id bigint,
    repository_name text NOT NULL,
    branch text,
    commit_sha text,
    status text NOT NULL,
    conclusion text,
    started_at timestamp with time zone,
    finished_at timestamp with time zone,
    duration_seconds integer,
    html_url text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ci_runs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'success'::text, 'failed'::text, 'cancelled'::text, 'skipped'::text, 'unknown'::text])))
);

CREATE SEQUENCE public.ci_runs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.ci_runs_id_seq OWNED BY public.ci_runs.id;

CREATE TABLE public.commands (
    id bigint NOT NULL,
    command_id text NOT NULL,
    command_type text NOT NULL,
    target_type text NOT NULL,
    target_id text NOT NULL,
    action_type text NOT NULL,
    status text NOT NULL,
    actor_login text NOT NULL,
    reason text NOT NULL,
    dry_run boolean DEFAULT true NOT NULL,
    requires_approval boolean DEFAULT false NOT NULL,
    idempotency_key text,
    request_payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    result_payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT commands_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'running'::text, 'succeeded'::text, 'failed'::text, 'rejected'::text, 'cancelled'::text]))),
    CONSTRAINT commands_target_check CHECK ((target_type = ANY (ARRAY['risk'::text, 'service'::text, 'report'::text]))),
    CONSTRAINT commands_type_check CHECK ((command_type = ANY (ARRAY['risk_mitigation'::text, 'service_action'::text, 'weekly_report'::text])))
);

CREATE SEQUENCE public.commands_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.commands_id_seq OWNED BY public.commands.id;

CREATE TABLE public.dev_request_intake_tokens (
    token_id uuid DEFAULT gen_random_uuid() NOT NULL,
    client_label text NOT NULL,
    hashed_token text NOT NULL,
    allowed_ips jsonb DEFAULT '[]'::jsonb NOT NULL,
    source_system text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by text NOT NULL,
    last_used_at timestamp with time zone,
    revoked_at timestamp with time zone,
    expires_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT dev_request_intake_tokens_allowed_ips_array CHECK ((jsonb_typeof(allowed_ips) = 'array'::text)),
    CONSTRAINT dev_request_intake_tokens_label_format CHECK ((client_label ~ '^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$'::text))
);

CREATE TABLE public.dev_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title text NOT NULL,
    details text DEFAULT ''::text NOT NULL,
    requester text NOT NULL,
    assignee_user_id text,
    source_system text NOT NULL,
    external_ref text,
    status text NOT NULL,
    registered_target_type text,
    registered_target_id text,
    rejected_reason text,
    received_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT dev_requests_assignee_required_when_active CHECK (((status = ANY (ARRAY['rejected'::text, 'closed'::text, 'received'::text])) OR (assignee_user_id IS NOT NULL))),
    CONSTRAINT dev_requests_registered_consistency CHECK (((status = 'registered'::text) = ((registered_target_type IS NOT NULL) AND (registered_target_id IS NOT NULL)))),
    CONSTRAINT dev_requests_rejected_reason_required CHECK (((status = 'rejected'::text) = ((rejected_reason IS NOT NULL) AND (rejected_reason <> ''::text)))),
    CONSTRAINT dev_requests_status_check CHECK ((status = ANY (ARRAY['received'::text, 'pending'::text, 'in_review'::text, 'registered'::text, 'rejected'::text, 'closed'::text]))),
    CONSTRAINT dev_requests_target_type_check CHECK (((registered_target_type IS NULL) OR (registered_target_type = ANY (ARRAY['platform'::text, 'project'::text]))))
);

CREATE TABLE public.event_cursors (
    cursor_key text NOT NULL,
    last_event_at timestamp with time zone NOT NULL,
    last_event_hash text,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.external_task_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider_id uuid NOT NULL,
    external_id text NOT NULL,
    title text NOT NULL,
    description text,
    raw_status text NOT NULL,
    normalized_status text,
    priority text,
    assignee text,
    reporter text,
    url text,
    labels text[],
    raw_payload jsonb,
    webhook_seq bigint,
    fetched_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.gitea_users (
    id bigint NOT NULL,
    gitea_user_id bigint,
    login text NOT NULL,
    display_name text,
    avatar_url text,
    html_url text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE public.gitea_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.gitea_users_id_seq OWNED BY public.gitea_users.id;

CREATE TABLE public.infra_service_snapshots (
    ingest_id text NOT NULL,
    agent_id text NOT NULL,
    snapshot_at timestamp with time zone NOT NULL,
    trace_id text,
    nodes_payload jsonb NOT NULL,
    services_payload jsonb NOT NULL,
    degraded_providers jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.integration_bindings (
    binding_id uuid DEFAULT gen_random_uuid() NOT NULL,
    scope_type text NOT NULL,
    scope_id text NOT NULL,
    provider_id uuid NOT NULL,
    external_key text NOT NULL,
    policy text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT integration_bindings_policy_check CHECK ((policy = ANY (ARRAY['summary_only'::text, 'execution_system'::text]))),
    CONSTRAINT integration_bindings_scope_type_check CHECK ((scope_type = ANY (ARRAY['platform'::text, 'project'::text])))
);

CREATE TABLE public.integration_providers (
    provider_id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider_key text NOT NULL,
    provider_type text NOT NULL,
    display_name text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    auth_mode text NOT NULL,
    credentials_ref text NOT NULL,
    capabilities jsonb DEFAULT '[]'::jsonb NOT NULL,
    sync_status text DEFAULT 'requested'::text NOT NULL,
    last_sync_at timestamp with time zone,
    last_error_code text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    base_url text,
    api_token text,
    auth_username text,
    auth_client_id text,
    auth_token_url text,
    auth_secret text,
    webhook_secret text,
    pull_interval_seconds integer DEFAULT 1800 NOT NULL,
    last_pulled_at timestamp with time zone,
    CONSTRAINT integration_providers_auth_mode_check CHECK ((auth_mode = ANY (ARRAY['token'::text, 'basic'::text, 'oauth2'::text, 'app_password'::text, 'agent'::text]))),
    CONSTRAINT integration_providers_sync_status_check CHECK ((sync_status = ANY (ARRAY['requested'::text, 'verifying'::text, 'active'::text, 'degraded'::text, 'disconnected'::text]))),
    CONSTRAINT integration_providers_type_check CHECK ((provider_type = ANY (ARRAY['alm'::text, 'scm'::text, 'ci_cd'::text, 'doc'::text, 'infra'::text])))
);

CREATE TABLE public.integration_sync_jobs (
    job_id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider_id uuid NOT NULL,
    requested_by text,
    status text DEFAULT 'queued'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT integration_sync_jobs_status_check CHECK ((status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text])))
);

CREATE TABLE public.issues (
    id bigint NOT NULL,
    gitea_issue_id bigint,
    repository_id bigint NOT NULL,
    number bigint NOT NULL,
    title text NOT NULL,
    state text NOT NULL,
    author_login text,
    assignee_login text,
    html_url text,
    opened_at timestamp with time zone,
    closed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT issues_state_check CHECK ((state = ANY (ARRAY['open'::text, 'closed'::text])))
);

CREATE SEQUENCE public.issues_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.issues_id_seq OWNED BY public.issues.id;

CREATE TABLE public.org_units (
    id bigint NOT NULL,
    unit_id text NOT NULL,
    parent_unit_id text,
    unit_type text NOT NULL,
    label text NOT NULL,
    leader_user_id text,
    position_x integer DEFAULT 0 NOT NULL,
    position_y integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT org_units_type_check CHECK ((unit_type = ANY (ARRAY['division'::text, 'team'::text, 'group'::text, 'part'::text, 'company'::text])))
);

CREATE SEQUENCE public.org_units_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.org_units_id_seq OWNED BY public.org_units.id;

CREATE TABLE public.unit_appointments (
    id bigint NOT NULL,
    user_id text NOT NULL,
    unit_id text NOT NULL,
    appointment_role text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT unit_appointments_role_check CHECK ((appointment_role = ANY (ARRAY['leader'::text, 'member'::text])))
);

CREATE MATERIALIZED VIEW public.org_units_total_count AS
 WITH RECURSIVE descendants(root_unit_id, descendant_unit_id) AS (
         SELECT org_units.unit_id,
            org_units.unit_id
           FROM public.org_units
        UNION ALL
         SELECT d_1.root_unit_id,
            child.unit_id
           FROM (descendants d_1
             JOIN public.org_units child ON ((child.parent_unit_id = d_1.descendant_unit_id)))
        )
 SELECT u.unit_id,
    count(DISTINCT ua.user_id) AS total_count
   FROM ((public.org_units u
     LEFT JOIN descendants d ON ((d.root_unit_id = u.unit_id)))
     LEFT JOIN public.unit_appointments ua ON ((ua.unit_id = d.descendant_unit_id)))
  GROUP BY u.unit_id
  WITH NO DATA;

CREATE TABLE public.pr_activities (
    id bigint NOT NULL,
    repository_id bigint NOT NULL,
    external_pr_id text NOT NULL,
    event_type text NOT NULL,
    actor_login text,
    occurred_at timestamp with time zone NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pr_activities_event_type_check CHECK ((event_type = ANY (ARRAY['opened'::text, 'reviewed'::text, 'commented'::text, 'closed'::text, 'merged'::text, 'reopened'::text, 'updated'::text])))
);

CREATE SEQUENCE public.pr_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.pr_activities_id_seq OWNED BY public.pr_activities.id;

CREATE TABLE public.project_integrations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    scope text NOT NULL,
    project_id uuid,
    platform_id uuid,
    integration_type text NOT NULL,
    external_key text NOT NULL,
    url text NOT NULL,
    policy text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_integrations_policy_check CHECK ((policy = ANY (ARRAY['summary_only'::text, 'execution_system'::text]))),
    CONSTRAINT project_integrations_scope_check CHECK ((scope = ANY (ARRAY['platform'::text, 'project'::text]))),
    CONSTRAINT project_integrations_scope_target_consistency CHECK ((((scope = 'platform'::text) AND (platform_id IS NOT NULL) AND (project_id IS NULL)) OR ((scope = 'project'::text) AND (project_id IS NOT NULL) AND (platform_id IS NULL)))),
    CONSTRAINT project_integrations_type_check CHECK ((integration_type = ANY (ARRAY['jira'::text, 'confluence'::text])))
);

CREATE TABLE public.project_members (
    project_id uuid NOT NULL,
    user_id text NOT NULL,
    project_role text NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_members_role_check CHECK ((project_role = ANY (ARRAY['lead'::text, 'contributor'::text, 'observer'::text])))
);

CREATE TABLE public.project_repositories (
    project_id uuid NOT NULL,
    repository_id bigint NOT NULL,
    role text DEFAULT 'linked'::text NOT NULL,
    linked_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT project_repositories_role_check CHECK ((role = ANY (ARRAY['primary'::text, 'linked'::text, 'shared'::text])))
);

CREATE TABLE public.projects (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    platform_id uuid,
    repository_id bigint,
    key text NOT NULL,
    name text NOT NULL,
    description text,
    status text NOT NULL,
    visibility text NOT NULL,
    owner_user_id text,
    start_date date,
    due_date date,
    archived_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT projects_archived_consistency CHECK ((((status = 'archived'::text) AND (archived_at IS NOT NULL)) OR ((status <> 'archived'::text) AND (archived_at IS NULL)))),
    CONSTRAINT projects_due_date_after_start CHECK (((start_date IS NULL) OR (due_date IS NULL) OR (due_date >= start_date))),
    CONSTRAINT projects_status_check CHECK ((status = ANY (ARRAY['planning'::text, 'active'::text, 'on_hold'::text, 'closed'::text, 'archived'::text]))),
    CONSTRAINT projects_visibility_check CHECK ((visibility = ANY (ARRAY['public'::text, 'internal'::text, 'restricted'::text])))
);

CREATE TABLE public.pull_requests (
    id bigint NOT NULL,
    gitea_pull_request_id bigint,
    repository_id bigint NOT NULL,
    number bigint NOT NULL,
    title text NOT NULL,
    state text NOT NULL,
    author_login text,
    head_branch text,
    base_branch text,
    head_sha text,
    html_url text,
    merged_at timestamp with time zone,
    closed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pull_requests_state_check CHECK ((state = ANY (ARRAY['open'::text, 'closed'::text, 'merged'::text])))
);

CREATE SEQUENCE public.pull_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.pull_requests_id_seq OWNED BY public.pull_requests.id;

CREATE TABLE public.quality_snapshots (
    id bigint NOT NULL,
    repository_id bigint NOT NULL,
    tool text NOT NULL,
    ref_name text NOT NULL,
    commit_sha text,
    score numeric(6,2),
    gate_passed boolean,
    metric_payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    measured_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT quality_snapshots_score_range CHECK (((score IS NULL) OR ((score >= (0)::numeric) AND (score <= (100)::numeric))))
);

CREATE SEQUENCE public.quality_snapshots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.quality_snapshots_id_seq OWNED BY public.quality_snapshots.id;

CREATE TABLE public.rbac_policies (
    role_id text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    is_system boolean DEFAULT false NOT NULL,
    permissions jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rbac_policies_audit_invariant CHECK (((COALESCE((((permissions -> 'audit'::text) ->> 'create'::text))::boolean, false) = false) AND (COALESCE((((permissions -> 'audit'::text) ->> 'edit'::text))::boolean, false) = false) AND (COALESCE((((permissions -> 'audit'::text) ->> 'delete'::text))::boolean, false) = false))),
    CONSTRAINT rbac_policies_role_id_format CHECK (((role_id = ANY (ARRAY['developer'::text, 'team_manager'::text, 'system_admin'::text])) OR (role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'::text)))
);

CREATE TABLE public.realtime_tickets (
    ticket text NOT NULL,
    actor_login text NOT NULL,
    actor_role text DEFAULT ''::text NOT NULL,
    source_type text DEFAULT 'oidc'::text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE public.repositories (
    id bigint NOT NULL,
    gitea_repository_id bigint,
    full_name text NOT NULL,
    owner_login text,
    name text NOT NULL,
    clone_url text,
    html_url text,
    default_branch text,
    private boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    source text,
    provider_id uuid,
    description text,
    repository_status text DEFAULT 'active'::text NOT NULL,
    publish_requested_at timestamp with time zone,
    published_at timestamp with time zone,
    CONSTRAINT repositories_status_check CHECK ((repository_status = ANY (ARRAY['draft'::text, 'active'::text])))
);

CREATE SEQUENCE public.repositories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.repositories_id_seq OWNED BY public.repositories.id;

CREATE TABLE public.risks (
    id bigint NOT NULL,
    risk_key text NOT NULL,
    title text NOT NULL,
    reason text NOT NULL,
    impact text NOT NULL,
    status text NOT NULL,
    owner_login text,
    source_type text NOT NULL,
    source_id text,
    suggested_actions jsonb DEFAULT '[]'::jsonb NOT NULL,
    detected_at timestamp with time zone DEFAULT now() NOT NULL,
    mitigated_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT risks_impact_check CHECK ((impact = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text, 'critical'::text]))),
    CONSTRAINT risks_status_check CHECK ((status = ANY (ARRAY['detected'::text, 'investigation'::text, 'action_required'::text, 'mitigated'::text, 'dismissed'::text])))
);

CREATE SEQUENCE public.risks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.risks_id_seq OWNED BY public.risks.id;

CREATE TABLE public.scm_providers (
    provider_key text NOT NULL,
    display_name text NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    adapter_version text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT scm_providers_key_format CHECK ((provider_key ~ '^[a-z][a-z0-9_-]{0,30}$'::text)),
    CONSTRAINT scm_providers_version_format CHECK ((adapter_version ~ '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$'::text))
);

CREATE SEQUENCE public.task_webhook_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE public.unit_appointments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.unit_appointments_id_seq OWNED BY public.unit_appointments.id;

CREATE TABLE public.users (
    id bigint NOT NULL,
    user_id text NOT NULL,
    email text NOT NULL,
    display_name text NOT NULL,
    role text NOT NULL,
    status text NOT NULL,
    primary_unit_id text,
    current_unit_id text,
    is_seconded boolean DEFAULT false NOT NULL,
    joined_at date NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    user_type character varying(20) DEFAULT 'human'::character varying NOT NULL,
    idp_subject text,
    onboarding_completed_at timestamp with time zone,
    review_status text,
    CONSTRAINT users_onboarding_review_consistency CHECK (((onboarding_completed_at IS NULL) = (review_status IS NULL))),
    CONSTRAINT users_review_status_check CHECK (((review_status IS NULL) OR (review_status = ANY (ARRAY['pending_review'::text, 'reviewed'::text])))),
    CONSTRAINT users_status_check CHECK ((status = ANY (ARRAY['active'::text, 'pending'::text, 'deactivated'::text])))
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE public.webhook_events (
    id bigint NOT NULL,
    event_type text NOT NULL,
    delivery_id text,
    dedupe_key text NOT NULL,
    repository_id bigint,
    repository_name text,
    sender_login text,
    payload jsonb NOT NULL,
    status text DEFAULT 'received'::text NOT NULL,
    error_message text,
    retry_count integer DEFAULT 0 NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    validated_at timestamp with time zone,
    processed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT webhook_events_status_check CHECK ((status = ANY (ARRAY['received'::text, 'validated'::text, 'processed'::text, 'failed'::text, 'ignored'::text])))
);

CREATE SEQUENCE public.webhook_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.webhook_events_id_seq OWNED BY public.webhook_events.id;

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);

ALTER TABLE ONLY public.build_runs ALTER COLUMN id SET DEFAULT nextval('public.build_runs_id_seq'::regclass);

ALTER TABLE ONLY public.ci_runs ALTER COLUMN id SET DEFAULT nextval('public.ci_runs_id_seq'::regclass);

ALTER TABLE ONLY public.commands ALTER COLUMN id SET DEFAULT nextval('public.commands_id_seq'::regclass);

ALTER TABLE ONLY public.gitea_users ALTER COLUMN id SET DEFAULT nextval('public.gitea_users_id_seq'::regclass);

ALTER TABLE ONLY public.issues ALTER COLUMN id SET DEFAULT nextval('public.issues_id_seq'::regclass);

ALTER TABLE ONLY public.org_units ALTER COLUMN id SET DEFAULT nextval('public.org_units_id_seq'::regclass);

ALTER TABLE ONLY public.pr_activities ALTER COLUMN id SET DEFAULT nextval('public.pr_activities_id_seq'::regclass);

ALTER TABLE ONLY public.pull_requests ALTER COLUMN id SET DEFAULT nextval('public.pull_requests_id_seq'::regclass);

ALTER TABLE ONLY public.quality_snapshots ALTER COLUMN id SET DEFAULT nextval('public.quality_snapshots_id_seq'::regclass);

ALTER TABLE ONLY public.repositories ALTER COLUMN id SET DEFAULT nextval('public.repositories_id_seq'::regclass);

ALTER TABLE ONLY public.risks ALTER COLUMN id SET DEFAULT nextval('public.risks_id_seq'::regclass);

ALTER TABLE ONLY public.unit_appointments ALTER COLUMN id SET DEFAULT nextval('public.unit_appointments_id_seq'::regclass);

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.webhook_events ALTER COLUMN id SET DEFAULT nextval('public.webhook_events_id_seq'::regclass);

ALTER TABLE ONLY hrdb.persons
    ADD CONSTRAINT persons_employee_id_key UNIQUE (employee_id);

ALTER TABLE ONLY hrdb.persons
    ADD CONSTRAINT persons_pkey PRIMARY KEY (system_id);

ALTER TABLE ONLY public.platform_repositories
    ADD CONSTRAINT platform_repositories_pkey PRIMARY KEY (platform_id, repo_provider, repo_full_name);

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_key_key UNIQUE (key);

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_audit_id_key UNIQUE (audit_id);

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.build_runs
    ADD CONSTRAINT build_runs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.build_runs
    ADD CONSTRAINT build_runs_run_external_id_key UNIQUE (run_external_id);

ALTER TABLE ONLY public.ci_runs
    ADD CONSTRAINT ci_runs_external_id_key UNIQUE (external_id);

ALTER TABLE ONLY public.ci_runs
    ADD CONSTRAINT ci_runs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.commands
    ADD CONSTRAINT commands_command_id_key UNIQUE (command_id);

ALTER TABLE ONLY public.commands
    ADD CONSTRAINT commands_idempotency_key_key UNIQUE (idempotency_key);

ALTER TABLE ONLY public.commands
    ADD CONSTRAINT commands_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.dev_request_intake_tokens
    ADD CONSTRAINT dev_request_intake_tokens_hashed_token_key UNIQUE (hashed_token);

ALTER TABLE ONLY public.dev_request_intake_tokens
    ADD CONSTRAINT dev_request_intake_tokens_pkey PRIMARY KEY (token_id);

ALTER TABLE ONLY public.dev_requests
    ADD CONSTRAINT dev_requests_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.event_cursors
    ADD CONSTRAINT event_cursors_pkey PRIMARY KEY (cursor_key);

ALTER TABLE ONLY public.external_task_items
    ADD CONSTRAINT external_task_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.external_task_items
    ADD CONSTRAINT external_task_items_provider_id_external_id_key UNIQUE (provider_id, external_id);

ALTER TABLE ONLY public.gitea_users
    ADD CONSTRAINT gitea_users_gitea_user_id_key UNIQUE (gitea_user_id);

ALTER TABLE ONLY public.gitea_users
    ADD CONSTRAINT gitea_users_login_key UNIQUE (login);

ALTER TABLE ONLY public.gitea_users
    ADD CONSTRAINT gitea_users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.infra_service_snapshots
    ADD CONSTRAINT infra_service_snapshots_pkey PRIMARY KEY (ingest_id);

ALTER TABLE ONLY public.integration_bindings
    ADD CONSTRAINT integration_bindings_pkey PRIMARY KEY (binding_id);

ALTER TABLE ONLY public.integration_bindings
    ADD CONSTRAINT integration_bindings_scope_type_scope_id_provider_id_extern_key UNIQUE (scope_type, scope_id, provider_id, external_key);

ALTER TABLE ONLY public.integration_providers
    ADD CONSTRAINT integration_providers_pkey PRIMARY KEY (provider_id);

ALTER TABLE ONLY public.integration_providers
    ADD CONSTRAINT integration_providers_provider_key_key UNIQUE (provider_key);

ALTER TABLE ONLY public.integration_sync_jobs
    ADD CONSTRAINT integration_sync_jobs_pkey PRIMARY KEY (job_id);

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_gitea_issue_id_key UNIQUE (gitea_issue_id);

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_repository_number_unique UNIQUE (repository_id, number);

ALTER TABLE ONLY public.org_units
    ADD CONSTRAINT org_units_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.org_units
    ADD CONSTRAINT org_units_unit_id_key UNIQUE (unit_id);

ALTER TABLE ONLY public.pr_activities
    ADD CONSTRAINT pr_activities_event_unique UNIQUE (repository_id, external_pr_id, event_type, occurred_at);

ALTER TABLE ONLY public.pr_activities
    ADD CONSTRAINT pr_activities_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_integrations
    ADD CONSTRAINT project_integrations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_pkey PRIMARY KEY (project_id, user_id);

ALTER TABLE ONLY public.project_repositories
    ADD CONSTRAINT project_repositories_pkey PRIMARY KEY (project_id, repository_id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pull_requests
    ADD CONSTRAINT pull_requests_gitea_pull_request_id_key UNIQUE (gitea_pull_request_id);

ALTER TABLE ONLY public.pull_requests
    ADD CONSTRAINT pull_requests_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.pull_requests
    ADD CONSTRAINT pull_requests_repository_number_unique UNIQUE (repository_id, number);

ALTER TABLE ONLY public.quality_snapshots
    ADD CONSTRAINT quality_snapshots_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rbac_policies
    ADD CONSTRAINT rbac_policies_pkey PRIMARY KEY (role_id);

ALTER TABLE ONLY public.realtime_tickets
    ADD CONSTRAINT realtime_tickets_pkey PRIMARY KEY (ticket);

ALTER TABLE ONLY public.repositories
    ADD CONSTRAINT repositories_full_name_key UNIQUE (full_name);

ALTER TABLE ONLY public.repositories
    ADD CONSTRAINT repositories_gitea_repository_id_key UNIQUE (gitea_repository_id);

ALTER TABLE ONLY public.repositories
    ADD CONSTRAINT repositories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.risks
    ADD CONSTRAINT risks_risk_key_key UNIQUE (risk_key);

ALTER TABLE ONLY public.scm_providers
    ADD CONSTRAINT scm_providers_pkey PRIMARY KEY (provider_key);

ALTER TABLE ONLY public.unit_appointments
    ADD CONSTRAINT unit_appointments_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.unit_appointments
    ADD CONSTRAINT unit_appointments_unique UNIQUE (user_id, unit_id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_user_id_key UNIQUE (user_id);

ALTER TABLE ONLY public.webhook_events
    ADD CONSTRAINT webhook_events_pkey PRIMARY KEY (id);

CREATE INDEX persons_employee_id_idx ON hrdb.persons USING btree (employee_id);

CREATE INDEX persons_name_lower_idx ON hrdb.persons USING btree (lower(name));

CREATE INDEX platform_repositories_external_repo_id_idx ON public.platform_repositories USING btree (repo_provider, external_repo_id) WHERE (external_repo_id IS NOT NULL);

CREATE INDEX platform_repositories_provider_repo_idx ON public.platform_repositories USING btree (repo_provider, repo_full_name);

CREATE INDEX platform_repositories_sync_status_idx ON public.platform_repositories USING btree (sync_status);

CREATE INDEX platforms_archived_at_idx ON public.platforms USING btree (archived_at) WHERE (archived_at IS NOT NULL);

CREATE INDEX platforms_dev_unit_idx ON public.platforms USING btree (development_unit_id);

CREATE INDEX platforms_leader_idx ON public.platforms USING btree (leader_user_id);

CREATE INDEX platforms_owner_idx ON public.platforms USING btree (owner_user_id);

CREATE INDEX platforms_status_idx ON public.platforms USING btree (status);

CREATE INDEX platforms_visibility_idx ON public.platforms USING btree (visibility);

CREATE INDEX audit_logs_command_id_idx ON public.audit_logs USING btree (command_id);

CREATE INDEX audit_logs_request_id_idx ON public.audit_logs USING btree (request_id) WHERE (request_id IS NOT NULL);

CREATE UNIQUE INDEX audit_logs_source_event_id_uniq ON public.audit_logs USING btree (source_type, source_event_id) WHERE ((source_event_id IS NOT NULL) AND (source_type IS NOT NULL));

CREATE INDEX audit_logs_source_type_idx ON public.audit_logs USING btree (source_type) WHERE (source_type IS NOT NULL);

CREATE INDEX audit_logs_target_created_at_idx ON public.audit_logs USING btree (target_type, target_id, created_at DESC);

CREATE INDEX build_runs_repository_started_at_idx ON public.build_runs USING btree (repository_id, started_at DESC);

CREATE INDEX build_runs_repository_status_idx ON public.build_runs USING btree (repository_id, status);

CREATE INDEX ci_runs_repository_status_updated_at_idx ON public.ci_runs USING btree (repository_name, status, updated_at DESC);

CREATE INDEX commands_idempotency_key_idx ON public.commands USING btree (idempotency_key) WHERE (idempotency_key IS NOT NULL);

CREATE INDEX commands_status_updated_at_idx ON public.commands USING btree (status, updated_at DESC);

CREATE INDEX commands_target_updated_at_idx ON public.commands USING btree (target_type, target_id, updated_at DESC);

CREATE INDEX dev_request_intake_tokens_active_idx ON public.dev_request_intake_tokens USING btree (revoked_at) WHERE (revoked_at IS NULL);

CREATE INDEX dev_requests_assignee_status_idx ON public.dev_requests USING btree (assignee_user_id, status);

CREATE UNIQUE INDEX dev_requests_idempotency_uniq ON public.dev_requests USING btree (source_system, external_ref) WHERE (external_ref IS NOT NULL);

CREATE INDEX dev_requests_source_system_idx ON public.dev_requests USING btree (source_system);

CREATE INDEX dev_requests_status_idx ON public.dev_requests USING btree (status);

CREATE INDEX event_cursors_updated_at_idx ON public.event_cursors USING btree (updated_at);

CREATE UNIQUE INDEX external_task_items_webhook_seq_uniq ON public.external_task_items USING btree (provider_id, webhook_seq) WHERE (webhook_seq IS NOT NULL);

CREATE INDEX idx_infra_service_snapshots_snapshot_at ON public.infra_service_snapshots USING btree (snapshot_at DESC, created_at DESC);

CREATE INDEX idx_repositories_provider_id ON public.repositories USING btree (provider_id);

CREATE INDEX integration_bindings_provider_idx ON public.integration_bindings USING btree (provider_id);

CREATE INDEX integration_bindings_scope_idx ON public.integration_bindings USING btree (scope_type, scope_id);

CREATE INDEX integration_providers_type_enabled_idx ON public.integration_providers USING btree (provider_type, enabled);

CREATE INDEX integration_sync_jobs_provider_created_idx ON public.integration_sync_jobs USING btree (provider_id, created_at DESC);

CREATE INDEX issues_repository_state_updated_at_idx ON public.issues USING btree (repository_id, state, updated_at DESC);

CREATE INDEX org_units_parent_idx ON public.org_units USING btree (parent_unit_id);

CREATE UNIQUE INDEX org_units_total_count_pk ON public.org_units_total_count USING btree (unit_id);

CREATE INDEX org_units_type_idx ON public.org_units USING btree (unit_type);

CREATE INDEX pr_activities_external_pr_idx ON public.pr_activities USING btree (repository_id, external_pr_id);

CREATE INDEX pr_activities_repository_occurred_at_idx ON public.pr_activities USING btree (repository_id, occurred_at DESC);

CREATE UNIQUE INDEX project_integrations_platform_unique ON public.project_integrations USING btree (platform_id, integration_type, external_key) WHERE (platform_id IS NOT NULL);

CREATE UNIQUE INDEX project_integrations_project_unique ON public.project_integrations USING btree (project_id, integration_type, external_key) WHERE (project_id IS NOT NULL);

CREATE INDEX project_members_user_idx ON public.project_members USING btree (user_id);

CREATE INDEX project_repositories_repository_idx ON public.project_repositories USING btree (repository_id);

CREATE INDEX projects_platform_idx ON public.projects USING btree (platform_id);

CREATE INDEX projects_owner_idx ON public.projects USING btree (owner_user_id);

CREATE INDEX projects_repository_idx ON public.projects USING btree (repository_id);

CREATE UNIQUE INDEX projects_repository_key_active_uniq ON public.projects USING btree (repository_id, key) WHERE ((repository_id IS NOT NULL) AND (archived_at IS NULL));

CREATE UNIQUE INDEX projects_standalone_key_active_uniq ON public.projects USING btree (key) WHERE ((repository_id IS NULL) AND (archived_at IS NULL));

CREATE INDEX projects_status_idx ON public.projects USING btree (status);

CREATE INDEX projects_visibility_idx ON public.projects USING btree (visibility);

CREATE INDEX pull_requests_repository_state_updated_at_idx ON public.pull_requests USING btree (repository_id, state, updated_at DESC);

CREATE INDEX quality_snapshots_repository_measured_at_idx ON public.quality_snapshots USING btree (repository_id, measured_at DESC);

CREATE INDEX quality_snapshots_repository_tool_idx ON public.quality_snapshots USING btree (repository_id, tool);

CREATE INDEX rbac_policies_is_system_idx ON public.rbac_policies USING btree (is_system);

CREATE INDEX realtime_tickets_expires_at_idx ON public.realtime_tickets USING btree (expires_at);

CREATE INDEX repositories_owner_login_idx ON public.repositories USING btree (owner_login);

CREATE INDEX repositories_status_updated_at_idx ON public.repositories USING btree (repository_status, updated_at DESC);

CREATE INDEX risks_status_impact_updated_at_idx ON public.risks USING btree (status, impact, updated_at DESC);

CREATE INDEX scm_providers_enabled_idx ON public.scm_providers USING btree (enabled);

CREATE INDEX unit_appointments_unit_idx ON public.unit_appointments USING btree (unit_id);

CREATE INDEX unit_appointments_user_idx ON public.unit_appointments USING btree (user_id);

CREATE UNIQUE INDEX unit_single_leader_idx ON public.unit_appointments USING btree (unit_id) WHERE (appointment_role = 'leader'::text);

CREATE UNIQUE INDEX users_idp_subject_unique ON public.users USING btree (idp_subject) WHERE (idp_subject IS NOT NULL);

CREATE INDEX users_primary_unit_idx ON public.users USING btree (primary_unit_id);

CREATE INDEX users_review_status_idx ON public.users USING btree (review_status) WHERE (review_status IS NOT NULL);

CREATE INDEX users_role_idx ON public.users USING btree (role);

CREATE INDEX users_status_idx ON public.users USING btree (status);

CREATE UNIQUE INDEX webhook_events_dedupe_key_idx ON public.webhook_events USING btree (dedupe_key);

CREATE INDEX webhook_events_event_type_received_at_idx ON public.webhook_events USING btree (event_type, received_at DESC);

CREATE INDEX webhook_events_repository_name_received_at_idx ON public.webhook_events USING btree (repository_name, received_at DESC);

CREATE INDEX webhook_events_status_received_at_idx ON public.webhook_events USING btree (status, received_at DESC);

ALTER TABLE ONLY public.platform_repositories
    ADD CONSTRAINT platform_repositories_platform_id_fkey FOREIGN KEY (platform_id) REFERENCES public.platforms(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.platform_repositories
    ADD CONSTRAINT platform_repositories_repo_provider_fkey FOREIGN KEY (repo_provider) REFERENCES public.scm_providers(provider_key) ON DELETE RESTRICT;

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_development_unit_id_fkey FOREIGN KEY (development_unit_id) REFERENCES public.org_units(unit_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_leader_user_id_fkey FOREIGN KEY (leader_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_command_id_fkey FOREIGN KEY (command_id) REFERENCES public.commands(command_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.build_runs
    ADD CONSTRAINT build_runs_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.ci_runs
    ADD CONSTRAINT ci_runs_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.dev_request_intake_tokens
    ADD CONSTRAINT dev_request_intake_tokens_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(user_id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.dev_requests
    ADD CONSTRAINT dev_requests_assignee_user_id_fkey FOREIGN KEY (assignee_user_id) REFERENCES public.users(user_id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.external_task_items
    ADD CONSTRAINT external_task_items_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.integration_providers(provider_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.integration_bindings
    ADD CONSTRAINT integration_bindings_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.integration_providers(provider_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.integration_sync_jobs
    ADD CONSTRAINT integration_sync_jobs_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.integration_providers(provider_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.issues
    ADD CONSTRAINT issues_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.org_units
    ADD CONSTRAINT org_units_parent_unit_id_fkey FOREIGN KEY (parent_unit_id) REFERENCES public.org_units(unit_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.pr_activities
    ADD CONSTRAINT pr_activities_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_integrations
    ADD CONSTRAINT project_integrations_platform_id_fkey FOREIGN KEY (platform_id) REFERENCES public.platforms(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_integrations
    ADD CONSTRAINT project_integrations_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_members
    ADD CONSTRAINT project_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_repositories
    ADD CONSTRAINT project_repositories_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.project_repositories
    ADD CONSTRAINT project_repositories_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_platform_id_fkey FOREIGN KEY (platform_id) REFERENCES public.platforms(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(user_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.pull_requests
    ADD CONSTRAINT pull_requests_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.quality_snapshots
    ADD CONSTRAINT quality_snapshots_repository_id_fkey FOREIGN KEY (repository_id) REFERENCES public.repositories(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.repositories
    ADD CONSTRAINT repositories_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES public.integration_providers(provider_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.unit_appointments
    ADD CONSTRAINT unit_appointments_unit_id_fkey FOREIGN KEY (unit_id) REFERENCES public.org_units(unit_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.unit_appointments
    ADD CONSTRAINT unit_appointments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE;

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_current_unit_id_fkey FOREIGN KEY (current_unit_id) REFERENCES public.org_units(unit_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_primary_unit_id_fkey FOREIGN KEY (primary_unit_id) REFERENCES public.org_units(unit_id) ON DELETE SET NULL;

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES public.rbac_policies(role_id);

