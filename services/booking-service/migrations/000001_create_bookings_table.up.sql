CREATE TABLE bookings (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    car_id TEXT NOT NULL,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    total_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT bookings_valid_dates CHECK (end_date > start_date),
    CONSTRAINT bookings_non_negative_total_price CHECK (total_price >= 0),
    CONSTRAINT bookings_valid_status CHECK (
        status IN ('pending', 'confirmed', 'cancelled', 'active', 'completed')
    )
);

CREATE INDEX bookings_user_id_created_at_idx
    ON bookings (user_id, created_at DESC);

CREATE INDEX bookings_car_id_dates_idx
    ON bookings (car_id, start_date, end_date);
