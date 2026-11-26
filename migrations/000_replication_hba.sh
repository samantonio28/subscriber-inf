#!/bin/sh
# WebLab #5 — правило стриминговой репликации в pg_hba.conf (primary-replica).
# Выполняется при инициализации primary (exec-бит установлен).
set -e

if [ -n "$PGDATA" ] && [ -f "$PGDATA/pg_hba.conf" ] && ! grep -q "host replication replicator" "$PGDATA/pg_hba.conf"; then
    echo "host replication replicator all scram-sha-256" >> "$PGDATA/pg_hba.conf"
fi
