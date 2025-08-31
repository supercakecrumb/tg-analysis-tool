-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Organizations table
CREATE TABLE orgs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Organization members
CREATE TABLE org_members (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, org_id)
);

-- Telegram accounts
CREATE TABLE telegram_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    phone VARCHAR(20) NOT NULL,
    session_blob_enc BYTEA NOT NULL,
    session_kek_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    last_sync_at TIMESTAMPTZ,
    UNIQUE(org_id, phone)
);

-- Dialogs (chats)
CREATE TABLE dialogs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID REFERENCES telegram_accounts(id) ON DELETE CASCADE,
    tg_peer_type VARCHAR(20) NOT NULL CHECK (tg_peer_type IN ('user', 'chat', 'channel')),
    tg_peer_id BIGINT NOT NULL,
    title VARCHAR(255),
    is_excluded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, account_id, tg_peer_type, tg_peer_id)
);

-- Sync cursors for incremental updates
CREATE TABLE sync_cursors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    account_id UUID REFERENCES telegram_accounts(id) ON DELETE CASCADE,
    dialog_id UUID REFERENCES dialogs(id) ON DELETE CASCADE,
    pts INT,
    qts INT,
    date INT,
    seq INT,
    max_message_id BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(account_id, dialog_id)
);

-- Shares for public dashboards
CREATE TABLE shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN ('dashboard', 'report')),
    resource_id UUID NOT NULL,
    visibility VARCHAR(20) NOT NULL CHECK (visibility IN ('private', 'org', 'public')),
    token VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    view_count INT NOT NULL DEFAULT 0,
    last_viewed_at TIMESTAMPTZ
);

-- API keys (optional feature)
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID REFERENCES orgs(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Sessions for authentication
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_orgs_slug ON orgs(slug);
CREATE INDEX idx_telegram_accounts_user_id ON telegram_accounts(user_id);
CREATE INDEX idx_telegram_accounts_org_id ON telegram_accounts(org_id);
CREATE INDEX idx_dialogs_account_id ON dialogs(account_id);
CREATE INDEX idx_dialogs_org_id ON dialogs(org_id);
CREATE INDEX idx_sync_cursors_account_id ON sync_cursors(account_id);
CREATE INDEX idx_shares_token ON shares(token);
CREATE INDEX idx_shares_expires_at ON shares(expires_at);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orgs_updated_at BEFORE UPDATE ON orgs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dialogs_updated_at BEFORE UPDATE ON dialogs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security Policies
ALTER TABLE telegram_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE dialogs ENABLE ROW LEVEL SECURITY;
ALTER TABLE sync_cursors ENABLE ROW LEVEL SECURITY;
ALTER TABLE shares ENABLE ROW LEVEL SECURITY;

-- RLS Policies (to be configured based on application context)
-- Example policy for telegram_accounts
CREATE POLICY telegram_accounts_org_isolation ON telegram_accounts
    FOR ALL
    USING (org_id IN (
        SELECT org_id FROM org_members 
        WHERE user_id = current_setting('app.current_user_id')::UUID
    ));

-- Comments for documentation
COMMENT ON TABLE users IS 'Application users';
COMMENT ON TABLE orgs IS 'Organizations for multi-tenancy';
COMMENT ON TABLE telegram_accounts IS 'Telegram accounts linked to users';
COMMENT ON TABLE dialogs IS 'Telegram chats/channels/groups';
COMMENT ON TABLE sync_cursors IS 'Sync state for incremental updates';
COMMENT ON TABLE shares IS 'Shared dashboards and reports';
COMMENT ON COLUMN telegram_accounts.session_blob_enc IS 'AES-GCM encrypted MTProto session';
COMMENT ON COLUMN telegram_accounts.session_kek_version IS 'Key encryption key version for rotation';