package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type Lab6Handler struct {
	db     *pgxpool.Pool
	logger *logger.LogrusLogger
}

func NewLab6Handler(db *pgxpool.Pool, logger *logger.LogrusLogger) *Lab6Handler {
	return &Lab6Handler{
		db:     db,
		logger: logger,
	}
}

// 1. Скалярный запрос
func (h *Lab6Handler) ScalarQuery(w http.ResponseWriter, r *http.Request) {
	var result int
	err := h.db.QueryRow(r.Context(), `
		SELECT COUNT(*) FROM subscriptions WHERE price > 1000
	`).Scan(&result)

	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":  "scalar",
		"result":      result,
		"description": "Количество подписок дороже 1000",
	})
}

// 2. Запрос с несколькими JOIN
func (h *Lab6Handler) JoinQuery(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		UserName    string `json:"user_name"`
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
		SubType     string `json:"sub_type"`
	}

	var results []Result

	rows, err := h.db.Query(r.Context(), `
		SELECT u.user_name, s.service_name, sub.price, sub.sub_type
		FROM subscriptions sub
		JOIN users u ON sub.user_id = u.user_id
		JOIN services s ON sub.service_id = s.service_id
		WHERE sub.price > 500
		LIMIT 10
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var res Result
		if err := rows.Scan(&res.UserName, &res.ServiceName, &res.Price, &res.SubType); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		results = append(results, res)
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type": "join",
		"results":    results,
	})
}

// 3. Запрос с CTE и оконными функциями
func (h *Lab6Handler) CTEQuery(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		ServiceName string `json:"service_name"`
		UserName    string `json:"user_name"`
		Price       int    `json:"price"`
		AvgPrice    int    `json:"avg_price"`
		Rank        int    `json:"rank"`
	}

	var results []Result

	rows, err := h.db.Query(r.Context(), `
		WITH service_stats AS (
			SELECT 
				service_id,
				AVG(price) as avg_price
			FROM subscriptions 
			GROUP BY service_id
		)
		SELECT 
			s.service_name,
			u.user_name,
			sub.price,
			ROUND(ss.avg_price) as avg_price,
			RANK() OVER (PARTITION BY s.service_id ORDER BY sub.price DESC) as price_rank
		FROM subscriptions sub
		JOIN users u ON sub.user_id = u.user_id
		JOIN services s ON sub.service_id = s.service_id
		JOIN service_stats ss ON s.service_id = ss.service_id
		WHERE sub.price > ss.avg_price
		ORDER BY s.service_name, price_rank
		LIMIT 15
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var res Result
		if err := rows.Scan(&res.ServiceName, &res.UserName, &res.Price, &res.AvgPrice, &res.Rank); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		results = append(results, res)
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type": "cte_window",
		"results":    results,
	})
}

