-- 0006_reconciliation_logs.sql
-- Logs used to verify ledger correctness

CREATE TABLE IF NOT EXISTS reconciliation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL, -- 'transactions', 'settlements', etc
    entity_id UUID,
    expected_amount BIGINT,
    actual_amount BIGINT,
    status VARCHAR(20) NOT NULL, -- matched / mismatch / resolved
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
