-- +goose Up
CREATE TABLE chirps (
    id VARCHAR(50) PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    CONSTRAINT chirps_user_id_foreign FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;