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
	Username     string          `json:"username"`
	RestaurantID string          `json:"restaurant_id"`
	Vote         domain.VoteType `json:"vote"`
}

type WSHandler struct {
	hub          *Hub
	submitVoteUC *application.SubmitVoteUseCase
	sessionRepo  application.SessionRepository
}

func NewWSHandler(hub *Hub, submitVoteUC *application.SubmitVoteUseCase, sessionRepo application.SessionRepository) *WSHandler {
	return &WSHandler{
		hub:          hub,
		submitVoteUC: submitVoteUC,
		sessionRepo:  sessionRepo,
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

	session, err := h.sessionRepo.GetByID(context.Background(), sessionID)
	if err == nil && session != nil {
		conn.WriteJSON(map[string]interface{}{
			"event":    "SESSION_UPDATED",
			"session":  session,
			"is_match": session.MatchedID != "",
		})
	}

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ClientMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			continue
		}

		if msg.Action == "JOIN_ROOM" {
			session, err := h.sessionRepo.GetByID(context.Background(), sessionID)
			if err == nil {
				session.AddParticipant(domain.Participant{ID: msg.UserID, Username: msg.Username})
				h.sessionRepo.Save(context.Background(), session)

				log.Printf("User %s joined room %s", msg.Username, sessionID)

				h.hub.Broadcast(sessionID, map[string]interface{}{
					"event":    "SESSION_UPDATED",
					"session":  session,
					"is_match": session.MatchedID != "",
				})
			}
		} else if msg.Action == "SWIPE" {
			log.Printf("User %s voted %s on restaurant %s", msg.UserID, msg.Vote, msg.RestaurantID)

			input := application.SubmitVoteInput{
				SessionID:    sessionID,
				UserID:       msg.UserID,
				RestaurantID: msg.RestaurantID,
				Vote:         msg.Vote,
			}

			updatedSession, isMatch, err := h.submitVoteUC.Execute(context.Background(), input)
			if err != nil {
				log.Printf("Error processing vote: %v", err)
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
