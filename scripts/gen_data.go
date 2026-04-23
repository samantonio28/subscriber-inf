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
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/service"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

// addUsers добавляет 1000 пользователей и создает реферальные связи
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
		age := 18 + rand.Intn(50)
		balance := rand.Intn(10000)

		_, err := pool.Exec(context.Background(), `
			INSERT INTO users (user_id, email, password, user_name, age, balance)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, userID, email, password, userName, age, balance)

		if err != nil {
			return nil, fmt.Errorf("failed to insert user %d: %w", i, err)
		}
	}

	log.Printf("Added %d users", len(userIDs))

	// Теперь создаем реферальные связи
	if err := addReferrals(pool, userIDs); err != nil {
		return nil, fmt.Errorf("failed to add referrals: %w", err)
	}

	return userIDs, nil
}

// addReferrals создает реферальные связи для случайных пользователей
func addReferrals(pool *pgxpool.Pool, userIDs []uuid.UUID) error {
	// Выбираем 50 случайных пользователей, которые будут приглашать других
	referrers := make([]uuid.UUID, 50)
	for i := range referrers {
		referrers[i] = userIDs[rand.Intn(len(userIDs))]
	}

	referralCount := 0

	// Для каждого приглашающего создаем несколько приглашенных
	for _, referrerID := range referrers {
		// Каждый приглашающий приглашает от 3 до 10 человек
		numReferrals := 3 + rand.Intn(8)

		for j := 0; j < numReferrals; j++ {
			// Выбираем случайного пользователя как приглашенного
			referredID := userIDs[rand.Intn(len(userIDs))]

			// Проверяем, чтобы не приглашать самого себя
			if referredID == referrerID {
				continue
			}

			// Проверяем, чтобы пользователь не был уже чьим-то рефералом
			var exists bool
			err := pool.QueryRow(context.Background(), `
				SELECT EXISTS(SELECT 1 FROM user_referrals WHERE referred_id = $1)
			`, referredID).Scan(&exists)

			if err != nil {
				return fmt.Errorf("failed to check referral existence: %w", err)
			}

			if exists {
				continue // этот пользователь уже чей-то реферал
			}

			// Добавляем реферальную связь
			_, err = pool.Exec(context.Background(), `
				INSERT INTO user_referrals (referrer_id, referred_id, created_at)
				VALUES ($1, $2, $3)
			`, referrerID, referredID, time.Now().Add(-time.Duration(rand.Intn(365))*24*time.Hour))

			if err != nil {
				log.Printf("Warning: failed to insert referral: %v", err)
			} else {
				referralCount++
			}
		}
	}

	log.Printf("Added %d referral relationships", referralCount)

	// Логируем статистику
	if err := logReferralStats(pool); err != nil {
		return fmt.Errorf("failed to log referral stats: %w", err)
	}

	return nil
}

// logReferralStats логирует статистику по рефералам
func logReferralStats(pool *pgxpool.Pool) error {
	// Топ пользователей по количеству приглашенных
	rows, err := pool.Query(context.Background(), `
		SELECT u.user_name, COUNT(ur.referred_id) as referral_count
		FROM users u
		JOIN user_referrals ur ON u.user_id = ur.referrer_id
		GROUP BY u.user_id, u.user_name
		ORDER BY referral_count DESC
		LIMIT 10
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	log.Println("Top referrers:")
	for rows.Next() {
		var userName string
		var count int
		if err := rows.Scan(&userName, &count); err != nil {
			return err
		}
		log.Printf("  %s: %d referrals", userName, count)
	}

	return nil
}

// generatePlanName создает название плана на основе имени сервиса и длительности
func generatePlanName(serviceName string, durationDays int) string {
	switch {
	case durationDays <= 30:
		return serviceName + " M"
	case durationDays <= 90:
		return serviceName + " Q"
	case durationDays <= 365:
		return serviceName + " A"
	default:
		return serviceName + " C"
	}
}

