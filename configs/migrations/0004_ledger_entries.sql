-- 0004_ledger_entries.sql
-- Double-entry ledger entries

CREATE TABLE IF NOT EXISTS ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    debit_account_id UUID REFERENCES accounts(id),
    credit_account_id UUID REFERENCES accounts(id),
    amount BIGINT NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'NGN',
    transaction_id UUID,  -- points to transactions table
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
