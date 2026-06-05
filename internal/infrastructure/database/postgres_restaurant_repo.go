package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/CraveCompass/api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRestaurantRepo struct {
	db *pgxpool.Pool
}

type OverpassResponse struct {
	Elements []struct {
		ID   int64             `json:"id"`
		Lat  float64           `json:"lat"`
		Lon  float64           `json:"lon"`
		Tags map[string]string `json:"tags"`
	} `json:"elements"`
}

func NewPostgresRestaurantRepo(db *pgxpool.Pool) *PostgresRestaurantRepo {
	return &PostgresRestaurantRepo{
		db: db,
	}
}

func (r *PostgresRestaurantRepo) GetByLocation(ctx context.Context, lat, lon float64, radiusMeters int, priceTiers []int, minRating *float64, cuisines []string) ([]domain.Restaurant, error) {
	query := `
		SELECT id, osm_id, name, ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon, cuisine_tags, google_place_id, rating, user_ratings_total, price_level, photo_reference, formatted_address, opening_hours
		FROM restaurants
		WHERE ST_DWithin(
			location, 
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, 
			$3
		)
	`

	args := []interface{}{lon, lat, radiusMeters}
	argCount := 3

	if len(priceTiers) > 0 {
		argCount++
		query += fmt.Sprintf(" AND (price_level = ANY($%d) OR price_tier = ANY($%d) OR price_level IS NULL)", argCount, argCount)
		args = append(args, priceTiers)
	}

	if minRating != nil && *minRating > 0 {
		argCount++
		query += fmt.Sprintf(" AND (rating >= $%d OR rating IS NULL)", argCount)
		args = append(args, *minRating)
	}

	if len(cuisines) > 0 {
		argCount++
		query += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM unnest(cuisine_tags) db_tag, unnest($%d::text[]) filter_tag 
			WHERE db_tag ILIKE '%%' || filter_tag || '%%'
		)`, argCount)
		args = append(args, cuisines)
	}

	query += " LIMIT 30;"

	rows, err := r.db.Query(ctx, query, args...)
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
			&rest.GooglePlaceID,
			&rest.Rating,
			&rest.UserRatingsTotal,
			&rest.PriceLevel,
			&rest.PhotoReference,
			&rest.FormattedAddress,
			&rest.OpeningHours,
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

func (r *PostgresRestaurantRepo) FetchAndSaveFromOSM(ctx context.Context, lat, lon float64, radiusMeters int) error {
	log.Printf("CACHE MISS: Fetching live OSM data within %dm of %f, %f...\n", radiusMeters, lat, lon)

	query := fmt.Sprintf(`
		[out:json][timeout:25];
		(
		  node["amenity"="restaurant"](around:%d,%f,%f);
		  node["amenity"="cafe"](around:%d,%f,%f);
		);
		out body;
	`, radiusMeters, lat, lon, radiusMeters, lat, lon)

	reqURL := "https://overpass.openstreetmap.fr/api/interpreter?data=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "CraveCompass-Production/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch data from OSM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		return fmt.Errorf("Overpass API returned status %d: %s", resp.StatusCode, buf.String())
	}

	var osmData OverpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&osmData); err != nil {
		return fmt.Errorf("failed to parse OSM JSON: %w", err)
	}

	insertedCount := 0
	for _, el := range osmData.Elements {
		name, exists := el.Tags["name"]
		if !exists || name == "" {
			continue
		}

		cuisines := []string{}
		if cuisineStr, ok := el.Tags["cuisine"]; ok {
			cuisines = strings.Split(cuisineStr, ";")
		}

		internalID := fmt.Sprintf("osm_%d", el.ID)

		insertSQL := `
			INSERT INTO restaurants (id, osm_id, name, location, cuisine_tags)
			VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326)::geography, $6)
			ON CONFLICT (osm_id) DO UPDATE 
			SET name = EXCLUDED.name, cuisine_tags = EXCLUDED.cuisine_tags;
		`

		_, err := r.db.Exec(ctx, insertSQL, internalID, el.ID, name, el.Lon, el.Lat, cuisines)
		if err == nil {
			insertedCount++
		}
	}

	log.Printf("Successfully cached %d new restaurants to PostGIS!", insertedCount)
	return nil
}

func (r *PostgresRestaurantRepo) UpdateGooglePlacesData(ctx context.Context, id string, googlePlaceID *string, rating *float64, userRatingsTotal *int, priceLevel *int, photoReference *string, formattedAddress *string, extraTags []string, openingHours []string) error {
	query := `
		UPDATE restaurants
		SET google_place_id = $2, 
		    rating = $3, 
		    user_ratings_total = $4, 
		    price_level = $5, 
		    photo_reference = $6, 
		    formatted_address = $7
			cuisine_tags = array_cat(cuisine_tags, $8)
			opening_hours = $9
		WHERE id = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		id,
		googlePlaceID,
		rating,
		userRatingsTotal,
		priceLevel,
		photoReference,
		formattedAddress,
		extraTags,
		openingHours,
	)

	return err
}

func (r *PostgresRestaurantRepo) GetUniqueCuisines(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT initcap(replace(tag, '_', ' '))
		FROM (SELECT unnest(cuisine_tags) AS tag FROM restaurants) sub 
		WHERE tag != '' 
		ORDER BY 1 
		LIMIT 50;
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err == nil {
			tags = append(tags, tag)
		}
	}
	return tags, nil
}
