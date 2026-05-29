import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("OnboardingService", () => {
  const apiClientMock = vi.fn();

  class MockApiError extends Error {
    status: number;
    payload: unknown;
    constructor(status: number, payload: unknown, message: string) {
      super(message);
      this.name = "ApiError";
      this.status = status;
      this.payload = payload;
    }
  }

  beforeEach(() => {
    vi.resetModules();
    apiClientMock.mockReset();
    vi.doMock("@/shared/api/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("submit", () => {
    it("POSTs payload to /api/v1/me/onboarding and returns data", async () => {
      const me = { user_id: "u1", display_name: "Alice", primary_unit_id: "team-a" };
      apiClientMock.mockResolvedValue({ data: me });

      const { onboardingService } = await import("./onboarding.service");
      const result = await onboardingService.submit({
        display_name: "Alice",
        primary_unit_id: "team-a",
      });

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/me/onboarding",
        { display_name: "Alice", primary_unit_id: "team-a" },
      );
      expect(result).toEqual(me);
    });

    it("throws ApiError(500) when response data is missing", async () => {
      apiClientMock.mockResolvedValue({ data: undefined });

      const { onboardingService } = await import("./onboarding.service");
      await expect(
        onboardingService.submit({ display_name: "x", primary_unit_id: "y" }),
      ).rejects.toMatchObject({ status: 500, message: "missing onboarding payload" });
    });
  });

  describe("patchMe", () => {
    it("PATCHes only the fields that were provided", async () => {
      const me = { user_id: "u1", display_name: "Bob" };
      apiClientMock.mockResolvedValue({ data: me });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.patchMe({ display_name: "Bob" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/me",
        { display_name: "Bob" },
      );
    });

    it("PATCHes both fields when both provided", async () => {
      apiClientMock.mockResolvedValue({ data: { user_id: "u1" } });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.patchMe({ display_name: "Carol", primary_unit_id: "team-c" });

      expect(apiClientMock).toHaveBeenCalledWith(
        "PATCH",
        "/api/v1/me",
        { display_name: "Carol", primary_unit_id: "team-c" },
      );
    });

    it("sends empty body when no fields are provided", async () => {
      apiClientMock.mockResolvedValue({ data: { user_id: "u1" } });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.patchMe({});

      expect(apiClientMock).toHaveBeenCalledWith("PATCH", "/api/v1/me", {});
    });

    it("throws ApiError(500) when response data is missing", async () => {
      apiClientMock.mockResolvedValue({ data: null });

      const { onboardingService } = await import("./onboarding.service");
      await expect(onboardingService.patchMe({ display_name: "x" })).rejects.toMatchObject({
        status: 500,
        message: "missing me payload",
      });
    });
  });

  describe("searchOrganizations", () => {
    it("issues GET with q + default limit=20", async () => {
      apiClientMock.mockResolvedValue({ data: [{ unit_id: "u1", name: "Team A" }] });

      const { onboardingService } = await import("./onboarding.service");
      const result = await onboardingService.searchOrganizations("team");

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/organizations/search?q=team&limit=20",
      );
      expect(result).toEqual([{ unit_id: "u1", name: "Team A" }]);
    });

    it("respects custom limit", async () => {
      apiClientMock.mockResolvedValue({ data: [] });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.searchOrganizations("q", 5);

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        "/api/v1/organizations/search?q=q&limit=5",
      );
    });

    it("URL-encodes the q parameter", async () => {
      apiClientMock.mockResolvedValue({ data: [] });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.searchOrganizations("팀 A&B", 10);

      expect(apiClientMock).toHaveBeenCalledWith(
        "GET",
        expect.stringContaining("q=%ED%8C%80+A%26B"),
      );
    });

    it("defaults to [] when API returns null data", async () => {
      apiClientMock.mockResolvedValue({ data: null });

      const { onboardingService } = await import("./onboarding.service");
      const result = await onboardingService.searchOrganizations("x");

      expect(result).toEqual([]);
    });
  });

  describe("confirmUserReview", () => {
    it("POSTs to /api/v1/users/:id/review and returns review payload", async () => {
      const review = {
        user_id: "u-1",
        review_status: "approved",
        reviewed_at: "2026-05-28T00:00:00Z",
        reviewed_by: "admin",
      };
      apiClientMock.mockResolvedValue({ data: review });

      const { onboardingService } = await import("./onboarding.service");
      const result = await onboardingService.confirmUserReview("u-1");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/users/u-1/review",
        {},
      );
      expect(result).toEqual(review);
    });

    it("URL-encodes user id with special characters", async () => {
      apiClientMock.mockResolvedValue({ data: { user_id: "x/y", review_status: "approved", reviewed_at: "", reviewed_by: "" } });

      const { onboardingService } = await import("./onboarding.service");
      await onboardingService.confirmUserReview("x/y");

      expect(apiClientMock).toHaveBeenCalledWith(
        "POST",
        "/api/v1/users/x%2Fy/review",
        {},
      );
    });

    it("throws ApiError(500) when response data is missing", async () => {
      apiClientMock.mockResolvedValue({ data: undefined });

      const { onboardingService } = await import("./onboarding.service");
      await expect(onboardingService.confirmUserReview("u-1")).rejects.toMatchObject({
        status: 500,
        message: "missing review payload",
      });
    });
  });

  describe("singleton", () => {
    it("getInstance returns the same instance on repeated calls", async () => {
      const { OnboardingService } = await import("./onboarding.service");
      const a = OnboardingService.getInstance();
      const b = OnboardingService.getInstance();
      expect(a).toBe(b);
    });
  });
});
