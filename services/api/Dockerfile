# syntax=docker/dockerfile:1.7

# Multi-module Go workspace build — context = repo root.
FROM golang:1.26-alpine AS build
WORKDIR /src

# Copy workspace + every module manifest so `go mod download` can cache.
COPY go.work go.work.sum* ./
COPY pkg pkg/
COPY services services/
COPY scripts scripts/

RUN go mod download

# Build the api binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./services/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
