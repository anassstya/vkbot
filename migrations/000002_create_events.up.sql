CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    department TEXT,
    gender VARCHAR(10),
    status TEXT NOT NULL DEFAULT 'scheduled',
    created_by TEXT,
    time_to_send TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);