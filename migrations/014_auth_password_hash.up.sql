BEGIN;

-- ЛР №5: пароли теперь хранятся как bcrypt-хэши (60 символов),
-- колонка password VARCHAR(30) не вмещает хэш — расширяем.
ALTER TABLE users ALTER COLUMN password TYPE VARCHAR(72);

COMMIT;
