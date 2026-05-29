-- name: CreateRecoveryCode :exec
INSERT INTO user_recovery_codes (user_id, code_hash)
VALUES ($1, $2);

-- name: ListUserRecoveryCodes :many
SELECT id, user_id, code_hash, used_at, created_at
FROM user_recovery_codes
WHERE user_id = $1 AND used_at IS NULL
ORDER BY created_at ASC;

-- name: MarkRecoveryCodeUsed :exec
UPDATE user_recovery_codes SET used_at = NOW()
WHERE id = $1;

-- name: DeleteUserRecoveryCodes :exec
DELETE FROM user_recovery_codes WHERE user_id = $1;
