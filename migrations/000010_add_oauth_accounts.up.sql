ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

CREATE TABLE oauth_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider    VARCHAR(50) NOT NULL CHECK (provider IN ('google', 'github')),
    provider_id VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (provider, provider_id),
    UNIQUE (user_id, provider)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider_id ON oauth_accounts(provider, provider_id);
