-- получение из csv
-- очищаем таблицы, чтоб вставить новые данные
TRUNCATE TABLE payments, cards, promocodes, users_subs, sub_durations, subscriptions, users, services CASCADE;

COPY services FROM '/tmp/services.csv' WITH CSV HEADER;
COPY users FROM '/tmp/users.csv' WITH CSV HEADER;
COPY subscriptions FROM '/tmp/subscriptions.csv' WITH CSV HEADER;
COPY users_subs FROM '/tmp/users_subs.csv' WITH CSV HEADER;
COPY sub_durations FROM '/tmp/sub_durations.csv' WITH CSV HEADER;
COPY promocodes FROM '/tmp/promocodes.csv' WITH CSV HEADER;
COPY cards FROM '/tmp/cards.csv' WITH CSV HEADER;
COPY payments FROM '/tmp/payments.csv' WITH CSV HEADER;

-- настраиваем последовательности (есть несколько generated as identity, сразу их правильно ставим согласно данным из csv)
SELECT setval(pg_get_serial_sequence('services', 'service_id'), coalesce(max(service_id), 0) + 1, false) FROM services;
SELECT setval(pg_get_serial_sequence('subscriptions', 'sub_id'), coalesce(max(sub_id), 0) + 1, false) FROM subscriptions;
