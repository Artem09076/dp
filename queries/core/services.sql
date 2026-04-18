-- name: ListServices :many
SELECT * FROM services;

-- name: CreateService :one
INSERT INTO services (
    performer_id, title, description, price, duration_minutes
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;


-- name: SearchServices :many
SELECT * FROM services
WHERE title ILIKE '%' || $1 || '%'
   OR description ILIKE '%' || $1 || '%'
LIMIT $2 OFFSET $3;


-- name: GetService :one
SELECT 
    s.*,
    COALESCE((
        SELECT AVG(r.rating)::FLOAT
        FROM bookings b
        JOIN reviews r ON r.booking_id = b.id
        WHERE b.service_id = s.id
    ), 0)::FLOAT AS average_rating
FROM services s
WHERE s.id = $1
ORDER BY s.created_at DESC;


-- name: GetServices :many
SELECT * FROM services
WHERE performer_id = $1
ORDER BY created_at DESC;

-- name: DeleteService :exec
DELETE FROM services WHERE id = $1;

-- name: UpdateService :exec
UPDATE services
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    price = COALESCE($4, price),
    duration_minutes = COALESCE($5, duration_minutes)
WHERE id = $1;