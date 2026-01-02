-- Хранимая процедура без параметров или с параметрами
-- Задача: Процедура, которая обновляет баланс всех пользователей, добавляя к нему 5% бонуса, но не более 100 единиц.
CREATE OR REPLACE PROCEDURE update_balances_with_bonus()
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE users
    SET balance = balance + LEAST(balance * 0.05, 100);
    COMMIT;
END;
$$;
