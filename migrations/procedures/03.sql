-- Хранимая процедура с курсором
-- Задача: Процедура, которая с помощью курсора проходит по всем подпискам типа 'promocode', у которых истек срок действия, и деактивирует их (устанавливает end_date в CURRENT_DATE).
CREATE OR REPLACE PROCEDURE deactivate_expired_promocode_subs()
LANGUAGE plpgsql
AS $$
DECLARE
    promo_cursor CURSOR FOR
        SELECT s.sub_id, p.expires_at
        FROM subscriptions s
        JOIN promocodes p ON s.sub_id = p.sub_id
        WHERE s.sub_type = 'promocode'
          AND p.expires_at < CURRENT_DATE
          AND s.end_date >= CURRENT_DATE; -- Деактивируем только активные

    sub_id_record INTEGER;
    expires_at_record DATE;
BEGIN
    OPEN promo_cursor;
    LOOP
        FETCH promo_cursor INTO sub_id_record, expires_at_record;
        EXIT WHEN NOT FOUND;

        RAISE NOTICE 'Деактивируем подписку с promocode ID: %, дата истечения: %', sub_id_record, expires_at_record;

        UPDATE subscriptions
        SET end_date = CURRENT_DATE
        WHERE sub_id = sub_id_record;

    END LOOP;
    CLOSE promo_cursor;
    COMMIT;
END;
$$;