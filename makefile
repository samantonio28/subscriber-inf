TEST_DIR=.
TEST_USECASE_DIR=./internal/usecase
TEST_SERVICE_DIR=./internal/service

test: integration

usecase:
	go clean -testcache
	@echo "Начало тестирования use case"
	go test -v $(TEST_USECASE_DIR)/...
	@echo "Конец тестирования use case"

integration:
	go clean -testcache
	@echo "Запуск интеграционных тестов (требуется запущенная БД)"
	go test -v $(TEST_SERVICE_DIR) -run ".*Integration.*"
	@echo "Интеграционные тесты завершены"

# OpenAPI-линтер (Redocly) по openapi.yaml
lint-api:
	npx -y @redocly/cli lint openapi.yaml

# Mock-сервер по спецификации (Prism), порт 4010
mock-api:
	npx -y @stoplight/prism-cli mock openapi.yaml --port 4010

# ===== Демонстрация API (реальный бэкенд :8080) =====
API_URL = http://localhost:8080

demo: demo-1 demo-2 demo-3 demo-4 demo-5

demo-1:
	@echo "=== 1. POST /api/v2/users (создать пользователя, валидно) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(API_URL)/api/v2/users -H 'Content-Type: application/json' -d '{"email":"demo@example.com","password":"secret","user_name":"Demo","age":25,"balance":10000}'

demo-2:
	@echo "=== 2. GET /api/v2/subscription-plans (тарифы) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' $(API_URL)/api/v2/subscription-plans

demo-3:
	@echo "=== 3. POST /api/v2/subscriptions (покупка по plan_id) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(API_URL)/api/v2/subscriptions -H 'Content-Type: application/json' -d '{"user_id":"11111111-1111-1111-1111-111111111111","plan_id":1,"start_date":"08-2024"}'

demo-4:
	@echo "=== 4. PATCH /api/v2/subscriptions/1 (применить промокод) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X PATCH $(API_URL)/api/v2/subscriptions/1 -H 'Content-Type: application/json' -d '{"promocode":"SUMMER2025"}'

demo-5:
	@echo "=== 5. POST /api/v2/users (НЕвалидно: нет email) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(API_URL)/api/v2/users -H 'Content-Type: application/json' -d '{"password":"secret","user_name":"Demo","age":25,"balance":10000}'


# ===== API-тесты (curl, реальный бэкенд :8080) =====
api-test: test-1 test-2 test-3 test-4 test-5 test-6 test-7 test-8 test-9 test-10 test-11 test-12 test-13 test-14 test-15

test-1:
	@echo "1. GET /api/v2/subscription-plans"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/subscription-plans

test-2:
	@echo "2. GET /api/v2/subscription-plans/1"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/subscription-plans/1

test-3:
	@echo "3. GET /api/v2/subscription-plans/999999 (ожид. 404)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/subscription-plans/999999

test-4:
	@echo "4. GET /api/v2/services"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/services

test-5:
	@echo "5. GET /api/v2/services/1"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/services/1

test-6:
	@echo "6. GET /api/v2/promocodes"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/promocodes

test-7:
	@echo "7. GET /api/v2/promocodes/1"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/promocodes/1

test-8:
	@echo "8. GET /api/v2/stats/overview"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v2/stats/overview

test-9:
	@echo "9. POST /api/v2/users (ожид. 201)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' -X POST $(API_URL)/api/v2/users -H 'Content-Type: application/json' -d "{\"email\":\"t9$$RANDOM@ex.com\",\"password\":\"secret\",\"user_name\":\"T9\",\"age\":20,\"balance\":1000}"

test-10:
	@echo "10. POST /api/v2/services (ожид. 201)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' -X POST $(API_URL)/api/v2/services -H 'Content-Type: application/json' -d "{\"service_name\":\"Svc$$RANDOM\"}"

test-11:
	@echo "11. POST /api/v2/subscription-plans (ожид. 201)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' -X POST $(API_URL)/api/v2/subscription-plans -H 'Content-Type: application/json' -d "{\"service_id\":1,\"name\":\"Plan$$RANDOM\",\"duration_days\":30,\"price\":$$((RANDOM + 1))}"

test-12:
	@echo "12. POST /api/v2/promocodes (ожид. 201)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' -X POST $(API_URL)/api/v2/promocodes -H 'Content-Type: application/json' -d "{\"service_id\":1,\"value\":\"T12$$RANDOM\",\"discount\":10,\"max_uses\":10}"

test-13:
	@echo "13. GET /api/v2/subscriptions?user_id"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' "$(API_URL)/api/v2/subscriptions?user_id=11111111-1111-1111-1111-111111111111"

test-14:
	@echo "14. GET /api/v2/total-costs"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' "$(API_URL)/api/v2/total-costs?user_id=11111111-1111-1111-1111-111111111111&start_date=01-2026&end_date=12-2026"

test-15:
	@echo "15. GET /api/v1/subscription-plans (старый API v1)"
	@curl -s -o /dev/null -w '   -> HTTP %{http_code}\n' $(API_URL)/api/v1/subscription-plans
