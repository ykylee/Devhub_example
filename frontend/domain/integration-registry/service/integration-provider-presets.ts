import type { IntegrationProviderType, IntegrationAuthMode } from "@/lib/services/integration.types";

// ADR-0024 외부 연동 등록 UX 고도화 — vendor 템플릿 + 가이드 자격증명 입력.
// credentials_ref 의 내부 인코딩(hmac_sha256:/provider_sdk:vendor:)을 사용자가 직접
// 타이핑하지 않도록, 전략 + secret 을 분리 입력받아 조합한다.
// 백엔드 verifyIntegrationWebhookSignature (integration_registry.go) 의 3 전략과 정합.

export type WebhookSignatureStrategy = "hmac_sha256" | "provider_sdk" | "shared_token";

// 백엔드 verifyProviderSDKWebhookSignature 가 인식하는 vendor (integration_registry.go).
export const SDK_VENDORS = [
  "gitea",
  "forgejo",
  "github",
  "gitlab",
  "jira",
  "bitbucket",
  "jenkins",
  "bamboo",
] as const;
export type SdkVendor = (typeof SDK_VENDORS)[number];

// 등록 UI 에서 체크박스로 노출하는 표준 capability 어휘 (백엔드는 자유 JSONB 이나
// 운영 일관성을 위해 큐레이션).
export const KNOWN_CAPABILITIES = ["webhook", "pull", "push", "snapshot", "sync"] as const;

export interface VendorPreset {
  id: string;
  label: string;
  providerType: IntegrationProviderType;
  authMode: IntegrationAuthMode;
  signatureStrategy: WebhookSignatureStrategy;
  sdkVendor?: SdkVendor;
  capabilities: string[];
  providerKeyHint: string;
  baseUrlHint?: string;
}

export const VENDOR_PRESETS: VendorPreset[] = [
  {
    id: "custom",
    label: "Custom (manual)",
    providerType: "scm",
    authMode: "token",
    signatureStrategy: "hmac_sha256",
    capabilities: [],
    providerKeyHint: "my_provider",
  },
  {
    id: "gitea",
    label: "Gitea",
    providerType: "scm",
    authMode: "token",
    signatureStrategy: "provider_sdk",
    sdkVendor: "gitea",
    capabilities: ["webhook", "pull"],
    providerKeyHint: "gitea_main",
    baseUrlHint: "https://gitea.example.com",
  },
  {
    id: "github",
    label: "GitHub",
    providerType: "scm",
    authMode: "token",
    signatureStrategy: "provider_sdk",
    sdkVendor: "github",
    capabilities: ["webhook", "pull"],
    providerKeyHint: "github_org",
    baseUrlHint: "https://api.github.com",
  },
  {
    id: "gitlab",
    label: "GitLab",
    providerType: "scm",
    authMode: "token",
    signatureStrategy: "provider_sdk",
    sdkVendor: "gitlab",
    capabilities: ["webhook", "pull"],
    providerKeyHint: "gitlab_main",
  },
  {
    id: "bitbucket",
    label: "Bitbucket",
    providerType: "scm",
    authMode: "app_password",
    signatureStrategy: "provider_sdk",
    sdkVendor: "bitbucket",
    capabilities: ["webhook"],
    providerKeyHint: "bitbucket_main",
  },
  {
    id: "jira",
    label: "Jira",
    providerType: "alm",
    authMode: "oauth2",
    signatureStrategy: "provider_sdk",
    sdkVendor: "jira",
    capabilities: ["webhook"],
    providerKeyHint: "jira_cloud",
    baseUrlHint: "https://your-org.atlassian.net",
  },
  {
    id: "jenkins",
    label: "Jenkins",
    providerType: "ci_cd",
    authMode: "token",
    signatureStrategy: "provider_sdk",
    sdkVendor: "jenkins",
    capabilities: ["webhook"],
    providerKeyHint: "jenkins_prod",
    baseUrlHint: "https://jenkins.example.com",
  },
  {
    id: "bamboo",
    label: "Bamboo",
    providerType: "ci_cd",
    authMode: "token",
    signatureStrategy: "provider_sdk",
    sdkVendor: "bamboo",
    capabilities: ["webhook"],
    providerKeyHint: "bamboo_ci",
  },
];

export function getVendorPreset(id: string): VendorPreset {
  return VENDOR_PRESETS.find((p) => p.id === id) ?? VENDOR_PRESETS[0];
}

export function composeCredentialsRef(
  strategy: WebhookSignatureStrategy,
  sdkVendor: SdkVendor | undefined,
  secret: string,
): string {
  const s = secret.trim();
  switch (strategy) {
    case "hmac_sha256":
      return `hmac_sha256:${s}`;
    case "provider_sdk":
      return `provider_sdk:${sdkVendor ?? "gitea"}:${s}`;
    case "shared_token":
    default:
      return s;
  }
}

export interface ParsedCredentials {
  strategy: WebhookSignatureStrategy;
  sdkVendor?: SdkVendor;
  hasSecret: boolean;
}

export function parseCredentialsRef(ref: string): ParsedCredentials {
  const v = (ref ?? "").trim();
  if (v.startsWith("hmac_sha256:")) {
    return { strategy: "hmac_sha256", hasSecret: v.slice("hmac_sha256:".length).length > 0 };
  }
  if (v.startsWith("provider_sdk:")) {
    const rest = v.slice("provider_sdk:".length);
    const firstColon = rest.indexOf(":");
    const vendor = firstColon >= 0 ? rest.slice(0, firstColon) : rest;
    const secret = firstColon >= 0 ? rest.slice(firstColon + 1) : "";
    const sv = (SDK_VENDORS as readonly string[]).includes(vendor) ? (vendor as SdkVendor) : undefined;
    return { strategy: "provider_sdk", sdkVendor: sv, hasSecret: secret.length > 0 };
  }
  return { strategy: "shared_token", hasSecret: v.length > 0 };
}
