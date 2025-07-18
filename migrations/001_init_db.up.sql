BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE services (
    service_id INTEGER PRIMARY KEY,
    service_name VARCHAR(50) NOT NULL
);

CREATE TABLE subscriptions (
    sub_id INTEGER PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(service_id),
    price INTEGER NOT NULL CHECK (price > 0),
    start_date DATE NOT NULL,
    end_date DATE,
    CONSTRAINT valid_start_date CHECK (EXTRACT(DAY FROM start_date) = 1),
    CONSTRAINT valid_end_date CHECK (
        end_date IS NULL OR 
        (EXTRACT(DAY FROM end_date) = 1 AND end_date >= start_date)
    )
);

CREATE TABLE users_subs (
    sub_id INTEGER PRIMARY KEY REFERENCES subscriptions(sub_id) ON DELETE CASCADE,
    user_id UUID NOT NULL
);

CREATE INDEX idx_subscriptions_service ON subscriptions(service_id);
CREATE INDEX idx_users_subs_user ON users_subs(user_id);
CREATE INDEX idx_subscriptions_dates ON subscriptions(start_date, end_date);

COMMIT;
