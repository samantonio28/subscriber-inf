package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/icrowley/fake"
	"github.com/samantonio28/subscriber-inf/pkg/config"
)

// DBType constants
const (
	DBTypePostgres = "postgres"
	DBTypeMySQL    = "mysql"
)

// adaptQuery converts PostgreSQL-style placeholders $1, $2, ... to ? for MySQL.
// Also handles RETURNING clause for MySQL (use SELECT LAST_INSERT_ID()).
func adaptQuery(query string, dbType string) string {
	if dbType == DBTypeMySQL {
		// Replace $1, $2, ... with ?
		for i := 1; strings.Contains(query, fmt.Sprintf("$%d", i)); i++ {
			query = strings.ReplaceAll(query, fmt.Sprintf("$%d", i), "?")
		}
		// Replace RETURNING ... with SELECT LAST_INSERT_ID()
		if strings.Contains(strings.ToUpper(query), "RETURNING") {
			// Simple replacement: assume returning a single column (like service_id, plan_id)
			// We'll capture the column name and later use sql.Result.LastInsertId
			// For simplicity, we remove RETURNING clause and rely on LastInsertId.
			// This requires the query to be executed with sql.Exec and then retrieving LastInsertId.
			// We'll handle this in the calling code.
			// For now, just remove RETURNING clause.
			start := strings.Index(strings.ToUpper(query), "RETURNING")
			end := len(query)
			for j := start; j < len(query); j++ {
				if query[j] == ')' || query[j] == ' ' && j > start+9 {
					end = j
					break
				}
			}
			query = query[:start] + query[end:]
		}
	}
	return query
}

