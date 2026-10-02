TEST_DIR=.
TEST_USECASE_DIR=./internal/usecase
TEST_SERVICE_DIR=./internal/service
ALLURE_OUTPUT_FOLDER=$(CURDIR)/allure-results

# unit-тесты не требуют запущенной БД: интеграционный тест
# (internal/service/sub_repo_integration_test.go) пропускается через testing.Short().
test: unit

unit:
	go clean -testcache
	@echo "Запуск unit-тестов (интеграционные пропускаются через -short)"
	go test -short ./...
	@echo "Unit-тесты завершены"

usecase:
	go clean -testcache
	@echo "Начало тестирования use case"
	go test -v $(TEST_USECASE_DIR)/...
	@echo "Конец тестирования use case"

# Allure-тесты: 10 юзкейсов (лондонские + классические), файл
# internal/usecase/allure_test.go собран с тегом allure; результаты пишутся
# в $(ALLURE_OUTPUT_FOLDER).
allure-test:
	go clean -testcache
	@echo "Запуск Allure-тестов (результаты: $(ALLURE_OUTPUT_FOLDER))"
	ALLURE_OUTPUT_FOLDER="$(ALLURE_OUTPUT_FOLDER)" go test -tags allure -v ./internal/usecase/...
	@echo "Allure-тесты завершены"

allure-report:
	allure generate "$(ALLURE_OUTPUT_FOLDER)" -o reports/allure-report --clean
	@echo "Отчёт: reports/allure-report/index.html"

allure-serve:
	allure serve "$(ALLURE_OUTPUT_FOLDER)"

cover:
	go clean -testcache
	@echo "Покрытие"
	go test -tags allure -cover -run 'TestAllure' -coverprofile=coverage.out ./internal/usecase/...
	go tool cover -func=coverage.out 

cover-html:
	go clean -testcache
	@echo "Покрытие html"
	go test -tags allure -run 'TestAllure' -coverprofile=coverage.out ./internal/usecase/...
	go tool cover -html=coverage.out

integration:
	go clean -testcache
	@echo "Запуск integration-тестов (требуется тестовый стенд: make test-up)"
	go test -p 1 -tags integration -run Integration ./...
	@echo "Integration-тесты завершены"

# E2E-тест демонстрационного сценария (требуется тестовый стенд + Redis).
e2e:
	go clean -testcache
	@echo "Запуск E2E-теста (требуется тестовый стенд: make test-up)"
	go test -p 1 -tags e2e ./test/e2e/...
	@echo "E2E-тест завершён"

# --- Тестовый стенд (ЛР №2): отдельные инстансы PostgreSQL и Redis ---
# Используется docker-compose.test.yml (не трогает дев-стенд из docker-compose.yml).
TEST_COMPOSE = docker-compose.test.yml

test-up:
	docker compose -f $(TEST_COMPOSE) up -d
	@echo "Тестовый стенд поднят: postgres_test -> localhost:8002 (БД test), redis_test -> localhost:6380"

test-down:
	docker compose -f $(TEST_COMPOSE) down

# Сброс тестового хранилища и кэша к чистому состоянию (локальная утилита;
# сами тесты тоже сбрасывают БД через testdb.Reset). Полезно для отката после
# аварийного прерывания прогона.
test-reset:
	@docker exec subscriber-inf-postgres-test psql -U postgres -d test -c \
		"TRUNCATE TABLE user_referrals, payments, cards, subscriptions, promocodes, subscription_plans, user_services, services, users RESTART IDENTITY CASCADE" || true
	@docker exec subscriber-inf-redis-test redis-cli FLUSHALL || true
	@echo "Тестовая БД и Redis очищены"

# Полный прогон в порядке unit → integration → e2e.
# Make останавливается на первой упавшей цели, поэтому если unit упал —
# integration и e2e не запустятся (требование ЛР №2).
test-all: unit integration e2e
	@echo "Все этапы тестирования (unit → integration → e2e) завершены"

# --- SonarQube: анализ качества кода + покрытия (Community Build) ---
# Одноразовая настройка:
#   1) brew install sonar-scanner
#   2) make sonar-up → http://localhost:9000 (admin/admin)
#   3) My Account → Security → Generate Token → export SONAR_TOKEN=<токен>
SONAR_URL=http://localhost:9000

sonar-up:
	docker compose up -d sonarqube
	@echo "Жду готовности SonarQube (до ~3 мин)…"
	@for i in $$(seq 1 36); do \
		if curl -sf http://localhost:9000/api/system/status 2>/dev/null | grep -q '"UP"'; then \
			echo "SonarQube готов."; break; \
		fi; \
		sleep 5; \
	done

sonar-down:
	docker compose down

sonar-scan:
	@echo "Сбор покрытия + запуск sonar-scanner"
	go test -short -tags allure -coverprofile=coverage.out -coverpkg=./... ./...
	sonar-scanner -Dsonar.login=$(SONAR_TOKEN)
	@echo "Готово. Результат: $(SONAR_URL)/dashboard?id=subscriber-inf"

# Одна команда: поднять сервер → покрытие → анализ → открыть дашборд в браузере.
sonar: sonar-up sonar-scan
	open $(SONAR_URL)/dashboard?id=subscriber-inf
	@echo "Дашборд открыт в браузере: $(SONAR_URL)/dashboard?id=subscriber-inf"
