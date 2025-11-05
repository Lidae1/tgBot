-- +goose Up
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active) WHERE active = true;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_currencies_updated ON currencies(updated);

-- +goose Down
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_currencies_updated;