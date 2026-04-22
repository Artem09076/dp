-- name: GetUnverifiedPerformers :many
SELECT 
    u.id, u.name, u.email, u.inn, u.business_type, u.verification_status, 
    u.role, u.created_at, u.updated_at,
    COUNT(DISTINCT s.id) as services_count,
    COUNT(DISTINCT b.id) as total_bookings
FROM users u
LEFT JOIN services s ON s.performer_id = u.id
LEFT JOIN bookings b ON b.service_id = s.id
WHERE u.role = 'performer' AND u.verification_status = 'pending'
GROUP BY u.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUnverifiedPerformers :one
SELECT COUNT(*) FROM users 
WHERE role = 'performer' AND verification_status = 'pending';

-- name: GetUsersWithFilters :many
SELECT id, name, email, inn, business_type, role, verification_status, created_at, updated_at
FROM users
WHERE (
    (sqlc.narg('role')::user_role IS NULL OR role = sqlc.narg('role')) AND
    (sqlc.narg('verification_status')::verification_status IS NULL OR verification_status = sqlc.narg('verification_status')) AND
    (sqlc.narg('search')::text IS NULL OR name ILIKE '%' || sqlc.narg('search') || '%' OR email ILIKE '%' || sqlc.narg('search') || '%')
)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsersWithFilters :one
SELECT COUNT(*)
FROM users
WHERE (
    (sqlc.narg('role')::user_role IS NULL OR role = sqlc.narg('role')) AND
    (sqlc.narg('verification_status')::verification_status IS NULL OR verification_status = sqlc.narg('verification_status')) AND
    (sqlc.narg('search')::text IS NULL OR name ILIKE '%' || sqlc.narg('search') || '%' OR email ILIKE '%' || sqlc.narg('search') || '%')
);

-- name: GetUserByID :one
SELECT id, name, email, inn, business_type, role, verification_status, created_at, updated_at
FROM users WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UpdateVerificationStatus :exec
UPDATE users SET verification_status = $2, updated_at = NOW() WHERE id = $1;

-- name: GetAllServices :many
SELECT s.id, s.performer_id, s.title, s.description, s.price, s.duration_minutes, s.created_at, s.updated_at, u.name as performer_name
FROM services s
JOIN users u ON u.id = s.performer_id
WHERE ($1::uuid IS NULL OR s.performer_id = $1)
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAllServices :one
SELECT COUNT(*) FROM services
WHERE ($1::uuid IS NULL OR performer_id = $1);

-- name: GetAllBookings :many
SELECT b.id, b.client_id, b.service_id, b.base_price, b.final_price, b.booking_time, b.status, b.created_at, b.updated_at,
       c.name as client_name, s.title as service_name
FROM bookings b
JOIN users c ON c.id = b.client_id
JOIN services s ON s.id = b.service_id
WHERE (
    (sqlc.narg('status')::booking_status IS NULL OR b.status = sqlc.narg('status')) AND
    (sqlc.narg('client_id')::uuid IS NULL OR b.client_id = sqlc.narg('client_id')) AND
    (sqlc.narg('performer_id')::uuid IS NULL OR s.performer_id = sqlc.narg('performer_id'))
)
ORDER BY b.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAllBookings :one
SELECT COUNT(*)
FROM bookings b
JOIN services s ON s.id = b.service_id
WHERE (
    (sqlc.narg('status')::booking_status IS NULL OR b.status = sqlc.narg('status')) AND
    (sqlc.narg('client_id')::uuid IS NULL OR b.client_id = sqlc.narg('client_id')) AND
    (sqlc.narg('performer_id')::uuid IS NULL OR s.performer_id = sqlc.narg('performer_id'))
);

-- name: GetAllReviews :many
SELECT r.id, r.booking_id, r.rating, r.comment, r.created_at, r.updated_at,
       u.name as client_name, s.id as service_id
FROM reviews r
JOIN bookings b ON b.id = r.booking_id
JOIN users u ON u.id = b.client_id
JOIN services s ON s.id = b.service_id
WHERE (
    ($1::uuid IS NULL OR s.id = $1) AND
    (sqlc.narg('rating')::int IS NULL OR r.rating = sqlc.narg('rating'))
)
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAllReviews :one
SELECT COUNT(*)
FROM reviews r
JOIN bookings b ON b.id = r.booking_id
JOIN services s ON s.id = b.service_id
WHERE (
    ($1::uuid IS NULL OR s.id = $1) AND
    (sqlc.narg('rating')::int IS NULL OR r.rating = sqlc.narg('rating'))
);