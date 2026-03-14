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


-- name: UpdateProfile :exec
UPDATE users 
SET email = $1, name = $2
WHERE id = $3;


-- name: DeleteProfile :exec
DELETE FROM users
WHERE id = $1;
