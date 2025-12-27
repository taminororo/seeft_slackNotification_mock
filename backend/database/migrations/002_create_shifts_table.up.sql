CREATE TABLE IF NOT EXISTS shifts (
    id SERIAL PRIMARY KEY,
    year_id INTEGER NOT NULL,
    time_id INTEGER NOT NULL,
    date VARCHAR(50) NOT NULL,
    weather VARCHAR(50) NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(year_id, time_id, date, weather, user_id)
);

CREATE INDEX idx_shifts_user_id ON shifts(user_id);
CREATE INDEX idx_shifts_year_time_date_weather ON shifts(year_id, time_id, date, weather);

