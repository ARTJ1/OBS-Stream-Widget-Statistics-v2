package hub

import (
	"encoding/json"
	"sync"

	"github.com/ARTJ1/OBS-Stream-Widget-Statistics-v2/internal/store"
)

type Message struct {
	Type     string         `json:"type"`
	State    store.State    `json:"state"`
	Settings store.Settings `json:"settings"`
}

type client chan []byte

type Hub struct {
	mu      sync.RWMutex
	clients map[client]struct{}
}

func New() *Hub {
	return &Hub{clients: make(map[client]struct{})}
}

func (h *Hub) Subscribe() client {
	ch := make(client, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch client) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Broadcast(msgType string, snap store.Snapshot) {
	payload, err := json.Marshal(Message{
		Type:     msgType,
		State:    snap.State,
		Settings: snap.Settings,
	})
	if err != nil {
		return
	}
	h.send(payload)
}

func (h *Hub) BroadcastShow(snap store.Snapshot) {
	h.Broadcast("show", snap)
}

func (h *Hub) send(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- payload:
		default:
		}
	}
}
