CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL,
    original_url TEXT NOT NULL,
    intent_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL,
    user_id TEXT NULL,
    CONSTRAINT links_code_key UNIQUE (code),
    CONSTRAINT links_intent_key_key UNIQUE (intent_key)
);

CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_links_user_id ON links (user_id);
