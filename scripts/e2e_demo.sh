#!/usr/bin/env bash
#
# Демонстрация E2E-сценария с помощью curl + захват трафика (ЛР №2, п.5).
#
# Сценарий повторяет test/e2e/subscription_flow_e2e_test.go, но вместо Go-теста
# выполняет те же HTTP-запросы утилитой curl, а «средством захвата трафика»
# выступает access-лог приложения (AccessLogMiddleware пишет каждый запрос в
# logs/access.log). Это тот же механизм, что и в E2E-тесте.
#
# Требования: поднят тестовый стенд (`make test-up`). Сервер стартует здесь же
# против тестового стенда (configs/postgres_test.yaml + redis_test), поэтому
# дев-стенд не затрагивается.
#
# Порядок шагов: покупка → список → стоимость → промокод → применение скидки.

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_UUID="11111111-1111-1111-1111-111111111111"
CUSTOMER_UUID="22222222-2222-2222-2222-222222222222"
ACCESS_LOG="logs/access.log"

echo "==> Сброс тестового стенда"
docker exec subscriber-inf-postgres-test psql -U postgres -d test -c \
  "TRUNCATE TABLE user_referrals, payments, cards, subscriptions, promocodes, subscription_plans, user_services, services, users RESTART IDENTITY CASCADE" >/dev/null
docker exec subscriber-inf-redis-test redis-cli FLUSHALL >/dev/null

echo "==> Сидирование: сервис, план, админ, покупатель"
SERVICE_ID=$(docker exec -i subscriber-inf-postgres-test psql -U postgres -d test -tA -c \
  "INSERT INTO services (service_name) VALUES ('Netflix') RETURNING service_id" | head -n1)
PLAN_ID=$(docker exec -i subscriber-inf-postgres-test psql -U postgres -d test -tA -c \
  "INSERT INTO subscription_plans (service_id, name, duration_days, price) VALUES ($SERVICE_ID, 'Netflix Basic', 30, 299) RETURNING plan_id" | head -n1)
docker exec subscriber-inf-postgres-test psql -U postgres -d test -c \
  "INSERT INTO users (user_id, email, password, user_name, age, balance, referral_code, role) VALUES
     ('$CUSTOMER_UUID', 'customer@example.com', 'secret', 'Customer', 25, 1000, 'CUSTREF1', 'user'),
     ('$ADMIN_UUID',    'admin@example.com',    'secret', 'Admin',    30, 0,    'ADMINREF1', 'admin')" >/dev/null

echo "==> Запуск сервера против тестового стенда"
go build -o /tmp/subscriber-inf-test cmd/main.go
CONFIG_PATH=configs/postgres_test.yaml REDIS_URL=localhost:6380 /tmp/subscriber-inf-test &
SERVER_PID=$!
trap 'kill $SERVER_PID 2>/dev/null || true' EXIT

echo "==> Ожидание готовности сервера"
for _ in $(seq 1 30); do
  if nc -z localhost 8080 2>/dev/null; then break; fi
  sleep 1
done

echo ""
echo "=== 1. Покупка подписки (customer) ==="
curl -s -X POST "$BASE_URL/subscriptions/purchase" \
  -H "Content-Type: application/json" -H "X-User-ID: $CUSTOMER_UUID" \
  -d "{\"user_id\":\"$CUSTOMER_UUID\",\"service_name\":\"Netflix\",\"plan_id\":$PLAN_ID,\"price\":299,\"duration_days\":30}"; echo

echo "=== 2. Список подписок (customer) ==="
curl -s "$BASE_URL/subscriptions?uuid=$CUSTOMER_UUID" -H "X-User-ID: $CUSTOMER_UUID"; echo

echo "=== 3. Общая стоимость за период (customer) ==="
START_DATE=$(date +%m-%Y)
END_DATE=$(date -v+1m +%m-%Y 2>/dev/null || date -d 'next month' +%m-%Y)
curl -s -X POST "$BASE_URL/total_costs" -H "Content-Type: application/json" -H "X-User-ID: $CUSTOMER_UUID" \
  -d "{\"start_date\":\"$START_DATE\",\"end_date\":\"$END_DATE\",\"filter\":{\"service_name\":\"Netflix\",\"user_id\":\"$CUSTOMER_UUID\"}}"; echo

echo "=== 4. Создание промокода (admin) ==="
curl -s -X POST "$BASE_URL/promocodes" -H "Content-Type: application/json" -H "X-User-ID: $ADMIN_UUID" \
  -d "{\"service_id\":$SERVICE_ID,\"value\":\"SAVE20\",\"discount\":20,\"max_uses\":5,\"duration_days\":30}"; echo

echo "=== 5. Применение промокода к подписке (customer) ==="
SUB_ID=$(docker exec subscriber-inf-postgres-test psql -U postgres -d test -tA -c "SELECT sub_id FROM subscriptions ORDER BY sub_id LIMIT 1")
curl -s -X POST "$BASE_URL/subscriptions/$SUB_ID/apply-promocode" \
  -H "Content-Type: application/json" -H "X-User-ID: $CUSTOMER_UUID" \
  -d '{"promocode":"SAVE20"}'; echo

echo ""
echo "=== Захваченный трафик (logs/access.log, последние 10 запросов) ==="
tail -n 10 "$ACCESS_LOG" 2>/dev/null || echo "(лог доступа не найден — сервер мог не успеть записать)"
