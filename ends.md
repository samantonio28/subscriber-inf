# Анализ эндпоинтов API (doc.yaml) и их использование во фронтенде

## Список эндпоинтов согласно doc.yaml

| Метод | Путь | Описание | Используется во фронтенде | Файл фронтенда | Примечания |
|-------|------|----------|---------------------------|----------------|------------|
| POST | `/subscriptions` | Добавить новую подписку | Да (создание подписки) | `Subscriptions.vue` (модальное окно) | Вызов через `subscriptionStore.createSubscription` |
| GET | `/subscriptions` | Получить подписки по user_id | Да (список подписок) | `Dashboard.vue`, `Subscriptions.vue`, `Analytics.vue` | Вызов через `subscriptionStore.fetchSubscriptions` |
| GET | `/subscriptions/{id}` | Получить подписку по ID | Нет | - | Не используется, но есть юзкейс `get_sub.go` |
| PUT | `/subscriptions/{id}` | Обновить подписку | Да (редактирование) | `Subscriptions.vue` (кнопка Edit) | Вызов через `subscriptionStore.updateSubscription` |
| DELETE | `/subscriptions/{id}` | Удалить подписку | Да (отмена) | `Subscriptions.vue` (кнопка Cancel) | Вызов через `subscriptionStore.deleteSubscription` |
| POST | `/subscriptions/{id}/apply-promocode` | Применить промокод к подписке | Да (применение промокода) | `Promocodes.vue` (секция Apply Promocode) | Вызов через `promocodeStore.applyPromocode` |
| POST | `/total_costs` | Получить общие затраты за период | Да (аналитика) | `Analytics.vue` (функция `fetchCosts`) | Вызов через `subscriptionStore.fetchTotalCosts` |
| POST | `/promocodes` | Создать промокод | Да (создание промокода) | `Promocodes.vue` (модальное окно) | Вызов через `promocodeStore.createPromocode` |
| GET | `/promocodes` | Получить промокоды (с фильтрацией) | Да (список промокодов) | `Promocodes.vue` (таблица) | Вызов через `promocodeStore.fetchPromocodes` |
| GET | `/promocodes/{id}` | Получить промокод по ID | Нет | - | Не используется, но есть юзкейс `get_promocode.go` |
| PUT | `/promocodes/{id}` | Обновить промокод | Нет | - | Не используется, но есть юзкейс `update_promocode.go` |
| DELETE | `/promocodes/{id}` | Удалить промокод | Да (удаление) | `Promocodes.vue` (кнопка Delete) | Вызов через `promocodeStore.deletePromocode` |
| GET | `/promocodes/code/{code}` | Получить промокод по коду | Нет | - | Не используется, но есть юзкейс `get_promocode.go` (метод GetByCode) |
| POST | `/subscription-plans` | Создать план подписки | Нет | - | Не используется, но есть юзкейс `create_subscription_plan.go` |
| GET | `/subscription-plans` | Получить планы подписки (с фильтрацией) | Нет | - | Не используется, но есть юзкейс `get_filtered_subscription_plans.go` |
| GET | `/subscription-plans/{id}` | Получить план подписки по ID | Нет | - | Не используется, но есть юзкейс `get_subscription_plan.go` |
| PUT | `/subscription-plans/{id}` | Обновить план подписки | Нет | - | Не используется, но есть юзкейс `update_subscription_plan.go` |
| DELETE | `/subscription-plans/{id}` | Удалить план подписки | Нет | - | Не используется, но есть юзкейс `delete_subscription_plan.go` |
| GET | `/stats/overview` | Получить сводную статистику | Нет | - | Не используется, но есть юзкейс `stats_overview.go` (заглушка) |

## Проблемы соответствия (ошибки 400)

### 1. DELETE /promocodes/{id}
Фронтенд отправляет DELETE запрос на `/promocodes/{id}`. Согласно спецификации, ответ при успешном удалении должен быть **204 No Content** (без тела). Однако в обработчике `DeletePromocode` в `server.go` возвращается `200 OK` с JSON-сообщением. Это может вызывать ошибку 400, если фронтенд ожидает пустой ответ.

**Решение**: Исправить обработчик, чтобы возвращал `204 No Content` без тела, либо согласовать с фронтендом ожидание JSON.

### 2. POST /subscriptions
Фронтенд отправляет поля `service_name`, `price`, `start_date` (формат MM-YYYY). Согласно спецификации, тело запроса должно соответствовать схеме `Subscription`, которая включает `user_id`, `service_name`, `price`, `sub_type`, `start_date`, `end_date`. Возможно, отсутствие `user_id` или неверный формат даты вызывает ошибку 400.

**Решение**: Проверить, что фронтенд отправляет все обязательные поля, и преобразовать формат даты.

### 3. POST /subscriptions/{id}/apply-promocode
Фронтенд отправляет `promocode` и `subscriptionId`. Спецификация ожидает JSON с полем `promocode`. Обработчик ожидает `promocode` в теле. Возможно, ошибка 400 возникает из-за неверного формата (например, отправка `subscriptionId` в теле, хотя он уже в пути).

**Решение**: Убедиться, что фронтенд отправляет корректный JSON: `{ "promocode": "CODE" }`.

### 4. GET /promocodes
Фронтенд вызывает `fetchPromocodes` без параметров. Спецификация допускает query-параметры `serviceId` и `status`. Если бэкенд ожидает эти параметры, но фронтенд не передаёт, может возникать ошибка 400? Нет, параметры необязательные. Однако наш обработчик `GetPromocodes` использует юзкейс `GetFilteredPromocodesUC`, который ожидает фильтр. Если фильтр пустой, возвращаются все промокоды. Ошибка 400 маловероятна.

### 5. GET /subscriptions
Фронтенд передаёт `uuid` как query-параметр. Спецификация требует `uuid` (строка UUID). Фронтенд использует жёстко заданный `USER_ID`. Возможно, ошибка 400 из-за неверного формата UUID? `USER_ID` = `'4228d7d5-2736-431f-950b-c169d5a77302'` — корректный UUID.

## Рекомендации по исправлению

1. **Исправить обработчик DELETE /promocodes/{id}** в `server.go`:
   - Вернуть `204 No Content` без тела.
   - Либо оставить `200 OK` с JSON, но фронтенд должен ожидать JSON.

2. **Проверить формат дат** в POST /subscriptions:
   - Преобразовать `start_date` из MM-YYYY в RFC3339 (или дату с первым днём месяца).
   - Добавить `user_id` из константы USER_ID.

3. **Проверить тело запроса apply-promocode**:
   - Убедиться, что фронтенд отправляет только `promocode`.

4. **Обновить stores фронтенда** для соответствия спецификации:
   - `promocodeStore.deletePromocode` должен обрабатывать 204.
   - `subscriptionStore.createSubscription` должен включать все обязательные поля.

5. **Добавить обработку ошибок 400** во фронтенде для отображения сообщений от сервера.

## Выводы
Большинство эндпоинтов используются фронтендом, но есть расхождения в форматах запросов/ответов. Основные источники ошибок 400:
- Несоответствие формата даты.
- Отсутствие обязательных полей.
- Неверный код статуса ответа.

Исправление этих несоответствий устранит ошибки 400.