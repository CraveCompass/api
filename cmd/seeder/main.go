package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type OverpassResponse struct {
	Elements []struct {
		ID   int64             `json:"id"`
		Lat  float64           `json:"lat"`
		Lon  float64           `json:"lon"`
		Tags map[string]string `json:"tags"`
	} `json:"elements"`
}

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	lat := 44.84044
	lon := -0.58050
	radiusMeters := 5000

	fmt.Printf("Fetching restaurants within %dm of %f, %f...\n", radiusMeters, lat, lon)

	query := fmt.Sprintf(`
		[out:json][timeout:60];
		(
		  node["amenity"="restaurant"](around:%d,%f,%f);
		  node["amenity"="cafe"](around:%d,%f,%f);
		);
		out body;
	`, radiusMeters, lat, lon, radiusMeters, lat, lon)

	reqURL := "https://overpass.openstreetmap.fr/api/interpreter?data=" + url.QueryEscape(query)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "CraveCompass-Dev-Seeder/1.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to fetch data from OSM: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		log.Fatalf("Overpass API returned status %d: %s", resp.StatusCode, buf.String())
	}

	var osmData OverpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&osmData); err != nil {
		log.Fatalf("Failed to parse OSM JSON: %v", err)
	}

	fmt.Printf("Found %d locations. Inserting into PostGIS...\n", len(osmData.Elements))

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

		_, err := dbPool.Exec(context.Background(), insertSQL, internalID, el.ID, name, el.Lon, el.Lat, cuisines)
		if err != nil {
			log.Printf("Failed to insert %s: %v", name, err)
			continue
		}
		insertedCount++
	}

	fmt.Printf("Successfully inserted %d restaurants into the database!\n", insertedCount)
}
