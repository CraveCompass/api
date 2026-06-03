package memory

import (
	"context"
	"time"

	"github.com/CraveCompass/api/internal/domain"
)

type InMemoryRestaurantRepo struct{}

func NewInMemoryRestaurantRepo() *InMemoryRestaurantRepo {
	return &InMemoryRestaurantRepo{}
}

func (r *InMemoryRestaurantRepo) GetByLocation(ctx context.Context, lat, lon float64, radiusMeters int) ([]domain.Restaurant, error) {
	return []domain.Restaurant{
		{
			ID:          "rest_1",
			Name:        "Luigi's Pizza",
			Latitude:    lat + 0.001,
			Longitude:   lon + 0.001,
			CuisineTags: []string{"Italian", "Pizza"},
			PriceTier:   2,
			Rating:      4.8,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "rest_2",
			Name:        "Spicy Thai Street",
			Latitude:    lat - 0.002,
			Longitude:   lon + 0.001,
			CuisineTags: []string{"Thai", "Spicy"},
			PriceTier:   1,
			Rating:      4.5,
			CreatedAt:   time.Now(),
		},
	}, nil
}

type InMemorySessionRepo struct {
	store map[string]*domain.Session
}

func NewInMemorySessionRepo() *InMemorySessionRepo {
	return &InMemorySessionRepo{
		store: make(map[string]*domain.Session),
	}
}

func (r *InMemorySessionRepo) Save(ctx context.Context, session *domain.Session) error {
	r.store[session.ID] = session
	return nil
}

func (r *InMemorySessionRepo) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	if session, exists := r.store[id]; exists {
		return session, nil
	}
	return nil, nil
}

func (r *InMemoryRestaurantRepo) FetchAndSaveFromOSM(ctx context.Context, lat, lon float64, radiusMeters int) error {
	return nil
}
