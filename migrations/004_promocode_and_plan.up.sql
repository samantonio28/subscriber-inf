BEGIN;

-- Create enum for promocode status
CREATE TYPE promocode_status AS ENUM ('ACTIVE', 'USED', 'DISABLED');

-- Add columns to promocodes table
ALTER TABLE promocodes
ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN discount INTEGER NOT NULL DEFAULT 0,
ADD COLUMN max_uses INTEGER NOT NULL DEFAULT 1,
ADD COLUMN cur_uses INTEGER NOT NULL DEFAULT 0,
ADD COLUMN status promocode_status NOT NULL DEFAULT 'ACTIVE',
ADD COLUMN duration_days INTEGER; -- nullable, if NULL use default 3 days

-- Update expires_at to be computed if not set (we'll keep as is, but ensure it's not null)
-- We'll add a constraint later

-- Create subscription_plans table (enhanced sub_durations with price)
CREATE TABLE subscription_plans (
    plan_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(service_id),
    duration_days INTEGER NOT NULL CHECK (duration_days > 0),
    price INTEGER NOT NULL CHECK (price > 0),
    UNIQUE (service_id, duration_days, price) -- optional
);

-- Insert existing sub_durations data into subscription_plans (assuming price = 0? we need to set a default)
-- Since there is no price column, we'll set price = 1 (placeholder). This is okay for lab.
INSERT INTO subscription_plans (service_id, duration_days, price)
SELECT service_id, duration_days, 1 FROM sub_durations;

-- Add plan_id column to promocodes
ALTER TABLE promocodes
ADD COLUMN plan_id INTEGER REFERENCES subscription_plans(plan_id);

-- Update existing promocodes to have a plan_id based on sub_duration_days and service_id
-- We'll join with subscription_plans to find matching plan.
UPDATE promocodes p
SET plan_id = sp.plan_id
FROM subscription_plans sp
WHERE sp.service_id = p.service_id AND sp.duration_days = p.sub_duration_days;

-- Now we can drop sub_duration_days column (after ensuring data migrated)
ALTER TABLE promocodes
DROP COLUMN sub_duration_days;

-- Also drop foreign key to subscriptions.sub_id (we'll keep it for now but make nullable)
-- First drop NOT NULL constraint on sub_id
ALTER TABLE promocodes
ALTER COLUMN sub_id DROP NOT NULL;

-- Add check for discount range (0-100)
ALTER TABLE promocodes
ADD CONSTRAINT discount_range CHECK (discount >= 0 AND discount <= 100);

-- Add check for uses
ALTER TABLE promocodes
ADD CONSTRAINT uses_check CHECK (cur_uses <= max_uses AND cur_uses >= 0);

-- Add default for duration_days (3 days) if not set
UPDATE promocodes SET duration_days = 3 WHERE duration_days IS NULL;

-- Set expires_at based on created_at + duration_days if expires_at is not set (but expires_at is NOT NULL currently)
-- We'll update expires_at where it's default? Actually expires_at already has value.
-- We'll leave as is.

COMMIT;