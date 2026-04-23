FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY common ./common
COPY services/backoffice ./services/backoffice

WORKDIR /app/services/backoffice
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/backoffice ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /bin/backoffice /app/backoffice

EXPOSE 8082

ENTRYPOINT ["/app/backoffice"]
