package database

import (
	"context"
	"fmt"

	"github.com/CraveCompass/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRestaurantRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRestaurantRepo(db *pgxpool.Pool) *PostgresRestaurantRepo {
	return &PostgresRestaurantRepo{
		db: db,
	}
}

func (r *PostgresRestaurantRepo) GetByLocation(ctx context.Context, lat, lon float64, radiusMeters int) ([]domain.Restaurant, error) {
	query := `
		SELECT id, osm_id, name, ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon, cuisine_tags, price_tier, rating
		FROM restaurants
		WHERE ST_DWithin(
			location, 
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, 
			$3
		)
		LIMIT 30; -- Cap the deck size to prevent massive payloads
	`

	rows, err := r.db.Query(ctx, query, lon, lat, radiusMeters)
	if err != nil {
		return nil, fmt.Errorf("failed to query restaurants: %w", err)
	}
	defer rows.Close()

	var restaurants []domain.Restaurant

	for rows.Next() {
		var rest domain.Restaurant
		if err := rows.Scan(
			&rest.ID,
			&rest.OSMID,
			&rest.Name,
			&rest.Latitude,
			&rest.Longitude,
			&rest.CuisineTags,
			&rest.PriceTier,
			&rest.Rating,
		); err != nil {
			return nil, fmt.Errorf("failed to scan restaurant row: %w", err)
		}
		restaurants = append(restaurants, rest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return restaurants, nil
}
