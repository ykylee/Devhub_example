package httpapi

import (
	"sync"
	"time"
)

// KratosSessionCache stores user_id → Kratos session_token mappings.
//
// DEC-D=α (work_26_05_11-e, L4-B): process-local in-memory map. Adequate for
// the single-instance PoC test server. When DevHub scales beyond one backend
// process, a stale session_token surfaces as a 401 from the settings flow —
// the handler returns code=REAUTH_REQUIRED and the frontend hands the user
// back through /login, which re-issues a fresh session_token. The β
// (DB-backed) and γ (Hydra access_token metadata) variants are deferred
// hardening items recorded in sprint backlog work_26_05_11-e.
//
// The cache key is the DevHub user_id (metadata_public.user_id from the
// Kratos identity), or the Kratos identity.id when the operator has not
// populated metadata yet — matching the subject the authLogin handler
// resolves before calling Put.
type KratosSessionCache struct {
	mu      sync.RWMutex
	entries map[string]kratosSessionEntry
}

type kratosSessionEntry struct {
	SessionToken string
	StoredAt     time.Time
}

func NewKratosSessionCache() *KratosSessionCache {
	return &KratosSessionCache{entries: map[string]kratosSessionEntry{}}
}

// Put records a fresh session_token for the given user_id. Empty inputs are
// ignored so handlers can call Put unconditionally on the login path.
func (c *KratosSessionCache) Put(userID, sessionToken string) {
	if c == nil || userID == "" || sessionToken == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[userID] = kratosSessionEntry{SessionToken: sessionToken, StoredAt: time.Now()}
}

// Get returns the most recently stored session_token for user_id. The
// boolean is false when no token has been cached (caller should return
// REAUTH_REQUIRED).
func (c *KratosSessionCache) Get(userID string) (string, bool) {
	if c == nil || userID == "" {
		return "", false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[userID]
	if !ok {
		return "", false
	}
	return entry.SessionToken, true
}

// Delete removes the cached session_token (e.g. on logout). Safe to call for
// unknown user_ids.
func (c *KratosSessionCache) Delete(userID string) {
	if c == nil || userID == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, userID)
}

// Len returns the number of cached entries. Test-only convenience.
func (c *KratosSessionCache) Len() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
