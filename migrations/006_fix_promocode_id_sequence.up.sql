BEGIN;

-- Изменяем promocode_id на автоинкрементный
ALTER TABLE promocodes 
ALTER COLUMN promocode_id ADD GENERATED ALWAYS AS IDENTITY;

COMMIT;