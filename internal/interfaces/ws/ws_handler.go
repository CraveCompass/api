package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/CraveCompass/api/internal/application"
	"github.com/CraveCompass/api/internal/domain"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ClientMessage struct {
	Action       string          `json:"action"`
	UserID       string          `json:"user_id"`
	RestaurantID string          `json:"restaurant_id"`
	Vote         domain.VoteType `json:"vote"`
}

type WSHandler struct {
	hub          *Hub
	submitVoteUC *application.SubmitVoteUseCase
}

func NewWSHandler(hub *Hub, submitVoteUC *application.SubmitVoteUseCase) *WSHandler {
	return &WSHandler{
		hub:          hub,
		submitVoteUC: submitVoteUC,
	}
}

func (h *WSHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id query parameter", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer func() {
		h.hub.RemoveClient(sessionID, conn)
		conn.Close()
	}()

	h.hub.AddClient(sessionID, conn)

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection closed or read error: %v", err)
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Invalid JSON payload received over socket: %v", err)
			continue
		}

		if msg.Action == "SWIPE" {
			input := application.SubmitVoteInput{
				SessionID:    sessionID,
				UserID:       msg.UserID,
				RestaurantID: msg.RestaurantID,
				Vote:         msg.Vote,
			}

			updatedSession, isMatch, err := h.submitVoteUC.Execute(context.Background(), input)
			if err != nil {
				log.Printf("Error processing vote usecase: %v", err)
				continue
			}

			h.hub.Broadcast(sessionID, map[string]interface{}{
				"event":    "SESSION_UPDATED",
				"session":  updatedSession,
				"is_match": isMatch,
			})
		}
	}
}
