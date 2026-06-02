package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/domain"
)

func TestShipmentStatus_Valid(t *testing.T) {
	for _, s := range []domain.ShipmentStatus{
		domain.ShipmentPacking, domain.ShipmentPacked,
		domain.ShipmentDispatched, domain.ShipmentCancelled,
	} {
		assert.True(t, s.Valid(), string(s))
	}
	assert.False(t, domain.ShipmentStatus("bogus").Valid())
}

func TestCanDispatch(t *testing.T) {
	assert.True(t, domain.CanDispatch(domain.ShipmentPacked))
	assert.False(t, domain.CanDispatch(domain.ShipmentPacking))
	assert.False(t, domain.CanDispatch(domain.ShipmentDispatched))
}

func TestCanCancelShipment(t *testing.T) {
	assert.True(t, domain.CanCancelShipment(domain.ShipmentPacking))
	assert.True(t, domain.CanCancelShipment(domain.ShipmentPacked))
	assert.False(t, domain.CanCancelShipment(domain.ShipmentDispatched))
	assert.False(t, domain.CanCancelShipment(domain.ShipmentCancelled))
}
