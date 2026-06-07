FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ingest ./services/ingest

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/ingest /ingest
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/ingest"]