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

# ===== Демонстрация mock-сервера (2-й терминал, mock-api уже поднят) =====
MOCK_URL = http://localhost:4010

demo: demo-1 demo-2 demo-3 demo-4 demo-5

demo-1:
	@echo "=== 1. POST /api/v1/users (создать пользователя, валидно) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(MOCK_URL)/api/v1/users -H 'Content-Type: application/json' -d '{"email":"demo@example.com","password":"secret","user_name":"Demo","age":25,"balance":10000}'

demo-2:
	@echo "=== 2. GET /api/v1/subscription-plans (тарифы) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' $(MOCK_URL)/api/v1/subscription-plans

demo-3:
	@echo "=== 3. POST /api/v1/subscriptions (покупка по plan_id) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(MOCK_URL)/api/v1/subscriptions -H 'Content-Type: application/json' -d '{"user_id":"11111111-1111-1111-1111-111111111111","plan_id":1,"start_date":"08-2024"}'

demo-4:
	@echo "=== 4. PATCH /api/v1/subscriptions/1 (применить промокод) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X PATCH $(MOCK_URL)/api/v1/subscriptions/1 -H 'Content-Type: application/json' -d '{"promocode":"SUMMER2025"}'

demo-5:
	@echo "=== 5. POST /api/v1/users (НЕвалидно: нет email) ==="
	@curl -s -w '\n[HTTP %{http_code}]\n' -X POST $(MOCK_URL)/api/v1/users -H 'Content-Type: application/json' -d '{"password":"secret","user_name":"Demo","age":25,"balance":10000}'

