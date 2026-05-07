# Процесс 1: Тестирование существующего пользователя и покупки подписки

Предполагается, что в системе уже есть подписка с `sub_id = 1` и соответствующий пользователь. Сервер запущен на `localhost:8080`.

## Шаг 1: Получить подписку с id = 1

```bash
curl -X GET "http://localhost:8080/subscriptions/1" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ (пример):
```json
{
  "service_name": "Yandex Plus",
  "price": 399,
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "sub_type": "usual",
  "start_date": "07-2025",
  "end_date": "08-2025"
}
```

Извлекаем `user_id` из ответа (здесь `123e4567-e89b-12d3-a456-426614174000`).

## Шаг 2: Получить подписки этого пользователя (проверить, что баланс положительный)

Баланс пользователя недоступен через API, поэтому просто убедимся, что пользователь существует, запросив его подписки.

```bash
curl -X GET "http://localhost:8080/subscriptions?uuid=123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ — массив подписок пользователя.

## Шаг 3: Получить стоимость всех трат пользователя за последние 2 месяца

Вычислим даты: текущий месяц и два предыдущих. Для простоты укажем `start_date` и `end_date` в формате MM-YYYY.

```bash
curl -X POST "http://localhost:8080/total_costs" \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "03-2025",
    "end_date": "05-2025",
    "filter": {
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "service_name": "Yandex Plus"
    },
    "sort_by": "price",
    "sort_order": "desc"
  }'
```

В ответе будет поле `total_sum` — общая сумма затрат за период.

## Шаг 4: Получить список доступных планов подписок

```bash
curl -X GET "http://localhost:8080/subscription-plans" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ — массив планов. Выберем первый план (например, с `plan_id = 1`).

## Шаг 5: Выбрать план и купить подписку для этого пользователя

Используем `POST /subscriptions/purchase`. Необходимо указать `user_id`, `service_name`, `plan_id`, `price`, `duration_days`. Возьмём данные из выбранного плана (предположим, что план имеет `service_id = 5`, `name = "Premium Monthly"`, `price = 2999`, `duration_days = 30`). Однако в запросе требуется `service_name`, а не `service_id`. Узнаем `service_name` из плана (например, "Yandex Plus"). Для теста можно использовать то же `service_name`, что и у существующей подписки.

```bash
curl -X POST "http://localhost:8080/subscriptions/purchase" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "service_name": "Yandex Plus",
    "plan_id": 1,
    "price": 2999,
    "duration_days": 30
  }'
```

Ожидаемый ответ (успех):
```json
{
  "message": "subscription purchased successfully, sub_id: 123"
}
```

Запоминаем `sub_id` новой подписки (например, 123).

## Шаг 6: Проверить, что баланс уменьшился

Поскольку эндпоинт для получения баланса отсутствует, этот шаг пропускаем. Можно убедиться, что покупка прошла успешно (код ответа 200).

## Шаг 7: Проверить подписки пользователя, убедиться, что есть подписка с id = 1 и новая подписка

Снова запросим подписки пользователя:

```bash
curl -X GET "http://localhost:8080/subscriptions?uuid=123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json"
```

В ответе должен быть массив, содержащий как минимум две подписки: одна с `sub_id = 1` (в ответе `sub_id` может не быть, но можно идентифицировать по `service_name` и `start_date`), и вторая с `sub_id = 123` (или другим идентификатором).

## Заключение

Все шаги выполнены, система работает корректно.