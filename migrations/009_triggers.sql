BEGIN;

CREATE OR REPLACE FUNCTION add_referral_promocodes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    referrer_service_id INTEGER;
    referred_service_id INTEGER;
    referrer_promocode VARCHAR(10);
    referred_promocode VARCHAR(10);
    referrer_discount INTEGER;
    referred_discount INTEGER;
BEGIN
    -- Выбираем случайный service_id
    SELECT service_id INTO referrer_service_id
    FROM services
    ORDER BY RANDOM()
    LIMIT 1;
    
    SELECT service_id INTO referred_service_id
    FROM services
    ORDER BY RANDOM()
    LIMIT 1;

    -- Генерируем уникальные промокоды
    referrer_promocode := 'REF' || substr(md5(random()::text || NEW.referrer_id::text || clock_timestamp()::text), 1, 7);
    referred_promocode := 'REF' || substr(md5(random()::text || NEW.referred_id::text || clock_timestamp()::text), 1, 7);

    -- Случайная скидка от 65% до 90%
    referrer_discount := floor(random() * 26) + 65;
    referred_discount := floor(random() * 26) + 65;

    -- Вставка промокодов
    INSERT INTO promocodes (
        service_id,
        promocode,
        sub_id,
        created_at,
        discount,
        max_uses,
        cur_uses,
        status,
        duration_days,
        plan_id,
        expires_at
    ) VALUES 
    (
        referrer_service_id,
        referrer_promocode,
        NULL,
        CURRENT_TIMESTAMP,
        referrer_discount,
        1,
        0,
        'ACTIVE',
        3,
        NULL,
        CURRENT_TIMESTAMP + INTERVAL '30 days'
    ),
    (
        referred_service_id,
        referred_promocode,
        NULL,
        CURRENT_TIMESTAMP,
        referred_discount,
        1,
        0,
        'ACTIVE',
        3,
        NULL,
        CURRENT_TIMESTAMP + INTERVAL '30 days'
    );

    RETURN NEW;
END;
$$;

-- Пересоздаем триггер
DROP TRIGGER IF EXISTS trg_add_referral_promocodes ON user_referrals;

CREATE TRIGGER trg_add_referral_promocodes
    AFTER INSERT ON user_referrals
    FOR EACH ROW
    EXECUTE FUNCTION add_referral_promocodes();

COMMIT;