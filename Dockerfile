FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /logs
COPY . .

# Generate OpenAPI server/types/client from doc.yaml + openapi.yaml (output is gitignored)
RUN go generate

# Собираем бинарник (а не `go run`) — быстрее стартует и устойчивее под нагрузкой.
RUN go build -ldflags "-s -w" -o /app/server ./cmd

CMD ["/app/server"]
