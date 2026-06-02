package application

import (
	"context"

	"github.com/CraveCompass/api/internal/domain"
)

type RestaurantRepository interface {
	GetByLocation(ctx context.Context, lat, lon float64, radiusMeters int) ([]domain.Restaurant, error)
}

type SessionRepository interface {
	Save(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
}
