BEGIN;

ALTER TABLE subscriptions
ADD COLUMN plan_id INTEGER,
ADD COLUMN promocode_id INTEGER;

ALTER TABLE subscriptions
ALTER COLUMN plan_id SET NOT NULL;

ALTER TABLE subscriptions
DROP CONSTRAINT IF EXISTS fk_subscriptions_service,
DROP CONSTRAINT IF EXISTS nn_service_id;

ALTER TABLE subscriptions
DROP COLUMN service_id;

ALTER TABLE subscriptions
ADD CONSTRAINT fk_subscriptions_plan
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(plan_id),
ADD CONSTRAINT fk_subscriptions_promocode
    FOREIGN KEY (promocode_id) REFERENCES promocodes(promocode_id);

ALTER TABLE subscriptions
DROP COLUMN price;

ALTER TABLE promocodes
ALTER COLUMN expires_at TYPE TIMESTAMP USING expires_at::TIMESTAMP;

ALTER TABLE promocodes
ALTER COLUMN expires_at DROP NOT NULL;

ALTER TABLE promocodes
DROP COLUMN expires_at;

ALTER TABLE promocodes
ADD COLUMN expires_at TIMESTAMP
GENERATED ALWAYS AS (created_at + (COALESCE(duration_days, 3) || ' days')::INTERVAL) STORED;

UPDATE promocodes SET duration_days = 3 WHERE duration_days IS NULL;
ALTER TABLE promocodes
ALTER COLUMN duration_days SET DEFAULT 3,
ALTER COLUMN duration_days SET NOT NULL;

COMMIT;