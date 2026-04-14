CREATE TABLE IF NOT EXISTS message_reads (
    id SERIAL PRIMARY KEY,

    notification_id INTEGER NOT NULL
                    REFERENCES notifications(id) ON DELETE CASCADE,

    chat_id TEXT NOT NULL
                REFERENCES users(chat_id) ON DELETE CASCADE,

    read_at TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(notification_id, chat_id)
    );