--  Рекурсивная функция или функция с рекурсивным ОТВ
-- Задача: Функция, которая с помощью рекурсивного ОТВ находит всех пользователей, приглашенных данным пользователем (и их приглашенных, и так далее), и уровень вложенности.
CREATE OR REPLACE FUNCTION get_referral_tree(start_referrer_id UUID)
RETURNS TABLE (
    referred_id UUID,
    referrer_id UUID,
    level INTEGER
)
LANGUAGE sql
AS $$
    WITH RECURSIVE referral_tree AS (
        -- Якорь рекурсии: непосредственные рефералы
        SELECT
            ur.referred_id,
            ur.referrer_id,
            1 AS level
        FROM user_referrals ur
        WHERE ur.referrer_id = start_referrer_id

        UNION ALL

        -- Рекурсивный шаг: рефералы рефералов
        SELECT
            ur.referred_id,
            ur.referrer_id,
            rt.level + 1
        FROM user_referrals ur
        INNER JOIN referral_tree rt ON ur.referrer_id = rt.referred_id
    )
    SELECT * FROM referral_tree;
$$;
