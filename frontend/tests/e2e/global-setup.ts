import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const KC_BASE_URL = (
  process.env.DEVHUB_E2E_KEYCLOAK_ADMIN_URL
  ?? process.env.DEVHUB_KEYCLOAK_ADMIN_URL
  ?? process.env.KEYCLOAK_URL
  ?? "http://localhost:8180/devhub/auth/keycloak"
).replace(/\/+$/, "");
const KC_REALM = (process.env.DEVHUB_KEYCLOAK_ADMIN_REALM ?? "devhub").trim();
const KC_ADMIN_CLIENT_ID = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_ID ?? "devhub-backend").trim();
const KC_ADMIN_CLIENT_SECRET = (process.env.DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET ?? "secret-change-me-backend").trim();

const DSN = process.env.DSN ?? "";
const SKIP_SEED = process.env.DEVHUB_E2E_SKIP_SEED === "1";

type Seed = {
  user_id: string;
  email: string;
  display_name: string;
  password: string;
  role: "developer" | "team_manager" | "system_admin";
};

const SEEDS: readonly Seed[] = [
  { user_id: "alice", email: "alice@example.com", display_name: "Alice", password: "ChangeMe-12345!", role: "developer" },
  { user_id: "bob", email: "bob@example.com", display_name: "Bob", password: "ChangeMe-12345!", role: "team_manager" },
  { user_id: "charlie", email: "charlie@example.com", display_name: "Charlie", password: "ChangeMe-12345!", role: "system_admin" },
];

type KeycloakUser = {
  id: string;
  username?: string;
  email?: string;
};

type KeycloakRole = {
  id: string;
  name: string;
};

async function fetchAdminTokenPassword(): Promise<string> {
  const adminUser = process.env.DEVHUB_KEYCLOAK_ADMIN_USERNAME ?? "admin";
  const adminPass = process.env.DEVHUB_KEYCLOAK_ADMIN_PASSWORD ?? "admin";
  // Bootstrap admin lives in the master realm; KC_REALM (devhub) may not have
  // a local admin user with password grant access.
  const tokenURL = `${KC_BASE_URL}/realms/master/protocol/openid-connect/token`;
  const resp = await fetch(tokenURL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "password",
      client_id: "admin-cli",
      username: adminUser,
      password: adminPass,
    }),
  });
  if (!resp.ok) {
    throw new Error(`Keycloak admin password grant failed ${resp.status}: ${await resp.text()}`);
  }
  const body = (await resp.json()) as { access_token?: string };
  if (!body.access_token) {
    throw new Error("Keycloak admin password grant response missing access_token");
  }
  return body.access_token;
}

async function fetchAdminToken(): Promise<string> {
  if (!KC_ADMIN_CLIENT_SECRET) {
    throw new Error("DEVHUB_KEYCLOAK_ADMIN_CLIENT_SECRET is required for e2e global setup");
  }
  const tokenURL = `${KC_BASE_URL}/realms/${encodeURIComponent(KC_REALM)}/protocol/openid-connect/token`;
  const resp = await fetch(tokenURL, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "client_credentials",
      client_id: KC_ADMIN_CLIENT_ID,
      client_secret: KC_ADMIN_CLIENT_SECRET,
    }),
  });
  if (!resp.ok) {
    throw new Error(`Keycloak token grant failed ${resp.status}: ${await resp.text()}`);
  }
  const body = (await resp.json()) as { access_token?: string };
  if (!body.access_token) {
    throw new Error("Keycloak token response missing access_token");
  }
  return body.access_token;
}

