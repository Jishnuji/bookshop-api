-- +goose Up
INSERT INTO "categories" ("id", "name", "created_at", "updated_at") VALUES (DEFAULT, 'Pulp fiction', DEFAULT, DEFAULT);

-- +goose Down

