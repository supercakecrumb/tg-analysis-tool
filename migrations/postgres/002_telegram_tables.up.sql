-- Telegram authentication table
CREATE TABLE IF NOT EXISTS telegram_auth (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    phone_number TEXT,
    session_data BYTEA, -- Encrypted session data
    is_authorized BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Telegram chats table
CREATE TABLE IF NOT EXISTS telegram_chats (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('private', 'group', 'supergroup', 'channel')),
    username TEXT,
    member_count INTEGER DEFAULT 0,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, chat_id)
);

-- Telegram messages table (partitioned by date)
CREATE TABLE IF NOT EXISTS telegram_messages (
    id BIGSERIAL,
    user_id TEXT NOT NULL,
    chat_id BIGINT NOT NULL,
    message_id INTEGER NOT NULL,
    from_id BIGINT,
    text TEXT,
    date TIMESTAMP NOT NULL,
    reply_to_id INTEGER,
    forward_id BIGINT,
    media_type TEXT,
    has_emoji BOOLEAN DEFAULT false,
    word_count INTEGER DEFAULT 0,
    is_outgoing BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, chat_id, message_id, date)
) PARTITION BY RANGE (date);

-- Create partitions for the last 12 months and next month
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN -12..1 LOOP
        start_date := DATE_TRUNC('month', CURRENT_DATE + (i || ' months')::INTERVAL);
        end_date := start_date + INTERVAL '1 month';
        partition_name := 'telegram_messages_' || TO_CHAR(start_date, 'YYYY_MM');
        
        EXECUTE format('
            CREATE TABLE IF NOT EXISTS %I PARTITION OF telegram_messages
            FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;

-- Sync progress table
CREATE TABLE IF NOT EXISTS sync_progress (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('idle', 'syncing', 'completed', 'failed')),
    total_chats INTEGER DEFAULT 0,
    synced_chats INTEGER DEFAULT 0,
    total_messages BIGINT DEFAULT 0,
    synced_messages BIGINT DEFAULT 0,
    current_chat TEXT,
    last_error TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_telegram_auth_user_id ON telegram_auth(user_id);
CREATE INDEX idx_telegram_chats_user_id ON telegram_chats(user_id);
CREATE INDEX idx_telegram_chats_last_sync ON telegram_chats(user_id, last_sync_at DESC);
CREATE INDEX idx_telegram_messages_chat ON telegram_messages(user_id, chat_id, date DESC);
CREATE INDEX idx_telegram_messages_from ON telegram_messages(user_id, from_id, date DESC);
CREATE INDEX idx_telegram_messages_text_search ON telegram_messages USING gin(to_tsvector('english', text));

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_telegram_auth_updated_at BEFORE UPDATE ON telegram_auth
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_telegram_chats_updated_at BEFORE UPDATE ON telegram_chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_progress_updated_at BEFORE UPDATE ON sync_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically create monthly partitions
CREATE OR REPLACE FUNCTION create_monthly_partition()
RETURNS void AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    start_date := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'telegram_messages_' || TO_CHAR(start_date, 'YYYY_MM');
    
    EXECUTE format('
        CREATE TABLE IF NOT EXISTS %I PARTITION OF telegram_messages
        FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;

-- Schedule monthly partition creation (requires pg_cron extension)
-- SELECT cron.schedule('create-monthly-partition', '0 0 1 * *', 'SELECT create_monthly_partition()');