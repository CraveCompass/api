package http

import (
	"encoding/json"
	"net/http"

	"github.com/CraveCompass/api/internal/application"
)

type SessionHandler struct {
	createSessionUC *application.CreateSessionUseCase
}

func NewSessionHandler(uc *application.CreateSessionUseCase) *SessionHandler {
	return &SessionHandler{
		createSessionUC: uc,
	}
}

// HandleCreateRoom handles POST /sessions
func (h *SessionHandler) HandleCreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input application.CreateSessionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	session, err := h.createSessionUC.Execute(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}
