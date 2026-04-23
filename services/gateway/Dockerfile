FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common ./common
COPY services/gateway ./services/gateway

WORKDIR /app/services/gateway
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/gateway ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/gateway /app/gateway

EXPOSE 8080

ENTRYPOINT ["/app/gateway"]
