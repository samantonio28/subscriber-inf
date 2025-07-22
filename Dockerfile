FROM golang:1.23-alpine
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /logs
COPY . .

CMD ["go", "run", "cmd/main.go"]
