import { spawnSync } from "node:child_process";
import path from "node:path";

// global-setup.ts (PR-T3.5, work_26_05_11-e): idempotently seed the three
// e2e users used by frontend/tests/e2e/fixtures.ts. Runs once per
// `npm run e2e` invocation. Skip via DEVHUB_E2E_SKIP_SEED=1 for CI matrix
// runs that drive seeding from a separate stage.
//
// The seed is split in two:
//   1. Kratos identities — POST /admin/identities when missing, PUT
//      /admin/identities/{id} to force-reset the seed password when the
//      identity already exists (PR-T3.5 hardening, work_260512-f). The PUT
//      keeps traits/state/metadata intact and only refreshes
//      credentials.password — Kratos hashes the plaintext on receive, and
//      other credential methods (if any) survive because PUT only updates
//      methods supplied in the payload. This closes the documented
//      stale-rotation gap: if password-change.spec's finally rollback fails
//      mid-run, the next `npm run e2e` restores the seed automatically.
//   2. DevHub users row — runs `go run ./cmd/idp-apply-schemas -sql
//      infra/idp/sql/002_seed_e2e_users.sql`. The helper is already used by
//      operators in the manual flow; reusing it keeps a single seeding
//      pathway. Requires DSN env so the helper can connect to PostgreSQL.

const KRATOS_ADMIN_URL = (process.env.KRATOS_ADMIN_URL ?? "http://localhost:4434").replace(/\/$/, "");
const DSN = process.env.DSN ?? "";
const SKIP_SEED = process.env.DEVHUB_E2E_SKIP_SEED === "1";

interface KratosSeed {
  user_id: string;
  email: string;
  display_name: string;
  password: string;
  role: string;
}

const SEEDS: readonly KratosSeed[] = [
  { user_id: "alice", email: "alice@example.com", display_name: "Alice", password: "ChangeMe-12345!", role: "developer" },
  { user_id: "bob", email: "bob@example.com", display_name: "Bob", password: "ChangeMe-12345!", role: "manager" },
  { user_id: "charlie", email: "charlie@example.com", display_name: "Charlie", password: "ChangeMe-12345!", role: "system_admin" },
];

interface KratosIdentityFull {
  id: string;
  schema_id: string;
  state: string;
  traits?: { email?: string; [k: string]: unknown };
  metadata_public?: Record<string, unknown> | null;
  metadata_admin?: Record<string, unknown> | null;
}

async function listExistingIdentities(): Promise<Map<string, KratosIdentityFull>> {
  // Returns email (lowercased) -> identity mapping. Kratos /admin/identities
  // pagination is 0-based — verified against v26.2.0 (page=0 returns first
  // batch). The earlier 1-based start silently returned an empty first page
  // and made the seed dedupe check think no identities existed, which then
  // 409'd on POST.
  const out = new Map<string, KratosIdentityFull>();
  let page = 0;
  const perPage = 250;
  while (page < 40) {
    const url = `${KRATOS_ADMIN_URL}/admin/identities?page=${page}&per_page=${perPage}`;
    const resp = await fetch(url, { headers: { Accept: "application/json" } });
    if (!resp.ok) {
      throw new Error(`Kratos admin list identities ${resp.status}: ${await resp.text()}`);
    }
    const batch = (await resp.json()) as KratosIdentityFull[];
    for (const ident of batch) {
      const email = ident.traits?.email?.toLowerCase();
      if (email && ident.id) out.set(email, ident);
    }
    if (batch.length < perPage) break;
    page += 1;
  }
  return out;
}

async function createKratosIdentity(seed: KratosSeed): Promise<void> {
  const payload = {
    schema_id: "devhub_user",
    state: "active",
    traits: { system_id: seed.user_id, email: seed.email, display_name: seed.display_name },
    metadata_public: { user_id: seed.user_id },
    credentials: { password: { config: { password: seed.password } } },
  };
  const resp = await fetch(`${KRATOS_ADMIN_URL}/admin/identities`, {
    method: "POST",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(payload),
  });
  if (!resp.ok) {
    throw new Error(`Kratos admin create identity ${seed.user_id} → ${resp.status}: ${await resp.text()}`);
  }
}

async function resetKratosPassword(identity: KratosIdentityFull, seed: KratosSeed): Promise<void> {
  // PUT /admin/identities/{id} replaces the identity. We echo schema_id /
  // state / traits / metadata from the list response so unmanaged fields
  // (e.g. metadata_admin populated by another process) survive, and supply
  // credentials.password as plaintext — Kratos hashes on receive. Kratos
  // only updates credential methods present in the payload, so non-password
  // methods (if ever added) are not wiped.
  const payload: Record<string, unknown> = {
    schema_id: identity.schema_id,
    state: identity.state,
    traits: identity.traits ?? {},
    credentials: { password: { config: { password: seed.password } } },
  };
  if (identity.metadata_public != null) payload.metadata_public = identity.metadata_public;
  if (identity.metadata_admin != null) payload.metadata_admin = identity.metadata_admin;
  const resp = await fetch(`${KRATOS_ADMIN_URL}/admin/identities/${identity.id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify(payload),
  });
  if (!resp.ok) {
    throw new Error(`Kratos admin reset password ${seed.email} → ${resp.status}: ${await resp.text()}`);
  }
}

async function seedKratos(): Promise<void> {
  const existing = await listExistingIdentities();
  for (const seed of SEEDS) {
    const identity = existing.get(seed.email.toLowerCase());
    if (identity) {
      await resetKratosPassword(identity, seed);
      console.log(`[e2e seed] kratos identity ${seed.email} present → password force-reset to seed value`);
      continue;
    }
    await createKratosIdentity(seed);
    console.log(`[e2e seed] created kratos identity ${seed.email}`);
  }
}

function seedDevhubUsers(): void {
  if (!DSN) {
    throw new Error("DSN env var is required for global-setup to seed DevHub users (see docs/setup/e2e-test-guide.md §2)");
  }
  const backendDir = path.resolve(__dirname, "..", "..", "..", "backend-core");
  const sqlPath = path.resolve(__dirname, "..", "..", "..", "infra", "idp", "sql", "002_seed_e2e_users.sql");
  const result = spawnSync("go", ["run", "./cmd/idp-apply-schemas", "-dsn", DSN, "-sql", sqlPath], {
    cwd: backendDir,
    env: { ...process.env, DSN, DEVHUB_DB_URL: DSN },
    stdio: "inherit",
    shell: process.platform === "win32",
  });
  if (result.status !== 0) {
    throw new Error(`idp-apply-schemas exited with status ${result.status} (DSN seed for 002_seed_e2e_users.sql)`);
  }
  console.log("[e2e seed] DevHub users row seeded via idp-apply-schemas");
}

export default async function globalSetup(): Promise<void> {
  if (SKIP_SEED) {
    console.log("[e2e seed] DEVHUB_E2E_SKIP_SEED=1 → skipping seed");
    return;
  }
  await seedKratos();
  seedDevhubUsers();
}
