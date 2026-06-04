package application

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/CraveCompass/api/internal/domain"
)

type EnrichRestaurantsUseCase struct {
	restaurantRepo RestaurantRepository
	placesClient   PlacesClient
}

func NewEnrichRestaurantsUseCase(restaurantRepo RestaurantRepository, placesClient PlacesClient) *EnrichRestaurantsUseCase {
	return &EnrichRestaurantsUseCase{
		restaurantRepo: restaurantRepo,
		placesClient:   placesClient,
	}
}

func (uc *EnrichRestaurantsUseCase) ExecuteAsynchronously(session *domain.Session, broadcaster SessionBroadcaster) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 2)

	for i := range session.Pool {
		if session.Pool[i].GooglePlaceID != nil && *session.Pool[i].GooglePlaceID != "" {
			continue
		}

		wg.Add(1)

		go func(index int) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			time.Sleep(250 * time.Millisecond)

			rest := session.Pool[index]

			details, err := uc.placesClient.FetchRestaurantDetails(ctx, rest.Name, rest.Latitude, rest.Longitude)
			if err != nil {
				log.Printf("[Enrichment Worker] Failed to fetch details for %s: %v", rest.Name, err)
				return
			}

			_ = uc.restaurantRepo.UpdateGooglePlacesData(
				ctx, rest.ID, &details.ID, details.Rating, details.UserRatingsTotal, details.PriceLevel, details.PhotoReference, details.FormattedAddress, // <-- Pass it here
			)

			session.Pool[index].GooglePlaceID = &details.ID

			session.Pool[index].Rating = details.Rating
			session.Pool[index].UserRatingsTotal = details.UserRatingsTotal
			session.Pool[index].PriceLevel = details.PriceLevel
			session.Pool[index].PhotoReference = details.PhotoReference
			session.Pool[index].FormattedAddress = details.FormattedAddress

			if broadcaster != nil {
				broadcaster.Broadcast(session.ID, map[string]interface{}{
					"event":    "SESSION_UPDATED",
					"session":  session,
					"is_match": session.MatchedID != "",
				})
			}
		}(i)
	}

	wg.Wait()
	log.Println("[Enrichment Worker] Finished updating session pool.")
}
