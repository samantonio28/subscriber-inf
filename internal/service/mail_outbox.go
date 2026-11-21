package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/redis"
)

// OutboxMailer — реализация domain.Mailer для dev/тестового окружения.
// Вместо реального SMTP «письмо» (конверт с получателем, темой и телом)
// кладётся в Redis под ключом mail:outbox:<email> с TTL MailOutboxTTL —
// это эмуляция почтового ящика получателя: любой клиент, у которого есть
// доступ к Redis (в том числе E2E-тесты), может прочитать письмо и извлечь
// из него одноразовый код.
//
// В продакшне на это место подключается SMTP-отправка: интерфейс
// domain.Mailer изолирует usecase-слой от способа доставки.
type OutboxMailer struct {
	client *redis.Client
}

var _ domain.Mailer = (*OutboxMailer)(nil)

// mailEnvelope — структура «письма», сохраняемого в outbox.
type mailEnvelope struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewOutboxMailer(client *redis.Client) (*OutboxMailer, error) {
	if client == nil {
		return nil, domain.ErrInvalidCache
	}
	return &OutboxMailer{client: client}, nil
}

// Send сохраняет письмо в Redis-outbox (эмуляция доставки на email).
func (m *OutboxMailer) Send(ctx context.Context, to, subject, body string) error {
	if to == "" {
		return fmt.Errorf("mailer: recipient is empty")
	}
	envelope := mailEnvelope{To: to, Subject: subject, Body: body}
	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("mailer: marshal envelope: %w", err)
	}
	if err := m.client.Set(ctx, domain.MailOutboxKey(to), string(data), domain.MailOutboxTTL); err != nil {
		return fmt.Errorf("mailer: store outbox letter: %w", err)
	}
	return nil
}

// ReadOutbox — чтение «почтового ящика» (используется E2E-тестами, которые
// исполняют роль получателя письма). Возвращает последнюю конвертацию письма.
func ReadOutbox(ctx context.Context, client *redis.Client, email string) (mailEnvelope, error) {
	var envelope mailEnvelope
	raw, err := client.Get(ctx, domain.MailOutboxKey(email))
	if err != nil {
		return envelope, fmt.Errorf("no letter in outbox for %s: %w", email, err)
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return envelope, fmt.Errorf("parse outbox letter: %w", err)
	}
	return envelope, nil
}
