import { render, screen } from "@testing-library/react";
import { vi, beforeEach } from "vitest";

// Mocks for next/navigation + zustand store + downstream services. Kept
// intentionally minimal — full coverage of AuthGuard's role-routing logic
// is carve out (sprint claude/work_260513-i § 미진입). This file pins the
// initial loading state so future regressions of the suspense path
// (Verifying Identity...) are caught.

const routerReplace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: routerReplace, push: vi.fn() }),
  usePathname: () => "/",
}));

const storeState = {
  actor: null,
  setActor: vi.fn(),
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

// identityService.whoAmI returns a pending promise so the loader state
// stays mounted for the assertion. AuthGuard's full happy/sad paths are
// carve out for the follow-up Vitest sprint.
vi.mock("@/lib/services/identity.service", () => ({
  identityService: {
    whoAmI: () => new Promise(() => {}),
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

import { AuthGuard } from "./AuthGuard";

describe("AuthGuard (smoke)", () => {
  beforeEach(() => {
    routerReplace.mockClear();
  });

  it("renders the loading state while whoAmI is in-flight", () => {
    render(
      <AuthGuard>
        <div data-testid="protected">secret</div>
      </AuthGuard>
    );

    // Verifying Identity copy is the public marker of the suspense path.
    expect(screen.getByText(/Verifying Identity/i)).toBeInTheDocument();
    // Children must NOT render while isAuthorized=false.
    expect(screen.queryByTestId("protected")).toBeNull();
  });
});
