-- 0002_accounts.sql
-- Account table 

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL,  -- user or merchant
    owner_type VARCHAR(20) NOT NULL,  -- 'user', 'merchant', 'system'
    account_type VARCHAR(20) NOT NULL DEFAULT 'wallet',
    currency VARCHAR(10) NOT NULL DEFAULT 'NGN',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
