FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common ./common
COPY core ./core
COPY services/auth-service ./services/auth-service

WORKDIR /app/services/auth-service
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/auth-service ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/auth-service /app/auth-service

EXPOSE 8081

ENTRYPOINT ["/app/auth-service"]
