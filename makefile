TEST_DIR=.
COVERAGE_TMP=coverage.out.tmp
FILES_TO_CLEAN=*.out.tmp  

test:
	@echo "Начало тестирования"
	go test -v -race -coverpkg=./... -coverprofile=$(COVERAGE_TMP) $(TEST_DIR)/...
	rm -f $(FILES_TO_CLEAN)
	@echo "Конец тестирования"
