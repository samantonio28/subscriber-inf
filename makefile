TEST_DIR=.

test:
	@echo "Начало тестирования"
	go test -v $(TEST_DIR)/...
	@echo "Конец тестирования"
