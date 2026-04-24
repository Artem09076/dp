-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1;

-- name: CancelBooking :exec
UPDATE bookings
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1;

-- name: CompletedBooking :exec
UPDATE bookings
SET status = 'completed', updated_at = NOW()
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
SELECT 
    b.*,
    s.title AS service_title,
    c.name AS client_name,
    c.email AS client_email,
    p.name AS performer_name,
    p.email AS performer_email,
    s.performer_id,
    d.type AS discount_type,
    d.value AS discount_value
FROM bookings b
JOIN services s ON b.service_id = s.id
JOIN users c ON b.client_id = c.id
JOIN users p ON s.performer_id = p.id
LEFT JOIN discounts d ON b.discount_id = d.id
WHERE b.id = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: SubmitBooking :exec
UPDATE bookings
SET status = 'confirmed', updated_at = NOW()
WHERE id = $1;


-- name: GetBookingsForUpdate :many
SELECT 
    b.id,
    b.booking_time,
    s.duration_minutes
FROM bookings b
JOIN services s ON s.id = b.service_id
WHERE s.performer_id = $1
  AND b.status IN ('pending', 'confirmed')
  AND b.booking_time < $3
  AND b.booking_time + (s.duration_minutes || ' minutes')::interval > $2  -- new_start
FOR UPDATE;


-- name: UpdateBookingTime1 :exec
UPDATE bookings
SET booking_time = $1, updated_at = NOW()
WHERE id = $2;

-- name: UpdateBookingTime2 :exec
UPDATE bookings
SET booking_time = $1,
    status = 'pending',
    updated_at = NOW()
WHERE id = $2;


-- name: GetBookingByClientID :many
SELECT * FROM bookings WHERE client_id = $1;

-- name: GetBookingByPerformerID :many
SELECT b.* FROM bookings b
JOIN services s ON s.id = b.service_id
WHERE s.performer_id = $1;

-- name: ServiceExists :one
SELECT EXISTS (SELECT * FROM services WHERE id = $1);

-- name: IncreaseDiscountUsage :exec
UPDATE discounts
SET used_count = used_count + 1
WHERE id = $1;


-- name: DeleteBooking :exec
DELETE FROM bookings WHERE id = $1;


-- name: GetBookingsWithServiceInfo :many
SELECT 
    b.*,
    s.title AS service_title,
    s.performer_id
FROM bookings b
JOIN services s ON b.service_id = s.id
WHERE b.client_id = $1
ORDER BY b.booking_time DESC;

-- name: GetBookingsByPerformerWithServiceInfo :many
SELECT 
    b.*,
    s.title AS service_title
FROM bookings b
JOIN services s ON b.service_id = s.id
WHERE s.performer_id = $1
ORDER BY b.booking_time DESC;