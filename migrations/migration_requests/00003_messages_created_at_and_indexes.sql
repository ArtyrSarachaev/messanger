-- +goose Up
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE messages RENAME COLUMN time_to_send TO created_at;

ALTER TABLE messages
    ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN created_at SET NOT NULL;

-- Ускоряет выборку переписки двух пользователей с сортировкой по времени.
CREATE INDEX IF NOT EXISTS idx_messages_sender_receiver_created
    ON messages (sender_id, receiver_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_sender_receiver_created;

ALTER TABLE messages
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE messages RENAME COLUMN created_at TO time_to_send;

ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;
