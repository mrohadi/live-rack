-- +goose Up
-- +goose StatementBegin

-- pg_trgm powers fuzzy ⌘K search over SKUs, item names, and zone names.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_items_name_trgm ON items USING gin (name gin_trgm_ops);
CREATE INDEX idx_items_sku_trgm  ON items USING gin (sku  gin_trgm_ops);
CREATE INDEX idx_zones_name_trgm ON zones USING gin (name gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_items_name_trgm;
DROP INDEX IF EXISTS idx_items_sku_trgm;
DROP INDEX IF EXISTS idx_zones_name_trgm;
DROP EXTENSION IF EXISTS pg_trgm;

-- +goose StatementEnd
