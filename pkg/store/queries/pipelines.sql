-- name: CreatePipeline :one
INSERT INTO pipelines (org_id, store_id, key, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPipeline :one
SELECT * FROM pipelines
WHERE org_id = $1 AND id = $2;

-- name: ListPipelinesByStore :many
SELECT * FROM pipelines
WHERE org_id = $1 AND store_id = $2
ORDER BY created_at;

-- name: CreateStage :one
INSERT INTO pipeline_stages (org_id, pipeline_id, position, name, sla_seconds)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListStagesByPipeline :many
SELECT * FROM pipeline_stages
WHERE org_id = $1 AND pipeline_id = $2
ORDER BY position;

-- name: CreateCard :one
INSERT INTO pipeline_cards (org_id, pipeline_id, stage_position, title, sku, priority, owner_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetCard :one
SELECT * FROM pipeline_cards
WHERE org_id = $1 AND id = $2;

-- name: ListCardsByPipeline :many
SELECT * FROM pipeline_cards
WHERE org_id = $1 AND pipeline_id = $2
ORDER BY stage_position, priority DESC, entered_stage_at;

-- name: MoveCard :one
UPDATE pipeline_cards
SET stage_position = $3, entered_stage_at = NOW()
WHERE org_id = $1 AND id = $2
RETURNING *;
