FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/userservice ./cmd/userservice

FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /out/userservice /app/userservice
COPY migrations /app/migrations

RUN chown -R appuser:appuser /app

USER appuser

ENV PORT=8080 \
    GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["./userservice"]
