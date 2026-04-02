package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // MVP: In production, check origin strictly
	},
}

type WSHub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var Hub = &WSHub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *WSHub) Broadcast(entry LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()

	message, err := json.Marshal(entry)
	if err != nil {
		return
	}

	for client := range h.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(h.clients, client)
		}
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// auth has already been handled by AuthMiddleware
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	Hub.mu.Lock()
	Hub.clients[ws] = true
	Hub.mu.Unlock()

	// Keep alive loop
	go func() {
		defer func() {
			Hub.mu.Lock()
			delete(Hub.clients, ws)
			Hub.mu.Unlock()
			ws.Close()
		}()
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
