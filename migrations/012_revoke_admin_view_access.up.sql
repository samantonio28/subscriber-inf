BEGIN;

-- Отзываем у администратора права SELECT на VIEW аналитика
-- (администратор сохраняет все права на таблицы, но не на эти VIEW)
REVOKE SELECT ON user_statistics FROM app_admin;
REVOKE SELECT ON referral_statistics FROM app_admin;

-- Примечание: если администратору нужен доступ к статистике, можно оставить,
-- но по требованиям доступ только у аналитика.

COMMIT;