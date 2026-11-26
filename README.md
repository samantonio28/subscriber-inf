# subscriber-inf

Сервис управления онлайн-подписками: пользователь отслеживает подписки, пополняет баланс, применяет промокоды, покупает подписки по тарифам и анализирует расходы. Backend — Go (чистая архитектура: `domain` → `usecase` → `service` → `delivery`), PostgreSQL + Redis.

## Запуск стенда (WebLab #5)

Полный стенд поднимается одной командой:

```bash
docker compose up -d --build
```

Единая точка входа — nginx (openresty) на `http://localhost:9090`. Всё остальное внутреннее.

Состав: nginx + `backend` (запись, вес 2) + `backend-read1`/`backend-read2` (только чтение, роль `reader`) +
`mirror-backend`/`mirror-backend-read` (отдельная версия за репликой) + `postgres` (primary) + `postgres-replica`
(standby) + redis + swagger-ui + adminer + prometheus + loki + promtail + grafana.

| Путь | Что |
|---|---|
| `/` | приветственная страница (эскизы WebLab #1) |
| `/api/v1/`, `/api/v2/` | REST API (v1/v2); **чтение балансируется 2:1:1**, запись на один узел |
| `/api/v2/docs/` | Swagger UI |
| `/mirror/`, `/mirror/api/...` | **отдельная версия** приложения за standby-репликой БД |
| `/monitoring/` | Grafana (метрики Prometheus + логи Loki со всех инстансов) |
| `/documentation` | требования WebLab #1 |
| `/legacy`, `/legacy/download` | консольная версия + бинарник CLI |
| `/status` | статус сервера |
| `/health`, `/ready` | liveness / readiness |
| `/admin/` | Adminer (basic-auth: `admin`/`secret`) |
| `/management` | список всех маршрутов |

Полные таблицы маршрутов и обоснования — в `WebLab4.md` и `WebLab5.md`.

## Проверки

```bash
make build-cli    # собрать бинарник CLI для /legacy/download
make web-check    # HTTP-проверки стенда (40 проверок: маршруты, /mirror, /monitoring, кэш, gzip, Server)
make load-test    # нагрузочный тест k6 (внутри compose-сети)
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
- `internal/` — `delivery` (handlers, метрики, middleware), `usecase`, `service` (репозитории), `domain`, `api` (кодогенерация)
- `deploy/` — `nginx.conf`, статика (`static/`), HTTP-проверки (`http-check.sh`), `k6/`, `prometheus/`, `promtail/`, `loki/`, `grafana/`, `postgres-replica/`
- `migrations/` — SQL-миграции Postgres (в т.ч. `014` — роли `reader`/`replicator`)
- `documentation/` — материалы WebLab #1 (диаграммы, эскизы)

Подробнее — `AGENTS.md`, `WebLab1.md` … `WebLab5.md`.
