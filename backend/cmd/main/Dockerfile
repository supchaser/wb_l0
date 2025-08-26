FROM golang:1.25 AS builder

RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o main ./cmd/main/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["/app/main"]
