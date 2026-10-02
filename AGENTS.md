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
- `openapi.yaml` — целевой REST OpenAPI 3.1 (`/api/v1`), результат WebLab #2; `WebLab2.md` — его документация.
- `documentation/project.md`, `documentation/diagrams/` — материалы из ППО.

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
  documentation/              # документация, диаграммы, скриншоты
  WebLab1.md                  # итоговый документ WebLab #1
subscriber-inf-frontend/      # Vue 3 SPA (отдельный репозиторий)
```

## Команды запуска и проверки

```bash
# Полный стек (из папки web/): backend 8080, frontend 3000, postgres 8001, redis 6380
cd ~/study/web && docker compose up -d --build

# Наполнение БД тестовыми данными
cd ~/study/web/subscriber-inf/scripts && go run gen_data.go

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
