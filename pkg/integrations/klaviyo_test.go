package integrations_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/integrations"
)

func TestKlaviyo_BuildTrackEvent(t *testing.T) {
	sale := integrations.Sale{
		Source: "shopify", OrderID: "#1001", SKU: "LR-1240", Qty: 2,
		AmountCents: 3998, Currency: "USD", Channel: "online",
		OccurredAt: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
	}
	req, err := integrations.NewKlaviyo().BuildTrackEvent("pk_test", "buyer@x.io", sale)
	require.NoError(t, err)

	assert.Equal(t, "https://a.klaviyo.com/api/events/", req.URL)
	assert.Equal(t, "Klaviyo-API-Key pk_test", req.Headers["Authorization"])

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(req.Body, &decoded))
	attrs := decoded["data"].(map[string]any)["attributes"].(map[string]any)
	assert.Equal(t, 39.98, attrs["value"])
	assert.Equal(t, "#1001", attrs["unique_id"])
	props := attrs["properties"].(map[string]any)
	assert.Equal(t, "LR-1240", props["sku"])
}

func TestKlaviyo_RequiresKeyAndEmail(t *testing.T) {
	k := integrations.NewKlaviyo()
	_, err := k.BuildTrackEvent("", "x@y.io", integrations.Sale{})
	assert.Error(t, err)
	_, err = k.BuildTrackEvent("pk", "", integrations.Sale{})
	assert.Error(t, err)
}
