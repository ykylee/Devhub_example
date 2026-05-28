import { apiClient, ApiError } from "./api-client";
import type { ApiResponse } from "./wire";
import type { MeResponse } from "./identity.service";

export interface OrgSearchResult {
  unit_id: string;
  name: string;
}

export interface OnboardingSubmitPayload {
  display_name: string;
  primary_unit_id: string;
}

export interface PatchMePayload {
  display_name?: string;
  primary_unit_id?: string;
}

export interface ReviewConfirmResult {
  user_id: string;
  review_status: string;
  reviewed_at: string;
  reviewed_by: string;
}

export class OnboardingService {
  private static instance: OnboardingService;
  private constructor() {}
  public static getInstance(): OnboardingService {
    if (!OnboardingService.instance) {
      OnboardingService.instance = new OnboardingService();
    }
    return OnboardingService.instance;
  }

  async submit(payload: OnboardingSubmitPayload): Promise<MeResponse> {
    const result = await apiClient<ApiResponse<MeResponse>>("POST", `/api/v1/me/onboarding`, payload);
    if (!result.data) throw new ApiError(500, result, "missing onboarding payload");
    return result.data;
  }

  async patchMe(payload: PatchMePayload): Promise<MeResponse> {
    const body: Record<string, string> = {};
    if (payload.display_name !== undefined) body.display_name = payload.display_name;
    if (payload.primary_unit_id !== undefined) body.primary_unit_id = payload.primary_unit_id;
    const result = await apiClient<ApiResponse<MeResponse>>("PATCH", `/api/v1/me`, body);
    if (!result.data) throw new ApiError(500, result, "missing me payload");
    return result.data;
  }

  async searchOrganizations(q: string, limit = 20): Promise<OrgSearchResult[]> {
    const qs = new URLSearchParams({ q, limit: String(limit) }).toString();
    const result = await apiClient<ApiResponse<OrgSearchResult[]>>("GET", `/api/v1/organizations/search?${qs}`);
    return result.data ?? [];
  }

  async confirmUserReview(userId: string): Promise<ReviewConfirmResult> {
    const result = await apiClient<ApiResponse<ReviewConfirmResult>>(
      "POST",
      `/api/v1/users/${encodeURIComponent(userId)}/review`,
      {},
    );
    if (!result.data) throw new ApiError(500, result, "missing review payload");
    return result.data;
  }
}

export const onboardingService = OnboardingService.getInstance();
