-- Drop RLS policies
DROP POLICY IF EXISTS telegram_accounts_org_isolation ON telegram_accounts;

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_orgs_updated_at ON orgs;
DROP TRIGGER IF EXISTS update_dialogs_updated_at ON dialogs;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (automatically dropped with tables, but being explicit)
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token_hash;
DROP INDEX IF EXISTS idx_shares_expires_at;
DROP INDEX IF EXISTS idx_shares_token;
DROP INDEX IF EXISTS idx_sync_cursors_account_id;
DROP INDEX IF EXISTS idx_dialogs_org_id;
DROP INDEX IF EXISTS idx_dialogs_account_id;
DROP INDEX IF EXISTS idx_telegram_accounts_org_id;
DROP INDEX IF EXISTS idx_telegram_accounts_user_id;
DROP INDEX IF EXISTS idx_orgs_slug;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS shares CASCADE;
DROP TABLE IF EXISTS sync_cursors CASCADE;
DROP TABLE IF EXISTS dialogs CASCADE;
DROP TABLE IF EXISTS telegram_accounts CASCADE;
DROP TABLE IF EXISTS org_members CASCADE;
DROP TABLE IF EXISTS orgs CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";