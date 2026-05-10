module github.com/live-rack/pkg/auth

go 1.22

require (
	github.com/clerk/clerk-sdk-go/v2 v2.3.0
	github.com/google/uuid v1.6.0
	github.com/live-rack/pkg/domain v0.0.0
)

require (
	github.com/go-jose/go-jose/v3 v3.0.3 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
)

replace github.com/live-rack/pkg/domain => ../domain
