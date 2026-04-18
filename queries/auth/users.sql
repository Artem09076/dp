-- name: CreateUser :one
INSERT INTO users (
    name,
    email,
    password_hash,
    role,
    inn,
    business_type,
    verification_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByInn :one
SELECT * FROM users u WHERE u.inn = $1;


-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

