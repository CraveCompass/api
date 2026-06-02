package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/CraveCompass/api/internal/domain"
)

type CreateSessionInput struct {
	HostID       string  `json:"host_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`
}

type CreateSessionUseCase struct {
	restaurantRepo RestaurantRepository
	sessionRepo    SessionRepository
}

func NewCreateSessionUseCase(rr RestaurantRepository, sr SessionRepository) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		restaurantRepo: rr,
		sessionRepo:    sr,
	}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, input CreateSessionInput) (*domain.Session, error) {
	if input.RadiusMeters <= 0 {
		return nil, errors.New("radius must be greater than zero")
	}

	restaurants, err := uc.restaurantRepo.GetByLocation(ctx, input.Latitude, input.Longitude, input.RadiusMeters)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(bytes)

	session := &domain.Session{
		ID:           sessionID,
		HostID:       input.HostID,
		Status:       domain.StatusActive,
		RadiusMeters: input.RadiusMeters,
		Participants: []domain.Participant{
			{ID: input.HostID, Username: "Host"},
		},
		Pool:      restaurants,
		CreatedAt: time.Now(),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}
