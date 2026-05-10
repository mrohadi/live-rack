module github.com/live-rack/services/api

go 1.22

require (
	github.com/clerk/clerk-sdk-go/v2 v2.3.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/labstack/echo/v4 v4.12.0
	github.com/live-rack/pkg/auth v0.0.0
	github.com/live-rack/pkg/domain v0.0.0
	github.com/live-rack/pkg/observability v0.0.0
	github.com/live-rack/pkg/store v0.0.0
	github.com/prometheus/client_golang v1.19.1
	github.com/svix/svix-webhooks/go v1.24.0
	go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho v0.52.0
	go.opentelemetry.io/otel v1.27.0
)

replace (
	github.com/live-rack/pkg/auth        => ../../pkg/auth
	github.com/live-rack/pkg/domain      => ../../pkg/domain
	github.com/live-rack/pkg/observability => ../../pkg/observability
	github.com/live-rack/pkg/store       => ../../pkg/store
)
