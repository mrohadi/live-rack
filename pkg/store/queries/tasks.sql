-- name: CreateTask :one
INSERT INTO tasks (org_id, store_id, zone_id, title, status, priority, assignee_id, due_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE org_id = $1 AND id = $2;

-- name: ListTasksByStore :many
SELECT * FROM tasks
WHERE org_id = $1 AND store_id = $2
ORDER BY status, priority DESC, created_at;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $3
WHERE org_id = $1 AND id = $2
RETURNING *;

-- name: AssignTask :one
UPDATE tasks
SET assignee_id = $3
WHERE org_id = $1 AND id = $2
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE org_id = $1 AND id = $2;

-- name: CountOpenTasksByTitle :one
SELECT count(*) FROM tasks
WHERE org_id = $1 AND store_id = $2 AND zone_id = $3
  AND title = $4 AND status <> 'done';