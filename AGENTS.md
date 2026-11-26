# AGENTS.md — контекст для кодирующего агента

## Цель проекта

**Сервис подписок (Subscriber Inf)** — web-приложение для управления онлайн-подписками: пользователь отслеживает подписки, пополняет баланс, применяет промокоды, покупает подписки по тарифам и анализирует расходы; администратор управляет пользователями, тарифами и промокодами.

## Словарь предметной области

| Термин | Сущность/значение |
|--------|-------------------|
| Пользователь | `User` — владелец подписок и баланса |
| Подписка | `Subscription` — запись об оформленной подписке (`sub_type`: usual/promocode/family) |
| Сервис | `Service` — онлайн-сервис, на который оформляется подписка |
| Тариф (план) | `SubscriptionPlan` — шаблон подписки (сервис + длительность + цена) |
| Промокод | `Promocode` — код на скидку (дискаунт, лимит, срок) |
| Платёж | `Payment` — пополнение/списание средств |
| Реферал | `UserReferral` — реферальная связь между пользователями |

## Границы системы

**Входит в scope:** CRUD подписок, применение промокодов, покупка по тарифу, расчёт затрат, промокоды, тарифы, пользователи, сводная статистика.

**Вне scope (на текущий этап):**
- авторизация/разграничение ролей — WebLab #3+;
- декомпозиция на сервисы (SOA) — WebLab #6;
- полноценный UI/UX — WebLab #7–#9.

## Ссылки на требования и ADR

- `WebLab1.md` — цель, требования (функц./нефункц.), сценарии, диаграммы.
- ADR — раздел 11 в `WebLab1.md` (ADR-1: PostgreSQL+Redis; ADR-2: модульный монолит).
- `doc.yaml` — OpenAPI 3.0, текущий источник истины для кодогенерации (`go generate`).
- `openapi.yaml` — REST OpenAPI 3.1 (`/api/v2`), спроектирован в WebLab #2 (`WebLab2.md`), реализован в WebLab #3 (`WebLab3.md`).
- `internal/api/v2/`, `internal/delivery/server_v2.go` — реализация нового API; `cmd/cli` — CLI поверх HTTP-клиента.
- `documentation/project.md`, `documentation/diagrams/` — материалы из ППО.
- `WebLab5.md` — итоговый документ WebLab #5 (балансировка 2:1:1, `/mirror`, репликация, кэш/gzip, `/monitoring`, k6).

## Структура репозитория

```
subscriber-inf/               # бэкенд (Go)
  cmd/main.go                 # точка входа HTTP-сервера
  cmd/cli/main.go             # CLI-интерфейс
  internal/domain/            # сущности и интерфейсы
  internal/usecase/           # бизнес-логика
  internal/service/           # репозитории (Postgres) + mysql/ (адаптер MySQL)
  internal/delivery/          # HTTP-обработчики (server.go и др.)
  internal/api/               # сгенерированный код (gitignored, go generate)
  migrations/                 # SQL-миграции Postgres
  migrations_mysql/           # MySQL-схема (экспериментально)
  configs/                    # конфиги (database.yaml — postgres)
  deploy/                     # WebLab #4+#5: nginx.conf, static/, http-check.sh, k6/, prometheus/, promtail/, loki/, grafana/, postgres-replica/
  documentation/              # документация, диаграммы, скриншоты
  WebLab1.md                  # итоговый документ WebLab #1
  WebLab4.md                  # итоговый документ WebLab #4 (стенд, маршруты)
  WebLab5.md                  # итоговый документ WebLab #5 (балансировка, /mirror, репликация, мониторинг)
subscriber-inf-frontend/      # Vue 3 SPA (отдельный репозиторий)
```

## Команды запуска и проверки

```bash
# Стенд WebLab #5 (nginx/openresty — единая точка входа :9090; всё остальное внутреннее)
# Состав: nginx + backend(запись) + backend-read1/read2(только чтение, роль reader)
#         + mirror-backend/mirror-backend-read(/mirror за репликой) + postgres(primary)
#         + postgres-replica(standby) + redis + swagger-ui + adminer
#         + prometheus + loki + promtail + grafana(/monitoring)
cd ~/study/web/subscriber-inf && docker compose up -d --build

# HTTP-проверки стенда (40 проверок: маршруты, /mirror, /monitoring, кэш, gzip, Server)
cd ~/study/web/subscriber-inf && make web-check

# Нагрузочный тест k6 (внутри compose-сети, бьёт по nginx:80)
cd ~/study/web/subscriber-inf && make load-test

# Наблюдаемость: Grafana на http://localhost:9090/monitoring/ (admin/admin)
# Prometheus метрики инстансов — по /metrics (backend:8080, read1:8081, read2:8082, mirror*)

# Полный стек dev (из папки web/): backend 8080, frontend 3000, postgres 8001, redis 6380
cd ~/study/web && docker compose up -d --build

# Наполнение БД тестовыми данными (postgres внутренний — запускать ВНУТРИ контейнера)
docker compose exec backend sh -c "cd scripts && go run gen_data.go"

# Генерация кода из OpenAPI (internal/api, pkg/clients, generated)
cd ~/study/web/subscriber-inf && go generate

# Тесты юзкейсов (unit)
cd ~/study/web/subscriber-inf && make usecase

# Интеграционные тесты (требуют запущенную БД)
cd ~/study/web/subscriber-inf && make integration

# OpenAPI-линтер (WebLab #2)
cd ~/study/web/subscriber-inf && make lint-api

# Mock-сервер (Prism) на :4010 — 1-й терминал, оставить висеть
cd ~/study/web/subscriber-inf && make mock-api

# Демонстрация сценария — 2-й терминал (mock уже поднят)
cd ~/study/web/subscriber-inf && make demo     # все 5 шагов подряд
cd ~/study/web/subscriber-inf && make demo-1   # ...или пошагово: demo-1 .. demo-5
```

## Границы изменений / правила

- `internal/api/`, `pkg/clients/`, `generated/` — сгенерированный код, руками не править (перегенерировать `go generate`).
- Источник истины по API — `doc.yaml`; при изменении контракта сначала править его, потом код.
- Секреты и реальные персональные данные в задания агенту не включать.
- **WebLab #5:** env бэкенда — `INSTANCE` (имя для метрик/`X-Backend`), `READ_ONLY=true` (блок записи 403),
  `POOL_MAX_CONNS` (размер пула БД), `DB_USER`/`DB_PASSWORD` (роль: `postgres` на запись, `reader` на чтение).
- **Роли БД:** `reader` (только SELECT, миграция `014`) — для `backend-read1/read2`; `replicator` — для стриминга.
- **Репликация:** `postgres` = primary, `postgres-replica` = standby (entrypoint `deploy/postgres-replica/entrypoint.sh`);
  `/mirror` читает с реплики, пишет в primary. При пересборке бэкендов — `docker compose restart nginx`
  (nginx кэширует IP upstream-контейнеров).
