-- +goose Up
CREATE TABLE currencies (
                            symbol VARCHAR(10) PRIMARY KEY,
                            price TEXT NOT NULL
);

-- +goose Down
DROP TABLE currencies;