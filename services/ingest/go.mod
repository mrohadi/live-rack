module github.com/live-rack/services/ingest

go 1.26.3

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/live-rack/pkg/events v0.0.0
	github.com/live-rack/pkg/store v0.0.0
	github.com/nats-io/nats.go v1.52.0
	github.com/stretchr/testify v1.11.1
)

replace (
	github.com/live-rack/pkg/domain => ../../pkg/domain
	github.com/live-rack/pkg/events => ../../pkg/events
	github.com/live-rack/pkg/store => ../../pkg/store
)
