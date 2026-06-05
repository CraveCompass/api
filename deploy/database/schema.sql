CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS restaurants (
    id VARCHAR(50) PRIMARY KEY,
    osm_id BIGINT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    location GEOGRAPHY(Point, 4326) NOT NULL, 
    cuisine_tags TEXT[] NOT NULL DEFAULT '{}',
    price_tier INT DEFAULT 1,
    rating NUMERIC(3, 2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_restaurants_location ON restaurants USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_restaurants_tags ON restaurants USING GIN (cuisine_tags);

ALTER TABLE restaurants
ADD COLUMN IF NOT EXISTS google_place_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS rating NUMERIC(3, 2),
ADD COLUMN IF NOT EXISTS user_ratings_total INT,
ADD COLUMN IF NOT EXISTS price_level INT,
ADD COLUMN IF NOT EXISTS photo_reference TEXT,
ADD COLUMN IF NOT EXISTS formatted_address TEXT;

ALTER TABLE restaurants
ADD COLUMN IF NOT EXISTS opening_hours TEXT[] DEFAULT '{}';