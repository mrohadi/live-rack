.PHONY: dev dev-status seed seed-reset test lint typecheck build clean hooks-install notion-seed

# Start full local dev stack
dev:
	docker compose -f deploy/docker/docker-compose.yml up -d

dev-status:
	docker compose -f deploy/docker/docker-compose.yml ps

# Load fixture data from references/live-rack/project/data.jsx
seed:
	go run ./scripts/seed/...

seed-reset:
	psql $$DATABASE_URL -c "TRUNCATE zones, items, item_locations, tasks, pipelines, pipeline_stages, pipeline_cards, integrations, scan_events, sales_events CASCADE;"
	$(MAKE) seed

# Run all tests
test:
	go test -race ./...
	pnpm -F web test --run

# Lint
lint:
	golangci-lint run ./...
	pnpm -F web lint

# Type check
typecheck:
	pnpm -F web typecheck

# Build all
build:
	go build -o bin/api ./services/api/...
	go build -o bin/ingest ./services/ingest/...
	go build -o bin/rollup ./services/rollup/...
	pnpm -F web build

# Generate sqlc
generate:
	sqlc generate

# Run DB migrations
migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-status:
	goose -dir migrations postgres "$$DATABASE_URL" status

hooks-install:
	go install github.com/evilmartians/lefthook@latest
	lefthook install

notion-seed:
	NOTION_API_KEY=$$NOTION_API_KEY NOTION_PARENT_PAGE_ID=$$NOTION_PARENT_PAGE_ID \
	go run -C scripts/notion-seed .

clean:
	rm -rf bin/ apps/web/dist/ coverage.out
	docker compose -f deploy/docker/docker-compose.yml down -v
