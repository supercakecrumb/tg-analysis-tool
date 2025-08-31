-- Create database if not exists
CREATE DATABASE IF NOT EXISTS tg_analytics;

USE tg_analytics;

-- Main messages table
CREATE TABLE IF NOT EXISTS messages (
    tenant_id UUID,
    account_id UUID,
    chat_id UUID,
    message_id UInt64,
    date DateTime('Europe/Moscow'),
    from_id UInt64,
    to_thread_id Nullable(UInt64),
    type Enum8(
        'text' = 1,
        'sticker' = 2,
        'photo' = 3,
        'video' = 4,
        'voice' = 5,
        'file' = 6,
        'service' = 7
    ),
    text Nullable(String),
    sticker_emoji Nullable(String),
    has_media UInt8,
    media_bytes Nullable(UInt64),
    voice_secs Nullable(UInt16),
    video_secs Nullable(UInt16),
    reply_to Nullable(UInt64),
    entities Array(String),
    emojis Array(String),
    is_deleted UInt8 DEFAULT 0,
    updated_at DateTime DEFAULT now()
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, chat_id, date, message_id)
SETTINGS index_granularity = 8192;

-- Tokenized words for text analysis
CREATE TABLE IF NOT EXISTS words (
    tenant_id UUID,
    message_id UInt64,
    chat_id UUID,
    word String,
    position UInt16,
    date Date
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, word, date)
SETTINGS index_granularity = 8192;

-- Materialized view for daily message aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS agg_daily_messages
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (tenant_id, chat_id, type, day)
AS SELECT
    tenant_id,
    chat_id,
    type,
    toDate(date) as day,
    count() as message_count,
    sum(voice_secs) as total_voice_secs,
    sum(video_secs) as total_video_secs,
    uniqExact(from_id) as unique_senders
FROM messages
WHERE is_deleted = 0
GROUP BY tenant_id, chat_id, type, day;

-- Materialized view for hourly heatmap
CREATE MATERIALIZED VIEW IF NOT EXISTS agg_hourly_heatmap
ENGINE = SummingMergeTree()
ORDER BY (tenant_id, weekday, hour)
AS SELECT
    tenant_id,
    toDayOfWeek(date) as weekday,
    toHour(date) as hour,
    count() as message_count,
    uniqExact(from_id) as unique_senders
FROM messages
WHERE is_deleted = 0
GROUP BY tenant_id, weekday, hour;

-- Materialized view for response time analysis
CREATE MATERIALIZED VIEW IF NOT EXISTS agg_response_times
ENGINE = AggregatingMergeTree()
ORDER BY (tenant_id, chat_id, from_id, to_id)
AS SELECT
    tenant_id,
    chat_id,
    from_id,
    from_id as to_id,  -- Will be updated with proper logic
    quantileExact(0.5)(response_time_seconds) as median_response_time,
    quantileExact(0.95)(response_time_seconds) as p95_response_time,
    avg(response_time_seconds) as avg_response_time,
    count() as response_count
FROM (
    SELECT 
        tenant_id,
        chat_id,
        from_id,
        date,
        message_id,
        -- Calculate response time to previous message from different user
        dateDiff('second', 
            lagInFrame(date) OVER (PARTITION BY tenant_id, chat_id ORDER BY date),
            date
        ) as response_time_seconds
    FROM messages
    WHERE is_deleted = 0
)
WHERE response_time_seconds > 0 AND response_time_seconds < 86400  -- Within 24 hours
GROUP BY tenant_id, chat_id, from_id, to_id;

-- Materialized view for top words daily
CREATE MATERIALIZED VIEW IF NOT EXISTS agg_top_words_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(day)
ORDER BY (tenant_id, day, word)
AS SELECT
    tenant_id,
    toDate(date) as day,
    word,
    count() as word_count
FROM words
WHERE length(word) >= 3  -- Minimum word length
GROUP BY tenant_id, day, word;

-- Table for storing emoji usage
CREATE TABLE IF NOT EXISTS emoji_usage (
    tenant_id UUID,
    chat_id UUID,
    message_id UInt64,
    emoji String,
    count UInt8,
    date DateTime
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, emoji, date)
SETTINGS index_granularity = 8192;

-- Materialized view for emoji statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS agg_emoji_stats
ENGINE = SummingMergeTree()
ORDER BY (tenant_id, emoji)
AS SELECT
    tenant_id,
    emoji,
    sum(count) as total_count,
    uniqExact(chat_id) as unique_chats,
    uniqExact(message_id) as unique_messages
FROM emoji_usage
GROUP BY tenant_id, emoji;

-- Table for media metadata
CREATE TABLE IF NOT EXISTS media_metadata (
    tenant_id UUID,
    message_id UInt64,
    chat_id UUID,
    media_type String,
    file_size Nullable(UInt64),
    mime_type Nullable(String),
    width Nullable(UInt32),
    height Nullable(UInt32),
    duration_seconds Nullable(UInt32),
    date DateTime
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, chat_id, date)
SETTINGS index_granularity = 8192;

-- System table for tracking sync progress
CREATE TABLE IF NOT EXISTS sync_log (
    tenant_id UUID,
    account_id UUID,
    sync_id UUID,
    sync_type Enum8('full' = 1, 'incremental' = 2),
    status Enum8('started' = 1, 'in_progress' = 2, 'completed' = 3, 'failed' = 4),
    started_at DateTime,
    completed_at Nullable(DateTime),
    messages_processed UInt64,
    error_message Nullable(String),
    metadata String  -- JSON for additional data
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(started_at)
ORDER BY (tenant_id, account_id, started_at)
SETTINGS index_granularity = 8192;

-- Create indexes for better query performance
ALTER TABLE messages ADD INDEX idx_from_id (from_id) TYPE minmax GRANULARITY 4;
ALTER TABLE messages ADD INDEX idx_type (type) TYPE set(10) GRANULARITY 2;
ALTER TABLE messages ADD INDEX idx_has_media (has_media) TYPE set(2) GRANULARITY 1;

ALTER TABLE words ADD INDEX idx_word_bloom (word) TYPE bloom_filter(0.01) GRANULARITY 1;

-- Comments for documentation
COMMENT ON TABLE messages IS 'Main table storing all Telegram messages';
COMMENT ON TABLE words IS 'Tokenized words from messages for text analysis';
COMMENT ON TABLE emoji_usage IS 'Emoji usage tracking for analytics';
COMMENT ON TABLE media_metadata IS 'Metadata for media files (photos, videos, etc)';
COMMENT ON TABLE sync_log IS 'Log of synchronization operations';

COMMENT ON COLUMN messages.tenant_id IS 'Organization ID for multi-tenancy';
COMMENT ON COLUMN messages.type IS 'Message type: text, sticker, photo, video, voice, file, service';
COMMENT ON COLUMN messages.entities IS 'Array of entities like links, mentions, hashtags';
COMMENT ON COLUMN messages.is_deleted IS 'Soft delete flag for messages';