import { render, screen, waitFor } from "@testing-library/react";
import { vi, beforeEach, describe, it, expect } from "vitest";

// next/navigation 은 hook 형태로 사용되므로 vi.fn() 의 반환값을 테스트별로 재설정한다.
const routerReplace = vi.fn();
const pathnameRef = { current: "/" };
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: routerReplace, push: vi.fn() }),
  usePathname: () => pathnameRef.current,
}));

const storeState = {
  actor: null as null | Record<string, unknown>,
  setActor: vi.fn((next: Record<string, unknown>) => {
    storeState.actor = next;
  }),
  clearActor: vi.fn(),
  addToast: vi.fn(),
  incrementNotifications: vi.fn(),
};
vi.mock("@/lib/store", () => ({
  useStore: () => storeState,
}));

vi.mock("@/lib/services/websocket.service", () => ({
  websocketService: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
  },
}));

const whoAmIMock = vi.fn();
vi.mock("@/lib/services/identity.service", () => ({
  identityService: {
    whoAmI: () => whoAmIMock(),
  },
}));

vi.mock("@/lib/services/api-client", () => ({
  ApiError: class extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
    }
  },
}));

const skipFlag = { value: false };
vi.mock("@/lib/storage/onboardingSkip", () => ({
  isOnboardingSkipped: () => skipFlag.value,
  markOnboardingSkipped: vi.fn(),
  clearOnboardingSkip: vi.fn(),
}));

import { AuthGuard } from "./AuthGuard";

function buildResolvedActor(overrides: Record<string, unknown> = {}) {
  return {
    login: "alice",
    subject: "alice@example.com",
    role: "Developer",
    source: "keycloak",
    display_name: "Alice",
    email: "alice@example.com",
    primary_unit_id: null,
    onboarding_required: false,
    onboarding_completed_at: null,
    review_status: null,
    ...overrides,
  };
}

describe("AuthGuard (smoke)", () => {
  beforeEach(() => {
    routerReplace.mockClear();
    pathnameRef.current = "/";
    storeState.actor = null;
    storeState.setActor.mockClear();
    whoAmIMock.mockReset();
    skipFlag.value = false;
  });

  it("renders the loading state while whoAmI is in-flight", () => {
    whoAmIMock.mockReturnValue(new Promise(() => {}));
    render(
      <AuthGuard>
        <div data-testid="protected">secret</div>
      </AuthGuard>
    );

    expect(screen.getByText(/Verifying Identity/i)).toBeInTheDocument();
    expect(screen.queryByTestId("protected")).toBeNull();
  });
});

describe("AuthGuard limited-mode (skip) gating", () => {
  beforeEach(() => {
    routerReplace.mockClear();
    storeState.actor = null;
    storeState.setActor.mockClear();
    whoAmIMock.mockReset();
    skipFlag.value = false;
  });

  // 회귀 가드: PR #288 의 whitelist (["/onboarding", "/auth/"]) 가 default landing
  // (/developer, /manager) 까지 차단해 skip 직후 무한 redirect 루프가 발생했었다.
  // 본 케이스는 skip 사용자가 /developer 진입 시 통과해야 함을 고정한다.
  it("skip 사용자가 default landing (/developer) 으로 진입하면 redirect 가 일어나지 않는다", async () => {
    pathnameRef.current = "/developer";
    skipFlag.value = true;
    whoAmIMock.mockResolvedValue(buildResolvedActor({ onboarding_required: true }));

    render(
      <AuthGuard>
        <div data-testid="protected">child</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(screen.getByTestId("protected")).toBeInTheDocument();
    });
    expect(routerReplace).not.toHaveBeenCalled();
  });

  it("skip 사용자가 /account 진입 시 /onboarding 으로 hard redirect 된다 (TC-ONBOARD-SKIP-PROTECTED-01)", async () => {
    pathnameRef.current = "/account";
    skipFlag.value = true;
    whoAmIMock.mockResolvedValue(buildResolvedActor({ onboarding_required: true }));

    render(
      <AuthGuard>
        <div data-testid="protected">child</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(routerReplace).toHaveBeenCalledWith("/onboarding");
    });
  });

  it("skip 미실행 사용자는 /developer 진입 시에도 /onboarding 으로 redirect 된다", async () => {
    pathnameRef.current = "/developer";
    skipFlag.value = false;
    whoAmIMock.mockResolvedValue(buildResolvedActor({ onboarding_required: true }));

    render(
      <AuthGuard>
        <div data-testid="protected">child</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(routerReplace).toHaveBeenCalledWith("/onboarding");
    });
  });

  it("onboarding 완료 사용자는 자유롭게 통과한다", async () => {
    pathnameRef.current = "/developer";
    skipFlag.value = false;
    whoAmIMock.mockResolvedValue(buildResolvedActor({ onboarding_required: false }));

    render(
      <AuthGuard>
        <div data-testid="protected">child</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(screen.getByTestId("protected")).toBeInTheDocument();
    });
    expect(routerReplace).not.toHaveBeenCalled();
  });
});
