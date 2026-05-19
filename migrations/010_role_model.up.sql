BEGIN;

-- 1. Создание ENUM типа для ролей пользователей (для хранения в таблице users)
CREATE TYPE user_role AS ENUM ('user', 'admin', 'analyst');

-- 2. Добавление столбца role в таблицу users
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'user';

-- 3. Создание ролей PostgreSQL
CREATE ROLE app_user;
CREATE ROLE app_admin;
CREATE ROLE app_analyst;

-- 4. Создание VIEW для аналитика

-- 4.1. user_statistics - статистика по пользователям
CREATE VIEW user_statistics AS
SELECT 
    -- Распределение по возрастам
    CASE 
        WHEN age BETWEEN 18 AND 25 THEN '18-25'
        WHEN age BETWEEN 26 AND 35 THEN '26-35'
        WHEN age BETWEEN 36 AND 45 THEN '36-45'
        WHEN age > 45 THEN '45+'
        ELSE 'unknown'
    END as age_group,
    COUNT(*) as user_count,
    -- Траты на подписки
    COALESCE(SUM(subscriptions.price), 0) as total_spent,
    -- Соотношение использования промокодов
    COUNT(DISTINCT CASE WHEN promocodes.promocode_id IS NOT NULL THEN users.user_id END) as users_with_promocodes,
    COUNT(DISTINCT promocodes.promocode_id) as total_promocodes_used
FROM users
LEFT JOIN subscriptions ON users.user_id = subscriptions.user_id
LEFT JOIN promocodes ON subscriptions.promocode_id = promocodes.promocode_id
GROUP BY age_group
ORDER BY age_group;

-- 4.2. referral_statistics - статистика по реферальной программе
CREATE VIEW referral_statistics AS
SELECT 
    referrer.user_id as referrer_id,
    referrer.user_name as referrer_name,
    COUNT(referred.user_id) as referred_count,
    -- Конверсия в покупки
    COUNT(DISTINCT CASE WHEN subscriptions.sub_id IS NOT NULL THEN referred.user_id END) as converted_to_purchase,
    -- Среднее число приглашений
    AVG(sub_count.subscription_count) as avg_subscriptions_per_referred
FROM user_referrals
JOIN users referrer ON user_referrals.referrer_id = referrer.user_id
JOIN users referred ON user_referrals.referred_id = referred.user_id
LEFT JOIN (
    SELECT user_id, COUNT(*) as subscription_count
    FROM subscriptions
    GROUP BY user_id
) sub_count ON referred.user_id = sub_count.user_id
LEFT JOIN subscriptions ON referred.user_id = subscriptions.user_id
GROUP BY referrer.user_id, referrer.user_name
ORDER BY referred_count DESC;

-- 5. Назначение прав доступа для ролей PostgreSQL

-- 5.1. Права для app_user (обычный пользователь)
-- SELECT на свои данные (реализуется через RLS или проверку в приложении)
-- Для простоты даем SELECT на все таблицы, но ограничения будут на уровне приложения
GRANT SELECT ON users, services, subscription_plans, promocodes, cards, payments, user_referrals TO app_user;
GRANT INSERT, UPDATE, DELETE ON cards TO app_user;
GRANT INSERT ON subscriptions, payments TO app_user;
GRANT DELETE ON subscriptions TO app_user;

-- 5.2. Права для app_admin (администратор)
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO app_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO app_admin;

-- 5.3. Права для app_analyst (аналитик)
GRANT SELECT ON user_statistics, referral_statistics TO app_analyst;

-- 6. Обновление существующих пользователей
-- Первому пользователю назначаем роль 'admin' для тестирования
-- UPDATE users 
-- SET role = 'admin' 
-- WHERE user_id = (SELECT user_id FROM users ORDER BY created_at LIMIT 1);

-- Второму пользователю назначаем роль 'analyst' (если есть)
-- UPDATE users 
-- SET role = 'analyst' 
-- WHERE user_id = (SELECT user_id FROM users ORDER BY created_at OFFSET 1 LIMIT 1);

-- 7. Создание функции для переключения роли в зависимости от user_id
-- Эта функция будет вызываться приложением после аутентификации
CREATE OR REPLACE FUNCTION set_role_by_user_id(p_user_id UUID) RETURNS VOID AS $$
DECLARE
    v_role user_role;
BEGIN
    -- Получаем роль пользователя из таблицы users
    SELECT role INTO v_role
    FROM users
    WHERE user_id = p_user_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User with id % not found', p_user_id;
    END IF;
    
    -- Устанавливаем соответствующую роль PostgreSQL
    CASE v_role
        WHEN 'admin' THEN
            EXECUTE 'SET ROLE app_admin';
        WHEN 'analyst' THEN
            EXECUTE 'SET ROLE app_analyst';
        ELSE
            EXECUTE 'SET ROLE app_user';
    END CASE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 8. Создание пользователя по умолчанию для подключения приложения
-- Приложение будет подключаться от имени этого пользователя, затем вызывать set_role_by_user_id
-- Создаем пользователя 'app' (если еще не существует)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'app') THEN
        CREATE ROLE app WITH LOGIN PASSWORD 'app_password';
    END IF;
END
$$;

-- Даем пользователю app право на подключение и выполнение функций
GRANT CONNECT ON DATABASE dev TO app;
GRANT USAGE ON SCHEMA public TO app;
GRANT EXECUTE ON FUNCTION set_role_by_user_id(UUID) TO app;

-- 9. Включаем Row Level Security (RLS) для таблиц, где нужен контроль доступа на уровне строк
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE cards ENABLE ROW LEVEL SECURITY;

-- 10. Создаем политики для роли app_user (обычный пользователь)
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

-- 11. Для администратора и аналитика политики не нужны, так как они имеют права на все строки
-- через GRANT на уровне таблиц

COMMIT;