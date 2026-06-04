import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("IntegrationService", () => {
  const apiClientMock = vi.fn();

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({ apiClient: apiClientMock }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("listProviders", () => {
    it("issues GET without query when no params given", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.listProviders();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/integration/providers");
      expect(result.data).toEqual([]);
      expect(result.total).toBe(0);
    });

    it("encodes provider_type / enabled / limit query parameters", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });
      const { integrationService } = await import("./integration.service");

      await integrationService.listProviders({ provider_type: "scm", enabled: true, limit: 10 });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.pathname).toBe("/api/v1/integration/providers");
      expect(url.searchParams.get("provider_type")).toBe("scm");
      expect(url.searchParams.get("enabled")).toBe("true");
      expect(url.searchParams.get("limit")).toBe("10");
    });

    it("falls back to data.length when meta.total is missing", async () => {
      apiClientMock.mockResolvedValue({
        data: [{ provider_id: "p1" }, { provider_id: "p2" }],
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.listProviders();

      expect(result.total).toBe(2);
    });

    it("encodes enabled=false explicitly (not skipped)", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });
      const { integrationService } = await import("./integration.service");

      await integrationService.listProviders({ enabled: false });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.searchParams.get("enabled")).toBe("false");
    });
  });

  describe("createProvider", () => {
    it("issues POST and unwraps data envelope", async () => {
      const provider = { provider_id: "p-new", provider_key: "jira" };
      apiClientMock.mockResolvedValue({ data: provider });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.createProvider({
        provider_key: "jira",
        provider_type: "alm",
        display_name: "Jira",
        auth_mode: "token",
        credentials_ref: "vault://...",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/providers",
        expect.objectContaining({ provider_key: "jira", auth_mode: "token" }),
      );
      expect(result).toEqual(provider);
    });
  });

  describe("updateProvider", () => {
    it("issues PATCH with provider id segment and returns unwrapped data", async () => {
      apiClientMock.mockResolvedValue({ data: { provider_id: "p-1", enabled: false } });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.updateProvider("p-1", { enabled: false });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/integration/providers/p-1",
        { enabled: false },
      );
      expect(result.enabled).toBe(false);
    });
  });

  describe("syncProvider", () => {
    it("issues POST to sync endpoint and returns raw response (no envelope unwrap)", async () => {
      apiClientMock.mockResolvedValue({ status: "accepted", job_id: "job-1" });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.syncProvider("p-1");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/providers/p-1/sync",
      );
      expect(result).toEqual({ status: "accepted", job_id: "job-1" });
    });
  });

  describe("testConnection", () => {
    it("issues POST with base_url body to test-connection endpoint", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        reachable: true,
        status_code: 200,
        latency_ms: 42,
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.testConnection("https://gitea.example.com");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/test-connection",
        { base_url: "https://gitea.example.com" },
      );
      expect(result.reachable).toBe(true);
      expect(result.status_code).toBe(200);
    });

    it("returns reachable=false envelope with error field", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        reachable: false,
        error: "timeout",
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.testConnection("https://bad.example");

      expect(result.reachable).toBe(false);
      expect(result.error).toBe("timeout");
    });
  });

  describe("listScmRepositories", () => {
    it("issues GET to provider scm-repositories endpoint and unwraps data", async () => {
      apiClientMock.mockResolvedValue({
        data: [
          {
            full_name: "org/repo-a",
            name: "repo-a",
            clone_url: "https://example/org/repo-a.git",
            html_url: "https://example/org/repo-a",
            default_branch: "main",
            private: false,
            imported: false,
          },
        ],
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.listScmRepositories("p-1");

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/integration/providers/p-1/scm-repositories",
      );
      expect(result).toHaveLength(1);
      expect(result[0].full_name).toBe("org/repo-a");
    });
  });

  describe("importScmRepositories", () => {
    it("issues POST with full_names body and returns raw envelope", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        imported: 2,
        repositories: [
          { full_name: "org/r1", name: "r1" },
          { full_name: "org/r2", name: "r2" },
        ],
        not_found: [],
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.importScmRepositories("p-1", [
        "org/r1",
        "org/r2",
      ]);

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/providers/p-1/import-repositories",
        { full_names: ["org/r1", "org/r2"] },
      );
      expect(result.imported).toBe(2);
      expect(result.repositories).toHaveLength(2);
    });
  });

  describe("createScmRepository", () => {
    it("issues POST with input body to create-repository endpoint", async () => {
      apiClientMock.mockResolvedValue({
        status: "ok",
        repository: {
          full_name: "org/new-repo",
          name: "new-repo",
          clone_url: "https://example/org/new-repo.git",
          html_url: "https://example/org/new-repo",
          default_branch: "main",
          private: false,
          imported: true,
          source: "system",
        },
      });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.createScmRepository("p-1", {
        name: "new-repo",
        owner: "org",
        description: "Hello",
        private: false,
        auto_init: true,
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/providers/p-1/create-repository",
        {
          name: "new-repo",
          owner: "org",
          description: "Hello",
          private: false,
          auto_init: true,
        },
      );
      expect(result.repository.source).toBe("system");
    });
  });

  describe("deleteProvider", () => {
    it("issues DELETE to provider id endpoint and resolves void", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { integrationService } = await import("./integration.service");

      await expect(integrationService.deleteProvider("p-1")).resolves.toBeUndefined();
      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        "/api/v1/integration/providers/p-1",
      );
    });
  });

  describe("listBindings", () => {
    it("issues GET without query when no params given", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.listBindings();

      expect(apiClientMock).toHaveBeenCalledWith("GET", "/api/v1/integration/bindings");
      expect(result.total).toBe(0);
    });

    it("encodes all supported filter params", async () => {
      apiClientMock.mockResolvedValue({ data: [], meta: { total: 0 } });
      const { integrationService } = await import("./integration.service");

      await integrationService.listBindings({
        scope_type: "platform",
        scope_id: "app-1",
        provider_type: "alm",
        enabled: true,
        limit: 25,
        offset: 50,
      });

      const [, path] = apiClientMock.mock.calls[0];
      const url = new URL(path as string, "http://example");
      expect(url.pathname).toBe("/api/v1/integration/bindings");
      expect(url.searchParams.get("scope_type")).toBe("platform");
      expect(url.searchParams.get("scope_id")).toBe("app-1");
      expect(url.searchParams.get("provider_type")).toBe("alm");
      expect(url.searchParams.get("enabled")).toBe("true");
      expect(url.searchParams.get("limit")).toBe("25");
      expect(url.searchParams.get("offset")).toBe("50");
    });

    it("falls back to data.length when meta.total is missing", async () => {
      apiClientMock.mockResolvedValue({ data: [{ binding_id: "b-1" }] });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.listBindings();

      expect(result.total).toBe(1);
    });
  });

  describe("createBinding", () => {
    it("issues POST and unwraps data envelope", async () => {
      const binding = { binding_id: "b-new", scope_type: "platform" };
      apiClientMock.mockResolvedValue({ data: binding });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.createBinding({
        scope_type: "platform",
        scope_id: "app-1",
        provider_id: "p-1",
        external_key: "PROJ",
        policy: "summary_only",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/integration/bindings",
        expect.objectContaining({ scope_type: "platform", policy: "summary_only" }),
      );
      expect(result).toEqual(binding);
    });
  });

  describe("updateBinding", () => {
    it("issues PATCH with binding id segment and returns unwrapped data", async () => {
      apiClientMock.mockResolvedValue({ data: { binding_id: "b-1", enabled: false } });
      const { integrationService } = await import("./integration.service");

      const result = await integrationService.updateBinding("b-1", { enabled: false });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/integration/bindings/b-1",
        { enabled: false },
      );
      expect(result.enabled).toBe(false);
    });
  });

  describe("deleteBinding", () => {
    it("issues DELETE to binding id endpoint and resolves void", async () => {
      apiClientMock.mockResolvedValue({ status: "ok" });
      const { integrationService } = await import("./integration.service");

      await expect(integrationService.deleteBinding("b-1")).resolves.toBeUndefined();
      expect(apiClientMock).toHaveBeenCalledWith(
        "DELETE",
        "/api/v1/integration/bindings/b-1",
      );
    });
  });
});
