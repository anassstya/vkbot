CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL               PRIMARY KEY,
    title                   VARCHAR(255) NOT NULL,
    description TEXT        NOT NULL,

    recipients_department   VARCHAR(50)
                            CHECK (recipients_department IN ('it', 'analytics', 'design', 'management', 'all')),

    recipients_gender       VARCHAR(10)
                            CHECK (recipients_gender IN ('male', 'female', 'all')),

    status                  VARCHAR(20) DEFAULT 'scheduled',

    created_by              TEXT NOT NULL DEFAULT '',
    created_at              TIMESTAMPTZ  DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  DEFAULT NOW(),

    type                    VARCHAR(20) NOT NULL
                            CHECK (type IN ('one_time', 'event_trigger', 'recurring')),

    time_to_send            TIMESTAMPTZ,
    trigger_event           VARCHAR(50),
    trigger_user_id         TEXT,
    recurrence              VARCHAR(50),
    next_send               TIMESTAMPTZ,
    last_sent               TIMESTAMPTZ,

    CHECK (
        (type = 'one_time' AND time_to_send IS NOT NULL AND recurrence IS NULL AND trigger_event IS NULL)
            OR
        (type = 'event_trigger' AND trigger_event IS NOT NULL AND time_to_send IS NULL)
            OR
        (type = 'recurring' AND recurrence IS NOT NULL)
    )
);