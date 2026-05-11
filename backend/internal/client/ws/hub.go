package ws

import (
	"context"

	"dd-prediction-api/internal/storage/cache"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var hub *Hub

type Hub struct {
	clients    map[uuid.UUID]map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	// event streaming
	eventPubSub *cache.EventPubSub
	// per-user subscription cancelers
	userCancels map[uuid.UUID]context.CancelFunc
}

func init() {
	hub = &Hub{
		broadcast:   make(chan Message),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[uuid.UUID]map[*Client]bool),
		userCancels: make(map[uuid.UUID]context.CancelFunc),
	}

	go hub.Run()
}

func Stream() *Hub {
	return hub
}

type Message struct {
	Message []byte
	UserID  uuid.UUID
}

// SetEventPubSub injects the redis pub/sub dependency.
func (h *Hub) SetEventPubSub(ps *cache.EventPubSub) {
	h.eventPubSub = ps
}

func (h *Hub) ensureSubscription(userID uuid.UUID) {
	if h.eventPubSub == nil {
		return
	}
	if _, ok := h.userCancels[userID]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.userCancels[userID] = cancel

	sub := h.eventPubSub.Subscribe(ctx, userID)

	go func() {
		defer func() {
			if err := sub.Close(); err != nil {
				zap.L().Error("failed to close subscription", zap.Error(err))
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-sub.Channel():
				h.broadcast <- Message{UserID: userID, Message: []byte(msg.Payload)}
			}
		}
	}()
}

func (h *Hub) stopSubscriptionIfIdle(userID uuid.UUID) {
	if canc, ok := h.userCancels[userID]; ok {
		// stop only if there are no clients left for this user
		if userClients, has := h.clients[userID]; !has || len(userClients) == 0 {
			canc()
			delete(h.userCancels, userID)
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			userID := client.UserID
			if _, ok := h.clients[userID]; !ok {
				h.clients[userID] = make(map[*Client]bool)
			}
			h.clients[userID][client] = true
			// start per-user subscription on first client
			h.ensureSubscription(userID)

		case client := <-h.unregister:
			userID := client.UserID

			if userClients, ok := h.clients[userID]; ok {
				if _, ok = userClients[client]; ok {
					delete(userClients, client)
					close(client.Send)

					if len(userClients) == 0 {
						delete(h.clients, userID)
						// stop per-user subscription when last client disconnects
						h.stopSubscriptionIfIdle(userID)
					}
				}
			}

		case message := <-h.broadcast:
			userID := message.UserID

			if userClients, ok := h.clients[userID]; ok {
				for client := range userClients {
					select {
					case client.Send <- message.Message:
					default:
						close(client.Send)
						delete(userClients, client)

						if len(userClients) == 0 {
							delete(h.clients, userID)
							h.stopSubscriptionIfIdle(userID)
						}
					}
				}
			}
		}
	}
}
