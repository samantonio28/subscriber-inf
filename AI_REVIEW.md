# AI_REVIEW.md

## WebLab #1 — «I'm a startuper (system architect)»

- **Задача:** на базе проекта ППО (`subscriber-inf`) собрать согласованный комплект проектных материалов: цель, требования (функц./нефункц.), use-case, BPMN, сценарии с критериями приёмки, ER, стек, диаграмма БД, C4, экраны, 2 ADR + разбор противоречий.
- **Инструмент/модель:** Claude Code (анализ репозитория, генерация документации и BPMN); ранее для ППО использовался DeepSeek.
- **Ссылки на результат:** `WebLab1.md`, `AGENTS.md`, `documentation/project.md`, `documentation/diagrams/` (use-case, ER, BPMN ×3, C4 L1–L4, диаграмма БД), `documentation/screenshots/` (5 экранов SPA).
- **Проверки:** стек запущен (`docker compose up`), API отвечает `200` (`/subscription-plans`), скриншоты сняты с работающего SPA.
- **Принятое решение:** основная СУБД — PostgreSQL (не MySQL), т.к. миграции/функции/CLI написаны под неё → закреплено в ADR-1.
- **Отклонённое решение:** MySQL как основная СУБД — из-за готовых Postgres-миграций и функций; MySQL оставлен только как экспериментальный адаптер.
- **Уточнение (по разбору противоречий):** связь «подписка↔сервис» ведём через `plan_id`, оставляя `service_name` в API; переделка на `service_id` отложена до WebLab #2.

## WebLab #2 — REST API System Analyst

- **Задача:** на основе требований/сценариев WebLab #1 спроектировать публичный REST API (OpenAPI 3.1, `/api/v1/...`) и проверить его линтером и mock-сервером.
- **Инструмент/модель:** Claude Code (проектирование спека, правки) + Redocly CLI (линтер) + Prism (mock-сервер).
- **Ссылки на результат:** `openapi.yaml` (OpenAPI 3.1.0, 13 путей, 19 схем), `WebLab2.md`.
- **Проверки:** Redocly lint → 0 ошибок; Prism mock → 4-шаговый сценарий прошёл (201/200/201/200).
- **Принятое решение:** действия переосмыслены в REST — `apply-promocode` → `PATCH /subscriptions/{id}` (поле `promocode`), `purchase` → `POST /subscriptions` (`plan_id`), `total_costs` → `GET /total-costs`.
- **Уточнение (управление изменением):** разделены `UserCreate` (вход, с паролем) и `UserResponse` (выход, без пароля) — утечка `password` в ответе устранена.

## WebLab #3 — REST API Developer

- **Задача:** реализовать REST API по `openapi.yaml` (OpenAPI 3.1, `/api/v2`) + Swagger UI + перенос TUI на API + версии v1/v2 + тесты контракта.
- **Инструмент/модель:** Claude Code + oapi-codegen v2.8.0 (кодогенерация), Redocly (линтер), Prism (mock), gomock (тесты).
- **Ссылки на результат:** `openapi.yaml`, `internal/api/v2/`, `internal/delivery/server_v2.go`, `internal/domain/service.go`, `internal/service/service_repo.go`, `cmd/cli/main.go`, `internal/delivery/server_v2_test.go`, `WebLab3.md`.
- **Проверки:** Redocly → 0 ошибок; Prism → 4-шаговый сценарий; curl → все 26 методов v2; CLI → read+write; `go test` → ok; демонстрация намеренного нарушения контракта (красный → зелёный).
- **Принятое решение:** новый API на `/api/v2`, старый оставлен на `/api/v1` и `/` (для SPA); TUI переведён на HTTP-клиент (`pkg/clients/apiv2`).
- **Исправленные баги (ревью):** `PUT /subscriptions` (пустой `service_name`), `PUT /promocodes` (пустой `expires_at`), `total-costs` (фильтр `service_name` отбрасывал всё).
- **Уточнение:** конфиг БД сведён к одному `configs/database.yaml` (удалены `postgres.yaml` и осиротевшая структура `DatabaseConfig`).