async function findUserByEmail(token: string, email: string): Promise<KeycloakUser | null> {
  const url = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users?email=${encodeURIComponent(email)}&exact=true`;
  const resp = await fetch(url, {
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
  });
  if (!resp.ok) {
    throw new Error(`Keycloak list users failed ${resp.status}: ${await resp.text()}`);
  }
  const users = (await resp.json()) as KeycloakUser[];
  return users[0] ?? null;
}

async function createUser(token: string, seed: Seed): Promise<string> {
  const url = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users`;
  const payload = {
    username: seed.user_id,
    email: seed.email,
    firstName: seed.display_name,
    enabled: true,
    emailVerified: true,
    credentials: [{ type: "password", value: seed.password, temporary: false }],
    attributes: {
      user_id: [seed.user_id],
      employee_id: [seed.user_id],
    },
  };
  const resp = await fetch(url, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  });
  if (resp.status !== 201) {
    throw new Error(`Keycloak create user ${seed.user_id} failed ${resp.status}: ${await resp.text()}`);
  }
  const location = resp.headers.get("location") ?? "";
  const id = location.split("/").pop()?.trim();
  if (!id) {
    throw new Error(`Keycloak create user ${seed.user_id} missing Location header`);
  }
  return id;
}

async function resetPassword(token: string, userID: string, password: string): Promise<void> {
  const url = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users/${encodeURIComponent(userID)}/reset-password`;
  const resp = await fetch(url, {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ type: "password", value: password, temporary: false }),
  });
  if (!resp.ok) {
    throw new Error(`Keycloak reset password ${userID} failed ${resp.status}: ${await resp.text()}`);
  }
}

async function ensureRealmRole(token: string, userID: string, _roleName: string): Promise<void> {
  // ADR-0026: Keycloak has a single 'user' realm role. The actual DevHub
  // role is stored in DB users.role and seeded separately via SQL.
  const realmRole = "user";
  // Use the caller-provided token (client_credentials) for role operations.
  // Password grant (fetchAdminTokenPassword) may fail if admin-cli has
  // directAccessGrantsEnabled=false.
  const roleURL = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/roles/${encodeURIComponent(realmRole)}`;
  let roleResp = await fetch(roleURL, {
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
  });
  if (roleResp.status === 404) {
    // ADR-0026: Keycloak realm role is not the source of truth, but the JWT
    // token still needs the role claim for authenticateActor fallback. Create
    // the role on demand so E2E setup works without manual realm configuration.
    const createResp = await fetch(`${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/roles`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name: realmRole }),
    });
    if (!createResp.ok) {
      throw new Error(`Keycloak role create ${realmRole} failed ${createResp.status}: ${await createResp.text()}`);
    }
    roleResp = await fetch(roleURL, {
      headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
    });
  }
  if (!roleResp.ok) {
    throw new Error(`Keycloak role lookup ${realmRole} failed ${roleResp.status}: ${await roleResp.text()}`);
  }
  const role = (await roleResp.json()) as KeycloakRole;

  const mappingURL = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/users/${encodeURIComponent(userID)}/role-mappings/realm`;
  const mapResp = await fetch(mappingURL, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify([{ id: role.id, name: role.name }]),
  });
  if (!mapResp.ok) {
    throw new Error(`Keycloak role map ${userID} -> ${realmRole} failed ${mapResp.status}: ${await mapResp.text()}`);
  }
}

async function seedKeycloakUsers(): Promise<Record<string, string>> {
  const token = await fetchAdminToken();
  const idMap: Record<string, string> = {};
  for (const seed of SEEDS) {
    const existing = await findUserByEmail(token, seed.email);
    const userID = existing?.id ?? (await createUser(token, seed));
    try {
      await ensureRealmRole(token, userID, "user");
    } catch (err) {
      // ADR-0026: Keycloak realm role is not the RBAC source of truth.
      // Role mapping failure does not affect E2E tests; the JWT token's
      // realm role claim is only a fallback for authenticateActor.
      // DB users.role is the single source of truth.
      console.warn(`[e2e seed] realm role mapping skipped for ${seed.email}: ${err}`);
    }
    idMap[seed.email] = userID;
    console.log(`[e2e seed] keycloak user ${seed.email} ${existing ? "present" : "created"} (sub: ${userID}) → role mapped (${seed.role})`);
  }
  return idMap;
}

// PostgreSQL single-quote escape (' → ''). SEEDS 는 hardcoded array 라 실 위험
// 낮지만 prepared statement 패턴이 없는 idp-apply-schemas 경로라 명시 escape.
function sqlEscape(value: string): string {
  return value.replace(/'/g, "''");
}

function seedDevhubData(idMap: Record<string, string>): void {
  if (!DSN) {
    throw new Error("DSN env var is required for global-setup (DevHub users seed)");
  }

  // Dynamic SQL for UPSERTing users with correct idp_subject (sprint -m design Option 1).
  // PR #290 의 OnboardingGateEnabled default ON flip 이후, onboarding_completed_at NULL
  // 사용자는 backend `onboardingGate` 가 allowlist 외 endpoint 호출을 403 으로 차단한다.
  // e2e seed users 는 onboarding 완료 상태 (`reviewed`) 로 INSERT 해 admin/account 등
  // 보호 endpoint 접근을 허용한다 (codex hotfix PR #288 회귀 추적의 e2e 갈래).
  const values = SEEDS.map(s => {
    const sub = idMap[s.email] ?? "";
    return `('${sqlEscape(s.user_id)}', '${sqlEscape(s.email)}', '${sqlEscape(s.display_name)}', '${sqlEscape(s.role)}', 'active', '2026-01-01', 'human', '${sqlEscape(sub)}', NOW(), 'reviewed')`;
  }).join(",\n    ");

  const sql = `-- Dynamic seed from global-setup.ts
