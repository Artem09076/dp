-- name: CreateReview :one
INSERT INTO reviews (booking_id, rating, comment)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetBookingByID :one
SELECT b.*, s.title AS service_title, s.performer_id, d.type AS discount_type, d.value AS discount_value
FROM bookings b
JOIN services s ON b.service_id = s.id
LEFT JOIN discounts d ON b.discount_id = d.id
WHERE b.id = $1;

-- name: GetReviewByID :one
SELECT * FROM reviews WHERE id = $1;

-- name: GetReviewByBookingID :one
SELECT * FROM reviews WHERE booking_id = $1;


-- name: GetReviewsByServiceID :many
SELECT r.*
FROM reviews r
JOIN bookings b ON r.booking_id = b.id
WHERE b.service_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;


-- name: UpdateReview :exec
UPDATE reviews
SET rating = $2, comment = $3
WHERE id = $1;


-- name: DeleteReview :exec
DELETE FROM reviews WHERE id = $1;