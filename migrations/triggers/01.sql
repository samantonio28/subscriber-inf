-- Триггер AFTER
-- Задача: Триггер, который после вставки новой подписки типа 'family' проверяет, не превышает ли количество активных семейных подписок на этом сервисе лимит (users_count из таблицы services). Если превышает - отменяет вставку.
-- Сначала создадим функцию, которую будет вызывать триггер
CREATE OR REPLACE FUNCTION check_family_sub_limit()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    max_users INTEGER;
    current_users INTEGER;
BEGIN
    -- Получаем максимальное количество пользователей для этого сервиса
    SELECT users_count INTO max_users
    FROM services
    WHERE service_id = NEW.service_id;

    -- Считаем текущее количество активных подписок типа 'family' для этого сервиса
    SELECT COUNT(*) INTO current_users
    FROM subscriptions
    WHERE service_id = NEW.service_id
      AND sub_type = 'family'
      AND end_date >= CURRENT_DATE;

    -- Если лимит превышен, отменяем вставку
    IF current_users >= max_users THEN
        RAISE EXCEPTION 'Лимит активных семейных подписок для сервиса % достигнут. Максимум: %', NEW.service_id, max_users;
    END IF;

    RETURN NEW;
END;
$$;

-- Создаем сам триггер
CREATE OR REPLACE TRIGGER trg_check_family_sub_limit
    AFTER INSERT ON subscriptions
    FOR EACH ROW
    WHEN (NEW.sub_type = 'family')
    EXECUTE FUNCTION check_family_sub_limit();
