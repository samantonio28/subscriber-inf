BEGIN;

-- 1. Включаем Row Level Security (RLS) для таблиц, где нужен контроль доступа на уровне строк
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE cards ENABLE ROW LEVEL SECURITY;

-- 2. Создаем политики для роли app_user (обычный пользователь)
-- Пользователь может видеть только свои данные в таблице users
CREATE POLICY user_select_own ON users FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может видеть только свои подписки
CREATE POLICY subscription_select_own ON subscriptions FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может создавать только свои подписки
CREATE POLICY subscription_insert_own ON subscriptions FOR INSERT TO app_user
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может удалять только свои подписки
CREATE POLICY subscription_delete_own ON subscriptions FOR DELETE TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может видеть только свои платежи
CREATE POLICY payment_select_own ON payments FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может создавать только свои платежи
CREATE POLICY payment_insert_own ON payments FOR INSERT TO app_user
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может видеть только свои карты
CREATE POLICY card_select_own ON cards FOR SELECT TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID);

-- Пользователь может управлять только своими картами
CREATE POLICY card_all_own ON cards FOR ALL TO app_user
    USING (user_id = current_setting('app.current_user_id', TRUE)::UUID)
    WITH CHECK (user_id = current_setting('app.current_user_id', TRUE)::UUID);

COMMIT;