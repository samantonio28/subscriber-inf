package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/icrowley/fake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

// addUsers добавляет 1000 пользователей и возвращает их UUID
func addUsers(pool *pgxpool.Pool) ([]uuid.UUID, error) {
	userIDs := make([]uuid.UUID, 0, 1000)

	for i := range 1000 {
		userID := uuid.New()
		userIDs = append(userIDs, userID)

		email := fake.EmailAddress()
		password := fake.Password(8, 30, true, true, true)
		userName := fake.UserName()
		if len(userName) > 20 {
			userName = userName[:20]
		}
		age := 18 + rand.Intn(50)   // возраст от 18 до 67
		balance := rand.Intn(10000) // баланс от 0 до 9999

		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (user_id, email, password, user_name, age, balance)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, email, password, userName, age, balance)

		if err != nil {
			return nil, fmt.Errorf("failed to insert user %d: %w", i, err)
		}
	}

	log.Printf("Added %d users", len(userIDs))
	return userIDs, nil
}

// addServices добавляет сервисы
func addServices(pool *pgxpool.Pool) ([]int, error) {
	services := []struct {
		name       string
		duration   int
		usersCount int
		hasPromo   bool
	}{
		{"Netflix", 30, 1, true},
		{"Spotify", 30, 1, true},
		{"YouTube Premium", 30, 6, true},
		{"Disney+", 30, 4, false},
		{"HBO Max", 30, 3, true},
		{"Amazon Prime", 30, 2, false},
		{"Apple Music", 30, 1, true},
		{"Microsoft 365", 365, 6, true},
		{"Adobe Creative Cloud", 365, 1, false},
		{"PlayStation Plus", 90, 1, true},
	}

	serviceIDs := make([]int, 0, len(services))

	for _, service := range services {
		var serviceID int
		err := pool.QueryRow(context.Background(), `
			INSERT INTO services (service_name, sub_duration_id_default, users_count, has_promocodes)
			VALUES ($1, $2, $3, $4)
			RETURNING service_id
		`, service.name, 1, service.usersCount, service.hasPromo).Scan(&serviceID)

		if err != nil {
			return nil, fmt.Errorf("failed to insert service %s: %w", service.name, err)
		}

		serviceIDs = append(serviceIDs, serviceID)

		// Добавляем продолжительность подписки
		_, err = pool.Exec(context.Background(), `
			INSERT INTO sub_durations (service_id, duration_days)
			VALUES ($1, $2)
		`, serviceID, service.duration)

		if err != nil {
			return nil, fmt.Errorf("failed to insert sub duration for service %d: %w", serviceID, err)
		}
	}

	log.Printf("Added %d services", len(serviceIDs))
	return serviceIDs, nil
}

// addSubscriptions добавляет подписки
func addSubscriptions(pool *pgxpool.Pool, userIDs []uuid.UUID, serviceIDs []int) ([]int, error) {
	subTypes := []string{"usual", "promocode", "family"}
	subscriptionIDs := make([]int, 0, 3000)

	for i := range 3000 {
		userID := userIDs[rand.Intn(len(userIDs))]
		serviceID := serviceIDs[rand.Intn(len(serviceIDs))]
		price := 100 + rand.Intn(4900) // цена от 100 до 4999
		subType := subTypes[rand.Intn(len(subTypes))]

		startDate := time.Now().AddDate(0, -rand.Intn(12), 0) // подписка началась от 0 до 12 месяцев назад
		// Округляем до первого числа месяца
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

		var endDate time.Time
		if subType == "promocode" {
			// Для промокодов может быть NULL
			if rand.Float32() < 0.3 {
				endDate = startDate.AddDate(0, rand.Intn(12)+1, 0)
				endDate = time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
			}
		} else {
			endDate = startDate.AddDate(0, rand.Intn(12)+1, 0)
			endDate = time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		}

		var subID int
		var err error

		if subType == "promocode" && endDate.IsZero() {
			err = pool.QueryRow(context.Background(), `
				INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date)
				VALUES ($1, $2, $3, $4, NULL, NULL)
				RETURNING sub_id
			`, userID, serviceID, price, subType).Scan(&subID)
		} else if endDate.IsZero() {
			err = pool.QueryRow(context.Background(), `
				INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date)
				VALUES ($1, $2, $3, $4, $5, NULL)
				RETURNING sub_id
			`, userID, serviceID, price, subType, startDate).Scan(&subID)
		} else {
			err = pool.QueryRow(context.Background(), `
				INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING sub_id
			`, userID, serviceID, price, subType, startDate, endDate).Scan(&subID)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to insert subscription %d: %w", i, err)
		}

		subscriptionIDs = append(subscriptionIDs, subID)
	}

	log.Printf("Added %d subscriptions", len(subscriptionIDs))
	return subscriptionIDs, nil
}

