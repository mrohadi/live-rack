module github.com/live-rack/pkg/auth

go 1.22

require (
	github.com/clerk/clerk-sdk-go/v2 v2.3.0
	github.com/google/uuid v1.6.0
	github.com/live-rack/pkg/domain v0.0.0
)

replace github.com/live-rack/pkg/domain => ../domain
