-- Многооператорная табличная функция (Multi-Statement Table-Valued Function)
-- Задача: Функция, которая возвращает детализированный финансовый отчет по пользователю: его баланс, общую сумму пополнений и общую сумму списаний.
CREATE OR REPLACE FUNCTION get_user_financial_report(user_id_input UUID)
RETURNS TABLE (
    current_balance INTEGER,
    total_income BIGINT,
    total_expence BIGINT
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Получаем текущий баланс пользователя
    SELECT balance INTO current_balance
    FROM users
    WHERE user_id = user_id_input;

    -- Считаем общие пополнения (income)
    SELECT COALESCE(SUM(amount), 0) INTO total_income
    FROM payments
    WHERE user_id = user_id_input AND paym_type = 'income';

    -- Считаем общие списания (expence)
    SELECT COALESCE(SUM(amount), 0) INTO total_expence
    FROM payments
    WHERE user_id = user_id_input AND paym_type = 'expence';

    RETURN NEXT;
    RETURN;
END;
$$;
