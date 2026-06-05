import type { Page } from "@playwright/test";

type DogfoodApiResult<T> = {
  status: number;
  body: T;
};

type ProviderEnvelope = {
  data?: {
    provider_id?: string;
    provider_key?: string;
    display_name?: string;
    sync_status?: string;
  };
};

export type DogfoodProviderFixture = {
  providerID: string;
  providerKey: string;
  displayName: string;
  syncStatus: string | null;
};

export type DogfoodWorkspaceFixture = {
  platformID: string;
  platformKey: string;
  platformName: string;
  projectID: string;
  projectKey: string;
  projectName: string;
  repositoryID: number;
  repositoryKey: string;
  repositorySlug: string;
  platformRepoCount: number;
  projectRepoCount: number;
};

function requiredEnv(name: string): string {
  const value = process.env[name]?.trim();
  if (!value) {
    throw new Error(`${name} is required for dogfood scenario`);
  }
  return value;
}

async function runDogfoodApi<T>(page: Page, method: string, path: string, body?: unknown): Promise<DogfoodApiResult<T>> {
  const result = await page.evaluate(
    async ({ method, path, body }) => {
      const token = sessionStorage.getItem("devhub_access_token");
      if (!token) {
        throw new Error("missing access token");
      }

      const apiBasePath = window.location.pathname.startsWith("/devhub/") ? "/devhub" : "";
      const resp = await fetch(`${apiBasePath}${path}`, {
        method,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: body === undefined ? undefined : JSON.stringify(body),
      });

      const raw = await resp.text();
      let parsed: unknown = { raw };
      try {
        parsed = raw ? JSON.parse(raw) : {};
      } catch {
        parsed = { raw };
      }

      return {
        status: resp.status,
        body: parsed,
      };
    },
    { method, path, body },
  );

  return result as DogfoodApiResult<T>;
}

export async function createDogfoodScmProvider(page: Page): Promise<DogfoodProviderFixture> {
  const giteaURL = requiredEnv("GITEA_URL");
  const giteaToken = requiredEnv("GITEA_TOKEN");
  const webhookSecret = requiredEnv("GITEA_WEBHOOK_SECRET");
  const unique = `${Date.now()}`;
  const providerKey = `dogfood-gitea-${unique}`;
  const displayName = `Dogfood Gitea ${unique}`;

  const result = await runDogfoodApi<ProviderEnvelope>(page, "POST", "/api/v1/integration/providers", {
    provider_key: providerKey,
    provider_type: "scm",
    display_name: displayName,
    auth_mode: "token",
    credentials_ref: `hmac_sha256:${webhookSecret}`,
    capabilities: ["pull", "sync"],
    base_url: giteaURL,
    api_token: giteaToken,
    webhook_secret: webhookSecret,
  });

  if (result.status !== 201) {
    throw new Error(`create provider failed: ${result.status} ${JSON.stringify(result.body)}`);
  }

  const provider = result.body?.data;
  if (!provider?.provider_id || !provider?.provider_key || !provider?.display_name) {
    throw new Error(`provider payload incomplete: ${JSON.stringify(result.body)}`);
  }

  return {
    providerID: provider.provider_id,
    providerKey: provider.provider_key,
    displayName: provider.display_name,
    syncStatus: provider.sync_status ?? null,
  };
}

export async function syncDogfoodProvider(page: Page, providerID: string): Promise<{ jobID: string }> {
  const result = await runDogfoodApi<{ status?: string; job_id?: string }>(
    page,
    "POST",
    `/api/v1/integration/providers/${providerID}/sync`,
  );
  if (result.status !== 202 || result.body?.status !== "accepted" || !result.body?.job_id) {
    throw new Error(`sync provider failed: ${result.status} ${JSON.stringify(result.body)}`);
  }
  return { jobID: result.body.job_id };
}

export async function listDogfoodScmRepositories(
  page: Page,
  providerID: string,
): Promise<Array<{ full_name?: string; imported?: boolean }>> {
  const result = await runDogfoodApi<{ data?: Array<{ full_name?: string; imported?: boolean }> }>(
    page,
    "GET",
    `/api/v1/integration/providers/${providerID}/scm-repositories`,
  );
  if (result.status !== 200 || !Array.isArray(result.body?.data)) {
    throw new Error(`list scm repositories failed: ${result.status} ${JSON.stringify(result.body)}`);
  }
  return result.body.data;
}

