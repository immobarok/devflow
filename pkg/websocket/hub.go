package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"github.com/mobarok/devflow/pkg/pubsub"
)

// Hub maintains the set of active clients and broadcasts messages to clients.
type Hub struct {
	// Registered clients mapped by UserID -> Set of *Client (supports multiple tabs/devices)
	clients    map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	bus        pubsub.Bus
	mu         sync.RWMutex
}

func NewHub(bus pubsub.Bus) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		bus:        bus,
	}
}

// Run starts the Hub event loop and listens to Redis Pub/Sub broadcast channel in background.
func (h *Hub) Run(ctx context.Context) {
	// If Redis Pub/Sub bus is configured, subscribe to event channels in background
	if h.bus != nil {
		go h.listenRedisPubSub(ctx)
	}

	for {
		select {
		case <-ctx.Done():
			h.cleanup()
			return

		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			slog.Debug("websocket client registered", "user_id", client.UserID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if userClients, ok := h.clients[client.UserID]; ok {
				if _, exists := userClients[client]; exists {
					delete(userClients, client)
					close(client.Send)
					if len(userClients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			slog.Debug("websocket client unregistered", "user_id", client.UserID)
		}
	}
}

// SendToUser pushes an event payload directly to all active WebSocket connections for a user.
func (h *Hub) SendToUser(userID string, event pubsub.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range userClients {
		select {
		case client.Send <- data:
		default:
			// If buffer is blocked, drop to avoid hanging
		}
	}
}

// BroadcastAll sends an event to every connected client on this server.
func (h *Hub) BroadcastAll(event pubsub.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userClients := range h.clients {
		for client := range userClients {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) listenRedisPubSub(ctx context.Context) {
	msgChan := h.bus.SubscribePattern(ctx, "events:*")
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			var event pubsub.Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				continue
			}

			// Channel name format: events:user:<user_id>
			if strings.HasPrefix(msg.Channel, "events:user:") {
				targetUserID := strings.TrimPrefix(msg.Channel, "events:user:")
				h.SendToUser(targetUserID, event)
			} else {
				h.BroadcastAll(event)
			}
		}
	}
}

func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, userClients := range h.clients {
		for client := range userClients {
			close(client.Send)
		}
	}
	h.clients = make(map[string]map[*Client]bool)
}
