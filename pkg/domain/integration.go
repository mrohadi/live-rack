package domain

import (
	"time"

	"github.com/google/uuid"
)

// IntegrationKind identifies a third-party connector.
type IntegrationKind string

const (
	IntegrationShopify IntegrationKind = "shopify"
	IntegrationSquare  IntegrationKind = "square"
)

func (k IntegrationKind) Valid() bool {
	switch k {
	case IntegrationShopify, IntegrationSquare:
		return true
	default:
		return false
	}
}

// IntegrationStatus is the connection lifecycle state.
type IntegrationStatus string

const (
	IntegrationStatusConnected    IntegrationStatus = "connected"
	IntegrationStatusDisconnected IntegrationStatus = "disconnected"
	IntegrationStatusError        IntegrationStatus = "error"
)

// Integration is a configured third-party connection scoped to an org. ExternalID
// is the vendor account handle (Shopify shop domain, Square merchant id) used to
// route inbound webhooks back to the right org.
type Integration struct {
	ID         uuid.UUID         `json:"id"`
	OrgID      uuid.UUID         `json:"org_id"`
	Kind       IntegrationKind   `json:"kind"`
	Status     IntegrationStatus `json:"status"`
	ExternalID string            `json:"external_id"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// CanManageIntegrations reports whether the principal may connect/configure
// integrations. Admin-only, matching the design permission matrix.
func CanManageIntegrations(p *Principal) bool {
	return Can(p.Role, PermManageIntegrations)
}
