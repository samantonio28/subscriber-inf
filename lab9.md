Эндпоинт 1: Получить статистику из БД
```bash
curl -X GET http://localhost:8080/lab9/1 | jq
```
Эндпоинт 2: Получить статистику через Redis кэш
```bash
curl -X GET http://localhost:8080/lab9/2 | jq
```
Эндпоинт 3: Получить метрики производительности
```bash
curl -X GET http://localhost:8080/lab9/3 | jq
```
Эндпоинт 4: Запустить тест производительности
Без изменений данных:
```bash
curl -X POST http://localhost:8080/lab9/4 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "no_changes_test",
    "description": "Performance test without data changes",
    "interval": 5,
    "data_change": "none"
  }' | jq
```
С добавлением данных:

```bash
curl -X POST http://localhost:8080/lab9/4 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "with_additions",
    "description": "Performance test with data additions every 10s",
    "interval": 5, 
    "data_change": "add"
  }' | jq
```
С удалением данных:

```bash
curl -X POST http://localhost:8080/lab9/4 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "with_deletions",
    "description": "Performance test with data deletions every 10s",
    "interval": 5,
    "data_change": "delete"
  }' | jq
```
С обновлением данных:

```bash
curl -X POST http://localhost:8080/lab9/4 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "with_updates", 
    "description": "Performance test with data updates every 10s",
    "interval": 5,
    "data_change": "update"
  }' | jq
```
Эндпоинт 5: Остановить тест производительности
```bash
curl -X POST http://localhost:8080/lab9/5 | jq
```
Эндпоинт 6: Добавить тестовые данные
```bash
curl -X POST http://localhost:8080/lab9/6 | jq
```
Эндпоинт 7: Очистить тестовые данные
```bash
curl -X POST http://localhost:8080/lab9/7 | jq
```
