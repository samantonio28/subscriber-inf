BEGIN;

-- Таблица для связи многие-ко-многим между пользователями и сервисами (например, избранное)
CREATE TABLE user_services (
    user_id UUID NOT NULL,
    service_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, service_id),
    CONSTRAINT fk_user_services_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    CONSTRAINT fk_user_services_service FOREIGN KEY (service_id) REFERENCES services(service_id) ON DELETE CASCADE
);

COMMIT;
