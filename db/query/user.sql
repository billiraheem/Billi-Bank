-- name: CreateUser :one
INSERT INTO users (
  username, hashed_password, fullname, email
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password),
  fullname = COALESCE(sqlc.narg(fullname), fullname),
  email = COALESCE(sqlc.narg(email), email)
WHERE username = sqlc.arg(username) 
RETURNING *;