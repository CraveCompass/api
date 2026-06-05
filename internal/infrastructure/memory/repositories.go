package memory

import (
	"context"

	"github.com/CraveCompass/api/internal/domain"
)

type InMemoryRestaurantRepo struct{}

func NewInMemoryRestaurantRepo() *InMemoryRestaurantRepo {
	return &InMemoryRestaurantRepo{}
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
