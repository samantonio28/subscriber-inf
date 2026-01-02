-- Триггер INSTEAD OF
-- Задача: Триггер INSTEAD OF для представления, которое позволяет "безопасно" вставлять данные в payments, автоматически проверяя и корректируя баланс пользователя.
-- Сначала создадим представление для таблицы payments
CREATE OR REPLACE VIEW payments_view AS
SELECT paym_id, user_id, card_number, amount, paym_type
FROM payments;

-- Создаем функцию для триггера INSTEAD OF INSERT
CREATE OR REPLACE FUNCTION instead_of_insert_payment()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Вставляем запись в реальную таблицу
    INSERT INTO payments (user_id, card_number, amount, paym_type)
    VALUES (NEW.user_id, NEW.card_number, NEW.amount, NEW.paym_type);

    -- Обновляем баланс пользователя в зависимости от типа операции
    IF NEW.paym_type = 'income' THEN
        UPDATE users SET balance = balance + NEW.amount WHERE user_id = NEW.user_id;
    ELSIF NEW.paym_type = 'expence' THEN
        -- Проверяем, достаточно ли средств
        IF (SELECT balance FROM users WHERE user_id = NEW.user_id) < NEW.amount THEN
            RAISE EXCEPTION 'Недостаточно средств на балансе для списания';
        END IF;
        UPDATE users SET balance = balance - NEW.amount WHERE user_id = NEW.user_id;
    END IF;

    -- Возвращаем NEW, чтобы триггер завершился успешно
    RETURN NEW;
END;
$$;

-- Создаем сам триггер INSTEAD OF
CREATE OR REPLACE TRIGGER trg_instead_of_insert_payment
    INSTEAD OF INSERT ON payments_view
    FOR EACH ROW
    EXECUTE FUNCTION instead_of_insert_payment();
