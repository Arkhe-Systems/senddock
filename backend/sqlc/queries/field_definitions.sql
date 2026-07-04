-- name: CreateFieldDefinition :one
INSERT INTO subscriber_field_definitions (project_id, key, label, field_type, options, required)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFieldDefinitionsByProject :many
SELECT * FROM subscriber_field_definitions
WHERE project_id = $1
ORDER BY created_at ASC;

-- name: GetFieldDefinition :one
SELECT * FROM subscriber_field_definitions
WHERE id = $1 AND project_id = $2;

-- name: UpdateFieldDefinition :one
UPDATE subscriber_field_definitions SET
    label = $3,
    options = $4,
    required = $5
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteFieldDefinition :exec
DELETE FROM subscriber_field_definitions
WHERE id = $1 AND project_id = $2;
