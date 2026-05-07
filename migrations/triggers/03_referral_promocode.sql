CREATE OR REPLACE FUNCTION add_referral_promocodes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    -- Начисляем промокод пригласившему (referrer_id)
    INSERT INTO promocodes (service_id, promocode, sub_duration_days, expires_at, discount, max_uses, cur_uses, status, duration_days)
    VALUES (
        1, -- service_id по умолчанию (нужно уточнить)
        'REF' || NEW.referrer_id::text, -- пример промокода
        30, -- длительность подписки в днях
        CURRENT_DATE + INTERVAL '30 days', -- срок действия
        10, -- discount 10%
        1, -- max_uses
        0, -- cur_uses
        'ACTIVE', -- status
        30 -- duration_days
    );

    -- Начисляем промокод приглашенному (referred_id)
    INSERT INTO promocodes (service_id, promocode, sub_duration_days, expires_at, discount, max_uses, cur_uses, status, duration_days)
    VALUES (
        1,
        'REF' || NEW.referred_id::text,
        30,
        CURRENT_DATE + INTERVAL '30 days',
        10,
        1,
        0,
        'ACTIVE',
        30
    );

    RETURN NEW;
END;
$$;

-- Создаем триггер после вставки в user_referrals
CREATE OR REPLACE TRIGGER trg_add_referral_promocodes
    AFTER INSERT ON user_referrals
    FOR EACH ROW
    EXECUTE FUNCTION add_referral_promocodes();