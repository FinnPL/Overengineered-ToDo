FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /workspace/main ./serverless/dueNotifier

FROM public.ecr.aws/lambda/go:1

COPY --from=builder /workspace/main ${LAMBDA_RUNTIME_DIR}/main

CMD ["main"]