// addUsers добавляет 1000 пользователей и создает реферальные связи
func addUsers(db *sql.DB, dbType string) ([]uuid.UUID, error) {
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

		query := `
			INSERT INTO users (user_id, email, password, user_name, age, balance)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		query = adaptQuery(query, dbType)
		_, err := db.ExecContext(context.Background(), query, userID, email, password, userName, age, balance)

		if err != nil {
			return nil, fmt.Errorf("failed to insert user %d: %w", i, err)
		}
	}

	log.Printf("Added %d users", len(userIDs))

	// Теперь создаем реферальные связи
	if err := addReferrals(db, dbType, userIDs); err != nil {
		return nil, fmt.Errorf("failed to add referrals: %w", err)
	}

	return userIDs, nil
}

// addReferrals создает реферальные связи для случайных пользователей
func addReferrals(db *sql.DB, dbType string, userIDs []uuid.UUID) error {
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
			query := `
				SELECT EXISTS(SELECT 1 FROM user_referrals WHERE referred_id = $1)
			`
			query = adaptQuery(query, dbType)
			err := db.QueryRowContext(context.Background(), query, referredID).Scan(&exists)

			if err != nil {
				return fmt.Errorf("failed to check referral existence: %w", err)
			}

			if exists {
				continue // этот пользователь уже чей-то реферал
			}

			// Добавляем реферальную связь
			query = `
				INSERT INTO user_referrals (referrer_id, referred_id, created_at)
				VALUES ($1, $2, $3)
			`
			query = adaptQuery(query, dbType)
			_, err = db.ExecContext(context.Background(), query, referrerID, referredID, time.Now().Add(-time.Duration(rand.Intn(365))*24*time.Hour))

			if err != nil {
				log.Printf("Warning: failed to insert referral: %v", err)
			} else {
				referralCount++
			}
		}
	}

	log.Printf("Added %d referral relationships", referralCount)

	// Логируем статистику
	if err := logReferralStats(db, dbType); err != nil {
		return fmt.Errorf("failed to log referral stats: %w", err)
	}

	return nil
}

// logReferralStats логирует статистику по рефералам
func logReferralStats(db *sql.DB, dbType string) error {
	// Топ пользователей по количеству приглашенных
	query := `
		SELECT u.user_name, COUNT(ur.referred_id) as referral_count
		FROM users u
		JOIN user_referrals ur ON u.user_id = ur.referrer_id
		GROUP BY u.user_id, u.user_name
		ORDER BY referral_count DESC
		LIMIT 10
	`
	// No placeholders, no adaptation needed
	rows, err := db.QueryContext(context.Background(), query)
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
func addServices(db *sql.DB, dbType string) (serviceIDs []int, planIDs []int, err error) {
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

	for _, service := range services {
		var serviceID int64
		query := `
			INSERT INTO services (service_name, sub_duration_id_default, users_count, has_promocodes)
			VALUES ($1, $2, $3, $4)
		`
		if dbType == DBTypePostgres {
			query += " RETURNING service_id"
		}
		query = adaptQuery(query, dbType)
		if dbType == DBTypePostgres {
			err := db.QueryRowContext(context.Background(), query, service.name, 1, service.usersCount, service.hasPromo).Scan(&serviceID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert service %s: %w", service.name, err)
			}
		} else {
			res, err := db.ExecContext(context.Background(), query, service.name, 1, service.usersCount, service.hasPromo)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert service %s: %w", service.name, err)
			}
			serviceID, err = res.LastInsertId()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get last insert id for service %s: %w", service.name, err)
			}
		}

		serviceIDs = append(serviceIDs, int(serviceID))

		// Добавляем продолжительность подписки
		query = `
			INSERT INTO sub_durations (service_id, duration_days)
			VALUES ($1, $2)
		`
		query = adaptQuery(query, dbType)
		_, err = db.ExecContext(context.Background(), query, serviceID, service.duration)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to insert sub duration for service %d: %w", serviceID, err)
		}

		// Добавляем план подписки
		price := 100 + rand.Intn(4900) // цена от 100 до 4999
		planName := generatePlanName(service.name, service.duration)
		var planID int64
		query = `
			INSERT INTO subscription_plans (service_id, duration_days, name, price)
			VALUES ($1, $2, $3, $4)
		`
		if dbType == DBTypePostgres {
			query += " RETURNING plan_id"
		}
		query = adaptQuery(query, dbType)
		if dbType == DBTypePostgres {
			err = db.QueryRowContext(context.Background(), query, serviceID, service.duration, planName, price).Scan(&planID)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert subscription plan for service %d: %w", serviceID, err)
			}
		} else {
			res, err := db.ExecContext(context.Background(), query, serviceID, service.duration, planName, price)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to insert subscription plan for service %d: %w", serviceID, err)
			}
			planID, err = res.LastInsertId()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get last insert id for plan: %w", err)
			}
		}
		planIDs = append(planIDs, int(planID))
	}

	log.Printf("Added %d services and %d plans", len(serviceIDs), len(planIDs))
	return serviceIDs, planIDs, nil
}

// addSubscriptions добавляет подписки (raw SQL, без юзкейса)
func addSubscriptions(db *sql.DB, dbType string, userIDs []uuid.UUID, planIDs []int) ([]int, error) {
	subTypes := []string{"usual", "promocode", "family"}
	subscriptionIDs := make([]int, 0, 3000)

	// Кэшируем mapping planID -> serviceID (для MySQL) или serviceName (для PostgreSQL)
	planServiceMap := make(map[int]interface{})
	for _, planID := range planIDs {
		if dbType == DBTypePostgres {
			var serviceName string
			query := `
				SELECT s.service_name
				FROM subscription_plans sp
				JOIN services s ON sp.service_id = s.service_id
				WHERE sp.plan_id = $1
			`
			query = adaptQuery(query, dbType)
			err := db.QueryRowContext(context.Background(), query, planID).Scan(&serviceName)
			if err != nil {
				return nil, fmt.Errorf("failed to get service name for plan %d: %w", planID, err)
			}
			planServiceMap[planID] = serviceName
		} else {
			var serviceID int
			query := `
				SELECT sp.service_id
				FROM subscription_plans sp
				WHERE sp.plan_id = $1
			`
			query = adaptQuery(query, dbType)
			err := db.QueryRowContext(context.Background(), query, planID).Scan(&serviceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get service id for plan %d: %w", planID, err)
			}
			planServiceMap[planID] = serviceID
		}
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

		serviceVal, ok := planServiceMap[planID]
		if !ok {
			return nil, fmt.Errorf("service not found for plan %d", planID)
		}

		// Вставка подписки
		var subID int64
		var query string
		if dbType == DBTypePostgres {
			query = `
				INSERT INTO subscriptions (user_id, service_name, price, sub_type, start_date, end_date, plan_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`
			query += " RETURNING sub_id"
		} else {
			query = `
				INSERT INTO subscriptions (user_id, service_id, price, sub_type, start_date, end_date, plan_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`
		}
		query = adaptQuery(query, dbType)
		if dbType == DBTypePostgres {
			err := db.QueryRowContext(context.Background(), query, userID, serviceVal, price, subType, startDate, endDate, planID).Scan(&subID)
			if err != nil {
				return nil, fmt.Errorf("failed to create subscription %d: %w", i, err)
			}
		} else {
			res, err := db.ExecContext(context.Background(), query, userID, serviceVal, price, subType, startDate, endDate, planID)
			if err != nil {
				return nil, fmt.Errorf("failed to create subscription %d: %w", i, err)
			}
			subID, err = res.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failed to get last insert id for subscription %d: %w", i, err)
			}
		}
		subscriptionIDs = append(subscriptionIDs, int(subID))
	}

	log.Printf("Added %d subscriptions", len(subscriptionIDs))
	return subscriptionIDs, nil
}

// addPromocodes добавляет промокоды (raw SQL, без юзкейса)
func addPromocodes(db *sql.DB, dbType string, serviceIDs []int, subscriptionIDs []int) error {
	// Берем только подписки типа promocode
	var promocodeSubIDs []int
	query := `
		SELECT sub_id FROM subscriptions WHERE sub_type = 'promocode'
	`
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to get promocode subscriptions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var subID int
		if err := rows.Scan(&subID); err != nil {
			return fmt.Errorf("failed to scan sub_id: %w", err)
		}
		promocodeSubIDs = append(promocodeSubIDs, subID)
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
		selectQuery := `
			SELECT plan_id FROM subscription_plans WHERE service_id = $1 LIMIT 1
		`
		selectQuery = adaptQuery(selectQuery, dbType)
		err := db.QueryRowContext(context.Background(), selectQuery, serviceID).Scan(&planID)
		if err != nil {
			// Если план не найден, пропускаем или используем дефолтный? Пропустим с логом.
			log.Printf("Warning: no plan found for service %d, skipping promocode", serviceID)
			continue
		}

		// Вставка промокода
		var query string
		if dbType == DBTypePostgres {
			query = `
				INSERT INTO promocodes (service_id, plan_id, sub_id, value, expires_at, discount, max_uses, duration_days)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`
		} else {
			// MySQL использует столбец promocode вместо value, также добавляем status и cur_uses
			query = `
				INSERT INTO promocodes (service_id, plan_id, sub_id, promocode, expires_at, discount, max_uses, duration_days, status, cur_uses)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'ACTIVE', 0)
			`
		}
		query = adaptQuery(query, dbType)
		_, err = db.ExecContext(context.Background(), query, serviceID, planID, subID, promocodeValue, expiresAt, discount, maxUses, durationDays)
		if err != nil {
			return fmt.Errorf("failed to create promocode %d: %w", i, err)
		}
	}

	log.Printf("Added %d promocodes", min(500, len(promocodeSubIDs)))
	return nil
}

// addCards добавляет карты для пользователей
func addCards(db *sql.DB, dbType string, userIDs []uuid.UUID) error {
	for i, userID := range userIDs {
		cardNumber := fake.CreditCardNum("")
		query := `
			INSERT INTO cards (user_id, card_number)
			VALUES ($1, $2)
		`
		query = adaptQuery(query, dbType)
		_, err := db.ExecContext(context.Background(), query, userID, cardNumber)

		if err != nil {
			return fmt.Errorf("failed to insert card for user %d: %w", i, err)
		}
	}

	log.Printf("Added cards for %d users", len(userIDs))
	return nil
}

// addPayments добавляет платежи
func addPayments(db *sql.DB, dbType string, userIDs []uuid.UUID) error {
	rows, err := db.QueryContext(context.Background(), `
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
			query := `
				INSERT INTO payments (user_id, card_number, amount, paym_type)
				VALUES ($1, $2, $3, $4)
			`
			query = adaptQuery(query, dbType)
			_, err = db.ExecContext(context.Background(), query, userID, *cardNumber, amount, paymType)
		} else {
			query := `
				INSERT INTO payments (user_id, amount, paym_type)
				VALUES ($1, $2, $3)
			`
			query = adaptQuery(query, dbType)
			_, err = db.ExecContext(context.Background(), query, userID, amount, paymType)
		}

		if err != nil {
			return fmt.Errorf("failed to insert payment %d: %w", i, err)
		}
	}

	log.Printf("Added 5000 payments")
	return nil
}

