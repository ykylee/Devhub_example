import { WSEvent, WSEventHandler } from "./types";
import { useStore } from "@/lib/store";

const WS_BASE = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/realtime/ws';
const DEFAULT_EVENT_TYPES = [
  'command.status.updated',
  'infra.node.updated',
  'infra.edge.updated',
  'risk.critical.created',
  'notification.created'
];

export type ConnectionStatusEvent = { connected: boolean };

export class RealtimeService {
  private static instance: RealtimeService;
  private socket: WebSocket | null = null;
  private handlers: Map<string, Set<WSEventHandler<unknown>>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000;
  private currentUrl: string | null = null;
  public isConnected = false;

  private constructor() {
    if (typeof window !== 'undefined') {
      this.init();
    }
  }

  public static getInstance(): RealtimeService {
    if (!RealtimeService.instance) {
      RealtimeService.instance = new RealtimeService();
    }
    return RealtimeService.instance;
  }

  private init() {
    this.connect();

    // Watch for store changes to trigger reconnection if identity changes
    useStore.subscribe(
      (state) => ({ actor: state.actor, role: state.role }),
      (current, previous) => {
        if (
          current.actor?.login !== previous.actor?.login || 
          current.role !== previous.role
        ) {
          console.log('[RealtimeService] Identity changed, reconnecting...');
          this.reconnect();
        }
      }
    );
  }

  private connect() {
    try {
      const url = this.buildURL();
      if (this.socket && this.currentUrl === url) return;

      if (this.socket) {
        this.socket.close();
      }

      this.currentUrl = url;
      console.log(`[RealtimeService] Connecting to ${url}...`);
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        console.log('[RealtimeService] Connected.');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.dispatch({ 
          type: 'status.changed', 
          data: { connected: true },
          schema_version: '1',
          event_id: 'internal',
          occurred_at: new Date().toISOString()
        } as WSEvent);
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
        this.dispatch({ 
          type: 'status.changed', 
          data: { connected: false },
          schema_version: '1',
          event_id: 'internal',
          occurred_at: new Date().toISOString()
        } as WSEvent);
        
        // Only reconnect if it wasn't a clean close for identity change
        if (event.code !== 1000) {
          this.handleReconnect();
        }
      };

      this.socket.onerror = (error) => {
        console.error('[RealtimeService] WebSocket Error:', error);
      };
    } catch (error) {
      console.error('[RealtimeService] Connection Error:', error);
      this.handleReconnect();
    }
  }

  private reconnect() {
    this.reconnectAttempts = 0;
    if (this.socket) {
      this.socket.close(1000, "Identity change");
    }
    this.connect();
  }

  private buildURL() {
    const { actor, role } = useStore.getState();
    const separator = WS_BASE.includes('?') ? '&' : '?';
    const types = encodeURIComponent(DEFAULT_EVENT_TYPES.join(','));
    
    const actorParam = actor?.login || 'guest';
    
    const roleMap: Record<string, string> = {
      "System Admin": "system_admin",
      "Manager": "manager",
      "Developer": "developer"
    };
    const roleParam = role ? (roleMap[role] || role.toLowerCase()) : 'guest';

    return `${WS_BASE}${separator}types=${types}&actor=${actorParam}&role=${roleParam}`;
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

  public subscribe<T = unknown>(type: string, handler: WSEventHandler<T>) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler as WSEventHandler<unknown>);

    return () => this.unsubscribe(type, handler);
  }

  public unsubscribe<T = unknown>(type: string, handler: WSEventHandler<T>) {
    const eventHandlers = this.handlers.get(type);
    if (eventHandlers) {
      eventHandlers.delete(handler as WSEventHandler<unknown>);
      if (eventHandlers.size === 0) {
        this.handlers.delete(type);
      }
    }
  }

  public send(type: string, data: unknown) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify({ type, data }));
    } else {
      console.warn('[RealtimeService] Cannot send message: WebSocket is not open.');
    }
  }
}

export const realtimeService = RealtimeService.getInstance();

