/**
 * Last build status → UI 표시 helper.
 *
 * REQ-FR-APPDASH-001 결정 ("단순 빌드 성공률(%)보다 broken/red 상태 즉시 표기") 의
 * frontend 표기 일원화. Application 단의 derive 값 (target_branch_build_status:
 * healthy|broken|unknown) 과 Repository 단의 raw build_runs.status (queued|running|
 * success|failed|cancelled|skipped|unknown) 를 모두 단일 표시 모델로 normalize.
 */

export type LastBuildSurface = "healthy" | "broken" | "unknown" | "pending" | "success" | "failed" | "skipped" | "cancelled";

export interface LastBuildView {
  label: string;
  variant: "success" | "danger" | "warning" | "secondary";
  tone: "positive" | "negative" | "neutral";
}

/**
 * Application 단의 derive 값 (rollup 또는 dashboard 응답의 target_branch_build_status)
 * 을 표시 모델로 변환.
 */
export function applicationBuildStatusView(status: string | null | undefined): LastBuildView {
  switch (status) {
    case "healthy":
      return { label: "Healthy", variant: "success", tone: "positive" };
    case "broken":
      return { label: "Broken", variant: "danger", tone: "negative" };
    default:
      return { label: "없음", variant: "secondary", tone: "neutral" };
  }
}

/**
 * Repository activity 의 last_build_status (build_runs.status enum) 를 표시 모델로 변환.
 */
export function repositoryLastBuildView(status: string | null | undefined): LastBuildView {
  switch (status) {
    case "success":
      return { label: "Success", variant: "success", tone: "positive" };
    case "failed":
      return { label: "Failed", variant: "danger", tone: "negative" };
    case "cancelled":
      return { label: "Cancelled", variant: "warning", tone: "negative" };
    case "skipped":
      return { label: "Skipped", variant: "secondary", tone: "neutral" };
    case "running":
      return { label: "Running", variant: "warning", tone: "neutral" };
    case "queued":
      return { label: "Queued", variant: "secondary", tone: "neutral" };
    default:
      return { label: "없음", variant: "secondary", tone: "neutral" };
  }
}
