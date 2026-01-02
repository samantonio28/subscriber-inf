-- Рекурсивная хранимая процедура или хранимая процедура с рекурсивным ОТВ
-- Задача: Процедура, которая использует рекурсивное ОТВ для вывода иерархии рефералов заданного пользователя в виде читаемого дерева.
CREATE OR REPLACE PROCEDURE print_referral_hierarchy(IN start_user_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN (
        WITH RECURSIVE referral_hierarchy AS (
            SELECT
                referred_id,
                referrer_id,
                user_name,
                0 AS level
            FROM user_referrals ur
            JOIN users u ON ur.referred_id = u.user_id
            WHERE referrer_id = start_user_id

            UNION ALL

            SELECT
                ur.referred_id,
                ur.referrer_id,
                u.user_name,
                rh.level + 1
            FROM user_referrals ur
            JOIN referral_hierarchy rh ON ur.referrer_id = rh.referred_id
            JOIN users u ON ur.referred_id = u.user_id
        )
        SELECT
            rh.referred_id,
            rh.user_name,
            rh.level,
            REPEAT('  ', rh.level) || '-> ' || rh.user_name AS hierarchy_tree
        FROM referral_hierarchy rh
        ORDER BY level, user_name
    )
    LOOP
        RAISE NOTICE '%', rec.hierarchy_tree;
    END LOOP;
END;
$$;