// addPromocodes добавляет промокоды
func addPromocodes(pool *pgxpool.Pool, serviceIDs []int, _ []int) error {
	// Берем только подписки типа promocode
	var promocodeSubIDs []int
	err := pool.QueryRow(context.Background(), `
		SELECT array_agg(sub_id) FROM subscriptions WHERE sub_type = 'promocode'
	`).Scan(&promocodeSubIDs)

	if err != nil {
		return fmt.Errorf("failed to get promocode subscriptions: %w", err)
	}

	if len(promocodeSubIDs) == 0 {
		log.Println("No promocode subscriptions found")
		return nil
	}

	for i, subID := range promocodeSubIDs {
		if i >= 1000 {
			break
		}

		serviceID := serviceIDs[rand.Intn(len(serviceIDs))]
		promocode := fake.CharactersN(10)
		durationDays := 30 * (rand.Intn(12) + 1)              // от 1 до 12 месяцев
		expiresAt := time.Now().AddDate(0, rand.Intn(6)+1, 0) // истекает через 1-6 месяцев

		_, err := pool.Exec(context.Background(), `
			INSERT INTO promocodes (promocode_id, service_id, promocode, sub_duration_days, sub_id, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, i+1, serviceID, promocode, durationDays, subID, expiresAt)

		if err != nil {
			return fmt.Errorf("failed to insert promocode %d: %w", i, err)
		}
	}

	log.Printf("Added %d promocodes", min(500, len(promocodeSubIDs)))
	return nil
}

// addCards добавляет карты для пользователей
func addCards(pool *pgxpool.Pool, userIDs []uuid.UUID) error {
	for i, userID := range userIDs {
		cardNumber := fake.CreditCardNum("")
		_, err := pool.Exec(context.Background(), `
			INSERT INTO cards (user_id, card_number)
			VALUES ($1, $2)
		`, userID, cardNumber)

		if err != nil {
			return fmt.Errorf("failed to insert card for user %d: %w", i, err)
		}
	}

	log.Printf("Added cards for %d users", len(userIDs))
	return nil
}

// addPayments добавляет платежи
func addPayments(pool *pgxpool.Pool, userIDs []uuid.UUID) error {
	rows, err := pool.Query(context.Background(), `
		SELECT user_id, card_number FROM cards
	`)
	if err != nil {
		return fmt.Errorf("failed to get cards: %w", err)
	}
	defer rows.Close()

	userCards := make(map[uuid.UUID][]string)
	for rows.Next() {
		var userID uuid.UUID
		var cardNumber string
		if err := rows.Scan(&userID, &cardNumber); err != nil {
			return fmt.Errorf("failed to scan card: %w", err)
		}
		userCards[userID] = append(userCards[userID], cardNumber)
	}

	// Добавляем платежи
	for i := range 5000 {
		userID := userIDs[rand.Intn(len(userIDs))]
		amount := 50 + rand.Intn(4950) // сумма от 50 до 4999

		var cardNumber *string
		if cards, exists := userCards[userID]; exists && len(cards) > 0 && rand.Float32() < 0.7 {
			// 70% chance to use card for income
			cn := cards[rand.Intn(len(cards))]
			cardNumber = &cn
		}

		var paymType string
		if cardNumber != nil {
			paymType = "income"
		} else {
			paymType = "expence"
		}

		if cardNumber != nil {
			_, err = pool.Exec(context.Background(), `
				INSERT INTO payments (user_id, card_number, amount, paym_type)
				VALUES ($1, $2, $3, $4)
			`, userID, *cardNumber, amount, paymType)
		} else {
			_, err = pool.Exec(context.Background(), `
				INSERT INTO payments (user_id, amount, paym_type)
				VALUES ($1, $2, $3)
			`, userID, amount, paymType)
		}

		if err != nil {
			return fmt.Errorf("failed to insert payment %d: %w", i, err)
		}
	}

	log.Printf("Added 5000 payments")
	return nil
}

func genData() error {
	cfg, err := config.LoadConfig("../configs/postgres.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
		return err
	}

	poolConfig, err := cfg.Postgres.ToPgxPoolConfig()
	if err != nil {
		log.Fatal("Failed to create pool config:", err)
		return err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
		return err
	}
	defer pool.Close()

	log.Println("Successfully connected to PostgreSQL!")

	// Очищаем существующие данные 
	_, err = pool.Exec(context.Background(), `
		TRUNCATE TABLE payments, promocodes, subscriptions, cards, users, sub_durations, services RESTART IDENTITY CASCADE
	`)
	if err != nil {
		log.Println("Warning: could not truncate tables:", err)
	}

	userIDs, err := addUsers(pool)
	if err != nil {
		return fmt.Errorf("error adding users: %w", err)
	}

	serviceIDs, err := addServices(pool)
	if err != nil {
		return fmt.Errorf("error adding services: %w", err)
	}

	subscriptionIDs, err := addSubscriptions(pool, userIDs, serviceIDs)
	if err != nil {
		return fmt.Errorf("error adding subscriptions: %w", err)
	}

	if err := addCards(pool, userIDs); err != nil {
		return fmt.Errorf("error adding cards: %w", err)
	}

	if err := addPromocodes(pool, serviceIDs, subscriptionIDs); err != nil {
		return fmt.Errorf("error adding promocodes: %w", err)
	}

	if err := addPayments(pool, userIDs); err != nil {
		return fmt.Errorf("error adding payments: %w", err)
	}

	return nil
}

func main() {
	log.Println("starting generating data")
	if err := genData(); err != nil {
		log.Println("cant gen data", err)
	} else {
		log.Println("data generation completed successfully")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
