-- Enable the PostGIS extension for spatial queries
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS restaurants (
    id VARCHAR(50) PRIMARY KEY,
    osm_id BIGINT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    location GEOGRAPHY(Point, 4326) NOT NULL, 
    cuisine_tags TEXT[] NOT NULL DEFAULT '{}',
    price_tier INT DEFAULT 1,
    rating NUMERIC(3, 2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create a spatial index to make radius lookups blazingly fast
CREATE INDEX IF NOT EXISTS idx_restaurants_location ON restaurants USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_restaurants_tags ON restaurants USING GIN (cuisine_tags);