-- Ensure rbac_policies has the required role_ids before seeding users.
-- Bypass FK/CHECK temporarily to fix legacy rows from incomplete migrations.
SET session_replication_role = replica;
ALTER TABLE rbac_policies DROP CONSTRAINT IF EXISTS rbac_policies_role_id_format;
DELETE FROM rbac_policies WHERE role_id = 'pmo_manager';
UPDATE rbac_policies SET role_id = 'team_manager' WHERE role_id = 'manager';
INSERT INTO rbac_policies (role_id, name, description, is_system, permissions)
VALUES
    ('developer', 'Developer', 'Developer', true, '{}'::jsonb),
    ('team_manager', 'Manager', 'Team Manager', true, '{}'::jsonb),
    ('system_admin', 'System Admin', 'System Administrator', true, '{}'::jsonb)
ON CONFLICT (role_id) DO NOTHING;
ALTER TABLE rbac_policies ADD CONSTRAINT rbac_policies_role_id_format CHECK (
    role_id IN ('developer', 'team_manager', 'system_admin')
    OR role_id ~ '^custom-[a-z0-9][a-z0-9_-]{0,62}$'
);
SET session_replication_role = origin;

INSERT INTO users (user_id, email, display_name, role, status, joined_at, user_type, idp_subject, onboarding_completed_at, review_status)
VALUES
    ${values}
ON CONFLICT (user_id) DO UPDATE SET
    idp_subject = EXCLUDED.idp_subject,
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    onboarding_completed_at = EXCLUDED.onboarding_completed_at,
    review_status = EXCLUDED.review_status;

-- Ensure repository fixtures exist for project-management e2e scenarios.
INSERT INTO repositories (gitea_repository_id, full_name, owner_login, name, clone_url, html_url, default_branch, private)
VALUES
    (100001, 'devhub/e2e-repo-a', 'devhub', 'e2e-repo-a', 'https://example.invalid/devhub/e2e-repo-a.git', 'https://example.invalid/devhub/e2e-repo-a', 'main', false),
    (100002, 'devhub/e2e-repo-b', 'devhub', 'e2e-repo-b', 'https://example.invalid/devhub/e2e-repo-b.git', 'https://example.invalid/devhub/e2e-repo-b', 'main', false)
ON CONFLICT (full_name) DO UPDATE SET
    owner_login = EXCLUDED.owner_login,
    name = EXCLUDED.name,
    clone_url = EXCLUDED.clone_url,
    html_url = EXCLUDED.html_url,
    default_branch = EXCLUDED.default_branch,
    private = EXCLUDED.private,
    updated_at = NOW();

