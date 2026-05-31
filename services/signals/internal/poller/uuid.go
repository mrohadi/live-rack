package poller

import "github.com/google/uuid"

// parseUUID parses a canonical UUID string.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
