-- name: ListServices :many
SELECT * FROM services;

-- name: CreateService :one
INSERT INTO services (
    performer_id, title, description, price, duration_minutes
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
