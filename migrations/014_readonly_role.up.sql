BEGIN;

-- WebLab #5 — роли для балансировки чтения и стриминговой репликации.

-- 1. Роль только для чтения (используется read-only инстансами бэкенда).
--    Подключается к primary, может только SELECT.
CREATE ROLE reader WITH LOGIN PASSWORD 'readpass';
GRANT CONNECT ON DATABASE dev TO reader;
GRANT USAGE ON SCHEMA public TO reader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO reader;

-- 2. RLS включён на users/subscriptions/payments/cards (миграции 010/011).
--    Роль reader — не владелец и не суперпользователь, поэтому без политик
--    она видела бы 0 строк. Добавляем SELECT-политики на все строки:
--    read-only роль на то и read-only, что читать можно все строки.
CREATE POLICY reader_select_users ON users FOR SELECT TO reader USING (true);
CREATE POLICY reader_select_subscriptions ON subscriptions FOR SELECT TO reader USING (true);
CREATE POLICY reader_select_payments ON payments FOR SELECT TO reader USING (true);
CREATE POLICY reader_select_cards ON cards FOR SELECT TO reader USING (true);

-- 3. Роль для стриминговой репликации (pg_basebackup + wal sender).
CREATE ROLE replicator WITH LOGIN REPLICATION PASSWORD 'secret';

COMMIT;
