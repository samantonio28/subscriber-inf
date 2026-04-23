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
