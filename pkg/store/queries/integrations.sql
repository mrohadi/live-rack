-- name: UpsertIntegration :one
INSERT INTO integrations (org_id, kind, status, external_id, secret, config)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (org_id, kind) DO UPDATE
SET status = EXCLUDED.status, external_id = EXCLUDED.external_id,
    secret = EXCLUDED.secret, config = EXCLUDED.config
RETURNING *;

-- name: GetIntegration :one
SELECT * FROM integrations
WHERE org_id = $1 AND kind = $2;

-- name: ListIntegrations :many
SELECT id, org_id, kind, status, external_id, config, created_at, updated_at
FROM integrations
WHERE org_id = $1
ORDER BY kind;

-- name: InsertInboundWebhook :one
INSERT INTO webhooks_inbound (org_id, provider, event_id, topic, status)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (org_id, provider, event_id) DO NOTHING
RETURNING *;

-- name: MarkWebhookStatus :exec
UPDATE webhooks_inbound
SET status = $3
WHERE org_id = $1 AND id = $2;

-- name: ListInboundWebhooks :many
SELECT * FROM webhooks_inbound
WHERE org_id = $1
ORDER BY received_at DESC
LIMIT $2;
