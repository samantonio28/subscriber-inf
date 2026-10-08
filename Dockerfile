# Образ сервиса (тестируемой системы). Собирает бинарник приложения и запускает
# его в минимальном alpine-образе.
#
# Используется в CI/CD как «тестируемая система», развёрнутая в контейнере
# (требование ЛР №2 п.1: окружение, которое тестируется, тоже разворачивается в
# контейнере).

# --- Этап сборки ---
FROM golang:1.25-alpine AS build
WORKDIR /app

# Кэшируем зависимости.
COPY go.mod go.sum ./
RUN go mod download

# Генерируем OpenAPI-код (internal/api, pkg/clients — они в .gitignore) и собираем.
COPY . .
RUN go generate && go build -o subscriber-inf cmd/main.go

# --- Этап запуска ---
FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/subscriber-inf ./subscriber-inf
COPY configs/ ./configs/
RUN mkdir -p logs

EXPOSE 8080
CMD ["./subscriber-inf"]
