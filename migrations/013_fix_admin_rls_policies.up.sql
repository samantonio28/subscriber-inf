BEGIN;

-- Добавляем политики RLS для роли app_admin (администратор)
-- чтобы он мог видеть все строки в таблицах с включенным RLS

-- Политика для таблицы users: администратор может выполнять любые операции
CREATE POLICY admin_all_users ON users FOR ALL TO app_admin
    USING (true)
    WITH CHECK (true);

-- Политика для таблицы subscriptions
CREATE POLICY admin_all_subscriptions ON subscriptions FOR ALL TO app_admin
    USING (true)
    WITH CHECK (true);

-- Политика для таблицы payments
CREATE POLICY admin_all_payments ON payments FOR ALL TO app_admin
    USING (true)
    WITH CHECK (true);

-- Политика для таблицы cards
CREATE POLICY admin_all_cards ON cards FOR ALL TO app_admin
    USING (true)
    WITH CHECK (true);

-- Примечание: политики для аналитика не нужны, так как он не имеет прав на эти таблицы
-- (GRANT не выдан). Если аналитику потребуется доступ к своим данным, можно добавить политики позже.

COMMIT;