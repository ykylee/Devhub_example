import { act } from "react";
import { describe, expect, it, vi } from "vitest";
import { useStore, AuthenticatedActor } from "./store";

describe("AppState Global Store", () => {
  it("manages authenticated actor state", () => {
    const actor: AuthenticatedActor = {
      login: "test-user",
      role: "Developer",
    };

    act(() => {
      useStore.getState().setActor(actor);
    });

    expect(useStore.getState().actor).toEqual(actor);
    expect(useStore.getState().role).toBe("Developer");

    act(() => {
      useStore.getState().clearActor();
    });

    expect(useStore.getState().actor).toBeNull();
    expect(useStore.getState().role).toBeNull();
  });

  it("manages custom roles and deep focus", () => {
    act(() => {
      useStore.getState().setRole("Manager");
    });
    expect(useStore.getState().role).toBe("Manager");

    act(() => {
      useStore.getState().setDeepFocus(true);
    });
    expect(useStore.getState().isDeepFocus).toBe(true);
  });

  it("manages notifications increment and clear", () => {
    act(() => {
      useStore.setState({ notifications: 3 });
    });

    act(() => {
      useStore.getState().incrementNotifications();
    });
    expect(useStore.getState().notifications).toBe(4);

    act(() => {
      useStore.getState().clearNotifications();
    });
    expect(useStore.getState().notifications).toBe(0);
  });

  it("adds and removes toasts with automatic timeout", () => {
    vi.useFakeTimers();

    act(() => {
      useStore.getState().addToast("hello", "success");
    });

    const toasts = useStore.getState().toasts;
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe("hello");
    expect(toasts[0].type).toBe("success");

    // Timeout execution cover (Timeout callback covered!)
    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(useStore.getState().toasts.length).toBe(0);

    // Manual removal
    act(() => {
      useStore.getState().addToast("manual");
    });
    const newToasts = useStore.getState().toasts;
    expect(newToasts.length).toBe(1);
    
    act(() => {
      useStore.getState().removeToast(newToasts[0].id);
    });
    expect(useStore.getState().toasts.length).toBe(0);

    vi.useRealTimers();
  });

  it("manages logging out and sidebar collapse states", () => {
    act(() => {
      useStore.getState().setIsLoggingOut(true);
      useStore.getState().setSidebarOpen(true);
      useStore.getState().setSidebarCollapsed(true);
    });

    expect(useStore.getState().isLoggingOut).toBe(true);
    expect(useStore.getState().isSidebarOpen).toBe(true);
    expect(useStore.getState().isSidebarCollapsed).toBe(true);
  });
});
