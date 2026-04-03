CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    chat_id     VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255),
    role        VARCHAR(20)  DEFAULT 'employee',
    gender      VARCHAR(6),
    department  VARCHAR(30),
    last_seen   TIMESTAMP,
    created_at  TIMESTAMP DEFAULT NOW()
    );