FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common ./common
COPY services/auth-service ./services/auth-service

WORKDIR /app/services/auth-service
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/auth-service ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/auth-service /app/auth-service

EXPOSE 8081

ENTRYPOINT ["/app/auth-service"]
