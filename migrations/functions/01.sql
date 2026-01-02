-- Скалярная функция
-- Задача: Функция, которая вычисляет общую сумму доходов от подписок для указанного сервиса.
CREATE OR REPLACE FUNCTION get_total_income_for_service(service_id_input INTEGER)
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    total_income INTEGER;
BEGIN
    SELECT COALESCE(SUM(s.price), 0)
    INTO total_income
    FROM subscriptions s
    WHERE s.service_id = service_id_input
      AND s.sub_type != 'promocode'; -- Не считаем промокоды, так как они бесплатные

    RETURN total_income;
END;
$$;
