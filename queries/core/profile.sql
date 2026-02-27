-- name: GetProfile :one
SELECT 
    u.name,
    u.email,
    u.role,
    u.inn,
    u.business_type,
    u.verification_status 
FROM users u
WHERE u.id = $1;