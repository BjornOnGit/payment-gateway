-- 0007_audit_logs.sql

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID, -- admin or system user
    action VARCHAR(100) NOT NULL,
    entity VARCHAR(50),
    entity_id UUID,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
