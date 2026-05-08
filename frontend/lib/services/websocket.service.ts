/**
 * WebSocket Service
 * Handles real-time connection and event Pub/Sub for the frontend.
 */

export interface WsMessage<T = unknown> {
  schema_version: string;
  type: string;
  event_id: string;
  occurred_at: string;
  data: T;
}

type WsCallback = (payload: WsMessage) => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private isIntentionalClose = false;
  private listeners: Map<string, Set<WsCallback>> = new Map();
  private mockTimer: NodeJS.Timeout | null = null;

  constructor() {
    const httpUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
    this.url = httpUrl.replace(/^http/, 'ws') + "/api/v1/realtime/ws";
  }

  public connect() {
    if (this.ws && (this.ws.readyState === WebSocket.CONNECTING || this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    this.isIntentionalClose = false;
    try {
      console.log(`[WebSocket] Connecting to ${this.url}...`);
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log("[WebSocket] Connected successfully.");
        this.reconnectAttempts = 0;
        // this.startMockEvents(); // TEMP: For Phase 3 verification
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WsMessage = JSON.parse(event.data);
          this.dispatch(message);
        } catch (error) {
          console.error("[WebSocket] Failed to parse message:", error);
        }
      };

      this.ws.onclose = () => {
        this.stopMockEvents();
        if (!this.isIntentionalClose) {
          this.handleReconnect();
        } else {
          console.log("[WebSocket] Connection closed intentionally.");
        }
      };

      this.ws.onerror = (error) => {
        console.error("[WebSocket] Connection error:", error);
      };
    } catch (error) {
      console.error("[WebSocket] Setup failed:", error);
      this.handleReconnect();
    }
  }

  public disconnect() {
    this.isIntentionalClose = true;
    this.stopMockEvents();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
      console.log(`[WebSocket] Reconnecting in ${delay}ms (Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
      setTimeout(() => this.connect(), delay);
    } else {
      console.error("[WebSocket] Max reconnect attempts reached.");
    }
  }

  public subscribe(type: string, callback: WsCallback) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);
  }

  public unsubscribe(type: string, callback: WsCallback) {
    const typeListeners = this.listeners.get(type);
    if (typeListeners) {
      typeListeners.delete(callback);
      if (typeListeners.size === 0) {
        this.listeners.delete(type);
      }
    }
  }

  private dispatch(message: WsMessage) {
    const typeListeners = this.listeners.get(message.type);
    if (typeListeners) {
      typeListeners.forEach(cb => cb(message));
    }
    // Also dispatch to wildcard '*' if we ever need it
    const wildcardListeners = this.listeners.get('*');
    if (wildcardListeners) {
      wildcardListeners.forEach(cb => cb(message));
    }
  }

  /**
   * TEMP: Mock event generator to verify frontend reactivity without backend
   */
  private startMockEvents() {
    if (this.mockTimer) return;
    
    let counter = 0;
    this.mockTimer = setInterval(() => {
      counter++;
      const isCritical = counter % 3 === 0;
      
      const mockEvent: WsMessage = {
        schema_version: "1",
        type: isCritical ? "risk.critical.created" : "notification.created",
        event_id: `mock-evt-${Date.now()}`,
        occurred_at: new Date().toISOString(),
        data: {
          message: isCritical ? "Critical DB Latency Detected" : "New CI build finished",
        }
      };
      
      this.dispatch(mockEvent);
    }, 10000); // Fire every 10 seconds
  }

  private stopMockEvents() {
    if (this.mockTimer) {
      clearInterval(this.mockTimer);
      this.mockTimer = null;
    }
  }
}

export const websocketService = new WebSocketService();
