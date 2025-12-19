package api

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP
	},
}

func (rt *Router) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		rt.logger.Error("WebSocket upgrade failed", err)
		return
	}
	defer conn.Close()

	rt.logger.Info("WebSocket client connected", "remote", r.RemoteAddr)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				rt.logger.Error("WebSocket read error", err)
			}
			break
		}

		// Process message
		rt.logger.Info("Received WebSocket message", "len", len(p))

		// Echo back success (optional, for streaming proto)
		if err := conn.WriteMessage(messageType, []byte("ACK")); err != nil {
			rt.logger.Error("WebSocket write error", err)
			break
		}
	}
}
