# subscriber-inf

Проект запускается через 
```
sudo docker-compose up --build
```

Конфигурации в `postgres.yaml`


## Генерация кода

Для генерации кода из спецификации OpenAPI выполните:
```
go generate
```

Это сгенерирует:
- Клиент и модели в `generated/client.go` (на основе конфигурации в `configs/backend.yaml`)
- Типы и сервер (gorilla/mux) в `internal/api/`
- Клиент и типы для внешнего использования в `pkg/clients/api/`


## Проверка API (WebLab #2)

Спецификация `openapi.yaml` (OpenAPI 3.1, пути `/api/v1`). Линтер и mock-сервер:

```
make lint-api    # линтер Redocly (0 ошибок)
make mock-api    # mock-сервер (Prism) на :4010 — 1-й терминал, оставить висеть
make demo        # демонстрация сценария — 2-й терминал (или пошагово: make demo-1 .. demo-5)
```
