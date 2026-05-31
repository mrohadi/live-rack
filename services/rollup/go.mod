module github.com/live-rack/services/rollup

go 1.26.3

require (
	github.com/google/uuid v1.6.0
	github.com/live-rack/pkg/chstore v0.0.0
)

replace github.com/live-rack/pkg/chstore => ../../pkg/chstore
