// Package ws provides the per-org WebSocket fan-out hub.
package ws

import (
	"log/slog"
	"sync"
)

// Client is one connected WebSocket subscriber.
type Client struct {
	orgID string
	send  chan []byte
}

// Hub fans NATS-sourced events out to clients, scoped by org.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]struct{}
	log     *slog.Logger
}

// NewHub builds an empty hub.
func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{clients: map[string]map[*Client]struct{}{}, log: log}
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[c.orgID] == nil {
		h.clients[c.orgID] = map[*Client]struct{}{}
	}
	h.clients[c.orgID][c] = struct{}{}
}

func (h *Hub) unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.clients[c.orgID]
	if !ok {
		return
	}
	if _, ok := set[c]; ok {
		delete(set, c)
		close(c.send)
	}
	if len(set) == 0 {
		delete(h.clients, c.orgID)
	}
}

// Broadcast pushes data to every client in orgID. Slow clients are dropped, not blocked.
func (h *Hub) Broadcast(orgID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[orgID] {
		select {
		case c.send <- data:
		default:
			h.log.Warn("ws client slow, dropping message", "org_id", orgID)
		}
	}
}
