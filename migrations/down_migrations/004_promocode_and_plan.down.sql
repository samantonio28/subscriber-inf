BEGIN;

-- Remove checks
ALTER TABLE promocodes DROP CONSTRAINT IF EXISTS discount_range;
ALTER TABLE promocodes DROP CONSTRAINT IF EXISTS uses_check;

-- Restore sub_duration_days column (we need to derive from plan_id)
ALTER TABLE promocodes
ADD COLUMN sub_duration_days INTEGER;

UPDATE promocodes p
SET sub_duration_days = sp.duration_days
FROM subscription_plans sp
WHERE p.plan_id = sp.plan_id;

ALTER TABLE promocodes
ALTER COLUMN sub_duration_days SET NOT NULL;

-- Drop plan_id column
ALTER TABLE promocodes
DROP COLUMN plan_id;

-- Restore NOT NULL constraint on sub_id
ALTER TABLE promocodes
ALTER COLUMN sub_id SET NOT NULL;

-- Drop added columns
ALTER TABLE promocodes
DROP COLUMN created_at,
DROP COLUMN discount,
DROP COLUMN max_uses,
DROP COLUMN cur_uses,
DROP COLUMN status,
DROP COLUMN duration_days;

-- Drop subscription_plans table
DROP TABLE subscription_plans;

-- Drop enum type
DROP TYPE promocode_status;

COMMIT;