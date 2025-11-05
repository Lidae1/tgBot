-- +goose Up
CREATE TABLE IF NOT EXISTS users (
                                     chat_id BIGINT PRIMARY KEY,
                                     username TEXT,
                                     active BOOLEAN DEFAULT true,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS currencies (
                                          symbol VARCHAR(10) PRIMARY KEY,
    price TEXT NOT NULL,
    updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                          );

-- +goose Down
DROP TABLE IF EXISTS currencies;
DROP TABLE IF EXISTS users;