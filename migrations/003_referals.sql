BEGIN;

-- Таблица для хранения реферальных связей
CREATE TABLE user_referrals (
    referrer_id UUID NOT NULL,    -- кто пригласил
    referred_id UUID NOT NULL,    -- кого пригласили
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (referred_id),
    FOREIGN KEY (referrer_id) REFERENCES users(user_id),
    FOREIGN KEY (referred_id) REFERENCES users(user_id),
    
    -- Нельзя пригласить самого себя
    CONSTRAINT no_self_referral CHECK (referrer_id != referred_id)
);

COMMIT;
