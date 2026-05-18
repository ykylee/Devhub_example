package httpapi

import "testing"

func TestIDPSessionCache_PutGetDelete(t *testing.T) {
	c := NewIDPSessionCache()

	if _, ok := c.Get("alice"); ok {
		t.Errorf("empty cache should miss")
	}

	c.Put("alice", "tok-1")
	got, ok := c.Get("alice")
	if !ok {
		t.Fatalf("expected cache hit after Put")
	}
	if got != "tok-1" {
		t.Errorf("token = %q, want tok-1", got)
	}

	// Put overrides.
	c.Put("alice", "tok-2")
	got, _ = c.Get("alice")
	if got != "tok-2" {
		t.Errorf("token after override = %q, want tok-2", got)
	}

	c.Delete("alice")
	if _, ok := c.Get("alice"); ok {
		t.Errorf("Delete did not remove entry")
	}
}

func TestIDPSessionCache_IgnoresEmptyInputs(t *testing.T) {
	c := NewIDPSessionCache()
	c.Put("", "tok")
	c.Put("alice", "")
	if c.Len() != 0 {
		t.Errorf("empty inputs must not write; Len=%d", c.Len())
	}
	if _, ok := c.Get(""); ok {
		t.Errorf("Get on empty user_id should miss")
	}
	c.Delete("") // must not panic
}

func TestIDPSessionCache_NilReceiverSafe(t *testing.T) {
	var c *IDPSessionCache
	c.Put("a", "b")
	if _, ok := c.Get("a"); ok {
		t.Errorf("nil receiver Get should miss")
	}
	c.Delete("a")
	if c.Len() != 0 {
		t.Errorf("nil receiver Len should be 0")
	}
}
