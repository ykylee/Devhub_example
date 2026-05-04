import { WSEvent, WSEventHandler } from "./types";

const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/realtime/ws';

export class RealtimeService {
  private static instance: RealtimeService;
  private socket: WebSocket | null = null;
  private handlers: Map<string, Set<WSEventHandler>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000;
  public isConnected = false;

  private constructor() {
    if (typeof window !== 'undefined') {
      this.connect();
    }
  }

  public static getInstance(): RealtimeService {
    if (!RealtimeService.instance) {
      RealtimeService.instance = new RealtimeService();
    }
    return RealtimeService.instance;
  }

  private connect() {
    try {
      console.log(`[RealtimeService] Connecting to ${WS_BASE}...`);
      this.socket = new WebSocket(WS_BASE);

      this.socket.onopen = () => {
        console.log('[RealtimeService] Connected.');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.dispatch({ type: 'status.changed', data: { connected: true } } as any);
      };

      this.socket.onmessage = (event) => {
        try {
          const wsEvent: WSEvent = JSON.parse(event.data);
          this.dispatch(wsEvent);
        } catch (e) {
          console.error('[RealtimeService] Failed to parse message:', e);
        }
      };

      this.socket.onclose = (event) => {
        console.log(`[RealtimeService] Disconnected. Code: ${event.code}`);
        this.isConnected = false;
        this.dispatch({ type: 'status.changed', data: { connected: false } } as any);
        this.handleReconnect();
      };

      this.socket.onerror = (error) => {
        console.error('[RealtimeService] WebSocket Error:', error);
      };
    } catch (error) {
      console.error('[RealtimeService] Connection Error:', error);
      this.handleReconnect();
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`[RealtimeService] Reconnecting in ${this.reconnectInterval}ms... (Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
      setTimeout(() => this.connect(), this.reconnectInterval);
    } else {
      console.error('[RealtimeService] Max reconnect attempts reached.');
    }
  }

  private dispatch(event: WSEvent) {
    const eventHandlers = this.handlers.get(event.type);
    if (eventHandlers) {
      eventHandlers.forEach(handler => {
        try {
          handler(event);
        } catch (e) {
          console.error(`[RealtimeService] Error in handler for ${event.type}:`, e);
        }
      });
    }
  }

  public subscribe<T = any>(type: string, handler: WSEventHandler<T>) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler);

    return () => this.unsubscribe(type, handler);
  }

  public unsubscribe<T = any>(type: string, handler: WSEventHandler<T>) {
    const eventHandlers = this.handlers.get(type);
    if (eventHandlers) {
      eventHandlers.delete(handler);
      if (eventHandlers.size === 0) {
        this.handlers.delete(type);
      }
    }
  }

  public send(type: string, data: any) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type, data }));
    } else {
      console.warn('[RealtimeService] Cannot send message: WebSocket is not open.');
    }
  }
}

export const realtimeService = RealtimeService.getInstance();
