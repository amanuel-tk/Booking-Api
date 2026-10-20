# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build API
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" \
    -o /booking-api ./cmd/api

# Build migration CLI using your existing dependency
RUN CGO_ENABLED=0 GOOS=linux \
    go build -tags 'postgres' \
    -o /migrate \
    github.com/golang-migrate/migrate/v4/cmd/migrate


# Stage 2: Production
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

COPY --from=builder /booking-api /app/booking-api
COPY --from=builder /migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations /app/migrations

USER app

EXPOSE 3456

CMD ["/app/booking-api"]