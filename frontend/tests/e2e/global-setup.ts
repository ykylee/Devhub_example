import { spawnSync } from "node:child_process";
import path from "node:path";

const KC_BASE_URL = (process.env.DEVHUB_KEYCLOAK_ADMIN_URL ?? "http://localhost:8180/devhub/auth/keycloak").replace(/\/+$/, "");
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
  role: "developer" | "manager" | "system_admin";
};

const SEEDS: readonly Seed[] = [
  { user_id: "alice", email: "alice@example.com", display_name: "Alice", password: "ChangeMe-12345!", role: "developer" },
  { user_id: "bob", email: "bob@example.com", display_name: "Bob", password: "ChangeMe-12345!", role: "manager" },
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

async function ensureRealmRole(token: string, userID: string, roleName: Seed["role"]): Promise<void> {
  const roleURL = `${KC_BASE_URL}/admin/realms/${encodeURIComponent(KC_REALM)}/roles/${encodeURIComponent(roleName)}`;
  const roleResp = await fetch(roleURL, {
    headers: { Authorization: `Bearer ${token}`, Accept: "application/json" },
  });
  if (!roleResp.ok) {
    throw new Error(`Keycloak role lookup ${roleName} failed ${roleResp.status}: ${await roleResp.text()}`);
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
    throw new Error(`Keycloak role map ${userID} -> ${roleName} failed ${mapResp.status}: ${await mapResp.text()}`);
  }
}

async function seedKeycloakUsers(): Promise<Record<string, string>> {
  const token = await fetchAdminToken();
  const idMap: Record<string, string> = {};
  for (const seed of SEEDS) {
    const existing = await findUserByEmail(token, seed.email);
    const userID = existing?.id ?? (await createUser(token, seed));
    await resetPassword(token, userID, seed.password);
    await ensureRealmRole(token, userID, seed.role);
    idMap[seed.email] = userID;
    console.log(`[e2e seed] keycloak user ${seed.email} ${existing ? "present" : "created"} (sub: ${userID}) → password reset + role mapped (${seed.role})`);
  }
  return idMap;
}

function seedDevhubUsers(idMap: Record<string, string>): void {
  if (!DSN) {
    throw new Error("DSN env var is required for global-setup (DevHub users seed)");
  }

  // Create dynamic SQL for UPSERTing users with correct idp_subject (Option 1)
  const values = SEEDS.map(s => {
    const sub = idMap[s.email];
    return `('${s.user_id}', '${s.email}', '${s.display_name}', '${s.role}', 'active', '2026-01-01', 'human', '${sub}')`;
  }).join(",\n    ");

  const sql = `-- Dynamic seed from global-setup.ts
INSERT INTO users (user_id, email, display_name, role, status, joined_at, user_type, idp_subject)
VALUES
    ${values}
ON CONFLICT (user_id) DO UPDATE SET
    idp_subject = EXCLUDED.idp_subject,
    role = EXCLUDED.role,
    status = EXCLUDED.status;
`;

  const tempSqlPath = path.resolve(__dirname, "temp_seed_e2e_users.sql");
  require("node:fs").writeFileSync(tempSqlPath, sql);

  const backendDir = path.resolve(__dirname, "..", "..", "..", "backend-core");
  const result = spawnSync("go", ["run", "./cmd/idp-apply-schemas", "-dsn", DSN, "-sql", tempSqlPath], {
    cwd: backendDir,
    env: { ...process.env, DSN, DEVHUB_DB_URL: DSN },
    stdio: "inherit",
    shell: process.platform === "win32",
  });

  // Clean up temp SQL
  try { require("node:fs").unlinkSync(tempSqlPath); } catch {}

  if (result.status !== 0) {
    throw new Error(`idp-apply-schemas exited with status ${result.status}`);
  }
  console.log("[e2e seed] DevHub users row seeded with idp_subject sync");
}

export default async function globalSetup(): Promise<void> {
  if (SKIP_SEED) {
    console.log("[e2e seed] DEVHUB_E2E_SKIP_SEED=1 -> skipping seed");
    return;
  }
  const idMap = await seedKeycloakUsers();
  seedDevhubUsers(idMap);
}