export async function deleteDogfoodProvider(page: Page, providerID: string): Promise<void> {
  const result = await runDogfoodApi<{ status?: string }>(page, "DELETE", `/api/v1/integration/providers/${providerID}`);
  if (result.status !== 200 || result.body?.status !== "ok") {
    throw new Error(`delete provider failed: ${result.status} ${JSON.stringify(result.body)}`);
  }
}

export async function createSelfDogfoodWorkspace(
  page: Page,
  repositoryProviderKey = "gitea",
): Promise<DogfoodWorkspaceFixture> {
  const unique = Date.now().toString().slice(-6);
  const platformKey = `SD${unique.slice(-6, -2)}X`;
  const projectKey = `DG${unique.slice(-4)}`;
  const platformName = `DevHub Example Codex ${unique}`;
  const projectName = `Self Dogfood Project ${unique}`;
  const repositoryKey = `SDR${unique}`;
  const repositorySlug = `yklee/devhub-example-codex-dogfood-${unique}`;

  const platformResult = await runDogfoodApi<{ data?: { id?: string } }>(page, "POST", "/api/v1/platforms", {
    key: platformKey,
    name: platformName,
    description: "Self dogfood platform for the current DevHub example workspace",
    owner_user_id: "charlie",
    leader_user_id: "charlie",
    development_unit_id: "dept-eng",
    visibility: "internal",
    status: "planning",
  });
  const platformID = platformResult.body?.data?.id;
  if (platformResult.status !== 201 || !platformID) {
    throw new Error(`create platform failed: ${platformResult.status} ${JSON.stringify(platformResult.body)}`);
  }

  const repositoryResult = await runDogfoodApi<{ data?: { id?: number; full_name?: string } }>(
    page,
    "POST",
    "/api/v1/repositories",
    {
      key: repositoryKey,
      slug: repositorySlug,
      provider_key: repositoryProviderKey,
    },
  );
  const repositoryID = repositoryResult.body?.data?.id;
  const repositoryFullName = repositoryResult.body?.data?.full_name;
  if (repositoryResult.status !== 201 || !repositoryID || !repositoryFullName) {
    throw new Error(`create repository draft failed: ${repositoryResult.status} ${JSON.stringify(repositoryResult.body)}`);
  }

  const linkResult = await runDogfoodApi(page, "POST", `/api/v1/platforms/${platformID}/repositories`, {
    repo_provider: repositoryProviderKey,
    repo_full_name: repositoryFullName,
    role: "primary",
  });
  if (linkResult.status !== 201) {
    throw new Error(`link platform repository failed: ${linkResult.status} ${JSON.stringify(linkResult.body)}`);
  }

  const projectResult = await runDogfoodApi<{ data?: { id?: string } }>(
    page,
    "POST",
    `/api/v1/platforms/${platformID}/projects`,
    {
      key: projectKey,
      name: projectName,
      description: "Self dogfood project for the current DevHub example workspace",
      owner_user_id: "charlie",
      visibility: "internal",
      status: "planning",
      repository_id: repositoryID,
      repository_ids: [repositoryID],
    },
  );
  const projectID = projectResult.body?.data?.id;
  if (projectResult.status !== 201 || !projectID) {
    throw new Error(`create project failed: ${projectResult.status} ${JSON.stringify(projectResult.body)}`);
  }

  const platformRepos = await runDogfoodApi<{ data?: unknown[] }>(page, "GET", `/api/v1/platforms/${platformID}/repositories`);
  const projectRepos = await runDogfoodApi<{ data?: unknown[] }>(page, "GET", `/api/v1/projects/${projectID}/repositories`);
  if (platformRepos.status !== 200 || projectRepos.status !== 200) {
    throw new Error(
      `list linked repositories failed: platform=${platformRepos.status} project=${projectRepos.status}`,
    );
  }

  return {
    platformID,
    platformKey,
    platformName,
    projectID,
    projectKey,
    projectName,
    repositoryID,
    repositoryKey,
    repositorySlug,
    platformRepoCount: Array.isArray(platformRepos.body?.data) ? platformRepos.body.data.length : 0,
    projectRepoCount: Array.isArray(projectRepos.body?.data) ? projectRepos.body.data.length : 0,
  };
}
