package pipelines_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgauth "github.com/live-rack/pkg/auth"
	"github.com/live-rack/pkg/domain"
	"github.com/live-rack/pkg/store"
	"github.com/live-rack/services/api/internal/pipelines"
)

type fakeStore struct {
	pipelines  []store.Pipeline
	pipe       store.Pipeline
	stages     []store.PipelineStage
	cards      []store.PipelineCard
	moved      store.PipelineCard
	gotOrg     uuid.UUID
	gotMovePos int32
}

func (f *fakeStore) ListPipelinesByStore(_ context.Context, arg store.ListPipelinesByStoreParams) ([]store.Pipeline, error) {
	f.gotOrg = arg.OrgID
	return f.pipelines, nil
}
func (f *fakeStore) GetPipeline(_ context.Context, arg store.GetPipelineParams) (store.Pipeline, error) {
	f.gotOrg = arg.OrgID
	return f.pipe, nil
}
func (f *fakeStore) ListStagesByPipeline(_ context.Context, _ store.ListStagesByPipelineParams) ([]store.PipelineStage, error) {
	return f.stages, nil
}
func (f *fakeStore) ListCardsByPipeline(_ context.Context, _ store.ListCardsByPipelineParams) ([]store.PipelineCard, error) {
	return f.cards, nil
}
func (f *fakeStore) MoveCard(_ context.Context, arg store.MoveCardParams) (store.PipelineCard, error) {
	f.gotMovePos = arg.StagePosition
	f.moved.StagePosition = arg.StagePosition
	return f.moved, nil
}

func newContext(t *testing.T, e *echo.Echo, method, target, body string, p *domain.Principal) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetRequest(c.Request().WithContext(pkgauth.WithPrincipal(c.Request().Context(), p)))
	return c, rec
}

func TestPipelinesHandler_Board_FlagsAgeing(t *testing.T) {
	orgID, storeID, pipeID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC()

	fresh := store.PipelineCard{ID: uuid.New(), PipelineID: pipeID, StagePosition: 0, Title: "Fresh", Priority: "low", EnteredStageAt: now.Add(-30 * time.Minute)}
	stale := store.PipelineCard{ID: uuid.New(), PipelineID: pipeID, StagePosition: 0, Title: "Stale", Priority: "high", EnteredStageAt: now.Add(-3 * time.Hour)}

	fs := &fakeStore{
		pipe:   store.Pipeline{ID: pipeID, OrgID: orgID, StoreID: storeID, Key: "item-restoration", Name: "Item Restoration"},
		stages: []store.PipelineStage{{Position: 0, Name: "Intake", SlaSeconds: 3600}}, // 1h SLA
		cards:  []store.PipelineCard{fresh, stale},
	}
	e := echo.New()
	h := pipelines.New(fs)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodGet, "/api/v1/stores/"+storeID.String()+"/pipelines/"+pipeID.String(), "", p)
	c.SetParamNames("storeID", "id")
	c.SetParamValues(storeID.String(), pipeID.String())

	require.NoError(t, h.Board(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID, fs.gotOrg)

	var out pipelines.BoardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out.Stages, 1)
	require.Len(t, out.Cards, 2)

	byTitle := map[string]pipelines.CardRow{}
	for _, cd := range out.Cards {
		byTitle[cd.Title] = cd
	}
	assert.False(t, byTitle["Fresh"].Ageing, "30m < 1h SLA")
	assert.True(t, byTitle["Stale"].Ageing, "3h > 1h SLA")
}

func TestPipelinesHandler_MoveCard_OK(t *testing.T) {
	orgID, storeID, pipeID, cardID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{moved: store.PipelineCard{ID: cardID, PipelineID: pipeID, Title: "Move me", Priority: "med", EnteredStageAt: time.Now().UTC()}}
	e := echo.New()
	h := pipelines.New(fs)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleManager}
	c, rec := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/pipelines/"+pipeID.String()+"/cards/"+cardID.String(),
		`{"stage_position":2}`, p)
	c.SetParamNames("storeID", "id", "cardID")
	c.SetParamValues(storeID.String(), pipeID.String(), cardID.String())

	require.NoError(t, h.MoveCard(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int32(2), fs.gotMovePos)

	var out pipelines.CardRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.Equal(t, int32(2), out.StagePosition)
	assert.False(t, out.Ageing, "freshly moved card resets ageing")
}

func TestPipelinesHandler_MoveCard_ReadonlyForbidden(t *testing.T) {
	orgID, storeID, pipeID, cardID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	fs := &fakeStore{}
	e := echo.New()
	h := pipelines.New(fs)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleReadonly}
	c, _ := newContext(t, e, http.MethodPatch,
		"/api/v1/stores/"+storeID.String()+"/pipelines/"+pipeID.String()+"/cards/"+cardID.String(),
		`{"stage_position":1}`, p)
	c.SetParamNames("storeID", "id", "cardID")
	c.SetParamValues(storeID.String(), pipeID.String(), cardID.String())

	err := h.MoveCard(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusForbidden, he.Code)
}

func TestPipelinesHandler_List(t *testing.T) {
	orgID, storeID := uuid.New(), uuid.New()
	fs := &fakeStore{pipelines: []store.Pipeline{{ID: uuid.New(), OrgID: orgID, StoreID: storeID, Key: "item-restoration", Name: "Item Restoration"}}}
	e := echo.New()
	h := pipelines.New(fs)

	p := &domain.Principal{UserID: uuid.New(), OrgID: orgID, Role: domain.RoleStaff}
	c, rec := newContext(t, e, http.MethodGet, "/api/v1/stores/"+storeID.String()+"/pipelines", "", p)
	c.SetParamNames("storeID")
	c.SetParamValues(storeID.String())

	require.NoError(t, h.List(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	var out []pipelines.PipelineRow
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	require.Len(t, out, 1)
	assert.Equal(t, "item-restoration", out[0].Key)
}
