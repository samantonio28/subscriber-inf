BEGIN;

-- Добавляем столбец referral_code в таблицу users
ALTER TABLE users ADD COLUMN referral_code VARCHAR(20);

COMMIT;