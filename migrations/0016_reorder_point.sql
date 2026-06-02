-- +goose Up
-- +goose StatementBegin

-- Per-SKU reorder point: when on-hand qty in a zone falls to/below this, the
-- SKU is "low" and a restock task is auto-created. 0 disables the trigger.
ALTER TABLE items
    ADD COLUMN reorder_point INTEGER NOT NULL DEFAULT 0
    CHECK (reorder_point >= 0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE items DROP COLUMN IF EXISTS reorder_point;

-- +goose StatementEnd
