FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/todoservice ./cmd/todoservice

FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /out/todoservice /app/todoservice
COPY migrations /app/migrations

RUN chown -R appuser:appuser /app

USER appuser

ENV PORT=8081 \
    GIN_MODE=release

EXPOSE 8081

ENTRYPOINT ["./todoservice"]
