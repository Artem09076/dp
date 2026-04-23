-- name: CreateDiscount :one
INSERT INTO discounts (service_id, type, value, valid_from, valid_to, max_uses, used_count)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;


-- name: GetDiscountById :one
SELECT * FROM discounts WHERE id = $1;

-- name: UpdateDiscount :exec
UPDATE discounts
SET
    valid_from = COALESCE($2, valid_from),
    valid_to = COALESCE($3, valid_to),
    max_uses = COALESCE($4, max_uses)
WHERE id = $1;


-- name: DeleteDiscount :exec
DELETE FROM discounts WHERE id = $1;

-- name: GetDiscountsByServiceID :many
SELECT * FROM discounts 
WHERE service_id = $1 
  AND valid_from <= NOW() 
  AND valid_to >= NOW()
  AND used_count < max_uses
ORDER BY created_at DESC;