func genData() error {
	// Загружаем конфигурацию базы данных из database.yaml
	cfg, err := config.LoadDatabaseConfig("../configs/database.yaml")
	if err != nil {
		log.Fatal("Failed to load database config:", err)
		return err
	}

	var db *sql.DB
	var dbType string
	switch cfg.Type {
	case config.DBTypePostgres:
		dbType = DBTypePostgres
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password,
			cfg.Postgres.DBName, cfg.Postgres.SSLMode)
		db, err = sql.Open("pgx", connStr)
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL:", err)
			return err
		}
	case config.DBTypeMySQL:
		dbType = DBTypeMySQL
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port,
			cfg.MySQL.DBName, cfg.MySQL.Params)
		db, err = sql.Open("mysql", connStr)
		if err != nil {
			log.Fatal("Failed to connect to MySQL:", err)
			return err
		}
	default:
		log.Fatal("Unsupported database type:", cfg.Type)
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
	defer db.Close()

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
		return err
	}

	log.Printf("Successfully connected to %s!", cfg.Type)

	// Очищаем существующие данные
	if dbType == DBTypePostgres {
		// Для PostgreSQL используем TRUNCATE ... RESTART IDENTITY
		truncateQuery := `
			TRUNCATE TABLE payments, promocodes, subscriptions, cards, user_referrals, users, sub_durations, services, subscription_plans RESTART IDENTITY CASCADE
		`
		_, err = db.ExecContext(context.Background(), truncateQuery)
		if err != nil {
			log.Println("Warning: could not truncate tables:", err)
		}
	} else {
		// Для MySQL выполняем DELETE FROM каждой таблицы в правильном порядке (с учётом внешних ключей)
		// Порядок: удаляем дочерние таблицы перед родительскими
		tables := []string{
			"payments",
			"promocodes",
			"subscriptions",
			"cards",
			"user_referrals",
			"subscription_plans",
			"sub_durations",
			"services",
			"users",
		}
		for _, table := range tables {
			_, err = db.ExecContext(context.Background(), "DELETE FROM "+table)
			if err != nil {
				log.Printf("Warning: could not delete from %s: %v", table, err)
			}
		}
	}

	// Генерация данных
	userIDs, err := addUsers(db, dbType)
	if err != nil {
		return fmt.Errorf("error adding users: %w", err)
	}

	serviceIDs, planIDs, err := addServices(db, dbType)
	if err != nil {
		return fmt.Errorf("error adding services: %w", err)
	}

	subscriptionIDs, err := addSubscriptions(db, dbType, userIDs, planIDs)
	if err != nil {
		return fmt.Errorf("error adding subscriptions: %w", err)
	}

	if err := addCards(db, dbType, userIDs); err != nil {
		return fmt.Errorf("error adding cards: %w", err)
	}

	if err := addPromocodes(db, dbType, serviceIDs, subscriptionIDs); err != nil {
		return fmt.Errorf("error adding promocodes: %w", err)
	}

	if err := addPayments(db, dbType, userIDs); err != nil {
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
