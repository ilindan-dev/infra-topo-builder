-- name: CreateLog :one
INSERT INTO logs (status)
VALUES ($1)
RETURNING id, status, nodes_count, ports_count, created_at;

-- name: UpdateLogStatus :exec
UPDATE logs
SET status = $1
WHERE id = $2;

-- name: UpdateLogCounts :exec
UPDATE logs
SET nodes_count = $1,
    ports_count = $2
WHERE id = $3;

-- name: GetLogByID :one
SELECT id, status, nodes_count, ports_count, created_at
FROM logs
WHERE id = $1;

-- name: DeleteLog :exec
DELETE FROM logs WHERE id = $1;

-- name: DeleteLogData :exec
DELETE FROM nodes WHERE log_id = $1;

-- name: GetLogStatus :one
SELECT status FROM logs WHERE id = $1;