import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { WSEvent } from "@/lib/services/types";

// realtime.service.ts 는 module-load 시 RealtimeService.getInstance() 호출
// → init() → connect() → buildURL() → fetchTicket() (apiClient 호출) → new WebSocket(url)
// 흐름이 발생하므로 모든 의존성 (WebSocket / tokenStore / useStore / apiClient) 을
// import 이전에 stub 한다.

interface MockWsInstance {
  url: string;
  readyState: number;
  onopen: ((event: Record<string, unknown>) => void) | null;
  onclose: ((event: Record<string, unknown>) => void) | null;
  onmessage: ((event: { data: string }) => void) | null;
  onerror: ((event: Record<string, unknown>) => void) | null;
  send: ReturnType<typeof vi.fn>;
  close: (code?: number, reason?: string) => void;
  simulateOpen: () => void;
  receiveMessage: (data: string) => void;
}

let mockWsInstances: MockWsInstance[] = [];

class MockWebSocket {
  static CONNECTING = 0;
  static OPEN = 1;
  static CLOSING = 2;
  static CLOSED = 3;

  url: string;
  readyState = MockWebSocket.CONNECTING;
  onopen: ((event: Record<string, unknown>) => void) | null = null;
  onclose: ((event: Record<string, unknown>) => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onerror: ((event: Record<string, unknown>) => void) | null = null;
  send = vi.fn();

  constructor(url: string) {
    this.url = url;
    mockWsInstances.push(this as unknown as MockWsInstance);
  }

  close(code = 1000) {
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) this.onclose({ code, reason: "test", wasClean: code === 1000 });
  }

  simulateOpen() {
    this.readyState = MockWebSocket.OPEN;
    if (this.onopen) this.onopen({});
  }

  receiveMessage(data: string) {
    if (this.onmessage) this.onmessage({ data });
  }
}

