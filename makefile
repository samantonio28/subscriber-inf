TEST_DIR=.
TEST_USECASE_DIR=./internal/usecase

test: usecase

usecase:
	@echo "Начало тестирования"
	go test -v $(TEST_USECASE_DIR)/...
	@echo "Конец тестирования"
