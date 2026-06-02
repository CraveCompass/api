package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	rooms map[string]map[*websocket.Conn]bool
	mu    sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *Hub) AddClient(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[sessionID] == nil {
		h.rooms[sessionID] = make(map[*websocket.Conn]bool)
	}
	h.rooms[sessionID][conn] = true
	log.Printf("Client joined session %s", sessionID)
}

func (h *Hub) RemoveClient(sessionID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[sessionID]; ok {
		delete(h.rooms[sessionID], conn)
		if len(h.rooms[sessionID]) == 0 {
			delete(h.rooms, sessionID)
		}
	}
}

func (h *Hub) Broadcast(sessionID string, message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.rooms[sessionID]
	for conn := range clients {
		err := conn.WriteJSON(message)
		if err != nil {
			log.Printf("Error broadcasting to client: %v", err)
			conn.Close()
		}
	}
}
