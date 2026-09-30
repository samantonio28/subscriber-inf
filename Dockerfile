FROM golang:1.25-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /logs
COPY . .

# Generate OpenAPI server/types/client from doc.yaml (output is gitignored)
RUN go generate ./...

CMD ["go", "run", "cmd/main.go"]
