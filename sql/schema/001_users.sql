-- +goose Up
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    CONSTRAINT users_email_unique UNIQUE (email)
);

-- +goose Down
DROP TABLE users;