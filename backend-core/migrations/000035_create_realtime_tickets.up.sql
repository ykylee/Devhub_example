-- migration 000035: realtime_tickets table for multi-instance WebSocket auth ticket store
-- (sprint claude/work_260527-adr0024-ticket-store).
--
-- ADR-0024 §6 carve 6 — in-memory ticket store (internal/httpapi/realtime_ticket.go)
-- 는 single-instance 가정. sticky session 없이 horizontal scale 하면 인스턴스 A 가
-- 발급한 ticket 을 인스턴스 B 가 모르므로 WS handshake 가 실패한다. PG 백킹으로
-- 어느 인스턴스가 발급해도 다른 인스턴스가 single-use 로 소비 가능하게 한다.
--
-- single-use 보장: ConsumeRealtimeTicket 의 DELETE ... RETURNING 이 row 를 한 번만
-- 반환 (동시 consume 시 한쪽만 row 획득). expires_at > NOW() 조건으로 만료 ticket
-- 은 절대 honor 되지 않는다. 만료 row 는 InsertRealtimeTicket 의 opportunistic
-- cleanup (DeleteExpiredRealtimeTickets) 로 정리되어 별도 background sweeper 불요.

CREATE TABLE IF NOT EXISTS realtime_tickets (
    ticket      TEXT PRIMARY KEY,
    actor_login TEXT NOT NULL,
    actor_role  TEXT NOT NULL DEFAULT '',
    source_type TEXT NOT NULL DEFAULT 'oidc',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 만료 row opportunistic cleanup (DELETE WHERE expires_at <= NOW()) 의 인덱스.
CREATE INDEX IF NOT EXISTS realtime_tickets_expires_at_idx ON realtime_tickets (expires_at);
