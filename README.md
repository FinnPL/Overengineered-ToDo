# Overengineered ToDo Backend

This repository contains a lightweight microservice architecture for a ToDo application built with Go, Gin, and CockroachDB. The stack is split into:

- **User Service** (`api/cmd/userservice`): manages user accounts.
- **Todo Service** (`api/cmd/todoservice`): manages user tasks.
- **Due Notifier Lambda** (`api/serverless/dueNotifier`): example serverless function that finds upcoming todos.

## Local Development

### Requirements
- Docker & Docker Compose
- Go 1.25+ (only if running binaries locally)

### Bootstrapping CockroachDB and Services

```bash
docker compose up --build
```

This command starts:
- A single-node CockroachDB cluster (secure mode) exposed on `localhost:26257` with the DB Console on `localhost:8080`.
- A migrator job that creates the `todoapp` database and applies SQL migrations from `api/migrations`.
- Both HTTP services:
  - `User Service` on http://localhost:8082
  - `Todo Service` on http://localhost:8083
- TLS certificates are generated automatically and mounted at `/app/certs` inside each service container.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL-compatible connection string for CockroachDB | required |
| `PORT` | HTTP listen port (inside container) | users: `8080`, todos: `8081` |
| `GIN_MODE` | Gin runtime mode | `release` inside Docker |
| `SHUTDOWN_TIMEOUT_SECONDS` | Graceful shutdown timeout | `10` |

When running locally without Docker Compose, export a connection string such as:

```bash
export DATABASE_URL='postgresql://root@localhost:26257/todoapp?sslmode=disable'
```


Then launch a service:

```bash
cd api
go run ./cmd/userservice
```

## API Overview

- `POST /v1/users` – register a new user.
- `GET /v1/users/{id}` – fetch a user by ID.
- `GET /v1/users?limit=50` – list users (default limit 100).
- `DELETE /v1/users/{id}` – delete a user.
- `POST /v1/todos` – create a todo (requires `user_id`).
- `GET /v1/todos/{id}` – fetch a todo.
- `GET /v1/todos?user_id={uuid}` – list todos for a user.
- `PUT /v1/todos/{id}` – update fields (`title`, `description`, `due_date`, `completed`, `clear_due_date`).
- `PATCH /v1/todos/{id}/complete` – mark a todo as complete.
- Health probes for both services: `GET /healthz`.

## Serverless Function

The Lambda example aggregates todos due within a configurable time window.

```bash
cd api
GOOS=linux GOARCH=amd64 go build -o bin/due-notifier ./serverless/dueNotifier
```

To package as a Lambda container image:

```bash
cd api
docker build -t todo-due-notifier -f serverless/dueNotifier/Dockerfile .
```

Invoke locally using the AWS Lambda Runtime Interface Emulator:

```bash
docker run --rm -p 9000:8080 \
  -e DATABASE_URL='postgresql://root@host.docker.internal:26257/todoapp?sslmode=disable' \
  todo-due-notifier

curl -XPOST localhost:9000/2015-03-31/functions/function/invocations \
  -d '{"window_minutes": 120}'
```

## SQL Migrations

SQL definitions live under `api/migrations`. They are applied automatically by the `migrator` service in `docker-compose.yml`. Run them manually with:

```bash
docker compose run --rm migrator
```

## Testing

Run unit build checks:

```bash
cd api
go test ./...
```

---
