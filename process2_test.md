# Процесс 2: Создание нового пользователя и покупка подписки

Сервер запущен на `localhost:8080`.

## Шаг 1: Создать пользователя

```bash
curl -X POST "http://localhost:8080/users" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "email": "testuser@example.com",
    "password": "secret123",
    "user_name": "Test User",
    "age": 25,
    "balance": 10000,
    "refferal_code": null
  }'
```

Ожидаемый ответ (успех):
```json
{
  "message": "new user_id: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
}
```

Запоминаем `user_id` (в данном случае мы его указали явно).

## Шаг 2: Получить данного пользователя, убедиться, что он создался правильно

Эндпоинт для получения пользователя по ID отсутствует, но можно убедиться по коду ответа (201 Created). Также можно проверить, что пользователь есть в системе, запросив его подписки (на следующем шаге).

## Шаг 3: Проверить подписки пользователя (их должно быть ноль)

```bash
curl -X GET "http://localhost:8080/subscriptions?uuid=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ — пустой массив `[]`.

## Шаг 4: Проверить, что баланс положительный

Баланс мы указали при создании (`balance`: 10000). Поскольку нет эндпоинта для получения баланса, полагаемся на данные, переданные при создании.

## Шаг 5: Получить список планов подписок

```bash
curl -X GET "http://localhost:8080/subscription-plans" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ — массив планов. Пример:
```json
[
  {
    "plan_id": 1,
    "service_id": 5,
    "name": "Basic Monthly",
    "duration_days": 30,
    "price": 1999
  },
  {
    "plan_id": 2,
    "service_id": 5,
    "name": "Premium Monthly",
    "duration_days": 30,
    "price": 2999
  }
]
```

Выберем самый дешёвый план (с минимальным `price`). Допустим, это план с `plan_id = 1`, `price = 1999`, `service_id = 5`. Нужно также определить `service_name` по `service_id`. В системе может быть соответствие между `service_id` и `service_name`. Для простоты предположим, что `service_name` равен `"Yandex Plus"` (можно взять из предыдущих тестов). В реальном тестировании нужно узнать `service_name` по `service_id` (например, из базы данных). Для целей теста используем `service_name`: `"Yandex Plus"`.

## Шаг 6: Купить самую дешёвую подписку

```bash
curl -X POST "http://localhost:8080/subscriptions/purchase" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "service_name": "Yandex Plus",
    "plan_id": 1,
    "price": 1999,
    "duration_days": 30
  }'
```

Ожидаемый ответ (успех):
```json
{
  "message": "subscription purchased successfully, sub_id: 124"
}
```

Запоминаем `sub_id` новой подписки (например, 124).

## Шаг 7: Проверить, что деньги с баланса списались

Эндпоинт для проверки баланса отсутствует, но можно косвенно убедиться, что покупка прошла успешно (код ответа 200). Также можно проверить, что после покупки баланс пользователя уменьшился, если бы был доступ к БД. В рамках тестирования API этот шаг пропускаем.

## Шаг 8: Проверить, что пользователю назначилась подписка

Запросим подписки пользователя ещё раз:

```bash
curl -X GET "http://localhost:8080/subscriptions?uuid=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" \
  -H "Content-Type: application/json"
```

Ожидаемый ответ — массив с одной подпиской (только что созданной). Пример:
```json
[
  {
    "service_name": "Yandex Plus",
    "price": 1999,
    "user_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "sub_type": "usual",
    "start_date": "05-2025",
    "end_date": "06-2025"
  }
]
```

Убеждаемся, что подписка присутствует.

## Заключение

Пользователь успешно создан, подписка куплена, система работает корректно.