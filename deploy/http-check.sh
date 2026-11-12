#!/usr/bin/env bash
# HTTP-проверки стенда WebLab #4 (nginx — единая точка входа).
# Запуск: bash deploy/http-check.sh [BASE_URL] [--no-wait]
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
  local ok=0
  local e
  for e in $expected; do
    [[ "$code" == "$e" ]] && ok=1
  done
  if [[ $ok -eq 1 ]]; then
    printf '  PASS  %-34s -> %s\n' "$name" "$code"
    pass=$((pass+1))
  else
    printf '  FAIL  %-34s -> ожидалось [%s], получено %s\n' "$name" "$expected" "$code"
    fail=$((fail+1))
  fi
}

if [[ $WAIT -eq 1 ]]; then
  echo "Ожидание готовности ${BASE_URL}/ready ..."
  for _ in $(seq 1 60); do
    code=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}/ready" || true)
    [[ "$code" == "200" ]] && { echo "  стенд готов"; break; }
    sleep 2
  done
fi

echo "HTTP-проверки (${BASE_URL}):"

check "корень /"                 "200"     "/"
check "копия /reserved/"         "200"     "/reserved/"
check "статический css"          "200"     "/static/style.css"
check "статическое изображение"  "200"     "/static/img/dashboard.png"
check "API v2 список"            "200"     "/api/v2/subscription-plans"
check "API v2 (скачивание спеки)" "200"    "/api/v2"
check "API v2 404 (нет сущности)" "404"    "/api/v2/subscription-plans/999999"
check "API v1 список"            "200"     "/api/v1/subscription-plans"
check "спецификация openapi"     "200"     "/api/v2/openapi.yaml"
check "спецификация /openapi"    "200"     "/openapi.yaml"
check "Swagger UI редирект"      "301"     "/api/v2/docs"
check "Swagger UI"               "200"     "/api/v2/docs/"
check "документация"             "200"     "/documentation"
check "legacy"                   "200"     "/legacy"
check "legacy download (бинарник)" "200"    "/legacy/download"
check "status"                   "200"     "/status"
check "health (liveness)"        "200"     "/health"
check "ready (readiness)"        "200"     "/ready"
check "management"               "200"     "/management"
check "admin без авторизации"    "401"     "/admin/"
check "admin с авторизацией"     "200 302" "/admin/" -u "admin:secret"

echo
echo "Результат: PASS=${pass} FAIL=${fail}"
[[ $fail -eq 0 ]]
