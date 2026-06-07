# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.work go.work.sum* ./
COPY pkg pkg/
COPY services services/
COPY scripts scripts/

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/insight ./services/insight

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/insight /insight
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/insight"]
