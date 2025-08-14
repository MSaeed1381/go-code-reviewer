FROM golang:1.24.2-bullseye AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git ca-certificates gcc g++ librdkafka-dev pkg-config && rm -rf /var/lib/apt/lists/*

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOARCH=amd64 go build -o api-gateway ./services/api-gateway


FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y librdkafka1 ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/services/api-gateway/config.yaml .
COPY --from=builder /app .

CMD ["./api-gateway"]
