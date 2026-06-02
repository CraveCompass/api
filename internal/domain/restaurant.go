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
	Rating      float64   `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
}
