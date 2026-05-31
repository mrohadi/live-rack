package integrations_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/live-rack/pkg/integrations"
)

func TestNetSuite_Skeleton(t *testing.T) {
	ns := integrations.NewNetSuite()
	assert.Equal(t, "netsuite", ns.Kind())

	err := ns.Verify("s", []byte(`{}`), nil)
	assert.ErrorIs(t, err, integrations.ErrNotImplemented)

	_, err = ns.ParseSales([]byte(`{}`))
	assert.ErrorIs(t, err, integrations.ErrNotImplemented)
}

func TestNetSuite_SatisfiesAdapter(t *testing.T) {
	var _ integrations.Adapter = integrations.NewNetSuite()
}
