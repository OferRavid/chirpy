-- +goose Up
CREATE TABLE chirps(
    id UUID primary key,
    created_at TIMESTAMP not null,
    updated_at TIMESTAMP not null,
    body TEXT not null,
    user_id UUID not null REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;