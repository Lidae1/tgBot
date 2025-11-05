-- +goose Up
CREATE TABLE users (
                       chat_id BIGINT PRIMARY KEY,
                       username TEXT,
                       active BOOLEAN DEFAULT true
);

-- +goose Down
DROP TABLE users;