package sales_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/sales"
)

type fakeStore struct {
	summary store.SalesSummaryRow
	byDay   []store.SalesByDayRow
	gotOrg  uuid.UUID
}

func (f *fakeStore) SalesSummary(_ context.Context, arg store.SalesSummaryParams) (store.SalesSummaryRow, error) {
	f.gotOrg = arg.OrgID
	return f.summary, nil
}
func (f *fakeStore) SalesByDay(_ context.Context, _ store.SalesByDayParams) ([]store.SalesByDayRow, error) {
	return f.byDay, nil
}

func TestSalesHandler_Summary(t *testing.T) {
	orgID := uuid.New()
	now := time.Now().UTC()
	since := now.AddDate(0, 0, -7).Truncate(24 * time.Hour)

	fs := &fakeStore{
		summary: store.SalesSummaryRow{RevenueCents: 12345, Units: 9, Orders: 4},
		byDay: []store.SalesByDayRow{
			{Day: since.Add(24 * time.Hour), RevenueCents: 1000}, // idx 1
			{Day: since.Add(72 * time.Hour), RevenueCents: 3000}, // idx 3
		},
	}
	e := echo.New()
	h := sales.New(fs)
	h.Register(e.Group("/api/v1"))

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleReadonly}
	req := httptest.NewRequestWithContext(
		pkgauth.WithPrincipal(context.Background(), p), http.MethodGet, "/api/v1/sales/summary", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID, fs.gotOrg)

	var out sales.SummaryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, int64(12345), out.RevenueCents)
	assert.Equal(t, int64(9), out.Units)
	assert.Equal(t, int64(4), out.Orders)
	require.Len(t, out.Spark, 7)
	assert.Equal(t, int64(1000), out.Spark[1])
	assert.Equal(t, int64(3000), out.Spark[3])
	assert.Equal(t, int64(0), out.Spark[0])
}
