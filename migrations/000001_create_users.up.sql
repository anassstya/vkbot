CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    chat_id     TEXT UNIQUE NOT NULL,
    name        VARCHAR(255),

    role        VARCHAR(20)  DEFAULT 'employee'
                CHECK (role IN ('employee', 'admin')),

    gender      VARCHAR(10)
                CHECK (gender IN ('male', 'female')),

    department  VARCHAR(30)
                CHECK (department IN ('it', 'analytics', 'design', 'management', 'all')),

    last_seen   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW()
    );