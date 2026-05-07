-- MySQL schema for subscriber-inf project
-- Adapted from PostgreSQL migrations

-- Disable foreign key checks for easier initialization
SET FOREIGN_KEY_CHECKS = 0;

-- Create database if not exists (will be created by docker)
-- USE dev;

-- ENUM types are simulated using ENUM columns
-- No need to create separate types

CREATE TABLE services (
    service_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    sub_duration_id_default INTEGER NOT NULL DEFAULT 1,
    users_count INTEGER NOT NULL DEFAULT 1,
    has_promocodes BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE KEY unique_service_name (service_name)
);

CREATE TABLE users (
    user_id CHAR(36) NOT NULL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(30) NOT NULL,
    user_name VARCHAR(20) NOT NULL,
    age INTEGER NOT NULL,
    balance INTEGER NOT NULL DEFAULT 0,
    referral_code VARCHAR(20) NULL,
    CHECK (age >= 18),
    CHECK (balance >= 0)
);

CREATE TABLE subscriptions (
    sub_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    service_id INTEGER NOT NULL,
    price INTEGER NOT NULL,
    sub_type ENUM('usual', 'promocode', 'family') NOT NULL DEFAULT 'usual',
    start_date DATE NOT NULL,
    end_date DATE NULL,
    promocode_id INTEGER NULL,
    plan_id INTEGER NULL,
    CHECK (price > 0),
    CHECK (DAY(start_date) = 1),
    CHECK (end_date IS NULL OR (DAY(end_date) = 1 AND end_date >= start_date)),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

CREATE TABLE sub_durations (
    sub_duration_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    service_id INTEGER NOT NULL,
    duration_days INTEGER NOT NULL,
    CHECK (duration_days > 0),
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

CREATE TABLE promocodes (
    promocode_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    service_id INTEGER NOT NULL,
    promocode VARCHAR(10) NOT NULL,
    plan_id INTEGER NULL,
    sub_id INTEGER NULL,
    expires_at DATE NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    discount INTEGER NOT NULL DEFAULT 0,
    max_uses INTEGER NOT NULL DEFAULT 1,
    cur_uses INTEGER NOT NULL DEFAULT 0,
    status ENUM('ACTIVE', 'INACTIVE', 'EXPIRED') NOT NULL DEFAULT 'ACTIVE',
    duration_days INTEGER NULL,
    UNIQUE KEY unique_promocode (promocode),
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

CREATE TABLE cards (
    user_id CHAR(36) NOT NULL,
    card_number VARCHAR(30) NOT NULL PRIMARY KEY,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE payments (
    paym_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    card_number VARCHAR(30) NULL,
    amount INTEGER NOT NULL,
    paym_type ENUM('income', 'expence') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE subscription_plans (
    plan_id INTEGER AUTO_INCREMENT PRIMARY KEY,
    service_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    duration_days INTEGER NOT NULL,
    price INTEGER NOT NULL,
    FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

CREATE TABLE user_referrals (
    referrer_id CHAR(36) NOT NULL,
    referred_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (referred_id),
    FOREIGN KEY (referrer_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (referred_id) REFERENCES users(user_id) ON DELETE CASCADE
);

SET FOREIGN_KEY_CHECKS = 1;
