# Swagger / OpenAPI — live-rack

## Stack

`swaggo/swag` + `swaggo/echo-swagger` — embedded Swagger UI, auto-served on startup.

## Access

Server running → open `http://localhost:8080/swagger/index.html`

No extra commands. Same as .NET `UseSwaggerUI()`.

## Regenerate docs (after adding/changing annotations)

```bash
cd services/api
swag init -g main.go --output docs
```

Commit `services/api/docs/` — checked into the repo.

## Add annotations to a new handler

```go
// List godoc
//
//	@Summary		List widgets
//	@Tags			widgets
//	@Produce		json
//	@Param			storeID	path		string			true	"Store UUID"
//	@Success		200		{array}		WidgetResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/stores/{storeID}/widgets [get]
func (h *Handler) List(c echo.Context) error { ... }
```

## Rules

- Use local response types (`WidgetResponse`, `ErrorResponse`) — swag cannot parse `store.*` or `echo.HTTPError`
- `ErrorResponse` lives in `handler.go` of each package
- `ZoneType` and `json.RawMessage` fields → replace with `string` + `enums:` tag and `any` + `swaggertype:"object"` respectively
- Swagger UI route `/swagger/*` registered in `main.go` — no auth middleware on it

## Wire new handler group in main.go

```go
// 1. Register route group
widgets.New(q).Register(api.Group("/stores"))

// 2. Re-run swag init
// 3. Swagger UI updates on next go run
```

## File locations

| File | Purpose |
|------|---------|
| `services/api/main.go` | `@title`, `@host`, `@BasePath` annotations + `e.GET("/swagger/*", echoSwagger.WrapHandler)` |
| `services/api/docs/` | Generated — `docs.go`, `swagger.json`, `swagger.yaml` |
| `docs/openapi/zones.yaml` | Hand-written OpenAPI 3.1 spec (reference, not served by swaggo) |
