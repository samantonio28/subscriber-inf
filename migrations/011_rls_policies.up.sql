BEGIN;

-- 1. Включаем Row Level Security (RLS) для таблиц, где нужен контроль доступа на уровне строк.
--    (Идемпотентно: ENABLE на уже включённом RLS не ошибка.)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE cards ENABLE ROW LEVEL SECURITY;

-- 2. Политики для роли app_user (обычный пользователь).
--    Миграция 010 уже создаёт эти политики; здесь добавляем DROP IF EXISTS,
--    чтобы повторный проход не падал с «policy already exists» и не прерывал
--    последующие миграции (012, 013, 014).

DROP POLICY IF EXISTS user_select_own ON users;
CREATE POLICY user_select_own ON users FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS subscription_select_own ON subscriptions;
CREATE POLICY subscription_select_own ON subscriptions FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS subscription_insert_own ON subscriptions;
CREATE POLICY subscription_insert_own ON subscriptions FOR INSERT TO app_user
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS subscription_delete_own ON subscriptions;
CREATE POLICY subscription_delete_own ON subscriptions FOR DELETE TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS payment_select_own ON payments;
CREATE POLICY payment_select_own ON payments FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS payment_insert_own ON payments;
CREATE POLICY payment_insert_own ON payments FOR INSERT TO app_user
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS card_select_own ON cards;
CREATE POLICY card_select_own ON cards FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

DROP POLICY IF EXISTS card_all_own ON cards;
CREATE POLICY card_all_own ON cards FOR ALL TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID)
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

COMMIT;
