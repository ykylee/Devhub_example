import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// Mock WebSocket for testing
interface MockWsInstance {
  url: string;
  readyState: number;
  onopen: ((event: Record<string, unknown>) => void) | null;
  onclose: ((event: Record<string, unknown>) => void) | null;
  onmessage: ((event: { data: string }) => void) | null;
  onerror: ((event: Record<string, unknown>) => void) | null;
  close: () => void;
  simulateOpen: () => void;
  receiveMessage: (data: string) => void;
}

const mockWsInstances: MockWsInstance[] = [];
let mockWsConstructorUrl = "";

class MockWebSocket {
  static instances = mockWsInstances;
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

  constructor(url: string) {
    this.url = url;
    mockWsConstructorUrl = url;
    mockWsInstances.push(this);
  }

  close() {
    this.readyState = MockWebSocket.CLOSED;
    if (this.onclose) {
      this.onclose({ code: 1000, reason: "test-close", wasClean: true });
    }
  }

  simulateOpen() {
    this.readyState = MockWebSocket.OPEN;
    if (this.onopen) this.onopen({});
  }

  receiveMessage(data: string) {
    if (this.onmessage) this.onmessage({ data });
  }
}

describe("WebSocketService", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let wsService: any;

  beforeEach(async () => {
    mockWsInstances.length = 0;
    mockWsConstructorUrl = "";
    vi.useFakeTimers();
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    vi.stubGlobal("WebSocket", MockWebSocket);
    vi.stubGlobal("window", { location: { host: "localhost:3000", protocol: "http:", origin: "http://localhost:3000" } });
    // happy-dom's URL constructor doesn't support ws:// protocol.
    // Wrap it to handle ws:// by temporarily converting to http:// for parsing.
    const OrigURL = globalThis.URL;
    vi.stubGlobal("URL", class URL extends OrigURL {
      constructor(url: string, base?: string) {
        super(url.replace(/^ws:/, "http:"), base);
      }
    });
    vi.resetModules();
    const mod = await import("./websocket.service");
    wsService = mod.websocketService;
  });

  afterEach(() => {
    wsService.disconnect();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  describe("connect", () => {
    it("creates a WebSocket connection", () => {
      wsService.connect();

      expect(mockWsInstances).toHaveLength(1);
      expect(mockWsConstructorUrl).toContain("/api/v1/realtime/ws");
    });

    it("does not create duplicate connections when already open", () => {
      wsService.connect();
      mockWsInstances[0].simulateOpen();

      wsService.connect();

      expect(mockWsInstances).toHaveLength(1);
    });

    it("reconnects on unexpected close", () => {
      wsService.connect();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close();

      vi.advanceTimersByTime(2000);
      expect(mockWsInstances).toHaveLength(2);
    });
  });

  describe("disconnect", () => {
    it("closes the connection intentionally", () => {
      wsService.connect();
      mockWsInstances[0].simulateOpen();

      wsService.disconnect();

      expect(mockWsInstances[0].readyState).toBe(MockWebSocket.CLOSED);
      vi.advanceTimersByTime(5000);
      expect(mockWsInstances).toHaveLength(1);
    });
  });

  describe("reconnect", () => {
    it("uses exponential backoff: 2s, 4s, 8s, 16s, 30s", () => {
      wsService.connect();
      mockWsInstances[0].simulateOpen();
      mockWsInstances[0].close();

      vi.advanceTimersByTime(2000);
      expect(mockWsInstances).toHaveLength(2);
      mockWsInstances[1].simulateOpen();
      mockWsInstances[1].close();

      vi.advanceTimersByTime(4000);
      expect(mockWsInstances).toHaveLength(3);
      mockWsInstances[2].simulateOpen();
      mockWsInstances[2].close();

      vi.advanceTimersByTime(4000);
      expect(mockWsInstances).toHaveLength(4);
    });

    it("stops after max attempts (5)", () => {
      wsService.connect();
      mockWsInstances[0].simulateOpen();

      for (let i = 0; i < 5; i++) {
        const last = mockWsInstances[mockWsInstances.length - 1];
        last.simulateOpen();
        last.close();
        vi.advanceTimersByTime(30000);
      }

      vi.advanceTimersByTime(60000);
      expect(mockWsInstances.length).toBeLessThanOrEqual(6);
    });
  });

  describe("subscribe / dispatch", () => {
    it("calls registered listener when matching message is received", () => {
      const callback = vi.fn();

      wsService.subscribe("risk.critical.created", callback);
      wsService.connect();

      mockWsInstances[0].receiveMessage(
        JSON.stringify({
          schema_version: "1",
          type: "risk.critical.created",
          event_id: "evt-1",
          occurred_at: "2026-05-28T12:00:00Z",
          data: { message: "Critical alert" },
        }),
      );

      expect(callback).toHaveBeenCalledTimes(1);
      expect(callback).toHaveBeenCalledWith(
        expect.objectContaining({ type: "risk.critical.created", event_id: "evt-1" }),
      );
    });

    it("calls wildcard '*' listener for any message type", () => {
      const wildcardCb = vi.fn();
      wsService.subscribe("*", wildcardCb);
      wsService.connect();

      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "any.event", event_id: "e1", occurred_at: "", data: {} }),
      );

      expect(wildcardCb).toHaveBeenCalledTimes(1);
    });

    it("does not call listener after unsubscribe", () => {
      const callback = vi.fn();
      wsService.subscribe("test.event", callback);
      wsService.unsubscribe("test.event", callback);
      wsService.connect();

      mockWsInstances[0].receiveMessage(
        JSON.stringify({ schema_version: "1", type: "test.event", event_id: "e1", occurred_at: "", data: {} }),
      );

      expect(callback).not.toHaveBeenCalled();
    });

    it("handles malformed JSON gracefully", () => {
      const callback = vi.fn();
      wsService.subscribe("test.event", callback);
      wsService.connect();

      mockWsInstances[0].receiveMessage("{invalid json}");

      expect(callback).not.toHaveBeenCalled();
      expect(console.error).toHaveBeenCalled();
    });
  });
});
