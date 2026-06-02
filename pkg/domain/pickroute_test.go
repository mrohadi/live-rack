package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/domain"
)

func stop(sku string, x, y float64) domain.PickStop {
	return domain.PickStop{ZoneID: uuid.New(), SKU: sku, QtyRequested: 1, X: x, Y: y}
}

func TestOptimizePickRoute_NearestNeighbourFromOrigin(t *testing.T) {
	stops := []domain.PickStop{
		stop("FAR", 100, 100),
		stop("NEAR", 10, 0),
		stop("MID", 50, 0),
	}
	got := domain.OptimizePickRoute(stops, 0, 0)
	require.Len(t, got, 3)
	assert.Equal(t, "NEAR", got[0].SKU)
	assert.Equal(t, "MID", got[1].SKU)
	assert.Equal(t, "FAR", got[2].SKU)
}

func TestOptimizePickRoute_DoesNotMutateInput(t *testing.T) {
	stops := []domain.PickStop{stop("B", 100, 0), stop("A", 1, 0)}
	_ = domain.OptimizePickRoute(stops, 0, 0)
	assert.Equal(t, "B", stops[0].SKU)
	assert.Equal(t, "A", stops[1].SKU)
}

func TestOptimizePickRoute_EmptyAndSingle(t *testing.T) {
	assert.Empty(t, domain.OptimizePickRoute(nil, 0, 0))
	one := []domain.PickStop{stop("X", 5, 5)}
	got := domain.OptimizePickRoute(one, 0, 0)
	require.Len(t, got, 1)
	assert.Equal(t, "X", got[0].SKU)
}

func TestRouteLength_SumsLegs(t *testing.T) {
	// origin -> (3,4)=5 -> (3,4)+(0,0)... use simple right-angle legs.
	stops := []domain.PickStop{stop("A", 3, 4), stop("B", 3, 0)}
	// leg1 = 5, leg2 = sqrt(0 + 16) = 4
	assert.InDelta(t, 9.0, domain.RouteLength(stops, 0, 0), 1e-9)
}
