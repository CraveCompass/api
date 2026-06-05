package places

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type GooglePlaceResult struct {
	ID               string
	Rating           *float64
	UserRatingsTotal *int
	PriceLevel       *int
	PhotoReference   *string
	FormattedAddress *string
	Tags             []string
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
	TextQuery           string              `json:"textQuery"`
	LocationRestriction locationRestriction `json:"locationRestriction"`
}

type locationRestriction struct {
	Rectangle rectangle `json:"rectangle"`
}

type rectangle struct {
	Low  coordinates `json:"low"`
	High coordinates `json:"high"`
}

type coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type searchTextResponse struct {
	Places []place `json:"places"`
}

type place struct {
	Id               string   `json:"id"`
	Rating           float64  `json:"rating"`
	UserRatingCount  int      `json:"userRatingCount"`
	PriceLevel       string   `json:"priceLevel"`
	FormattedAddress string   `json:"formattedAddress"`
	Types            []string `json:"types"`
	Photos           []photo  `json:"photos"`
}

type photo struct {
	Name string `json:"name"`
}

func (c *GooglePlacesClient) FetchRestaurantDetails(ctx context.Context, name string, lat, lon float64) (*GooglePlaceResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_PLACES_API_KEY is not set")
	}

	url := "https://places.googleapis.com/v1/places:searchText"

	latOffset := 0.005
	lonOffset := 0.005

	reqBody := searchTextRequest{
		TextQuery: name,
		LocationRestriction: locationRestriction{
			Rectangle: rectangle{
				Low: coordinates{
					Latitude:  lat - latOffset,
					Longitude: lon - lonOffset,
				},
				High: coordinates{
					Latitude:  lat + latOffset,
					Longitude: lon + lonOffset,
				},
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

	req.Header.Set("X-Goog-FieldMask", "places.id,places.rating,places.userRatingCount,places.priceLevel,places.photos,places.formattedAddress,places.types")

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

	for _, t := range p.Types {
		if t != "restaurant" && t != "food" && t != "point_of_interest" && t != "establishment" {
			cleanTag := strings.ReplaceAll(t, "_", " ")
			result.Tags = append(result.Tags, cleanTag)
		}
	}

	return result, nil
}
