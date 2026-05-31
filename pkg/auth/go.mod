module github.com/live-rack/pkg/auth

go 1.25.0

require (
	github.com/coreos/go-oidc/v3 v3.18.0
	github.com/google/uuid v1.6.0
	github.com/live-rack/pkg/domain v0.0.0
)

require (
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
)

replace github.com/live-rack/pkg/domain => ../domain
