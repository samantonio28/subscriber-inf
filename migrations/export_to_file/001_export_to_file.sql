-- копирование в csv
COPY services TO '/tmp/services.csv' WITH CSV HEADER;
COPY users TO '/tmp/users.csv' WITH CSV HEADER;
COPY subscriptions TO '/tmp/subscriptions.csv' WITH CSV HEADER;
COPY users_subs TO '/tmp/users_subs.csv' WITH CSV HEADER;
COPY sub_durations TO '/tmp/sub_durations.csv' WITH CSV HEADER;
COPY promocodes TO '/tmp/promocodes.csv' WITH CSV HEADER;
COPY cards TO '/tmp/cards.csv' WITH CSV HEADER;
COPY payments TO '/tmp/payments.csv' WITH CSV HEADER;