-- Ensure integration provider 'gitea' exists for publish e2e scenario.
INSERT INTO integration_providers (
    provider_id, provider_key, provider_type, display_name, enabled, auth_mode,
    credentials_ref, capabilities, sync_status, base_url, api_token,
    auth_username, auth_client_id, auth_token_url, auth_secret,
    webhook_secret, pull_interval_seconds
) VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'gitea', 'scm', 'Local Gitea', true, 'token',
    'credentials_ref_gitea', '["push"]'::jsonb, 'active', 'http://localhost:3000', 'gitea-token',
    null, null, null, null,
    null, 1800
)
ON CONFLICT (provider_key) DO UPDATE SET
    provider_type = EXCLUDED.provider_type,
    display_name = EXCLUDED.display_name,
    enabled = EXCLUDED.enabled,
    auth_mode = EXCLUDED.auth_mode,
    base_url = EXCLUDED.base_url,
    api_token = EXCLUDED.api_token,
    updated_at = NOW();

-- Ensure platform fixture for e2e details exists
INSERT INTO platforms (id, key, name, description, status, visibility, owner_user_id, start_date, due_date)
VALUES
    ('e8a9bc11-a89c-4cb1-8071-8890ab2345ef', 'DEVHUBAPP1', 'DevHub Simulation App', 'DevHub Simulation App for E2E testing', 'active', 'public', 'charlie', '2026-01-01', '2026-12-31')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    key = EXCLUDED.key,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    visibility = EXCLUDED.visibility,
    owner_user_id = EXCLUDED.owner_user_id;

-- Ensure project fixture for e2e details exists
INSERT INTO projects (id, platform_id, repository_id, key, name, description, status, visibility, owner_user_id, start_date, due_date)
VALUES
    ('31b9e2cb-b1b0-466a-bb10-ea00ee1234a1', 'e8a9bc11-a89c-4cb1-8071-8890ab2345ef', (SELECT id FROM repositories WHERE full_name = 'devhub/e2e-repo-a'), 'DEVHUBPROJ', 'DevHub Simulation Project', 'DevHub Simulation Project for E2E testing', 'active', 'public', 'charlie', '2026-01-01', '2026-12-31')
ON CONFLICT (id) DO UPDATE SET
    platform_id = EXCLUDED.platform_id,
    repository_id = EXCLUDED.repository_id,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    visibility = EXCLUDED.visibility,
    owner_user_id = EXCLUDED.owner_user_id;

-- Ensure project repository association exists
INSERT INTO project_repositories (project_id, repository_id, role)
VALUES
    ('31b9e2cb-b1b0-466a-bb10-ea00ee1234a1', (SELECT id FROM repositories WHERE full_name = 'devhub/e2e-repo-a'), 'primary')
ON CONFLICT (project_id, repository_id) DO NOTHING;

-- Ensure platform repository association exists
INSERT INTO platform_repositories (platform_id, repo_provider, repo_full_name, role, sync_status)
VALUES
    ('e8a9bc11-a89c-4cb1-8071-8890ab2345ef', 'gitea', 'devhub/e2e-repo-a', 'primary', 'active')
ON CONFLICT (platform_id, repo_provider, repo_full_name) DO NOTHING;
`;

  // Unique file name — parallel e2e shard 간 collision 회피 (process.pid + timestamp).
  const tempSqlPath = path.resolve(__dirname, `temp_seed_e2e_users_${process.pid}_${Date.now()}.sql`);
  fs.writeFileSync(tempSqlPath, sql);

  try {
    const backendDir = path.resolve(__dirname, "..", "..", "..", "backend-core");
    const result = spawnSync("go", ["run", "./cmd/idp-apply-schemas", "-dsn", DSN, "-sql", tempSqlPath], {
      cwd: backendDir,
      env: { ...process.env, DSN, DEVHUB_DB_URL: DSN },
      stdio: "inherit",
      shell: process.platform === "win32",
    });
    if (result.status !== 0) {
      throw new Error(`idp-apply-schemas exited with status ${result.status}`);
    }
    console.log("[e2e seed] DevHub users + repositories fixture seeded");
  } finally {
    // Always clean up temp SQL even on error.
    try { fs.unlinkSync(tempSqlPath); } catch { /* ignore */ }
  }
}

export default async function globalSetup(): Promise<void> {
  if (SKIP_SEED) {
    console.log("[e2e seed] DEVHUB_E2E_SKIP_SEED=1 -> skipping seed");
    return;
  }
  const idMap = await seedKeycloakUsers();
  seedDevhubData(idMap);
}
