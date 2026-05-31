package chsink_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/live-rack/pkg/chstore"
	"github.com/live-rack/pkg/events"
	"github.com/live-rack/services/ingest/internal/chsink"
)

func TestDecodeDemographics(t *testing.T) {
	org, store := uuid.New(), uuid.New()
	data, _ := json.Marshal(events.Demographics{
		OrgID: org, StoreID: store, Segment: "catchment", Metric: "median_income", Value: 64250, Day: "2026-06-01",
	})
	row, err := chsink.DecodeDemographics(data)
	require.NoError(t, err)
	assert.Equal(t, org.String(), row["org_id"])
	assert.Equal(t, "median_income", row["metric"])
	assert.Equal(t, 64250.0, row["value"])
	assert.Equal(t, "2026-06-01", row["day"])
}

func TestDecodeDemographics_DefaultsDay(t *testing.T) {
	data, _ := json.Marshal(events.Demographics{OrgID: uuid.New(), Metric: "foot_traffic", Value: 1})
	row, err := chsink.DecodeDemographics(data)
	require.NoError(t, err)
	assert.Len(t, row["day"], 10) // YYYY-MM-DD
}

func TestDecodeDemographics_Invalid(t *testing.T) {
	missingOrg, _ := json.Marshal(events.Demographics{Metric: "x"})
	_, err := chsink.DecodeDemographics(missingOrg)
	assert.Error(t, err)

	missingMetric, _ := json.Marshal(events.Demographics{OrgID: uuid.New()})
	_, err = chsink.DecodeDemographics(missingMetric)
	assert.Error(t, err)
}

// TestDemographics_Integration writes a snapshot into live ClickHouse and reads
// it back. Skipped unless CLICKHOUSE_URL is set.
func TestDemographics_Integration(t *testing.T) {
	rawURL := os.Getenv("CLICKHOUSE_URL")
	if rawURL == "" {
		t.Skip("CLICKHOUSE_URL not set; skipping ClickHouse integration test")
	}
	db := os.Getenv("CLICKHOUSE_DB")
	if db == "" {
		db = "liverack"
	}
	cfg, err := chstore.ParseConfig(rawURL, db)
	require.NoError(t, err)
	ch := chstore.New(cfg)
	ctx := context.Background()
	require.NoError(t, ch.Migrate(ctx))

	org := uuid.New()
	data, _ := json.Marshal(events.Demographics{
		OrgID: org, StoreID: uuid.New(), Segment: "catchment", Metric: "median_income", Value: 64250, Day: "2026-06-01",
	})
	require.NoError(t, chsink.New(ch).HandleDemographics(ctx, data))

	body, err := ch.Query(ctx, "SELECT metric, value FROM demographics FINAL WHERE org_id='"+org.String()+"' FORMAT TSV")
	require.NoError(t, err)
	assert.Equal(t, "median_income\t64250", strings.TrimSpace(string(body)))
}
