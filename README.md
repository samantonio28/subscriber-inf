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
