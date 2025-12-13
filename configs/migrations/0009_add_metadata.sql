ALTER TABLE settlements
ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}':: jsonb;

ALTER TABLE transactions
ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}':: jsonb;

ALTER TABLE reconciliation_logs
ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}':: jsonb;

ALTER TABLE ledger_entries
ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}':: jsonb;