// 4. Запрос к метаданным
func (h *Lab6Handler) MetadataQuery(w http.ResponseWriter, r *http.Request) {
	type TableInfo struct {
		TableName  string `json:"table_name"`
		ColumnName string `json:"column_name"`
		DataType   string `json:"data_type"`
	}

	var results []TableInfo

	rows, err := h.db.Query(r.Context(), `
		SELECT 
			table_name,
			column_name,
			data_type
		FROM information_schema.columns 
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position
		LIMIT 20
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var res TableInfo
		if err := rows.Scan(&res.TableName, &res.ColumnName, &res.DataType); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		results = append(results, res)
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type": "metadata",
		"results":    results,
	})
}

func (h *Lab6Handler) ScalarFunction(w http.ResponseWriter, r *http.Request) {
	// Сначала создаем функцию в БД
	_, err := h.db.Exec(r.Context(), `
		CREATE OR REPLACE FUNCTION get_total_income_for_service(service_id_input INTEGER)
		RETURNS INTEGER
		LANGUAGE plpgsql
		AS $$
		DECLARE
			total_income INTEGER;
		BEGIN
			SELECT COALESCE(SUM(s.price), 0)
			INTO total_income
			FROM subscriptions s
			WHERE s.service_id = service_id_input
			  AND s.sub_type != 'promocode'; -- промокоды делают бесплатной покупку

			RETURN total_income;
		END;
		$$;
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create function: " + err.Error()})
		return
	}

	// Берем первый сервис из БД для демонстрации
	var serviceID int
	err = h.db.QueryRow(r.Context(), `SELECT service_id FROM services LIMIT 1`).Scan(&serviceID)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "No services found"})
		return
	}

	var totalIncome int
	// Вызываем только что созданную функцию
	err = h.db.QueryRow(r.Context(), `SELECT get_total_income_for_service($1)`, serviceID).Scan(&totalIncome)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to call function: " + err.Error()})
		return
	}

	// Получаем название сервиса для красивого ответа
	var serviceName string
	err = h.db.QueryRow(r.Context(), `SELECT service_name FROM services WHERE service_id = $1`, serviceID).Scan(&serviceName)
	if err != nil {
		serviceName = "Unknown"
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":    "scalar_function",
		"service_id":    serviceID,
		"service_name":  serviceName,
		"total_income":  totalIncome,
		"description":   "Общий доход от сервиса (исключая промокоды)",
		"function_name": "get_total_income_for_service",
		"status":        "Function created and executed successfully",
	})
}