// addServices добавляет сервисы и возвращает ID сервисов и ID планов
func addServices(pool *pgxpool.Pool) (serviceIDs []int, planIDs []int, err error) {
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

	serviceIDs = make([]int, 0, len(services))
	planIDs = make([]int, 0, len(services))

	// Убедимся, что необходимые таблицы существуют (на случай, если миграции не применились)
	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS services (
			service_id INTEGER GENERATED ALWAYS AS IDENTITY,
			service_name VARCHAR,
			sub_duration_id_default INTEGER,
			users_count INTEGER,
			has_promocodes BOOLEAN
		)
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ensure services table: %w", err)
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS sub_durations (
			sub_duration_id INTEGER GENERATED BY DEFAULT AS IDENTITY,
			service_id INTEGER,
			duration_days INTEGER
		)
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ensure sub_durations table: %w", err)
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS subscription_plans (
			plan_id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			service_id INTEGER NOT NULL,
			duration_days INTEGER NOT NULL CHECK (duration_days > 0),
			name VARCHAR(255) NOT NULL,
			price INTEGER NOT NULL CHECK (price > 0),
			UNIQUE (service_id, duration_days, price)
		)
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ensure subscription_plans table: %w", err)
	}

	for _, service := range services {
		var serviceID int
		err := pool.QueryRow(context.Background(), `
			INSERT INTO services (service_name, sub_duration_id_default, users_count, has_promocodes)
			VALUES ($1, $2, $3, $4)
			RETURNING service_id
		`, service.name, 1, service.usersCount, service.hasPromo).Scan(&serviceID)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to insert service %s: %w", service.name, err)
		}

		serviceIDs = append(serviceIDs, serviceID)

		// Добавляем продолжительность подписки
		_, err = pool.Exec(context.Background(), `
			INSERT INTO sub_durations (service_id, duration_days)
			VALUES ($1, $2)
		`, serviceID, service.duration)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to insert sub duration for service %d: %w", serviceID, err)
		}

		// Добавляем план подписки
		price := 100 + rand.Intn(4900) // цена от 100 до 4999
		planName := generatePlanName(service.name, service.duration)
		var planID int
		err = pool.QueryRow(context.Background(), `
			INSERT INTO subscription_plans (service_id, duration_days, name, price)
			VALUES ($1, $2, $3, $4)
			RETURNING plan_id
		`, serviceID, service.duration, planName, price).Scan(&planID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to insert subscription plan for service %d: %w", serviceID, err)
		}
		planIDs = append(planIDs, planID)
	}

	log.Printf("Added %d services and %d plans", len(serviceIDs), len(planIDs))
	return serviceIDs, planIDs, nil
}

// addSubscriptions добавляет подписки с использованием юзкейса
func addSubscriptions(pool *pgxpool.Pool, userIDs []uuid.UUID, planIDs []int, createSubUC *usecase.CreateSubUC) ([]int, error) {
	subTypes := []string{"usual", "promocode", "family"}
	subscriptionIDs := make([]int, 0, 3000)

	// Кэшируем mapping planID -> serviceName для быстрого доступа
	planServiceMap := make(map[int]string)
	for _, planID := range planIDs {
		var serviceName string
		err := pool.QueryRow(context.Background(), `
			SELECT s.service_name
			FROM subscription_plans sp
			JOIN services s ON sp.service_id = s.service_id
			WHERE sp.plan_id = $1
		`, planID).Scan(&serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get service name for plan %d: %w", planID, err)
		}
		planServiceMap[planID] = serviceName
	}

	for i := range 3000 {
		userID := userIDs[rand.Intn(len(userIDs))]
		planID := planIDs[rand.Intn(len(planIDs))]
		price := 100 + rand.Intn(4900) // цена от 100 до 4999
		subType := subTypes[rand.Intn(len(subTypes))]

		startDate := time.Now().AddDate(0, -rand.Intn(12), 0) // подписка началась от 0 до 12 месяцев назад
		// Округляем до первого числа месяца
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

		var endDate time.Time
		endDate = startDate.AddDate(0, rand.Intn(12)+1, 0)
		endDate = time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)

		serviceName, ok := planServiceMap[planID]
		if !ok {
			return nil, fmt.Errorf("service name not found for plan %d", planID)
		}

		// Создаем DTO для подписки
		dto := usecase.SubscriptionDTO{
			SubId:       0, // будет сгенерировано
			UserId:      userID,
			ServiceName: serviceName,
			Price:       price,
			SubType:     subType,
			StartDate:   startDate,
			EndDate:     endDate,
			PlanID:      planID,
			PromocodeID: nil,
		}

		// Используем юзкейс для создания подписки
		subID, err := createSubUC.NewSub(context.Background(), dto)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription via usecase %d: %w", i, err)
		}

		subscriptionIDs = append(subscriptionIDs, subID)
	}

	log.Printf("Added %d subscriptions using usecase", len(subscriptionIDs))
	return subscriptionIDs, nil
}

