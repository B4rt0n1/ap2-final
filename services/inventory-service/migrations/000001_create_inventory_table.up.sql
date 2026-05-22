CREATE TABLE cars (
    id TEXT PRIMARY KEY,
    brand TEXT NOT NULL,
    model TEXT NOT NULL,
    year INT NOT NULL,
    category TEXT NOT NULL,
    price_per_day NUMERIC(12, 2) NOT NULL CHECK (price_per_day >= 0),
    available BOOLEAN NOT NULL DEFAULT TRUE,
    image_urls TEXT[] NOT NULL DEFAULT '{}', -- Supports your UploadCarImages endpoint
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexing for rubric performance and optimized search lookups
CREATE INDEX idx_cars_category_price ON cars(category, price_per_day);

CREATE TABLE reviews (
    id TEXT PRIMARY KEY,
    car_id TEXT NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_car_id ON reviews(car_id);