describe("RealtimeService", () => {
  const tokenStoreMock = { getAccessToken: vi.fn<() => string | null>() };
  const apiClientMock = vi.fn();

  // useStore mock — Zustand-like. subscribe 콜백을 트래킹해 identity 변경 시뮬레이션.
  type Snapshot = { actor: { login: string } | null; role: string | null };
  let currentSnapshot: Snapshot = { actor: { login: "alice" }, role: "Developer" };
  let subscribeListeners: Array<{
    selector: (s: Snapshot) => unknown;
    cb: (current: unknown, previous: unknown) => void;
  }> = [];
  const useStoreMock = {
    getState: () => currentSnapshot,
    subscribe: vi.fn((selector: (s: Snapshot) => unknown, cb: (c: unknown, p: unknown) => void) => {
      subscribeListeners.push({ selector, cb });
      return () => {
        subscribeListeners = subscribeListeners.filter((l) => l.selector !== selector);
      };
    }),
  };

  // ApiError mock
  class MockApiError extends Error {
    status: number;
    payload: unknown;
    constructor(status: number, payload: unknown, message: string) {
      super(message);
      this.status = status;
      this.payload = payload;
    }
  }

  beforeEach(() => {
    vi.resetModules();
    mockWsInstances = [];
    subscribeListeners = [];
    currentSnapshot = { actor: { login: "alice" }, role: "Developer" };
    tokenStoreMock.getAccessToken.mockReset();
    apiClientMock.mockReset();
    useStoreMock.subscribe.mockClear();

    tokenStoreMock.getAccessToken.mockReturnValue("access-token-1");
    apiClientMock.mockResolvedValue({ ticket: "ticket-abc" });

    vi.useFakeTimers();
    vi.stubGlobal("WebSocket", MockWebSocket);
    vi.spyOn(console, "log").mockImplementation(() => undefined);
    vi.spyOn(console, "warn").mockImplementation(() => undefined);
    vi.spyOn(console, "error").mockImplementation(() => undefined);

    vi.doMock("@/domain/auth-session/service/token-store", () => ({ tokenStore: tokenStoreMock }));
    vi.doMock("@/lib/services/api-client", () => ({
      apiClient: apiClientMock,
      ApiError: MockApiError,
    }));
    vi.doMock("@/lib/store", () => ({ useStore: useStoreMock }));
    vi.doMock("@/shared/config/endpoints", () => ({
      WS_BASE_URL: "ws://localhost:8080/api/v1/realtime/ws",
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  async function loadService() {
    const mod = await import("./realtime.service");
    // module-load 시 init→connect 시작 (async). 첫 tick 진행으로 fetchTicket await 완료.
    await vi.advanceTimersByTimeAsync(0);
    await Promise.resolve();
    await Promise.resolve();
    return mod;
  }

  describe("singleton + initial connect", () => {
    it("getInstance returns same instance + module export", async () => {
      const mod = await loadService();
      expect(mod.realtimeService).toBe(mod.RealtimeService.getInstance());
    });

    it("initial connect fetches ticket and opens WS with ticket param", async () => {
      await loadService();

      expect(apiClientMock).toHaveBeenCalledWith("POST", "/api/v1/realtime/ticket");
      expect(mockWsInstances).toHaveLength(1);
      expect(mockWsInstances[0].url).toContain("ticket=ticket-abc");
      expect(mockWsInstances[0].url).toContain("types=");
      expect(mockWsInstances[0].url).toContain("actor=alice");
      expect(mockWsInstances[0].url).toContain("role=developer"); // "Developer" → "developer"
    });

    it("skips connect when no access token (auth-dead guard)", async () => {
      tokenStoreMock.getAccessToken.mockReturnValue(null);
      await loadService();

      expect(mockWsInstances).toHaveLength(0);
      // apiClient ticket 도 호출 안 됨.
      expect(apiClientMock).not.toHaveBeenCalled();
    });
  });

  describe("buildURL — role / actor encoding", () => {
    it("maps role 'System Admin' → 'system_admin' wire param", async () => {
      currentSnapshot = { actor: { login: "bob" }, role: "System Admin" };
      await loadService();

      expect(mockWsInstances[0].url).toContain("role=system_admin");
    });

    it("maps role 'Manager' → 'manager'", async () => {
      currentSnapshot = { actor: { login: "bob" }, role: "Manager" };
      await loadService();
      expect(mockWsInstances[0].url).toContain("role=manager");
    });

    it("non-recognized role lowercased fallback", async () => {
      currentSnapshot = { actor: { login: "x" }, role: "CustomRole" };
      await loadService();
      expect(mockWsInstances[0].url).toContain("role=customrole");
    });

    it("actor=guest + role=guest when actor/role both null", async () => {
      currentSnapshot = { actor: null, role: null };
      // tokenStore returns a token, so connect proceeds but builds guest URL.
      await loadService();
      expect(mockWsInstances[0].url).toContain("actor=guest");
      expect(mockWsInstances[0].url).toContain("role=guest");
    });
  });

  describe("fetchTicket failure handling", () => {
    it("connects without ticket when apiClient throws ApiError (401)", async () => {
      apiClientMock.mockRejectedValue(new MockApiError(401, null, "unauth"));
      await loadService();

      // connect 가 ticket 없이 진행 (URL 에 ticket 없음).
      expect(mockWsInstances).toHaveLength(1);
      expect(mockWsInstances[0].url).not.toContain("ticket=");
    });

    it("connects without ticket on network error", async () => {
      apiClientMock.mockRejectedValue(new Error("network down"));
      await loadService();

      expect(mockWsInstances).toHaveLength(1);
      expect(mockWsInstances[0].url).not.toContain("ticket=");
    });
  });

  describe("onopen — heartbeat + stable-open timer + dispatch", () => {
    it("dispatches status.changed (connected=true) and starts heartbeat ticking every 25s", async () => {
      const { realtimeService } = await loadService();
      const statusCb = vi.fn();
      realtimeService.subscribe("status.changed", statusCb);

      mockWsInstances[0].simulateOpen();
      expect(statusCb).toHaveBeenCalledWith(
        expect.objectContaining({ type: "status.changed", data: { connected: true } }),
      );

      // 25s 이후 heartbeat send.
      vi.advanceTimersByTime(25_000);
      expect(mockWsInstances[0].send).toHaveBeenCalledTimes(1);
      const payload = JSON.parse(mockWsInstances[0].send.mock.calls[0][0] as string);
      expect(payload.type).toBe("ping");

      // 50s → 2 calls.
      vi.advanceTimersByTime(25_000);
      expect(mockWsInstances[0].send).toHaveBeenCalledTimes(2);

      realtimeService.unsubscribe("status.changed", statusCb);
    });

    it("heartbeat tolerates send throw without crashing interval", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].send.mockImplementation(() => {
        throw new Error("send blew");
      });
      vi.advanceTimersByTime(25_000);
      // No throw — warn logged, interval continues.
      expect(console.warn).toHaveBeenCalled();
    });

    it("stable-open timer resets reconnectAttempts after stableOpenMs", async () => {
      const { realtimeService } = await loadService();
      // close to bump attempts up.
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close(1006);

      // reconnectAttempts 가 증가했음.
      const svc = realtimeService as { reconnectAttempts: number; stableOpenMs: number };
      expect(svc.reconnectAttempts).toBeGreaterThan(0);

      // 3s 진행 → 새 connect → simulateOpen → stable-open 5분 후 reset.
      await vi.advanceTimersByTimeAsync(3_000);
      // 두 번째 instance 가 생성됐을 수도. 마지막 인스턴스 open.
      const last = mockWsInstances[mockWsInstances.length - 1];
      last.simulateOpen();
      // 5분 + 1s 진행 → stable-open timer fire → reconnectAttempts 0 reset.
      vi.advanceTimersByTime(svc.stableOpenMs + 1000);
      expect(svc.reconnectAttempts).toBe(0);
    });
  });

  describe("onclose — reconnect logic", () => {
    it("non-clean close (code !== 1000) triggers reconnect with exponential backoff", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close(1006);

      // 3s base backoff.
      await vi.advanceTimersByTimeAsync(3_000);
      expect(mockWsInstances).toHaveLength(2);
    });

    it("clean close (code === 1000) does NOT reconnect", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close(1000);

      await vi.advanceTimersByTimeAsync(60_000);
      // No additional WS created.
      expect(mockWsInstances).toHaveLength(1);
    });

    it("dispatches status.changed (connected=false) on close", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn();
      realtimeService.subscribe("status.changed", cb);
      mockWsInstances[0].simulateOpen();
      cb.mockClear();
      mockWsInstances[0].close(1006);
      expect(cb).toHaveBeenCalledWith(
        expect.objectContaining({ type: "status.changed", data: { connected: false } }),
      );
    });

    it("stops heartbeat on close", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close(1006);

      // After close — even if send is still on the (closed) socket, no more heartbeats fire.
      const before = mockWsInstances[0].send.mock.calls.length;
      vi.advanceTimersByTime(60_000);
      expect(mockWsInstances[0].send.mock.calls.length).toBe(before);
    });
  });

  describe("handleReconnect — max attempts ceiling", () => {
    it("stops reconnecting after maxReconnectAttempts", async () => {
      const { realtimeService } = await loadService();
      const svc = realtimeService as {
        reconnectAttempts: number;
        maxReconnectAttempts: number;
        handleReconnect: () => void;
      };
      svc.reconnectAttempts = svc.maxReconnectAttempts;
      svc.handleReconnect();
      // No new WS scheduled.
      const before = mockWsInstances.length;
      await vi.advanceTimersByTimeAsync(120_000);
      expect(mockWsInstances.length).toBe(before);
      expect(console.error).toHaveBeenCalledWith(expect.stringContaining("Max reconnect attempts reached"));
    });

    it("halts reconnect when token is null (auth-dead)", async () => {
      const { realtimeService } = await loadService();
      tokenStoreMock.getAccessToken.mockReturnValue(null);
      const before = mockWsInstances.length;
      (realtimeService as { handleReconnect: () => void }).handleReconnect();
      await vi.advanceTimersByTimeAsync(60_000);
      expect(mockWsInstances.length).toBe(before);
    });
  });

  describe("subscribe / unsubscribe / dispatch", () => {
    it("subscribe returns unsubscribe function that removes the handler", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn();
      const unsub = realtimeService.subscribe("test.event", cb);
      unsub();

      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "test.event", event_id: "e1", occurred_at: "", data: {} }),
      );
      expect(cb).not.toHaveBeenCalled();
    });

    it("dispatches event to registered handler", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn();
      realtimeService.subscribe("alert.fired", cb);
      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "alert.fired", event_id: "e1", occurred_at: "x", data: { level: "high" } }),
      );
      expect(cb).toHaveBeenCalledWith(expect.objectContaining({ type: "alert.fired" }));
    });

    it("catches handler exceptions without aborting dispatch", async () => {
      const { realtimeService } = await loadService();
      const goodCb = vi.fn();
      realtimeService.subscribe("x", () => {
        throw new Error("bad handler");
      });
      realtimeService.subscribe("x", goodCb);

      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "x", event_id: "e", occurred_at: "", data: {} }),
      );
      expect(goodCb).toHaveBeenCalled();
      expect(console.error).toHaveBeenCalled();
    });

    it("no-op on dispatch when no handlers for type", async () => {
      await loadService();
      // No subscribe — receive should not throw.
      expect(() =>
        mockWsInstances[0].receiveMessage(
          JSON.stringify({ schema_version: "1", type: "no.one.cares", event_id: "x", occurred_at: "", data: {} }),
        ),
      ).not.toThrow();
    });

    it("malformed JSON does not throw — logs error", async () => {
      await loadService();
      mockWsInstances[0].receiveMessage("{not json");
      expect(console.error).toHaveBeenCalledWith(expect.stringContaining("Failed to parse"), expect.anything());
    });

    it("removes handler set entirely after last unsubscribe", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn();
      const unsub = realtimeService.subscribe("removable", cb);
      unsub();
      const handlers = (realtimeService as { handlers: Map<string, Set<unknown>> }).handlers;
      expect(handlers.has("removable")).toBe(false);
    });

    it("unsubscribe with unknown type is a silent no-op", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn();
      expect(() => realtimeService.unsubscribe("never.subscribed", cb)).not.toThrow();
    });
  });

  describe("send", () => {
    it("sends JSON when socket is OPEN", async () => {
      const { realtimeService } = await loadService();
      mockWsInstances[0].simulateOpen();

      realtimeService.send("custom", { hello: "world" });
      const payload = JSON.parse(mockWsInstances[0].send.mock.calls[0][0] as string);
      expect(payload.type).toBe("custom");
      expect(payload.data).toEqual({ hello: "world" });
    });

    it("logs warn when socket is not open", async () => {
      const { realtimeService } = await loadService();
      // socket created but never opened (still CONNECTING).
      realtimeService.send("won't send", {});
      expect(console.warn).toHaveBeenCalledWith(expect.stringContaining("Cannot send"));
    });
  });

  describe("identity change → reconnect", () => {
    it("changes actor login → previous socket closed with 1000 and new connect attempted", async () => {
      const { realtimeService } = await loadService();
      mockWsInstances[0].simulateOpen();

      // Drive a store subscribe callback as if identity changed.
      const listener = subscribeListeners[0];
      expect(listener).toBeDefined();
      // Trigger the watcher with a new actor login.
      listener.cb(
        { actor: { login: "bob" }, role: "Developer" },
        { actor: { login: "alice" }, role: "Developer" },
      );

      // The current socket should have been closed cleanly.
      expect(mockWsInstances[0].readyState).toBe(MockWebSocket.CLOSED);

      // Allow connect chain to progress and a fresh ws to be created.
      await vi.advanceTimersByTimeAsync(0);
      await Promise.resolve();
      await Promise.resolve();
      expect(mockWsInstances.length).toBeGreaterThan(1);
    });

    it("changes role → reconnect", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();

      const listener = subscribeListeners[0];
      listener.cb(
        { actor: { login: "alice" }, role: "Manager" },
        { actor: { login: "alice" }, role: "Developer" },
      );

      expect(mockWsInstances[0].readyState).toBe(MockWebSocket.CLOSED);
    });

    it("no-op when neither login nor role changed", async () => {
      await loadService();
      mockWsInstances[0].simulateOpen();

      const before = mockWsInstances.length;
      const listener = subscribeListeners[0];
      listener.cb(
        { actor: { login: "alice" }, role: "Developer" },
        { actor: { login: "alice" }, role: "Developer" },
      );

      // No close, no extra connect.
      expect(mockWsInstances[0].readyState).toBe(MockWebSocket.OPEN);
      expect(mockWsInstances.length).toBe(before);
    });
  });

  describe("connect — skip when same URL + OPEN", () => {
    it("skips when already OPEN to same URL", async () => {
      const { realtimeService } = await loadService();
      mockWsInstances[0].simulateOpen();
      const before = mockWsInstances.length;

      // Direct private call — connect() should detect same URL OPEN and bail.
      await (realtimeService as { connect: () => Promise<void> }).connect();
      // Give microtasks a chance.
      await vi.advanceTimersByTimeAsync(0);
      expect(mockWsInstances.length).toBe(before);
    });
  });

  describe("WSEvent type generic — handler matches union", () => {
    it("typed subscribe carries through dispatched payload type", async () => {
      const { realtimeService } = await loadService();
      const cb = vi.fn<(e: WSEvent<{ value: number }>) => void>();
      realtimeService.subscribe<{ value: number }>("metric", cb);
      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "metric", event_id: "e", occurred_at: "", data: { value: 42 } }),
      );
      expect(cb).toHaveBeenCalled();
      expect((cb.mock.calls[0][0] as WSEvent<{ value: number }>).data.value).toBe(42);
    });
  });
});
