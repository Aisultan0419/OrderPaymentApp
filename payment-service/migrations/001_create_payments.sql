-- migrations/001_create_payments.sql
-- Run this against your payments_db database.

CREATE TABLE IF NOT EXISTS payments (
    id             VARCHAR(36)  PRIMARY KEY,
    order_id       VARCHAR(36)  NOT NULL,
    transaction_id VARCHAR(36)  NOT NULL DEFAULT '',
    amount         BIGINT       NOT NULL CHECK (amount > 0),
    status         VARCHAR(20)  NOT NULL DEFAULT 'Authorized'
);

CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);
