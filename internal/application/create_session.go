package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/CraveCompass/api/internal/domain"
)

type CreateSessionInput struct {
	HostID       string  `json:"host_id"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeters int     `json:"radius_meters"`

	PriceTiers []int    `json:"price_tiers"`
	MinRating  *float64 `json:"min_rating"`
	Cuisines   []string `json:"cuisines"`
}

type CreateSessionUseCase struct {
	restaurantRepo RestaurantRepository
	sessionRepo    SessionRepository
	enrichUC       *EnrichRestaurantsUseCase
	broadcaster    SessionBroadcaster
}

func NewCreateSessionUseCase(rr RestaurantRepository, sr SessionRepository, enrichUC *EnrichRestaurantsUseCase, broadcaster SessionBroadcaster) *CreateSessionUseCase {
	return &CreateSessionUseCase{
		restaurantRepo: rr,
		sessionRepo:    sr,
		enrichUC:       enrichUC,
		broadcaster:    broadcaster,
	}
}

func (uc *CreateSessionUseCase) Execute(ctx context.Context, input CreateSessionInput) (*domain.Session, error) {
	if input.RadiusMeters <= 0 {
		return nil, errors.New("radius must be greater than zero")
	}

	restaurants, err := uc.restaurantRepo.GetByLocation(ctx, input.Latitude, input.Longitude, input.RadiusMeters, input.PriceTiers, input.MinRating, input.Cuisines)
	if err != nil {
		return nil, err
	}

	if len(restaurants) < 5 {
		err := uc.restaurantRepo.FetchAndSaveFromOSM(ctx, input.Latitude, input.Longitude, input.RadiusMeters)
		if err != nil {
			log.Printf("Warning: OSM Fetch failed: %v", err)
		}
		restaurants, err = uc.restaurantRepo.GetByLocation(ctx, input.Latitude, input.Longitude, input.RadiusMeters, input.PriceTiers, input.MinRating, input.Cuisines)
		if err != nil {
			return nil, err
		}
	}

	if len(restaurants) == 0 {
		return nil, errors.New("no restaurants found in this area")
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(bytes)

	minRatingVal := 0.0
	if input.MinRating != nil {
		minRatingVal = *input.MinRating
	}

	session := &domain.Session{
		ID:           sessionID,
		HostID:       input.HostID,
		Status:       domain.StatusActive,
		RadiusMeters: input.RadiusMeters,
		Participants: []domain.Participant{
			{ID: input.HostID, Username: "Host"},
		},
		Pool: restaurants,
		Filters: domain.SessionFilters{
			PriceTiers: input.PriceTiers,
			MinRating:  minRatingVal,
			Cuisines:   input.Cuisines,
		},
		CreatedAt: time.Now(),
	}

	if err := uc.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	if uc.enrichUC != nil {
		go uc.enrichUC.ExecuteAsynchronously(session, uc.broadcaster)
	}

	return session, nil
}
