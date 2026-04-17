-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, provider, provider_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByProvider :one
SELECT * FROM users WHERE provider = $1 AND provider_id = $2;

-- name: UpdateUserPlan :exec
UPDATE users SET plan = $2, plan_changed_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: IncrementEmailsSent :exec
UPDATE users SET monthly_emails_sent = monthly_emails_sent +1, updated_at = NOW()
WHERE id = $1;

-- name: ResetMonthlyUsage :exec
UPDATE users SET monthly_emails_sent = 0,
    monthly_reset_at = date_trunc('month', NOW()) + INTERVAL '1 month',
    updated_at = NOW()
WHERE id = $1;