BEGIN;

-- services
ALTER TABLE services ADD PRIMARY KEY (service_id);
ALTER TABLE services ADD CONSTRAINT unique_service_name UNIQUE (service_name);
ALTER TABLE services ADD CONSTRAINT nn_service_name CHECK (service_name IS NOT NULL);
ALTER TABLE services ADD CONSTRAINT nn_sub_duration_id_default CHECK (sub_duration_id_default IS NOT NULL);
ALTER TABLE services ADD CONSTRAINT nn_users_count CHECK (users_count IS NOT NULL);
ALTER TABLE services ALTER COLUMN has_promocodes SET DEFAULT FALSE;
ALTER TABLE services ALTER COLUMN sub_duration_id_default SET DEFAULT 1;
ALTER TABLE services ALTER COLUMN users_count SET DEFAULT 1;

-- users
ALTER TABLE users ADD PRIMARY KEY (user_id);
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT check_age_adult 
    CHECK (age >= 18);
ALTER TABLE users ADD CONSTRAINT check_balance_non_negative 
    CHECK (balance >= 0);
ALTER TABLE users ADD CONSTRAINT nn_user_id CHECK (user_id IS NOT NULL);
ALTER TABLE users ADD CONSTRAINT nn_email CHECK (email IS NOT NULL);
ALTER TABLE users ADD CONSTRAINT nn_password CHECK (password IS NOT NULL);
ALTER TABLE users ADD CONSTRAINT nn_age CHECK (age IS NOT NULL);
ALTER TABLE users ADD CONSTRAINT nn_balance CHECK (balance IS NOT NULL);
ALTER TABLE users ADD CONSTRAINT nn_user_name CHECK (user_name IS NOT NULL);

-- subscriptions
ALTER TABLE subscriptions ADD PRIMARY KEY (sub_id);
ALTER TABLE subscriptions ADD CONSTRAINT fk_subscriptions_service 
    FOREIGN KEY (service_id) REFERENCES services(service_id);
ALTER TABLE subscriptions ADD CONSTRAINT fk_subscriptions_users
    FOREIGN KEY (user_id) REFERENCES users(user_id);
ALTER TABLE subscriptions ADD CONSTRAINT nn_service_id CHECK (service_id IS NOT NULL);
ALTER TABLE subscriptions ADD CONSTRAINT nn_price CHECK (price IS NOT NULL);
ALTER TABLE subscriptions ADD CONSTRAINT check_price_positive 
    CHECK (price > 0);
ALTER TABLE subscriptions ADD CONSTRAINT valid_start_date 
    CHECK (
        (sub_type = 'promocode' AND start_date IS NULL) OR
        EXTRACT(DAY FROM start_date) = 1
    );
ALTER TABLE subscriptions ADD CONSTRAINT valid_end_date 
    CHECK (
        (end_date IS NULL AND start_date IS NULL) OR 
        (EXTRACT(DAY FROM end_date) = 1 AND end_date >= start_date)
    );
ALTER TABLE subscriptions ADD CONSTRAINT nn_sub_type CHECK (sub_type IS NOT NULL);
ALTER TABLE subscriptions ALTER COLUMN sub_type SET DEFAULT 'usual';

-- cards
ALTER TABLE cards ADD PRIMARY KEY (card_number);
ALTER TABLE cards ADD CONSTRAINT unique_card_number UNIQUE (card_number);
ALTER TABLE cards ADD CONSTRAINT fk_cards_user 
    FOREIGN KEY (user_id) REFERENCES users(user_id);
ALTER TABLE cards ADD CONSTRAINT nn_user_id CHECK (user_id IS NOT NULL);
ALTER TABLE cards ADD CONSTRAINT nn_card_number CHECK (card_number IS NOT NULL);

-- sub_durations
ALTER TABLE sub_durations ADD PRIMARY KEY (sub_duration_id);
ALTER TABLE sub_durations ADD CONSTRAINT fk_sub_durations_service 
    FOREIGN KEY (service_id) REFERENCES services(service_id);
ALTER TABLE sub_durations ADD CONSTRAINT check_duration_positive 
    CHECK (duration_days > 0);
ALTER TABLE sub_durations ADD CONSTRAINT nn_sub_duration_id CHECK (sub_duration_id IS NOT NULL);
ALTER TABLE sub_durations ADD CONSTRAINT nn_service_id CHECK (service_id IS NOT NULL);
ALTER TABLE sub_durations ADD CONSTRAINT nn_duration_days CHECK (duration_days IS NOT NULL);

-- promocodes
ALTER TABLE promocodes ADD PRIMARY KEY (promocode_id);
ALTER TABLE promocodes ADD CONSTRAINT nn_service_id CHECK (service_id IS NOT NULL);
ALTER TABLE promocodes ADD CONSTRAINT nn_promocode CHECK (promocode IS NOT NULL);
ALTER TABLE promocodes ADD CONSTRAINT nn_sub_duration_days CHECK (sub_duration_days IS NOT NULL);
ALTER TABLE promocodes ADD CONSTRAINT nn_sub_id CHECK (sub_id IS NOT NULL);
ALTER TABLE promocodes ADD CONSTRAINT nn_expires_at CHECK (expires_at IS NOT NULL);
ALTER TABLE promocodes ADD CONSTRAINT fk_promocodes_service 
    FOREIGN KEY (service_id) REFERENCES services(service_id);
ALTER TABLE promocodes ADD CONSTRAINT fk_promocodes_subscription 
    FOREIGN KEY (sub_id) REFERENCES subscriptions(sub_id);

-- payments
ALTER TABLE payments ADD PRIMARY KEY (paym_id);
ALTER TABLE payments ADD CONSTRAINT nn_user_id CHECK (user_id IS NOT NULL);
ALTER TABLE payments ADD CONSTRAINT nn_amount CHECK (amount IS NOT NULL);
ALTER TABLE payments ADD CONSTRAINT nn_paym_type CHECK (paym_type IS NOT NULL);
ALTER TABLE payments ADD CONSTRAINT fk_payments_user 
    FOREIGN KEY (user_id) REFERENCES users(user_id);
ALTER TABLE payments ADD CONSTRAINT fk_payments_card 
    FOREIGN KEY (card_number) REFERENCES cards(card_number);
ALTER TABLE payments ADD CONSTRAINT check_amount_positive 
    CHECK (amount > 0);
ALTER TABLE payments ADD CONSTRAINT valid_card_number 
    CHECK (
        (paym_type = 'expence' AND card_number IS NULL) OR 
        (paym_type = 'income' AND card_number IS NOT NULL)
    );

COMMIT;
