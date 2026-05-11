package ws

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 20000
)

var (
	newline = []byte{'\n'}
)

type Client struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
}

func (h *Hub) Conn(conn *websocket.Conn, userID uuid.UUID) {
	if h == nil {
		return
	}

	client := New(userID, conn)
	h.register <- client

	go h.writeMessengerPump(client)
	h.readMessengerPump(client, userID)

}

func New(userID uuid.UUID, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
}

func (h *Hub) readMessengerPump(c *Client, userID uuid.UUID) {
	defer func() {
		h.unregister <- c
		if err := c.Conn.Close(); err != nil {
			zap.L().Error("failed to close conn after unregistering client", zap.Error(err))
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(
		func(string) error {
			return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		},
	)
	for {
		if _, _, err := c.Conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Error(
					"websocket conn unexpected close",
					zap.String("user_id", userID.String()),
					zap.Error(err),
				)
			}
			break
		}
	}
}

func (h *Hub) writeMessengerPump(c *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write(newline)
				_, _ = w.Write(<-c.Send)
			}

			if err = w.Close(); err != nil {
				return
			}

		}
	}
}
