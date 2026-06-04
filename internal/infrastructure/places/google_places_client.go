package places

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type GooglePlaceResult struct {
	ID               string
	Rating           *float64
	UserRatingsTotal *int
	PriceLevel       *int
	PhotoReference   *string
	FormattedAddress *string
}

type GooglePlacesClient struct {
	apiKey string
}

func NewGooglePlacesClient() *GooglePlacesClient {
	return &GooglePlacesClient{
		apiKey: os.Getenv("GOOGLE_PLACES_API_KEY"),
	}
}

type searchTextRequest struct {
	TextQuery    string       `json:"textQuery"`
	LocationBias locationBias `json:"locationBias"`
}

type locationBias struct {
	Circle circle `json:"circle"`
}

type circle struct {
	Center center  `json:"center"`
	Radius float64 `json:"radius"`
}

type center struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type searchTextResponse struct {
	Places []place `json:"places"`
}

type place struct {
	Id               string  `json:"id"`
	Rating           float64 `json:"rating"`
	UserRatingCount  int     `json:"userRatingCount"`
	PriceLevel       string  `json:"priceLevel"`
	FormattedAddress string  `json:"formattedAddress"`
	Photos           []photo `json:"photos"`
}

type photo struct {
	Name string `json:"name"`
}

func (c *GooglePlacesClient) FetchRestaurantDetails(ctx context.Context, name string, lat, lon float64) (*GooglePlaceResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_PLACES_API_KEY is not set")
	}

	url := "https://places.googleapis.com/v1/places:searchText"

	reqBody := searchTextRequest{
		TextQuery: name,
		LocationBias: locationBias{
			Circle: circle{
				Center: center{
					Latitude:  lat,
					Longitude: lon,
				},
				Radius: 100.0,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)

	req.Header.Set("X-Goog-FieldMask", "places.id,places.rating,places.userRatingCount,places.priceLevel,places.photos,places.formattedAddress")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google places API returned status: %d", resp.StatusCode)
	}

	var resData searchTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&resData); err != nil {
		return nil, err
	}

	if len(resData.Places) == 0 {
		return nil, fmt.Errorf("no places found for %s", name)
	}

	p := resData.Places[0]

	result := &GooglePlaceResult{
		ID:               p.Id,
		Rating:           &p.Rating,
		UserRatingsTotal: &p.UserRatingCount,
	}

	priceMap := map[string]int{
		"PRICE_LEVEL_INEXPENSIVE":    1,
		"PRICE_LEVEL_MODERATE":       2,
		"PRICE_LEVEL_EXPENSIVE":      3,
		"PRICE_LEVEL_VERY_EXPENSIVE": 4,
	}

	if val, ok := priceMap[p.PriceLevel]; ok {
		priceLvl := val
		result.PriceLevel = &priceLvl
	}

	if len(p.Photos) > 0 {
		result.PhotoReference = &p.Photos[0].Name
	}

	if p.FormattedAddress != "" {
		result.FormattedAddress = &p.FormattedAddress
	}

	return result, nil
}
