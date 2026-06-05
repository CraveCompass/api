package domain

import "time"

type Restaurant struct {
	ID          string    `json:"id"`
	OSMID       int64     `json:"osm_id"`
	Name        string    `json:"name"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CuisineTags []string  `json:"cuisine_tags"`
	PriceTier   int       `json:"price_tier"`
	CreatedAt   time.Time `json:"created_at"`

	GooglePlaceID    *string  `json:"google_place_id,omitempty"`
	Rating           *float64 `json:"rating,omitempty"`
	UserRatingsTotal *int     `json:"user_ratings_total,omitempty"`
	PriceLevel       *int     `json:"price_level,omitempty"`
	PhotoReference   *string  `json:"photo_reference,omitempty"`
	FormattedAddress *string  `json:"formatted_address,omitempty"`
	OpeningHours     []string `json:"opening_hours,omitempty"`
}
