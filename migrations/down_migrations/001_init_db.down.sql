BEGIN;

DROP INDEX IF EXISTS idx_subscriptions_dates;
DROP INDEX IF EXISTS idx_users_subs_user;
DROP INDEX IF EXISTS idx_subscriptions_service;

DROP TABLE IF EXISTS users_subs;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS services;

COMMIT;
