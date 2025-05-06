FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /loadbalancer ./cmd/http-load-balancer/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /loadbalancer ./
COPY config.yaml ./
COPY .env ./

RUN ls -la

EXPOSE 8090
CMD ["sh", "-c", "ls -la /app && echo 'Config content:' && cat /app/config.yaml && sleep 5 && ./loadbalancer --config=./config.yaml"]