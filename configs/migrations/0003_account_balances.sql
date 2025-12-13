-- 0003_account_balances.sql

CREATE TABLE IF NOT EXISTS account_balances (
    account_id UUID PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
