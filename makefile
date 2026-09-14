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

integration:
	go clean -testcache
	@echo "Запуск интеграционных тестов (требуется запущенная БД)"
	go test -v $(TEST_SERVICE_DIR) -run ".*Integration.*"
	@echo "Интеграционные тесты завершены"
