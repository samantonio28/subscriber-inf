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
