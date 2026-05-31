package ws

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHub_BroadcastScopedByOrg(t *testing.T) {
	h := NewHub(nil)

	a := &Client{orgID: "org-a", send: make(chan []byte, 1)}
	b := &Client{orgID: "org-b", send: make(chan []byte, 1)}
	h.register(a)
	h.register(b)

	h.Broadcast("org-a", []byte(`{"sku":"X"}`))

	select {
	case msg := <-a.send:
		assert.Equal(t, `{"sku":"X"}`, string(msg))
	case <-time.After(time.Second):
		t.Fatal("org-a client got no message")
	}

	select {
	case <-b.send:
		t.Fatal("org-b client must not receive org-a message")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHub_UnregisterDropsClient(t *testing.T) {
	h := NewHub(nil)
	c := &Client{orgID: "org-a", send: make(chan []byte, 1)}
	h.register(c)
	h.unregister(c)

	h.mu.RLock()
	defer h.mu.RUnlock()
	assert.Empty(t, h.clients["org-a"])
}
