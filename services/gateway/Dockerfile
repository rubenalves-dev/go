FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common ./common
COPY core ./core
COPY services/gateway ./services/gateway

WORKDIR /app/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/gateway ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/gateway /app/gateway

EXPOSE 8080

ENTRYPOINT ["/app/gateway"]
