package integrations_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

type stubAdapter struct{ kind string }

func (s stubAdapter) Kind() string                                   { return s.kind }
func (s stubAdapter) EventID(http.Header) string                     { return "" }
func (s stubAdapter) Verify(string, []byte, http.Header) error       { return nil }
func (s stubAdapter) ParseSales([]byte) ([]integrations.Sale, error) { return nil, nil }

func TestRegistry_GetKnownAndUnknown(t *testing.T) {
	r := integrations.NewRegistry(stubAdapter{kind: "shopify"}, stubAdapter{kind: "square"})

	a, ok := r.Get("shopify")
	require.True(t, ok)
	assert.Equal(t, "shopify", a.Kind())

	_, ok = r.Get("stripe")
	assert.False(t, ok)
}
