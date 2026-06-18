# Multi-stage Distroless

# Step 1: Build
FROM --platform=linux/amd64 golang:1.26.1-bookworm AS builder
WORKDIR /app

# cashe
COPY go.mod go.sum ./
RUN go mod download

# source code copy
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# go build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" -o /med-tracker cmd/app/main.go

# Step 2: Minimal image
FROM gcr.io/distroless/static-debian12:latest-amd64
WORKDIR /

# moving binaries
COPY --from=builder /med-tracker /med-tracker

EXPOSE 8080
ENTRYPOINT ["/med-tracker"]