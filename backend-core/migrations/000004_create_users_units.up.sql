CREATE TABLE org_units (
    id BIGSERIAL PRIMARY KEY,
    unit_id TEXT NOT NULL UNIQUE,
    parent_unit_id TEXT REFERENCES org_units(unit_id) ON DELETE SET NULL,
    unit_type TEXT NOT NULL,
    label TEXT NOT NULL,
    leader_user_id TEXT,
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT org_units_type_check CHECK (unit_type IN ('division', 'team', 'group', 'part', 'company'))
);

CREATE INDEX org_units_parent_idx ON org_units (parent_unit_id);
CREATE INDEX org_units_type_idx ON org_units (unit_type);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL,
    status TEXT NOT NULL,
    primary_unit_id TEXT REFERENCES org_units(unit_id) ON DELETE SET NULL,
    current_unit_id TEXT REFERENCES org_units(unit_id) ON DELETE SET NULL,
    is_seconded BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_role_check CHECK (role IN ('developer', 'manager', 'system_admin')),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'pending', 'deactivated'))
);

CREATE INDEX users_role_idx ON users (role);
CREATE INDEX users_status_idx ON users (status);
CREATE INDEX users_primary_unit_idx ON users (primary_unit_id);

CREATE TABLE unit_appointments (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    unit_id TEXT NOT NULL REFERENCES org_units(unit_id) ON DELETE CASCADE,
    appointment_role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unit_appointments_role_check CHECK (appointment_role IN ('leader', 'member')),
    CONSTRAINT unit_appointments_unique UNIQUE (user_id, unit_id)
);

CREATE INDEX unit_appointments_unit_idx ON unit_appointments (unit_id);
CREATE INDEX unit_appointments_user_idx ON unit_appointments (user_id);

-- Seed (frontend lib/services/identity.service.ts mock과 일치)
INSERT INTO org_units (unit_id, parent_unit_id, unit_type, label, leader_user_id, position_x, position_y) VALUES
    ('org-root',     NULL,         'company',  'DevHub Global',   'u1', 400, 0),
    ('dept-eng',     'org-root',   'division', 'Engineering',     'u1', 200, 150),
    ('dept-prod',    'org-root',   'division', 'Product',         'u2', 600, 150),
    ('team-infra',   'dept-eng',   'team',     'Infrastructure',  'u1', 50,  300),
    ('team-frontend','dept-eng',   'team',     'Frontend',        'u3', 350, 300),
    ('team-ux',      'dept-prod',  'team',     'UX Strategy',     'u2', 600, 300),
    ('part-security','team-infra', 'part',     'Security Part',   NULL, 50,  450);

INSERT INTO users (user_id, email, display_name, role, status, primary_unit_id, current_unit_id, is_seconded, joined_at) VALUES
    ('u1', 'yklee@example.com', 'YK Lee',   'system_admin', 'active', 'dept-eng',   'dept-eng',  FALSE, '2026-01-15'),
    ('u2', 'alex@example.com',  'Alex Kim', 'manager',      'active', 'dept-prod',  'team-ux',   TRUE,  '2026-02-01'),
    ('u3', 'sam@example.com',   'Sam Jones','developer',    'active', 'team-infra', 'team-infra',FALSE, '2026-05-01');

INSERT INTO unit_appointments (user_id, unit_id, appointment_role) VALUES
    ('u1', 'org-root',   'leader'),
    ('u1', 'dept-eng',   'leader'),
    ('u2', 'dept-prod',  'leader'),
    ('u3', 'team-infra', 'member');
