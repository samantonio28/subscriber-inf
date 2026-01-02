-- Подставляемая табличная функция (Inline Table-Valued Function)
-- Задача: Функция, которая возвращает всех активных (не истекших на текущую дату) подписчиков указанного сервиса.
CREATE OR REPLACE FUNCTION get_active_subscribers(service_id_input INTEGER)
RETURNS TABLE (
    user_id UUID,
    user_name VARCHAR(20),
    email VARCHAR,
    sub_type subscription_type,
    end_date DATE
)
LANGUAGE sql
AS $$
    SELECT u.user_id, u.user_name, u.email, s.sub_type, s.end_date
    FROM subscriptions s
    JOIN users u ON s.user_id = u.user_id
    WHERE s.service_id = service_id_input
      AND s.end_date >= CURRENT_DATE;
$$;
