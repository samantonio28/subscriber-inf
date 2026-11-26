#!/bin/sh
# WebLab #5 — entrypoint standby PostgreSQL (primary-replica, стриминговая репликация).
# Первый запуск: pg_basebackup с primary + standby.signal + primary_conninfo.
# Последующие запуски: обычный старт standby (данные уже есть).
#
# Примечание: вызываем `docker-entrypoint.sh postgres` явно (без опоры на CMD),
# т.к. при override entrypoint в compose CMD образа не наследуется.
set -e

PGDATA="${PGDATA:-/var/lib/postgresql/data}"
PRIMARY_HOST="${PRIMARY_HOST:-postgres}"
REPLICATION_PASSWORD="${REPLICATION_PASSWORD:-secret}"

if [ ! -s "$PGDATA/PG_VERSION" ]; then
    echo "[replica] Жду primary ${PRIMARY_HOST} (TCP + роль replicator)..."
    # pg_isready -h — только по TCP; временный init-сервер primary слушает только
    # unix-сокет, поэтому цикл завершится лишь после старта основного сервера.
    until pg_isready -h "$PRIMARY_HOST" -U postgres >/dev/null 2>&1 && \
          PGPASSWORD="$REPLICATION_PASSWORD" psql -h "$PRIMARY_HOST" -U postgres -d dev -tAc \
            "SELECT 1 FROM pg_roles WHERE rolname = 'replicator'" 2>/dev/null | grep -q 1; do
        sleep 2
    done

    echo "[replica] Базовая резервная копия с primary (pg_basebackup)..."
    mkdir -p "$PGDATA"
    chown -R postgres:postgres "$PGDATA" 2>/dev/null || true
    gosu postgres sh -c "PGPASSWORD='${REPLICATION_PASSWORD}' pg_basebackup -h '${PRIMARY_HOST}' -U replicator -D '${PGDATA}' -X stream -P -v"

    echo "[replica] Создаю standby.signal..."
    touch "$PGDATA/standby.signal"

    # Важно: ищем НЕзакомментированную primary_conninfo (в дефолтном
    # postgresql.conf есть строка "#primary_conninfo = ''", её не считаем).
    if ! grep -q '^primary_conninfo' "$PGDATA/postgresql.conf" 2>/dev/null; then
        cat >> "$PGDATA/postgresql.conf" <<EOF

# --- streaming replication (primary-replica) ---
primary_conninfo = 'host=${PRIMARY_HOST} port=5432 user=replicator password=${REPLICATION_PASSWORD}'
hot_standby = on
EOF
    fi
    chown -R postgres:postgres "$PGDATA"
fi

# Аргументы берём из CMD compose (например: postgres -c max_connections=200);
# параметры standby должны быть не меньше, чем у primary.
exec docker-entrypoint.sh "$@"