// addPromocodes добавляет промокоды с использованием юзкейса
func addPromocodes(pool *pgxpool.Pool, serviceIDs []int, subscriptionIDs []int, createPromoUC *usecase.CreatePromocodeUC) error {
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
		promocodeValue := fake.CharactersN(10)
		expiresAt := time.Now().AddDate(0, rand.Intn(6)+1, 0) // истекает через 1-6 месяцев
		discount := rand.Intn(101)                            // 0-100
		maxUses := 1 + rand.Intn(10)                          // 1-10
		durationDays := rand.Intn(365) + 1                    // 1-365 дней

		// Получаем plan_id для данного сервиса
		var planID int
		err := pool.QueryRow(context.Background(), `
			SELECT plan_id FROM subscription_plans WHERE service_id = $1 LIMIT 1
		`, serviceID).Scan(&planID)
		if err != nil {
			// Если план не найден, пропускаем или используем дефолтный? Пропустим с логом.
			log.Printf("Warning: no plan found for service %d, skipping promocode", serviceID)
			continue
		}

		// Используем юзкейс для создания промокода
		input := usecase.CreatePromocodeInput{
			ServiceID:    serviceID,
			Value:        promocodeValue,
			PlanID:       &planID,
			SubID:        &subID,
			ExpiresAt:    expiresAt,
			Discount:     discount,
			MaxUses:      maxUses,
			DurationDays: durationDays,
		}

		_, err = createPromoUC.Create(context.Background(), input)
		if err != nil {
			return fmt.Errorf("failed to create promocode %d via usecase: %w", i, err)
		}
	}

	log.Printf("Added %d promocodes using usecase", min(500, len(promocodeSubIDs)))
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

	// Инициализация репозиториев и юзкейсов
	subRepo, err := service.NewSubRepo(pool)
	if err != nil {
		return fmt.Errorf("failed to create sub repo: %w", err)
	}

	promoRepo, err := service.NewPromocodeRepo(pool)
	if err != nil {
		return fmt.Errorf("failed to create promocode repo: %w", err)
	}

	logger, err := logger.NewLogrusLogger("logs/gen_data.log")
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	createSubUC, err := usecase.NewCreateSubUC(subRepo, logger)
	if err != nil {
		return fmt.Errorf("failed to create CreateSubUC: %w", err)
	}

	createPromoUC, err := usecase.NewCreatePromocodeUC(promoRepo, logger)
	if err != nil {
		return fmt.Errorf("failed to create CreatePromocodeUC: %w", err)
	}

	userIDs, err := addUsers(pool)
	if err != nil {
		return fmt.Errorf("error adding users: %w", err)
	}

	serviceIDs, planIDs, err := addServices(pool)
	if err != nil {
		return fmt.Errorf("error adding services: %w", err)
	}

	subscriptionIDs, err := addSubscriptions(pool, userIDs, planIDs, createSubUC)
	if err != nil {
		return fmt.Errorf("error adding subscriptions: %w", err)
	}

	if err := addCards(pool, userIDs); err != nil {
		return fmt.Errorf("error adding cards: %w", err)
	}

	if err := addPromocodes(pool, serviceIDs, subscriptionIDs, createPromoUC); err != nil {
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
