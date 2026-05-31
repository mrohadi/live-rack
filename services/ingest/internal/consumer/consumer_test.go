package consumer_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/events"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/ingest/internal/consumer"
)

type fakeRecorder struct {
	gotOrg uuid.UUID
	gotArg store.CreateSaleEventParams
	calls  int
}

func (f *fakeRecorder) RecordSale(_ context.Context, orgID uuid.UUID, arg store.CreateSaleEventParams) error {
	f.gotOrg = orgID
	f.gotArg = arg
	f.calls++
	return nil
}

func TestDecodeSale(t *testing.T) {
	org := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	data, _ := json.Marshal(events.POSSale{
		OrgID: org, Source: "shopify", OrderID: "#1001", SKU: "LR-1240",
		Qty: 2, AmountCents: 3998, Currency: "USD", Channel: "online", OccurredAt: now,
	})

	gotOrg, arg, err := consumer.DecodeSale(data)
	require.NoError(t, err)
	assert.Equal(t, org, gotOrg)
	assert.Equal(t, "LR-1240", arg.Sku)
	assert.Equal(t, int32(2), arg.Qty)
	assert.Equal(t, int64(3998), arg.AmountCents)
	assert.False(t, arg.StoreID.Valid)
	assert.Equal(t, now, arg.Ts)
}

func TestDecodeSale_MissingOrg(t *testing.T) {
	data, _ := json.Marshal(events.POSSale{Source: "shopify"})
	_, _, err := consumer.DecodeSale(data)
	assert.Error(t, err)
}

func TestConsumer_Handle(t *testing.T) {
	org := uuid.New()
	data, _ := json.Marshal(events.POSSale{OrgID: org, Source: "square", SKU: "SKU1", Qty: 1, AmountCents: 500})

	rec := &fakeRecorder{}
	require.NoError(t, consumer.New(rec).Handle(context.Background(), data))
	assert.Equal(t, 1, rec.calls)
	assert.Equal(t, org, rec.gotOrg)
	assert.Equal(t, "SKU1", rec.gotArg.Sku)
}
