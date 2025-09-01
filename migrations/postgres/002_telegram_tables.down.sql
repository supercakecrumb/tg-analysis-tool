-- Drop triggers
DROP TRIGGER IF EXISTS update_telegram_auth_updated_at ON telegram_auth;
DROP TRIGGER IF EXISTS update_telegram_chats_updated_at ON telegram_chats;
DROP TRIGGER IF EXISTS update_sync_progress_updated_at ON sync_progress;

-- Drop functions
DROP FUNCTION IF EXISTS create_monthly_partition();

-- Drop tables (partitions will be dropped automatically)
DROP TABLE IF EXISTS sync_progress;
DROP TABLE IF EXISTS telegram_messages;
DROP TABLE IF EXISTS telegram_chats;
DROP TABLE IF EXISTS telegram_auth;