# Start full local dev stack
.PHONY: dev
dev:
	docker compose -f deploy/docker/docker-compose.yml up -d

.PHONY: dev-status
dev-status:
	docker compose -f deploy/docker/docker-compose.yml ps
	
.PHONY: sync
sync:
	go work sync

.PHONY: build-go
build-go:
	for d in pkg/auth pkg/domain pkg/events pkg/observability pkg/store services/api; do \
		(cd $$d && go build ./...) || exit 1; \
	done

# Load fixture data from references/live-rack/project/data.jsx
.PHONY: seed
seed:
	go run ./scripts/seed/...

.PHONY: seed-reset
seed-reset:
	psql $$DATABASE_URL -c "TRUNCATE zones, items, item_locations, tasks, pipelines, pipeline_stages, pipeline_cards, integrations, scan_events, sales_events CASCADE;"
	$(MAKE) seed

# Run all tests
.PHONY: test
test:
	go test -race github.com/live-rack/services/api/... github.com/live-rack/pkg/auth/... github.com/live-rack/pkg/domain/... github.com/live-rack/pkg/observability/... github.com/live-rack/pkg/store/...
	pnpm -F web exec vitest run

# Lint
.PHONY: lint
lint:
	golangci-lint run --config .golangci.yml ./services/api/...
	golangci-lint run --config .golangci.yml ./pkg/auth/...
	golangci-lint run --config .golangci.yml ./pkg/events/...
	golangci-lint run --config .golangci.yml ./pkg/domain/...
	golangci-lint run --config .golangci.yml ./pkg/observability/...
	golangci-lint run --config .golangci.yml ./pkg/store/...
	pnpm -F web lint

# Prettier format check
.PHONY: prettier-check
prettier-check:
	pnpm -F web exec prettier --check "src/**/*.{ts,tsx,css}"

# Prettier auto-fix
.PHONY: prettier-fix
prettier-fix:
	pnpm -F web exec prettier --write "src/**/*.{ts,tsx,css}"

# Type check
.PHONY: typecheck
typecheck:
	pnpm -F web typecheck

# Build all
.PHONY: build
build:
	go build -o bin/api github.com/live-rack/services/api/...
	pnpm -F web build

# Generate sqlc
.PHONY: generate
generate:
	sqlc generate

# Run DB migrations
.PHONY: migrate-up
migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres "$$DATABASE_URL" status

.PHONY: migrate-down
migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

.PHONY: hooks-install
hooks-install:
	go install github.com/evilmartians/lefthook@latest
	lefthook install

.PHONY: notion-seed
notion-seed:
	GOWORK=off NOTION_API_KEY=$$NOTION_API_KEY NOTION_PARENT_PAGE_ID=$$NOTION_PARENT_PAGE_ID \
	go run -C scripts/notion-seed .

.PHONY: clean
clean:
	rm -rf bin/ apps/web/dist/ coverage.out
	docker compose -f deploy/docker/docker-compose.yml down -v
