-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1;

-- name: CancelBooking :exec
UPDATE bookings
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1;

-- name: CreateBooking :one
INSERT INTO bookings (
    client_id,
    service_id,
    base_price,
    discount_id,
    final_price,
    booking_time,
    status
)
VALUES ($1, $2, $3, $4, $5, $6, 'pending')
RETURNING id;

-- name: GetServiceById :one
SELECT * FROM services WHERE id = $1;

-- name: GetDiscountById :one
SELECT * FROM discounts WHERE id = $1;

-- name: GetBookingByID :one
SELECT b.*, s.title AS service_title, s.performer_id, d.type AS discount_type, d.value AS discount_value
FROM bookings b
JOIN services s ON b.service_id = s.id
LEFT JOIN discounts d ON b.discount_id = d.id
WHERE b.id = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: SubmitBooking :exec
UPDATE bookings
SET status = 'confirmed', updated_at = NOW()
WHERE id = $1;


-- name: GetBookingsForUpdate :many
SELECT b.id, b.booking_time, s.duration_minutes from bookings b
JOIN services s ON s.id = b.service_id
WHERE b.service_id = $1 AND b.status IN ('pending', 'confirmed')
FOR UPDATE;