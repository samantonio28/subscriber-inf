#!/usr/bin/env bash
# HTTP-проверки стенда WebLab #5 (nginx/openresty — единая точка входа).
# Запуск: bash deploy/http-check.sh [--no-wait]
# По умолчанию BASE_URL=http://localhost:9090; перед проверками ждёт /ready.
set -u

BASE_URL="${BASE_URL:-http://localhost:9090}"
if [[ "${1:-}" == "--no-wait" ]]; then WAIT=0; else WAIT=1; fi

pass=0
fail=0

check() {
  local name="$1" expected="$2" path="$3"
  shift 3
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' "$@" "${BASE_URL}${path}")
  local ok=0 e
  for e in $expected; do
    [[ "$code" == "$e" ]] && ok=1
  done
  if [[ $ok -eq 1 ]]; then
    printf '  PASS  %-40s -> %s\n' "$name" "$code"
    pass=$((pass+1))
  else
    printf '  FAIL  %-40s -> ожидалось [%s], получено %s\n' "$name" "$expected" "$code"
    fail=$((fail+1))
  fi
}

# Проверка значения заголовка ответа (contains).
check_header() {
  local name="$1" header="$2" want="$3" path="$4"
  shift 4
  local val
  val=$(curl -s -D- -o /dev/null "$@" "${BASE_URL}${path}" | tr -d '\r' \
        | grep -i "^${header}:" | head -1 | sed "s/^[^:]*: *//I")
  if [[ "$val" == *"$want"* ]]; then
    printf '  PASS  %-40s -> %s: %s\n' "$name" "$header" "$val"
    pass=$((pass+1))
  else
    printf '  FAIL  %-40s -> %s: ожидалось [%s], получено [%s]\n' "$name" "$header" "$want" "$val"
    fail=$((fail+1))
  fi
}

if [[ $WAIT -eq 1 ]]; then
  echo "Ожидание готовности ${BASE_URL}/ready ..."
  for _ in $(seq 1 90); do
    code=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}/ready" || true)
    [[ "$code" == "200" ]] && { echo "  стенд готов"; break; }
    sleep 2
  done
fi

echo "HTTP-проверки (${BASE_URL}):"

# --- Базовые маршруты WebLab #4 ---
check "корень /"                        "200"       "/"
check "копия /reserved/"                "200"       "/reserved/"
check "статический css"                 "200"       "/static/style.css"
check "статическое изображение"         "200"       "/static/img/dashboard.png"
check "API v2 список"                   "200"       "/api/v2/subscription-plans"
check "API v2 (скачивание спеки)"       "200"       "/api/v2"
check "API v2 404 (нет сущности)"       "404"       "/api/v2/subscription-plans/999999"
check "API v1 список"                   "200"       "/api/v1/subscription-plans"
check "спецификация openapi"            "200"       "/api/v2/openapi.yaml"
check "спецификация /openapi"           "200"       "/openapi.yaml"
check "Swagger UI редирект"             "301"       "/api/v2/docs"
check "Swagger UI"                      "200"       "/api/v2/docs/"
check "документация"                    "200"       "/documentation"
check "legacy"                          "200"       "/legacy"
check "legacy download (бинарник)"      "200"       "/legacy/download"
check "status"                          "200"       "/status"
check "health (liveness)"               "200"       "/health"
check "ready (readiness)"               "200"       "/ready"
check "management"                      "200"       "/management"
check "admin без авторизации"           "401"       "/admin/"
check "admin с авторизацией"            "200 302"   "/admin/" -u "admin:secret"

# --- WebLab #5: запись на основной инстанс (POST) ---
check "запись POST /api/v2/users"       "201"       "/api/v2/users" -X POST -H "Content-Type: application/json" -d "{\"email\":\"check5-${RANDOM}@example.com\",\"password\":\"secret\",\"user_name\":\"check5\",\"age\":20,\"balance\":10}"

# --- WebLab #5: /mirror (отдельная версия приложения за репликой) ---
check "/mirror редирект"                "301"       "/mirror"
check "/mirror страница"                "200"       "/mirror/"
check "/mirror API чтение"              "200"       "/mirror/api/v2/subscription-plans"
check "/mirror API v1 чтение"           "200"       "/mirror/api/v1/subscription-plans"
check "/mirror Swagger UI редирект"     "301"       "/mirror/api/v2/docs"
check "/mirror Swagger UI"              "200"       "/mirror/api/v2/docs/"
check "/mirror health"                  "200"       "/mirror/health"
check "/mirror ready"                   "200"       "/mirror/ready"
check "/mirror запись (POST)"           "201"       "/mirror/api/v2/users" -X POST -H "Content-Type: application/json" -d "{\"email\":\"check5m-${RANDOM}@example.com\",\"password\":\"secret\",\"user_name\":\"check5m\",\"age\":20,\"balance\":10}"

# --- WebLab #5: /monitoring (Grafana) ---
check "/monitoring редирект"            "301"       "/monitoring"
check "/monitoring (Grafana)"           "200 302"   "/monitoring/"

# --- WebLab #5: подмена имени сервера (заголовок Server) ---
check_header "Server = SubscriberInf"   "Server"    "SubscriberInf" "/"

# --- WebLab #5: кэширование публичной статики ---
check_header "статический css кэш public" "Cache-Control" "public" "/static/style.css"

# --- WebLab #5: исключения из кэша (API/служебные страницы) ---
check_header "API без кэша"             "Cache-Control" "no-store" "/api/v2/subscription-plans"
check_header "status без кэша"          "Cache-Control" "no-store" "/status"
check_header "management без кэша"      "Cache-Control" "no-store" "/management"
check_header "health без кэша"          "Cache-Control" "no-store" "/health"

# --- WebLab #5: gzip сжатие текстовых ответов ---
check_header "gzip статического css"    "Content-Encoding" "gzip" "/static/style.css" -H "Accept-Encoding: gzip"

echo
echo "Результат: PASS=${pass} FAIL=${fail}"
[[ $fail -eq 0 ]]
