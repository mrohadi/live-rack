-- +goose Up
-- +goose StatementBegin

-- Unit price in integer cents — avoids float rounding. Stock value is
-- qty * price_cents, summed per SKU/store for dead-stock and on-hand $ views.
ALTER TABLE items
    ADD COLUMN price_cents INTEGER NOT NULL DEFAULT 0
    CHECK (price_cents >= 0);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE items DROP COLUMN IF EXISTS price_cents;

-- +goose StatementEnd
