import { WSEvent, WSEventHandler } from "@/shared/api/types";
import { useStore } from "@/lib/store";
import { apiClient, ApiError } from "@/shared/api/api-client";
import { tokenStore } from "@/domain/auth-session/service/token-store";

import { WS_BASE_URL as WS_BASE } from "@/shared/config/endpoints";
// codex P1 (PR #252 review): `infra.service.updated` 는 backend
// handleRealtimeWebSocket 의 supported types 에 없음 — default 에 포함 시
// secured 환경에서 WS handshake fail. topology-v2 page 가 ad-hoc subscribe.
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
  // STABLE_OPEN_MS 이상 연결이 유지된 뒤에만 reconnectAttempts 를 0 으로 리셋한다.
  // (인증 실패로 즉시 1006-close 되는 연결이 onopen→close 를 반복하며 max-5 cap 을
  //  우회해 영구 재연결 루프에 빠지던 #387 ③ 버그 차단.)
  // 30초 → 5분 (#388 hotfix): 일부 환경에서 30초 이상 연결 유지 후 idle drop 되는
  // 패턴이 있어 attempts 가 매 cycle 리셋되어 사용자엔 "주기적 connect↔disconnect
  // 반복" 으로 관측됐음. 5분으로 늘리고, 동시에 heartbeat 로 idle drop 자체를 예방.
  private stableOpenMs = 5 * 60_000;
  private stableOpenTimer: ReturnType<typeof setTimeout> | null = null;
  // Heartbeat (#388 hotfix): nginx/proxy/CDN idle timeout (보통 60s) 으로 WS 가
  // 끊기는 것을 방지하기 위해 클라이언트가 25초마다 작은 ping 메시지를 흘려보낸다.
  // backend (`HandleWebSocket` 의 `conn.ReadMessage()`) 가 메시지를 읽고 폐기하므로
  // 서버 측 추가 처리 불필요. WebSocket 표준 ping/pong 프레임은 브라우저 API 에서
  // 직접 보낼 수 없어 데이터 메시지(text) 로 대체.
  private heartbeatIntervalMs = 25_000;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
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

  private async fetchTicket(): Promise<string | null> {
    // ADR-0024 §3.2 ticket pattern. 401 시 refresh-then-retry 는 apiClient 가
    // 내부에서 `refreshAccessToken()` (단일 single-flight mutex, #388 codex P1)
    // 으로 처리한다. 본 함수는 결과만 받음 — 이중 refresh 시도(레거시 authService.
    // refreshTokens 직접 호출) 를 제거해 Keycloak `Refresh Token Max Reuse=0`
    // 환경에서 동시 invalid_grant 가 터지던 race 차단.
    try {
      const resp = await apiClient<{ ticket: string }>("POST", "/api/v1/realtime/ticket");
      return resp.ticket;
    } catch (e) {
      if (e instanceof ApiError) {
        console.warn('[RealtimeService] Ticket fetch failed (status %d); WS will retry on next reconnect.', e.status);
      } else {
        console.warn('[RealtimeService] Ticket fetch network error:', e);
      }
      return null;
    }
  }

  private async connect() {
    // Auth-dead 가드: access token 이 없으면 연결 시도 자체를 건너뜀. 서버가
    // 즉시 1006-close 하여 재연결 루프에 빠지는 것을 차단 (#387 ③). 로그인 후
    // store identity 변경이 발생하면 reconnect() 가 다시 진입.
    if (typeof window !== 'undefined' && tokenStore.getAccessToken() === null) {
      console.log('[RealtimeService] No access token; skipping connect (will retry on login).');
      return;
    }
    try {
      const url = await this.buildURL();
      if (this.socket && this.socket.readyState === WebSocket.OPEN && this.currentUrl === url) return;

      if (this.socket) {
        this.socket.close();
      }

      this.currentUrl = url;
      console.log(`[RealtimeService] Connecting to ${url}...`);
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        console.log('[RealtimeService] Connected.');
        this.isConnected = true;
        // 즉시 reconnectAttempts=0 리셋 금지 — STABLE_OPEN_MS 이상 유지된 뒤에만 리셋.
        // 인증 실패로 onopen→1006 close 가 반복되는 경우 attempts 가 누적되어 max-5
        // cap 이 동작 (#387 ③).
        if (this.stableOpenTimer) clearTimeout(this.stableOpenTimer);
        this.stableOpenTimer = setTimeout(() => {
          this.reconnectAttempts = 0;
          this.stableOpenTimer = null;
        }, this.stableOpenMs);
        // Heartbeat 시작 — idle timeout (nginx/proxy 기본 60s) 으로 인한 drop 예방.
        this.startHeartbeat();
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
        // stable-open 도달 전에 close 되면 타이머 취소 — attempts 누적 유지.
        if (this.stableOpenTimer) {
          clearTimeout(this.stableOpenTimer);
          this.stableOpenTimer = null;
        }
        // Heartbeat 정지 — 연결이 죽었으므로 더 보낼 필요 없음.
        this.stopHeartbeat();
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

  private async buildURL(): Promise<string> {
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

    // ADR-0024 §6 carve 5 (ticket-only 컷오버): ticket (single-use + 60s TTL) 만
    // 사용. 레거시 `?access_token=` query fallback 은 제거 — browser WS API 의
    // Authorization header 제약 우회는 ticket pattern 으로 일원화 (URL/log token
    // 노출 위협 제거). ticket fetch 실패 시 token 미첨부 → backend 401 →
    // handleReconnect 가 재시도 (ticket 재발급 + 401 refresh-then-retry).
    let tokenParam = '';
    const ticket = await this.fetchTicket();
    if (ticket) {
      tokenParam = `&ticket=${encodeURIComponent(ticket)}`;
    }

    return `${WS_BASE}${separator}types=${types}&actor=${actorParam}&role=${roleParam}${tokenParam}`;
  }

  // Heartbeat 제어 — onopen 에서 startHeartbeat, onclose 에서 stopHeartbeat.
  // 25초 (heartbeatIntervalMs) 마다 작은 ping 메시지를 보내 nginx/proxy/CDN 의
  // idle timeout (보통 60s) 에 걸려 연결이 끊기는 것을 예방. backend `ReadMessage()`
  // 는 메시지를 읽고 폐기하므로 서버 측 추가 처리 불필요.
  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.socket && this.socket.readyState === WebSocket.OPEN) {
        try {
          this.socket.send(JSON.stringify({ type: 'ping', ts: Date.now() }));
        } catch (err) {
          console.warn('[RealtimeService] heartbeat send failed', err);
        }
      }
    }, this.heartbeatIntervalMs);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private handleReconnect() {
    // Auth-dead 가드: 세션이 죽었으면 재연결 시도 자체를 중단. 재로그인 시
    // store identity 변경으로 reconnect() 가 재진입.
    if (typeof window !== 'undefined' && tokenStore.getAccessToken() === null) {
      console.log('[RealtimeService] No access token; halting reconnect (will resume on login).');
      return;
    }
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[RealtimeService] Max reconnect attempts reached.');
      return;
    }
    this.reconnectAttempts++;
    // Exponential backoff: base * 2^(attempts-1), cap 60s (3s → 6s → 12s → 24s → 48s).
    // 백엔드/네트워크 hammering 방지.
    const delay = Math.min(60_000, this.reconnectInterval * Math.pow(2, this.reconnectAttempts - 1));
    console.log(`[RealtimeService] Reconnecting in ${delay}ms... (Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
    setTimeout(() => this.connect(), delay);
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