// 6. Табличная функция - создаем и вызываем функцию
func (h *Lab6Handler) TableFunction(w http.ResponseWriter, r *http.Request) {
	// Сначала создаем функцию в БД
	_, err := h.db.Exec(r.Context(), `
		CREATE OR REPLACE FUNCTION get_active_subscribers(service_id_input INTEGER)
		RETURNS TABLE (
			user_id UUID,
			user_name VARCHAR(20),
			email VARCHAR,
			sub_type subscription_type,
			end_date DATE
		)
		LANGUAGE sql
		AS $$
			SELECT u.user_id, u.user_name, u.email, s.sub_type, s.end_date
			FROM subscriptions s
			JOIN users u ON s.user_id = u.user_id
			WHERE s.service_id = service_id_input
			  AND s.end_date >= CURRENT_DATE;
		$$;
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create function: " + err.Error()})
		return
	}

	// Берем первый сервис из БД для демонстрации
	var serviceID int
	err = h.db.QueryRow(r.Context(), `SELECT service_id FROM services LIMIT 1`).Scan(&serviceID)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "No services found"})
		return
	}

	// Получаем название сервиса
	var serviceName string
	err = h.db.QueryRow(r.Context(), `SELECT service_name FROM services WHERE service_id = $1`, serviceID).Scan(&serviceName)
	if err != nil {
		serviceName = "Unknown"
	}

	type ActiveSubscriber struct {
		UserID   string    `json:"user_id"`
		UserName string    `json:"user_name"`
		Email    string    `json:"email"`
		SubType  string    `json:"sub_type"`
		EndDate  time.Time `json:"end_date"`
	}

	var results []ActiveSubscriber

	// Вызываем табличную функцию
	rows, err := h.db.Query(r.Context(), `SELECT * FROM get_active_subscribers($1)`, serviceID)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to call function: " + err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var res ActiveSubscriber
		if err := rows.Scan(&res.UserID, &res.UserName, &res.Email, &res.SubType, &res.EndDate); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		results = append(results, res)
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":    "table_function",
		"service_id":    serviceID,
		"service_name":  serviceName,
		"subscribers":   results,
		"count":         len(results),
		"description":   "Активные подписчики сервиса (подписка не истекла)",
		"function_name": "get_active_subscribers",
		"status":        "Function created and executed successfully",
	})
}

// 7. Хранимая процедура - используем существующий UpdateSubscription
// 7. Хранимая процедура - создаем и вызываем процедуру
func (h *Lab6Handler) StoredProcedure(w http.ResponseWriter, r *http.Request) {
	// Сначала создаем процедуру в БД
	_, err := h.db.Exec(r.Context(), `
		CREATE OR REPLACE PROCEDURE update_balances_with_bonus()
		LANGUAGE plpgsql
		AS $$
		BEGIN
			UPDATE users
			SET balance = balance + LEAST(balance * 0.05, 100);
			COMMIT;
		END;
		$$;
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create procedure: " + err.Error()})
		return
	}

	// Получаем статистику ДО выполнения процедуры
	var beforeStats struct {
		TotalUsers   int     `json:"total_users"`
		TotalBalance int     `json:"total_balance"`
		AvgBalance   float64 `json:"avg_balance"`
	}

	err = h.db.QueryRow(r.Context(), `
		SELECT 
			COUNT(*) as total_users,
			SUM(balance) as total_balance,
			AVG(balance) as avg_balance
		FROM users
	`).Scan(&beforeStats.TotalUsers, &beforeStats.TotalBalance, &beforeStats.AvgBalance)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get before stats: " + err.Error()})
		return
	}

	// Вызываем хранимую процедуру
	_, err = h.db.Exec(r.Context(), `CALL update_balances_with_bonus()`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to call procedure: " + err.Error()})
		return
	}

	// Получаем статистику ПОСЛЕ выполнения процедуры
	var afterStats struct {
		TotalUsers   int     `json:"total_users"`
		TotalBalance int     `json:"total_balance"`
		AvgBalance   float64 `json:"avg_balance"`
	}

	err = h.db.QueryRow(r.Context(), `
		SELECT 
			COUNT(*) as total_users,
			SUM(balance) as total_balance,
			AVG(balance) as avg_balance
		FROM users
	`).Scan(&afterStats.TotalUsers, &afterStats.TotalBalance, &afterStats.AvgBalance)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get after stats: " + err.Error()})
		return
	}

	// Получаем несколько пользователей для демонстрации изменений
	type UserBalance struct {
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
		Balance  int    `json:"balance"`
	}

	var userBalances []UserBalance

	rows, err := h.db.Query(r.Context(), `
		SELECT user_id, user_name, balance 
		FROM users 
		ORDER BY balance DESC 
		LIMIT 5
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get user balances: " + err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user UserBalance
		if err := rows.Scan(&user.UserID, &user.UserName, &user.Balance); err != nil {
			continue
		}
		userBalances = append(userBalances, user)
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":       "stored_procedure",
		"procedure_name":   "update_balances_with_bonus",
		"description":      "Начисление бонуса 5% на баланс (макс. 100)",
		"before_execution": beforeStats,
		"after_execution":  afterStats,
		"balance_increase": afterStats.TotalBalance - beforeStats.TotalBalance,
		"top_users":        userBalances,
		"status":           "Procedure created and executed successfully",
	})
}

// 8. Системная функция
func (h *Lab6Handler) SystemFunction(w http.ResponseWriter, r *http.Request) {
	var (
		dbName            string
		userName          string
		version           string
		activeConnections int
	)

	err := h.db.QueryRow(r.Context(), `
		SELECT 
			current_database(),
			current_user,
			version(),
			(SELECT count(*) FROM pg_stat_activity)
	`).Scan(&dbName, &userName, &version, &activeConnections)

	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":         "system_function",
		"database":           dbName,
		"user":               userName,
		"version":            version,
		"active_connections": activeConnections,
	})
}

// 9. Создание таблицы
func (h *Lab6Handler) CreateTable(w http.ResponseWriter, r *http.Request) {
	_, err := h.db.Exec(r.Context(), `
		CREATE TABLE IF NOT EXISTS subscription_audit (
			audit_id SERIAL PRIMARY KEY,
			sub_id INTEGER REFERENCES subscriptions(sub_id),
			action_type VARCHAR(10) NOT NULL,
			old_price INTEGER,
			new_price INTEGER,
			changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			changed_by VARCHAR(50)
		)
	`)

	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "Таблица subscription_audit создана успешно",
	})
}

// 10. Вставка данных
func (h *Lab6Handler) InsertData(w http.ResponseWriter, r *http.Request) {
	type InsertRequest struct {
		SubID     int    `json:"sub_id"`
		Action    string `json:"action"`
		OldPrice  *int   `json:"old_price"`
		NewPrice  *int   `json:"new_price"`
		ChangedBy string `json:"changed_by"`
	}

	var req InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}

	_, err := h.db.Exec(r.Context(), `
		INSERT INTO subscription_audit (sub_id, action_type, old_price, new_price, changed_by)
		VALUES ($1, $2, $3, $4, $5)
	`, req.SubID, req.Action, req.OldPrice, req.NewPrice, req.ChangedBy)

	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "Данные успешно вставлены в subscription_audit",
	})
}

type AuditRecord struct {
	AuditID    int       `json:"audit_id"`
	SubID      int       `json:"sub_id"`
	ActionType string    `json:"action_type"`
	OldPrice   *int      `json:"old_price,omitempty"`
	NewPrice   *int      `json:"new_price,omitempty"`
	ChangedAt  time.Time `json:"changed_at"`
	ChangedBy  string    `json:"changed_by"`
}

// 11. Показать всю таблицу subscription_audit
func (h *Lab6Handler) ShowAuditTable(w http.ResponseWriter, r *http.Request) {
	// type AuditRecord struct {
	// 	AuditID    int        `json:"audit_id"`
	// 	SubID      int        `json:"sub_id"`
	// 	ActionType string     `json:"action_type"`
	// 	OldPrice   *int       `json:"old_price,omitempty"`
	// 	NewPrice   *int       `json:"new_price,omitempty"`
	// 	ChangedAt  time.Time  `json:"changed_at"`
	// 	ChangedBy  string     `json:"changed_by"`
	// }

	var results []AuditRecord

	// Получаем все записи из таблицы audit
	rows, err := h.db.Query(r.Context(), `
		SELECT 
			audit_id, 
			sub_id, 
			action_type, 
			old_price, 
			new_price, 
			changed_at, 
			changed_by
		FROM subscription_audit 
		ORDER BY changed_at DESC, audit_id DESC
	`)
	if err != nil {
		// Если таблицы не существует, создаем ее и возвращаем пустой результат
		_, createErr := h.db.Exec(r.Context(), `
			CREATE TABLE IF NOT EXISTS subscription_audit (
				audit_id SERIAL PRIMARY KEY,
				sub_id INTEGER REFERENCES subscriptions(sub_id),
				action_type VARCHAR(10) NOT NULL,
				old_price INTEGER,
				new_price INTEGER,
				changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				changed_by VARCHAR(50)
			)
		`)
		if createErr != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to create audit table: " + createErr.Error(),
			})
			return
		}

		utils.MakeResponse(w, http.StatusOK, map[string]any{
			"table_name": "subscription_audit",
			"records":    []AuditRecord{},
			"count":      0,
			"message":    "Table created successfully, but no records found",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var record AuditRecord
		var oldPrice, newPrice *int

		if err := rows.Scan(
			&record.AuditID,
			&record.SubID,
			&record.ActionType,
			&oldPrice,
			&newPrice,
			&record.ChangedAt,
			&record.ChangedBy,
		); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to scan record: " + err.Error(),
			})
			return
		}

		// Копируем указатели
		record.OldPrice = oldPrice
		record.NewPrice = newPrice

		results = append(results, record)
	}

	if err := rows.Err(); err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Error iterating rows: " + err.Error(),
		})
		return
	}

	// Форматируем даты для красивого вывода
	formattedResults := make([]map[string]any, len(results))
	for i, record := range results {
		formattedResults[i] = map[string]any{
			"audit_id":           record.AuditID,
			"sub_id":             record.SubID,
			"action_type":        record.ActionType,
			"old_price":          record.OldPrice,
			"new_price":          record.NewPrice,
			"changed_at":         record.ChangedAt.Format("2006-01-02 15:04:05"),
			"changed_by":         record.ChangedBy,
			"change_description": h.getChangeDescription(record),
		}
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"table_name": "subscription_audit",
		"records":    formattedResults,
		"count":      len(results),
		"summary": map[string]any{
			"total_records": len(results),
			"actions_count": h.countActions(results),
			"last_update":   h.getLastUpdate(results),
		},
	})
}

// Вспомогательная функция для описания изменений
func (h *Lab6Handler) getChangeDescription(record AuditRecord) string {
	switch record.ActionType {
	case "create":
		if record.NewPrice != nil {
			return fmt.Sprintf("Создана подписка с ценой %d", *record.NewPrice)
		}
		return "Создана подписка"
	case "update":
		if record.OldPrice != nil && record.NewPrice != nil {
			return fmt.Sprintf("Цена изменена с %d на %d", *record.OldPrice, *record.NewPrice)
		}
		return "Обновление подписки"
	case "delete":
		if record.OldPrice != nil {
			return fmt.Sprintf("Удалена подписка с ценой %d", *record.OldPrice)
		}
		return "Удалена подписка"
	default:
		return "Изменение подписки"
	}
}

// Подсчет действий по типам
func (h *Lab6Handler) countActions(records []AuditRecord) map[string]int {
	counts := make(map[string]int)
	for _, record := range records {
		counts[record.ActionType]++
	}
	return counts
}

// Получение времени последнего обновления
func (h *Lab6Handler) getLastUpdate(records []AuditRecord) string {
	if len(records) == 0 {
		return "No records"
	}
	return records[0].ChangedAt.Format("2006-01-02 15:04:05")
}

// 12. Действительные промокоды Netflix для пользователей с истекающими подписками
type PromocodeInfo struct {
	ServiceName         string    `json:"service_name"`
	UserName            string    `json:"user_name"`
	UserEmail           string    `json:"user_email"`
	Promocode           string    `json:"promocode"`
	ExpiresAt           time.Time `json:"expires_at"`
	SubscriptionEndDate time.Time `json:"subscription_end_date"`
	DaysUntilExpiry     int       `json:"days_until_expiry"`
}

// 12. Действительные промокоды Netflix для пользователей с истекающими подписками
func (h *Lab6Handler) GetNetflixPromocodes(w http.ResponseWriter, r *http.Request) {
	// type PromocodeInfo struct {
	// 	ServiceName string    `json:"service_name"`
	// 	UserName    string    `json:"user_name"`
	// 	UserEmail   string    `json:"user_email"`
	// 	Promocode   string    `json:"promocode"`
	// 	ExpiresAt   time.Time `json:"expires_at"`
	// 	SubscriptionEndDate time.Time `json:"subscription_end_date"`
	// 	DaysUntilExpiry int  `json:"days_until_expiry"`
	// }

	var results []PromocodeInfo

	rows, err := h.db.Query(r.Context(), `
		SELECT 
			s.service_name,
			u.user_name,
			u.email,
			p.promocode,
			p.expires_at,
			sub.end_date as subscription_end_date,
			(sub.end_date - CURRENT_DATE) as days_until_expiry
		FROM promocodes p
		JOIN services s ON p.service_id = s.service_id
		JOIN subscriptions sub ON p.sub_id = sub.sub_id
		JOIN users u ON sub.user_id = u.user_id
		WHERE s.service_name = 'Netflix'
		  AND p.expires_at >= CURRENT_DATE
		  AND sub.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '70 days'
		  AND sub.end_date IS NOT NULL
		ORDER BY days_until_expiry ASC, p.expires_at ASC
	`)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to query promocodes: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result PromocodeInfo
		var daysUntilExpiry int

		if err := rows.Scan(
			&result.ServiceName,
			&result.UserName,
			&result.UserEmail,
			&result.Promocode,
			&result.ExpiresAt,
			&result.SubscriptionEndDate,
			&daysUntilExpiry,
		); err != nil {
			utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to scan record: " + err.Error(),
			})
			return
		}

		result.DaysUntilExpiry = daysUntilExpiry
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "Error iterating rows: " + err.Error(),
		})
		return
	}

	// Форматируем даты для красивого вывода
	formattedResults := make([]map[string]any, len(results))
	for i, record := range results {
		formattedResults[i] = map[string]any{
			"service_name":      record.ServiceName,
			"user_name":         record.UserName,
			"user_email":        record.UserEmail,
			"promocode":         record.Promocode,
			"promocode_expires": record.ExpiresAt.Format("2006-01-02"),
			"subscription_ends": record.SubscriptionEndDate.Format("2006-01-02"),
			"days_until_expiry": record.DaysUntilExpiry,
			"urgency_level":     h.getUrgencyLevel(record.DaysUntilExpiry),
			"message":           h.generateMessage(record),
		}
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"query_type":  "netflix_promocodes",
		"description": "Действительные промокоды Netflix для пользователей с истекающими подписками (в течение 70 дней)",
		"results":     formattedResults,
		"summary": map[string]any{
			"total_promocodes": len(results),
			"users_at_risk":    h.countUniqueUsers(results),
			"urgency_stats":    h.getUrgencyStats(results),
		},
	})
}

// Вспомогательная функция для определения уровня срочности
func (h *Lab6Handler) getUrgencyLevel(days int) string {
	switch {
	case days <= 1:
		return "CRITICAL"
	case days <= 3:
		return "HIGH"
	case days <= 5:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

// Генерация сообщения для пользователя
func (h *Lab6Handler) generateMessage(record PromocodeInfo) string {
	return fmt.Sprintf("У %s подписка истекает через %d дней. Промокод %s действителен до %s",
		record.UserName, record.DaysUntilExpiry, record.Promocode, record.ExpiresAt.Format("2006-01-02"))
}

// Подсчет уникальных пользователей
func (h *Lab6Handler) countUniqueUsers(records []PromocodeInfo) int {
	users := make(map[string]bool)
	for _, record := range records {
		users[record.UserEmail] = true
	}
	return len(users)
}

// Статистика по срочности
func (h *Lab6Handler) getUrgencyStats(records []PromocodeInfo) map[string]int {
	stats := map[string]int{
		"CRITICAL": 0,
		"HIGH":     0,
		"MEDIUM":   0,
		"LOW":      0,
	}
	for _, record := range records {
		stats[h.getUrgencyLevel(record.DaysUntilExpiry)]++
	}
	return stats
}

// Универсальный обработчик для всех пунктов
func (h *Lab6Handler) Lab6Handler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	numStr := vars["num"]

	num, err := strconv.Atoi(numStr)
	if err != nil || num < 1 || num > 12 {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{"error": "Invalid task number (1-10)"})
		return
	}

	switch num {
	case 1:
		h.ScalarQuery(w, r)
	case 2:
		h.JoinQuery(w, r)
	case 3:
		h.CTEQuery(w, r)
	case 4:
		h.MetadataQuery(w, r)
	case 5:
		h.ScalarFunction(w, r)
	case 6:
		h.TableFunction(w, r)
	case 7:
		h.StoredProcedure(w, r)
	case 8:
		h.SystemFunction(w, r)
	case 9:
		h.CreateTable(w, r)
	case 10:
		h.InsertData(w, r)
	case 11:
		h.ShowAuditTable(w, r)
	case 12:
		h.GetNetflixPromocodes(w, r)
	}
}

// Вспомогательная структура для перехвата ответов
type responseWriter struct {
	http.ResponseWriter
	body []byte
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
}
