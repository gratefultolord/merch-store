FROM golang:1.22.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/merch-store

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/merch-store .

COPY .env .

COPY migrations/001_init.up.sql /docker-entrypoint-initdb.d/001_init.up.sql

COPY migrations/002_seed_items.up.sql /docker-entrypoint-initdb.d/002_seed_items.up.sql

CMD ["./merch-store"]