## WebLab #4 — WEB SERVER Junior

- **Задача:** настроить асинхронный web-сервер (nginx) как единую точку входа: маршрутизация `/api/v1`/`/api/v2`, Swagger UI, статика, `/documentation`, `/legacy`, `/status`, `/management`, ограниченный `/admin`, `/health`/`/ready`, воспроизводимый запуск через Docker Compose.
- **Инструмент/модель:** Claude Code + nginx (конфиг), Docker Compose, Adminer, curl (HTTP-проверки).
- **Ссылки на результат:** `deploy/nginx.conf`, `deploy/static/*`, `deploy/http-check.sh`, `docker-compose.yml` (полный стенд), `.env.example`, `internal/delivery/app.go` (+`/health`,`/ready`,`PORT`), `WebLab4.md`.
- **Проверки:** `make web-check` → 19/19 зелёных (маршруты, коды, статика, префиксы, `/health`/`/ready`, `/admin` 401/200).
- **Принятое решение:** nginx как единая точка входа, всё остальное — внутреннее; Swagger UI за `/api/v2/docs` через контейнер `swagger-ui` + статическая спецификация (`/openapi.yaml`); `/admin` → Adminer с basic-auth.
- **Исправленные ошибки (запуск/отладка):** (1) вложенный read-only mount `/srv/static/img` → отдельный `/srv/img`; (2) спецификация не отдавалась для Swagger UI → добавлен `location = /openapi.yaml`.
- **Управляемость изменений:** внутренний порт бэкенда `8080 → 9090` (env `PORT` + `proxy_pass backend:9090`) при неизменных внешних URL — проверки остались зелёными; финал на `8080`.

## WebLab #5 — WEB SERVER Senior DevOps

- **Задача:** расширить стенд до v2.0: балансировка чтения 2:1:1, запись на один узел, read-only роли БД, `/mirror` за standby-репликой, кэш+gzip, подмена `Server`, `/monitoring` (Prometheus+Loki+Grafana), нагрузка k6, сценарий отказа.
- **Инструмент/модель:** Claude Code + openresty (nginx+headers-more) + PostgreSQL streaming replication + k6 + Prometheus/Loki/Promtail/Grafana.
- **Ссылки на результат:** `deploy/nginx.conf`, `deploy/http-check.sh` (40 проверок), `deploy/k6/load.js`, `deploy/prometheus|promtail|loki|grafana/*`, `deploy/postgres-replica/entrypoint.sh`, `migrations/000_replication_hba.sh`, `migrations/014_readonly_role.up.sql`, `internal/delivery/metrics.go` + `middlewares.go` + `app.go`, `internal/logger/logger.go`, `Dockerfile` (бинарник), `pkg/config/env.go` (`POOL_MAX_CONNS`), `WebLab5.md`.
- **Проверки:** `make web-check` 40/40; 200 GET → `100/50/50` (2:1:1); POST → только `backend`; POST на read-only → 403; репликация `state=streaming` + запись в реплику → «read-only transaction»; k6 402 тыс. запросов, 0 ошибок, p95 60 мс; Prometheus 5/5 targets up; сценарий отказа без ошибок.
- **Принятое решение:** primary-replica (streaming) для `/mirror` + read-only БД-роль `reader` для основных read-инстансов — два разных механизма «только чтение» (роль на primary и физический standby).
- **Исправленные баги (запуск/отладка):** (1) миграция `011` дублировала RLS-политики `010` и роняла цепочку миграций → сделана идемпотентной, иначе `012/013/014` не применялись; (2) в реплике `grep primary_conninfo` находил закомментированную строку → `^primary_conninfo`; (3) compose-`entrypoint` не наследует CMD → `exec docker-entrypoint.sh postgres` явно; (4) `max_fails=1` выбивал read-only из ротации → `max_fails=3`; (5) `go run` хрупок → бинарник; (6) пул БД 10 → `POOL_MAX_CONNS=40` на write-инстансах; (7) Grafana под-subpath redirect-петля → `proxy_pass` без среза.


