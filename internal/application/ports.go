package application

import (
	"context"

	"github.com/CraveCompass/api/internal/domain"
	"github.com/CraveCompass/api/internal/infrastructure/places"
)

type RestaurantRepository interface {
	GetByLocation(ctx context.Context, lat, lon float64, radiusMeters int, priceTiers []int, minRating *float64, cuisines []string) ([]domain.Restaurant, error)
	FetchAndSaveFromOSM(ctx context.Context, lat, lon float64, radiusMeters int) error
	UpdateGooglePlacesData(ctx context.Context, id string, googlePlaceID *string, rating *float64, userRatingsTotal *int, priceLevel *int, photoReference *string, formattedAddress *string, extraTags []string, openingHours []string) error
	GetUniqueCuisines(ctx context.Context) ([]string, error)
}

type SessionRepository interface {
	Save(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
}

type PlacesClient interface {
	FetchRestaurantDetails(ctx context.Context, name string, lat, lon float64) (*places.GooglePlaceResult, error)
}

type SessionBroadcaster interface {
	Broadcast(sessionID string, message interface{})
}
