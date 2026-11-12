# subscriber-inf

Сервис управления онлайн-подписками: пользователь отслеживает подписки, пополняет баланс, применяет промокоды, покупает подписки по тарифам и анализирует расходы. Backend — Go (чистая архитектура: `domain` → `usecase` → `service` → `delivery`), PostgreSQL + Redis.

## Запуск стенда (WebLab #4)

Полный стенд (nginx + backend + postgres + redis + swagger-ui + adminer) поднимается одной командой:

```bash
docker compose up -d --build
```

Единая точка входа — nginx на `http://localhost:9090`. Все остальные сервисы внутренние.

| Путь | Что |
|---|---|
| `/` | приветственная страница (эскизы WebLab #1) |
| `/api/v1/`, `/api/v2/` | REST API (v1/v2) |
| `/api/v2/docs/` | Swagger UI |
| `/documentation` | требования WebLab #1 |
| `/legacy` | консольная версия (ППО LR5) |
| `/legacy/download` | скачать бинарник CLI |
| `/status` | статус сервера |
| `/health`, `/ready` | liveness / readiness |
| `/admin/` | Adminer (basic-auth: `admin`/`secret`) |
| `/management` | список всех маршрутов |

Полная таблица маршрутов и обоснования — в `WebLab4.md`.

## Проверки

```bash
make build-cli    # собрать бинарник CLI для /legacy/download
make web-check    # HTTP-проверки стенда (20 проверок: маршруты, коды, статика, /admin)
make api-test     # curl-сценарии REST API v2 (реальный бэкенд)
make lint-api     # линтер OpenAPI (Redocly)
make usecase      # unit-тесты юзкейсов
make integration  # интеграционные тесты (нужна запущенная БД)
```

## Генерация кода из OpenAPI

Спецификация `openapi.yaml` (OpenAPI 3.1) и `doc.yaml` (OpenAPI 3.0) — источники истины. Сгенерированный код (`internal/api/`, `pkg/clients/`, `generated/`) в git не хранится:

```bash
go generate
```

## Структура (кратко)

- `cmd/` — точки входа (HTTP-сервер, CLI поверх API)
- `internal/` — `delivery` (handlers), `usecase`, `service` (репозитории), `domain`, `api` (кодогенерация)
- `deploy/` — `nginx.conf`, статика (`static/`), HTTP-проверки (`http-check.sh`)
- `migrations/` — SQL-миграции Postgres
- `documentation/` — материалы WebLab #1 (диаграммы, эскизы)

Подробнее — `AGENTS.md`, `WebLab1.md` … `WebLab4.md`.
