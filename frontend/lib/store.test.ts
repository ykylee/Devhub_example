import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { useStore, AuthenticatedActor } from "./store";

describe("Zustand Global Store (useStore)", () => {
  beforeEach(() => {
    // 매 테스트 시작 전 스토어 상태를 초기화합니다.
    useStore.setState({
      role: null,
      actor: null,
      isDeepFocus: false,
      notifications: 3,
      toasts: [],
      isLoggingOut: false,
      isSidebarOpen: false,
      isSidebarCollapsed: false,
    });
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("초기 기본 상태가 설계 스펙과 일치해야 합니다", () => {
    const state = useStore.getState();
    expect(state.role).toBeNull();
    expect(state.actor).toBeNull();
    expect(state.isDeepFocus).toBe(false);
    expect(state.notifications).toBe(3);
    expect(state.toasts).toEqual([]);
    expect(state.isLoggingOut).toBe(false);
    expect(state.isSidebarOpen).toBe(false);
    expect(state.isSidebarCollapsed).toBe(false);
  });

  it("setActor 및 clearActor가 정상적으로 상태를 갱신해야 합니다", () => {
    const mockActor: AuthenticatedActor = {
      login: "testuser",
      user_id: "usr-12345",
      subject: "sub-99999",
      role: "Developer",
      display_name: "Test User",
      email: "test@example.com",
    };

    // 1. setActor
    useStore.getState().setActor(mockActor);
    expect(useStore.getState().actor).toEqual(mockActor);
    expect(useStore.getState().role).toBe("Developer");

    // 2. clearActor
    useStore.getState().clearActor();
    expect(useStore.getState().actor).toBeNull();
    expect(useStore.getState().role).toBeNull();
  });

  // UT-store-01 (2026-06-08 PR #497 hotfix #2):
  // logout 502 (Keycloak unreachable) 분기에서 backend 가 access token revoke
  // 를 못 하는 환경 (CI E2E container Keycloak flake 등) 에서, AuthGuard 의
  // pathname 변화 반응 useEffect 가 logout 진행 중에 whoAmI() 를 재호출해
  // stale session (revoke 안 된 token) 을 actor store 에 다시 박는 race.
  // spec #488 정합을 위해 setActor 는 isLoggingOut=true 동안 no-op 처리.
  it("setActor는 isLoggingOut=true 동안 외부 호출이 와도 stale actor를 박지 않아야 합니다 (logout 502 분기 race 차단)", () => {
    const initialActor: AuthenticatedActor = {
      login: "alice",
      user_id: "usr-alice",
      subject: "sub-alice",
      role: "Developer",
    };
    const staleActor: AuthenticatedActor = {
      login: "alice",
      user_id: "usr-alice",
      subject: "sub-alice",
      role: "Developer",
      display_name: "stale-actor (revoke 안 된 backend session)",
    };

    // 1. 정상 로그인 → actor 박힘
    useStore.getState().setActor(initialActor);
    expect(useStore.getState().actor?.display_name).toBeUndefined();

    // 2. logout 시작 — clearActor + isLoggingOut=true (auth.service.ts:206-207 시퀀스)
    useStore.getState().setIsLoggingOut(true);
    useStore.getState().clearActor();
    expect(useStore.getState().actor).toBeNull();
    expect(useStore.getState().isLoggingOut).toBe(true);

    // 3. logout 진행 중 AuthGuard 가 pathname 변경에 반응해 whoAmI() 재호출,
    //    backend 가 502 + revoke 실패 → 200 으로 stale 응답. setActor 시도:
    useStore.getState().setActor(staleActor);

    // 4. spec #488 정합: isLoggingOut=true 동안 setActor 는 no-op 이므로
    //    stale actor 가 박히지 않아야 함. window.location.assign('/login') 가
    //    발사한 후에도 AuthGuard 가 stale actor 를 못 보고 /developer 로
    //    redirect 시키는 deadlock 가 풀린다.
    expect(useStore.getState().actor).toBeNull();

    // 5. logout 완료 → isLoggingOut=false (실제 login 페이지 도착 시점에
    //    overlay 가 풀리며 setIsLoggingOut(false) 호출된다고 가정).
    useStore.getState().setIsLoggingOut(false);

    // 6. 다음 login-success path 에서는 setActor 가 정상 동작.
    useStore.getState().setActor(staleActor);
    expect(useStore.getState().actor).toEqual(staleActor);
    expect(useStore.getState().actor?.display_name).toBe("stale-actor (revoke 안 된 backend session)");
  });

  it("setRole이 정상적으로 롤(Role) 상태를 직접 갱신해야 합니다", () => {
    useStore.getState().setRole("Manager");
    expect(useStore.getState().role).toBe("Manager");

    useStore.getState().setRole(null);
    expect(useStore.getState().role).toBeNull();
  });

  it("setDeepFocus가 딥 포커스 상태를 정상적으로 제어해야 합니다", () => {
    useStore.getState().setDeepFocus(true);
    expect(useStore.getState().isDeepFocus).toBe(true);

    useStore.getState().setDeepFocus(false);
    expect(useStore.getState().isDeepFocus).toBe(false);
  });

  it("알림 개수를 초기화하거나(clear) 1씩 증가(increment)시킬 수 있어야 합니다", () => {
    // 1. increment
    useStore.getState().incrementNotifications();
    expect(useStore.getState().notifications).toBe(4);

    useStore.getState().incrementNotifications();
    expect(useStore.getState().notifications).toBe(5);

    // 2. clear
    useStore.getState().clearNotifications();
    expect(useStore.getState().notifications).toBe(0);
  });

  it("addToast가 새로운 토스트를 추가하고, 5초 후 자동으로 제거해야 합니다", () => {
    useStore.getState().addToast("Welcome to DevHub!", "success");
    
    let toasts = useStore.getState().toasts;
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe("Welcome to DevHub!");
    expect(toasts[0].type).toBe("success");
    expect(toasts[0].id).toBeDefined();

    // 5초 미만 대기 시 토스트는 그대로 유지되어야 합니다.
    vi.advanceTimersByTime(4900);
    expect(useStore.getState().toasts.length).toBe(1);

    // 5초 도달 시 자동 제거되어야 합니다.
    vi.advanceTimersByTime(100);
    expect(useStore.getState().toasts.length).toBe(0);
  });

  it("removeToast를 통해 특정 토스트를 강제로 즉시 제거할 수 있어야 합니다", () => {
    useStore.getState().addToast("Transient warning", "warning");
    const toastId = useStore.getState().toasts[0].id;
    expect(useStore.getState().toasts.length).toBe(1);

    useStore.getState().removeToast(toastId);
    expect(useStore.getState().toasts.length).toBe(0);
  });

  it("로그아웃 상태(isLoggingOut)를 제어할 수 있어야 합니다", () => {
    useStore.getState().setIsLoggingOut(true);
    expect(useStore.getState().isLoggingOut).toBe(true);

    useStore.getState().setIsLoggingOut(false);
    expect(useStore.getState().isLoggingOut).toBe(false);
  });

  it("사이드바 열림(isSidebarOpen) 및 접힘(isSidebarCollapsed) 상태를 제어해야 합니다", () => {
    // sidebar open
    useStore.getState().setSidebarOpen(true);
    expect(useStore.getState().isSidebarOpen).toBe(true);

    // sidebar collapsed
    useStore.getState().setSidebarCollapsed(true);
    expect(useStore.getState().isSidebarCollapsed).toBe(true);
  });

  it("persist 미들웨어가 영속화 시 isLoggingOut, toasts, isSidebarOpen 상태를 제외해야 합니다 (partialize)", () => {
    const partializeFn = useStore.persist.getOptions().partialize;
    expect(partializeFn).toBeDefined();

    if (partializeFn) {
      const fullState = {
        role: "System Admin" as const,
        actor: { login: "admin", role: "System Admin" as const },
        isDeepFocus: true,
        notifications: 10,
        toasts: [{ id: "t1", message: "Error", type: "error" as const }],
        isLoggingOut: true,
        isSidebarOpen: true,
        isSidebarCollapsed: true,
        setActor: () => {},
        clearActor: () => {},
        setRole: () => {},
        setDeepFocus: () => {},
        clearNotifications: () => {},
        incrementNotifications: () => {},
        addToast: () => {},
        removeToast: () => {},
        setIsLoggingOut: () => {},
        setSidebarOpen: () => {},
        setSidebarCollapsed: () => {},
      };

      const partialized = partializeFn(fullState);
      
      // 제외 대상들이 제대로 빠졌는지 단언합니다.
      expect(partialized).not.toHaveProperty("isLoggingOut");
      expect(partialized).not.toHaveProperty("toasts");
      expect(partialized).not.toHaveProperty("isSidebarOpen");

      // 보존 대상들이 제대로 남아있는지 단언합니다.
      expect(partialized.role).toBe("System Admin");
      expect(partialized.actor).toEqual({ login: "admin", role: "System Admin" });
      expect(partialized.isDeepFocus).toBe(true);
      expect(partialized.notifications).toBe(10);
      expect(partialized.isSidebarCollapsed).toBe(true);
    }
  });
});
